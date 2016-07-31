package daemons

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
	"io/ioutil"
	"os"
)

func AutoUpdate(chBreaker chan bool, chAnswer chan string) {
	defer func() {
		if r := recover(); r != nil {
			logger.Error("daemon Recovered", r)
			panic(r)
		}
	}()

	GoroutineName := "AutoUpdate"
	d := new(daemon)
	d.DCDB = DbConnect(chBreaker, chAnswer, GoroutineName)
	if d.DCDB == nil {
		return
	}
	d.goRoutineName = GoroutineName
	d.chAnswer = chAnswer
	d.chBreaker = chBreaker
	d.sleepTime = 3600
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

		config, err := d.GetNodeConfig()
		if err != nil {
			if d.dPrintSleep(err, d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}

		if config["auto_update"] == "1" {
			updTime, _ := ioutil.ReadFile(*utils.Dir + "/auto_update")
			logger.Debug("updTime %v / ", utils.BytesToInt64(updTime))
			//fmt.Println(utils.BytesToInt64(updTime))
			if utils.Time()-utils.BytesToInt64(updTime) < int64(d.sleepTime) {
				logger.Debug("sleepTime")
				//fmt.Println("sleepTime")
				if d.dSleep(d.sleepTime) {
					break BEGIN
				}
				continue BEGIN
			}
			_, url, err := utils.GetUpdVerAndUrl(config["auto_update_url"])
			//fmt.Println("url", url)
			if err != nil {
				if d.dPrintSleep(err, d.sleepTime) {
					break BEGIN
				}
				continue BEGIN
			}
			if len(url) > 0 {
				f, _ := os.OpenFile(*utils.Dir+"/auto_update", os.O_WRONLY|os.O_CREATE, 0600)
				f.WriteString(utils.Int64ToStr(utils.Time()))
				f.Close()
				err = utils.DcoinUpd(url)
				if err != nil {
					if d.dPrintSleep(err, d.sleepTime) {
						break BEGIN
					}
					continue BEGIN
				}
			}
		}

		if d.dSleep(d.sleepTime) {
			break BEGIN
		}
	}
	logger.Debug("break BEGIN %v", GoroutineName)
}
