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

package main

import (
	"bufio"
	"crypto/aes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/EGaaS/go-egaas-mvp/packages/consts"

	"github.com/EGaaS/go-egaas-mvp/packages/lib"
	logger "github.com/EGaaS/go-egaas-mvp/packages/log"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"

	"github.com/go-yaml/yaml"
)

// Settings contains options of the program
type Settings struct {
	NodeURL string // URL of EGAAS node
	Log     bool   // if true then the program writes log data
	Cookie  string //*http.Cookie
	Key     string // decrypted private key
	Address int64  // wallet id
	State   int64  // state id
}

var (
	gSettings Settings
	gPrivate  []byte // private key
	gPublic   []byte
)

func logOut(format string, params ...interface{}) {
	logger.LogDebug(consts.FuncStarted, "")
	if !gSettings.Log {
		return
	}
	logger.LogDebug(consts.DebugMessage, fmt.Sprintf(format, params...))
}

func saveSetting() {
	logger.LogDebug(consts.FuncStarted, "")
	out, err := json.Marshal(gSettings)
	if err != nil {
		logger.LogError(consts.JSONError, err)
		logOut(`saveSetting`, err)
	}
	ioutil.WriteFile(`settings.json`, out, 0600)
}

func checkKey() bool {
	logger.LogDebug(consts.FuncStarted, "")
	var privKey, pass []byte
	var err error
	// Reads the hex private key from the file
	for len(gSettings.Key) == 0 {
		var (
			filename string
		)
		fmt.Println(`Enter the filename with the private key:`)
		n, err := fmt.Scanln(&filename)
		if err != nil || n == 0 {
			logger.LogError(consts.InputError, err)
			fmt.Println(err)
			continue
		}
		if privKey, err = ioutil.ReadFile(filename); err != nil {
			logger.LogError(consts.IOError, err)
			fmt.Println(err)
			continue
		}
		privKey, err = hex.DecodeString(strings.TrimSpace(string(privKey)))
		if err != nil {
			logger.LogError(consts.CryptoError, err)
			fmt.Println(err)
			continue
		}
		if len(privKey) != 32 {
			logger.LogError(consts.CryptoError, fmt.Sprintf(`Wrong the length of private key: %d`, len(privKey)))
			fmt.Println(`Wrong the length of private key`, len(privKey))
			continue
		}
		fmt.Println(`Enter a new password:`)
		n, err = fmt.Scanln(&pass)
		if err != nil || n == 0 {
			logger.LogError(consts.InputError, err)
			fmt.Println(err)
			continue
		}
		gSettings.Address = lib.Address(lib.PrivateToPublic(privKey))
		hash := sha256.Sum256(pass)
		privKey, err = lib.CBCEncrypt(hash[:], privKey, make([]byte, aes.BlockSize))
		if err != nil {
			fmt.Println(err)
			continue
		}
		gSettings.Key = hex.EncodeToString(privKey[aes.BlockSize:])
	}
	if privKey, err = hex.DecodeString(gSettings.Key); err != nil {
		fmt.Println(err)
		return false
	}
	for {
		if len(pass) == 0 {
			fmt.Println(`Enter the password:`)
			n, err := fmt.Scanln(&pass)
			if err != nil || n == 0 {
				fmt.Println(err)
				continue
			}
		}
		hash := sha256.Sum256(pass)
		pass = pass[:0]
		gPrivate, err = lib.CBCDecrypt(hash[:], privKey, make([]byte, aes.BlockSize))
		if err != nil {
			fmt.Println(err)
			continue
		}
		gPublic = lib.PrivateToPublic(gPrivate)
		if gSettings.Address != lib.Address(gPublic) {
			fmt.Println(`Wrong password`)
			continue
		}
		break
	}
	return true
}

func login() error {
	ret, err := sendGet(`ajax_get_uid`, nil)
	if err != nil {
		return err
	}
	if len(ret[`uid`].(string)) == 0 {
		return fmt.Errorf(`Unknown uid`)
	}
	sign, err := lib.SignECDSA(hex.EncodeToString(gPrivate), ret[`uid`].(string))
	if err != nil {
		return err
	}
	var state string
	fmt.Println(`Enter a state id:`)
	_, err = fmt.Scanln(&state)
	if err != nil {
		return err
	}
	form := url.Values{"key": {hex.EncodeToString(gPublic)}, "sign": {hex.EncodeToString(sign)},
		`state_id`: {utils.Int64ToStr(utils.StrToInt64(state))}, `citizen_id`: {utils.Int64ToStr(gSettings.Address)}}

	ret, err = sendPost(`ajax_sign_in`, &form)
	if err != nil {
		return err
	}
	if ret[`result`].(bool) != true {
		return fmt.Errorf(`Login is incorrect`)
	}
	fmt.Println(`Address: `, ret[`address`])
	saveSetting()
	return nil
}

func map2yaml(in map[string]string, filename string) error {
	data, err := yaml.Marshal(in)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, data, 0600)
}

func yaml2map(filename string, out *map[string]string) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, out)
}

func main() {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(`Dir`, err)
	}
	params, err := ioutil.ReadFile(filepath.Join(dir, `settings.json`))
	if err != nil {
		log.Fatalln(dir, `Settings.json`, err)
	}
	if err = json.Unmarshal(params, &gSettings); err != nil {
		log.Fatalln(`Unmarshall`, err)
	}
	if gSettings.Log {
		logfile, err := os.OpenFile(filepath.Join(dir, "egcons.log"),
			os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalln(`Egcons log`, err)
		}
		defer logfile.Close()
		log.SetOutput(logfile)
	}
	os.Chdir(dir)
	if !checkKey() {
		return
	}
	if err = login(); err != nil {
		fmt.Println(`ERROR:`, err)
		return
	}

cmd:
	for {
		var cmd string
		var pars []string

		fmt.Printf(`>`)
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		cmd = scanner.Text()
		for _, val := range strings.Split(cmd, ` `) {
			val = strings.TrimSpace(val)
			if len(val) > 0 {
				pars = append(pars, val)
			}
		}
		if len(pars) == 0 {
			continue
		}
		switch {
		case pars[0] == `cntfields`:
			if len(pars) != 3 {
				fmt.Println(`cntfields <ContractName> <Filename>`)
			} else {
				ret, err := sendPost(`ajax_contract_info`, &url.Values{`name`: {pars[1]}})
				if err != nil {
					fmt.Println(`ERROR`, err)
				} else {
					out := make([]string, 0)
					out = append(out, fmt.Sprintf(`TxName: %s`, ret[`name`]))
					for _, field := range ret[`fields`].([]interface{}) {
						tmp := field.(map[string]interface{})
						//	out = append(out, fmt.Sprintf(`%s: #%s %s`, field[`name`], field[`type`], field[`tagsb `]))
						out = append(out, fmt.Sprintf(`%v: #%v %v`, tmp[`name`], tmp[`type`], tmp[`tags`]))
					}
					err = ioutil.WriteFile(pars[2], []byte(strings.Join(out, "\r\n")), 0600)
					if err != nil {
						fmt.Println(err)
					}
				}
			}
		case pars[0] == `quit`:
			break cmd
		default:
			fmt.Println(`Unknown command`)
		}
	}
}
