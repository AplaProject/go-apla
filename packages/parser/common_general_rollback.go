package parser

import (
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/utils"
)

func (p *Parser) generalRollback(table string, whereUserId_ interface{}, addWhere string, AI bool) error {
	var whereUserId int64
	switch whereUserId_.(type) {
		case string:
		whereUserId = utils.StrToInt64(whereUserId_.(string))
		case []byte:
		whereUserId = utils.BytesToInt64(whereUserId_.([]byte))
		case int:
		whereUserId = int64(whereUserId_.(int))
		case int64:
		whereUserId = whereUserId_.(int64)
	}

	where := ""
	if whereUserId > 0 {
		where = fmt.Sprintf(" WHERE user_id = %d ", whereUserId)
	}
	// получим rb_id, по которому можно найти данные, которые были до этого
	logId, err := p.Single("SELECT rb_id FROM " + table + " " + where + addWhere).Int64()
	if err != nil {
		return utils.ErrInfo(err)
	}
	// если $rb_id = 0, значит восстанавливать нечего и нужно просто удалить запись
	if logId == 0 {
		err = p.ExecSql("DELETE FROM " + table + " " + where + addWhere)
		if err != nil {
			return utils.ErrInfo(err)
		}
	} else {
		// данные, которые восстановим
		data, err := p.OneRow("SELECT * FROM rb_"+table+" WHERE rb_id = ?", logId).String()
		if err != nil {
			return utils.ErrInfo(err)
		}
		addSql := ""
		for k, v := range data {
			// block_id т.к. в rb_ он нужен для удаления старых данных, а в обычной табле не нужен
			if k == "rb_id" || k == "prev_rb_id" || k == "block_id" {
				continue
			}
			if k == "node_public_key" {
				switch p.ConfigIni["db_type"] {
				case "sqlite":
					addSql += fmt.Sprintf("%v='%x',", k, v)
				case "postgresql":
					addSql += fmt.Sprintf("%v=decode('%x','HEX'),", k, v)
				case "mysql":
					addSql += fmt.Sprintf("%v=UNHEX('%x'),", k, v)
				}
			} else {
				addSql += fmt.Sprintf("%v = '%v',", k, v)
			}
		}
		// всегда пишем предыдущий rb_id
		addSql += fmt.Sprintf("rb_id = %v,", data["prev_rb_id"])
		addSql = addSql[0 : len(addSql)-1]
		err = p.ExecSql("UPDATE " + table + " SET " + addSql + where + addWhere)
		if err != nil {
			return utils.ErrInfo(err)
		}
		// подчищаем log
		err = p.ExecSql("DELETE FROM rb_"+table+" WHERE rb_id= ?", logId)
		if err != nil {
			return utils.ErrInfo(err)
		}
		err = p.rollbackAI("rb_"+table, 1)
		if err != nil {
			return utils.ErrInfo(err)
		}
	}
	return nil
}