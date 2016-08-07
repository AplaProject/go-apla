package controllers

import (
	"bytes"
	"github.com/DayLightProject/go-daylight/packages/static"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"html/template"
	"regexp"
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
	WalletId       int64
	CitizenId      int64
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
	Desktop        bool
	Pct            float64
	Amount         float64
	IsRestricted   bool
	Wallets        []utils.DCAmounts
	CurrencyList   map[int64]string
}

func (c *Controller) Menu() (string, error) {

	if !c.dbInit || (c.SessWalletId == 0 && c.SessCitizenId == 0) {
		return "", nil
	}

	status, err := c.DCDB.Single("SELECT status FROM " + c.MyPrefix + "my_table").String()
	if err != nil {
		log.Error("%v", err)
	}
	if status == "waiting_set_new_key" || status == "waiting_accept_new_key" {
		return "", nil
	}

	// ID блока вверху
	blockId, err := c.GetBlockId()

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

	var exchangeEnable bool
	exchangeEnable_, err := c.Single(`SELECT value FROM e_config WHERE name='enable'`).Int64()
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	if exchangeEnable_ == 1 {
		exchangeEnable = true
	}

	var (
		isRestricted bool
		pct          float64
	)

	currencyList, err := c.GetCurrencyList(true)
	if err != nil {
		return "", utils.ErrInfo(err)
	}

	funcMap := template.FuncMap{
		"noescape": func(s string) template.HTML {
			return template.HTML(s)
		},
	}
	t := template.Must(template.New("template").Funcs(funcMap).Parse(string(data)))
	//	t = template.Must(t.Parse(string(modal)))
	b := new(bytes.Buffer)
	err = t.ExecuteTemplate(b, "menu", &menuPage{Desktop: utils.Desktop(),
		ExchangeEnable: exchangeEnable, Mobile: mobile, SetupPassword: false,
		MyModalIdName: "myModal", Lang: c.Lang, PoolAdmin: c.PoolAdmin,
		Community: c.Community, LangInt: c.LangInt,
		WalletId: c.SessWalletId, CitizenId: c.SessCitizenId,
		MyNotice: c.MyNotice, BlockId: blockId,
		IsRestricted: isRestricted,
		CurrencyList: currencyList,
		Pct:          pct})
	if err != nil {
		log.Error("%s", utils.ErrInfo(err))
		return "", utils.ErrInfo(err)
	}
	log.Debug("b.String():\n %s", b.String())
	return b.String(), nil
}
