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
	"time"

	"github.com/EGaaS/go-egaas-mvp/packages/config/syspar"
	"github.com/EGaaS/go-egaas-mvp/packages/consts"
	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/crypto"
	"github.com/EGaaS/go-egaas-mvp/packages/model"
	"github.com/EGaaS/go-egaas-mvp/packages/template"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
	"github.com/EGaaS/go-egaas-mvp/packages/utils/tx"

	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/shopspring/decimal"
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
	data := p.TxPtr.(*consts.FirstBlock)
	myAddress := crypto.Address(data.PublicKey)
	err := model.ExecSchemaEcosystem(1, myAddress, ``)
	if err != nil {
		return p.ErrInfo(err)
	}
	key := &model.Key{
		ID:        myAddress,
		PublicKey: data.PublicKey,
		Amount:    decimal.NewFromFloat(consts.FIRST_QDLT).String(),
	}
	if err = key.SetTablePrefix(consts.MainEco).Create(); err != nil {
		return p.ErrInfo(err)
	}
	err = template.LoadContract(p.DbTransaction, `1`)
	if err != nil {
		return p.ErrInfo(err)
	}
	node := &model.SystemParameterV2{Name: `full_nodes`}
	if err = node.SaveArray([][]string{{data.Host, converter.Int64ToStr(myAddress),
		hex.EncodeToString(data.NodePublicKey)}}); err != nil {
		return p.ErrInfo(err)
	}
	syspar.SysUpdate()
	fullNode := &model.FullNode{WalletID: myAddress, Host: data.Host}
	err = fullNode.Create(p.DbTransaction)
	if err != nil {
		return p.ErrInfo(err)
	}

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
	log.Debug("FirstBlock")

	if len(*utils.FirstBlockPublicKey) == 0 {
		log.Debug("len(*FirstBlockPublicKey) == 0")
		priv, pub, _ := crypto.GenHexKeys()
		err := ioutil.WriteFile(*utils.Dir+"/PrivateKey", []byte(priv), 0644)
		if err != nil {
			log.Error("write publick key failed: %v", utils.ErrInfo(err))
			return
		}
		log.Debugf("public key: %s", pub)
		*utils.FirstBlockPublicKey = pub
	}
	if len(*utils.FirstBlockNodePublicKey) == 0 {
		log.Debug("len(*FirstBlockNodePublicKey) == 0")
		priv, pub, _ := crypto.GenHexKeys()
		err := ioutil.WriteFile(*utils.Dir+"/NodePrivateKey", []byte(priv), 0644)
		if err != nil {
			log.Error("write private kery failed: %v", utils.ErrInfo(err))
			return
		}
		*utils.FirstBlockNodePublicKey = pub
	}

	PublicKey := *utils.FirstBlockPublicKey
	PublicKeyBytes, err := hex.DecodeString(string(PublicKey))
	if err != nil {
		log.Errorf("can't generate key, decode string failed: %s", err)
		return
	}

	NodePublicKey := *utils.FirstBlockNodePublicKey
	NodePublicKeyBytes, err := hex.DecodeString(string(NodePublicKey))
	if err != nil {
		log.Errorf("can't generate key, decode string failed: %s", err)
		return
	}

	Host := *utils.FirstBlockHost
	if len(Host) == 0 {
		Host = "127.0.0.1"
	}

	iAddress := int64(crypto.Address(PublicKeyBytes))
	now := time.Now().Unix()

	header := &utils.BlockData{
		BlockID:  1,
		Time:     now,
		WalletID: iAddress,
		Version:  consts.BLOCK_VERSION,
	}
	var tx []byte
	_, err = converter.BinMarshal(&tx,
		&consts.FirstBlock{
			TxHeader: consts.TxHeader{
				Type:      1, // FirstBlock
				Time:      uint32(now),
				WalletID:  iAddress,
				CitizenID: 0,
			},
			PublicKey:     PublicKeyBytes,
			NodePublicKey: NodePublicKeyBytes,
			Host:          string(Host),
		},
	)
	if err != nil {
		log.Errorf("first block body marshal error: %v", utils.ErrInfo(err))
		return
	}

	log.Debugf("start marshalling first block")
	block, err := MarshallBlock(header, [][]byte{tx}, []byte("0"), "")
	if err != nil {
		log.Errorf("block marshalling failed: %s", err)
		return
	}

	firstBlockDir := ""
	if len(*utils.FirstBlockDir) == 0 {
		firstBlockDir = *utils.Dir
	} else {
		firstBlockDir = filepath.Join("", *utils.FirstBlockDir)
		if _, err := os.Stat(firstBlockDir); os.IsNotExist(err) {
			if err = os.Mkdir(firstBlockDir, 0755); err != nil {
				log.Error("can't create directory for 1block: %v", utils.ErrInfo(err))
				return
			}
		}
	}
	log.Debugf("write first block to: %s/1block", firstBlockDir)
	ioutil.WriteFile(filepath.Join(firstBlockDir, "1block"), block, 0644)
}
