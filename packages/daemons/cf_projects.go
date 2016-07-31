package daemons

import (
	"encoding/json"
	"github.com/DayLightProject/go-daylight/packages/utils"
)

func CfProjects(chBreaker chan bool, chAnswer chan string) {
	defer func() {
		if r := recover(); r != nil {
			logger.Error("daemon Recovered", r)
			panic(r)
		}
	}()

	const GoroutineName = "CfProjects"
	d := new(daemon)
	d.DCDB = DbConnect(chBreaker, chAnswer, GoroutineName)
	if d.DCDB == nil {
		return
	}
	d.goRoutineName = GoroutineName
	d.chAnswer = chAnswer
	d.chBreaker = chBreaker
	if utils.Mobile() {
		d.sleepTime = 1800
	} else {
		d.sleepTime = 60
	}
	if !d.CheckInstall(chBreaker, chAnswer, GoroutineName) {
		return
	}
	d.DCDB = DbConnect(chBreaker, chAnswer, GoroutineName)
	if d.DCDB == nil {
		return
	}

BEGIN:
	for {
		logger.Info(GoroutineName)
		MonitorDaemonCh <- []string{GoroutineName, utils.Int64ToStr(utils.Time())}

		// проверим, не нужно ли нам выйти из цикла
		if CheckDaemonsRestart(chBreaker, chAnswer, GoroutineName) {
			break BEGIN
		}

		err, restart := d.dbLock()
		if restart {
			break BEGIN
		}
		if err != nil {
			if d.dPrintSleep(err, d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}

		// гео-декодирование
		all, err := d.GetAll(`
				SELECT id,
							latitude,
							longitude
				FROM cf_projects
				WHERE geo_checked= 0
				`, -1)
		for _, cf_projects := range all {
			gmapData, err := utils.GetHttpTextAnswer("http://maps.googleapis.com/maps/api/geocode/json?latlng=" + cf_projects["latitude"] + "," + cf_projects["longitude"] + "&sensor=true_or_false")
			if err != nil {
				if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
					break BEGIN
				}
				continue BEGIN
			}
			var gmap map[string][]map[string][]map[string]string
			json.Unmarshal([]byte(gmapData), &gmap)
			if len(gmap["results"]) > 1 && len(gmap["results"][len(gmap["results"])-2]["address_components"]) > 0 {
				country := gmap["results"][len(gmap["results"])-2]["address_components"][0]["long_name"]
				city := gmap["results"][len(gmap["results"])-2]["address_components"][1]["short_name"]
				err = d.ExecSql("UPDATE cf_projects SET country = ?, city = ?, geo_checked= 1 WHERE id = ?", country, city, cf_projects["id"])
				if err != nil {
					if d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime) {
						break BEGIN
					}
					continue BEGIN
				}
			}
		}

		// финансирование проектов
		cf_funding, err := d.GetAll(`
				SELECT  id,
							 project_id,
							 amount
				FROM cf_funding
				WHERE checked= 0
				`, -1)
		for _, data := range cf_funding {
			// отмечаем, чтобы больше не брать
			err = d.ExecSql("UPDATE cf_funding SET checked = 1 WHERE id = ?", data["id"])
			if err != nil {
				if d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime) {
					break BEGIN
				}
				continue BEGIN
			}
			// сколько собрано средств
			funding, err := d.Single("SELECT sum(amount) FROM cf_funding WHERE project_id  =  ? AND del_block_id  =  0", data["project_id"]).Float64()
			if err != nil {
				if d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime) {
					break BEGIN
				}
				continue BEGIN
			}

			// сколько всего фундеров
			countFunders, err := d.Single("SELECT count(id) FROM cf_funding WHERE project_id  = ? AND del_block_id  =  0 GROUP BY user_id", data["project_id"]).Int64()
			if err != nil {
				if d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime) {
					break BEGIN
				}
				continue BEGIN
			}

			// обновляем кол-во фундеров и собранные средства
			err = d.ExecSql("UPDATE cf_projects SET funding = ?, funders = ? WHERE id = ?", funding, countFunders, data["project_id"])
			if err != nil {
				if d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime) {
					break BEGIN
				}
				continue BEGIN
			}
		}

		d.dbUnlock()

		if d.dSleep(d.sleepTime) {
			break BEGIN
		}
	}
	logger.Debug("break BEGIN %v", GoroutineName)
}
