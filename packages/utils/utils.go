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

package utils

import (
	"archive/zip"
	"bytes"
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	//	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"math"
	"math/rand"
	"net"
	"net/http"

	"github.com/EGaaS/go-egaas-mvp/packages/consts"
	"github.com/EGaaS/go-egaas-mvp/packages/lib"
	"github.com/EGaaS/go-egaas-mvp/packages/static"
	"github.com/EGaaS/go-egaas-mvp/packages/textproc"
	"github.com/kardianos/osext"
	//	_ "github.com/lib/pq"
	"github.com/mcuadros/go-version"
	"github.com/shopspring/decimal"
	//	"net/mail"
	//  "net/smtp"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
)

// BlockData is a structure of the block's header
type BlockData struct {
	BlockId  int64
	Time     int64
	WalletId int64
	StateID  int64
	Sign     []byte
	Hash     []byte
}

// DaemonsChansType is a structure for deamons
type DaemonsChansType struct {
	ChBreaker chan bool
	ChAnswer  chan string
}

var (
	// FirstBlockDir is a folder where 1block file will be stored
	FirstBlockDir = flag.String("firstBlockDir", "", "FirstBlockDir")
	// FirstBlockPublicKey is the private key
	FirstBlockPublicKey = flag.String("firstBlockPublicKey", "", "FirstBlockPublicKey")
	// FirstBlockNodePublicKey is the node private key
	FirstBlockNodePublicKey = flag.String("firstBlockNodePublicKey", "", "FirstBlockNodePublicKey")
	// FirstBlockHost is the host of the first block
	FirstBlockHost = flag.String("firstBlockHost", "", "FirstBlockHost")
	// WalletAddress is a wallet address for forging
	WalletAddress = flag.String("walletAddress", "", "walletAddress for forging ")
	// TCPHost is the tcp host
	TCPHost = flag.String("tcpHost", "", "tcpHost (e.g. 127.0.0.1)")
	// ListenHTTPPort is HTTP port
	ListenHTTPPort = flag.String("listenHttpPort", "7079", "ListenHTTPPort")
	// GenerateFirstBlock show if the first block must be generated
	GenerateFirstBlock = flag.Int64("generateFirstBlock", 0, "generateFirstBlock")
	// OldVersion is the number of the old version
	OldVersion = flag.String("oldVersion", "", "")
	// TestRollBack equals 1 for testing rollback
	TestRollBack = flag.Int64("testRollBack", 0, "testRollBack")
	// Dir is EGAAS folder
	Dir = flag.String("dir", GetCurrentDir(), "DayLight directory")
	// OldFileName is the old file name
	OldFileName = flag.String("oldFileName", "", "")
	// LogLevel is the log level
	LogLevel = flag.String("logLevel", "", "DayLight LogLevel")
	// Console equals 1 for starting in console
	Console = flag.Int64("console", 0, "Start from console")
	// StartBlockID is the start block
	StartBlockID = flag.Int64("startBlockId", 0, "Start block for blockCollection daemon")
	// EndBlockID is the end block
	EndBlockID = flag.Int64("endBlockId", 0, "End block for blockCollection daemon")
	// RollbackToBlockID is the target block for rollback
	RollbackToBlockID = flag.Int64("rollbackToBlockId", 0, "Rollback to block_id")
	// TLS is a directory for .well-known and keys. It is required for https
	TLS = flag.String("tls", "", "Support https. Specify directory for .well-known")
	// DevTools switches on dev tools in thrust shell
	DevTools = flag.Int64("devtools", 0, "Devtools in thrust-shell")
	// BoltDir is the edir for BoltDb folder
	BoltDir = flag.String("boltDir", GetCurrentDir(), "Bolt directory")
	// BoltPsw is the password for BoltDB
	BoltPsw = flag.String("boltPsw", "", "Bolt password")
	// APIToken is an api token for exchange api
	APIToken = flag.String("apiToken", "", "API Token")
	// OneCountry is the country which is supported
	OneCountry int64
	// PrivCountry is protect system from registering
	PrivCountry bool
	//	OutFile            *os.File

	// LogoExt is the extension of the logotype
	LogoExt = `png`
	// DltWalletID is the wallet identifier
	DltWalletID = flag.Int64("dltWalletId", 0, "DltWalletID")

	// DaemonsChans is a slice of DaemonsChansType
	DaemonsChans []*DaemonsChansType
	// Thrust is true for thrust shell
	Thrust bool
)

func init() {
	flag.Parse()
}

// IOS checks if the app runs on iOS
func IOS() bool {
	if (runtime.GOARCH == "arm" || runtime.GOARCH == "arm64") && runtime.GOOS == "darwin" {
		return true
	}
	return false
}

// Desktop checks if the app runs on the desktop with thrust_shell
func Desktop() bool {
	thrustShell := "thrust_shell"
	if runtime.GOOS == "windows" {
		thrustShell = "thrust_shell.exe"
	} else if runtime.GOOS == "darwin" {
		thrustShell = "ThrustShell"
	}
	if _, err := os.Stat(*Dir + "/" + thrustShell); err == nil {
		return true
	}
	return false
}

// Mobile checks if the app runs on Android or iOS
func Mobile() bool {
	if IOS() || runtime.GOOS == "android" {
		return true
	}
	return false
}

// Android checks if the app runs on Android
func Android() bool {
	if runtime.GOOS == "android" {
		return true
	}
	return false
}

// Sleep makes a pause during sec seconds
func Sleep(sec time.Duration) {
	time.Sleep(sec * time.Second)
}

// ParseBlockHeader parses the header of the block
func ParseBlockHeader(binaryBlock *[]byte) *BlockData {
	result := new(BlockData)
	// распарсим заголовок блока // parse the heading of a block
	/*
		Заголовок // the heading
		TYPE (0-блок, 1-тр-я)        1 // TYPE(0-block, 1-transaction)
		BLOCK_ID   				       4
		TIME       					       4
		WALLET_ID                         1-8
		state_id                              1
		SIGN                               от 128 до 512 байт. Подпись от TYPE, BLOCK_ID, PREV_BLOCK_HASH, TIME, WALLET_ID, state_id, MRKL_ROOT // from 128 to 512 байт. Signature from TYPE, BLOCK_ID, PREV_BLOCK_HASH, TIME, WALLET_ID, state_id, MRKL_ROOT
Далее - тело блока (Тр-ии) // further is body block (transaction)
	*/
	result.BlockId = BinToDecBytesShift(binaryBlock, 4)
	result.Time = BinToDecBytesShift(binaryBlock, 4)
	result.WalletId, _ = lib.DecodeLenInt64(binaryBlock) //BytesToInt64(BytesShift(binaryBlock, DecodeLength(binaryBlock)))
	// Delete after re-build blocks
	/*	if result.WalletId == 0x31 {
		result.WalletId = 1
	}*/
	result.StateID = BinToDecBytesShift(binaryBlock, 1)
	if result.BlockId > 1 {
		signSize := DecodeLength(binaryBlock)
		result.Sign = BytesShift(binaryBlock, signSize)
	} else {
		*binaryBlock = (*binaryBlock)[1:]
	}
	log.Debug("result.BlockId: %v / result.Time: %v / result.WalletId: %v / result.StateID: %v / result.Sign: %v", result.BlockId, result.Time, result.WalletId, result.StateID, result.Sign)
	return result
}

/*
func Round(f float64, places int) (float64) {
	if places==0 {
		return math.Floor(f + .5)
	} else {
		shift := math.Pow(10, float64(places))
		return math.Floor((f * shift)+.5) / shift;
	}
}
*/

func round(num float64) int64 {
	//log.Debug("num", num)
	//num += ROUND_FIX
	//	return int(StrToFloat64(Float64ToStr(num)) + math.Copysign(0.5, num))
	//log.Debug("num", num)
	return int64(num + math.Copysign(0.5, num))
}

// Round rounds float64 value
func Round(num float64, precision int) float64 {
	num += consts.ROUND_FIX
	output := math.Pow(10, float64(precision))
	return float64(round(num*output)) / output
}

// RandInt returns a random integer between min and max
func RandInt(min int, max int) int {
	if max-min <= 0 {
		return 1
	}
	return min + rand.Intn(max-min)
}

// CheckInputData checks the input data
func CheckInputData(idata interface{}, dataType string) bool {
	var data string
	switch idata.(type) {
	case int:
		data = IntToStr(idata.(int))
	case int64:
		data = Int64ToStr(idata.(int64))
	case float64:
		data = Float64ToStr(idata.(float64))
	case string:
		data = idata.(string)
	case []byte:
		data = string(idata.([]byte))
	}
	log.Debug("CheckInputData:" + data)
	log.Debug("dataType:" + dataType)
	switch dataType {
	case "arbitration_trust_list":
		if ok, _ := regexp.MatchString(`^\[[0-9]{1,10}(,[0-9]{1,10}){0,100}\]$`, data); ok {
			return true
		}
	case "abuse_comment", "vote_comment":
		if ok, _ := regexp.MatchString(`^(?i)[0-9a-z\,\s\.\-]{1,255}$`, data); ok {
			return true
		}
	case "private_key":
		if ok, _ := regexp.MatchString(`^[0-9a-fA-F]+$`, data); ok {
			if len(data) == 64 {
				return true
			}
		}
	case "votes_comment", "cf_comment":
		if ok, _ := regexp.MatchString(`^[\pL0-9\,\s\.\-\:\=\;\?\!\%\)\(\@\/\n\r]{1,140}$`, data); ok {
			return true
		}
	case "type":
		if ok, _ := regexp.MatchString(`^[\w]+$`, data); ok {
			if StrToInt(data) <= 30 {
				return true
			}
		}
	case "word":
		if ok, _ := regexp.MatchString(`^(?i)[a-z]+$`, data); ok {
			if StrToInt(data) <= 1024 {
				return true
			}
		}
	case "currency_name", "state_name":
		if ok, _ := regexp.MatchString(`^[\pL0-9\,\s\.\-\:\=\;\?\!\%\)\(\@\/\n\r]{1,20}$`, data); ok {
			if StrToInt(data) <= 1024 {
				return true
			}
		}
	case "string":
		if ok, _ := regexp.MatchString(`^[\w]+$`, data); ok {
			if StrToInt(data) <= 1024 {
				return true
			}
		}
	case "referral":
		if ok, _ := regexp.MatchString(`^[0-9]{1,2}$`, data); ok {
			if StrToInt(data) <= 30 {
				return true
			}
		}
	case "currency_id":
		if ok, _ := regexp.MatchString(`^[0-9]{1,3}$`, data); ok {
			if StrToInt(data) <= 255 {
				return true
			}
		}
	case "system_commission":
		if ok, _ := regexp.MatchString(`^[0-9]{1,3}$`, data); ok {
			if StrToInt(data) <= 15 && StrToInt(data) >= 5 {
				return true
			}
		}
	case "tinyint":
		if ok, _ := regexp.MatchString(`^[0-9]{1,3}$`, data); ok {
			if StrToInt(data) <= 127 {
				return true
			}
		}
	case "smallint":
		if ok, _ := regexp.MatchString(`^[0-9]{1,5}$`, data); ok {
			if StrToInt(data) <= 65535 {
				return true
			}
		}
	case "column_type":
		if ok, _ := regexp.MatchString(`^(text|int64|time|hash|money|double)$`, data); ok {
			return true
		}
	case "avatar":
		regex := `https?\:\/\/`        // SCHEME
		regex += `[\w-.]*\.[a-z]{2,4}` // Host or IP
		regex += `(\:[0-9]{2,5})?`     // Port
		regex += `(\/[\w_-]+)*\/?`     // Path
		regex += `\.(png|jpg)`         // Img
		if ok, _ := regexp.MatchString(`^`+regex+`$`, data); ok {
			if len(data) < 100 {
				return true
			}
		}
	case "img_url":
		regex := `https?\:\/\/`        // SCHEME
		regex += `[\w-.]*\.[a-z]{2,4}` // Host or IP
		regex += `(\:[0-9]{2,5})?`     // Port
		regex += `(\/[\w_-]+)*\/?`     // Path
		regex += `\.(png|jpg)`         // Img
		if ok, _ := regexp.MatchString(`^`+regex+`$`, data); ok {
			if len(data) < 50 {
				return true
			}
		}
	case "ca_url", "arbitrator_url":
		regex := `https?\:\/\/`        // SCHEME
		regex += `[\w-.]*\.[a-z]{2,4}` // Host or IP
		regex += `(\:[0-9]{2,5})?`     // Port
		regex += `(\/[\w_-]+)*\/?`     // Path
		if ok, _ := regexp.MatchString(`^`+regex+`$`, data); ok {
			if len(data) <= 30 {
				return true
			}
		}
	case "credit_pct", "pct":
		if ok, _ := regexp.MatchString(`^[0-9]{1,3}(\.[0-9]{2})?$`, data); ok {
			return true
		}
	case "user_name":
		if ok, _ := regexp.MatchString(`^[\w\s]{1,30}$`, data); ok {
			return true
		}
	case "admin_currency_list":
		if ok, _ := regexp.MatchString(`^((\d{1,3}\,){0,9}\d{1,3}|ALL)$`, data); ok {
			return true
		}
	case "users_ids":
		if ok, _ := regexp.MatchString(`^([0-9]{1,12},){0,1000}[0-9]{1,12}$`, data); ok {
			return true
		}
	case "version":
		if ok, _ := regexp.MatchString(`^[0-9]{1,2}\.[0-9]{1,2}\.[0-9]{1,2}([a-z]{1,2}[0-9]{1,2})?$`, data); ok {
			return true
		}
	case "soft_type":
		if ok, _ := regexp.MatchString(`^[a-z]{3,10}$`, data); ok {
			return true
		}
	case "currency_full_name":
		if ok, _ := regexp.MatchString(`^[a-zA-Z\s]{3,50}$`, data); ok {
			return true
		}
	case "currency_commission":
		if ok, _ := regexp.MatchString(`^[0-9]{1,7}(\.[0-9]{1,2})?$`, data); ok {
			return true
		}
	case "sell_rate":
		if ok, _ := regexp.MatchString(`^[0-9]{0,10}(\.[0-9]{0,10})?$`, data); ok {
			return true
		}
	case "amount":
		if ok, _ := regexp.MatchString(`^[0-9]{0,10}(\.[0-9]{0,2})?$`, data); ok {
			return true
		}
	case "amount_btc":
		if ok, _ := regexp.MatchString(`^[0-9]{0,10}(\.[0-9]{0,5})?$`, data); ok {
			return true
		}
	case "tpl_name":
		if ok, _ := regexp.MatchString("^[\\w]{1,30}$", data); ok {
			return true
		}
	case "example_spots":
		r1 := `"\d{1,2}":\["\d{1,3}","\d{1,3}",(\[("[a-z_]{1,30}",?){0,20}\]|""),"\d{1,2}","\d{1,2}"\]`
		reg := `^\{(\"(face|profile)\":\{(` + r1 + `,?){1,20}\},?){2}}$`
		if ok, _ := regexp.MatchString(reg, data); ok {
			return true
		}
	case "segments":
		r1 := `"\d{1,2}":\["\d{1,2}","\d{1,2}"\]`
		face := `"face":\{(` + r1 + `\,){1,20}` + r1 + `\}`
		profile := `"profile":\{(` + r1 + `\,){1,20}` + r1 + `\}`
		reg := `^\{` + face + `,` + profile + `\}$`
		if ok, _ := regexp.MatchString(reg, data); ok {
			return true
		}
	case "tolerances":
		r1 := `"\d{1,2}":"0\.\d{1,2}"`
		face := `"face":\{(` + r1 + `\,){1,50}` + r1 + `\}`
		profile := `"profile":\{(` + r1 + `\,){1,50}` + r1 + `\}`
		reg := `^\{` + face + `,` + profile + `\}$`
		if ok, _ := regexp.MatchString(reg, data); ok {
			return true
		}
	case "compatibility":
		if ok, _ := regexp.MatchString(`^\[(\d{1,5},)*\d{1,5}\]$`, data); ok {
			return true
		}
	case "race":
		if ok, _ := regexp.MatchString("^[1-3]$", data); ok {
			return true
		}
	case "country":
		if ok, _ := regexp.MatchString("^[0-9]{1,3}$", data); ok {
			return true
		}
	case "vote", "boolean":
		if ok, _ := regexp.MatchString(`^0|1$`, data); ok {
			return true
		}
	case "coordinate":
		if ok, _ := regexp.MatchString(`^\-?[0-9]{1,3}(\.[0-9]{1,5})?$`, data); ok {
			return true
		}
	case "cf_links":
		regex := `\["https?\:\/\/(goo\.gl|bit\.ly|t\.co)\/[\w-]+",[0-9]+,[0-9]+,[0-9]+,[0-9]+\]`
		if ok, _ := regexp.MatchString(`^\[`+regex+`(\,`+regex+`)*\]$`, data); ok {
			if len(data) < 512 {
				return true
			}
		}
	case "http_host":
		if ok, _ := regexp.MatchString(`^https?:\/\/[0-9a-z\_\.\-\/:]{1,100}[\/]$`, data); ok {
			return true
		}
	case "e_host":
		if ok, _ := regexp.MatchString(`^https?:\/\/[0-9a-z\_\.\-\/:]{1,100}[\/]$`, data); ok || data == "0" {
			return true
		}
	case "host":
		if ok, _ := regexp.MatchString(`^(?i)[0-9a-z\_\.\-]{1,100}$`, data); ok {
			return true
		}
	case "tcp_host":
		if ok, _ := regexp.MatchString(`^(?i)[0-9a-z\_\.\-]{1,100}:[0-9]+$`, data); ok {
			return true
		}
	case "coords":
		xy := `\[\d{1,3}\,\d{1,3}\]`
		r := `^\[(` + xy + `\,){}` + xy + `\]$`
		if ok, _ := regexp.MatchString(r, data); ok {
			return true
		}
	case "lang":
		if ok, _ := regexp.MatchString("^(en|ru)$", data); ok {
			return true
		}
	case "payment_systems_ids":
		if ok, _ := regexp.MatchString("^([0-9]{1,4},){0,4}[0-9]{1,4}$", data); ok {
			return true
		}
	case "video_type":
		if ok, _ := regexp.MatchString("^(youtube|vimeo|youku|null)$", data); ok {
			return true
		}
	case "video_url_id", "sn_url_id":
		if ok, _ := regexp.MatchString("^(?i)(null|[0-9a-z_\\-\\.]{2,32})$", data); ok {
			return true
		}
	case "sn_type":
		if ok, _ := regexp.MatchString("^(vk|fb|qq)$", data); ok {
			return true
		}
	case "sha1":
		if ok, _ := regexp.MatchString("^[0-9a-z]{40}$", data); ok {
			return true
		}

	case "walletAddress":
		if ok, _ := regexp.MatchString("^(?i)[0-9]{20}$", strings.Replace(data, `-`, ``, -1)); ok {
			return true
		}
	case "photo_hash", "sha256":
		if ok, _ := regexp.MatchString("^[0-9a-z]{64}$", data); ok {
			return true
		}
	case "cash_code":
		if ok, _ := regexp.MatchString("^[0-9a-z]{32}$", data); ok {
			return true
		}
	case "alert":
		if ok, _ := regexp.MatchString("^[\\pL0-9\\,\\s\\.\\-\\:\\=\\;\\?\\!\\%\\)\\(\\@\\/]{1,512}$", data); ok {
			return true
		}
	case "int":
		if ok, _ := regexp.MatchString("^[0-9]{1,10}$", data); ok {
			return true
		}
	case "float":
		if ok, _ := regexp.MatchString(`^[0-9]{1,5}(\.[0-9]{1,5})?$`, data); ok {
			return true
		}
	case "sleep_var":
		if ok, _ := regexp.MatchString(`^\{\"is_ready\"\:\[([0-9]{1,5},){1,100}[0-9]{1,5}\],\"generator\"\:\[([0-9]{1,5},){1,100}[0-9]{1,5}\]\}$`, data); ok {
			return true
		}
	case "int64", "bigint", "user_id":
		if ok, _ := regexp.MatchString("^-?[0-9]{1,20}$", data); ok {
			return true
		}
	case "decimal": // 1.2345678e+25
		if ok, _ := regexp.MatchString(`^([0-9]{1,30})|([0-9]+\.[0-9]+[e]\+\[0-9]+)$`, data); ok {
			return true
		}
	case "level":
		if StrToInt(data) >= 0 && StrToInt(data) <= 34 {
			return true
		}
	case "comment":
		if len(data) >= 1 && len(data) <= 512 {
			return true
		}
	case "conditions":
		if len(data) <= 1024 {
			return true
		}
	case "hex_sign", "hex", "public_key":
		if ok, _ := regexp.MatchString("^[0-9a-z]+$", data); ok {
			if len(data) < 2048 {
				return true
			}
		}
	case "account":
		if ok, _ := regexp.MatchString(`^[0-9a-zA-Z\-\s_\+\#\:]{1,50}$`, data); ok {
			return true
		}
	case "method":
		if ok, _ := regexp.MatchString(`^[0-9a-zA-Z\-\_]{1,30}$`, data); ok {
			return true
		}
	}

	return false
}

// Time returns th ecurrent Unix time
func Time() int64 {
	return time.Now().Unix()
}

// ValidateEmail validates email
func ValidateEmail(email string) bool {
	Re := regexp.MustCompile(`^(?i)[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	return Re.MatchString(email)
}

// GetHTTPTextAnswer returns HTTP answer as a string
func GetHTTPTextAnswer(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	htmlData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode == 404 {
		err = fmt.Errorf(`404`)
	}
	return string(htmlData), err
}

// StrToInt64 converts string to int64
func StrToInt64(s string) int64 {
	int64, _ := strconv.ParseInt(s, 10, 64)
	return int64
}

// BytesToInt64 converts []bytes to int64
func BytesToInt64(s []byte) int64 {
	int64, _ := strconv.ParseInt(string(s), 10, 64)
	return int64
}

// StrToUint64 converts string to the unsinged int64
func StrToUint64(s string) uint64 {
	ret, _ := strconv.ParseUint(s, 10, 64)
	return ret
}

// StrToInt converts string to integer
func StrToInt(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}

// Float64ToStr converts float64 to string
func Float64ToStr(f float64) string {
	return strconv.FormatFloat(f, 'f', 13, 64)
}

// StrToFloat64 converts string to float64
func StrToFloat64(s string) float64 {
	Float64, _ := strconv.ParseFloat(s, 64)
	return Float64
}

// BytesToFloat64 converts []byte to float64
func BytesToFloat64(s []byte) float64 {
	Float64, _ := strconv.ParseFloat(string(s), 64)
	return Float64
}

// BytesToInt converts []byte to integer
func BytesToInt(s []byte) int {
	i, _ := strconv.Atoi(string(s))
	return i
}

// StrToMoney rounds money string to float64
func StrToMoney(str string) float64 {
	ind := strings.Index(str, ".")
	new := ""
	if ind != -1 {
		end := 2
		if len(str[ind+1:]) > 1 {
			end = 3
		}
		new = str[:ind] + "." + str[ind+1:ind+end]
	} else {
		new = str
	}
	return StrToFloat64(new)
}

// GetEndBlockID returns the end block id
func GetEndBlockID() (int64, error) {

	if _, err := os.Stat(*Dir + "/public/blockchain"); os.IsNotExist(err) {
		return 0, nil
	}

	// размер блока, записанный в 5-и последних байтах файла blockchain
	// size of a block recorded into the last 5 bytes of blockchain file
	fname := *Dir + "/public/blockchain"
	file, err := os.Open(fname)
	if err != nil {
		return 0, ErrInfo(err)
	}
	defer file.Close()

	fi, err := file.Stat()
	if err != nil {
		log.Fatal(err)
	}
	if fi.Size() == 0 {
		return 0, ErrInfo("/public/blockchain size=0")
	}

	// размер блока, записанный в 5-и последних байтах файла blockchain
	// size of a block recorded into the last 5 bytes of blockchain file
	_, err = file.Seek(-5, 2)
	if err != nil {
		return 0, ErrInfo(err)
	}
	buf := make([]byte, 5)
	_, err = file.Read(buf)
	if err != nil {
		return 0, ErrInfo(err)
	}
	size := BinToDec(buf)
	if size > consts.MAX_BLOCK_SIZE {
		return 0, ErrInfo("size > conts.MAX_BLOCK_SIZE")
	}
	// сам блок
	// block itself
	_, err = file.Seek(-(size + 5), 2)
	if err != nil {
		return 0, ErrInfo(err)
	}
	dataBinary := make([]byte, size+5)
	_, err = file.Read(dataBinary)
	if err != nil {
		return 0, ErrInfo(err)
	}
	// размер (id блока + тело блока)
	// size (block id + body of a block)
	BinToDecBytesShift(&dataBinary, 5)
	return BinToDecBytesShift(&dataBinary, 5), nil
}

// DownloadToFile downloads and saves the specified file
func DownloadToFile(url, file string, timeoutSec int64, DaemonCh chan bool, AnswerDaemonCh chan string, GoroutineName string) (int64, error) {

	f, err := os.Create(file)
	if err != nil {
		return 0, ErrInfo(err)
	}
	defer f.Close()

	timeout := time.Duration(time.Duration(timeoutSec) * time.Second)
	client := http.Client{
		Timeout: timeout,
	}
	resp, err := client.Get(url)
	if err != nil {
		return 0, ErrInfo(err)
	}
	defer resp.Body.Close()

	var offset int64
	for {
		if DaemonCh != nil {
			select {
			case <-DaemonCh:
				if GoroutineName == "NodeVoting" {
					DB.DbUnlock(GoroutineName)
				}
				AnswerDaemonCh <- GoroutineName
				return offset, fmt.Errorf("daemons restart")
			default:
			}
		}
		data, err := ioutil.ReadAll(io.LimitReader(resp.Body, 10000))
		if err != nil {
			return offset, ErrInfo(err)
		}
		f.WriteAt(data, offset)
		offset += int64(len(data))
		if len(data) == 0 {
			break
		}
		log.Debug("read %s", url)
	}
	return offset, nil
}

// ErrInfoFmt fomats the error message
func ErrInfoFmt(err string, a ...interface{}) error {
	return fmt.Errorf("%s (%s)", fmt.Sprintf(err, a...), Caller(1))
}

// ErrInfo formats the error message
func ErrInfo(verr interface{}, additionally ...string) error {
	var err error
	switch verr.(type) {
	case error:
		err = verr.(error)
	case string:
		err = errors.New(verr.(string))
	}
	if err != nil {
		if len(additionally) > 0 {
			return fmt.Errorf("%s # %s (%s)", err, additionally, Caller(1))
		}
		return fmt.Errorf("%s (%s)", err, Caller(1))
	}
	return err
}

// CallMethod calls the function by its name
func CallMethod(i interface{}, methodName string) interface{} {
	var ptr reflect.Value
	var value reflect.Value
	var finalMethod reflect.Value

	value = reflect.ValueOf(i)

	// if we start with a pointer, we need to get value pointed to
	// if we start with a value, we need to get a pointer to that value
	if value.Type().Kind() == reflect.Ptr {
		ptr = value
		value = ptr.Elem()
	} else {
		ptr = reflect.New(reflect.TypeOf(i))
		temp := ptr.Elem()
		temp.Set(value)
	}

	// check for method on value
	method := value.MethodByName(methodName)
	if method.IsValid() {
		finalMethod = method
	}
	// check for method on pointer
	method = ptr.MethodByName(methodName)
	if method.IsValid() {
		finalMethod = method
	}

	if finalMethod.IsValid() {
		return finalMethod.Call([]reflect.Value{})[0].Interface()
	}

	// return or panic, method not found of either type
	return fmt.Errorf("method %s not found", methodName)
}

// Caller returns the name of the latest function
func Caller(steps int) string {
	name := "?"
	if pc, _, num, ok := runtime.Caller(steps + 1); ok {
		//fmt.Println(num)
		name = fmt.Sprintf("%s :  %d", filepath.Base(runtime.FuncForPC(pc).Name()), num)
	}
	return name
}

// InSliceString searches the string in the slice of strings
func InSliceString(search string, slice []string) bool {
	for _, v := range slice {
		if v == search {
			return true
		}
	}
	return false
}

// EncodeLengthPlusData encoding interface into []byte
func EncodeLengthPlusData(idata interface{}) []byte {
	var data []byte
	switch idata.(type) {
	case int64:
		data = Int64ToByte(idata.(int64))
	case string:
		data = []byte(idata.(string))
	case []byte:
		data = idata.([]byte)
	}
	//log.Debug("data: %x", data)
	//log.Debug("len data: %d", len(data))
	return append(lib.EncodeLength(int64(len(data))), data...)
}

// UInt32ToStr converts uint32 to string
func UInt32ToStr(num uint32) string {
	return strconv.FormatInt(int64(num), 10)
}

// Int64ToStr converts int64 to string
func Int64ToStr(num int64) string {
	return strconv.FormatInt(num, 10)
}

// Int64ToByte converts int64 to []byte
func Int64ToByte(num int64) []byte {
	return []byte(strconv.FormatInt(num, 10))
}

// IntToStr converts integer to string
func IntToStr(num int) string {
	return strconv.Itoa(num)
}

// DecToBin converts interface to []byte
func DecToBin(v interface{}, sizeBytes int64) []byte {
	var dec int64
	switch v.(type) {
	case int:
		dec = int64(v.(int))
	case int64:
		dec = v.(int64)
	case string:
		dec = StrToInt64(v.(string))
	}
	Hex := fmt.Sprintf("%0"+Int64ToStr(sizeBytes*2)+"x", dec)
	return HexToBin([]byte(Hex))
}

// BinToHex converts interface to hex []byte
func BinToHex(v interface{}) []byte {
	var bin []byte
	switch v.(type) {
	case []byte:
		bin = v.([]byte)
	case int64:
		bin = Int64ToByte(v.(int64))
	case string:
		bin = []byte(v.(string))
	}
	return []byte(fmt.Sprintf("%x", bin))
}

// HexToBin converts hex interface to binary []byte
func HexToBin(ihexdata interface{}) []byte {
	var hexdata string
	switch ihexdata.(type) {
	case []byte:
		hexdata = string(ihexdata.([]byte))
	case int64:
		hexdata = Int64ToStr(ihexdata.(int64))
	case string:
		hexdata = ihexdata.(string)
	}
	var str []byte
	str, err := hex.DecodeString(hexdata)
	if err != nil {
		log.Error("%v / %v", err, GetParent())
	}
	return str
}

// BinToDec converts input binary []byte to int64
func BinToDec(bin []byte) int64 {
	var a uint64
	l := len(bin)
	for i, b := range bin {
		shift := uint64((l - i - 1) * 8)
		a |= uint64(b) << shift
	}
	return int64(a)
}

// BinToDecBytesShift converts the input binary []byte to int64 and shifts the input bin
func BinToDecBytesShift(bin *[]byte, num int64) int64 {
	return BinToDec(BytesShift(bin, num))
}

// BytesShift returns the index bytes of the input []byte and shift str pointer
func BytesShift(str *[]byte, index int64) (ret []byte) {
	if int64(len(*str)) < index || index == 0 {
		*str = (*str)[:0]
		return []byte{}
	}
	ret, *str = (*str)[:index], (*str)[index:]
	return
}

// InterfaceToStr converts the interfaces to the string
func InterfaceToStr(v interface{}) string {
	var str string
	switch v.(type) {
	case int:
		str = IntToStr(v.(int))
	case float64:
		str = Float64ToStr(v.(float64))
	case int64:
		str = Int64ToStr(v.(int64))
	case string:
		str = v.(string)
	case []byte:
		str = string(v.([]byte))
	default:
		if reflect.TypeOf(v).String() == `decimal.Decimal` {
			str = v.(decimal.Decimal).String()
		}
	}
	return str
}

// InterfaceSliceToStr converts the slice of interfaces to the slice of strings
func InterfaceSliceToStr(i []interface{}) []string {
	var str []string
	for _, v := range i {
		str = append(str, InterfaceToStr(v))
	}
	return str
}

// InterfaceToFloat64 converts the interfaces to the float64
func InterfaceToFloat64(i interface{}) float64 {
	var result float64
	switch i.(type) {
	case int:
		result = float64(i.(int))
	case float64:
		result = i.(float64)
	case int64:
		result = float64(i.(int64))
	case string:
		result = StrToFloat64(i.(string))
	case []byte:
		result = BytesToFloat64(i.([]byte))
	}
	return result
}

// BytesShiftReverse gets []byte from the end of the input and cut the input pointer to []byte
func BytesShiftReverse(str *[]byte, v interface{}) []byte {
	var index int64
	switch v.(type) {
	case int:
		index = int64(v.(int))
	case int64:
		index = v.(int64)
	}

	var substr []byte
	slen := int64(len(*str))
	if slen < index {
		index = slen
	}
	substr = (*str)[slen-index:]
	*str = (*str)[:slen-index]
	return substr
}

// CopyFileContents copy files
func CopyFileContents(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return ErrInfo(err)
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return ErrInfo(err)
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()
	if _, err = io.Copy(out, in); err != nil {
		return ErrInfo(err)
	}
	err = out.Sync()
	return ErrInfo(err)
}

// RandSeq generates a random string
func RandSeq(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	rand.Seed(time.Now().UnixNano())
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// CheckSign checks the signature
func CheckSign(publicKeys [][]byte, forSign string, signs []byte, nodeKeyOrLogin bool) (bool, error) {
	defer func() {
		if r := recover(); r != nil {
			log.Error("Panic CheckECDSA %v", r)
		}
	}()

	var signsSlice [][]byte
	if len(forSign) == 0 {
		return false, ErrInfoFmt("len(forSign) == 0")
	}
	if len(publicKeys) == 0 {
		return false, ErrInfoFmt("len(publicKeys) == 0")
	}
	if len(signs) == 0 {
		return false, ErrInfoFmt("len(signs) == 0")
	}
	// у нода всегда 1 подпись
	// node always has olny one signature
	if nodeKeyOrLogin {
		signsSlice = append(signsSlice, signs)
	} else {
		if length := DecodeLength(&signs); length > 0 {
			signsSlice = append(signsSlice, BytesShift(&signs, length))
		}
		if len(publicKeys) != len(signsSlice) {
			return false, fmt.Errorf("sign error %d!=%d", len(publicKeys), len(signsSlice))
		}
	}
	return lib.CheckECDSA(publicKeys[0], forSign, signsSlice[0])
}

// Md5 returns the hex MD5 hash
func Md5(v interface{}) []byte {
	var msg []byte
	switch v.(type) {
	case string:
		msg = []byte(v.(string))
	case []byte:
		msg = v.([]byte)
	}
	sh := crypto.MD5.New()
	sh.Write(msg)
	hash := sh.Sum(nil)
	return BinToHex(hash)
}

// DSha256 returns the double calculation of SHA256 hash
func DSha256(v interface{}) []byte {
	var data []byte
	switch v.(type) {
	case string:
		data = []byte(v.(string))
	case []byte:
		data = v.([]byte)
	}
	isha256 := sha256.New()
	isha256.Write(data)
	hashSha256 := fmt.Sprintf("%x", isha256.Sum(nil))
	isha256 = sha256.New()
	isha256.Write([]byte(hashSha256))
	return []byte(fmt.Sprintf("%x", isha256.Sum(nil)))
}

// Sha256 returns SHA256 hash
func Sha256(v interface{}) []byte {
	var data []byte
	switch v.(type) {
	case string:
		data = []byte(v.(string))
	case []byte:
		data = v.([]byte)
	}
	isha256 := sha256.New()
	isha256.Write(data)
	return []byte(fmt.Sprintf("%x", isha256.Sum(nil)))
}

// GetMrklroot returns MerkleTreeRoot
func GetMrklroot(binaryData []byte, first bool) ([]byte, error) {
	var mrklSlice [][]byte
	var txSize int64
	// [error] парсим после вызова функции
	// parse [error] after the calling of a function
	if len(binaryData) > 0 {
		for {
			// чтобы исключить атаку на переполнение памяти
			// to exclude an attack on memory overflow
			if !first {
				if txSize > consts.MAX_TX_SIZE {
					return nil, ErrInfoFmt("[error] MAX_TX_SIZE")
				}
			}
			txSize = DecodeLength(&binaryData)

			// отчекрыжим одну транзакцию от списка транзакций
			// separate one transaction from the list of transactions
			if txSize > 0 {
				transactionBinaryData := BytesShift(&binaryData, txSize)
				dSha256Hash := DSha256(transactionBinaryData)
				mrklSlice = append(mrklSlice, dSha256Hash)
				//if len(transactionBinaryData) > 500000 {
				//	ioutil.WriteFile(string(dSha256Hash)+"-"+Int64ToStr(txSize), transactionBinaryData, 0644)
				//}
			}

			// чтобы исключить атаку на переполнение памяти
			// to exclude an attack on memory overflow
			if !first {
				if len(mrklSlice) > consts.MAX_TX_COUNT {
					return nil, ErrInfo(fmt.Errorf("[error] MAX_TX_COUNT (%v > %v)", len(mrklSlice), consts.MAX_TX_COUNT))
				}
			}
			if len(binaryData) == 0 {
				break
			}
		}
	} else {
		mrklSlice = append(mrklSlice, []byte("0"))
	}
	log.Debug("mrklSlice: %s", mrklSlice)
	if len(mrklSlice) == 0 {
		mrklSlice = append(mrklSlice, []byte("0"))
	}
	log.Debug("mrklSlice: %s", mrklSlice)
	return MerkleTreeRoot(mrklSlice), nil
}

// SliceReverse reverses the slice of int64
func SliceReverse(s []int64) []int64 {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}

// MerkleTreeRoot rertun Merkle value
func MerkleTreeRoot(dataArray [][]byte) []byte {
	log.Debug("dataArray: %s", dataArray)
	result := make(map[int32][][]byte)
	for _, v := range dataArray {
		result[0] = append(result[0], DSha256(v))
	}
	var j int32
	for len(result[j]) > 1 {
		for i := 0; i < len(result[j]); i = i + 2 {
			if len(result[j]) <= (i + 1) {
				if _, ok := result[j+1]; !ok {
					result[j+1] = [][]byte{result[j][i]}
				} else {
					result[j+1] = append(result[j+1], result[j][i])
				}
			} else {
				if _, ok := result[j+1]; !ok {
					result[j+1] = [][]byte{DSha256(append(result[j][i], result[j][i+1]...))}
				} else {
					result[j+1] = append(result[j+1], DSha256([]byte(append(result[j][i], result[j][i+1]...))))
				}
			}
		}
		j++
	}

	log.Debug("result: %s", result)
	ret := result[int32(len(result)-1)]
	log.Debug("result_: %s", ret)
	return []byte(ret[0])
}

// TypeInt returns the identifier of the embedded transaction
func TypeInt(txType string) int64 {
	for k, v := range consts.TxTypes {
		if v == txType {
			return int64(k)
		}
	}
	return 0
}

// EncryptCFB encrypts the text with AES CFB
func EncryptCFB(text, key, iv []byte) ([]byte, []byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, ErrInfo(err)
	}
	str := text
	if len(iv) == 0 {
		ciphertext := []byte(RandSeq(16))
		iv = ciphertext[:16]
	}
	encrypter := cipher.NewCFBEncrypter(block, iv)
	encrypted := make([]byte, len(str))
	encrypter.XORKeyStream(encrypted, str)

	return append(iv, encrypted...), iv, nil
}

// DecryptCFB decrypts the ciphertext with AES CFB
func DecryptCFB(iv, encrypted, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	decrypter := cipher.NewCFBDecrypter(block, iv)
	decrypted := make([]byte, len(encrypted))
	decrypter.XORKeyStream(decrypted, encrypted)

	return decrypted, nil
}

/*
// http://stackoverflow.com/a/18411978
func VersionOrdinal(version string) string {
	// ISO/IEC 14651:2011
	const maxByte = 1<<8 - 1
	vo := make([]byte, 0, len(version)+8)
	j := -1
	for i := 0; i < len(version); i++ {
		b := version[i]
		if '0' > b || b > '9' {
			vo = append(vo, b)
			j = -1
			continue
		}
		if j == -1 {
			vo = append(vo, 0x00)
			j = len(vo) - 1
		}
		if vo[j] == 1 && vo[j+1] == '0' {
			vo[j+1] = b
			continue
		}
		if vo[j]+1 > maxByte {
			panic("VersionOrdinal: invalid version")
		}
		vo = append(vo, b)
		vo[j]++
	}
	return string(vo)
}*/

// GetNetworkTime returns the network time
func GetNetworkTime() (*time.Time, error) {

	ntpAddr := []string{"0.pool.ntp.org", "europe.pool.ntp.org", "asia.pool.ntp.org", "oceania.pool.ntp.org", "north-america.pool.ntp.org", "south-america.pool.ntp.org", "africa.pool.ntp.org"}
	for i := 0; i < len(ntpAddr); i++ {
		host := ntpAddr[i]
		raddr, err := net.ResolveUDPAddr("udp", host+":123")
		if err != nil {
			continue
		}

		data := make([]byte, 48)
		data[0] = 3<<3 | 3

		con, err := net.DialUDP("udp", nil, raddr)
		if err != nil {
			continue
		}

		defer con.Close()

		_, err = con.Write(data)
		if err != nil {
			continue
		}

		con.SetDeadline(time.Now().Add(5 * time.Second))

		_, err = con.Read(data)
		if err != nil {
			continue
		}

		var sec, frac uint64
		sec = uint64(data[43]) | uint64(data[42])<<8 | uint64(data[41])<<16 | uint64(data[40])<<24
		frac = uint64(data[47]) | uint64(data[46])<<8 | uint64(data[45])<<16 | uint64(data[44])<<24

		nsec := sec * 1e9
		nsec += (frac * 1e9) >> 32

		t := time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC).Add(time.Duration(nsec)).Local()
		return &t, nil
	}
	return nil, errors.New("unable connect to NTP")

}

// SortMap sorts map to the slice of maps
func SortMap(m map[int64]string) []map[int64]string {
	var keys []int
	for k := range m {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	var result []map[int64]string
	for _, k := range keys {
		result = append(result, map[int64]string{int64(k): m[int64(k)]})
	}
	return result
}

// RSortMap sorts map to the reversed slice of maps
func RSortMap(m map[int64]string) []map[int64]string {

	var keys []int
	for k := range m {
		keys = append(keys, int(k))
	}
	sort.Sort(sort.Reverse(sort.IntSlice(keys)))
	var result []map[int64]string
	for _, k := range keys {
		result = append(result, map[int64]string{int64(k): m[int64(k)]})
	}
	return result
}

// TCPConn connects to the address
func TCPConn(Addr string) (net.Conn, error) {
	// шлем данные указанному хосту
	// send data to the specified host
	/*tcpAddr, err := net.ResolveTCPAddr("tcp", Addr)
	if err != nil {
		return nil, ErrInfo(err)
	}
	conn, err := net.DialTCP("tcp", nil, tcpAddr)*/
	conn, err := net.DialTimeout("tcp", Addr, 10*time.Second)
	if err != nil {
		return nil, ErrInfo(err)
	}
	conn.SetReadDeadline(time.Now().Add(consts.READ_TIMEOUT * time.Second))
	conn.SetWriteDeadline(time.Now().Add(consts.WRITE_TIMEOUT * time.Second))
	return conn, nil
}

// WriteSizeAndData writes []byte to the connection
func WriteSizeAndData(binaryData []byte, conn net.Conn) error {
	// в 4-х байтах пишем размер данных, которые пошлем далее
	// record the data size in 4 bytes, which will send further
	size := DecToBin(len(binaryData), 4)
	fmt.Println("len(binaryData)", len(binaryData))
	_, err := conn.Write(size)
	if err != nil {
		return ErrInfo(err)
	}
	// далее шлем сами данные
	// further send data itself 
	if len(binaryData) > 0 {
		/*if len(binaryData) > 500000 {
			ioutil.WriteFile("WriteSizeAndData-7-block-"+IntToStr(len(binaryData))+string(DSha256(binaryData)), binaryData, 0644)
		}*/
		_, err = conn.Write(binaryData)
		if err != nil {
			return ErrInfo(err)
		}
	}
	return nil
}

// GetCurrentDir returns the current directory
func GetCurrentDir() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return "."
	}
	return dir
}

// GetBlockBody gets the block data
func GetBlockBody(host string, blockID int64, dataTypeBlockBody int64) ([]byte, error) {
	conn, err := TCPConn(host)
	if err != nil {
		return nil, ErrInfo(err)
	}
	defer conn.Close()

	log.Debug("dataTypeBlockBody: %v", dataTypeBlockBody)
	// шлем тип данных
	// send the type of data
	_, err = conn.Write(DecToBin(dataTypeBlockBody, 2))
	if err != nil {
		return nil, ErrInfo(err)
	}

	log.Debug("blockID: %v", blockID)

	// шлем номер блока
	// send the number of a block
	_, err = conn.Write(DecToBin(blockID, 4))
	if err != nil {
		return nil, ErrInfo(err)
	}

	// в ответ получаем размер данных, которые нам хочет передать сервер
	// recieve the data size as a response that server wants to transfer
	buf := make([]byte, 4)
	n, err := conn.Read(buf)
	if err != nil {
		return nil, ErrInfo(err)
	}
	log.Debug("dataSize buf: %x / get: %v", buf, n)

	// и если данных менее 10мб, то получаем их
	// if the data size is less than 10mb, we will receive them
	dataSize := BinToDec(buf)
	var binaryBlock []byte
	log.Debug("dataSize: %v", dataSize)
	if dataSize < 10485760 && dataSize > 0 {
		binaryBlock = make([]byte, dataSize)
		/*n, err := conn.Read(binaryBlock)
		log.Debug("dataSize: %v / get: %v", dataSize, n)
		if err != nil {
			return nil, ErrInfo(err)
		}
		if len(binaryBlock) > 500000 {
			ioutil.WriteFile(IntToStr(n)+"-block-"+string(DSha256(binaryBlock)), binaryBlock, 0644)
		}*/
		//binaryBlock, err = ioutil.ReadAll(conn)
		_, err = io.ReadFull(conn, binaryBlock)
		if err != nil {
			return nil, ErrInfo(err)
		}
	} else {
		return nil, ErrInfo("null block")
	}
	return binaryBlock, nil

}

// WriteSelectiveLog writes info into SelectiveLog.txt
func WriteSelectiveLog(text interface{}) {
	if *LogLevel == "DEBUG" {
		var stext string
		switch text.(type) {
		case string:
			stext = text.(string)
		case []byte:
			stext = string(text.([]byte))
		case error:
			stext = fmt.Sprintf("%v", text)
		}
		allTransactionsStr := ""
		allTransactions, _ := DB.GetAll("SELECT hex(hash) as hex_hash, verified, used, high_rate, for_self_use, user_id, third_var, counter, sent FROM transactions", 100)
		for _, data := range allTransactions {
			allTransactionsStr += data["hex_hash"] + "|" + data["verified"] + "|" + data["used"] + "|" + data["high_rate"] + "|" + data["for_self_use"] + "|" + consts.TxTypes[StrToInt(data["type"])] + "|" + data["user_id"] + "|" + data["third_var"] + "|" + data["counter"] + "|" + data["sent"] + "\n"
		}
		t := time.Now()
		data := allTransactionsStr + GetParent() + " ### " + t.Format(time.StampMicro) + " ### " + stext + "\n\n"
		//ioutil.WriteFile(*Dir+"/SelectiveLog.txt", []byte(data), 0644)
		f, err := os.OpenFile(*Dir+"/SelectiveLog.txt", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
		if err != nil {
			panic(err)
		}

		defer f.Close()

		if _, err = f.WriteString(data); err != nil {
			panic(err)
		}
	}
}

/*
func DaylightRestart() error {
	log.Debug("exec", os.Args[0])
	err := exec.Command(os.Args[0]).Start()
	if err != nil {
		return ErrInfo(err)
	}
	return nil
}*/

// GetUpdVerAndURL downloads the information about the version
func GetUpdVerAndURL(host string) (updinfo *lib.Update, err error) {

	update, err := GetHTTPTextAnswer(host + "/update.json")
	//update, err := ioutil.ReadFile(`c:\egaas\update.json`)
	if len(update) > 0 {
		updateData := make(map[string]lib.Update)
		err = json.Unmarshal([]byte(update), &updateData)
		if err != nil {
			return
		}
		if upd, ok := updateData[runtime.GOOS+`_`+runtime.GOARCH]; ok && version.Compare(upd.Version, consts.VERSION, ">") {
			updinfo = &upd
		}
	}
	return
}

// ShellExecute runs cmdline
func ShellExecute(cmdline string) {
	time.Sleep(500 * time.Millisecond)
	switch runtime.GOOS {
	case "linux":
		exec.Command("xdg-open", cmdline).Start()
	case "windows":
		exec.Command(`rundll32.exe`, `url.dll,FileProtocolHandler`, cmdline).Start()
	case "darwin":
		exec.Command("open", cmdline).Start()
	}
}

// DecodeLength decodes length from []byte
func DecodeLength(buf *[]byte) (ret int64) {
	ret, _ = lib.DecodeLength(buf)
	return
}

// CreateHTMLFromTemplate gets the template of the page from the table and proceeds it
func CreateHTMLFromTemplate(page string, citizenID, stateID int64, params *map[string]string) (string, error) {
	query := `SELECT value FROM "` + Int64ToStr(stateID) + `_pages" WHERE name = ?`
	if (*params)[`global`] == `1` {
		query = `SELECT value FROM global_pages WHERE name = ?`
	}

	data, err := DB.Single(query, page).String()
	if err != nil {
		return "", err
	}
	/*	qrx := regexp.MustCompile(`CitizenId`)
		data = qrx.ReplaceAllString(data, Int64ToStr(citizenId))
		qrx = regexp.MustCompile(`AccountId`)
		data = qrx.ReplaceAllString(data, Int64ToStr(accountId))*/
	(*params)[`page`] = page
	(*params)[`state_id`] = Int64ToStr(stateID)
	(*params)[`citizen`] = Int64ToStr(citizenID)
	if len(data) > 0 {
		templ := textproc.Process(data, params)
		if (*params)[`isrow`] == `opened` {
			templ += `</div>`
			(*params)[`isrow`] = ``
		}
		templ = LangMacro(templ, int(stateID), (*params)[`accept_lang`])
		getHeight := func() int64 {
			height := int64(100)
			if h, ok := (*params)[`hmap`]; ok {
				height = StrToInt64(h)
			}
			return height
		}
		if len((*params)[`wisource`]) > 0 {
			templ += fmt.Sprintf(`<script language="JavaScript" type="text/javascript">
			var editor = ace.edit("textEditor");
	var ContractMode = ace.require("ace/mode/c_cpp").Mode;
	ace.require("ace/ext/language_tools");
	$(".textEditor code").html(editor.getValue());
	$("#%s").val(editor.getValue());
	editor.setTheme("ace/theme/chrome");
    editor.session.setMode(new ContractMode());
	editor.setShowPrintMargin(false);
	editor.getSession().setTabSize(4);
	editor.getSession().setUseWrapMode(true);
	editor.getSession().on('change', function(e) {
		$(".textEditor code").html(editor.getValue());
		$("#%s").val(editor.getValue());
		editor.resize();
	});
	editor.setOptions({
		enableBasicAutocompletion: true,
		enableSnippets: true,
		enableLiveAutocompletion: true
	});
			</script>`, (*params)[`wisource`], (*params)[`wisource`])
		}
		if (*params)[`wimoney`] == `1` {
			templ += fmt.Sprintf(`<script language="JavaScript" type="text/javascript">
				$(".inputmask").inputmask({'autoUnmask': true});</script>`)
		}
		if (*params)[`widate`] == `1` {
			templ += fmt.Sprintf(`<script language="JavaScript" type="text/javascript">
						$(document).ready(function() {
							$.datetimepicker.setLocale('en');
							$(".datetimepicker").datetimepicker();
						})
				</script>`)
		}
		if (*params)[`wiaddress`] == `1` {
			templ += fmt.Sprintf(`<script language="JavaScript" type="text/javascript">
				$(".address").prop("autocomplete", "off").inputmask({mask: "9999-9999-9999-9999-9999", autoUnmask: true }).focus();
	$(".address").typeahead({
		minLength: 1,
		items: 10,
		source: function (query, process) {
			return $.get('ajax?json=ajax_addresses', { 'address': query }, function (data) {
				return process(data.address);
			});
		}
	}).focus();</script>`)
		}
		if (*params)[`wimap`] == `1` {
			templ += fmt.Sprintf(`<script language="JavaScript" type="text/javascript">
			miniMap("wimap", "100%%", "%dpx");</script>`, getHeight())
		}
		if (*params)[`wicitizen`] == `1` {
			templ += fmt.Sprintf(`<script language="JavaScript" type="text/javascript">(function($, window, document){
'use strict';
  var Selector = '[data-notify]',
      autoloadSelector = '[data-onload]',
      doc = $(document);

  $(function() {
    $(Selector).each(function(){
      var $this  = $(this),
          onload = $this.data('onload');
      if(onload !== undefined) {
        setTimeout(function(){
          notifyNow($this);
        }, 800);
      }
      $this.on('click', function (e) {
        e.preventDefault();
        notifyNow($this);
      });
    });
  });
  function notifyNow($element) {
      var message = $element.data('message'),
          options = $element.data('options');
 	 if(!message)
        $.error('Notify: No message specified');
      $.notify(message, options || {});
  }
}(jQuery, window, document));</script>`)
		}
		if (*params)[`wimappoint`] == `1` {
			templ += fmt.Sprintf(`<script language="JavaScript" type="text/javascript">
			userLocation("wimappoint", "100%%", "%dpx");</script>`, getHeight())
		}
		if (*params)[`wibtncont`] == `1` {
			var unique int64
			if uval, ok := (*params)[`tx_unique`]; ok {
				unique = StrToInt64(uval) + 1
			}
			(*params)[`tx_unique`] = Int64ToStr(unique)
			funcMap := template.FuncMap{
				"sum": func(a, b interface{}) float64 {
					return InterfaceToFloat64(a) + InterfaceToFloat64(b)
				},
				"noescape": func(s string) template.HTML {
					return template.HTML(s)
				},
			}
			data, err := static.Asset("static/tx_btncont.html")
			if err != nil {
				return ``, err
			}
			sign, err := static.Asset("static/signatures_new.html")
			if err != nil {
				return ``, err
			}

			t := template.New("template").Funcs(funcMap)
			if t, err = t.Parse(string(data)); err != nil {
				return ``, err
			}
			t = template.Must(t.Parse(string(sign)))
			b := new(bytes.Buffer)

			finfo := TxBtnCont{ //Class: class, ClassBtn: classBtn, Name: LangRes(vars, btnName),
				Unique: template.JS((*params)[`tx_unique`]), // OnSuccess: template.JS(onsuccess),
				//Fields: make([]TxInfo, 0), AutoClose: (*pars)[`AutoClose`] != `0`,
				/*Silent: (*pars)[`Silent`] == `1`*/}
			if err = t.Execute(b, finfo); err != nil {
				return ``, err
			}
			templ += b.String()
		}
		return ProceedTemplate(`page_template`, &PageTpl{Page: page, Template: templ})
	}
	return ``, nil
}

// FirstBlock generates the first block
func FirstBlock(exit bool) {
	log.Debug("FirstBlock")

	if *GenerateFirstBlock == 1 {

		log.Debug("GenerateFirstBlock == 1")

		if len(*FirstBlockPublicKey) == 0 {
			log.Debug("len(*FirstBlockPublicKey) == 0")
			priv, pub, _ := lib.GenHexKeys()
			err := ioutil.WriteFile(*Dir+"/PrivateKey", []byte(priv), 0644)
			if err != nil {
				log.Error("%v", ErrInfo(err))
			}
			*FirstBlockPublicKey = pub
		}
		if len(*FirstBlockNodePublicKey) == 0 {
			log.Debug("len(*FirstBlockNodePublicKey) == 0")
			priv, pub, _ := lib.GenHexKeys()
			err := ioutil.WriteFile(*Dir+"/NodePrivateKey", []byte(priv), 0644)
			if err != nil {
				log.Error("%v", ErrInfo(err))
			}
			*FirstBlockNodePublicKey = pub
		}

		PublicKey := *FirstBlockPublicKey
		log.Debug("PublicKey", PublicKey)
		//		PublicKeyBytes, _ := base64.StdEncoding.DecodeString(string(PublicKey))
		PublicKeyBytes, _ := hex.DecodeString(string(PublicKey))

		NodePublicKey := *FirstBlockNodePublicKey
		log.Debug("NodePublicKey", NodePublicKey)
		//		NodePublicKeyBytes, _ := base64.StdEncoding.DecodeString(string(NodePublicKey))
		NodePublicKeyBytes, _ := hex.DecodeString(string(NodePublicKey))
		Host := *FirstBlockHost
		if len(Host) == 0 {
			Host = "127.0.0.1"
		}

		var block, tx []byte
		iAddress := int64(lib.Address(PublicKeyBytes))
		now := lib.Time32()
		_, err := lib.BinMarshal(&block, &consts.BlockHeader{Type: 0, BlockID: 1, Time: now, WalletID: iAddress})
		if err != nil {
			log.Error("%v", ErrInfo(err))
		}
		_, err = lib.BinMarshal(&tx, &consts.FirstBlock{TxHeader: consts.TxHeader{Type: 1,
			Time: now, WalletID: iAddress, CitizenID: 0},
			PublicKey: PublicKeyBytes, NodePublicKey: NodePublicKeyBytes, Host: string(Host)})
		if err != nil {
			log.Error("%v", ErrInfo(err))
		}
		lib.EncodeLenByte(&block, tx)

		firstBlockDir := ""
		if len(*FirstBlockDir) == 0 {
			firstBlockDir = *Dir
		} else {
			firstBlockDir = filepath.Join("", *FirstBlockDir)
			if _, err := os.Stat(firstBlockDir); os.IsNotExist(err) {
				if err = os.Mkdir(firstBlockDir, 0755); err != nil {
					log.Error("%v", ErrInfo(err))
				}
			}
		}
		ioutil.WriteFile(filepath.Join(firstBlockDir, "1block"), block, 0644)
		if exit {
			os.Exit(0)
		}
	}
}

// EgaasUpdate decompresses and updates executable file
func EgaasUpdate(url string) error {
	//	GetUpdVerAndURL(host string) (updinfo *lib.Update, err error)

	zipfile := filepath.Join(*Dir, "egaas.zip")
	/*	_, err := DownloadToFile(url, zipfile, 3600, nil, nil, "upd")
		if err != nil {
			return ErrInfo(err)
		}
		fmt.Println(zipfile)*/
	reader, err := zip.OpenReader(zipfile)
	if err != nil {
		return ErrInfo(err)
	}
	appname := filepath.Base(os.Args[0])
	tmpname := filepath.Join(*Dir, `tmp_`+appname)

	ftemp := reader.Reader.File
	f := ftemp[0]
	zipped, err := f.Open()
	if err != nil {
		return ErrInfo(err)
	}

	writer, err := os.OpenFile(tmpname, os.O_WRONLY|os.O_CREATE, f.Mode())
	if err != nil {
		return ErrInfo(err)
	}

	if _, err = io.Copy(writer, zipped); err != nil {
		return ErrInfo(err)
	}
	reader.Close()
	zipped.Close()
	writer.Close()

	/*	pwd, err := os.Getwd()
		if err != nil {
			return ErrInfo(err)
		}
		fmt.Print(pwd)*/

	folderPath, err := osext.ExecutableFolder()
	if err != nil {
		return ErrInfo(err)
	}

	old := ""
	if _, err := os.Stat(os.Args[0]); err == nil {
		old = os.Args[0]
	} else if _, err := os.Stat(filepath.Join(folderPath, appname)); err == nil {
		old = filepath.Join(folderPath, appname)
	} else {
		old = filepath.Join(*Dir, appname)
	}
	//	log.Debug(tmpname, "-oldFileName", old, "-dir", *Dir, "-oldVersion", consts.VERSION)
	err = exec.Command(tmpname, "-oldFileName", old, "-dir", *Dir, "-oldVersion", consts.VERSION).Start()
	if err != nil {
		return ErrInfo(err)
	}
	return nil
}

/*
func OutInit() {
	odir, _ := filepath.Abs(os.Args[0])
	OutFile, _ = os.OpenFile(odir+`.txt`, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	//	defer utils.OutFile.Close()
}

func Out(pars ...interface{}) {
	OutFile.WriteString(fmt.Sprint(pars...) + "\r\n")
}*/

// GetPrefix возвращает префикс у таблицы. При этом идет проверка, чтобы префикс был global или совпадал
// GetPrefix returns the prefix of the table. In this case it is checked that the prefix was global or matched
// с идентифкатором государства
// with the identifier of the state
func GetPrefix(tableName, stateID string) (string, error) {
	s := strings.Split(tableName, "_")
	if len(s) < 2 {
		return "", ErrInfo("incorrect table name")
	}
	prefix := s[0]
	if prefix != "global" && prefix != stateID {
		return "", ErrInfo("incorrect table name")
	}
	return prefix, nil
}
