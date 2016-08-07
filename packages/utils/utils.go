package utils

import (
	"archive/zip"
	"bytes"
	"crypto"
	"crypto/aes"
	"crypto/cipher"
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
	"github.com/DayLightProject/go-daylight/packages/consts"
	"github.com/DayLightProject/go-daylight/packages/static"
	"github.com/golang/freetype"
	"github.com/kardianos/osext"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"github.com/mcuadros/go-version"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"io/ioutil"
	"math"
	"math/big"
	"math/rand"
	"net"
	"net/http"
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
)

type BlockData struct {
	BlockId       int64
	Time          int64
	WalletId      int64
	CBID          int64
	CurrentUserId int64
	Sign          []byte
	Hash          []byte
	HeadHash      []byte
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
	ChAnswer chan string
}
var (
	GenerateFirstBlock = flag.Int64("generateFirstBlock", 0, "generateFirstBlock")
	OldVersion = flag.String("oldVersion", "", "")
	TestRollBack = flag.Int64("testRollBack", 0, "testRollBack")
	Dir = flag.String("dir", GetCurrentDir(), "Dcoin directory")
	OldFileName = flag.String("oldFileName", "", "")
	LogLevel = flag.String("logLevel", "", "Dcoin LogLevel")
	Console = flag.Int64("console", 0, "Start from console")
	SqliteDbUrl string
	StartBlockId = flag.Int64("startBlockId", 0, "Start block for blockCollection daemon")
	EndBlockId = flag.Int64("endBlockId", 0, "End block for blockCollection daemon")
	RollbackToBlockId = flag.Int64("rollbackToBlockId", 0, "Rollback to block_id")
	ListenHttpHost = flag.String("listenHttpHost", "8089", "ListenHttpHost")
	DaemonsChans []*DaemonsChansType
	eWallets = &sync.Mutex{}
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
	if _, err := os.Stat(*Dir+"/"+thrust_shell); err == nil {
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

func KeyToImg(key, resultPath string, userId int64, timeFormat string, param ParamType) (*bytes.Buffer, error) {

	keyBin, _ := base64.StdEncoding.DecodeString(key)
	keyHex := append(BinToHex(keyBin), []byte("00000000")...)
	keyBin = HexToBin(keyHex)

	w, h := getImageDimension(param.Bg_path)
	/*fSrc, err := os.Open(param.Bg_path)
	if err != nil {
		return nil, ErrInfo(err)
	}
	defer fSrc.Close()*/
	data, _ := static.Asset(param.Bg_path)
	fSrc := bytes.NewReader(data)

	src, err := png.Decode(fSrc)
	if err != nil {
		return nil, ErrInfo(err)
	}

	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	white := image.NewUniform(color.White)
	black := image.NewUniform(color.Black)
	draw.Draw(dst, dst.Bounds(), src, image.Point{0, 0}, draw.Src)

	x := param.X
	y := param.Y
	var color *image.Uniform
	for i := 0; i < len(keyBin); i++ {
		b := fmt.Sprintf("%08b ", keyBin[i])
		for j := 0; j < 8; j++ {
			if b[j:j+1] == "0" {
				color = black
			} else {
				color = white
			}
			dst.Set(int(x), int(y), color)
			x++
			if (x+1-param.X)%param.Width == 0 {
				x = param.X
				y++
			}
		}
	}
	param.Height = y - param.Y + 1

	x = 0
	// теперь пишем инфу, где искать квадрат
	info := fmt.Sprintf("%016s", strconv.FormatInt(param.X, 2)) + fmt.Sprintf("%016s", strconv.FormatInt(param.Y, 2)) + fmt.Sprintf("%016s", strconv.FormatInt(param.Width, 2)) + fmt.Sprintf("%016s", strconv.FormatInt(param.Height, 2))
	for i := 0; i < len(info); i++ {
		if info[i:i+1] == "0" {
			color = black
		} else {
			color = white
		}
		dst.Set(int(x), 0, color)
		x++
	}

	fontBytes, err := static.Asset("static/fonts/luxisr.ttf")
	if err != nil {
		return nil, ErrInfo(err)
	}
	font, err := freetype.ParseFont(fontBytes)
	if err != nil {
		return nil, ErrInfo(err)
	}

	c := freetype.NewContext()
	c.SetDPI(72)
	c.SetFont(font)
	c.SetFontSize(15)
	c.SetClip(dst.Bounds())
	c.SetDst(dst)
	c.SetSrc(image.Black)

	// Draw the text.
	pt := freetype.Pt(13, 300)
	_, err = c.DrawString("User ID: "+Int64ToStr(userId), pt)
	if err != nil {
		return nil, ErrInfo(err)
	}

	t := time.Unix(time.Now().Unix(), 0)
	txTime := t.Format(timeFormat)
	pt = freetype.Pt(300, 300)
	_, err = c.DrawString(txTime, pt)
	if err != nil {
		return nil, ErrInfo(err)
	}

	if len(resultPath) > 0 {
		fDst, err := os.Create(resultPath)
		if err != nil {
			return nil, ErrInfo(err)
		}
		defer fDst.Close()
		err = png.Encode(fDst, dst)
		if err != nil {
			return nil, ErrInfo(err)
		}
	}
	buffer := new(bytes.Buffer)
	err = png.Encode(buffer, dst)
	if err != nil {
		return nil, ErrInfo(err)
	}

	return buffer, nil
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
		CB_ID                              1
		SIGN                               от 128 до 512 байт. Подпись от TYPE, BLOCK_ID, PREV_BLOCK_HASH, TIME, WALLET_ID, CB_ID, MRKL_ROOT
		Далее - тело блока (Тр-ии)
	*/
	result.BlockId = BinToDecBytesShift(binaryBlock, 4)
	result.Time = BinToDecBytesShift(binaryBlock, 4)
	result.WalletId = BytesToInt64(BytesShift(binaryBlock, DecodeLength(binaryBlock)))
	result.CBID = BinToDecBytesShift(binaryBlock, 1)
	if result.BlockId > 1 {
		signSize := DecodeLength(binaryBlock)
		result.Sign = BytesShift(binaryBlock, signSize)
	}
	log.Debug("result: %v", result)
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

func CalcProfit_24946(amount float64, timeStart, timeFinish int64, pctArray []map[int64]map[string]float64, pointsStatusArray []map[int64]string, holidaysArray [][]int64, maxPromisedAmountArray []map[int64]string, currencyId int64, repaidAmount float64) (float64, error) {

	log.Debug("amount", amount)
	log.Debug("timeStart", timeStart)
	log.Debug("timeFinish", timeFinish)
	log.Debug("%v", pctArray)
	log.Debug("%v", pointsStatusArray)
	log.Debug("%v", holidaysArray)
	log.Debug("%v", maxPromisedAmountArray)
	log.Debug("currencyId", currencyId)
	log.Debug("repaidAmount", repaidAmount)

	var lastStatus string = ""
	var findMinArray []map[string]string
	var newArr []map[int64]float64
	var statusPctArray_ map[string]float64
	// нужно получить массив вида time=>pct, совместив $pct_array и $points_status_array

	findTime := func(key int64, arr []map[int64]float64) bool {
		for i := 0; i < len(arr); i++ {
			if arr[i][key] != 0 {
				return true
			}
		}
		return false
	}

	log.Debug("pctArray", pctArray)
	for i := 0; i < len(pctArray); i++ {
		for time, statusPctArray := range pctArray[i] {
			log.Debug("i=", i, "pctArray[i]=", pctArray[i])
			findMinArray, pointsStatusArray = findMinPointsStatus(time, pointsStatusArray, "status")
			//log.Debug("i", i)
			log.Debug("time", time)
			log.Debug("findMinArray", findMinArray)
			log.Debug("pointsStatusArray", pointsStatusArray)
			for j := 0; j < len(findMinArray); j++ {
				if StrToInt64(findMinArray[j]["time"]) < time {
					findMinPct := findMinPct_24946(StrToInt64(findMinArray[j]["time"]), pctArray, findMinArray[j]["status"])
					if !findTime(StrToInt64(findMinArray[j]["time"]), newArr) {
						newArr = append(newArr, map[int64]float64{StrToInt64(findMinArray[j]["time"]): findMinPct})
						log.Debug("findMinPct", findMinPct)
					}
					lastStatus = findMinArray[j]["status"]
				}
			}
			if len(findMinArray) == 0 && len(lastStatus) == 0 {
				findMinArray = append(findMinArray, map[string]string{"status": "user"})
			} else if len(findMinArray) == 0 && len(lastStatus) != 0 { // есть проценты, но кончились points_status
				findMinArray = append(findMinArray, map[string]string{"status": "miner"})
			}
			if !findTime(time, newArr) {
				newArr = append(newArr, map[int64]float64{time: statusPctArray[findMinArray[len(findMinArray)-1]["status"]]})
			}
			statusPctArray_ = statusPctArray
		}
	}

	// если в points больше чем в pct
	if len(pointsStatusArray) > 0 {
		for i := 0; i < len(pointsStatusArray); i++ {
			for time, status := range pointsStatusArray[i] {
				if !findTime(time, newArr) {
					newArr = append(newArr, map[int64]float64{time: statusPctArray_[status]})
				}
			}
		}
	}

	log.Debug("newArr", newArr)

	// newArr - массив, где ключи - это время из pct и points_status, а значения - проценты.

	// $max_promised_amount_array + $pct_array
	/*
		 * в $pct_array сейчас
				[1394308000] =>  0,05
				[1394308100] =>  0,1

			после обработки станет

				[1394308000] => Array
					(
						[pct] => 0,05
						[amount] => 1000
					)
				[1394308005] => Array
					(
						[pct] => 0,05
						[amount] => 100
					)
				[1394308100] => Array
					(
						[pct] => 0,1
						[amount] => 100
					)

		 * */

	findTime2 := func(key int64, arr []map[int64]pctAmount) bool {
		for i := 0; i < len(arr); i++ {
			if arr[i][key].pct != 0 {
				return true
			}
		}
		return false
	}

	var newArr2 []map[int64]pctAmount
	var lastAmount float64
	var amount_ float64
	var pct_ float64
	if len(maxPromisedAmountArray) == 0 {
		lastAmount = amount
	}

	// нужно получить массив вида time=>pct, совместив newArr и $max_promised_amount_array
	for i := 0; i < len(newArr); i++ {
		log.Debug("i ", i)
		for time, pct := range newArr[i] {
			findMinArray, maxPromisedAmountArray = findMinPointsStatus(time, maxPromisedAmountArray, "amount")
			for j := 0; j < len(findMinArray); j++ {
				if amount+repaidAmount > StrToFloat64(findMinArray[j]["amount"]) {
					amount_ = StrToFloat64(findMinArray[j]["amount"]) - repaidAmount
				} else if amount < StrToFloat64(findMinArray[j]["amount"]) && currencyId == 1 {
					amount_ = StrToFloat64(findMinArray[j]["amount"])
				} else {
					amount_ = amount
				}
				if StrToInt64(findMinArray[j]["time"]) <= time {
					minPct := findMinPct1_24946(StrToInt64(findMinArray[j]["time"]), newArr)
					if !findTime2(StrToInt64(findMinArray[j]["time"]), newArr2) {
						newArr2 = append(newArr2, map[int64]pctAmount{StrToInt64(findMinArray[j]["time"]): {pct: minPct, amount: amount_}})
					}
					lastAmount = amount_

				}
			}
			if !findTime2(time, newArr2) {
				newArr2 = append(newArr2, map[int64]pctAmount{time: {pct: pct, amount: lastAmount}})
			}
			pct_ = pct
		}
	}

	log.Debug("newArr2", newArr2)

	if !findTime2(timeFinish, newArr2) {
		newArr2 = append(newArr2, map[int64]pctAmount{timeFinish: {pct: pct_, amount: 0}})
	}

	var workTime, oldTime int64
	var resultArr []resultArrType
	var oldPctAndAmount pctAmount
	var startHolidays bool
	var finishHolidaysElement int64
	//START:
	for i := 0; i < len(newArr2); i++ {

		for time, pctAndAmount := range newArr2[i] {

			log.Debug("pctAndAmount", pctAndAmount)

			if time > timeStart {
				workTime = time
				for j := 0; j < len(holidaysArray); j++ {

					if holidaysArray[j][1] <= oldTime {
						continue
					}

					log.Debug("holidaysArray[j]", holidaysArray[j])

					// полные каникулы в промежутке между time и old_time
					if holidaysArray[j][0] != -1 && workTime >= holidaysArray[j][0] && holidaysArray[j][1] != -1 && workTime >= holidaysArray[j][1] {
						time = holidaysArray[j][0]
						holidaysArray[j][0] = -1
						resultArr = append(resultArr, resultArrType{num_sec: (time - oldTime), pct: oldPctAndAmount.pct, amount: oldPctAndAmount.amount})
						log.Debug("resultArr append")
						oldTime = holidaysArray[j][1]
						holidaysArray[j][1] = -1
					}
					if holidaysArray[j][0] != -1 && workTime >= holidaysArray[j][0] {
						startHolidays = true                        // есть начало каникул, но есть ли конец?
						finishHolidaysElement = holidaysArray[j][1] // для записи в лог
						time = holidaysArray[j][0]
						if time < timeStart {
							time = timeStart
						}
						holidaysArray[j][0] = -1
					} else if holidaysArray[j][1] != -1 && workTime < holidaysArray[j][1] && holidaysArray[j][0] == -1 {
						// конец каникул заканчивается после $work_time
						time = oldTime
						continue
					} else if holidaysArray[j][1] != -1 && workTime >= holidaysArray[j][1] {
						oldTime = holidaysArray[j][1]
						holidaysArray[j][1] = -1
						startHolidays = false // конец каникул есть
					} else if j == len(holidaysArray)-1 && !startHolidays {
						// если это последний полный внутрений холидей, то time должен быть равен текущему workTime
						time = workTime
					}
				}
				if time > timeFinish {
					time = timeFinish
				}
				resultArr = append(resultArr, resultArrType{num_sec: (time - oldTime), pct: oldPctAndAmount.pct, amount: oldPctAndAmount.amount})
				log.Debug("new", (time - oldTime))
				oldTime = time
			} else {
				oldTime = timeStart
			}
			oldPctAndAmount = pctAndAmount
		}
	}

	log.Debug("resultArr", resultArr)

	if startHolidays && finishHolidaysElement > 0 {
		log.Debug("finishHolidaysElement:", finishHolidaysElement)
	}

	// время в процентах меньше, чем нужное нам конечное время
	if oldTime < timeFinish && !startHolidays {
		// просто берем последний процент и добиваем его до нужного $time_finish
		sec := timeFinish - oldTime
		resultArr = append(resultArr, resultArrType{num_sec: sec, pct: oldPctAndAmount.pct, amount: oldPctAndAmount.amount})
	}

	var profit, amountAndProfit float64
	for i := 0; i < len(resultArr); i++ {
		pct := 1 + resultArr[i].pct
		num := resultArr[i].num_sec
		amountAndProfit = profit + resultArr[i].amount
		//$profit = ( floor( round( $amount_and_profit*pow($pct, $num), 3)*100 ) / 100 ) - $new[$i]['amount'];
		// из-за того, что в front был подсчет без обновления points, а в рабочем методе уже с обновлением points, выходило, что в рабочем методе было больше мелких временных промежуток, и получалось profit <0.01, из-за этого было расхождение в front и попадание минуса в БД
		profit = amountAndProfit*math.Pow(pct, float64(num)) - resultArr[i].amount
	}
	log.Debug("profit", profit)

	return profit, nil
}

// только для блоков до 24946
func findMinPct1_24946(needTime int64, pctArray []map[int64]float64) float64 {
	var findTime int64 = 0
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

// только для блоков до 24946
func findMinPct_24946(needTime int64, pctArray []map[int64]map[string]float64, status string) float64 {
	var findTime int64 = 0
	var pct float64 = 0
	log.Debug("pctArray findMinPct_24946", pctArray)
BR:
	for i := 0; i < len(pctArray); i++ {
		for time, _ := range pctArray[i] {
			log.Debug("%v", time, ">", needTime, "?")
			if time > needTime {
				log.Debug("break")
				break BR
			}
			findTime = int64(i)
		}
	}
	log.Debug("findTime", findTime)
	if findTime > 0 {
		for _, arr := range pctArray[findTime] {
			pct = arr[status]
		}
	}
	return pct
}
func CalcProfit(amount float64, timeStart, timeFinish int64, pctArray []map[int64]map[string]float64, pointsStatusArray []map[int64]string, holidaysArray [][]int64, maxPromisedAmountArray []map[int64]string, currencyId int64, repaidAmount float64) (float64, error) {

	log.Debug("CalcProfit")
	log.Debug("amount", amount)
	log.Debug("timeStart", timeStart)
	log.Debug("timeFinish", timeFinish)
	log.Debug("%v", pctArray)
	log.Debug("%v", pointsStatusArray)
	log.Debug("%v", holidaysArray)
	log.Debug("%v", maxPromisedAmountArray)
	log.Debug("currencyId", currencyId)
	log.Debug("repaidAmount", repaidAmount)

	if timeStart >= timeFinish {
		return 0, nil
	}

	// для WOC майнинг останавливается только если майнера забанил админ, каникулы на WOC не действуют
	if currencyId == 1 {
		holidaysArray = nil
	}

	/* $max_promised_amount_array имеет дефолтные значения от времени = 0
	 * $pct_array имеет дефолтные значения 0% для user/miner от времени = 0
	 * в $points_status_array крайний элемент массива всегда будет относиться к текущим 30-и дням т.к. перед calc_profit всегда идет вызов points_update
	 * */

	var lastStatus string = ""
	var findMinArray []map[string]string
	var newArr []map[int64]float64
	var statusPctArray_ map[string]float64
	// нужно получить массив вида time=>pct, совместив $pct_array и $points_status_array

	findTime := func(key int64, arr []map[int64]float64) bool {
		for i := 0; i < len(arr); i++ {
			if arr[i][key] != 0 {
				return true
			}
		}
		return false
	}

	log.Debug("pctArray", pctArray)
	for i := 0; i < len(pctArray); i++ {
		for time, statusPctArray := range pctArray[i] {
			log.Debug("i=", i, "pctArray[i]=", pctArray[i])
			findMinArray, pointsStatusArray = findMinPointsStatus(time, pointsStatusArray, "status")
			//log.Debug("i", i)
			log.Debug("time", time)
			log.Debug("findMinArray", findMinArray)
			log.Debug("pointsStatusArray", pointsStatusArray)
			for j := 0; j < len(findMinArray); j++ {
				if StrToInt64(findMinArray[j]["time"]) <= time {
					findMinPct := findMinPct(StrToInt64(findMinArray[j]["time"]), pctArray, findMinArray[j]["status"])
					if !findTime(StrToInt64(findMinArray[j]["time"]), newArr) {
						newArr = append(newArr, map[int64]float64{StrToInt64(findMinArray[j]["time"]): findMinPct})
						log.Debug("findMinPct", findMinPct)
					}
					lastStatus = findMinArray[j]["status"]
				}
			}
			if len(findMinArray) == 0 && len(lastStatus) == 0 {
				findMinArray = append(findMinArray, map[string]string{"status": "user"})
			} else if len(findMinArray) == 0 && len(lastStatus) != 0 { // есть проценты, но кончились points_status
				findMinArray = append(findMinArray, map[string]string{"status": lastStatus})
			}
			if !findTime(time, newArr) {
				newArr = append(newArr, map[int64]float64{time: statusPctArray[findMinArray[len(findMinArray)-1]["status"]]})
			}
			statusPctArray_ = statusPctArray
		}
	}

	// если в points больше чем в pct
	if len(pointsStatusArray) > 0 {
		for i := 0; i < len(pointsStatusArray); i++ {
			for time, status := range pointsStatusArray[i] {
				if !findTime(time, newArr) {
					newArr = append(newArr, map[int64]float64{time: statusPctArray_[status]})
				}
			}
		}
	}

	log.Debug("newArr", newArr)

	// newArr - массив, где ключи - это время из pct и points_status, а значения - проценты.

	// $max_promised_amount_array + $pct_array
	/*
		 * в $pct_array сейчас
				[1394308000] =>  0,05
				[1394308100] =>  0,1

			после обработки станет

				[1394308000] => Array
					(
						[pct] => 0,05
						[amount] => 1000
					)
				[1394308005] => Array
					(
						[pct] => 0,05
						[amount] => 100
					)
				[1394308100] => Array
					(
						[pct] => 0,1
						[amount] => 100
					)

		 * */

	findTime2 := func(key int64, arr []map[int64]pctAmount) bool {
		for i := 0; i < len(arr); i++ {
			if arr[i][key].pct != 0 {
				return true
			}
		}
		return false
	}

	var newArr2 []map[int64]pctAmount
	var lastAmount float64
	var amount_ float64
	var pct_ float64
	if len(maxPromisedAmountArray) == 0 {
		lastAmount = amount
	}

	log.Debug("newArr201", newArr)

	// нужно получить массив вида time=>pct, совместив newArr и $max_promised_amount_array
	for i := 0; i < len(newArr); i++ {
		log.Debug("i ", i)
		for time, pct := range newArr[i] {
			findMinArray, maxPromisedAmountArray = findMinPointsStatus(time, maxPromisedAmountArray, "amount")
			for j := 0; j < len(findMinArray); j++ {
				amount_ = getMaxPromisedAmountCalcProfit(amount, repaidAmount, StrToFloat64(findMinArray[j]["amount"]), currencyId)
				if StrToInt64(findMinArray[j]["time"]) <= time {
					minPct := findMinPct1(StrToInt64(findMinArray[j]["time"]), newArr)
					if !findTime2(StrToInt64(findMinArray[j]["time"]), newArr2) {
						newArr2 = append(newArr2, map[int64]pctAmount{StrToInt64(findMinArray[j]["time"]): {pct: minPct, amount: amount_}})
					}
					lastAmount = amount_

				}
			}
			if !findTime2(time, newArr2) {
				newArr2 = append(newArr2, map[int64]pctAmount{time: {pct: pct, amount: lastAmount}})
				log.Debug("findTime2", time, pct)
			}
			pct_ = pct
		}
	}

	log.Debug("newArr21", newArr2)

	// если в max_promised_amount больше чем в pct
	if len(maxPromisedAmountArray) > 0 {
		log.Debug("maxPromisedAmountArray", maxPromisedAmountArray)

		for i := 0; i < len(maxPromisedAmountArray); i++ {
			for time, maxPromisedAmount := range maxPromisedAmountArray[i] {
				MaxPromisedAmountCalcProfit := getMaxPromisedAmountCalcProfit(amount, repaidAmount, StrToFloat64(maxPromisedAmount), currencyId)
				amount_ = MaxPromisedAmountCalcProfit
				if !findTime2(time, newArr2) {
					newArr2 = append(newArr2, map[int64]pctAmount{time: {pct: pct_, amount: MaxPromisedAmountCalcProfit}})
				}
			}
		}
	}

	maxTimeInNewArr2 := func(newArr2 []map[int64]pctAmount) int64 {
		var max int64
		for i := 0; i < len(newArr2); i++ {
			for time, _ := range newArr2[i] {
				if time > max {
					max = time
				}
			}
		}
		return max
	}

	if maxTimeInNewArr2(newArr2) < timeFinish {
		// добавим сразу время окончания
		//newArr2[timeFinish] = pct;
		if !findTime2(timeFinish, newArr2) {
			newArr2 = append(newArr2, map[int64]pctAmount{timeFinish: {pct: pct_, amount: 0}})
		}
	}

	var workTime, oldTime int64
	var resultArr []resultArrType
	var oldPctAndAmount pctAmount
	var startHolidays bool
	var finishHolidaysElement int64
	log.Debug("newArr2", newArr2)
START:
	for i := 0; i < len(newArr2); i++ {

		for time, pctAndAmount := range newArr2[i] {

			log.Debug("%v", time, timeFinish)
			log.Debug("pctAndAmount", pctAndAmount)
			if time > timeFinish {
				log.Debug("continue START", time, timeFinish)
				continue START
			}
			if time > timeStart {
				workTime = time
				for j := 0; j < len(holidaysArray); j++ {

					if holidaysArray[j][1] <= oldTime {
						continue
					}

					log.Debug("holidaysArray[j]", holidaysArray[j])

					// полные каникулы в промежутке между time и old_time
					if holidaysArray[j][0] != -1 && oldTime <= holidaysArray[j][0] && holidaysArray[j][1] != -1 && workTime >= holidaysArray[j][1] {
						time = holidaysArray[j][0]
						holidaysArray[j][0] = -1
						resultArr = append(resultArr, resultArrType{num_sec: (time - oldTime), pct: oldPctAndAmount.pct, amount: oldPctAndAmount.amount})
						log.Debug("resultArr append")
						oldTime = holidaysArray[j][1]
						holidaysArray[j][1] = -1
					}
					if holidaysArray[j][0] != -1 && workTime >= holidaysArray[j][0] {
						startHolidays = true                        // есть начало каникул, но есть ли конец?
						finishHolidaysElement = holidaysArray[j][1] // для записи в лог
						time = holidaysArray[j][0]
						if time < timeStart {
							time = timeStart
						}
						holidaysArray[j][0] = -1
					} else if holidaysArray[j][1] != -1 && workTime < holidaysArray[j][1] && holidaysArray[j][0] == -1 {
						// конец каникул заканчивается после $work_time
						time = oldTime
						continue
					} else if holidaysArray[j][1] != -1 && workTime >= holidaysArray[j][1] {
						oldTime = holidaysArray[j][1]
						holidaysArray[j][1] = -1
						startHolidays = false // конец каникул есть
					} else if j == len(holidaysArray)-1 && !startHolidays {
						// если это последний полный внутрений холидей, то time должен быть равен текущему workTime
						time = workTime
					}
				}
				if time > timeFinish {
					time = timeFinish
				}
				resultArr = append(resultArr, resultArrType{num_sec: (time - oldTime), pct: oldPctAndAmount.pct, amount: oldPctAndAmount.amount})
				log.Debug("new", (time - oldTime))
				oldTime = time
			} else {
				oldTime = timeStart
			}
			oldPctAndAmount = pctAndAmount
			log.Debug("oldPctAndAmount", oldPctAndAmount)
		}
	}
	log.Debug("oldTime", oldTime)
	log.Debug("timeFinish", timeFinish)

	log.Debug("resultArr", resultArr)

	if startHolidays && finishHolidaysElement > 0 {
		log.Debug("finishHolidaysElement:", finishHolidaysElement)
	}

	// время в процентах меньше, чем нужное нам конечное время
	if oldTime < timeFinish && !startHolidays {
		log.Debug("oldTime < timeFinish")
		// просто берем последний процент и добиваем его до нужного $time_finish
		sec := timeFinish - oldTime
		resultArr = append(resultArr, resultArrType{num_sec: sec, pct: oldPctAndAmount.pct, amount: oldPctAndAmount.amount})
	}

	log.Debug("resultArr", resultArr)

	var profit, amountAndProfit float64
	for i := 0; i < len(resultArr); i++ {
		pct := 1 + resultArr[i].pct
		num := resultArr[i].num_sec
		amountAndProfit = profit + resultArr[i].amount
		//$profit = ( floor( round( $amount_and_profit*pow($pct, $num), 3)*100 ) / 100 ) - $new[$i]['amount'];
		// из-за того, что в front был подсчет без обновления points, а в рабочем методе уже с обновлением points, выходило, что в рабочем методе было больше мелких временных промежуток, и получалось profit <0.01, из-за этого было расхождение в front и попадание минуса в БД
		profit = amountAndProfit*math.Pow(pct, float64(num)) - resultArr[i].amount
	}

	log.Debug("total profit w/o amount:", profit)

	return profit, nil
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
		if ok, _ := regexp.MatchString(`^(?i)[0-9a-z\+\-\s\=\/\n\r]+$`, data); ok {
			if len(data) > 256 && len(data) < 3072 {
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
			if StrToInt(data) <= 15 && StrToInt(data) >=5 {
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
	case "tcp_host":
		if ok, _ := regexp.MatchString(`^(?i)[0-9a-z\_\.\-\]{1,100}:[0-9]+$`, data); ok {
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

func SendSms(sms_http_get_request, text string) (string, error) {
	html, err := GetHttpTextAnswer(sms_http_get_request + text)
	if err != nil {
		result, _ := json.Marshal(map[string]string{"error": fmt.Sprintf(`%s`, err)})
		return string(result), nil
	}
	result, _ := json.Marshal(map[string]string{"success": html})
	return string(result), nil
}
/*
func sendMail(body, subj string, To string, mailData map[string]string) error {

	smtpHostAndPort := mailData["smtp_server"] + ":" + mailData["smtp_port"]

	from := mail.Address{"", mailData["smtp_username"]}
	to := mail.Address{"", To}

	// Setup headers
	headers := make(map[string]string)
	headers["From"] = from.String()
	headers["To"] = to.String()
	headers["Subject"] = subj

	// Setup message
	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + body

	auth := smtp.PlainAuth("", mailData["smtp_username"], mailData["smtp_password"], mailData["smtp_server"])

	// TLS config
	tlsconfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         mailData["smtp_server"],
	}

	// Here is the key, you need to call tls.Dial instead of smtp.Dial
	// for smtp servers running on 465 that require an ssl connection
	// from the very beginning (no starttls)
	conn, err := tls.Dial("tcp", smtpHostAndPort, tlsconfig)
	if err != nil {
		return ErrInfo(err)
	}

	c, err := smtp.NewClient(conn, mailData["smtp_server"])
	if err != nil {
		return ErrInfo(err)
	}

	// Auth
	if err = c.Auth(auth); err != nil {
		return ErrInfo(err)
	}

	// To && From
	if err = c.Mail(from.Address); err != nil {
		return ErrInfo(err)
	}

	if err = c.Rcpt(to.Address); err != nil {
		return ErrInfo(err)
	}

	// Data
	w, err := c.Data()
	if err != nil {
		return ErrInfo(err)
	}

	_, err = w.Write([]byte(message))
	if err != nil {
		return ErrInfo(err)
	}

	err = w.Close()
	if err != nil {
		return ErrInfo(err)
	}

	c.Quit()
*/
	/*
		if len(mailData["use_smtp"]) > 0 && len(mailData["smtp_server"]) > 0 {
			e := email.NewEmail()
			e.From = "Dcoin <" + mailData["smtp_username"] + ">"
			e.To = []string{To}
			e.Subject = subject
			e.HTML = []byte(`<table width="100%" cellspacing="0" cellpadding="0" border="0">
	        <tr>
	                 <td style="font-family: 'helvetica neue', 'helvetica', 'arial', 'sans-serif'; font-size: 14px;">
	                          <table width="100%" bgcolor="f0f0f0" color="000000" cellspacing="0" cellpadding="0" border="0">
	                                   <tr>
	                                            <td>
	                                                     <table width="560" align="center" cellspacing="0" cellpadding="8" border="0">
	                                                     <tr>
															<td><img src="http://dcoin.club/email/logo.png" alt="Dcoin" style="width: 280px; height: 62px; margin: 10px 0 15px;" />
																<table width="100%" bgcolor="ffffff" style="border: 1px solid #eeeeee; margin-bottom: 10px; padding: 30px 16px; box-shadow: 0 1px 2px rgba(0,0,0,0.07); line-height: 1.4;" cellspacing="0" cellpadding="0" border="0">
																<tr>
																	<td>
																	<table width="100%" cellspacing="0" cellpadding="0" border="0"><tr><td valign="middle" align="center" height="200" style="font-size: 20px; text-decoration: none; color: #111111;">` + message + `</td></tr></table>
																	</td>
																</tr>
																</table>
															</td>
	                                                     </tr>
	                                                     <tr>
															<td><p style="margin-bottom: 20px; text-align: center; font-size: 11px; color: #555555;">You can cut off the e-mail notifications here: '.$node_url.' -> Settings -> Sms and email notifications</p>
															</td>
	                                                     </tr>
														</table>
	                                            </td>
	                                   </tr>
	                          </table>
	                 </td>
	        </tr>
	</table>';`)
			fmt.Println("smtp_server", mailData["smtp_server"]+":"+mailData["smtp_port"])
			fmt.Println("PlainAuth", smtp.PlainAuth("", mailData["smtp_username"], mailData["smtp_password"], mailData["smtp_server"]+":"+mailData["smtp_port"]))

			err := e.Send(mailData["smtp_server"]+":"+mailData["smtp_port"], smtp.PlainAuth("", mailData["smtp_username"], mailData["smtp_password"], mailData["smtp_server"]+":"+mailData["smtp_port"]))
			if err != nil {
				return ErrInfo(err)
			}
		}*/
/*	return nil
}*/

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

func GetEntropy(hash string) int64 {
	if len(hash) >= 6 {
		result, _ := strconv.ParseInt(hash[0:6], 16, 0)
		return result
	} else {
		return 0
	}
}

/**
 * Определяем, какой юзер должен генерить блок
 *
 * @param int $max_miner_id Общее число майнеров
 * @param string $ctx Энтропия
 * @return int ID майнера
 */
func GetBlockGeneratorMinerId(maxMinerId, ctx int64) int64 {
	var x, hi, lo float64
	hi = float64(ctx) / 127773
	lo = float64(ctx % 127773)
	x = 16807*lo - 2836*hi
	if x <= 0 {
		x += 0x7fffffff
	}
	rez := int64(x) % (maxMinerId + 1)
	if rez == 0 {
		rez = 1
	}
	return rez
}

/*
new
static function get_block_generator_miner_id ($max_miner_id, $ctx)
{
	$n = ceil( log($max_miner_id) / log(16) );
	$hash = $ctx;
	do {
	  $hash = hash('sha256', $hash);
	  $c = substr($hash, 0, $n);
	  $level_0_miner_id = hexdec($c);
	} while( $level_0_miner_id > $max_miner_id || !$level_0_miner_id );

	return $level_0_miner_id;
}
*/

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

func GetBlockGeneratorMinerIdRange(currentMinerId, maxMinerId int64) [][][]int64 {
	var end int64
	var begin int64
	minus1Ok := 0
	minusStop := false
	//var result [][][2]int64{1, 2, 3}
	result := [][][]int64{{{currentMinerId, currentMinerId}}}
	// на верхнем уровне тот, кто генерит блок первым
	var i float64 = 1
	for {
		log.Debug("GetBlockGeneratorMinerIdRange")
		needUsers := math.Pow(2, i)
		//fmt.Println("needUsers", needUsers)
		if begin == 0 {
			var x int64
			if i > 1 {
				x = int64(math.Pow(2, i-1))
			}
			begin = currentMinerId + 1 + x
		} else {
			begin = end + 1
		}
		//fmt.Println("begin", begin)
		if begin == maxMinerId+1 && !minusStop && currentMinerId > 1 && begin != 2 {
			begin = 1
			//fmt.Println("begin ", begin)
			minusStop = true
		} else {
			if begin == currentMinerId || end == currentMinerId || begin > maxMinerId {
				break
			}
		}

		end = begin + int64(needUsers) - 1
		if end > currentMinerId && minus1Ok > 0 {
			//fmt.Println("$end > $cur_miner_id && $minus_1_ok ")
			end = currentMinerId - 1
		}

		end_p := end
		if end_p > maxMinerId {
			//fmt.Println("$end_p > $max_miner_id")
			end_p = maxMinerId
		}

		if end_p == maxMinerId && end_p == currentMinerId {
			//fmt.Println("$end_p == $max_miner_id && $end_p == $cur_miner_id")
			end_p = currentMinerId - 1
		}

		result = append(result, [][]int64{{begin, end_p}})
		var minus int64 = 0
		if end > maxMinerId && !minusStop {
			//fmt.Println("$end > $max_miner_id && !$minus_stop")
			minus = maxMinerId - end
			if int64(math.Abs(float64(minus))) >= currentMinerId {
				minus = -(currentMinerId - 1)
			}
			end = int64(math.Abs(float64(minus)))
			minus1Ok = 1
		}

		if minus != 0 {
			result[int(i)] = append(result[int(i)], []int64{1, int64(math.Abs(float64(minus)))})
		}
		i++
	}
	return result
}

func FindMinerIdLevel(minersIds []int64, levelsRange [][][]int64) (int64, int64) {
	for level, ranges := range levelsRange {
		for _, minerId := range minersIds {
			if minerId >= ranges[0][0] && minerId <= ranges[0][1] {
				return minerId, int64(level)
			}
			if len(ranges) == 2 {
				if minerId >= ranges[1][0] && minerId <= ranges[1][1] {
					return minerId, int64(level)
				}
			}
		}
	}
	return 0, -1
}

func GetIsReadySleep0(level int64, data []int64) int64 {
	if int64(len(data)) > level {
		return data[level]
	} else {
		return 0
	}
}

func GetOurLevelNodes(level int64, levelsRange [][][]int64) []int64 {
	var result []int64
	if level != -1 {
		for i := levelsRange[level][0][0]; i <= levelsRange[level][0][1]; i++ {
			result = append(result, i)
		}
		if len(levelsRange[level]) == 2 {
			for i := levelsRange[level][1][0]; i <= levelsRange[level][1][1]; i++ {
				result = append(result, i)
			}
		}
	}
	return result
}

// на 0-м уровне всегда большее значение, чтобы успели набраться тр-ии
// на остальных уровнях - это время, за которое нужно успеть получить новый блок и занести его в БД
func GetGeneratorSleep(level int64, data []int64) int64 {
	var sleep int64
	if int64(len(data)) > level {
		// суммируем время со всех уровней, которые не успели сгенерить блок до нас
		for i := 0; i <= int(level); i++ {
			sleep += data[i]
		}
	}
	return sleep
}

// сумма is_ready всех предыдущих уровней, которые не успели сгенерить блок
func GetIsReadySleepSum(level int64, data []int64) int64 {
	var sum int64
	if len(data) <= int(level) {
		log.Error("data: %v / level: %d", data, level)
	}
	for i := 0; i < int(level); i++ {
		sum += data[i]
	}
	return sum
}

func EncodeLengthPlusData(data_ interface{}) []byte {
	var data []byte
	switch data_.(type) {
	case string:
		data = []byte(data_.(string))
	case []byte:
		data = data_.([]byte)
	}
	//log.Debug("data: %x", data)
	//log.Debug("len data: %d", len(data))
	return append(EncodeLength(int64(len(data))), data...)
}

func EncodeLength(len0 int64) []byte {
	//log.Debug("len0: %v", len0)
	if len0 <= 127 {
		if len0 < 16 {
			result := HexToBin([]byte(fmt.Sprintf("0%x", len0)))
			log.Debug("%x", result)
			return result
		}
		result := HexToBin([]byte(fmt.Sprintf("%x", len0)))
		log.Debug("%s", fmt.Sprintf("%x", len0))
		log.Debug("%x", []byte(fmt.Sprintf("%x", len0)))
		log.Debug("%x", result)
		return result
	}
	temphex := fmt.Sprintf("%x", len0)
	if len(temphex)%2 > 0 {
		temphex = "0" + temphex
	}
	str, _ := hex.DecodeString(temphex)
	temp := string(str)
	t1 := (0x80 | len(temp))
	t1hex := fmt.Sprintf("%x", t1)
	len_and_t1 := t1hex + temphex
	len_and_t1_bin, _ := hex.DecodeString(len_and_t1)
	//fmt.Println("len_and_t1_bin", len_and_t1_bin)
	//fmt.Printf("len_and_t1_bin %x\n", len_and_t1_bin)
	return len_and_t1_bin
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

func DecodeLength(str *[]byte) int64 {
	var str_ []byte
	str_ = *str
	if len(str_) == 0 {
		return 0
	}
	length_ := []byte(BytesShift(&str_, 1))
	*str = str_
	length := int64(length_[0])
	//fmt.Println(length)
	t1 := (length & 0x80)
	//fmt.Printf("length&0x80 %x", t1)
	if t1 > 0 {
		//fmt.Println("1")
		length &= 0x7F
		//fmt.Printf("length %x\n", length)
		temp := BytesShift(&str_, length)
		*str = str_
		//fmt.Printf("temp %x\n", temp)
		temp2 := fmt.Sprintf("%08x", temp)
		//fmt.Println("temp2", temp2)
		temp3 := HexToDec(temp2)
		//fmt.Println("temp3", temp3)
		return temp3
	}
	return length
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

	log.Debug("forSign", forSign)
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
	return true, nil
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

/*
func DbConnect(configIni map[string]string) *DCDB {
	for {
		db, err := NewDbConnect(configIni)
		if err == nil {
			return db
		} else {
			Sleep(1)
		}
	}
	return nil
}
*/
func DbClose(c *DCDB) {
	err := c.Close()
	if err != nil {
		log.Debug("%v", err)
	}
}

func GetAllMaxPromisedAmount() []int64 {
	var arr []int64
	var iEnd int64
	for i := int64(1); i < 1000000000; i = i * 10 {
		if i == 10 {
			continue
		}
		if i < 100 {
			iEnd = 100
		} else {
			iEnd = 90
		}
		for j := int64(0); j < iEnd; j++ {
			if i < 100 {
				arr = append(arr, int64(i+j))
			} else {
				arr = append(arr, int64(i+j*i/10))
			}
		}
	}
	return arr
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

// время выполнения - 0,01сек на 1000. т.е. на 100 валют майнеру и юзеру уйдет 2 сек.
func GetMaxVote(array []map[int64]int64, min, max, step int64) int64 {

	// если <=10 то тупо берем максимальное кол-во голосов
	var maxPct int64
	if len(array) <= 10 {
		if len(array) > 0 {
			_, maxPct = MaxInSliceMap(array)
			//fmt.Println("0maxPct", maxPct)
		} else {
			maxPct = 0
			//fmt.Println("0maxPct", maxPct)
		}
		return maxPct
	}

	// делим набор данных от $min до $max на секции кратные $step
	// таких секций будет 100 при $max=1000
	// цель такого деления - найти ту секцию, где больше всего голосов

	var sums []map[int64]int64
	dataBank := make(map[int64][]map[int64]int64)
	for i := min; i < max; i = i + step/10 {
		//dataBank[i] = make(map[int64]int64)
		min0 := i
		max0 := i + step
		//fmt.Println("min0", min0)
		//fmt.Println("max0", max0)
		// берем из массива те данные, которые попадают в данную секцию
		for i0 := 0; i0 < len(array); i0++ {
			for number, votes := range array[i0] {
				if number >= min0 && number < max0 {
					dataBank[i] = append(dataBank[i], map[int64]int64{number: votes})
					//fmt.Println("i", i, "number", number, "votes", votes)
				}
			}
		}
		if len(dataBank[i]) > 0 {
			sums = append(sums, map[int64]int64{i: arraySum(dataBank[i])})
			//fmt.Println("i", i, dataBank[i])
		}
	}

	// ищем id секции, где больше всего голосов
	_, maxI := MaxInSliceMap(sums)
	//fmt.Println(sums)
	//fmt.Println(dataBank)
	//fmt.Println(maxI)

	// если в этой секции <= 10-и элементов, то просто выбираем максимальное кол-во голосов
	if len(dataBank[maxI]) <= 10 {
		_, maxPct := MaxInSliceMap(dataBank[maxI])
		return maxPct
	} else { // без рекурсии, просто один раз вызываем свою функцию
		return GetMaxVote(dataBank[maxI], maxI, maxI+step, step/10)
	}
	return 0
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

func GetMinersKeepers(ctx0, maxMinerId0, minersKeepers0 string, arr0 bool) map[int]int {
	ctx := StrToInt(ctx0)
	maxMinerId := StrToInt(maxMinerId0)
	minersKeepers := StrToInt(minersKeepers0)
	result := make(map[int]int)
	newResult := make(map[int]int)
	var ctx_ float64
	ctx_ = float64(ctx)
	for i := 0; i < minersKeepers; i++ {
		//log.Debug("ctx", ctx)
		//var hi float34
		hi := ctx_ / float64(127773)
		//log.Debug("hi", hi)
		lo := int(ctx_) % 127773
		//log.Debug("lo", lo)
		x := (float64(16807) * float64(lo)) - (float64(2836) * hi)
		//log.Debug("x", x, float64(16807), float64(lo), float64(2836), hi)
		if x <= 0 {
			x += 0x7fffffff
		}
		ctx_ = x
		rez := int(ctx_) % (maxMinerId + 1)
		//log.Debug("rez", rez)
		if rez == 0 {
			rez = 1
		}
		result[rez] = 1
	}
	if arr0 {
		i := 0
		for k, _ := range result {
			newResult[i] = k
			i++
		}
	} else {
		newResult = result
	}
	return newResult
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

func GenKeys() (string, string) {
	privatekey, _ := rsa.GenerateKey(crand.Reader, 2048)
	var pemkey = &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privatekey)}
	PrivBytes0 := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: pemkey.Bytes})

	PubASN1, _ := x509.MarshalPKIXPublicKey(&privatekey.PublicKey)
	pubBytes := pem.EncodeToMemory(&pem.Block{Type: "RSA PUBLIC KEY", Bytes: PubASN1})
	s := strings.Replace(string(pubBytes), "-----BEGIN RSA PUBLIC KEY-----", "", -1)
	s = strings.Replace(s, "-----END RSA PUBLIC KEY-----", "", -1)
	sDec, _ := base64.StdEncoding.DecodeString(s)

	return string(PrivBytes0), fmt.Sprintf("%x", sDec)
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

func MakeUpgradeMenu(cur int) ( result string, full bool, next string ) {
	// Если количество майнеров <= 1000, то тогда не требуем профиль
	next = IntToStr( cur + 2 )
	if count, _ := DB.GetCountMiners(); count > 1000 {
		full = true
	} else if cur > 0 {
		next = IntToStr( cur + 1 )
	}
	
	for i := 0; i <= 6; i++ {
		active := ""
		if i <= cur {
			active = `complete`
		} else {
			active = `disabled`
		}
		ind := i + 1
		if !full {
			if i == 1 {
				continue
			} else if i > 1 {
				ind--
			}
		}	
		result += `
			<div class="col-xs-2 col-md-2 bs-wizard-step ` + active + `">
				<div class="text-xs-center bs-wizard-stepnum">` + IntToStr(ind) + `</div>
				<div class="progress"><span class="progress-bar"></span></div>
				<a href="#upgrade` + IntToStr(i+1) + `" class="bs-wizard-dot"></a>
			</div>`
	}
	return
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

func GetBlockBody(host string, blockId int64, dataTypeBlockBody int64, nodeHost string) ([]byte, error) {

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
	if len(nodeHost) > 0 { // защищенный режим
		err = WriteSizeAndDataTCPConn([]byte(nodeHost), conn)
		if err != nil {
			return nil, ErrInfo(err)
		}
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

func EGetReductionLock() (int64, error) {
	return DB.Single("SELECT time FROM e_reduction_lock").Int64()
}

func EUserAmountAndProfit(userId, currencyId int64) float64 {
	var UserCurrencyId, UserLastUpdate int64
	var UserAmount float64
	err := DB.QueryRow(DB.FormatQuery("SELECT currency_id, amount, last_update FROM e_wallets WHERE user_id  =  ? AND currency_id  =  ?"), userId, currencyId).Scan(&UserCurrencyId, &UserAmount, &UserLastUpdate)
	if err != nil {
		return 0
	}
	if UserAmount <= 0 {
		return 0
	}
	profit, err := DB.CalcProfitGen(UserCurrencyId, UserAmount, 0, UserLastUpdate, Time(), "wallet")
	log.Debug("profit: %v / UserCurrencyId: %v / UserAmount: %v / UserLastUpdate: %v", profit, UserCurrencyId, UserAmount, UserLastUpdate)
	return UserAmount + profit
}

func EGetCurrencyList() (map[int64]string, error) {
	rez := make(map[int64]string)
	list, err := DB.GetMap("SELECT id, name FROM e_currency ORDER BY name", "id", "name")
	if err != nil {
		return rez, ErrInfo(err)
	}
	for id, name := range list {
		rez[StrToInt64(id)] = name
	}
	return rez, nil
}

func UpdEWallet(userId, currencyId, lastUpdate int64, amount float64, newAmount bool) error {
	eWallets.Lock()
	exists, err := DB.Single(`SELECT user_id FROM e_wallets WHERE user_id = ? AND currency_id = ?`, userId, currencyId).Int64()
	if err != nil {
		eWallets.Unlock()
		return ErrInfo(err)
	}
	if exists == 0 {
		err = DB.ExecSql("INSERT INTO e_wallets ( user_id, currency_id, amount, last_update ) VALUES ( ?, ?, ?, ? )", userId, currencyId, amount, lastUpdate)
		if err != nil {
			eWallets.Unlock()
			return ErrInfo(err)
		}
	} else {
		if newAmount {
			err = DB.ExecSql("UPDATE e_wallets SET amount = ?, last_update = ? WHERE user_id = ? AND currency_id = ?", amount, lastUpdate, userId, currencyId)
		} else {
			err = DB.ExecSql("UPDATE e_wallets SET amount = amount + ?, last_update = ? WHERE user_id = ? AND currency_id = ?", amount, lastUpdate, userId, currencyId)
		}
		if err != nil {
			eWallets.Unlock()
			return ErrInfo(err)
		}
	}
	eWallets.Unlock()
	return nil
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
			allTransactionsStr+=data["hex_hash"]+"|"+data["verified"]+"|"+data["used"]+"|"+data["high_rate"]+"|"+data["for_self_use"]+"|"+consts.TxTypes[StrToInt(data["type"])]+"|"+data["user_id"]+"|"+data["third_var"]+"|"+data["counter"]+"|"+data["sent"]+"\n"
		}
		t := time.Now()
		data := allTransactionsStr + GetParent() + " ### "+ t.Format(time.StampMicro)+" ### "+text_+"\n\n"
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

func DcoinUpd(url string) error {
	zipfile := filepath.Join(*Dir, "dc.zip" )
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
	tmpname := filepath.Join(*Dir, `tmp_`+appname )
	
	f_ := reader.Reader.File
	f := f_[0]
	zipped, err := f.Open()
	if err != nil {
		return ErrInfo(err)
	}
	writer, err := os.OpenFile( tmpname, os.O_WRONLY|os.O_CREATE, f.Mode())
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
	} else if _, err := os.Stat( filepath.Join( folderPath, appname )); err == nil {
		old = filepath.Join( folderPath, appname )
	} else {
		old = filepath.Join( *Dir, appname )
	}
	log.Debug( tmpname, "-oldFileName", old, "-dir", *Dir, "-oldVersion", consts.VERSION)
	err = exec.Command( tmpname, "-oldFileName", old, "-dir", *Dir, "-oldVersion", consts.VERSION).Start()
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

func ShellExecute( cmdline string ) {
	time.Sleep(500 * time.Millisecond)
	switch runtime.GOOS {
		case "linux":
			exec.Command("xdg-open", cmdline).Start()
		case "windows":
			exec.Command(`rundll32.exe`, `url.dll,FileProtocolHandler`,	cmdline).Start()
		case "darwin":
			exec.Command("open", cmdline).Start()
	}
}