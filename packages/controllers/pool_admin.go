package controllers

import (
	"errors"
	"github.com/DayLightProject/go-daylight/packages/consts"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"encoding/json"
)

type poolAdminPage struct {
	Alert        string
	SignData     string
	ShowSignData bool
	CountSignArr []int
	Config       map[string]string
	WaitingList  []map[string]string
	UserId       int64
	Lang         map[string]string
	Users        []map[int64]map[string]string
	ModeError    string
	MyMode       string
	CurrentFillingPool string
	TxType       string
	TxTypeId     int64
	TimeNow      int64
}

func (c *Controller) PoolAdminControl() (string, error) {

	if !c.PoolAdmin {
		return "", utils.ErrInfo(errors.New("access denied"))
	}

	allTable, err := c.GetAllTables()
	if err != nil {
		return "", utils.ErrInfo(err)
	}

	txType := "SwitchPool"
	txTypeId := utils.TypeInt(txType)
	timeNow := utils.Time()

	// удаление юзера с пула
	delId := int64(utils.StrToFloat64(c.Parameters["del_id"]))
	if delId > 0 {

		for _, table := range consts.MyTables {
			if utils.InSliceString(utils.Int64ToStr(delId)+"_"+table, allTable) {
				err = c.ExecSql("DROP TABLE " + utils.Int64ToStr(delId) + "_" + table)
				if err != nil {
					return "", utils.ErrInfo(err)
				}
			}
		}
		err = c.ExecSql("DELETE FROM community WHERE user_id = ?", delId)
		if err != nil {
			return "", utils.ErrInfo(err)
		}
	}

	if _, ok := c.Parameters["pool_tech_works"]; ok {
		poolTechWorks := int64(utils.StrToFloat64(c.Parameters["pool_tech_works"]))
		poolMaxUsers := int64(utils.StrToFloat64(c.Parameters["pool_max_users"]))
		commission := c.Parameters["commission"]

		//if len(commission) > 0 && !utils.CheckInputData(commission, "commission") {
		//	return "", utils.ErrInfo(errors.New("incorrect commission"))
		//}
		err = c.ExecSql("UPDATE config SET pool_tech_works = ?, pool_max_users = ?, commission = ?", poolTechWorks, poolMaxUsers, commission)
		if err != nil {
			return "", utils.ErrInfo(err)
		}
	}

	community, err := c.GetCommunityUsers() // получаем новые данные, т.к. выше было удаление
	var users []map[int64]map[string]string
	for _, uid := range community {
		if uid != c.SessUserId {
			if utils.InSliceString(utils.Int64ToStr(uid)+"_my_table", allTable) {
				data, err := c.OneRow("SELECT miner_id, email FROM " + utils.Int64ToStr(uid) + "_my_table LIMIT 1").String()
				if err != nil {
					return "", utils.ErrInfo(err)
				}
				users = append(users, map[int64]map[string]string{uid: data})
			}
		}
	}
	log.Debug("users", users)

	// лист ожидания попадания в пул
	waitingList, err := c.GetAll("SELECT * FROM pool_waiting_list", -1)

	myMode := ""
	modeError := ""
	if _, ok := c.Parameters["switch_pool_mode"]; ok {
		dq := c.GetQuotes()
		log.Debug("c.Community", c.Community)
		if !c.Community { // сингл-мод

			myUserId, err := c.GetMyUserId("")
			commission, err := c.Single("SELECT commission FROM commission WHERE user_id = ?", myUserId).String()
			if err != nil {
				return "", utils.ErrInfo(err)
			}
			// без комиссии не получится генерить блоки и пр., TestBlock() будет выдавать ошибку
			if len(commission) == 0 {
				modeError = "empty commission"
				myMode = "Single"
			} else {
				// переключаемся в пул-мод
				for _, table := range consts.MyTables {

					err = c.ExecSql("ALTER TABLE " + dq + table + dq + " RENAME TO " + dq + utils.Int64ToStr(myUserId) + "_" + table + dq)
					if err != nil {
						return "", utils.ErrInfo(err)
					}
				}
				err = c.ExecSql("INSERT INTO community (user_id) VALUES (?)", myUserId)
				if err != nil {
					return "", utils.ErrInfo(err)
				}

				log.Debug("UPDATE config SET pool_admin_user_id = ?, pool_max_users = 100, commission = ?", myUserId, commission)
				err = c.ExecSql("UPDATE config SET pool_admin_user_id = ?, pool_max_users = 100, commission = ?", myUserId, commission)
				if err != nil {
					return "", utils.ErrInfo(err)
				}

				// восстановим тех, кто ранее был на пуле
				backup_community, err := c.Single("SELECT data FROM backup_community").Bytes()
				if err != nil {
					return "", utils.ErrInfo(err)
				}
				if len(backup_community) > 0 {
					var community []int
					err = json.Unmarshal(backup_community, &community)
					if err != nil {
						return "", utils.ErrInfo(err)
					}
					for i := 0; i < len(community); i++ {
						// тут дубль при инсерте, поэтому без обработки ошибок
						c.ExecSql("INSERT INTO community (user_id) VALUES (?)", community[i])
					}
				}
				myMode = "Pool"
			}
		} else {

			// бэкап, чтобы при возврате пул-мода, можно было восстановить
			communityUsers := c.CommunityUsers
			jsonData, _ := json.Marshal(communityUsers)
			backup_community, err := c.Single("SELECT data FROM backup_community").String()
			if err != nil {
				return "", utils.ErrInfo(err)
			}
			if len(backup_community) > 0 {
				err := c.ExecSql("UPDATE backup_community SET data = ?", jsonData)
				if err != nil {
					return "", utils.ErrInfo(err)
				}
			} else {
				err = c.ExecSql("INSERT INTO backup_community (data) VALUES (?)", jsonData)
				if err != nil {
					return "", utils.ErrInfo(err)
				}
			}
			myUserId, err := c.GetPoolAdminUserId()
			for _, table := range consts.MyTables {
				err = c.ExecSql("ALTER TABLE " + dq + utils.Int64ToStr(myUserId) + "_" + table + dq + " RENAME TO " + dq + table + dq)
				if err != nil {
					return "", utils.ErrInfo(err)
				}
			}
			err = c.ExecSql("DELETE FROM community")
			if err != nil {
				return "", utils.ErrInfo(err)
			}
			myMode = "Single"
		}
	}


	if myMode == "" && c.Community {
		myMode = "Pool"
	} else if myMode == "" {
		myMode = "Single"
	}

	config, err := c.GetNodeConfig()

	// текущее заполнение пулов
	poolUsers, err := c.Single(`SELECT sum(pool_count_users) / sum(i_am_pool) FROM miners_data`).Float64()
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	currentFillingPool := utils.Float64ToStr(poolUsers / float64(c.Variables.Int64["max_pool_users"]) * 100)
	currentFillingPool = utils.ClearNull(currentFillingPool, 0)

	TemplateStr, err := makeTemplate("pool_admin", "poolAdmin", &poolAdminPage{
		Alert:        c.Alert,
		Lang:         c.Lang,
		ShowSignData: c.ShowSignData,
		SignData:     "",
		Config:       config,
		Users:        users,
		UserId:       c.SessUserId,
		WaitingList:  waitingList,
		MyMode:       myMode,
		ModeError:    modeError,
		TimeNow:      timeNow,
		TxType:       txType,
		TxTypeId:     txTypeId,
		CurrentFillingPool : currentFillingPool,
		CountSignArr: c.CountSignArr})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
