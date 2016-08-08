//calcprofit
package calcprofit

import (
	"fmt"
	"time"
	"log"
	"testing"
	"github.com/astaxie/beego/config"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"github.com/DayLightProject/go-daylight/packages/parser"
	"path/filepath"
)

/*
Проверка
go test -bench=. >out1.txt
go test -bench=. -cpuprofile=cpu.log >>out1.txt
go tool pprof -text calcprofitgen.test.exe cpu.log >>out1.txt
*/
// Путь к базе с lite.db и config.ini
const PATH = `k:\dcoin\my1413`

var p *parser.Parser

func init() {
	*utils.Dir = PATH
	configIni := make(map[string]string)
	configIni_, err := config.NewConfig("ini", filepath.Join( PATH, `config.ini`))
	if err != nil {
		log.Fatalln(`Config`, err)
	} else {
		configIni, err = configIni_.GetSection("default")
	}
	if utils.DB, err = utils.NewDbConnect(configIni); err != nil {
		log.Fatalln(`Utils connect`, err)
	}
	p = new(parser.Parser)
	p.DCDB = utils.DB
}

func BenchmarkCalcProfit(b *testing.B) {
	p.TxUserID = 1

	start := time.Now()
	for i:=0; i<700; i++{
	
		restrictedPA, err := utils.DB.OneRow(`SELECT * from promised_amount_restricted WHERE currency_id = 72 AND user_id = ?`, p.TxUserID).String()
		if err != nil {
			p.ErrInfo(err)
		}
		if len(restrictedPA) == 0 {
			p.ErrInfo("promised_amount_restricted == 0")
		}
		pct, err := p.GetPct()
		if err != nil {
			p.ErrInfo(err)
		}
		startTime := utils.StrToInt64(restrictedPA["last_update"])
		var txTime int64
	/*	if p.BlockData != nil { // тр-ия пришла в блоке
			txTime = p.BlockData.Time
		} else {*/
			txTime = utils.Time() - 300 // просто на всякий случай небольшой запас
	//	}
		p.CalcProfit_(utils.StrToFloat64(restrictedPA["amount"]), startTime, 
		     txTime, pct[72], []map[int64]string{{0: "user"}}, [][]int64{}, []map[int64]string{}, 0, 0)
			
			//fmt.Println(i, utils.Round(test_data[i].result, 8), utils.Round(profit, 8))
	}
	fmt.Println(`Duration`, time.Since( start ))
}
