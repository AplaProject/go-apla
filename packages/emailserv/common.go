// common
package main

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
	"fmt"
	"log"
	"bytes"
	"hash/crc32"
	"strconv"
	"html/template"
)

func Currency(currency int64) string {
	ret,_ := utils.DB.Single(`SELECT name FROM currency where id=?`,currency ).String()
	return ret
}

func RoundMoney(in float64) (out float64) {
	off := float64(10)
	for k:=0; k<5; k++ {
		if in < off {
			out = utils.Round( in, 4 - k )
			break
		}
		off *= 10
	}
	if out == 0 {
		out = utils.Round( in, 0 )
	}
	return
}

func CheckUser( userId int64 ) (map[string]interface{}, error) {
	data := make(map[string]interface{})
	if userId == utils.EXCHANGE_USER {
		data[`lang`] = int64(1)
		data[`nobcc`] = 1
		return data, nil
	}
	
	user, err := GDB.OneRow("select * from users where user_id=?", userId ).String()
	if err != nil {
		return nil, err
	}
	if len(user) == 0 {
		return nil, fmt.Errorf(`The user has no email`)
	} else if utils.StrToInt( user[`verified`] ) < 0 {
		return nil, fmt.Errorf(`The user in the stop-list`)
	}
	data[`email`] = user[`email`]
	data[`Unsubscribe`] = fmt.Sprintf( `%s/unsubscribe?uid=%d-%s`, 
		 	utils.EMAIL_SERVER, userId, strconv.FormatUint( uint64( crc32.ChecksumIEEE([]byte(user[`email`]))), 32 ))
			
	lang := utils.StrToInt64( user[`lang`] )
	if lang == 0 {
		if country, err := utils.DB.Single(`select country from miners_data where user_id=?`, userId ).Int64(); err == nil {
			switch country {
				case 10, 14, 19, 67, 80, 112, 119, 125, 180, 214, 224, 230, 235:
					lang = 42
			}	
		} else {
			lang = 1
		}
		if err := GDB.ExecSql(`update users set lang=? where user_id=?`, lang, userId ); err != nil {
			return nil, err
		}
	}
	data[`lang`] = lang
	return data, nil
}

func EmailUser( userId int64, data map[string]interface{}, cmd int ) bool {

	result := func( msg string ) bool {
		log.Println( fmt.Sprintf( `Error: user_id=%d %s`, userId, msg ))
		return false
	}
	
	patterns := []string{ `unknown`, `new`, `test`, `adminmsg`, `cashreq`, `changestat`,
		`dccame`, `dcsent`, `updprimary`,`updemail`, `updsms`, `voteres`,
		`votetime`, `newver`, `nodetime`, `signup`, `balance`, `exrequest`, `exanswer`,
		`refready`}
	pattern := patterns[cmd]
	if len(pattern) == 0 {
		pattern = data[`pattern`].(string)
	}
	common := `common`
	if userId != utils.EXCHANGE_USER {
		data[`UserId`] = userId

		data[`Status`],_ = utils.DB.Single(`select status from miners_data where user_id=?`, userId ).String()
		if len(data[`Status`].(string)) == 0 {
			data[`Status`],_ = utils.DB.Single(`select status from users where user_id=?`, userId ).String()
		}
	} else {
		common = `exchange`
	}
	subject := new(bytes.Buffer)
	html := new(bytes.Buffer)
	lang := utils.Int64ToStr( data[`lang`].(int64) )
	if data[`lang`].(int64) > 1 {
		GPagePattern.ExecuteTemplate(subject, pattern + `Subject` + lang, data )
		GPagePattern.ExecuteTemplate(html, pattern + `HTML` + lang, data )
	}
	if len( subject.String()) == 0 {
		GPagePattern.ExecuteTemplate(subject, pattern + `Subject`, data )
	}
	if len( html.String()) == 0 {
		if err := GPagePattern.ExecuteTemplate(html, pattern + `HTML`, data ); err!=nil {
			return result( err.Error() )
		}
	}
	if len( html.String()) == 0 {
		return result( `Empty pattern ` + pattern )
	}
	data[`Body`] = template.HTML(html.String())
	html.Reset()
	if data[`lang`].(int64) > 1 {
		if len( subject.String()) == 0 {
			GPagePattern.ExecuteTemplate(subject, common + `Subject` + lang, data )
		}
		GPagePattern.ExecuteTemplate(html, common + `HTML` + lang, data )
	}
	if len( subject.String()) == 0 {
		GPagePattern.ExecuteTemplate(subject, common + `Subject`, data )
	}
	if len( html.String()) == 0 {
		GPagePattern.ExecuteTemplate(html, common + `HTML`, data )
	}
	if len( subject.String()) == 0 {
		subject.WriteString(`DCoin notifications`)
	}
	bcc := GSettings.CopyTo
	if _, ok := data[`nobcc`]; ok {
		GSettings.CopyTo = ``
	}
	err := GEmail.SendEmail( html.String(), ``, subject.String(),
		[]*Email{&Email{``, data[`email`].(string) }})
	if _, ok := data[`nobcc`]; ok {
		GSettings.CopyTo = bcc
	}
	if err != nil && userId != utils.EXCHANGE_USER {
		GDB.ExecSql(`UPDATE users SET verified=? WHERE user_id=?`, -1, userId )
		return result( fmt.Sprintf(`SendPulse %s`, err.Error()))
	}

	log.Println( `Daemon Sent:`, cmd, data[`email`].(string), userId )
	if err := GDB.ExecSql(`INSERT INTO log ( user_id, email, cmd, params, uptime, ip )
				 VALUES ( ?, ?, ?, ?, datetime('now'), ? )`,
		         userId, data[`email`].(string), cmd, ``, 1 ); err != nil {
					return result( err.Error() )
			}
	return true
}
