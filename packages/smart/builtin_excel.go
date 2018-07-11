package smart

import (
	"bytes"

	"encoding/json"

	xl "github.com/360EntSecGroup-Skylar/excelize"
	"github.com/GenesisKernel/go-genesis/packages/model"
	log "github.com/sirupsen/logrus"
)

// GetJSONFromExcel returns json by parameters range
func GetJSONFromExcel(sc *SmartContract, binaryID, startLine, linesCount, sheetNum int64) (data []byte, err error) {
	book, err := excelBookFromStoredBinary(binaryID)
	if err != nil || book == nil {
		return nil, err
	}

	sheetName := book.GetSheetName(int(sheetNum))
	rows := book.GetRows(sheetName)
	endLine := startLine + linesCount
	processedRows := []interface{}{}
	for ; startLine <= endLine; startLine++ {
		processedRows = append(processedRows, processRow(rows[startLine]))
	}

	return json.Marshal(processedRows)
}

// GetRowsCount returns count of rows from excel file
func GetRowsCount(sc *SmartContract, binaryID, sheetNum int64) (int, error) {
	book, err := excelBookFromStoredBinary(binaryID)
	if err != nil {
		return -1, err
	}

	sheetName := book.GetSheetName(int(sheetNum))
	rows := book.GetRows(sheetName)
	return len(rows), nil
}

func processRow(row []string) interface{} {
	return nil
}

func excelBookFromStoredBinary(binaryID int64) (*xl.File, error) {
	bin := &model.Binary{}
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
