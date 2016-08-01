package daemons

import (
	"errors"
	"flag"
	"github.com/astaxie/beego/config"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"github.com/op/go-logging"
	"os"
	"strings"
	"regexp"
	"github.com/DayLightProject/go-daylight/packages/stopdaemons"
	"fmt"
	"time"
)

var (
	logger = logging.MustGetLogger("daemons")
	/*DaemonCh        chan bool     = make(chan bool, 100)
	AnswerDaemonCh  chan string   = make(chan string, 100)*/
	MonitorDaemonCh chan []string = make(chan []string, 100)
	configIni       map[string]string
)

type daemon struct {
	*utils.DCDB
	goRoutineName  string
	/*DaemonCh       chan bool
	AnswerDaemonCh chan string*/
	chBreaker chan bool
	chAnswer chan string
	sleepTime      int
}

func (d *daemon) dbLock() (error, bool) {
	return d.DbLock(d.chBreaker, d.chAnswer, d.goRoutineName)
}

func (d *daemon) dbUnlock() error {
	logger.Debug("dbUnlock %v", utils.Caller(1))
	return d.DbUnlock(d.goRoutineName)
}

func (d *daemon) dSleep(sleep int) bool {
	for i := 0; i < sleep; i++ {
		if CheckDaemonsRestart(d.chBreaker, d.chAnswer, d.goRoutineName) {
			return true
		}
		utils.Sleep(1)
	}
	return false
}

func (d *daemon) dPrintSleep(err_ interface{}, sleep int) bool {
	var err error
	switch err_.(type) {
	case string:
		err = errors.New(err_.(string))
	case error:
		err = err_.(error)
	}

	if err!=nil {
		logger.Error("%v (%v)", err, utils.GetParent())
	}
	if d.dSleep(sleep) {
		return true
	}
	return false
}

func (d *daemon) unlockPrintSleep(err error, sleep int) bool {
	if err != nil {
		logger.Error("%v", err)
	}
	err = d.DbUnlock(d.goRoutineName)
	if err != nil {
		logger.Error("%v", err)
	}
	for i := 0; i < sleep; i++ {
		if CheckDaemonsRestart(d.chBreaker, d.chAnswer, d.goRoutineName) {
			return true
		}
		utils.Sleep(1)
	}
	return false
}

func (d *daemon) unlockPrintSleepInfo(err error, sleep int) bool {
	if err != nil {
		logger.Debug("%v", err)
	}
	err = d.DbUnlock(d.goRoutineName)
	if err != nil {
		logger.Error("%v", err)
	}

	for i := 0; i < sleep; i++ {
		if CheckDaemonsRestart(d.chBreaker, d.chAnswer, d.goRoutineName) {
			return true
		}
		utils.Sleep(1)
	}
	return false
}

func (d *daemon) notMinerSetSleepTime(sleep int) error {
	community, err := d.GetCommunityUsers()
	if err != nil {
		return err
	}
	logger.Debug("community", community)
	if len(community) == 0 {
		userId, err := d.GetMyUserId("")
		if err != nil {
			return err
		}
		minerId, err := d.GetMinerId(userId)
		if minerId == 0 {
			d.sleepTime = sleep
		}
		logger.Debug("minerId", minerId)
	}
	return nil
}

func ConfigInit() {
	// мониторим config.ini на наличие изменений
	go func() {
		for {
			logger.Debug("ConfigInit monitor")
			if _, err := os.Stat(*utils.Dir + "/config.ini"); os.IsNotExist(err) {
				utils.Sleep(1)
				continue
			}
			configIni_, err := config.NewConfig("ini", *utils.Dir+"/config.ini")
			if err != nil {
				logger.Error("%v", utils.ErrInfo(err))
			}
			configIni, err = configIni_.GetSection("default")
			if err != nil {
				logger.Error("%v", utils.ErrInfo(err))
			}
			if len(configIni["db_type"]) > 0 {
				break
			}
			utils.Sleep(3)
		}
	}()
}

func init() {
	flag.Parse()

}

func CheckDaemonsRestart(chBreaker chan bool, chAnswer chan string, goRoutineName string) bool {
	logger.Debug("CheckDaemonsRestart %v %v", goRoutineName, utils.Caller(2))
	select {
	case <-chBreaker:
		logger.Debug("DaemonCh true %v", goRoutineName)
		chAnswer <- goRoutineName
		return true
	default:
	}
	return false
}

func DbConnect(chBreaker chan bool, chAnswer chan string, goRoutineName string) *utils.DCDB {
	for {
		if CheckDaemonsRestart(chBreaker, chAnswer, goRoutineName) {
			return nil
		}
		if utils.DB == nil || utils.DB.DB == nil {
			utils.Sleep(1)
		} else {
			return utils.DB
		}
	}
	return nil
}


func StartDaemons() {
	utils.DaemonsChans = nil
	daemonsStart := map[string]func(chBreaker chan bool, chAnswer chan string){"TestblockIsReady": TestblockIsReady, "TestblockGenerator": TestblockGenerator, "TestblockDisseminator": TestblockDisseminator, "ReductionGenerator": ReductionGenerator, "QueueParserTx": QueueParserTx, "QueueParserTestblock": QueueParserTestblock, "QueueParserBlocks": QueueParserBlocks,  "Notifications": Notifications, "Disseminator": Disseminator, "Confirmations": Confirmations, "Connector": Connector, "Clear": Clear, "CleaningDb": CleaningDb, "BlocksCollection": BlocksCollection, "Exchange": Exchange, "AutoUpdate": AutoUpdate, "Stats": Stats}
	if utils.Mobile() {
		daemonsStart = map[string]func(chBreaker chan bool, chAnswer chan string){"QueueParserTx": QueueParserTx, "Notifications": Notifications, "Disseminator": Disseminator, "Confirmations": Confirmations, "Connector": Connector, "Clear": Clear, "CleaningDb": CleaningDb, "BlocksCollection": BlocksCollection}
	}
	if *utils.TestRollBack == 1 {
		daemonsStart = map[string]func(chBreaker chan bool, chAnswer chan string){"BlocksCollection": BlocksCollection, "Connector": Connector, "Confirmations": Confirmations}
	}

	if len(configIni["daemons"]) > 0 && configIni["daemons"] != "null" {
		daemonsConf := strings.Split(configIni["daemons"], ",")
		for _, fns := range daemonsConf {
			logger.Debug("start daemon %s", fns)
			fmt.Println("start daemon ", fns)
			var chBreaker chan bool = make(chan bool, 1)
			var chAnswer chan string = make(chan string, 1)
			utils.DaemonsChans = append(utils.DaemonsChans, &utils.DaemonsChansType{ChBreaker: chBreaker, ChAnswer: chAnswer})
			go daemonsStart[fns](chBreaker, chAnswer)
		}
	} else if configIni["daemons"] != "null" {
		for dName, fns := range daemonsStart {
			logger.Debug("start daemon %s", dName)
			fmt.Println("start daemon ", fns)
			var chBreaker chan bool = make(chan bool, 1)
			var chAnswer chan string = make(chan string, 1)
			utils.DaemonsChans = append(utils.DaemonsChans, &utils.DaemonsChansType{ChBreaker: chBreaker, ChAnswer: chAnswer})
			go fns(chBreaker, chAnswer)
		}
	}
	if ischeck, ok := configIni["check_blocks"]; ok && ischeck == `1` {
		time.AfterFunc( 10*time.Second, CheckBlocks )
	}
}


func ClearDb(ChAnswer chan string, goroutineName string) error {

	// остановим демонов, иначе будет паника, когда таблы обнулятся
	fmt.Println("ClearDb() Stop_daemons from DB!")
	for _, ch := range utils.DaemonsChans {
		fmt.Println("ch.ChBreaker<-true")
		ch.ChBreaker<-true
	}
	if len(goroutineName) > 0 {
		ChAnswer<-goroutineName
	}
	for _, ch := range utils.DaemonsChans {
		fmt.Println(<-ch.ChAnswer)
	}

	fmt.Println("ClearDb() Stop_daemons from DB OK")

	// на всякий случай пометим, что работаем
	err = utils.DB.ExecSql("UPDATE main_lock SET script_name = 'cleaning_db'")
	if err != nil {
		return utils.ErrInfo(err)
	}
	err = utils.DB.ExecSql("UPDATE config SET pool_tech_works = 1")
	if err != nil {
		return utils.ErrInfo(err)
	}
	allTables, err := utils.DB.GetAllTables()
	if err != nil {
		return utils.ErrInfo(err)
	}
	for _, table := range allTables {
		logger.Debug("table: %s", table)
		if ok, _ := regexp.MatchString(`^[0-9_]*my_|^e_|install|^config|daemons|payment_systems|community|cf_lang|main_lock`, table); !ok {
			logger.Debug("DELETE FROM %s", table)
			err = utils.DB.ExecSql("DELETE FROM " + table)
			if err != nil {
				return utils.ErrInfo(err)
			}
			if table == "cf_currency" {
				if utils.DB.ConfigIni["db_type"] == "sqlite" {
					err = utils.DB.SetAI("cf_currency", 999)
				} else {
					err = utils.DB.SetAI("cf_currency", 1000)
				}
				if err != nil {
					return utils.ErrInfo(err)
				}
			} else if table == "admin" {
				err = utils.DB.ExecSql("INSERT INTO admin (user_id) VALUES (1)")
				if err != nil {
					return utils.ErrInfo(err)
				}
			} else {
				logger.Debug("SET AI %s", table)
				if utils.DB.ConfigIni["db_type"] == "sqlite" {
					err = utils.DB.SetAI(table, 0)
				} else {
					err = utils.DB.SetAI(table, 1)
				}
				// только логируем, т.к. тут ошибка - это норм
				if err != nil {
					logger.Error("%v", err)
				}
			}
		}
	}

	err = utils.DB.ExecSql("DELETE FROM main_lock")
	if err != nil {
		return utils.ErrInfo(err)
	}

	// запустим демонов
	StartDaemons()
	stopdaemons.Signals()
	utils.Sleep(1)
// мониторим сигнал из БД о том, что демонам надо завершаться
// Похоже это не нужно так как WaitStopTime не прекращает работу и от демонов не зависит
//	go stopdaemons.WaitStopTime()
	return nil
}