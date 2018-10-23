package smart

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/model"
	log "github.com/sirupsen/logrus"
)

type smartQueryBuilder struct {
	tableID               string
	table                 string
	isKeyTable            bool
	keyEcosystem, keyName string
}

func CreateQueryBuilder(table string, fields, whereFields, whereValues []string, ivalues []interface{}) smartQueryBuilder {
	builder := smartQueryBuilder{}

	idNames := strings.SplitN(table, `_`, 2)
	if len(idNames) == 2 {
		builder.keyName = idNames[1]
		if v, ok := model.FirstEcosystemTables[builder.keyName]; ok && !v {
			builder.isKeyTable = true
			builder.keyEcosystem = idNames[0]
			builder.table = `1_` + builder.keyName

			if contains, ecosysIndx := isParamsContainsEcosystem(fields, ivalues); contains {
				if whereFields == nil {
					builder.keyEcosystem = fmt.Sprint(ivalues[ecosysIndx])
				}
			} else {
				fields = append(fields, "ecosystem")
				ivalues = append(ivalues, builder.keyEcosystem)
			}
		}
	}

	return builder
}

func (b smartQueryBuilder) getSelectExpr(fields, whereFields, whereValues []string) string {
	fieldsExpr := b.getSQLSelectFieldsExpr(fields)
	whereExpr := b.getSQLWhereExpr(whereFields, whereValues)
	return `SELECT ` + fieldsExpr + ` FROM "` + b.table + `" ` + whereExpr
}

func (b smartQueryBuilder) getSQLSelectFieldsExpr(fields []string) string {
	sqlFields := make([]string, len(fields)+1)
	sqlFields = append(sqlFields, "id")

	for i, _ := range fields {
		fields[i] = strings.TrimSpace(strings.ToLower(fields[i]))
		sqlFields = append(sqlFields, toSQLField(fields[i]))

	}

	return strings.Join(sqlFields, ",")
}

func (b smartQueryBuilder) getSQLWhereExpr(fields, values []string) string {
	if fields == nil || values == nil {
		return ""
	}

	expressions := make([]string, len(fields))
	for i := 0; i < len(fields); i++ {
		if val := converter.StrToInt64(values[i]); val != 0 {
			expressions = append(expressions, fields[i]+" = "+escapeSingleQuotes(values[i]))
		} else {
			expressions = append(expressions, fields[i]+" = "+wrapString(escapeSingleQuotes(values[i]), "'"))
		}
	}

	if b.isKeyTable {
		expressions = append(expressions, fmt.Sprintf("ecosystem = '%s'", b.keyEcosystem))
	}

	if len(expressions) > 0 {
		return " WHERE " + strings.Join(expressions, " AND ") + "\n"
	}

	return ""
}

func (b smartQueryBuilder) getSQLUpdateExpr(fields, values []string, logData map[string]string) (string, error) {
	expressions := make([]string, len(fields))
	jsonFields := make(map[string]map[string]string)

	for i := 0; i < len(fields); i++ {
		if b.isKeyTable && fields[i] == "ecosystem" {
			continue
		}

		if strings.Contains(fields[i], `->`) {
			colfield := strings.Split(fields[i], `->`)
			if len(colfield) == 2 {
				if jsonFields[colfield[0]] == nil {
					jsonFields[colfield[0]] = make(map[string]string)
				}
				jsonFields[colfield[0]][colfield[1]] = values[i]
				continue
			}
		}

		if converter.IsByteColumn(b.table, fields[i]) && len(values[i]) != 0 {
			expressions = append(expressions, fields[i]+"="+toSQLHexExpr(values[i]))
		} else if fields[i][:1] == "+" || fields[i][:1] == "-" {
			expressions = append(expressions, toArithmeticUpdateExpr(fields[i], values[i]))
		} else if values[i] == `NULL` {
			expressions = append(expressions, fields[i]+"= NULL")
		} else if strings.HasPrefix(fields[i], prefTimestampSpace) {
			expressions = append(expressions, toTimestampUpdateExpr(fields[i], values[i]))
		} else if strings.HasPrefix(values[i], prefTimestampSpace) {
			expressions = append(expressions, fields[i]+`= timestamp '`+escapeSingleQuotes(values[i][len(`timestamp `):])+`'`)
		} else {
			expressions = append(expressions, `"`+fields[i]+`"='`+escapeSingleQuotes(values[i])+`'`)
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

func (b smartQueryBuilder) getSQLInsertStatement(fields, values, whereFields []string) (string, error) {
	isID := false
	insFields := []string{}
	insValues := []string{}
	jsonFields := make(map[string]map[string]string)

	for i := 0; i < len(fields); i++ {
		if fields[i] == `id` {
			isID = true
			b.tableID = escapeSingleQuotes(values[i])
		}

		if strings.Contains(fields[i], `->`) {
			colfield := strings.Split(fields[i], `->`)
			if len(colfield) == 2 {
				if jsonFields[colfield[0]] == nil {
					jsonFields[colfield[0]] = make(map[string]string)
				}
				jsonFields[colfield[0]][colfield[1]] = escapeSingleQuotes(values[i])
				continue
			}
		}

		insFields = append(insFields, toSQLField(fields[i]))
		insValues = append(insValues, b.toSQLValue(values[i], fields[i]))
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

	if whereFields != nil && whereValues != nil {
		for i := 0; i < len(whereFields); i++ {
			if whereFields[i] == `id` {
				isID = true
				b.tableID = whereValues[i]
			}
			insFields = append(insFields, whereFields[i])
			insValues = append(insValues, escapeSingleQuotes(whereValues[i]))
		}
	}
	if !isID {
		id, err := model.GetNextID(sc.DbTransaction, b.table)
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting next id for table")
			return 0, ``, err
		}
		tableID = converter.Int64ToStr(id)
		addSQLIns0 = append(addSQLIns0, `id`)
		addSQLIns1 = append(addSQLIns1, `'`+tableID+`'`)
	}
	insertQuery := `INSERT INTO "` + table + `" (` + strings.Join(addSQLIns0, ",") +
		`) VALUES (` + strings.Join(addSQLIns1, ",") + `)`
}

func (b smartQueryBuilder) generateRollBackInfoString(logData map[string]string) (string, error) {
	rollbackInfo := make(map[string]string)
	for k, v := range logData {
		if k == `id` || (isKeyTable && k == "ecosystem") {
			continue
		}
		if converter.IsByteColumn(table, k) && v != "" {
			rollbackInfo[k] = string(converter.BinToHex([]byte(v)))
		} else {
			rollbackInfo[k] = v
		}
	}

	jsonRollbackInfo, err := json.Marshal(rollbackInfo)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.JSONMarshallError, "error": err}).Error("marshalling rollback info to json")
		return "", err
	}

	return string(jsonRollbackInfo)
}

func (b smartQueryBuilder) toSQLValue(rawValue, rawField string) string {
	if converter.IsByteColumn(b.table, rawField) && len(rawValue) != 0 {
		return toSQLHexExpr(values[i])
	}

	if values[i] == `NULL` {
		return `NULL`
	}

	if strings.HasPrefix(fields[i], prefTimestamp) {
		return toWrapedTimestamp(values[i])
	}

	if strings.HasPrefix(values[i], prefTimestamp) {
		return toTimestamp(values[i])
	}

	return wrapString(escapeSingleQuotes(values[i]), "'")
}

func isParamsContainsEcosystem(fields []string, ivalues []interface{}) (bool, int) {
	ecosysIndx := getFieldIndex(fields, `ecosystem`)
	if ecosysIndx >= 0 && len(ivalues) > ecosysIndx && converter.StrToInt64(fmt.Sprint(ivalues[ecosysIndx])) > 0 {
		return true, ecosysIndx
	}

	return false, -1
}

func normalizeValues(fields []string, values []interface{}) error {
	for i, v := range values {
		switch val := v.(type) {
		case string:
			if strings.HasPrefix(strings.TrimSpace(val), prefTimestamp) {
				if err := checkNow(val); err != nil {
					return err
				}
			}

			if len(fields) > i && converter.IsByteColumn(table, fields[i]) {
				if vbyte, err := hex.DecodeString(val); err == nil {
					values[i] = vbyte
				}
			}
		}
	}
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
	return prefTimestampSpace + wrapString(escapeSingleQuotes(values[i][len(prefTimestampSpace):]), "'")
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
	len := len([]byte(raw)) + len([]byte(wrapper))*2
	bts := make([]byte, len)
	buf := bytes.NewBuffer(bts)
	buf.WriteString(wrapper)
	buf.WriteString(raw)
	buf.WriteString(wrapper)
	return buf.String()
}
