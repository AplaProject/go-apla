package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
	"html/template"
	"io/ioutil"
	"net"
	"os"
	"strings"
	"time"
)

type upgrade6Page struct {
	SignData        string
	ShowSignData    bool
	Alert           string
	UserId          int64
	Lang            map[string]string
	HttpHost        string
	TcpHost         string
	SaveAndGotoStep string
	UpgradeMenu     string
	Community       bool
	HostType        string
	NodePrivateKey  string
	CountSignArr    []int
	ProfileHash     string
	FaceHash        string
	Pools           template.JS
	VideoHash       string
	Mobile          bool
	VideoUrlId      string
	PoolAdminUserId string
}

func (c *Controller) Upgrade6() (string, error) {

	log.Debug("Upgrade6")

	var hostType string

	hostData, err := c.OneRow("SELECT http_host, tcp_host FROM " + c.MyPrefix + "my_table").String()
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	hostData["PoolAdminUserId"] = "0"

	// в режиме пула выдаем только хост ноды
	log.Debug("c.Community: %v", c.Community)
	log.Debug("c.PoolAdminUserId: %v", c.PoolAdminUserId)
	if c.Community /*&& len(data["http_host"]) == 0 && len(data["tcp_host"]) == 0*/ {
		hostType = "pool"
		hostData, err = c.OneRow("SELECT http_host, tcp_host FROM miners_data WHERE user_id  =  ?", c.PoolAdminUserId).String()
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		if len(hostData) == 0 {
			hostData["http_host"] = "null http_host in miners_data"
			hostData["tcp_host"] = "null tcp_host in miners_data"
		}
		hostData["PoolAdminUserId"] = utils.Int64ToStr(c.PoolAdminUserId)
	} else {
		// если смогли подключиться из вне
		ip, err := utils.GetHttpTextAnswer("http://api.ipify.org")
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		/*httpHost, err := c.Single("SELECT http_host FROM "+c.MyPrefix+"my_table").String()
		if err!=nil {
			return "", utils.ErrInfo(err)
		}
		port := "8089"
		if len(httpHost) > 0 {
			re := regexp.MustCompile(`https?:\/\/(?:[0-9a-z\_\.\-]+):([0-9]+)`)
			match := re.FindStringSubmatch(httpHost)
			if len(match) != 0 {
				port = match[1];
			}
		}*/
		conn, err := net.DialTimeout("tcp", ip+":8089", 3*time.Second)
		log.Debug("ip: %v", ip)
		if err != nil {
			// если не смогли подключиться, то в JS будем искать рабочий пул и региться на нем. и дадим юзеру указать другие хост:ip
			hostType = "findPool"

		} else {
			hostType = "normal"
			defer conn.Close()
			hostData["http_host"] = ip + ":8089"
			hostData["tcp_host"] = ip + ":8088"
		}
	}

	// проверим, есть ли необработанные ключи в локальной табле
	nodePrivateKey, err := c.Single("SELECT private_key FROM " + c.MyPrefix + "my_node_keys WHERE block_id  =  0").String()
	if err != nil {
		return "", utils.ErrInfo(err)
	}

	if len(nodePrivateKey) == 0 {
		//  сгенерим ключ для нода
		priv, pub := utils.GenKeys()
		err = c.ExecSql("INSERT INTO "+c.MyPrefix+"my_node_keys ( public_key, private_key ) VALUES ( [hex], ? )", pub, priv)
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		nodePrivateKey = priv
	}

	var profileHash, faceHash, videoHash string

	if _, err := os.Stat(*utils.Dir + "/public/" + utils.Int64ToStr(c.SessUserId) + "_user_face.jpg"); err == nil {
		file, err := ioutil.ReadFile(*utils.Dir + "/public/" + utils.Int64ToStr(c.SessUserId) + "_user_face.jpg")
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		faceHash = string(utils.DSha256(file))
	}
	if _, err := os.Stat(*utils.Dir + "/public/" + utils.Int64ToStr(c.SessUserId) + "_user_profile.jpg"); err == nil {
		file, err := ioutil.ReadFile(*utils.Dir + "/public/" + utils.Int64ToStr(c.SessUserId) + "_user_profile.jpg")
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		profileHash = string(utils.DSha256(file))
	}

	if _, err := os.Stat(*utils.Dir + "/public/" + utils.Int64ToStr(c.SessUserId) + "_user_video.mp4"); err == nil {
		file, err := ioutil.ReadFile(*utils.Dir + "/public/" + utils.Int64ToStr(c.SessUserId) + "_user_video.mp4")
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		videoHash = string(utils.DSha256(file))
	}


	rows, err := c.Query(c.FormatQuery(`
			SELECT user_id, http_host
			FROM miners_data
			WHERE i_am_pool = 1 AND pool_count_users < ?`), c.Variables.Int64["max_pool_users"])
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	defer rows.Close()
	pools := make(map[string]string)
	for rows.Next() {
		var user_id, http_host string
		err = rows.Scan(&user_id, &http_host)
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		pools[user_id] = http_host
	}
	poolsJs := ""
	for userId, httpHost := range pools {
		poolsJs = poolsJs + "[" + userId + ",'" + httpHost + "'],"
	}
	if len(poolsJs) > 0 {
		poolsJs = poolsJs[:len(poolsJs)-1]
	}
	videoUrlId, err := c.Single("SELECT video_url_id FROM " + c.MyPrefix + "my_table").String()
	if err != nil {
		return "", utils.ErrInfo(err)
	}

	upgradeMenu,_,next := utils.MakeUpgradeMenu(5)
	saveAndGotoStep := strings.Replace(c.Lang["save_and_goto_step"], "[num]", next, -1)

	TemplateStr, err := makeTemplate("upgrade_6", "upgrade6", &upgrade6Page{
		Alert:           c.Alert,
		Lang:            c.Lang,
		SaveAndGotoStep: saveAndGotoStep,
		UpgradeMenu:     upgradeMenu,
		ShowSignData:    c.ShowSignData,
		SignData:        "",
		CountSignArr:    c.CountSignArr,
		HttpHost:        hostData["http_host"],
		TcpHost:         hostData["tcp_host"],
		PoolAdminUserId:         hostData["PoolAdminUserId"],
		Community:       c.Community,
		HostType:        hostType,
		ProfileHash:     profileHash,
		FaceHash:        faceHash,
		VideoHash:       videoHash,
		NodePrivateKey:  nodePrivateKey,
		Pools:           template.JS(poolsJs),
		UserId:          c.SessUserId,
		VideoUrlId:      videoUrlId,
		Mobile:          utils.Mobile()})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
