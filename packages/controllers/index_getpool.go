package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
	"net/http"
	"fmt"
)


func IndexGetPool(w http.ResponseWriter, r *http.Request) {

	fmt.Println("IndexGetPool");
	if utils.DB != nil && utils.DB.DB != nil {

		var err error
		var poolHttpHost string
		var getUserId int64
		
		publicKey := r.FormValue("public_key")
		if len( publicKey ) > 0 {
			getUserId, err = utils.DB.Single("SELECT user_id FROM users WHERE hex(public_key_0) = ?", publicKey).Int64()
			if err != nil {
				log.Error("%v", err)
			}
		} else {
			getUserId = utils.StrToInt64(r.FormValue("user_id"))
		}
		if getUserId == 0 {
			variables, err := utils.DB.GetAllVariables()
			poolHttpHost, err = utils.DB.Single(`SELECT http_host FROM miners_data WHERE i_am_pool = 1 AND pool_count_users < ?`, variables.Int64["max_pool_users"]).String()
			if err != nil {
				log.Error("%v", err)
			}
		} else {
			poolHttpHost, err = utils.DB.Single("SELECT CASE WHEN m.pool_user_id > 0 then (SELECT http_host FROM miners_data WHERE user_id = m.pool_user_id) ELSE http_host end as http_host FROM miners_data as m WHERE m.user_id = ?", getUserId).String()
			if err != nil {
				log.Error("%v", err)
			}
		}
		w.Header().Set("Access-Control-Allow-Origin", "*")
		answer := `{"pool":"`+poolHttpHost+`"}`
		if len( publicKey ) > 0 {
			answer = `{"pool":"`+poolHttpHost+`", "user_id":`+utils.Int64ToStr(getUserId)+`}`
		}
		if _, err := w.Write([]byte(answer)); err != nil {
			log.Error("%v", err)
		}

	}
}