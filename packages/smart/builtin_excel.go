// Apla Software includes an integrated development
// environment with a multi-level system for the management
// of access rights to data, interfaces, and Smart contracts. The
// technical characteristics of the Apla Software are indicated in
// Apla Technical Paper.
//
// Apla Users are granted a permission to deal in the Apla
// Software without restrictions, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of Apla Software, and to permit persons
// to whom Apla Software is furnished to do so, subject to the
// following conditions:
// * the copyright notice of GenesisKernel and EGAAS S.A.
// and this permission notice shall be included in all copies or
// substantial portions of the software;
// * a result of the dealing in Apla Software cannot be
// implemented outside of the Apla Platform environment.
//
// THE APLA SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY
// OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED
// TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A
// PARTICULAR PURPOSE, ERROR FREE AND NONINFRINGEMENT. IN
// NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
// LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR
// THE USE OR OTHER DEALINGS IN THE APLA SOFTWARE.

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
