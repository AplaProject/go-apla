package main

import (
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/parser"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"os"
	"log"
	//"io"
)

type calcProfitTest struct {
	amount float64
	repaid_amount float64
	time_start int64
	time_finish int64
	pct_array []map[int64]map[string]float64
	points_status_array []map[int64]string
	holidays_array [][]int64
	max_promised_amount_array []map[int64]string
	currency_id int64
	result float64
}

func main() {

	var test_data [21]*calcProfitTest


	test_data[0] = new(calcProfitTest)
	test_data[0].amount = 10
	test_data[0].repaid_amount = 0
	test_data[0].time_start = 0
	test_data[0].time_finish = 300
	test_data[0].pct_array = []map[int64]map[string]float64{
		{0:{"user":0.0059, "miner":0.05}},
	}
	test_data[0].points_status_array = []map[int64]string{
		{0:"miner"},
		{101:"user"},
		{200:"miner"},
		{203:"miner"},
	}
	test_data[0].holidays_array = [][]int64 {
		{130,150},
	}
	test_data[0].max_promised_amount_array = []map[int64]string {
		{0:"1000"},
	}
	test_data[0].currency_id = 10
	test_data[0].result = 288977.43266019

	test_data[1] = new(calcProfitTest)
	test_data[1].amount = 10
	test_data[1].repaid_amount = 0
	test_data[1].time_start = 0
	test_data[1].time_finish = 300
	test_data[1].pct_array = []map[int64]map[string]float64{
		{0:{"user":0.0059, "miner":0.05}},
		{250:{"user":0.0019, "miner":0.01}},
		{300:{"user":0.0029, "miner":0.02}},
		{301:{"user":0.0029, "miner":0.03}},
	}
	test_data[1].points_status_array = []map[int64]string{
		{0:"miner"},
		{101:"user"},
		{200:"miner"},
		{203:"miner"},
	}
	test_data[1].holidays_array = [][]int64 {
		{130,150},
	}
	test_data[1].max_promised_amount_array = []map[int64]string {
		{0:"1000"},
	}
	test_data[1].currency_id = 10
	test_data[1].result = 41436.006657618

	test_data[2] = new(calcProfitTest)
	test_data[2].amount = 10
	test_data[2].repaid_amount = 0
	test_data[2].time_start = 0
	test_data[2].time_finish = 300
	test_data[2].pct_array = []map[int64]map[string]float64{
		{0:{"user":0.0059, "miner":0.05}},
		{250:{"user":0.0019, "miner":0.01}},
		{301:{"user":0.0029, "miner":0.03}},
		{300:{"user":0.0029, "miner":0.02}},
	}
	test_data[2].points_status_array = []map[int64]string{
		{0:"miner"},
		{101:"user"},
		{200:"miner"},
		{203:"miner"},
	}
	test_data[2].holidays_array = [][]int64 {
		{130,500},
	}
	test_data[2].max_promised_amount_array = []map[int64]string {
		{0:"1000"},
	}
	test_data[2].currency_id = 10
	test_data[2].result = 1627.6030645193

	test_data[3] = new(calcProfitTest)
	test_data[3].amount = 10
	test_data[3].repaid_amount = 0
	test_data[3].time_start = 0
	test_data[3].time_finish = 300
	test_data[3].pct_array = []map[int64]map[string]float64{
		{0:{"user":0.0059, "miner":0.05}},
		{250:{"user":0.0019, "miner":0.01}},
		{301:{"user":0.0029, "miner":0.03}},
		{300:{"user":0.0029, "miner":0.02}},
	}
	test_data[3].points_status_array = []map[int64]string{
		{0:"miner"},
		{101:"user"},
		{200:"miner"},
		{203:"miner"},
	}
	test_data[3].holidays_array = [][]int64 {
		{130,140},
		{150,160},
		{170,210},
	}
	test_data[3].max_promised_amount_array = []map[int64]string {
		{0:"1000"},
	}
	test_data[3].currency_id = 10
	test_data[3].result = 21317.770423946

	test_data[4] = new(calcProfitTest)
	test_data[4].amount = 10
	test_data[4].repaid_amount = 0
	test_data[4].time_start = 0
	test_data[4].time_finish = 300
	test_data[4].pct_array = []map[int64]map[string]float64{
		{0:{"user":0.0059, "miner":0.05}},
		{250:{"user":0.0019, "miner":0.01}},
		{300:{"user":0.0029, "miner":0.02}},
	}
	test_data[4].points_status_array = []map[int64]string{
		{0:"miner"},
		{101:"user"},
		{200:"miner"},
		{203:"miner"},
		{210:"user"},
	}
	test_data[4].holidays_array = [][]int64 {
		{130,140},
		{150,160},
		{170,210},
	}
	test_data[4].max_promised_amount_array = []map[int64]string {
		{0:"1000"},
	}
	test_data[4].currency_id = 10
	test_data[4].result = 2552.8073541488

	test_data[5] = new(calcProfitTest)
	test_data[5].amount = 10
	test_data[5].repaid_amount = 0
	test_data[5].time_start = 100
	test_data[5].time_finish = 300
	test_data[5].pct_array = []map[int64]map[string]float64{
		{0:{"user":0.0059, "miner":0.05}},
		{150:{"user":0.0029, "miner":0.02}},
	}
	test_data[5].points_status_array = []map[int64]string{
		{0:"miner"},
		{101:"user"},
		{200:"miner"},
	}
	test_data[5].holidays_array = [][]int64 {
		{20,30},
		{90,100},
	}
	test_data[5].max_promised_amount_array = []map[int64]string {
		{0:"1000"},
	}
	test_data[5].currency_id = 10
	test_data[5].result = 107.29341748147

	test_data[6] = new(calcProfitTest)
	test_data[6].amount = 1500
	test_data[6].repaid_amount = 0
	test_data[6].time_start = 100
	test_data[6].time_finish = 300
	test_data[6].pct_array = []map[int64]map[string]float64{
		{0:{"user":0.0059, "miner":0.05}},
		{150:{"user":0.0029, "miner":0.02}},
	}
	test_data[6].points_status_array = []map[int64]string{
		{0:"miner"},
		{101:"user"},
		{200:"miner"},
	}
	test_data[6].holidays_array = [][]int64 {
		{20,30},
		{90,150},
	}
	test_data[6].max_promised_amount_array = []map[int64]string {
		{0:"1000"},
		{150:"1600"},
	}
	test_data[6].currency_id = 1
	test_data[6].result = 15153.345929561

	test_data[7] = new(calcProfitTest)
	test_data[7].amount = 1500
	test_data[7].repaid_amount = 0
	test_data[7].time_start = 100
	test_data[7].time_finish = 300
	test_data[7].pct_array = []map[int64]map[string]float64{
		{0:{"user":0.0059, "miner":0.05}},
		{150:{"user":0.0029, "miner":0.02}},
	}
	test_data[7].points_status_array = []map[int64]string{
		{0:"miner"},
		{101:"user"},
		{200:"miner"},
	}
	test_data[7].holidays_array = [][]int64 {
		{20,30},
		{90,150},
	}
	test_data[7].max_promised_amount_array = []map[int64]string {
		{0:"1000"},
		{150:"1600"},
		{210:"100"},
	}
	test_data[7].currency_id = 10
	test_data[7].result = 4139.6240767059

	test_data[8] = new(calcProfitTest)
	test_data[8].amount = 1500
	test_data[8].repaid_amount = 0
	test_data[8].time_start = 100
	test_data[8].time_finish = 300
	test_data[8].pct_array = []map[int64]map[string]float64{
		{0:{"user":0.0059, "miner":0.05}},
		{150:{"user":0.0029, "miner":0.02}},
	}
	test_data[8].points_status_array = []map[int64]string{
		{0:"miner"},
		{101:"user"},
		{200:"miner"},
	}
	test_data[8].holidays_array = [][]int64 {
		{20,30},
		{90,150},
	}
	test_data[8].max_promised_amount_array = []map[int64]string {
		{0:"1000"},
		{150:"1600"},
		{210:"100"},
	}
	test_data[8].currency_id = 1
	test_data[8].result = 7738.6462401027

	test_data[9] = new(calcProfitTest)
	test_data[9].amount = 1500
	test_data[9].repaid_amount = 0
	test_data[9].time_start = 100
	test_data[9].time_finish = 300
	test_data[9].pct_array = []map[int64]map[string]float64{
		{0:{"user":0.0059, "miner":0.05}},
		{150:{"user":0.0029, "miner":0.02}},
	}
	test_data[9].points_status_array = []map[int64]string{
		{0:"miner"},
		{101:"user"},
		{120:"miner"},
	}
	test_data[9].holidays_array = [][]int64 {
		{20,30},
		{90,101},
	}
	test_data[9].max_promised_amount_array = []map[int64]string {
		{0:"1000"},
		{150:"1600"},
		{210:"100"},
		{220:"10000"},
	}
	test_data[9].currency_id = 10
	test_data[9].result = 100997.3763937

	test_data[10] = new(calcProfitTest)
	test_data[10].amount = 1500
	test_data[10].repaid_amount = 0
	test_data[10].time_start = 100
	test_data[10].time_finish = 300
	test_data[10].pct_array = []map[int64]map[string]float64{
		{0:{"user":0.0059, "miner":0.05}},
		{150:{"user":0.0029, "miner":0.02}},
	}
	test_data[10].points_status_array = []map[int64]string{
		{0:"miner"},
		{101:"user"},
		{120:"miner"},
	}
	test_data[10].holidays_array = [][]int64 {
		{20,30},
		{90,101},
		{299,300},
		{330,350},
	}
	test_data[10].max_promised_amount_array = []map[int64]string {
		{0:"1000"},
		{150:"1600"},
		{210:"100"},
		{220:"10000"},
	}
	test_data[10].currency_id = 10
	test_data[10].result = 98987.623915392

	test_data[11] = new(calcProfitTest)
	test_data[11].amount = 1500
	test_data[11].repaid_amount = 0
	test_data[11].time_start = 100
	test_data[11].time_finish = 300
	test_data[11].pct_array = []map[int64]map[string]float64{
		{0:{"user":0.0059, "miner":0.05}},
		{150:{"user":0.0029, "miner":0.02}},
	}
	test_data[11].points_status_array = []map[int64]string{
		{0:"miner"},
		{101:"user"},
		{120:"miner"},
	}
	test_data[11].holidays_array = [][]int64 {
		{20,350},
	}
	test_data[11].max_promised_amount_array = []map[int64]string {
		{0:"1000"},
		{150:"1600"},
		{210:"100"},
		{220:"10000"},
	}
	test_data[11].currency_id = 10
	test_data[11].result = 0

	test_data[12] = new(calcProfitTest)
	test_data[12].amount = 1500
	test_data[12].repaid_amount = 0
	test_data[12].time_start = 0
	test_data[12].time_finish = 300
	test_data[12].pct_array = []map[int64]map[string]float64{
		{0:{"user":0.0059, "miner":0.05}},
		{200:{"user":0.0029, "miner":0.02}},
	}
	test_data[12].points_status_array = []map[int64]string{
		{0:"miner"},
		{101:"user"},
		{295:"miner"},
	}
	test_data[12].holidays_array = [][]int64 {
		{0,10},
		{10,20},
		{30,40},
		{290,10000000},
	}
	test_data[12].max_promised_amount_array = []map[int64]string {
		{0:"1000"},
		{220:"10000"},
	}
	test_data[12].currency_id = 10
	test_data[12].result = 73337.843828611

	test_data[13] = new(calcProfitTest)
	test_data[13].amount = 1500
	test_data[13].repaid_amount = 0
	test_data[13].time_start = 300
	test_data[13].time_finish = 400
	test_data[13].pct_array = []map[int64]map[string]float64{
		{0:{"user":0.0059, "miner":0.05}},
		{200:{"user":0.0029, "miner":0.02}},
	}
	test_data[13].points_status_array = []map[int64]string{
		{0:"miner"},
		{101:"user"},
		{295:"miner"},
	}
	test_data[13].holidays_array = [][]int64 {
		{0,10},
		{10,20},
		{30,40},
		{290,295},
	}
	test_data[13].max_promised_amount_array = []map[int64]string {
		{0:"1000"},
		{220:"500"},
	}
	test_data[13].currency_id = 10
	test_data[13].result = 3122.3230591262

	test_data[14] = new(calcProfitTest)
	test_data[14].amount = 1500
	test_data[14].repaid_amount = 0
	test_data[14].time_start = 50
	test_data[14].time_finish = 51
	test_data[14].pct_array = []map[int64]map[string]float64{
		{0:{"user":0.0059, "miner":0.05}},
		{200:{"user":0.0029, "miner":0.02}},
	}
	test_data[14].points_status_array = []map[int64]string{
		{0:"miner"},
		{101:"user"},
		{295:"miner"},
	}
	test_data[14].holidays_array = [][]int64 {
		{0,10},
		{10,20},
		{30,40},
		{290,295},
	}
	test_data[14].max_promised_amount_array = []map[int64]string {
		{0:"1000"},
		{220:"500"},
	}
	test_data[14].currency_id = 10
	test_data[14].result = 50

	test_data[15] = new(calcProfitTest)
	test_data[15].amount = 1500
	test_data[15].repaid_amount = 0
	test_data[15].time_start = 50
	test_data[15].time_finish = 51
	test_data[15].pct_array = []map[int64]map[string]float64{
		{0:{"user":0.0059, "miner":0.05}},
		{10:{"user":0.0049, "miner":0.04}},
		{11:{"user":0.0088, "miner":0.08}},
		{200:{"user":0.0029, "miner":0.02}},
	}
	test_data[15].points_status_array = []map[int64]string{
		{0:"miner"},
		{101:"user"},
		{295:"miner"},
	}
	test_data[15].holidays_array = [][]int64 {
		{0,10},
		{10,20},
		{30,40},
		{51,250},
		{290,295},
		{500,600},
	}
	test_data[15].max_promised_amount_array = []map[int64]string {
		{0:"1000"},
		{220:"500"},
	}
	test_data[15].currency_id = 10
	test_data[15].result = 80

	test_data[16] = new(calcProfitTest)
	test_data[16].amount = 1500
	test_data[16].repaid_amount = 0
	test_data[16].time_start = 50
	test_data[16].time_finish = 51
	test_data[16].pct_array = []map[int64]map[string]float64{
		{0:{"user":0.0059, "miner":0.05}},
		{10:{"user":0.0049, "miner":0.04}},
		{11:{"user":0.0088, "miner":0.08}},
		{200:{"user":0.0029, "miner":0.02}},
	}
	test_data[16].points_status_array = []map[int64]string{
		{0:"miner"},
		{101:"user"},
		{295:"miner"},
	}
	test_data[16].holidays_array = [][]int64 {
		{0,10},
		{10,20},
		{30,40},
		{51,250},
		{290,295},
		{500,600},
	}
	test_data[16].max_promised_amount_array = []map[int64]string {
		{0:"1000"},
		{220:"500"},
	}
	test_data[16].currency_id = 1
	test_data[16].result = 80

	test_data[17] = new(calcProfitTest)
	test_data[17].amount = 1500
	test_data[17].repaid_amount = 0
	test_data[17].time_start = 1000
	test_data[17].time_finish = 1001
	test_data[17].pct_array = []map[int64]map[string]float64{
		{0:{"user":0.0059, "miner":0.05}},
		{10:{"user":0.0049, "miner":0.04}},
		{11:{"user":0.0088, "miner":0.08}},
		{200:{"user":0.0029, "miner":0.02}},
	}
	test_data[17].points_status_array = []map[int64]string{
		{0:"miner"},
		{101:"user"},
		{295:"miner"},
	}
	test_data[17].holidays_array = [][]int64 {
		{0,10},
		{10,20},
		{30,40},
		{51,250},
		{290,295},
		{500,600},
	}
	test_data[17].max_promised_amount_array = []map[int64]string {
		{0:"1000"},
		{220:"500"},
	}
	test_data[17].currency_id = 10
	test_data[17].result = 10

	test_data[18] = new(calcProfitTest)
	test_data[18].amount = 1500
	test_data[18].repaid_amount = 0
	test_data[18].time_start = 50
	test_data[18].time_finish = 140
	test_data[18].pct_array = []map[int64]map[string]float64{
		{0:{"user":0.0059, "miner":0.05}},
		{36:{"user":0.0088, "miner":0.08}},
		{164:{"user":0.0049, "miner":0.04}},
		{223:{"user":0.0029, "miner":0.02}},
	}
	test_data[18].points_status_array = []map[int64]string{
		{0:"miner"},
		{98:"miner"},
		{101:"user"},
		{295:"miner"},
	}
	test_data[18].holidays_array = [][]int64 {
		{0,10},
		{10,20},
		{30,30},
		{40,50},
		{66,99},
		{233,1999},
	}
	test_data[18].max_promised_amount_array = []map[int64]string {
		{0:"1000"},
		{63:"3333"},
		{156:"899"},
		{220:"500"},
	}
	test_data[18].currency_id = 10
	test_data[18].result = 5157.6623487708

	test_data[19] = new(calcProfitTest)
	test_data[19].amount = 1500
	test_data[19].repaid_amount = 0
	test_data[19].time_start = 50
	test_data[19].time_finish = 140
	test_data[19].pct_array = []map[int64]map[string]float64{
		{0:{"user":0.0059, "miner":0.05}},
		{36:{"user":0.0088, "miner":0.08}},
		{164:{"user":0.0049, "miner":0.04}},
		{223:{"user":0.0029, "miner":0.02}},
	}
	test_data[19].points_status_array = []map[int64]string{
		{0:"miner"},
		{98:"miner"},
		{101:"user"},
		{295:"miner"},
	}
	test_data[19].holidays_array = [][]int64 {
		{0,10},
		{10,20},
		{30,30},
		{40,50},
		{66,99},
		{233,1999},
	}
	test_data[19].max_promised_amount_array = []map[int64]string {
		{0:"1000"},
		{63:"3333"},
		{156:"899"},
		{220:"500"},
	}
	test_data[19].currency_id = 1
	test_data[19].result = 129106.50065867

	test_data[20] = new(calcProfitTest)
	test_data[20].amount = 1500
	test_data[20].repaid_amount = 50
	test_data[20].time_start = 50
	test_data[20].time_finish = 140
	test_data[20].pct_array = []map[int64]map[string]float64{
		{0:{"user":0.0059, "miner":0.05}},
		{36:{"user":0.0088, "miner":0.08}},
		{164:{"user":0.0049, "miner":0.04}},
		{223:{"user":0.0029, "miner":0.02}},
	}
	test_data[20].points_status_array = []map[int64]string{
		{0:"miner"},
		{98:"miner"},
		{101:"user"},
		{295:"miner"},
	}
	test_data[20].holidays_array = [][]int64 {
		{0,10},
		{10,20},
		{30,30},
		{40,50},
		{66,99},
		{233,1999},
	}
	test_data[20].max_promised_amount_array = []map[int64]string {
		{0:"1000"},
		{63:"1525"},
		{64:"1550"},
		{139:"500"},
	}
	test_data[20].currency_id = 10
	test_data[20].result = 4966.7977985526

	f, _ := os.OpenFile("dclog.txt", os.O_WRONLY | os.O_APPEND | os.O_CREATE, 0777)
	defer f.Close()
//	log.SetOutput(io.MultiWriter(f, os.Stdout))
	log.SetOutput(f)
	log.SetFlags(log.LstdFlags | log.Lshortfile)


	for i:=0; i<21; i++{
		p:=new(parser.Parser)
		profit, _ := p.CalcProfit_24946(test_data[i].amount, test_data[i].time_start, test_data[i].time_finish, test_data[i].pct_array, test_data[i].points_status_array, test_data[i].holidays_array, test_data[i].max_promised_amount_array, test_data[i].currency_id, test_data[i].repaid_amount)
		//fmt.Println(i, utils.Round(test_data[i].result, 8), utils.Round(profit, 8))
	}
}
