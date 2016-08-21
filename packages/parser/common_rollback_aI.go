package parser

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
)

// откатываем ID на кол-во затронутых строк
func (p *Parser) rollbackAI(table string, num int64) error {

	if num == 0 {
		return nil
	}

	AiId, err := p.GetAiId(table)
	if err != nil {
		return utils.ErrInfo(err)
	}
	log.Debug("AiId: %s", AiId)
	// если табла была очищена, то тут будет 0, поэтому нелья чистить таблы под нуль
	current, err := p.Single("SELECT " + AiId + " FROM " + table + " ORDER BY " + AiId + " DESC LIMIT 1").Int64()
	if err != nil {
		return utils.ErrInfo(err)
	}
	NewAi := current + num
	log.Debug("NewAi: %d", NewAi)

	if p.ConfigIni["db_type"] == "postgresql" {
		pg_get_serial_sequence, err := p.Single("SELECT pg_get_serial_sequence('" + table + "', '" + AiId + "')").String()
		if err != nil {
			return utils.ErrInfo(err)
		}
		err = p.ExecSql("ALTER SEQUENCE " + pg_get_serial_sequence + " RESTART WITH " + utils.Int64ToStr(NewAi))
		if err != nil {
			return utils.ErrInfo(err)
		}
	} else if p.ConfigIni["db_type"] == "mysql" {
		err := p.ExecSql("ALTER TABLE " + table + " AUTO_INCREMENT = " + utils.Int64ToStr(NewAi))
		if err != nil {
			return utils.ErrInfo(err)
		}
	} else if p.ConfigIni["db_type"] == "sqlite" {
		NewAi--
		err := p.ExecSql("UPDATE SQLITE_SEQUENCE SET seq = ? WHERE name = ?", NewAi, table)
		if err != nil {
			return utils.ErrInfo(err)
		}
	}
	return nil
}
