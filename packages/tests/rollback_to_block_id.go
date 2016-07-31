package main

import (
	"fmt"
//	"github.com/DayLightProject/go-daylight/packages/utils"
	"github.com/DayLightProject/go-daylight/packages/tests_utils"
	"github.com/DayLightProject/go-daylight/packages/dcparser"
)

func main() {

	f:=tests_utils.InitLog()
	defer f.Close()

	db := tests_utils.DbConn()
	parser := new(dcparser.Parser)
	parser.DCDB = db
	err := parser.RollbackToBlockId(261950)
	if err!=nil {
		fmt.Println(err)
	}

}
