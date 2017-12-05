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

package parser

import (
	"encoding/hex"
	"errors"
	"io/ioutil"
	"path/filepath"
	"time"

	"github.com/AplaProject/go-apla/packages/conf"

	"github.com/AplaProject/go-apla/packages/config/syspar"
	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/crypto"
	"github.com/AplaProject/go-apla/packages/model"
	"github.com/AplaProject/go-apla/packages/smart"
	"github.com/AplaProject/go-apla/packages/utils"
	"github.com/AplaProject/go-apla/packages/utils/tx"

	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
)

// FirstBlockParser is parser wrapper
type FirstBlockParser struct {
	*Parser
}

// Init first block
func (p *FirstBlockParser) Init() error {
	return nil
}

// Validate first block
func (p *FirstBlockParser) Validate() error {
	return nil
}

// Action is fires first block
func (p *FirstBlockParser) Action() error {
	logger := p.GetLogger()
	data := p.TxPtr.(*consts.FirstBlock)
	myAddress := crypto.Address(data.PublicKey)
	err := model.ExecSchemaEcosystem(1, myAddress, ``)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("executing ecosystem schema")
		return p.ErrInfo(err)
	}
	err = model.GetDB(p.DbTransaction).Exec(`insert into "1_keys" (id,pub,amount) values(?, ?,?)`,
		myAddress, data.PublicKey, decimal.NewFromFloat(consts.FIRST_QDLT).String()).Error
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("inserting default page")
		return p.ErrInfo(err)
	}
	err = model.GetDB(p.DbTransaction).Exec(`insert into "1_pages" (id,name,menu,value,conditions) values('1', 'default_page',
		  'default_menu', ?, 'ContractAccess("@1EditPage")')`, syspar.SysString(`default_ecosystem_page`)).Error
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("inserting default page")
		return p.ErrInfo(err)
	}
	err = model.GetDB(p.DbTransaction).Exec(`insert into "1_menu" (id,name,value,title,conditions) values('1', 'default_menu', ?, ?, 'ContractAccess("@1EditMenu")')`,
		syspar.SysString(`default_ecosystem_menu`), `default`).Error
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("inserting default menu")
		return p.ErrInfo(err)
	}
	err = smart.LoadContract(p.DbTransaction, `1`)
	if err != nil {
		return p.ErrInfo(err)
	}
	node := &model.SystemParameterV2{Name: `full_nodes`}
	if err = node.SaveArray([][]string{{data.Host, converter.Int64ToStr(myAddress),
		hex.EncodeToString(data.NodePublicKey)}}); err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("saving node array")
		return p.ErrInfo(err)
	}
	commission := &model.SystemParameterV2{Name: `commission_wallet`}
	if err = commission.SaveArray([][]string{{"1", converter.Int64ToStr(myAddress)}}); err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("saving commission_wallet array")
		return p.ErrInfo(err)
	}
	syspar.SysUpdate()
	return nil
}

// Rollback first block
func (p *FirstBlockParser) Rollback() error {
	return nil
}

// Header is returns first block header
func (p FirstBlockParser) Header() *tx.Header {
	return nil
}

// GenerateFirstBlock generates the first block
func GenerateFirstBlock() error {
	if len(*utils.FirstBlockPublicKey) == 0 {
		priv, pub, _ := crypto.GenHexKeys()
		pk := filepath.Join(conf.Config.PrivateDir, "/PrivateKey")
		if err := ioutil.WriteFile(pk, []byte(priv), 0644); err != nil {
			log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("writing private key file")
			return err
		}
		pubk := filepath.Join(conf.Config.PrivateDir, "/PublicKey")
		if err := ioutil.WriteFile(pubk, []byte(pub), 0644); err != nil {
			log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("writing public key file")
			return err
		}
		*utils.FirstBlockPublicKey = pub
	}

	if len(*utils.FirstBlockNodePublicKey) == 0 {
		priv, pub, _ := crypto.GenHexKeys()
		nvk := filepath.Join(conf.Config.PrivateDir, "/NodePrivateKey")
		if err := ioutil.WriteFile(nvk, []byte(priv), 0644); err != nil {
			log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("writing node private key file")
			return err
		}
		npk := filepath.Join(conf.Config.PrivateDir, "/NodePublicKey")
		if err := ioutil.WriteFile(npk, []byte(pub), 0644); err != nil {
			log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("writing node public key file")
			return err
		}
		*utils.FirstBlockNodePublicKey = pub
	}

	PublicKey := *utils.FirstBlockPublicKey
	PublicKeyBytes, err := hex.DecodeString(string(PublicKey))
	if err != nil {
		log.WithFields(log.Fields{"type": consts.ConversionError, "error": err}).Error("decoding public key from hex to string")
		return err
	}

	NodePublicKey := *utils.FirstBlockNodePublicKey
	NodePublicKeyBytes, err := hex.DecodeString(string(NodePublicKey))
	if err != nil {
		log.WithFields(log.Fields{"type": consts.ConversionError, "error": err}).Error("decoding node public key from hex to string")
		return err
	}

	Host := *utils.FirstBlockHost
	if len(Host) == 0 {
		return errors.New("FirstBlockHost is empty")
	}

	iAddress := int64(crypto.Address(PublicKeyBytes))
	conf.Config.KeyID = iAddress

	now := time.Now().Unix()

	header := &utils.BlockData{
		BlockID:      1,
		Time:         now,
		EcosystemID:  0,
		KeyID:        iAddress,
		NodePosition: 0,
		Version:      consts.BLOCK_VERSION,
	}
	var tx []byte
	_, err = converter.BinMarshal(&tx,
		&consts.FirstBlock{
			TxHeader: consts.TxHeader{
				Type:  1, // FirstBlock
				Time:  uint32(now),
				KeyID: iAddress,
			},
			PublicKey:     PublicKeyBytes,
			NodePublicKey: NodePublicKeyBytes,
			Host:          string(Host),
		},
	)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.MarshallingError, "error": err}).Error("first block body bin marshalling")
		return err
	}

	block, err := MarshallBlock(header, [][]byte{tx}, []byte("0"), "")
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(*conf.FirstBlockPath, block, 0644); err != nil {
		log.WithFields(log.Fields{"type": consts.IOError, "error": err, "file": *conf.FirstBlockPath}).Error("first block write")
		return err
	}

	return nil
}

// GetKeyIDFromPrivateKey load KeyID fron PrivateKey file
func GetKeyIDFromPrivateKey() (int64, error) {

	key, err := ioutil.ReadFile(conf.Config.PrivateDir + "/PrivateKey")
	if err != nil {
		log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("reading private key file")
		return 0, err
	}
	key, err = hex.DecodeString(string(key))
	if err != nil {
		log.WithFields(log.Fields{"type": consts.ConversionError, "error": err}).Error("decoding private key from hex")
		return 0, err
	}
	key, err = crypto.PrivateToPublic(key)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.CryptoError, "error": err}).Error("converting private key to public")
		return 0, err
	}

	return crypto.Address(key), nil
}

// GetKeyIDFromPublicKey load KeyID fron PublicKey file
func GetKeyIDFromPublicKey() (int64, error) {

	key, err := ioutil.ReadFile(conf.Config.PrivateDir + "/PublicKey")
	if err != nil {
		log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("reading public key file")
		return 0, err
	}
	return crypto.Address(key), nil
}
