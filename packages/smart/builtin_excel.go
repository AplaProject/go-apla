// Copyright (C) 2017, 2018, 2019 EGAAS S.A.
//
// This program is free software; you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation; either version 2 of the License, or (at
// your option) any later version.
//
// This program is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
// General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program; if not, write to the Free Software
// Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301, USA.

package smart

import (
	"bytes"

	"github.com/AplaProject/go-apla/packages/converter"

	xl "github.com/360EntSecGroup-Skylar/excelize"
	"github.com/AplaProject/go-apla/packages/model"
	log "github.com/sirupsen/logrus"
)

// GetDataFromXLSX returns json by parameters range
func GetDataFromXLSX(sc *SmartContract, binaryID, startLine, linesCount, sheetNum int64) (data []interface{}, err error) {
	book, err := excelBookFromStoredBinary(sc, binaryID)
	if err != nil || book == nil {
		return nil, err
	}

	sheetName := book.GetSheetName(int(sheetNum))
	rows := book.GetRows(sheetName)
	endLine := startLine + linesCount
	if endLine > int64(len(rows)) {
		endLine = int64(len(rows))
	}
	processedRows := []interface{}{}
	for ; startLine < endLine; startLine++ {
		var row []interface{}
		for _, item := range rows[startLine] {
			row = append(row, item)
		}
		processedRows = append(processedRows, row)
	}
	return processedRows, nil
}

// GetRowsCountXLSX returns count of rows from excel file
func GetRowsCountXLSX(sc *SmartContract, binaryID, sheetNum int64) (int64, error) {
	book, err := excelBookFromStoredBinary(sc, binaryID)
	if err != nil {
		return -1, err
	}

	sheetName := book.GetSheetName(int(sheetNum))
	rows := book.GetRows(sheetName)
	return int64(len(rows)), nil
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
