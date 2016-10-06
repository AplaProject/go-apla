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
	"crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	//	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"flag"
	"fmt"
	"image"
	"io"
	"io/ioutil"
	"math"
	"math/big"
	"math/rand"
	"net"
	"net/http"

	"github.com/DayLightProject/go-daylight/packages/consts"
	"github.com/DayLightProject/go-daylight/packages/lib"
	"github.com/DayLightProject/go-daylight/packages/script"
	"github.com/DayLightProject/go-daylight/packages/smart"
	"github.com/DayLightProject/go-daylight/packages/static"
	b58 "github.com/jbenet/go-base58"
	"github.com/kardianos/osext"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"github.com/mcuadros/go-version"
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
	"sync"
	"time"

	"github.com/russross/blackfriday"
)

type BlockData struct {
	BlockId       int64
	Time          int64
	WalletId      int64
	CBID          int64
	CurrentUserId int64
	Sign          []byte
	Hash          []byte
}

type prevBlockType struct {
	Hash     string
	HeadHash string
	BlockId  int64
	Time     int64
	Level    int64
}
type DaemonsChansType struct {
	ChBreaker chan bool
	ChAnswer  chan string
}

var (
	TcpHost            = flag.String("tcpHost", "", "tcpHost (e.g. 127.0.0.1)")
	ListenHttpPort     = flag.String("listenHttpPort", "7079", "ListenHttpPort")
	GenerateFirstBlock = flag.Int64("generateFirstBlock", 0, "generateFirstBlock")
	OldVersion         = flag.String("oldVersion", "", "")
	TestRollBack       = flag.Int64("testRollBack", 0, "testRollBack")
	Dir                = flag.String("dir", GetCurrentDir(), "DayLight directory")
	OldFileName        = flag.String("oldFileName", "", "")
	LogLevel           = flag.String("logLevel", "", "DayLight LogLevel")
	Console            = flag.Int64("console", 0, "Start from console")
	SqliteDbUrl        string
	StartBlockId       = flag.Int64("startBlockId", 0, "Start block for blockCollection daemon")
	EndBlockId         = flag.Int64("endBlockId", 0, "End block for blockCollection daemon")
	RollbackToBlockId  = flag.Int64("rollbackToBlockId", 0, "Rollback to block_id")
	Tls                = flag.String("tls", "", "Support https. Specify directory for .well-known")
	DaemonsChans       []*DaemonsChansType
	eWallets           = &sync.Mutex{}
)

func init() {
	flag.Parse()
}

func IOS() bool {
	if (runtime.GOARCH == "arm" || runtime.GOARCH == "arm64") && runtime.GOOS == "darwin" {
		return true
	}
	return false
}
func Desktop() bool {
	thrust_shell := "thrust_shell"
	if runtime.GOOS == "windows" {
		thrust_shell = "thrust_shell.exe"
	} else if runtime.GOOS == "darwin" {
		thrust_shell = "ThrustShell"
	}
	if _, err := os.Stat(*Dir + "/" + thrust_shell); err == nil {
		return true
	}
	return false
}
func Mobile() bool {
	if IOS() || runtime.GOOS == "android" {
		return true
	}
	return false
}
func Android() bool {
	if runtime.GOOS == "android" {
		return true
	}
	return false
}
func Sleep(sec time.Duration) {
	//log.Debug("time.Duration(sec): %v / %v",sec, GetParent())
	time.Sleep(sec * time.Second)
}

type SortCfCatalog []map[string]string

func (s SortCfCatalog) Len() int {
	return len(s)
}
func (s SortCfCatalog) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s SortCfCatalog) Less(i, j int) bool {
	return s[i]["name"] < s[j]["name"]
}
func MakeCfCategories(lang map[string]string) []map[string]string {
	var cfCategory []map[string]string
	for i := 0; i < 18; i++ {
		cfCategory = append(cfCategory, map[string]string{"id": IntToStr(i), "name": lang["cf_category_"+IntToStr(i)]})
	}
	sort.Sort(SortCfCatalog(cfCategory))
	return cfCategory
}

func getImageDimension(imagePath string) (int, int) {
	/*file, err := os.Open(imagePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
	}
	defer file.Close()*/
	data, _ := static.Asset(imagePath)
	file := bytes.NewReader(data)
	image, _, err := image.DecodeConfig(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", imagePath, err)
	}
	return image.Width, image.Height
}

type ParamType struct {
	X, Y, Width, Height int64
	Bg_path             string
}

func ParseBlockHeader(binaryBlock *[]byte) *BlockData {
	result := new(BlockData)
	// распарсим заголовок блока
	/*
		Заголовок
		TYPE (0-блок, 1-тр-я)        1
		BLOCK_ID   				       4
		TIME       					       4
		WALLET_ID                         1-8
		state_id                              1
		SIGN                               от 128 до 512 байт. Подпись от TYPE, BLOCK_ID, PREV_BLOCK_HASH, TIME, WALLET_ID, state_id, MRKL_ROOT
		Далее - тело блока (Тр-ии)
	*/
	result.BlockId = BinToDecBytesShift(binaryBlock, 4)
	result.Time = BinToDecBytesShift(binaryBlock, 4)
	result.WalletId, _ = DecodeLenInt64(binaryBlock) //BytesToInt64(BytesShift(binaryBlock, DecodeLength(binaryBlock)))
	// Delete after re-build blocks
	/*	if result.WalletId == 0x31 {
		result.WalletId = 1
	}*/
	result.CBID = BinToDecBytesShift(binaryBlock, 1)
	if result.BlockId > 1 {
		signSize := DecodeLength(binaryBlock)
		result.Sign = BytesShift(binaryBlock, signSize)
	} else {
		*binaryBlock = (*binaryBlock)[1:]
	}
	log.Debug("result.BlockId: %v / result.Time: %v / result.WalletId: %v / result.CBID: %v / result.Sign: %v", result.BlockId, result.Time, result.WalletId, result.CBID, result.Sign)
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

// ищем ближайшее время в $points_status_array или $max_promised_amount_array
// $type - status для $points_status_array / amount - для $max_promised_amount_array
func findMinPointsStatus(needTime int64, pointsStatusArray []map[int64]string, pType string) ([]map[string]string, []map[int64]string) {
	var findTime []int64
	newPointsStatusArray := pointsStatusArray
	var timeStatusArr []map[string]string
BR:
	for i := 0; i < len(pointsStatusArray); i++ {
		for time, _ := range pointsStatusArray[i] {
			if time > needTime {
				break BR
			}
			findTime = append(findTime, time)
			start := i + 1
			if i+1 > len(pointsStatusArray) {
				start = len(pointsStatusArray)
			}
			newPointsStatusArray = pointsStatusArray[start:]
		}
	}
	if len(findTime) > 0 {
		for i := 0; i < len(findTime); i++ {
			for _, status := range pointsStatusArray[i] {
				timeStatusArr = append(timeStatusArr, map[string]string{"time": Int64ToStr(findTime[i]), pType: status})
			}
		}
	}
	return timeStatusArr, newPointsStatusArray
}

func findMinPct(needTime int64, pctArray []map[int64]map[string]float64, status string) float64 {
	var findTime int64 = -1
	var pct float64 = 0
BR:
	for i := 0; i < len(pctArray); i++ {
		for time, _ := range pctArray[i] {
			if time > needTime {
				break BR
			}
			findTime = int64(i)
		}
	}
	if findTime >= 0 {
		for _, arr := range pctArray[findTime] {
			pct = arr[status]
		}
	}
	return pct
}

func findMinPct1(needTime int64, pctArray []map[int64]float64) float64 {
	var findTime int64 = -1
	var pct float64 = 0
BR:
	for i := 0; i < len(pctArray); i++ {
		for time, _ := range pctArray[i] {
			if time > needTime {
				break BR
			}
			findTime = int64(i)
		}
	}
	if findTime >= 0 {
		for _, pct0 := range pctArray[findTime] {
			pct = pct0
		}
	}
	return pct
}

func getMaxPromisedAmountCalcProfit(amount, repaidAmount, maxPromisedAmount float64, currencyId int64) float64 {
	// для WOC $repaid_amount всегда = 0, т.к. cash_request на WOC послать невозможно
	// если наша сумма больше, чем максимально допустимая ($find_min_array[$i]['amount'])
	var result float64
	if amount+repaidAmount > maxPromisedAmount {
		result = maxPromisedAmount - repaidAmount
	} else if amount < maxPromisedAmount && currencyId == 1 { // для WOC разрешено брать maxPromisedAmount вместо promisedAmount, если promisedAmount < maxPromisedAmount
		result = maxPromisedAmount
	} else {
		result = amount
	}
	return result
}

type resultArrType struct {
	num_sec int64
	pct     float64
	amount  float64
}

type pctAmount struct {
	pct    float64
	amount float64
}

func round(num float64) int64 {
	//log.Debug("num", num)
	//num += ROUND_FIX
	//	return int(StrToFloat64(Float64ToStr(num)) + math.Copysign(0.5, num))
	//log.Debug("num", num)
	return int64(num + math.Copysign(0.5, num))
}

func Round(num float64, precision int) float64 {
	num += consts.ROUND_FIX
	//log.Debug("num", num)
	//num = StrToFloat64(Float64ToStr(num))
	//log.Debug("precision", precision)
	//log.Debug("float64(precision)", float64(precision))
	output := math.Pow(10, float64(precision))
	//log.Debug("output", output)
	return float64(round(num*output)) / output
}

func RandSlice(min, max, count int64) []string {
	var result []string
	for i := 0; i < int(count); i++ {
		result = append(result, IntToStr(RandInt(int(min), int(max))))
	}
	return result
}

func RandInt(min int, max int) int {
	if max-min <= 0 {
		return 1
	}
	return min + rand.Intn(max-min)
}

func PpLenght(p1, p2 [2]int) float64 {
	return math.Sqrt(math.Pow(float64(p1[0]-p2[0]), 2) + math.Pow(float64(p1[1]-p2[1]), 2))
}

func CheckInputData(data_ interface{}, dataType string) bool {
	return CheckInputData_(data_, dataType, "")
}

// функция проверки входящих данных
func CheckInputData_(data_ interface{}, dataType string, info string) bool {
	var data string
	switch data_.(type) {
	case int:
		data = IntToStr(data_.(int))
	case int64:
		data = Int64ToStr(data_.(int64))
	case float64:
		data = Float64ToStr(data_.(float64))
	case string:
		data = data_.(string)
	case []byte:
		data = string(data_.([]byte))
	}
	log.Debug("CheckInputData_:" + data)
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
	case "reduction_type":
		if ok, _ := regexp.MatchString(`^(manual|promised_amount)$`, data); ok {
			if StrToInt(data) <= 30 {
				return true
			}
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
	case "cf_currency_name":
		if ok, _ := regexp.MatchString(`^[A-Z0-9]{7}$`, data); ok {
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
	case "currency_name":
		if ok, _ := regexp.MatchString(`^[A-Z]{3}$`, data); ok {
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
		r := `^\[(` + xy + `\,){` + info + `}` + xy + `\]$`
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
		if ok, _ := regexp.MatchString("^(?i)[0-9a-z]{25,34}$", data); ok {
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
		if ok, _ := regexp.MatchString("^[0-9]{1,15}$", data); ok {
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

func Time() int64 {
	return time.Now().Unix()
}

func TimeF(timeFormat string) string {
	t := time.Unix(time.Now().Unix(), 0)
	return t.Format(timeFormat)
}

func ValidateEmail(email string) bool {
	Re := regexp.MustCompile(`^(?i)[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	return Re.MatchString(email)
}

func GetHttpTextAnswer(url string) (string, error) {
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

func RemoteAddrFix(addr string) string {
	if ok, _ := regexp.MatchString(`(\:\:)|(127\.0\.0\.1)`, addr); ok {
		return ""
	} else {
		return addr
	}
}

// без проверки на ошибки т.к. тут ошибки не могут навредить
func StrToInt64(s string) int64 {
	int64, _ := strconv.ParseInt(s, 10, 64)
	return int64
}
func BytesToInt64(s []byte) int64 {
	int64, _ := strconv.ParseInt(string(s), 10, 64)
	return int64
}
func StrToUint64(s string) uint64 {
	int64, _ := strconv.ParseInt(s, 10, 64)
	return uint64(int64)
}
func StrToInt(s string) int {
	int_, _ := strconv.Atoi(s)
	return int_
}
func Float64ToStr(f float64) string {
	return strconv.FormatFloat(f, 'f', 13, 64)
}
func Float64ToStrGeo(f float64) string {
	return strconv.FormatFloat(f, 'f', 5, 64)
}
func Float64ToBytes(f float64) []byte {
	return []byte(strconv.FormatFloat(f, 'f', 13, 64))
}
func Float64ToStrPct(f float64) string {
	if f == 0 {
		return "0"
	} else {
		return strconv.FormatFloat(f, 'f', 2, 64)
	}
}
func StrToFloat64(s string) float64 {
	Float64, _ := strconv.ParseFloat(s, 64)
	return Float64
}
func BytesToFloat64(s []byte) float64 {
	Float64, _ := strconv.ParseFloat(string(s), 64)
	return Float64
}
func BytesToInt(s []byte) int {
	int_, _ := strconv.Atoi(string(s))
	return int_
}
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

func GetEndBlockId() (int64, error) {

	if _, err := os.Stat(*Dir + "/public/blockchain"); os.IsNotExist(err) {
		return 0, nil
	} else {

		// размер блока, записанный в 5-и последних байтах файла blockchain
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
		BinToDecBytesShift(&dataBinary, 5)
		blockId := BinToDecBytesShift(&dataBinary, 5)
		return blockId, nil

	}
	return 0, nil
}

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

func CheckErr(err error) {
	if err != nil {
		panic(fmt.Sprintf("%s", err))
	}
}

func ErrInfoFmt(err string, a ...interface{}) error {
	err_ := fmt.Sprintf(err, a...)
	return fmt.Errorf("%s (%s)", err_, Caller(1))
}

func ErrInfo(err_ interface{}, additionally ...string) error {
	var err error
	switch err_.(type) {
	case error:
		err = err_.(error)
	case string:
		err = errors.New(err_.(string))
	}
	if err != nil {
		if len(additionally) > 0 {
			return fmt.Errorf("%s # %s (%s)", err, additionally, Caller(1))
		} else {
			return fmt.Errorf("%s (%s)", err, Caller(1))
		}
	}
	return err
}

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

func Caller(steps int) string {
	name := "?"
	if pc, _, num, ok := runtime.Caller(steps + 1); ok {
		//fmt.Println(num)
		name = fmt.Sprintf("%s :  %d", filepath.Base(runtime.FuncForPC(pc).Name()), num)
	}
	return name
}

func SliceInt64ToString(int64 []int64) []string {
	result := make([]string, len(int64))
	for i, v := range int64 {
		result[i] = strconv.FormatInt(v, 10)
	}
	return result
}

func RemoveInt64Slice(slice *[]int64, pos int) {
	sl := *slice
	*slice = append(sl[:pos], sl[pos+1:]...)
}

func DelUserIdFromArray(array *[]int64, userId int64) {
	for i, v := range *array {
		if v == userId {
			RemoveInt64Slice(&*array, i)
		}
	}
}

func InSliceInt64(search int64, slice []int64) bool {
	for _, v := range slice {
		if v == search {
			return true
		}
	}
	return false
}

func InSliceString(search string, slice []string) bool {
	for _, v := range slice {
		if v == search {
			return true
		}
	}
	return false
}

func EncodeLengthPlusData(data_ interface{}) []byte {
	var data []byte
	switch data_.(type) {
	case int64:
		data = Int64ToByte(data_.(int64))
	case string:
		data = []byte(data_.(string))
	case []byte:
		data = data_.([]byte)
	}
	//log.Debug("data: %x", data)
	//log.Debug("len data: %d", len(data))
	return append(EncodeLength(int64(len(data))), data...)
}

func DecToHex(dec int64) string {
	return strconv.FormatInt(dec, 16)
}

func HexToDec(h string) int64 {
	int64, _ := strconv.ParseInt(h, 16, 0)
	return int64
}

func HexToDecBig(hex string) string {
	i := new(big.Int)
	i.SetString(hex, 16)
	return fmt.Sprintf("%d", i)
}

func DecToHexBig(hex string) string {
	i := new(big.Int)
	i.SetString(hex, 10)
	hex = fmt.Sprintf("%x", i)
	if len(hex)%2 > 0 {
		hex = "0" + hex
	}
	return hex
}

func UInt32ToStr(num uint32) string {
	return strconv.FormatInt(int64(num), 10)
}
func Int64ToStr(num int64) string {
	return strconv.FormatInt(num, 10)
}
func Int64ToByte(num int64) []byte {
	return []byte(strconv.FormatInt(num, 10))
}

func IntToStr(num int) string {
	return strconv.Itoa(num)
}

func DecToBin(dec_ interface{}, sizeBytes int64) []byte {
	var dec int64
	switch dec_.(type) {
	case int:
		dec = int64(dec_.(int))
	case int64:
		dec = dec_.(int64)
	case string:
		dec = StrToInt64(dec_.(string))
	}
	Hex := fmt.Sprintf("%0"+Int64ToStr(sizeBytes*2)+"x", dec)
	//fmt.Println("Hex", Hex)
	return HexToBin([]byte(Hex))
}
func BinToHex(bin_ interface{}) []byte {
	var bin []byte
	switch bin_.(type) {
	case []byte:
		bin = bin_.([]byte)
	case int64:
		bin = Int64ToByte(bin_.(int64))
	case string:
		bin = []byte(bin_.(string))
	}
	return []byte(fmt.Sprintf("%x", bin))
}

func HexToBin(hexdata_ interface{}) []byte {
	var hexdata string
	switch hexdata_.(type) {
	case []byte:
		hexdata = string(hexdata_.([]byte))
	case int64:
		hexdata = Int64ToStr(hexdata_.(int64))
	case string:
		hexdata = hexdata_.(string)
	}
	var str []byte
	str, err := hex.DecodeString(hexdata)
	if err != nil {
		log.Error("%v / %v", err, GetParent())
	}
	return str
}

func BinToDec(bin []byte) int64 {
	var a uint64
	l := len(bin)
	for i, b := range bin {
		shift := uint64((l - i - 1) * 8)
		a |= uint64(b) << shift
	}
	return int64(a)
}

func BinToDecBytesShift(bin *[]byte, num int64) int64 {
	return BinToDec(BytesShift(bin, num))
}

func BytesShift(str *[]byte, index int64) []byte {
	if int64(len(*str)) < index {
		return []byte("")
	}
	var substr []byte
	var str_ []byte
	substr = *str
	substr = substr[0:index]
	str_ = *str
	str_ = str_[index:]
	*str = str_
	return substr
}

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
	}
	return str
}
func InterfaceSliceToStr(i []interface{}) []string {
	var str []string
	for _, v := range i {
		switch v.(type) {
		case int:
			str = append(str, IntToStr(v.(int)))
		case float64:
			str = append(str, Float64ToStr(v.(float64)))
		case int64:
			str = append(str, Int64ToStr(v.(int64)))
		case string:
			str = append(str, v.(string))
		case []byte:
			str = append(str, string(v.([]byte)))
		}
	}
	return str
}

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

func BytesShiftReverse(str *[]byte, index_ interface{}) []byte {
	var index int64
	switch index_.(type) {
	case int:
		index = int64(index_.(int))
	case int64:
		index = index_.(int64)
	}

	var substr []byte
	var str_ []byte
	substr = *str
	substr = substr[int64(len(substr))-index:]
	//fmt.Println(substr)
	str_ = *str
	if int64(len(str_)) < int64(len(str_))-index {
		return []byte("")
	}
	str_ = str_[0 : int64(len(str_))-index]
	*str = str_
	//fmt.Println(utils.BinToHex(str_))
	return substr
}

func SleepDiff(sleep *int64, diff int64) {
	// вычитаем уже прошедшее время
	if *sleep > diff {
		*sleep = *sleep - diff
	} else {
		*sleep = 0
	}
}

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

func PKCS5Padding(src []byte, blockSize int) []byte {
	padding := blockSize - len(src)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(src, padtext...)
}

func PKCS5UnPadding(src []byte) []byte {
	length := len(src)
	unpadding := int(src[length-1])
	return src[:(length - unpadding)]
}

func RandSeq(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	rand.Seed(time.Now().UnixNano())
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func GetPublicFromPrivate(key string) ([]byte, error) {
	block, _ := pem.Decode([]byte(key))
	if block == nil {
		return nil, errors.New("bad key data")
	}
	log.Debug("%v", block)
	if got, want := block.Type, "RSA PRIVATE KEY"; got != want {
		return nil, errors.New("unknown key type " + got + ", want " + want)
	}
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	log.Debug("privateKey %v", privateKey)
	if err != nil {
		return nil, ErrInfo(err)
	}
	e := fmt.Sprintf("%x", privateKey.PublicKey.E)
	if len(e)%2 > 0 {
		e = "0" + e
	}
	n := BinToHex(privateKey.PublicKey.N.Bytes())
	n = append([]byte("00"), n...)
	log.Debug("%s / %v", n, e)
	publicKeyAsn := MakeAsn1(n, []byte(e))
	return publicKeyAsn, nil
}

func MakeAsn1(hex_n, hex_e []byte) []byte {
	//hex_n = append([]byte("00"), hex_n...)
	n_ := []byte(HexToBin(hex_n))
	n_ = append([]byte("02"), BinToHex(EncodeLength(int64(len(HexToBin(hex_n)))))...)
	//log.Debug("n_length", string(n_))
	n_ = append(n_, hex_n...)
	//log.Debug("n_", string(n_))
	e_ := append([]byte("02"), BinToHex(EncodeLength(int64(len(HexToBin(hex_e)))))...)
	e_ = append(e_, hex_e...)
	//log.Debug("e_", string(e_))
	length := BinToHex(EncodeLength(int64(len(HexToBin(append(n_, e_...))))))
	//log.Debug("length", string(length))
	rez := append([]byte("30"), length...)
	rez = append(rez, n_...)
	rez = append(rez, e_...)
	rez = append([]byte("00"), rez...)
	//log.Debug("%v", string(rez))
	//log.Debug("%v", len(string(rez)))
	//log.Debug("%v", len(HexToBin(rez)))
	rez = append(BinToHex(EncodeLength(int64(len(HexToBin(rez))))), rez...)
	rez = append([]byte("03"), rez...)
	//log.Debug("%v", string(rez))
	rez = append([]byte("300d06092a864886f70d0101010500"), rez...)
	//log.Debug("%v", string(rez))
	rez = append(BinToHex(EncodeLength(int64(len(HexToBin(rez))))), rez...)
	//log.Debug("%v", string(rez))
	rez = append([]byte("30"), rez...)

	//log.Debug("hex_n: %s", hex_n)
	//log.Debug("hex_e: %s", hex_e)
	//log.Debug("%v", string(rez))

	return rez
	//b64:=base64.StdEncoding.EncodeToString([]byte(utils.HexToBin("30"+length+bin_enc)))
	//fmt.Println(b64)
}

func BinToRsaPubKey(publicKey []byte) (*rsa.PublicKey, error) {
	key := base64.StdEncoding.EncodeToString(publicKey)
	key = "-----BEGIN PUBLIC KEY-----\n" + key + "\n-----END PUBLIC KEY-----"
	//fmt.Printf("%x\n", publicKeys[i])
	log.Debug("key", key)
	block, _ := pem.Decode([]byte(key))
	if block == nil {
		return nil, ErrInfo(fmt.Errorf("incorrect key"))
	}
	re, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, ErrInfo(err)
	}
	pub := re.(*rsa.PublicKey)
	if err != nil {
		return nil, ErrInfo(err)
	}
	return pub, nil
}

func CheckSign(publicKeys [][]byte, forSign string, signs []byte, nodeKeyOrLogin bool) (bool, error) {

	/*	log.Debug("forSign", forSign)
		//fmt.Println("publicKeys", publicKeys)
		var signsSlice [][]byte
		// у нода всегда 1 подпись
		if nodeKeyOrLogin {
			signsSlice = append(signsSlice, signs)
		} else {
			// в 1 signs может быть от 1 до 3-х подписей
			for {
				if len(signs) == 0 {
					break
				}
				length := DecodeLength(&signs)
				//fmt.Println("length", length)
				//fmt.Printf("signs %x", signs)
				signsSlice = append(signsSlice, BytesShift(&signs, length))
			}
			if len(publicKeys) != len(signsSlice) {
				log.Debug("signsSlice", signsSlice)
				log.Debug("publicKeys", publicKeys)
				return false, fmt.Errorf("sign error %d!=%d", len(publicKeys), len(signsSlice))
			}
		}

		for i := 0; i < len(publicKeys); i++ {
			pub, err := BinToRsaPubKey(publicKeys[i])
			if err != nil {
				return false, ErrInfo(err)
			}
			err = rsa.VerifyPKCS1v15(pub, crypto.SHA1, HashSha1(forSign), signsSlice[i])
			if err != nil {
				log.Error("pub %v", pub)
				log.Error("publicKeys[i] %x", publicKeys[i])
				log.Error("crypto.SHA1", crypto.SHA1)
				log.Error("HashSha1(forSign)", HashSha1(forSign))
				log.Error("HashSha1(forSign)", string(HashSha1(forSign)))
				log.Error("forSign", forSign)
				log.Error("sign: %x\n", signsSlice[i])
				return false, ErrInfoFmt("incorrect sign:  hash = %x; forSign = %v, publicKeys[i] = %x, sign = %x", HashSha1(forSign), forSign, publicKeys[i], signsSlice[i])
			}
		}
		return true, nil*/
	return CheckECDSA(publicKeys, forSign, signs, nodeKeyOrLogin)
}

func SignECDSA(privateKey string, forSign string) (ret []byte, err error) {
	pubkeyCurve := elliptic.P256()

	b, err := hex.DecodeString(privateKey)
	if err != nil {
		log.Error("SignECDSA 0 %v", err)
		return
	}
	bi := new(big.Int).SetBytes(b)
	priv := new(ecdsa.PrivateKey)
	priv.PublicKey.Curve = pubkeyCurve
	priv.D = bi
	priv.PublicKey.X, priv.PublicKey.Y = pubkeyCurve.ScalarBaseMult(bi.Bytes())

	signhash := sha256.Sum256([]byte(forSign))
	r, s, err := ecdsa.Sign(crand.Reader, priv, signhash[:])
	if err != nil {
		log.Error("SignECDSA 0 %v", err)
		return
	}
	ret = FillLeft(r.Bytes())
	ret = append(ret, FillLeft(s.Bytes())...)
	return
}

func ParseSign(sign string) (r, s *big.Int) {
	var off int
	if len(sign) > 128 {
		off = 8
		if sign[7] == '1' {
			off = 10
		}
	} else if len(sign) < 128 {
		return
	}
	all, err := hex.DecodeString(string(sign[off:]))
	if err != nil {
		return
	}
	r = new(big.Int).SetBytes(all[:32])
	s = new(big.Int).SetBytes(all[len(all)-32:])
	return
}

func ConvertJSSign(in string) string {
	if len(in) == 0 {
		return ``
	}
	r, s := ParseSign(in)
	return hex.EncodeToString(append(FillLeft(r.Bytes()), FillLeft(s.Bytes())...))
}

func CheckECDSA(publicKeys [][]byte, forSign string, signs []byte, nodeKeyOrLogin bool) (bool, error) {
	var signsSlice [][]byte
	// у нода всегда 1 подпись
	if nodeKeyOrLogin {
		signsSlice = append(signsSlice, signs)
	} else {

		log.Debug("signs %x", signs)
		// в 1 signs может быть от 1 до 3-х подписей
		for {
			if len(signs) == 0 {
				break
			}
			length := DecodeLength(&signs)
			log.Debug("length %d", length)
			signsSlice = append(signsSlice, BytesShift(&signs, length))
		}
		if len(publicKeys) != len(signsSlice) {
			return false, fmt.Errorf("sign error %d!=%d", len(publicKeys), len(signsSlice))
		}
	}
	log.Debug("publicKeys %v", publicKeys)
	pubkeyCurve := elliptic.P256()
	signhash := sha256.Sum256([]byte(forSign))

	for i := 0; i < len(publicKeys); i++ {
		/*public, err := hex.DecodeString(string(publicKeys[i]))
		if err != nil {
			return false, ErrInfo(err)
		}*/
		public := publicKeys[i]
		pubkey := new(ecdsa.PublicKey)
		pubkey.Curve = pubkeyCurve
		pubkey.X = new(big.Int).SetBytes(public[0:32])
		pubkey.Y = new(big.Int).SetBytes(public[32:])

		r, s := ParseSign(hex.EncodeToString(signsSlice[i]))
		verifystatus := ecdsa.Verify(pubkey, signhash[:], r, s)
		if !verifystatus {
			log.Error("Check sign: %i %s\n", i, signsSlice[i])
			return false, ErrInfoFmt("incorrect sign:  hash = %x; forSign = %v, publicKeys[i] = %x, sign = %x",
				signhash, forSign, publicKeys[i], signsSlice[i])
		}

	}
	return true, nil
}

func B54Decode(b54_ interface{}) string {
	var b54 string
	switch b54_.(type) {
	case string:
		b54 = b54_.(string)
	case []byte:
		b54 = string(b54_.([]byte))
	}
	return string(b58.Decode(b54))
}

func HashSha1(msg string) []byte {
	sh := crypto.SHA1.New()
	sh.Write([]byte(msg))
	hash := sh.Sum(nil)
	return hash
}

func HashSha1Hex(msg []byte) string {
	sh := crypto.SHA1.New()
	sh.Write(msg)
	hash := sh.Sum(nil)
	return string(BinToHex(hash))
}

func Md5(msg_ interface{}) []byte {
	var msg []byte
	switch msg_.(type) {
	case string:
		msg = []byte(msg_.(string))
	case []byte:
		msg = msg_.([]byte)
	}
	sh := crypto.MD5.New()
	sh.Write(msg)
	hash := sh.Sum(nil)
	return BinToHex(hash)
}

func DSha256(data_ interface{}) []byte {
	var data []byte
	switch data_.(type) {
	case string:
		data = []byte(data_.(string))
	case []byte:
		data = data_.([]byte)
	}
	sha256_ := sha256.New()
	sha256_.Write(data)
	hashSha256 := fmt.Sprintf("%x", sha256_.Sum(nil))
	sha256_ = sha256.New()
	sha256_.Write([]byte(hashSha256))
	return []byte(fmt.Sprintf("%x", sha256_.Sum(nil)))
}

func Sha256(data_ interface{}) []byte {
	var data []byte
	switch data_.(type) {
	case string:
		data = []byte(data_.(string))
	case []byte:
		data = data_.([]byte)
	}
	sha256_ := sha256.New()
	sha256_.Write(data)
	return []byte(fmt.Sprintf("%x", sha256_.Sum(nil)))
}

func DeleteHeader(binaryData []byte) []byte {
	/*
		TYPE (0-блок, 1-тр-я)     1
		BLOCK_ID   				       4
		TIME       					       4
		USER_ID                         5
		LEVEL                              1
		SIGN                               от 128 до 512 байт. Подпись от TYPE, BLOCK_ID, PREV_BLOCK_HASH, TIME, USER_ID, LEVEL, MRKL_ROOT
		Далее - тело блока (Тр-ии)
	*/
	BytesShift(&binaryData, 15)
	size := DecodeLength(&binaryData)
	BytesShift(&binaryData, size)
	return binaryData
}

func GetMrklroot(binaryData []byte, first bool) ([]byte, error) {

	var mrklSlice [][]byte
	var txSize int64
	// [error] парсим после вызова функции
	if len(binaryData) > 0 {
		for {
			// чтобы исключить атаку на переполнение памяти
			if !first {
				if txSize > consts.MAX_TX_SIZE {
					return nil, ErrInfoFmt("[error] MAX_TX_SIZE")
				}
			}
			txSize = DecodeLength(&binaryData)

			// отчекрыжим одну транзакцию от списка транзакций
			if txSize > 0 {
				transactionBinaryData := BytesShift(&binaryData, txSize)
				dSha256Hash := DSha256(transactionBinaryData)
				mrklSlice = append(mrklSlice, dSha256Hash)
				//if len(transactionBinaryData) > 500000 {
				//	ioutil.WriteFile(string(dSha256Hash)+"-"+Int64ToStr(txSize), transactionBinaryData, 0644)
				//}
			}

			// чтобы исключить атаку на переполнение памяти
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

func SliceReverse(s []int64) []int64 {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}

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
	result_ := result[int32(len(result)-1)]
	log.Debug("result_: %s", result_)
	return []byte(result_[0])
}

func DbClose(c *DCDB) {
	err := c.Close()
	if err != nil {
		log.Debug("%v", err)
	}
}

func MaxInMap(m map[int64]int64) (int64, int64) {
	var max int64
	var maxK int64
	for k, v := range m {
		if max == 0 {
			max = v
			maxK = k
		} else if v > max {
			max = v
			maxK = k
		}
	}
	return max, maxK
}

func arraySum(m []map[int64]int64) int64 {
	var sum int64
	for i := 0; i < len(m); i++ {
		for _, v := range m[i] {
			sum += v
		}
	}
	return sum
}

func MaxInSliceMap(m []map[int64]int64) (int64, int64) {
	var max int64
	var maxK int64
	for i := 0; i < len(m); i++ {
		for k, v := range m[i] {
			if max == 0 {
				max = v
				maxK = k
			} else if v > max {
				max = v
				maxK = k
			}
		}
	}
	return max, maxK
}

func TypesToIds(arr []string) []int64 {
	var result []int64
	for _, v := range arr {
		result = append(result, TypeInt(v))
	}
	return result
}

func TypeInt(txType string) int64 {
	for k, v := range consts.TxTypes {
		if v == txType {
			return int64(k)
		}
	}
	return 0
}

func IntSliceToStr(Int []int) []string {
	var result []string
	for _, v := range Int {
		result = append(result, IntToStr(v))
	}
	return result
}

func MakePrivateKey(key string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(key))
	if block == nil {
		return nil, errors.New("bad key data")
	}
	if got, want := block.Type, "RSA PRIVATE KEY"; got != want {
		return nil, errors.New("unknown key type " + got + ", want " + want)
	}
	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

func JoinInts(arr map[int]int, sep string) string {
	var arrStr []string
	for _, v := range arr {
		arrStr = append(arrStr, IntToStr(v))
	}
	return strings.Join(arrStr, sep)
}

func JoinInt64Slice(arr []int64, sep string) string {
	var arrStr []string
	for _, v := range arr {
		arrStr = append(arrStr, Int64ToStr(v))
	}
	return strings.Join(arrStr, sep)
}

func JoinIntsK(arr map[int]int, sep string) string {
	var arrStr []string
	for k, _ := range arr {
		arrStr = append(arrStr, IntToStr(k))
	}
	return strings.Join(arrStr, sep)
}
func JoinInts64(arr map[int64]int, sep string) string {
	var arrStr []string
	for k, _ := range arr {
		arrStr = append(arrStr, Int64ToStr(k))
	}
	return strings.Join(arrStr, sep)
}

func TimeLeft(sec int64, lang map[string]string) string {
	result := ""
	if sec > 0 {
		days := int64(math.Floor(float64(sec / 86400)))
		sec -= days * 86400
		result += fmt.Sprintf(`%d %s `, days, lang["time_days"])
	}
	if sec > 0 {
		hours := int64(math.Floor(float64(sec / 3600)))
		sec -= hours * 3600
		result += fmt.Sprintf(`%d %s `, hours, lang["time_hours"])
	}
	if sec > 0 {
		minutes := int64(math.Floor(float64(sec / 60)))
		sec -= minutes * 3600
		result += fmt.Sprintf(`%d %s `, minutes, lang["time_minutes"])
	}
	return result
}

func MakeLastTx(lastTx []map[string]string, lng map[string]string) (string, map[int64]int64) {
	pendingTx := make(map[int64]int64)
	result := `<h3>` + lng["transactions"] + `</h3><table class="table" style="width:500px;">`
	result += `<tr><th>` + lng["time"] + `</th><th>` + lng["result"] + `</th></tr>`
	for _, data := range lastTx {
		result += "<tr>"
		result += "<td class='unixtime'>" + data["time_int"] + "</td>"
		if StrToInt64(data["block_id"]) > 0 {
			result += "<td>" + lng["in_the_block"] + " " + data["block_id"] + "</td>"
		} else if len(data["error"]) > 0 {
			result += "<td>Error: " + data["error"] + "</td>"
		} else if (len(data["queue_tx"]) == 0 && len(data["tx"]) == 0) || time.Now().Unix()-StrToInt64(data["time_int"]) > 7200 {
			result += "<td>" + lng["lost"] + "</td>"
		} else {
			result += "<td>" + lng["status_pending"] + "</td>"
			pendingTx[StrToInt64(data["type"])] = 1
		}
		result += "</tr>"
	}
	result += "</table>"
	return result, pendingTx
}

func Encrypt(key, text []byte) ([]byte, error) {
	iv := []byte(RandSeq(aes.BlockSize))
	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, ErrInfo(err)
	}
	plaintext := PKCS5Padding([]byte(text), c.BlockSize())
	cfbdec := cipher.NewCBCEncrypter(c, iv)
	EncPrivateKeyBin := make([]byte, len(plaintext))
	cfbdec.CryptBlocks(EncPrivateKeyBin, plaintext)
	EncPrivateKeyBin = append(iv, EncPrivateKeyBin...)
	//EncPrivateKeyB64 := base64.StdEncoding.EncodeToString(EncPrivateKeyBin)
	return EncPrivateKeyBin, nil
}

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

func EncryptData(data, publicKey []byte, randcandidateBlockHash string) ([]byte, []byte, []byte, error) {

	// генерим ключ
	key := Md5(DSha256([]byte(RandSeq(32) + randcandidateBlockHash)))

	// шифруем ключ публичным ключем получателя
	pub, err := BinToRsaPubKey(publicKey)
	if err != nil {
		return nil, nil, nil, ErrInfo(err)
	}
	encKey, err := rsa.EncryptPKCS1v15(crand.Reader, pub, key)
	if err != nil {
		return nil, nil, nil, ErrInfo(err)
	}

	// шифруем сам блок/тр-ии. Вначале encData добавляется IV
	encData, iv, err := EncryptCFB(data, key, []byte(""))
	if err != nil {
		return nil, nil, nil, ErrInfo(err)
	}
	log.Debug("encData %x", encData)

	// возвращаем ключ + IV + encData
	return append(EncodeLengthPlusData(encKey), encData...), key, iv, nil
}

func strpad(text string) string {
	length := aes.BlockSize - (len(text) % aes.BlockSize)
	for i := 0; i < length; i++ {
		text += "0"
	}
	return text
}

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
}

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

func TcpConn(Addr string) (net.Conn, error) {
	// шлем данные указанному хосту
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

func ProtectedCheckRemoteAddrAndGetHost(binaryData *[]byte, conn net.Conn) (string, error) {
	if ok, _ := regexp.MatchString(`^192\.168`, conn.RemoteAddr().String()); !ok {
		return "", ErrInfo("not local")
	}
	size := DecodeLength(&*binaryData)
	if int64(len(*binaryData)) < size {
		return "", ErrInfo("int64(len(binaryData)) < size")
	}
	host := string(BytesShift(&*binaryData, size))
	if ok, _ := regexp.MatchString(`^(?i)[0-9a-z\_\.\-\]{1,100}:[0-9]+$`, host); !ok {
		return "", ErrInfo("incorrect host " + host)
	}
	return host, nil

}

func WriteSizeAndData(binaryData []byte, conn net.Conn) error {
	// в 4-х байтах пишем размер данных, которые пошлем далее
	size := DecToBin(len(binaryData), 4)
	fmt.Println("len(binaryData)", len(binaryData))
	_, err := conn.Write(size)
	if err != nil {
		return ErrInfo(err)
	}
	// далее шлем сами данные
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

func WriteSizeAndDataTCPConn(binaryData []byte, conn net.Conn) error {
	// в 4-х байтах пишем размер данных, которые пошлем далее
	size := DecToBin(len(binaryData), 4)
	_, err := conn.Write(size)
	if err != nil {
		return ErrInfo(err)
	}
	// далее шлем сами данные
	if len(binaryData) > 0 {
		_, err = conn.Write(binaryData)
		if err != nil {
			return ErrInfo(err)
		}
	}
	return nil
}

func GetCurrentDir() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return "."
	}
	return dir
}

func GetBlockBody(host string, blockId int64, dataTypeBlockBody int64) ([]byte, error) {

	conn, err := TcpConn(host)
	if err != nil {
		return nil, ErrInfo(err)
	}
	defer conn.Close()

	log.Debug("dataTypeBlockBody: %v", dataTypeBlockBody)
	// шлем тип данных
	_, err = conn.Write(DecToBin(dataTypeBlockBody, 2))
	if err != nil {
		return nil, ErrInfo(err)
	}

	log.Debug("blockId: %v", blockId)

	// шлем номер блока
	_, err = conn.Write(DecToBin(blockId, 4))
	if err != nil {
		return nil, ErrInfo(err)
	}

	// в ответ получаем размер данных, которые нам хочет передать сервер
	buf := make([]byte, 4)
	n, err := conn.Read(buf)
	if err != nil {
		return nil, ErrInfo(err)
	}
	log.Debug("dataSize buf: %x / get: %v", buf, n)

	// и если данных менее 10мб, то получаем их
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

type jsonAnswer struct {
	err error
}

func (r *jsonAnswer) String() string {
	return fmt.Sprintf("%s", r.err)
}
func (r *jsonAnswer) Error() error {
	return r.err
}
func JsonAnswer(err interface{}, answType string) *jsonAnswer {
	var error_ string
	switch err.(type) {
	case string:
		error_ = err.(string)
	case error:
		error_ = fmt.Sprintf("%v", err)
	}
	result, _ := json.Marshal(map[string]string{answType: fmt.Sprintf("%v", error_)})
	return &jsonAnswer{errors.New(string(result))}
}

func TCPGetSizeAndData(conn net.Conn, maxSize int64) ([]byte, error) {
	// получаем размер данных
	buf := make([]byte, 4)
	_, err := conn.Read(buf)
	if err != nil {
		return nil, ErrInfo(err)
	}
	size := BinToDec(buf)
	fmt.Println("size: ", size)

	// получаем сами данные
	if size > maxSize || size == 0 {
		return nil, ErrInfo("incorrect size")
	}
	binaryData := make([]byte, size)
	_, err = io.ReadFull(conn, binaryData)
	if err != nil {
		return nil, ErrInfo("incorrect binaryData")
	}
	return binaryData, nil
}

func ClearNullFloat64(number float64, n int) float64 {
	return StrToFloat64(ClearNull(Float64ToStr(number), n))
}

func ClearNull(str string, n int) string {
	//str := Float64ToStr(num)
	ind := strings.Index(str, ".")
	new := ""
	if ind != -1 {
		end := n
		if len(str[ind+1:]) > 1 {
			end = n + 1
		}
		if n > 0 {
			new = str[:ind] + "." + str[ind+1:ind+end]
		} else {
			new = str[:ind]
		}
	} else {
		new = str
	}
	return new
}

func WriteSelectiveLog(text interface{}) {
	if *LogLevel == "DEBUG" {
		var text_ string
		switch text.(type) {
		case string:
			text_ = text.(string)
		case []byte:
			text_ = string(text.([]byte))
		case error:
			text_ = fmt.Sprintf("%v", text)
		}
		allTransactionsStr := ""
		allTransactions, _ := DB.GetAll("SELECT hex(hash) as hex_hash, verified, used, high_rate, for_self_use, user_id, third_var, counter, sent FROM transactions", 100)
		for _, data := range allTransactions {
			allTransactionsStr += data["hex_hash"] + "|" + data["verified"] + "|" + data["used"] + "|" + data["high_rate"] + "|" + data["for_self_use"] + "|" + consts.TxTypes[StrToInt(data["type"])] + "|" + data["user_id"] + "|" + data["third_var"] + "|" + data["counter"] + "|" + data["sent"] + "\n"
		}
		t := time.Now()
		data := allTransactionsStr + GetParent() + " ### " + t.Format(time.StampMicro) + " ### " + text_ + "\n\n"
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

func IPwoPort(ipport string) string {
	r, _ := regexp.Compile(`^([0-9\.]+)`)
	match := r.FindStringSubmatch(ipport)
	if len(match) == 0 {
		return ""
	}
	return match[1]
}

func daylightUpd(url string) error {
	zipfile := filepath.Join(*Dir, "dc.zip")
	_, err := DownloadToFile(url, zipfile, 3600, nil, nil, "upd")
	if err != nil {
		return ErrInfo(err)
	}
	fmt.Println(zipfile)
	reader, err := zip.OpenReader(zipfile)
	if err != nil {
		return ErrInfo(err)
	}
	appname := filepath.Base(os.Args[0])
	tmpname := filepath.Join(*Dir, `tmp_`+appname)

	f_ := reader.Reader.File
	f := f_[0]
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

	pwd, err := os.Getwd()
	if err != nil {
		return ErrInfo(err)
	}
	fmt.Print(pwd)

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
	log.Debug(tmpname, "-oldFileName", old, "-dir", *Dir, "-oldVersion", consts.VERSION)
	err = exec.Command(tmpname, "-oldFileName", old, "-dir", *Dir, "-oldVersion", consts.VERSION).Start()
	if err != nil {
		return ErrInfo(err)
	}
	return nil
}

func GetUpdVerAndUrl(host string) (string, string, error) {

	update, err := GetHttpTextAnswer(host + "/update.json")
	if len(update) > 0 {

		updateData := new(updateType)
		err = json.Unmarshal([]byte(update), &updateData)
		if err != nil {
			return "", "", ErrInfo(err)
		}

		//fmt.Println(updateData)

		dataJson, err := json.Marshal(updateData.Message)
		if err != nil {
			return "", "", ErrInfo(err)
		}

		pub, err := BinToRsaPubKey(HexToBin(consts.ALERT_KEY))
		if err != nil {
			return "", "", ErrInfo(err)
		}
		//fmt.Println(updateData.Signature)
		//fmt.Println(string(dataJson))
		err = rsa.VerifyPKCS1v15(pub, crypto.SHA1, HashSha1(string(dataJson)), []byte(HexToBin(updateData.Signature)))
		if err != nil {
			return "", "", ErrInfo(err)
		}

		//fmt.Println(runtime.GOOS+"_"+runtime.GOARCH)
		//fmt.Println(updateData.Message)
		//fmt.Println(updateData.Message[runtime.GOOS+"_"+runtime.GOARCH])
		//fmt.Println(updateData.Message["version"], consts.VERSION)
		if len(updateData.Message[runtime.GOOS+"_"+runtime.GOARCH]) > 0 && version.Compare(updateData.Message["version"], consts.VERSION, ">") {
			return updateData.Message["version"], updateData.Message[runtime.GOOS+"_"+runtime.GOARCH], nil
		}
	}
	return "", "", nil
}

type updateType struct {
	Message   map[string]string
	Signature string
}

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

// temporary

func EncodeLength(length int64) []byte {
	return lib.EncodeLength(length)
}

func DecodeLength(buf *[]byte) (ret int64) {
	ret, _ = lib.DecodeLength(buf)
	return
}

func DecodeLenInt64(data *[]byte) (int64, error) {
	return lib.DecodeLenInt64(data)
}

func FillLeft(slice []byte) []byte {
	return lib.FillLeft(slice)
}

func CreateHtmlFromTemplate(page string, citizenId, accountId, stateId int64) (string, error) {
	data, err := DB.Single(`SELECT value FROM "`+Int64ToStr(stateId)+`_pages" WHERE name = ?`, page).String()
	if err != nil {
		return "", err
	}

	qrx := regexp.MustCompile(`CitizenId`)
	data = qrx.ReplaceAllString(data, Int64ToStr(citizenId))
	qrx = regexp.MustCompile(`AccountId`)
	data = qrx.ReplaceAllString(data, Int64ToStr(accountId))

	qrx = regexp.MustCompile(`(?is).*\{\{table\.([\w\d_]*)\[([^\].]*)\]\.([\w\d_]*)\}\}.*`)
	sql := qrx.ReplaceAllString(data, `SELECT $3 FROM "$1" WHERE $2`)
	singleData, err := DB.Single(sql).String()
	if err != nil {
		log.Error("%v", err)
	}
	qrx = regexp.MustCompile(`(?is)\{\{table\.([\w\d_]*)\[[^\].]*\]\.[\w\d_]*\}\}`)
	data = qrx.ReplaceAllString(data, singleData)

	sys_navigate := func(row map[string]string, nav []string) string {
		if len(nav) != 3 {
			return fmt.Sprintf(`<td>%s</td><td>%s</td>`, row[`id`], strings.Join(nav, `,`))
		}
		pars := strings.Split(nav[2], `&`)
		parsout := make([]string, 0)
		for _, ipar := range pars {
			lr := strings.Split(ipar, `=`)
			if len(lr) == 2 {
				value := lr[1]
				if val, ok := row[value]; ok {
					value = val
				}
				parsout = append(parsout, lr[0]+`:`+value)
			}
		}
		return fmt.Sprintf(`<td>%s</td><td><a href="#" onclick="load_page('%s', {%s} )">%s</a></td>`,
			row[`id`], nav[1], strings.Join(parsout, `,`), nav[0])
	}
	navigate := func(row map[string]string, nav []string) string {
		if len(nav) != 2 {
			return fmt.Sprintf(`<td>%s</td><td>%s</td>`, row[`id`], strings.Join(nav, `,`))
		}
		value := nav[1]
		if val, ok := row[value]; ok {
			value = val
		}
		return fmt.Sprintf(`<td>%s</td><td><a href="#" onclick="load_template('%s')">%s</a></td>`,
			row[`id`], value, nav[0])
	}

	qrx = regexp.MustCompile(`(?is).*\{\{table\.([\w\d_]*)\.\(([\w\d_\,]*)\)\.([\w\d_]*)\(([\w\d_\s=\,)]*)\)\}\}.*`)
	sql = qrx.ReplaceAllString(data, `SELECT $2 FROM "$1"|$3|$4`)
	pars := strings.Split(sql, `|`)
	if len(pars) == 3 {
		nav := strings.Split(pars[2], `,`)
		dataTable, err := DB.GetAll(pars[0], 1000)
		if err != nil {
			log.Error("%v", err)
		}
		table := `<table  class="table table-striped table-bordered table-hover">`
		for _, row := range dataTable {
			table += `<tr>`
			switch pars[1] {
			case `sys_navigate`:
				table += sys_navigate(row, nav)
			case `navigate`:
				table += navigate(row, nav)
			}
			table += `</tr>`
		}
		table += `</table>`
		qrx = regexp.MustCompile(`(?is)\{\{table\.([\w\d_]*)\.\(([\w\d_\,]*)\)\.([\w\d_]*)\(([\w\d_\s=\,)]*)\)\}\}`)
		data = qrx.ReplaceAllString(data, table)
	}
	qrx = regexp.MustCompile(`(?is).*\{\{table\.([\w\d_]*)\}\}.*`)
	sql = qrx.ReplaceAllString(data, `SELECT * FROM "$1"`)
	dataTable, err := DB.GetAll(sql, 1000)
	if err != nil {
		log.Error("%v", err)
	}
	table := `<table class="table table-striped table-bordered table-hover">`
	for _, row := range dataTable {
		table += `<tr>`
		for _, cell := range row {
			table += `<td>` + cell + `</td>`
		}
		table += `</tr>`
	}
	table += `</table>`

	qrx = regexp.MustCompile(`(?is)\{\{table\.([\w\d_]*)\}\}`)
	data = qrx.ReplaceAllString(data, table)

	qrx = regexp.MustCompile(`(?is)\{\{contract\.([\w\d_]*)\}\}`)
	data = qrx.ReplaceAllStringFunc(data, func(match string) string {
		name := match[strings.Index(match, `.`)+1 : len(match)-2]
		contract := smart.GetContract(name)
		out := `<form role="form">` + name + `<div id="fields">`
		if contract == nil || contract.Block.Info.(*script.ContractInfo).Tx == nil {
			err = fmt.Errorf(`there is not %s contract or parameters`, name)
		} else {
			for _, fitem := range *(*contract).Block.Info.(*script.ContractInfo).Tx {
				if fitem.Type.String() == `string` {
					out += fmt.Sprintf(`<div class="form-group"><label for="%s">%s</label>
					<input id="%s" name="%s" type="text" class="form-control"></div>`,
						fitem.Name, fitem.Name, fitem.Name)
				}
			}
		}
		fmt.Println(`Name Contract`, name, contract)
		return out + `</div></form>`
	})

	qrx = regexp.MustCompile(`(?is)\{\{table\.([\w\d_]*)\}\}`)
	data = qrx.ReplaceAllString(data, table)

	qrx = regexp.MustCompile(`(?is)\[([\w\s]*)\]\(([\w\s]*)\)`)
	data = qrx.ReplaceAllString(data, "<li><a href='#' onclick=\"load_template('$2'); HideMenu();\"><span>$1</span></a></li>")
	qrx = regexp.MustCompile(`(?is)\[([\w\s]*)\]\(sys.([\w\s]*)\)`)
	data = qrx.ReplaceAllString(data, "<li><a href='#' onclick=\"load_page('$2'); HideMenu();\"><span>$1</span></a></li>")

	qrx = regexp.MustCompile(`(?is)\{\{Title=([\w\s]+)\}\}`)
	data = qrx.ReplaceAllString(data, `<div class="content-heading">$1</div>`)
	qrx = regexp.MustCompile(`(?is)\{\{Navigation=(.*?)\}\}`)
	data = qrx.ReplaceAllString(data, `<ol class="breadcrumb"><span class="pull-right"><a href='#' onclick="load_page('editPage', {name: '`+page+`'} )">Edit</a></span>$1</ol>`)
	qrx = regexp.MustCompile(`(?is)\{\{PageTitle=([\w\s]+)\}\}`)
	data = qrx.ReplaceAllString(data, `<div class="panel panel-default"><div class="panel-heading"><div class="panel-title">$1</div></div><div class="panel-body">`)

	unsafe := string(blackfriday.MarkdownCommon([]byte(data)))
	//html := string(bluemonday.UGCPolicy().SanitizeBytes(unsafe))

	// removing <p></p>
	//unsafe = unsafe[3 : len(unsafe)-5]
	return unsafe, nil
}
