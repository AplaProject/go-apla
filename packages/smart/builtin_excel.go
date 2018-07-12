package smart

import (
	"bytes"
	"encoding/json"

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"

	xl "github.com/360EntSecGroup-Skylar/excelize"
	"github.com/GenesisKernel/go-genesis/packages/model"
	log "github.com/sirupsen/logrus"
)

// GetJSONFromExcel returns json by parameters range
func GetJSONFromExcel(sc *SmartContract, binaryID, startLine, linesCount, sheetNum int64) (data string, err error) {
	book, err := excelBookFromStoredBinary(sc, binaryID)
	if err != nil || book == nil {
		return ``, err
	}

	sheetName := book.GetSheetName(int(sheetNum))
	rows := book.GetRows(sheetName)
	endLine := startLine + linesCount
	processedRows := []interface{}{}
	for ; startLine < endLine; startLine++ {
		processedRows = append(processedRows, rows[startLine])
	}
	jsonData, err := json.Marshal(processedRows)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.JSONMarshallError, "error": err}).Error("marshalling excel data")
		return ``, err
	}
	return string(jsonData), nil
}

// GetRowsCount returns count of rows from excel file
func GetRowsCount(sc *SmartContract, binaryID, sheetNum int64) (int, error) {
	book, err := excelBookFromStoredBinary(sc, binaryID)
	if err != nil {
		return -1, err
	}

	sheetName := book.GetSheetName(int(sheetNum))
	rows := book.GetRows(sheetName)
	return len(rows), nil
}

func excelBookFromStoredBinary(sc *SmartContract, binaryID int64) (*xl.File, error) {
	bin := &model.Binary{}
	bin.SetTablePrefix(converter.Int64ToStr(sc.TxSmart.EcosystemID))
	found, err := bin.GetByID(binaryID)
	if err != nil {
		return nil, err
	}

	if !found {
		log.WithFields(log.Fields{"binary_id": binaryID}).Error("binary_id not found")
		return nil, nil
	}

	return xl.OpenReader(bytes.NewReader(bin.Data))
}
