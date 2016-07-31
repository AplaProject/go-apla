package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
	"io/ioutil"
	"os"
	"strings"
)

type newUserPage struct {
	SignData        string
	ShowSignData    bool
	TxType          string
	TxTypeId        int64
	TimeNow         int64
	UserId          int64
	Alert           string
	Lang            map[string]string
	CountSignArr    []int
	MyRefs          map[int64]myRefsType
	GlobalRefs      map[int64]globalRefsType
	CurrencyList    map[int64]string
	PoolUrl         string
	RefPhotos       map[int64][]string
//	LastTxFormatted string
}

func (c *Controller) NewUser() (string, error) {

	txType := "NewUser"
	txTypeId := utils.TypeInt(txType)
	timeNow := utils.Time()

	param := utils.ParamType{X: 176, Y: 100, Width: 100, Bg_path: "static/img/k_bg.png"}

	refPhotos := make(map[int64][]string)
	myRefsKeys := make(map[int64]map[string]string)
	if c.SessRestricted == 0 {
		join := c.MyPrefix + `my_new_users.user_id`
		if c.ConfigIni["db_type"] == "sqlite" || c.ConfigIni["db_type"] == "postgresql" {
			join = `"` + c.MyPrefix + `my_new_users".user_id`
		}
		rows, err := c.Query(c.FormatQuery(`
				SELECT users.user_id,	private_key,  log_id
				FROM ` + c.MyPrefix + `my_new_users
				LEFT JOIN users ON users.user_id = ` + join + `
				WHERE ` + c.MyPrefix + `my_new_users.status = 'approved'
				`))
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		defer rows.Close()
		for rows.Next() {
			var user_id, log_id int64
			var private_key string
			err = rows.Scan(&user_id, &private_key, &log_id)
			if err != nil {
				return "", utils.ErrInfo(err)
			}
			// проверим, не сменил ли уже юзер свой ключ
			StrUserId := utils.Int64ToStr(user_id)
			if log_id != 0 {
				myRefsKeys[user_id] = map[string]string{"user_id": StrUserId}
			} else {
				myRefsKeys[user_id] = map[string]string{"user_id": StrUserId, "private_key": private_key}
				md5 := string(utils.Md5(private_key))
				kPath := *utils.Dir + "/public/" + md5[0:16]
				kPathPng := kPath + ".png"
				kPathTxt := kPath + ".txt"
				if _, err := os.Stat(kPathPng); os.IsNotExist(err) {
					privKey := strings.Replace(private_key, "-----BEGIN RSA PRIVATE KEY-----", "", -1)
					privKey = strings.Replace(privKey, "-----END RSA PRIVATE KEY-----", "", -1)
					_, err = utils.KeyToImg(privKey, kPathPng, user_id, c.TimeFormat, param)
					if err != nil {
						return "", utils.ErrInfo(err)
					}
					err := ioutil.WriteFile(kPathTxt, []byte(privKey), 0644)
					if err != nil {
						return "", utils.ErrInfo(err)
					}
					/*$gd = key_to_img($private_key, $param, $row['user_id']);
					imagepng($gd, $k_path_png);
					file_put_contents($k_path_txt, trim($private_key));*/
				}
			}
		}
	}

	refs := make(map[int64]map[int64]float64)
	// инфа по рефам юзера
	rows, err := c.Query(c.FormatQuery(`
			SELECT referral, sum(amount) as amount, currency_id
			FROM referral_stats
			WHERE user_id = ?
			GROUP BY currency_id,  referral
			`), c.SessUserId)
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	defer rows.Close()
	for rows.Next() {
		var referral, currency_id int64
		var amount float64
		err = rows.Scan(&referral, &amount, &currency_id)
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		refs[referral] = map[int64]float64{currency_id: amount}
	}

	myRefsAmounts := make(map[int64]myRefsType)
	for refUserId, refData := range refs {
		data, err := c.OneRow("SELECT * FROM miners_data WHERE user_id  =  ?", refUserId).String()
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		// получим ID майнеров, у которых лежат фото нужного нам юзера
		if len(data) == 0 {
			continue
		}
		minersIds := utils.GetMinersKeepers(data["photo_block_id"], data["photo_max_miner_id"], data["miners_keepers"], true)
		if len(minersIds) > 0 {
			hosts, err := c.GetList("SELECT http_host FROM miners_data WHERE miner_id  IN (" + utils.JoinInts(minersIds, ",") + ")").String()
			if err != nil {
				return "", utils.ErrInfo(err)
			}
			myRefsAmounts[refUserId] = myRefsType{Amounts: refData, Hosts: hosts}
			refPhotos[refUserId] = hosts
		}
	}
	myRefs := make(map[int64]myRefsType)
	for refUserId, refData := range myRefsAmounts {
		myRefs[refUserId] = refData
	}
	for refUserId, refData := range myRefsKeys {
		md5 := string(utils.Md5(refData["private_key"]))
		myRefs[refUserId] = myRefsType{Key: refData["private_key"], KeyUrl: "public/" + md5[0:16]}
	}

	/*
	 * Общая стата по рефам
	 */
	globalRefs := make(map[int64]globalRefsType)
	// берем лидеров по USD
	rows, err = c.Query(c.FormatQuery(`
			SELECT user_id, sum(amount) as amount
			FROM referral_stats
			WHERE currency_id = 72
			GROUP BY user_id
			ORDER BY amount DESC
			`))
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	defer rows.Close()
	for rows.Next() {
		var user_id int64
		var amount float64
		err = rows.Scan(&user_id, &amount)
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		// вся прибыль с рефов у данного юзера
		refAmounts, err := c.GetAll(`
				SELECT ROUND(sum(amount)) as amount,  currency_id
				FROM referral_stats
				WHERE user_id = ?
				GROUP BY currency_id
				`, -1, user_id)
		if err != nil {
			return "", utils.ErrInfo(err)
		}

		data, err := c.OneRow("SELECT * FROM miners_data WHERE user_id  =  ?", user_id).String()
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		// получим ID майнеров, у которых лежат фото нужного нам юзера
		minersIds := utils.GetMinersKeepers(data["photo_block_id"], data["photo_max_miner_id"], data["miners_keepers"], true)
		hosts, err := c.GetList("SELECT http_host FROM miners_data WHERE miner_id  IN (" + utils.JoinInts(minersIds, ",") + ")").String()
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		globalRefs[user_id] = globalRefsType{Amounts: refAmounts, Hosts: hosts}
		refPhotos[user_id] = hosts
	}

/*	lastTx, err := c.GetLastTx(c.SessUserId, utils.TypesToIds([]string{"NewUser"}), 1, c.TimeFormat)
	lastTxFormatted := ""
	if len(lastTx) > 0 {
		lastTxFormatted, _ = utils.MakeLastTx(lastTx, c.Lang)
	}*/

	TemplateStr, err := makeTemplate("new_user", "newUser", &newUserPage{
		Alert:           c.Alert,
		Lang:            c.Lang,
		CountSignArr:    c.CountSignArr,
		ShowSignData:    c.ShowSignData,
		UserId:          c.SessUserId,
		TimeNow:         timeNow,
		TxType:          txType,
		TxTypeId:        txTypeId,
		SignData:        "",
//		LastTxFormatted: lastTxFormatted,
		MyRefs:          myRefs,
		GlobalRefs:      globalRefs,
		CurrencyList:    c.CurrencyList,
		RefPhotos:       refPhotos,
		PoolUrl:         c.NodeConfig["pool_url"]})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}

type myRefsType struct {
	Amounts map[int64]float64
	Hosts   []string
	Key     string
	KeyUrl  string
}

type globalRefsType struct {
	Amounts []map[string]string
	Hosts   []string
}
