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

package model

import (
	"fmt"
	"strings"
)

const maxBatchRows = 1000

// BatchModel allows bulk insert on BatchModel slice
type BatchModel interface {
	TableName() string
	FieldValue(fieldName string) (interface{}, error)
}

// BatchInsert create and execute batch queries from rows splitted by maxBatchRows and fields
func BatchInsert(rows []BatchModel, fields []string) error {
	queries, values, err := batchQueue(rows, fields)
	if err != nil {
		return err
	}

	for i := 0; i < len(queries); i++ {
		if err := DBConn.Exec(queries[i], values[i]...).Error; err != nil {
			return err
		}
	}

	return nil
}

func batchQueue(rows []BatchModel, fields []string) (queries []string, values [][]interface{}, err error) {
	for len(rows) > 0 {
		if len(rows) > maxBatchRows {
			q, vals, err := prepareQuery(rows[:maxBatchRows], fields)
			if err != nil {
				return queries, values, err
			}

			queries = append(queries, q)
			values = append(values, vals)
			rows = rows[maxBatchRows:]
			continue
		}

		q, vals, err := prepareQuery(rows, fields)
		if err != nil {
			return queries, values, err
		}

		queries = append(queries, q)
		values = append(values, vals)
		rows = nil
	}

	return
}

func prepareQuery(rows []BatchModel, fields []string) (query string, values []interface{}, err error) {
	valueTemplates := make([]string, 0, len(rows))
	values = make([]interface{}, 0, len(rows)*len(fields))
	query = fmt.Sprintf(`INSERT INTO "%s" (%s) VALUES `, rows[0].TableName(), strings.Join(fields, ","))

	rowQSlice := make([]string, 0, len(fields))
	for range fields {
		rowQSlice = append(rowQSlice, "?")
	}

	valueTemplate := fmt.Sprintf("(%s)", strings.Join(rowQSlice, ","))

	for _, row := range rows {
		valueTemplates = append(valueTemplates, valueTemplate)
		for _, field := range fields {
			val, err := row.FieldValue(field)
			if err != nil {
				return query, values, err
			}

			values = append(values, val)
		}
	}

	query += strings.Join(valueTemplates, ",")
	return
}
