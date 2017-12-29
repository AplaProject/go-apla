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
	"strconv"
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

// ErrFirstBlockHostIsEmpty host for first block is not specified
var ErrFirstBlockHostIsEmpty = errors.New("FirstBlockHost is empty")

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
	err := model.ExecSchemaEcosystem(nil, 1, myAddress, ``)
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
	node := &model.SystemParameter{Name: `full_nodes`}
	if err = node.SaveArray([][]string{{data.Host, converter.Int64ToStr(myAddress),
		hex.EncodeToString(data.NodePublicKey)}}); err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("saving node array")
		return p.ErrInfo(err)
	}
	commission := &model.SystemParameter{Name: `commission_wallet`}
	if err = commission.SaveArray([][]string{{"1", converter.Int64ToStr(myAddress)}}); err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("saving commission_wallet array")
		return p.ErrInfo(err)
	}
	if err = syspar.SysUpdate(nil); err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("updating syspar")
		return p.ErrInfo(err)
	}
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
	if len(*conf.FirstBlockPublicKey) == 0 {
		priv, pub, _ := crypto.GenHexKeys()

		privFile := filepath.Join(conf.Config.PrivateDir, consts.PrivateKeyFilename)
		if err := ioutil.WriteFile(privFile, []byte(priv), 0644); err != nil {
			log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("writing private key file")
			return err
		}
		pubFile := filepath.Join(conf.Config.PrivateDir, consts.PublicKeyFilename)
		if err := ioutil.WriteFile(pubFile, []byte(pub), 0644); err != nil {
			log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("writing public key file")
			return err
		}
		*conf.FirstBlockPublicKey = pub
	}

	if len(*conf.FirstBlockNodePublicKey) == 0 {
		priv, pub, _ := crypto.GenHexKeys()
		nodePrivFile := filepath.Join(conf.Config.PrivateDir, consts.NodePrivateKeyFilename)
		if err := ioutil.WriteFile(nodePrivFile, []byte(priv), 0644); err != nil {
			log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("writing node private key file")
			return err
		}
		nodePubFile := filepath.Join(conf.Config.PrivateDir, consts.NodePublicKeyFilename)
		if err := ioutil.WriteFile(nodePubFile, []byte(pub), 0644); err != nil {
			log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("writing node public key file")
			return err
		}
		*conf.FirstBlockNodePublicKey = pub
	}

	PublicKey := *conf.FirstBlockPublicKey
	PublicKeyBytes, err := hex.DecodeString(string(PublicKey))
	if err != nil {
		log.WithFields(log.Fields{"type": consts.ConversionError, "error": err}).Error("decoding public key from hex to string")
		return err
	}

	NodePublicKey := *conf.FirstBlockNodePublicKey
	NodePublicKeyBytes, err := hex.DecodeString(string(NodePublicKey))
	if err != nil {
		log.WithFields(log.Fields{"type": consts.ConversionError, "error": err}).Error("decoding node public key from hex to string")
		return err
	}

	Host := *conf.FirstBlockHost
	if len(Host) == 0 {
		return ErrFirstBlockHostIsEmpty
	}

	iAddress := int64(crypto.Address(PublicKeyBytes))
	conf.Config.KeyID = iAddress

	keyIDFile := filepath.Join(conf.Config.PrivateDir, consts.KeyIDFilename)
	if err := ioutil.WriteFile(keyIDFile, []byte(strconv.FormatInt(iAddress, 10)), 0644); err != nil {
		log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("writing keyID file")
		return err
	}

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

	key, err := ioutil.ReadFile(filepath.Join(conf.Config.PrivateDir, consts.PrivateKeyFilename))
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
