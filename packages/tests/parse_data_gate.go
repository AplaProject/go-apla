package main

import (
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"github.com/DayLightProject/go-daylight/packages/tests_utils"
	"github.com/DayLightProject/go-daylight/packages/dcparser"
)

func main() {

	f:=tests_utils.InitLog()
	defer f.Close()

	d := tests_utils.DbConn()
	p := new(dcparser.Parser)
	p.DCDB = d
	var mycandidateBlockBody []byte
	transactionscandidateBlock, err := d.GetAll("SELECT data FROM transactions_candidate_block ORDER BY id ASC", -1)
	if err != nil {
		fmt.Println(utils.ErrInfo(err))
	}
	for _, data := range transactionscandidateBlock {
		fmt.Printf("%x\n", data["data"])
		mycandidateBlockBody = append(mycandidateBlockBody, utils.EncodeLengthPlusData([]byte(data["data"]))...)
	}
	fmt.Println(utils.BinToHex(mycandidateBlockBody))

	if len(mycandidateBlockBody) > 0 {
		fmt.Printf("%x\n", mycandidateBlockBody)
		p.BinaryData = append(utils.DecToBin(0, 1), mycandidateBlockBody...)
		err = p.ParseDataGate(true)
		if err != nil {
			fmt.Println(utils.ErrInfo(err))
		}
	}
}
