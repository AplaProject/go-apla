package schema

import (
	"github.com/DayLightProject/go-daylight/packages/consts"
	"github.com/DayLightProject/go-daylight/packages/utils"
)

func Migration() {
	oldDbVersion, err := utils.DB.Single(`SELECT version FROM migration_history ORDER BY id DESC LIMIT 1`).String()
	if err != nil {
		log.Error("%v", utils.ErrInfo(err))
	}
	if len(*utils.OldVersion) == 0 && consts.VERSION != oldDbVersion {
		*utils.OldVersion = oldDbVersion
	}

	log.Debug("*utils.OldVersion %v", *utils.OldVersion)
	if len(*utils.OldVersion) > 0 {

		err = utils.DB.ExecSql(`INSERT INTO migration_history (version, date_applied) VALUES (?, ?)`, consts.VERSION, utils.Time())
		if err != nil {
			log.Error("%v", utils.ErrInfo(err))
		}
	}
}

