package controllers

import (
	"bytes"
	"github.com/DayLightProject/go-daylight/packages/static"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"html/template"
	"regexp"
	"strings"
)

type menuPage struct {
	MyModalIdName  string
	SetupPassword  bool
	Lang           map[string]string
	LangInt        int64
	PoolAdmin      bool
	Community      bool
	MinerId        int64
	Name           string
	UserId         int64
	DaemonsStatus  string
	MyNotice       map[string]string
	BlockId        int64
	Avatar         string
	NoAvatar       string
	FaceUrls       string
	Restricted     int64
	Mobile         bool
	ExchangeEnable bool
	Admin          bool
	Notifications  int64
	Desktop bool
	Pct               float64
	Amount            float64
	IsRestricted      bool
	Wallets           []utils.DCAmounts
	CurrencyList       map[int64]string
}

func (c *Controller) Menu() (string, error) {

	if !c.dbInit || c.SessUserId == 0 {
		return "", nil
	}

	status, err := c.DCDB.Single("SELECT status FROM " + c.MyPrefix + "my_table").String()
	if err != nil {
		log.Error("%v", err)
	}
	if status == "waiting_set_new_key" || status == "waiting_accept_new_key" {
		return "", nil
	}

	var name, avatar string
	if c.SessUserId > 0 {
		data, err := c.OneRow("SELECT name, avatar FROM users WHERE user_id =  ?", c.SessUserId).String()
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		name, avatar = data["name"], data["avatar"]
	}

	var face_urls []string
	if len(avatar) == 0 {
		data, err := c.OneRow("SELECT photo_block_id, photo_max_miner_id, miners_keepers FROM miners_data WHERE user_id  =  ?", c.SessUserId).String()
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		if len(data) > 0 {
			// получим ID майнеров, у которых лежат фото нужного нам юзера
			minersIds := utils.GetMinersKeepers(data["photo_block_id"], data["photo_max_miner_id"], data["miners_keepers"], true)
			if len(minersIds) > 0 {
				hosts, err := c.GetList("SELECT http_host as host FROM miners_data WHERE miner_id IN (" + utils.JoinInts(minersIds, ",") + ")").String()
				if err != nil {
					return "", utils.ErrInfo(err)
				}
				for i := 0; i < len(hosts); i++ {
					face_urls = append(face_urls, hosts[i]+"public/face_"+utils.Int64ToStr(c.SessUserId)+".jpg")
				}
			}
		}
	}

	noAvatar := "static/img/noavatar.png"
	minerId, err := c.GetMinerId(c.SessUserId)
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	if len(name) == 0 {
		if minerId > 0 {
			name = "ID " + utils.Int64ToStr(c.SessUserId) + " (miner)"
		} else {
			name = "ID " + utils.Int64ToStr(c.SessUserId)
		}
	}
	// ID блока вверху
	blockId, err := c.GetBlockId()

	daemonsStatus := ""
	if !c.Community {
		scriptName, err := c.Single("SELECT script_name FROM main_lock").String()
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		if scriptName == "my_lock" {
			daemonsStatus = `<li title="` + c.Lang["daemons_status_off"] + `"><a href="#" id="start_daemons" style="color:#C90600"><i class="fa fa-power-off" style="font-size: 20px"></i></a></li>`
		} else {
			daemonsStatus = `<li title="` + c.Lang["daemons_status_on"] + `"><a href="#" id="stop_daemons" style="color:#009804"><i class="fa fa-power-off" style="font-size: 20px"></i></a></li>`
		}
	}

	data, err := static.Asset("static/templates/menu.html")
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	log.Debug("menu ok : %d", len(data))
/*	modal, err := static.Asset("static/templates/modal.html")
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	log.Debug("modal ok : %d", len(modal))*/

	defer func() {
		if r := recover(); r != nil {
			log.Error("Recovered", r)
		}
	}()

	mobile := utils.Mobile()
	if ok, _ := regexp.MatchString("(?i)(iPod|iPhone|iPad|Android)", c.r.UserAgent()); ok {
		mobile = true
	}

	var admin bool
	if c.SessUserId == 1 {
		admin = true
	}

	var exchangeEnable bool
	exchangeEnable_, err := c.Single(`SELECT value FROM e_config WHERE name='enable'`).Int64()
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	if exchangeEnable_ == 1 {
		exchangeEnable = true
	}

	var ( isRestricted bool
		  profit, pct float64
	)
	myUserId, err := c.Single("SELECT user_id FROM miners_data WHERE user_id  =  ?", c.SessUserId).Int64()
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	if myUserId > 0 {
		idPromised,err := c.Single("SELECT id FROM promised_amount WHERE user_id = ? AND currency_id = ? AND status != 'pending' AND status != 'rejected'",
		                              c.SessUserId, 72 ).Int64()
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		if idPromised == 0 {
			profit, pct, err = c.GetPromisedAmountCounter(c.SessUserId)
			if err != nil {
				return "", utils.ErrInfo(err)
			}
			if profit > 0 {
				isRestricted = true
			}
		}
	}
	notifications,err := c.GetNotificationsCount(c.SessUserId)
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	currencyList,err := c.GetCurrencyList(true)
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	var wallets []utils.DCAmounts
	if c.SessUserId > 0 {
		wallets, err = c.GetBalances(c.SessUserId)
	}
		
	funcMap := template.FuncMap{
		"noescape": func(s string) template.HTML {
			return template.HTML(s)
		},
	}
	t := template.Must(template.New("template").Funcs(funcMap).Parse(string(data)))
//	t = template.Must(t.Parse(string(modal)))
	b := new(bytes.Buffer)
	err = t.ExecuteTemplate(b, "menu", &menuPage{Desktop: utils.Desktop(), Admin: admin, 
			ExchangeEnable: exchangeEnable, Mobile: mobile, SetupPassword: false, 
			MyModalIdName: "myModal", Lang: c.Lang, PoolAdmin: c.PoolAdmin, 
			Community: c.Community, MinerId: minerId, Name: name, LangInt: c.LangInt, 
			UserId: c.SessUserId, Restricted: c.SessRestricted, DaemonsStatus: daemonsStatus, 
			MyNotice: c.MyNotice, BlockId: blockId, Avatar: avatar, NoAvatar: noAvatar, 
			FaceUrls: strings.Join(face_urls, ","),
			Notifications:     notifications,
			IsRestricted:      isRestricted,
			Amount:            profit,
			Wallets:           wallets,
			CurrencyList:      currencyList,
			Pct:               pct })
	if err != nil {
		log.Error("%s", utils.ErrInfo(err))
		return "", utils.ErrInfo(err)
	}
	log.Debug("b.String():\n %s", b.String())
	return b.String(), nil
}
