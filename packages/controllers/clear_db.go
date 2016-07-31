package controllers

import (
	"errors"
	"github.com/astaxie/beego/config"
	"github.com/DayLightProject/go-daylight/packages/utils"
//	"github.com/DayLightProject/go-daylight/packages/stopdaemons"
	"github.com/DayLightProject/go-daylight/packages/daemons"
)

func (c *Controller) ReloadDb() (string, error) {
	// Обнуляем timeSynchro для Synchronization Blockchain
	timeSynchro = 0

	c.Logout()
	for _, ch := range utils.DaemonsChans {
		ch.ChBreaker<-true
	}
	utils.Sleep(2)
	// ClearDb может требоваться в случае оставшейся записи main_lock
	// поэтому ждем и очищаем её 
	c.ExecSql(`DELETE FROM main_lock`)
	for _, ch := range utils.DaemonsChans {
		 <-ch.ChAnswer
	}
	
	utils.Mutex.Lock()	
	defer utils.Mutex.Unlock()
//  Нужно удалять так как в противном случае появится две записи
	err := c.ExecSql(`DELETE FROM install`)
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	confIni, err := config.NewConfig("ini", *utils.Dir+"/config.ini")
	confIni.Set("db_type", "")
	err = confIni.SaveConfigFile(*utils.Dir + "/config.ini")

	daemons.StartDaemons()
//	stopdaemons.Signals()
	utils.Sleep(3)
	return "", nil
}


func (c *Controller) ClearDb() (string, error) {
	if !c.NodeAdmin || c.SessRestricted != 0 {
		return "", utils.ErrInfo(errors.New("Permission denied"))
	}
	return c.ReloadDb()
}
