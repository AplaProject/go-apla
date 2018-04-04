// Copyright 2016 The go-daylight Authors
// This file is part of the go-daylight library.
//
// The go-daylight library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-daylight library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-daylight library. If not, see <http://www.gnu.org/licenses/>.

package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/GenesisKernel/go-genesis/packages/conf"
	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/notificator"
	"github.com/GenesisKernel/go-genesis/packages/publisher"

	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/crypto"
	"github.com/GenesisKernel/go-genesis/packages/model"

	"encoding/hex"
	"encoding/json"

	"github.com/GenesisKernel/go-genesis/packages/script"
	"github.com/GenesisKernel/go-genesis/packages/smart"
	"github.com/GenesisKernel/go-genesis/packages/utils"
	"github.com/GenesisKernel/go-genesis/packages/utils/tx"
	"github.com/dgrijalva/jwt-go"
	log "github.com/sirupsen/logrus"
)

type loginResult struct {
	Token       string        `json:"token,omitempty"`
	Refresh     string        `json:"refresh,omitempty"`
	EcosystemID string        `json:"ecosystem_id,omitempty"`
	KeyID       string        `json:"key_id,omitempty"`
	Address     string        `json:"address,omitempty"`
	NotifyKey   string        `json:"notify_key,omitempty"`
	IsNode      bool          `json:"isnode,omitempty"`
	IsOwner     bool          `json:"isowner,omitempty"`
	IsVDE       bool          `json:"vde,omitempty"`
	Timestamp   string        `json:"timestamp,omitempty"`
	Roles       []rolesResult `json:"roles,omitempty"`
}

type rolesResult struct {
	RoleId   int64  `json:"role_id"`
	RoleName string `json:"role_name"`
}

func login(w http.ResponseWriter, r *http.Request, data *apiData, logger *log.Entry) error {
	var (
		pubkey []byte
		wallet int64
		msg    string
		err    error
	)

	if data.token != nil && data.token.Valid {
		if claims, ok := data.token.Claims.(*JWTClaims); ok {
			msg = claims.UID
		}
	}
	if len(msg) == 0 {
		logger.WithFields(log.Fields{"type": consts.EmptyObject}).Error("UID is empty")
		return errorAPI(w, `E_UNKNOWNUID`, http.StatusBadRequest)
	}
	state := data.ecosystemId
	if data.params[`ecosystem`].(int64) > 0 {
		state = data.params[`ecosystem`].(int64)
	}
	if state == 0 {
		logger.WithFields(log.Fields{"type": consts.EmptyObject}).Warning("state is empty, using 1 as a state")
		state = 1
	}
	if len(data.params[`key_id`].(string)) > 0 {
		wallet = converter.StringToAddress(data.params[`key_id`].(string))
	} else if len(data.params[`pubkey`].([]byte)) > 0 {
		wallet = crypto.Address(data.params[`pubkey`].([]byte))
	}
	account := &model.Key{}
	account.SetTablePrefix(state)
	isAccount, err := account.Get(wallet)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("selecting public key from keys")
		return errorAPI(w, err, http.StatusBadRequest)
	}
	if isAccount {
		pubkey = account.PublicKey
		if account.Delete == 1 {
			return errorAPI(w, `E_DELETEDKEY`, http.StatusForbidden)
		}
	} else {
		pubkey = data.params[`pubkey`].([]byte)
		if len(pubkey) == 0 {
			logger.WithFields(log.Fields{"type": consts.EmptyObject}).Error("public key is empty")
			return errorAPI(w, `E_EMPTYPUBLIC`, http.StatusBadRequest)
		}
		NodePrivateKey, NodePublicKey, err := utils.GetNodeKeys()
		if err != nil || len(NodePrivateKey) < 1 {
			if err == nil {
				log.WithFields(log.Fields{"type": consts.EmptyObject}).Error("node private key is empty")
			}
			return err
		}

		hexPubKey := hex.EncodeToString(pubkey)
		params := make([]byte, 0)
		params = append(append(params, converter.EncodeLength(int64(len(hexPubKey)))...), hexPubKey...)

		vm := smart.GetVM(false, 0)
		contract := smart.VMGetContract(vm, "NewUser", 1)
		info := contract.Block.Info.(*script.ContractInfo)

		err = tx.BuildTransaction(tx.SmartContract{
			Header: tx.Header{
				Type:        int(info.ID),
				Time:        time.Now().Unix(),
				EcosystemID: 1,
				KeyID:       conf.Config.KeyID,
			},
			SignedBy: smart.PubToID(NodePublicKey),
			Data:     params,
		}, NodePrivateKey, NodePublicKey, string(hexPubKey))
		if err != nil {
			log.WithFields(log.Fields{"type": consts.ContractError}).Error("Executing contract")
		}
	}
	if state > 1 && len(pubkey) == 0 {
		logger.WithFields(log.Fields{"type": consts.EmptyObject}).Error("public key is empty, and state is not default")
		return errorAPI(w, `E_STATELOGIN`, http.StatusForbidden, wallet, state)
	}

	if r, ok := data.params["role_id"]; ok {
		role := r.(int64)
		if role > 0 {
			ok, err := model.MemberHasRole(nil, state, wallet, role)
			if err != nil {
				logger.WithFields(log.Fields{
					"type":      consts.DBError,
					"member":    wallet,
					"role":      role,
					"ecosystem": state}).Error("check role")

				return errorAPI(w, "E_CHECKROLE", http.StatusInternalServerError)
			}

			if !ok {
				logger.WithFields(log.Fields{
					"type":      consts.NotFound,
					"member":    wallet,
					"role":      role,
					"ecosystem": state,
				}).Error("member hasn't role")

				return errorAPI(w, "E_CHECKROLE", http.StatusNotFound)
			}

			data.roleId = role
		}
	}

	verify, err := crypto.CheckSign(pubkey, msg, data.params[`signature`].([]byte))
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.CryptoError, "pubkey": pubkey, "msg": msg, "signature": string(data.params["signature"].([]byte))}).Error("checking signature")
		return errorAPI(w, err, http.StatusBadRequest)
	}
	if !verify {
		logger.WithFields(log.Fields{"type": consts.InvalidObject, "pubkey": pubkey, "msg": msg, "signature": string(data.params["signature"].([]byte))}).Error("incorrect signature")
		return errorAPI(w, `E_SIGNATURE`, http.StatusBadRequest)
	}

	address := crypto.KeyToAddress(pubkey)
	var (
		sp      model.StateParameter
		founder int64
	)
	sp.SetTablePrefix(converter.Int64ToStr(state))
	if ok, err := sp.Get(nil, "founder_account"); err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting founder_account parameter")
		return errorAPI(w, `E_SERVER`, http.StatusBadRequest)
	} else if ok {
		founder = converter.StrToInt64(sp.Value)
	}

	result := loginResult{EcosystemID: converter.Int64ToStr(state), KeyID: converter.Int64ToStr(wallet),
		Address: address, IsOwner: founder == wallet, IsNode: conf.Config.KeyID == wallet,
		IsVDE: model.IsTable(fmt.Sprintf(`%d_vde_tables`, state))}
	data.result = &result
	expire := data.params[`expire`].(int64)
	if expire == 0 {
		logger.WithFields(log.Fields{"type": consts.JWTError, "expire": jwtExpire}).Warning("using expire from jwt")
		expire = jwtExpire
	}

	isMobile := "0"
	if mob, ok := data.params[`mobile`]; ok && mob != nil {
		if mob.(string) == `1` || mob.(string) == `true` {
			isMobile = `1`
		}
	}
	claims := JWTClaims{
		KeyID:       result.KeyID,
		EcosystemID: result.EcosystemID,
		IsMobile:    isMobile,
		RoleID:      converter.Int64ToStr(data.roleId),
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Second * time.Duration(expire)).Unix(),
		},
	}
	result.Token, err = jwtGenerateToken(w, claims)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.JWTError, "error": err}).Error("generating jwt token")
		return errorAPI(w, err, http.StatusInternalServerError)
	}
	claims.StandardClaims.ExpiresAt = time.Now().Add(time.Hour * 30 * 24).Unix()
	result.Refresh, err = jwtGenerateToken(w, claims)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.JWTError, "error": err}).Error("generating jwt token")
		return errorAPI(w, err, http.StatusInternalServerError)
	}
	result.NotifyKey, result.Timestamp, err = publisher.GetHMACSign(wallet)
	if err != nil {
		return errorAPI(w, err, http.StatusInternalServerError)
	}
	notificator.AddUser(wallet, state)
	notificator.UpdateNotifications(state, []int64{wallet})

	ra := &model.RolesParticipants{}
	roles, err := ra.SetTablePrefix(state).GetActiveMemberRoles(wallet)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting roles")
		return errorAPI(w, `E_SERVER`, http.StatusBadRequest)
	}

	for _, r := range roles {
		var res map[string]string
		if err := json.Unmarshal([]byte(r.Role), &res); err != nil {
			log.WithFields(log.Fields{"type": consts.JSONUnmarshallError, "error": err}).Error("unmarshalling role")
			return errorAPI(w, `E_SERVER`, http.StatusInternalServerError)
		} else {
			result.Roles = append(result.Roles, rolesResult{RoleId: converter.StrToInt64(res["id"]), RoleName: res["name"]})
		}
	}

	return nil
}
