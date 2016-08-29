package daylight

import (
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/utils"
)


func Stop() {
	log.Debug("Stop()")
	IosLog("Stop()")
	var err error
	utils.DB, err = utils.NewDbConnect(configIni)
	log.Debug("DayLight Stop : %v", utils.DB)
	IosLog("utils.DB:" + fmt.Sprintf("%v", utils.DB))
	if err != nil {
		IosLog("err:" + fmt.Sprintf("%s", utils.ErrInfo(err)))
		log.Error("%v", utils.ErrInfo(err))
		//panic(err)
		//os.Exit(1)
	}
	err = utils.DB.ExecSql(`INSERT INTO stop_daemons(stop_time) VALUES (?)`, utils.Time())
	if err != nil {
		IosLog("err:" + fmt.Sprintf("%s", utils.ErrInfo(err)))
		log.Error("%v", utils.ErrInfo(err))
	}
	log.Debug("DayLight Stop")
	IosLog("DayLight Stop")
}
