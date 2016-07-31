// emailserv
package main

import (
	"crypto/md5"
/*	"crypto/rsa"
	"crypto/x509"*/
	"encoding/base64"
//	"golang.org/x/crypto/bcrypt"
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/config"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"html"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"html/template"
	//	"regexp"
	//	"net/url"
	"strings"
)

const (
	XForwardedFor = "X-Forwarded-For"
	XRealIP       = "X-Real-IP"
)

type Settings struct {
	Port      uint32 `json:"port"`
	Path      string `json:"path"`
	ApiId     string `json:"api_id"`
	ApiSecret string `json:"api_secret"`
	FromName  string `json:"from_name"`
	FromEmail string `json:"from_email"`
	Password  string `json:"password"`
	Admin     string `json:"admin"`
	CopyTo    string `json:"copy_to"`
	WhiteList []string `json:"white"`
}

var (
	GSettings Settings
	GDB       *utils.DCDB
	GEmail    *EmailClient
	GPageTpl  *template.Template
	GPagePattern  *template.Template
	GLatest    map[int]int64
)

func getIP(r *http.Request) (uint32, string) {
	var ipval uint32

	remoteAddr := r.RemoteAddr
	var ip string
	if ip = r.Header.Get(XRealIP); len(ip) > 6 {
		remoteAddr = ip
	} else if ip = r.Header.Get(XForwardedFor); len(ip) > 6 {
		remoteAddr = ip
	}
	if strings.Contains(remoteAddr, ":") {
		remoteAddr, _, _ = net.SplitHostPort(remoteAddr)
	}
	if ipb := net.ParseIP(remoteAddr).To4(); ipb != nil {
		ipval = uint32(ipb[3]) | (uint32(ipb[2]) << 8) |
			(uint32(ipb[1]) << 16) | (uint32(ipb[0]) << 24)
	}
	return ipval,remoteAddr
}

func AddToStopList( email string, userId int64 ) error {
	if userId == 0 {
		userId,_ = GDB.Single(`select user_id from users where email=?`, email ).Int64()
	}
	if userId > 0 {
//		return fmt.Errorf(`Unknown user_id for `.email )
		GDB.ExecSql(`UPDATE users SET verified=? WHERE user_id=?`, -1, userId )
	}

	isStop, _ := GDB.Single(`SELECT id FROM stoplist where email=?`, email).Int64()
	if isStop == 0 {
		log.Println( `Auto ban: `, email, userId )
		GDB.ExecSql(`INSERT INTO stoplist ( email, error, uptime, ip )
					VALUES ( ?, ?, datetime('now'), ? )`,
				   email, `Auto ban`, 1 )
	}
	
	return nil
}

func emailHandler(w http.ResponseWriter, r *http.Request) {
	var (
		jsonEmail             utils.JsonEmail
		err                   error
		/*publicKey,*/
	)

	answer := utils.Answer{false, ``}
	ipval, remoteAddr := getIP( r )

	result := func(msg string) {

		answer.Error = msg
		if !answer.Success {
			if len(jsonEmail.Email) == 0 {
				jsonEmail.Email = r.FormValue(`email`)
			}
			log.Println(remoteAddr, `Error:`, jsonEmail.Cmd, answer.Error, jsonEmail.Email, jsonEmail.UserId)
		} else {
			log.Println(remoteAddr, `Sent:`, jsonEmail.Cmd, jsonEmail.Email, jsonEmail.UserId)
		}

		ret, err := json.Marshal(answer)
		if err != nil {
			ret = []byte(`{"success": false,
"error":"Unknown error"}`)
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		//	w.WriteHeader(200)
		w.Write(ret)
	}
	checkParams := func(params ...string) error {
		for _, name := range params {
			if _, ok := (*jsonEmail.Params)[name]; !ok {
				return fmt.Errorf(`Empty %s parameter`, name)
			}
			(*jsonEmail.Params)[name] = html.EscapeString((*jsonEmail.Params)[name])
		}
		return nil
	}

	iplog, err := GDB.Single(`select count(id) from log where ip=? AND date( uptime, '+1 hour' ) > datetime('now')`, 
	                     ipval ).Int64()
	if err!=nil {
		log.Println("SQL Error", err )
	} else if iplog > 10 {
		result(`Anti-spam`)
		return
	}
	
	r.ParseForm()

	if len(r.URL.Path[1:]) > 0 || r.Method != `POST` {
		result(`Wrong method or path`)
		return
	}

	data := r.FormValue(`data`)
//	sign := r.FormValue(`sign`)
	if err = json.Unmarshal([]byte(data), &jsonEmail); err != nil ||
		jsonEmail.UserId == 0 || jsonEmail.Cmd == 0 {
		result(`Incorrect data`)
		return
	}
	var email string
	if jsonEmail.UserId != utils.EXCHANGE_USER {
		user,_ := GDB.OneRow(`SELECT * FROM users WHERE user_id=?`, jsonEmail.UserId ).String()
		if len(user[`user_id`]) > 0 {
			if utils.StrToInt(user[`verified`]) < 0 {
				result(`Stop list`)
				return
			}
			email = user[`email`]
			if len(jsonEmail.Email) > 0 && len(email)>0 && email != jsonEmail.Email {
				if jsonEmail.Cmd == utils.ECMD_NEW || jsonEmail.Cmd == utils.ECMD_SIGNUP {
					if err = GDB.ExecSql(`update users set newemail = '*' + email, email=?, verified=0 where user_id=?`, 
										jsonEmail.Email, jsonEmail.UserId ); err!=nil {
						log.Println(remoteAddr, `Error re-email user:`, err, jsonEmail.Email)
					}
				} else {
					result(`Overwrite email`)
					return
				}
				jsonEmail.Email = email
			}
		}
	}
	if len(jsonEmail.Email) == 0 && len(email) > 0 {
		jsonEmail.Email = email
	}

	//	re := regexp.MustCompile( `^([a-z0-9_\-]+\.)*[a-z0-9_\-]+@([a-z0-9][a-z0-9\-]*[a-z0-9]\.)+[a-z]{2,4}$` )
	//	if !re.MatchString( email ) {
	if !utils.ValidateEmail(jsonEmail.Email) {
		result(`Incorrect email`)
		return
	}
/*
	if publicKey, err = utils.DB.GetUserPublicKey(jsonEmail.UserId); err != nil || len(publicKey) == 0 {
		pubVal := r.FormValue(`public`)
		if (jsonEmail.Cmd == utils.ECMD_TEST || jsonEmail.Cmd == utils.ECMD_NEW) && len(pubVal) > 0 {
			public, _ := base64.StdEncoding.DecodeString(pubVal)
			publicKey = string(public)
		} else {
			result(`Incorrect user_id or public_key`)
			return
		}
	}
	//	fmt.Println(jsonEmail)
	signature, _ := base64.StdEncoding.DecodeString(sign)
	var re interface{}
	if re, err = x509.ParsePKIXPublicKey([]byte(publicKey)); err != nil {
		result(err.Error())
		return
	}
	if err = rsa.VerifyPKCS1v15(re.(*rsa.PublicKey), crypto.SHA1, utils.HashSha1(data),
		signature); err != nil {
		result(err.Error())
		return
	}*/
/*	var (
		params        []byte
	)*/
	if jsonEmail.Params == nil {
		jsonEmail.Params = &map[string]string{}
	}
/*	if len(*jsonEmail.Params) > 0 {
		params, _ = json.Marshal(jsonEmail.Params)
	}*/
	switch jsonEmail.Cmd {
	case utils.ECMD_NEW, utils.ECMD_TEST, utils.ECMD_SIGNUP, utils.ECMD_UPDPRIMARY, utils.ECMD_VOTETIME:
	case utils.ECMD_UPDEMAIL:
		if err := checkParams(`email`); err != nil {
			result(err.Error())
			return
		}
	case utils.ECMD_ADMINMSG:
		if err := checkParams(`msg`); err != nil {
			result(err.Error())
			return
		}
	case utils.ECMD_VOTERES:
		if err := checkParams(`text`); err != nil {
			result(err.Error())
			return
		}
	case utils.ECMD_UPDSMS:
		if err := checkParams(`sms`); err != nil {
			result(err.Error())
			return
		}
		(*jsonEmail.Params)[`sms`] = `New sms_http_get_request ` + (*jsonEmail.Params)[`sms`]
/*	Отправляем через daemon - notifications
	case utils.ECMD_CHANGESTAT:
		if err := checkParams(`status`); err != nil {
			result(err.Error())
			return
		}
		text = `New status: ` + (*jsonEmail.Params)[`status`]
	case utils.ECMD_DCCAME:
		if err := checkParams(`amount`, `currency`, `comment`); err != nil {
			result(err.Error())
			return
		}
	case utils.ECMD_DCSENT:
		if err := checkParams(`amount`, `currency`); err != nil {
			result(err.Error())
			return
		}*/
	case utils.ECMD_NODETIME:
		if err := checkParams(`dif`); err != nil {
			result(err.Error())
			return
		}
	case utils.ECMD_EXREQUEST,utils.ECMD_EXANSWER:
		if err := checkParams(`exchange`); err != nil {
			result(err.Error())
			return
		}
	case utils.ECMD_SENDKEY:
		if err := checkParams(`subject`,`text`, `txt_key`, `refid`); err != nil || 
			len((*jsonEmail.Params)[`subject`]) == 0 || len((*jsonEmail.Params)[`text`]) == 0 ||
			len((*jsonEmail.Params)[`txt_key`]) == 0 || utils.StrToInt64((*jsonEmail.Params)[`refid`]) == 0 {
			result( `Wrong email parameters` )
			return
		}
	case utils.ECMD_FORKBLOCK:
		if err := checkParams(`forks`); err != nil || len((*jsonEmail.Params)[`forks`]) == 0  {
			result( `Wrong email parameters` )
			return
		}
	default:
		result(fmt.Sprintf(`Unknown command %d`, jsonEmail.Cmd))
		return
	}
	isStop, _ := GDB.Single(`SELECT id FROM stoplist where email=?`, jsonEmail.Email).Int64()
	if isStop != 0 {
		result(fmt.Sprintf(`Email %s is in the stop-list`, jsonEmail.Email))
		return
	}
	if (jsonEmail.Cmd == utils.ECMD_NEW || jsonEmail.Cmd == utils.ECMD_SIGNUP ) && len(email)==0 {
		if err = GDB.ExecSql(`INSERT INTO users (user_id, email, newemail, verified, code, lang ) VALUES(?,?,'', 0, 0, 0)`, 
								jsonEmail.UserId, jsonEmail.Email ); err!=nil {
			log.Println(remoteAddr, `Error new user:`, err, jsonEmail.Email)
		}
	}
	if jsonEmail.Cmd == utils.ECMD_SENDKEY {
		bcc := GSettings.CopyTo
		GSettings.CopyTo = ``
		files := make( map[string][]byte )
		decoded, err := base64.StdEncoding.DecodeString((*jsonEmail.Params)[`txt_key`])
		if err != nil {
			result(err.Error())
			return
		}
		files[`dcoin-private-key-`+(*jsonEmail.Params)[`refid`]+`.txt`] = decoded
		decoded, err = base64.StdEncoding.DecodeString((*jsonEmail.Params)[`png_key`])
		if err != nil {
			result(err.Error())
			return
		}
		files[`dcoin-private-key-`+(*jsonEmail.Params)[`refid`]+`.png`] = decoded
		err = GEmail.SendEmailAttach(`<p>`+(*jsonEmail.Params)[`text`]+`</p>`, ``, (*jsonEmail.Params)[`subject`],
		      []*Email{ &Email{``, jsonEmail.Email }}, &files )
		GSettings.CopyTo = bcc

		if err != nil {
			result(err.Error())
			return
		}
	} else if jsonEmail.Cmd == utils.ECMD_FORKBLOCK {
		err = GEmail.SendEmail(`<p>Fork of blockchain has been detected<br>`+(*jsonEmail.Params)[`forks`]+`</p>`, ``, 
		     `Fork of blockchain`, []*Email{ &Email{``, jsonEmail.Email }})
		if err != nil {
			result(err.Error())
			return
		}
	} else if jsonEmail.Cmd != utils.ECMD_SIGNUP {
		data, err := CheckUser( jsonEmail.UserId )
		if err != nil {
			result( fmt.Sprintf(`EmailCheck %s`, err))
			return
		}
		for key, value := range *jsonEmail.Params {
			data[strings.ToUpper(key[:1]) + key[1:]] = value
		}
		if _,ok := data[`Msg`]; ok {
			data[`Msg`] = template.HTML(data[`Msg`].(string))
		}
		if _,ok := data[`Text`]; ok {
			data[`Text`] = template.HTML(data[`Text`].(string))
		}
		if jsonEmail.UserId == utils.EXCHANGE_USER {
			data[`email`] = jsonEmail.Email
		}
		if !EmailUser( jsonEmail.UserId, data, int(jsonEmail.Cmd) ) {
			result( `EmailUser`)
			return
		}
	}
	answer.Success = true

	result(``)
}
/*
func Send() {
	time.Sleep( 5 * time.Second )

	Client := &http.Client{
			Transport: http.DefaultTransport,
			Timeout:   20 * time.Second,
		}
	values := url.Values{}
	values.Set("email", "@mail.ru")
	values.Set("user_id", "1001" )
	values.Set("text", "Test" )
	values.Set("subject", "Test" )
	req, err := http.NewRequest("POST", `http://localhost:8090/setemail`,
	                          strings.NewReader(values.Encode()))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res, e := Client.Do(req)
	if e != nil {
		return
	}

	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	var answer Answer
	err = json.Unmarshal( body, &answer )
	if err != nil {
		return
	}
    fmt.Println( answer )
}

func Send() {
	time.Sleep(5 * time.Second)
	emailhere := ``
	var userId int64 = 0
	fmt.Println("Result", utils.SendEmail(emailhere, userId, utils.ECMD_TEST, nil ))
	fmt.Println("Result", utils.SendEmail(emailhere, userId, utils.ECMD_ADMINMSG, &map[string]string{`msg`: `<h1>Header</h1>`}))
	fmt.Println("Result", utils.SendEmail(`emailhere`, 3, utils.ECMD_CASHREQ, &map[string]string{`amount`: `<h1>Header</h1>`, `currency`: `USD`}))
	fmt.Println("Result", utils.SendEmail(`emailhere`, 3, utils.ECMD_CHANGESTAT, &map[string]string{`status`: `miner`}))
	fmt.Println("Result", utils.SendEmail(emailhere, userId, utils.ECMD_DCCAME, &map[string]string{ `amount`: `10`,
	                                  `currency`: `USD`, `comment`: `<h1>Header</h1>` }))
	fmt.Println("Result", utils.SendEmail(emailhere, userId, utils.ECMD_DCSENT, &map[string]string{`amount`: `111`, `currency`: `USD`}))
	fmt.Println("Result", utils.SendEmail(emailhere, userId, utils.ECMD_UPDPRIMARY, nil ))
//	fmt.Println("Result", utils.SendEmail(emailhere, userId, utils.ECMD_UPDEMAIL, &map[string]string{ `email`: `my@newemail.com` }))
//	fmt.Println("Result", utils.SendEmail(emailhere, userId, utils.ECMD_UPDSMS, &map[string]string{ `sms`: `my SMS` }))
//	fmt.Println("Result", utils.SendEmail(emailhere, userId, utils.ECMD_VOTERES, &map[string]string{ `text`: `Voting result` }))
//	fmt.Println("Result", utils.SendEmail(emailhere, userId, utils.ECMD_VOTETIME, nil ))
//	fmt.Println("Result", utils.SendEmail(emailhere, userId, utils.ECMD_NODETIME, &map[string]string{ `dif`: `7` } ))
}
*/

func checkLogin( w http.ResponseWriter, req *http.Request ) ( ip uint32, ips string, ok bool ) {
	ip, ips = getIP( req )
	if phash,err := req.Cookie(`admpass`); err == http.ErrNoCookie || 
	          fmt.Sprintf("%x", md5.Sum([]byte(GSettings.Password))) != phash.Value {
		http.Redirect(w, req, `/` + GSettings.Admin + `/login`, http.StatusFound)
		return
	}
	ok = true
	return
}

func main() {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(`Dir`, err)
	}
	//	os.Chdir(dir)
	logfile, err := os.OpenFile(filepath.Join(dir, "email.log"),
		os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln(`Email log`, err)
	}
	defer logfile.Close()
	log.SetOutput(logfile)
	params, err := ioutil.ReadFile(filepath.Join(dir, `settings.json`))
	if err != nil {
		log.Fatalln(dir, `Settings.json`, err)
	}
	if err = json.Unmarshal(params, &GSettings); err != nil {
		log.Fatalln(`Unmarshall`, err)
	}
	if len(GSettings.ApiId) == 0 || len(GSettings.ApiSecret) == 0 ||
		len(GSettings.FromEmail) == 0 {
		log.Fatalln(`api_id, api_secret, from_email are not defined`)
	}
	if err = os.Chdir(GSettings.Path); err != nil {
		log.Fatalln(`Chdir`, err)
	}
	if GDB, err = utils.NewDbConnect(map[string]string{
		"db_name": "", "db_password": ``, `db_port`: ``,
		`db_user`: ``, `db_host`: ``, `db_type`: `sqlite`}); err != nil {
		log.Fatalln(`Connect`, err)
	}

	*utils.Dir = GSettings.Path
	configIni := make(map[string]string)
	configIni_, err := config.NewConfig("ini", `config.ini`)
	if err != nil {
		log.Fatalln(`Config`, err)
	} else {
		configIni, err = configIni_.GetSection("default")
	}
	if utils.DB, err = utils.NewDbConnect(configIni); err != nil {
		log.Fatalln(`Utils connect`, err)
	}
	
	var list []string
	if list, err = GDB.GetAllTables(); err == nil && len(list) == 0 {
		if err = GDB.ExecSql(`CREATE TABLE log (
	id	INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
	user_id	INTEGER NOT NULL,
	email	TEXT NOT NULL,
	cmd     INTEGER NOT NULL,
	params  TEXT NOT NULL,
	ip	INTEGER NOT NULL,
	uptime	INTEGER NOT NULL
	)`); err != nil {
			//	verified	INTEGER NOT NULL,
			//	code	INTEGER NOT NULL,

			log.Fatalln(err)
		}
		if err = GDB.ExecSql(`CREATE INDEX userid ON log (user_id)`); err != nil {
			log.Fatalln(err)
		}
		if err = GDB.ExecSql(`CREATE INDEX ip ON log (ip)`); err != nil {
			log.Fatalln(err)
		}
		if err = GDB.ExecSql(`CREATE TABLE stoplist (
	id	INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
	email	TEXT NOT NULL,
	error   TEXT NOT NULL,
	ip	INTEGER NOT NULL,
	uptime	INTEGER NOT NULL
	)`); err != nil {
			log.Fatalln(err)
		}
		if err = GDB.ExecSql(`CREATE INDEX email ON stoplist (email)`); err != nil {
			log.Fatalln(err)
		}
	} 
	
	if !utils.InSliceString(`users`, list ) || len(list) == 0 {
		if err = GDB.ExecSql(`CREATE TABLE users (
	id	INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
	user_id	INTEGER NOT NULL,
	email	 TEXT NOT NULL,
	newemail TEXT NOT NULL,
	lang	 INTEGER NOT NULL,	
	code	 INTEGER NOT NULL,	
	verified  INTEGER NOT NULL
	)`); err != nil {
			log.Fatalln(err)
		}
		if err = GDB.ExecSql(`CREATE INDEX usersid ON users (user_id)`); err != nil {
			log.Fatalln(err)
		}

		curlist, _ := GDB.GetAll(`SELECT DISTINCT user_id, email FROM log order by user_id, email`, -1 )
		prev := 0
		for _, curi := range curlist {
			if prev == utils.StrToInt( curi[`user_id`] ) {
				continue
			}
			prev = utils.StrToInt( curi[`user_id`] )
			
			verified := 0
			if isStop, _ := GDB.Single(`SELECT id FROM stoplist where email=?`, curi[`email`]).Int64(); isStop > 0 {
				verified = -1
			}
			
			if err = GDB.ExecSql(`INSERT INTO users (user_id, email, newemail, verified, code, lang ) VALUES(?,?,'', ?, 0, 0 )`, 
			      curi[`user_id`], curi[`email`], verified ); err!=nil {
				log.Fatalln( err )
			}
		}
	}

	if !utils.InSliceString(`latest`, list ) || len(list) == 0 {
		if err = GDB.ExecSql(`CREATE TABLE latest (
	cmd_id	INTEGER NOT NULL,
	latest	INTEGER NOT NULL
	)`); err != nil {
			log.Fatalln(err)
		}
		if err = GDB.ExecSql(`CREATE INDEX cmdid ON latest (cmd_id)`); err != nil {
			log.Fatalln(err)
		}
	}
	GLatest = make(map[int]int64)
	if curlatest, err := GDB.GetAll(`SELECT * FROM latest`, -1 ); err == nil {
		for _, curi := range curlatest {
			GLatest[ utils.StrToInt(curi[`cmd_id`])] = utils.StrToInt64(curi[`latest`])
		}
		if _, ok := GLatest[utils.ECMD_CASHREQ]; !ok {
			if cash, err := utils.DB.Single(`SELECT max(id) FROM cash_requests` ).Int64(); err==nil {
				 GLatest[utils.ECMD_CASHREQ] = cash
				if err = GDB.ExecSql(`INSERT INTO latest ( cmd_id, latest ) VALUES(?,?)`, utils.ECMD_CASHREQ, cash ); err!=nil {
					log.Fatalln( err )
				}
			} else {
				log.Fatalln(err)
			}
		}
		if _, ok := GLatest[utils.ECMD_DCCAME]; !ok {
			if nfy, err := utils.DB.Single(`SELECT max(id) FROM notifications` ).Int64(); err==nil {
    			 GLatest[utils.ECMD_DCCAME] = nfy
				if err = GDB.ExecSql(`INSERT INTO latest (cmd_id, latest) VALUES(?,?)`, utils.ECMD_DCCAME, nfy ); err!=nil {
					log.Fatalln( err )
				}
			} else {
					log.Fatalln(err)
			}
		}
	} else {
		log.Fatalln( err )
	}
	
	if !utils.InSliceString(`balance`, list ) || len(list) == 0 {
		if err = GDB.ExecSql(`CREATE TABLE balance (
	id	INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
	user_id	INTEGER NOT NULL,
	balance	TEXT NOT NULL,
	uptime	INTEGER NOT NULL
	)`); err != nil {
			log.Fatalln(err)
		}
		if err = GDB.ExecSql(`CREATE INDEX balanceind ON balance (user_id,uptime)`); err != nil {
			log.Fatalln(err)
		}
	}

	os.Chdir(dir)	
	if GPageTpl,err =template.ParseGlob(`template/*.tpl`); err!=nil {
		log.Fatalln( err )
	}
	if GPagePattern,err =template.ParseGlob(`pattern/*.tpl`); err!=nil {
		log.Fatalln( err )
	}
//	answer, err := test()	
//	fmt.Println( answer, err )
/*	if imported, _ := ioutil.ReadFile( `emails.txt` ); len( imported ) > 0 {
		items := strings.Split( string(imported), "\n" )
		for i, cur := range items {
			pars := strings.Split( strings.TrimSpace( cur ), ` ` )
			if len( pars ) >= 3 {
				verified := 0
				if len(pars) == 4 && pars[3] == `-` {
					verified = -1
				}
				userId := utils.StrToInt64( pars[0] )
				email := strings.TrimSpace( pars[2] )
				user,err := GDB.Single(`select user_id from users where user_id=?`, userId ).Int64()
				if err != nil {
					log.Fatalln( err )
				}
				if user == 0 {
					if err = GDB.ExecSql(`INSERT INTO users (user_id, email, newemail, verified, code, lang ) VALUES(?,?,'', ?, 0, 0)`, 
								userId, email, verified ); err!=nil {
						log.Fatalln( err )
					}					
					fmt.Println( i, userId, email, verified )					
				}
			}
		}
	}*/
	
	go daemon()
	go sendDaemon()
	go balanceDaemon()

	GEmail = NewEmailClient(GSettings.ApiId, GSettings.ApiSecret,
		&Email{GSettings.FromName, GSettings.FromEmail})
	log.Println("Start")
//	go Send()

	http.HandleFunc( `/` + GSettings.Admin + `/sent`, sentHandler)
	http.HandleFunc( `/` + GSettings.Admin + `/send`, sendHandler)
	http.HandleFunc( `/` + GSettings.Admin + `/unban`, unbanHandler)
	http.HandleFunc( `/` + GSettings.Admin + `/ban`, banHandler)
	http.HandleFunc( `/` + GSettings.Admin + `/edit`, editHandler)
	http.HandleFunc( `/` + GSettings.Admin + `/new`, newHandler)
	http.HandleFunc( `/` + GSettings.Admin + `/patterns`, patternsHandler)
	http.HandleFunc( `/` + GSettings.Admin + `/list`, listHandler)
	http.HandleFunc( `/` + GSettings.Admin + `/login`, loginHandler)
	http.HandleFunc( `/` + GSettings.Admin + `/backup`, backupHandler)
	http.HandleFunc( `/` + GSettings.Admin + `/balance`, balanceHandler)
	http.HandleFunc( `/` + GSettings.Admin + `/notifications`, notificationsHandler)
	http.HandleFunc( `/` + GSettings.Admin + `/`, adminHandler)
	http.HandleFunc( `/unsubscribe`, unsubscribeHandler)
	http.HandleFunc( `/`, emailHandler)
	http.ListenAndServe(fmt.Sprintf(":%d", GSettings.Port), nil)
	log.Println("Finish")
}
