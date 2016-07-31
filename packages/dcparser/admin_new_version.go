package dcparser

import (
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/utils"
	//"log"
	//"encoding/json"
	//"regexp"
	//"math"
	//"strings"
	//	"os"
	//	"time"
	//"strings"
	//"bytes"
	//"github.com/DayLightProject/go-daylight/packages/consts"
	//	"math"
	//	"database/sql"
	//	"bytes"
	"io/ioutil"
)

func (p *Parser) AdminNewVersionInit() error {
	/*
		soft_type тип софта, например php/cppwin/cppnix
		version версия, например 0.0.10
		file запакованный файл
		format чем запакован файл или же просто exe
	*/
	fields := []map[string]string{{"soft_type": "string"}, {"version": "string"}, {"file": "bytes"}, {"format": "string"}, {"sign": "bytes"}}
	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) AdminNewVersionFront() error {

	err := p.generalCheckAdmin()
	if err != nil {
		return p.ErrInfo(err)
	}

	verifyData := map[string]string{"version": "version", "soft_type": "soft_type"}
	err = p.CheckInputData(verifyData)
	if err != nil {
		return p.ErrInfo(err)
	}

	version, err := p.Single("SELECT version FROM new_version WHERE version  =  ?", p.TxMap["version"]).String()
	if err != nil {
		return p.ErrInfo(err)
	}
	if len(version) > 0 {
		return p.ErrInfo("exists version")
	}

	forSign := fmt.Sprintf("%s,%s,%s,%s,%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxMap["user_id"], p.TxMap["soft_type"], p.TxMap["version"], utils.Sha256(p.TxMap["file"]), p.TxMap["format"])
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, forSign, p.TxMap["sign"], false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}
	return nil
}

func (p *Parser) AdminNewVersion() error {
	err := ioutil.WriteFile(*utils.Dir+"/public/"+p.TxMaps.String["version"]+"."+p.TxMaps.String["format"], p.TxMaps.Bytes["file"], 0644)
	if err != nil {
		return p.ErrInfo(err)
	}
	err = p.ExecSql("INSERT INTO new_version ( version ) VALUES ( ? )", p.TxMaps.String["version"])
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) AdminNewVersionRollback() error {
	err := p.ExecSql("DELETE FROM new_version WHERE version = ?", p.TxMaps.String["version"])
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) AdminNewVersionRollbackFront() error {
	return nil
}
