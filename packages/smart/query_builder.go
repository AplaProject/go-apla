package smart

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	log "github.com/sirupsen/logrus"
)

// KeyTableChecker checks table
type KeyTableChecker interface {
	IsKeyTable(string) bool
}

type NextIDGetter interface {
	GetNextID(string) (int64, error)
}

type smartQueryBuilder struct {
	*log.Entry
	tableID      string
	table        string
	isKeyTable   bool
	prepared     bool
	keyEcosystem string
	keyName      string
	Fields       []string
	FieldValues  []interface{}
	stringValues []string
	WhereFields  []string
	WhereValues  []string
	KeyTableChkr KeyTableChecker
	whereExpr    string
}

func (b *smartQueryBuilder) prepare() error {
	if b.prepared {
		return nil
	}

	idNames := strings.SplitN(b.table, `_`, 2)
	if len(idNames) == 2 {
		b.keyName = idNames[1]

		if b.KeyTableChkr.IsKeyTable(b.keyName) {
			b.isKeyTable = true
			b.keyEcosystem = idNames[0]
			b.table = `1_` + b.keyName

			if contains, ecosysIndx := isParamsContainsEcosystem(b.Fields, b.FieldValues); contains {
				if b.WhereFields == nil {
					b.keyEcosystem = fmt.Sprint(b.FieldValues[ecosysIndx])
				}
			} else {
				b.Fields = append(b.Fields, "ecosystem")
				b.FieldValues = append(b.FieldValues, b.keyEcosystem)
			}
		}
	}

	if err := b.normalizeValues(); err != nil {
		b.WithFields(log.Fields{"error": err}).Error("on normalize field values")
		return err
	}

	values, err := converter.InterfaceSliceToStr(b.FieldValues)
	if err != nil {
		b.WithFields(log.Fields{"type": consts.ConversionError, "error": err}).Error("on convert field values to string")
		return err
	}

	b.stringValues = values
	b.prepared = true
	return nil
}

func (b *smartQueryBuilder) getSelectExpr() (string, error) {
	if err := b.prepare(); err != nil {
		return "", err
	}

	fieldsExpr, err := b.GetSQLSelectFieldsExpr()
	if err != nil {
		b.WithFields(log.Fields{"error": err}).Error("on getting sql fields statement")
		return "", err
	}

	whereExpr, err := b.GetSQLWhereExpr()
	if err != nil {
		b.WithFields(log.Fields{"error": err}).Error("on getting sql where statement")
		return "", err
	}
	return fmt.Sprintf(`SELECT %s FROM "%s" %s`, fieldsExpr, b.table, whereExpr), nil
}

func (b *smartQueryBuilder) GetSQLSelectFieldsExpr() (string, error) {
	if err := b.prepare(); err != nil {
		return "", err
	}

	sqlFields := make([]string, 0, len(b.Fields)+1)
	sqlFields = append(sqlFields, "id")

	for i, _ := range b.Fields {
		b.Fields[i] = strings.TrimSpace(strings.ToLower(b.Fields[i]))
		sqlFields = append(sqlFields, toSQLField(b.Fields[i]))
	}

	return strings.Join(sqlFields, ","), nil
}

func (b *smartQueryBuilder) GetSQLWhereExpr() (string, error) {
	if err := b.prepare(); err != nil {
		return "", err
	}

	if b.WhereFields == nil || b.WhereValues == nil {
		return "", nil
	}

	if b.whereExpr != "" {
		return b.whereExpr, nil
	}

	expressions := make([]string, 0, len(b.WhereFields))
	for i := 0; i < len(b.WhereFields); i++ {
		if val := converter.StrToInt64(b.WhereValues[i]); val != 0 {
			expressions = append(expressions, b.WhereFields[i]+" = "+escapeSingleQuotes(b.WhereValues[i]))
		} else {
			expressions = append(expressions, b.WhereFields[i]+" = "+wrapString(escapeSingleQuotes(b.WhereValues[i]), "'"))
		}
	}

	if b.isKeyTable {
		expressions = append(expressions, fmt.Sprintf("ecosystem = '%s'", b.keyEcosystem))
	}

	if len(expressions) > 0 {
		b.whereExpr = " WHERE " + strings.Join(expressions, " AND ") + " "
		return b.whereExpr, nil
	}

	return "", nil
}

func (b *smartQueryBuilder) GetSQLUpdateExpr(logData map[string]string) (string, error) {
	if err := b.prepare(); err != nil {
		return "", err
	}

	expressions := make([]string, 0, len(b.Fields))
	jsonFields := make(map[string]map[string]string)

	for i := 0; i < len(b.Fields); i++ {
		if b.isKeyTable && b.Fields[i] == "ecosystem" {
			continue
		}

		if strings.Contains(b.Fields[i], `->`) {
			colfield := strings.Split(b.Fields[i], `->`)
			if len(colfield) == 2 {
				if jsonFields[colfield[0]] == nil {
					jsonFields[colfield[0]] = make(map[string]string)
				}
				jsonFields[colfield[0]][colfield[1]] = b.stringValues[i]
				continue
			}
		}

		if converter.IsByteColumn(b.table, b.Fields[i]) && len(b.stringValues[i]) != 0 {
			expressions = append(expressions, b.Fields[i]+"="+toSQLHexExpr(b.stringValues[i]))
		} else if b.Fields[i][:1] == "+" || b.Fields[i][:1] == "-" {
			expressions = append(expressions, toArithmeticUpdateExpr(b.Fields[i], b.stringValues[i]))
		} else if b.stringValues[i] == `NULL` {
			expressions = append(expressions, b.Fields[i]+"= NULL")
		} else if strings.HasPrefix(b.Fields[i], prefTimestampSpace) {
			expressions = append(expressions, toTimestampUpdateExpr(b.Fields[i], b.stringValues[i]))
		} else if strings.HasPrefix(b.stringValues[i], prefTimestampSpace) {
			expressions = append(expressions, b.Fields[i]+`= timestamp '`+escapeSingleQuotes(b.stringValues[i][len(`timestamp `):])+`'`)
		} else {
			expressions = append(expressions, `"`+b.Fields[i]+`"='`+escapeSingleQuotes(b.stringValues[i])+`'`)
		}
	}

	for colname, colvals := range jsonFields {
		var initial string
		out, err := json.Marshal(colvals)
		if err != nil {
			log.WithFields(log.Fields{"error": err, "type": consts.JSONMarshallError}).Error("marshalling update columns for jsonb")
			return "", err
		}

		if len(logData[colname]) > 0 && logData[colname] != `NULL` {
			initial = colname
		} else {
			initial = `'{}'`
		}

		expressions = append(expressions, fmt.Sprintf(`%s=%s::jsonb || '%s'::jsonb`, colname, initial, string(out)))
	}

	return strings.Join(expressions, ","), nil
}

func (b *smartQueryBuilder) GetSQLInsertQuery(idGetter NextIDGetter) (string, error) {
	if err := b.prepare(); err != nil {
		return "", err
	}

	isID := false
	insFields := []string{}
	insValues := []string{}
	jsonFields := make(map[string]map[string]string)

	for i := 0; i < len(b.Fields); i++ {
		if b.Fields[i] == `id` {
			isID = true
			b.tableID = escapeSingleQuotes(b.stringValues[i])
		}

		if strings.Contains(b.Fields[i], `->`) {
			colfield := strings.Split(b.Fields[i], `->`)
			if len(colfield) == 2 {
				if jsonFields[colfield[0]] == nil {
					jsonFields[colfield[0]] = make(map[string]string)
				}
				jsonFields[colfield[0]][colfield[1]] = escapeSingleQuotes(b.stringValues[i])
				continue
			}
		}

		insFields = append(insFields, toSQLField(b.Fields[i]))
		insValues = append(insValues, b.toSQLValue(b.stringValues[i], b.Fields[i]))
	}

	for colname, colvals := range jsonFields {
		out, err := json.Marshal(colvals)
		if err != nil {
			log.WithFields(log.Fields{"error": err, "type": consts.JSONMarshallError}).Error("marshalling update columns for jsonb")
			return "", err
		}

		insFields = append(insFields, colname)
		insValues = append(insValues, fmt.Sprintf(`'%s'::jsonb`, string(out)))
	}

	if b.WhereFields != nil && b.WhereValues != nil {
		for i := 0; i < len(b.WhereFields); i++ {
			if b.WhereFields[i] == `id` {
				isID = true
				b.tableID = b.WhereValues[i]
			}
			insFields = append(insFields, b.WhereFields[i])
			insValues = append(insValues, escapeSingleQuotes(b.WhereValues[i]))
		}
	}

	if !isID {
		id, err := idGetter.GetNextID(b.table)
		if err != nil {
			b.Logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting next id for table")
			return "", err
		}

		b.tableID = converter.Int64ToStr(id)
		insFields = append(insFields, `id`)
		insValues = append(insValues, wrapString(b.tableID, "'"))
	}

	flds := strings.Join(insFields, ",")
	vls := strings.Join(insValues, ",")

	return fmt.Sprintf(`INSERT INTO "%s" (%s) VALUES (%s)`, b.table, flds, vls), nil
}

func (b smartQueryBuilder) generateRollBackInfoString(logData map[string]string) (string, error) {
	rollbackInfo := make(map[string]string)
	for k, v := range logData {
		if k == `id` || (b.isKeyTable && k == "ecosystem") {
			continue
		}
		if converter.IsByteColumn(b.table, k) && v != "" {
			rollbackInfo[k] = string(converter.BinToHex([]byte(v)))
		} else {
			rollbackInfo[k] = v
		}
	}

	jsonRollbackInfo, err := json.Marshal(rollbackInfo)
	if err != nil {
		b.Logger.WithFields(log.Fields{"type": consts.JSONMarshallError, "error": err}).Error("marshalling rollback info to json")
		return "", err
	}

	return string(jsonRollbackInfo), nil
}

func (b smartQueryBuilder) toSQLValue(rawValue, rawField string) string {
	if converter.IsByteColumn(b.table, rawField) && len(rawValue) != 0 {
		return toSQLHexExpr(rawValue)
	}

	if rawValue == `NULL` {
		return `NULL`
	}

	if strings.HasPrefix(rawField, prefTimestamp) {
		return toWrapedTimestamp(rawValue)
	}

	if strings.HasPrefix(rawValue, prefTimestamp) {
		return toTimestamp(rawValue)
	}

	return wrapString(escapeSingleQuotes(rawValue), "'")
}

func (b smartQueryBuilder) normalizeValues() error {
	for i, v := range b.FieldValues {
		switch val := v.(type) {
		case string:
			if strings.HasPrefix(strings.TrimSpace(val), prefTimestamp) {
				if err := checkNow(val); err != nil {
					return err
				}
			}

			if len(b.Fields) > i && converter.IsByteColumn(b.table, b.Fields[i]) {
				if vbyte, err := hex.DecodeString(val); err == nil {
					b.FieldValues[i] = vbyte
				}
			}
		}
	}

	return nil
}

func isParamsContainsEcosystem(fields []string, ivalues []interface{}) (bool, int) {
	ecosysIndx := getFieldIndex(fields, `ecosystem`)
	if ecosysIndx >= 0 && len(ivalues) > ecosysIndx && converter.StrToInt64(fmt.Sprint(ivalues[ecosysIndx])) > 0 {
		return true, ecosysIndx
	}

	return false, -1
}

func toSQLHexExpr(value string) string {
	return fmt.Sprintf(" decode('%s','HEX')", hex.EncodeToString([]byte(value)))
}

func toArithmeticUpdateExpr(field, value string) string {
	return field[1:len(field)] + "=" + field[1:len(field)] + field[:1] + escapeSingleQuotes(value)
}

func toTimestampUpdateExpr(field, value string) string {
	return field[len(prefTimestampSpace):] + `= to_timestamp('` + value + `')`
}

func toWrapedTimestamp(value string) string {
	return `to_timestamp('` + escapeSingleQuotes(value) + `')`
}

func toTimestamp(value string) string {
	return prefTimestampSpace + wrapString(escapeSingleQuotes(value[len(prefTimestampSpace):]), "'")
}

func toSQLField(rawField string) string {
	if rawField[:1] == "+" || rawField[:1] == "-" {
		return rawField[1:]
	}

	if strings.HasPrefix(rawField, prefTimestampSpace) {
		return rawField[len(prefTimestampSpace):]
	}

	if strings.Contains(rawField, `->`) {
		return rawField[:strings.Index(rawField, `->`)]
	}

	return wrapString(rawField, `"`)
}

func wrapString(raw, wrapper string) string {
	return wrapper + raw + wrapper
}
