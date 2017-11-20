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
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

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

type FirstBlockParser struct {
	*Parser
}

func (p *FirstBlockParser) Init() error {
	return nil
}

func (p *FirstBlockParser) Validate() error {
	return nil
}

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

func (p *FirstBlockParser) Rollback() error {
	return nil
}

func (p FirstBlockParser) Header() *tx.Header {
	return nil
}

// FirstBlock generates the first block
func FirstBlock() {
	if len(*utils.FirstBlockPublicKey) == 0 {
		priv, pub, _ := crypto.GenHexKeys()
		err := ioutil.WriteFile(*utils.Dir+"/PrivateKey", []byte(priv), 0644)
		if err != nil {
			log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("writing private key file")
			return
		}
		*utils.FirstBlockPublicKey = pub
	}
	if len(*utils.FirstBlockNodePublicKey) == 0 {
		priv, pub, _ := crypto.GenHexKeys()
		err := ioutil.WriteFile(*utils.Dir+"/NodePrivateKey", []byte(priv), 0644)
		if err != nil {
			log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("writing node private key file")
			return
		}
		*utils.FirstBlockNodePublicKey = pub
	}

	PublicKey := *utils.FirstBlockPublicKey
	PublicKeyBytes, err := hex.DecodeString(string(PublicKey))
	if err != nil {
		log.WithFields(log.Fields{"type": consts.ConversionError, "error": err}).Error("decoding public key from hex to string")
		return
	}

	NodePublicKey := *utils.FirstBlockNodePublicKey
	NodePublicKeyBytes, err := hex.DecodeString(string(NodePublicKey))
	if err != nil {
		log.WithFields(log.Fields{"type": consts.ConversionError, "error": err}).Error("decoding node public key from hex to string")
		return
	}

	Host := *utils.FirstBlockHost
	if len(Host) == 0 {
		log.Info("first block host is empty, using localhost as host")
		Host = "127.0.0.1"
	}

	iAddress := int64(crypto.Address(PublicKeyBytes))
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
		return
	}

	block, err := MarshallBlock(header, [][]byte{tx}, []byte("0"), "")
	if err != nil {
		return
	}

	firstBlockDir := ""
	if len(*utils.FirstBlockDir) == 0 {
		firstBlockDir = *utils.Dir
	} else {
		firstBlockDir = filepath.Join("", *utils.FirstBlockDir)
		if _, err := os.Stat(firstBlockDir); os.IsNotExist(err) {
			if err = os.Mkdir(firstBlockDir, 0755); err != nil {
				log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("creating first block dir directory")
				return
			}
		}
	}
	ioutil.WriteFile(filepath.Join(firstBlockDir, "1block"), block, 0644)
}
