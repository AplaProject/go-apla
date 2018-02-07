//MIT License
//
//Copyright (c) 2016 GenesisKernel
//
//Permission is hereby granted, free of charge, to any person obtaining a copy
//of this software and associated documentation files (the "Software"), to deal
//in the Software without restriction, including without limitation the rights
//to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
//copies of the Software, and to permit persons to whom the Software is
//furnished to do so, subject to the following conditions:
//
//The above copyright notice and this permission notice shall be included in all
//copies or substantial portions of the Software.
//
//THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
//IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
//FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
//AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
//LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
//OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
//SOFTWARE.
package parser

import (
	"encoding/hex"
	"errors"
	"io/ioutil"
	"path/filepath"

	"github.com/GenesisCommunity/go-genesis/packages/conf"

	"github.com/GenesisCommunity/go-genesis/packages/config/syspar"
	"github.com/GenesisCommunity/go-genesis/packages/consts"
	"github.com/GenesisCommunity/go-genesis/packages/converter"
	"github.com/GenesisCommunity/go-genesis/packages/crypto"
	"github.com/GenesisCommunity/go-genesis/packages/model"
	"github.com/GenesisCommunity/go-genesis/packages/smart"
	"github.com/GenesisCommunity/go-genesis/packages/utils/tx"

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
	err := model.ExecSchemaEcosystem(nil, 1, myAddress, ``, myAddress)
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
