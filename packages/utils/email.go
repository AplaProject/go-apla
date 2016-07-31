// email
package utils

import (
//	"crypto"
//	"crypto/rand"
//	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"fmt"
)

const (
	EMAIL_SERVER = `http://email.dcoin.club:8200`
)

const (
	ECMD_UNKNOWN = iota
	ECMD_NEW     // Привязка email к пользователю, должно отправляться при подключении уведомлений
	// Отправляет тестовое сообщение
	ECMD_TEST       // Отправить тестовое сообщение
	ECMD_ADMINMSG   // Сообщение от администратора
	ECMD_CASHREQ    // Уведомление incoming_cash_requests
	ECMD_CHANGESTAT // Уведомление change_in_status
	ECMD_DCCAME     // Уведомление dc_came_from
	ECMD_DCSENT     // Уведомление dc_sent
	ECMD_UPDPRIMARY // Уведомление update_primary_key
	ECMD_UPDEMAIL   // Уведомление update_email
	ECMD_UPDSMS     // Уведомление update_sms_request
	ECMD_VOTERES    // Уведомление voting_results
	ECMD_VOTETIME   // Уведомление voting_time
	ECMD_NEWVER     // Уведомление new_version
	ECMD_NODETIME   // Уведомление node_time
	ECMD_SIGNUP     // Сообщаем Email с новой заявкой майнера, ничего не отправляется
	ECMD_BALANCE    // Отправляем информацию о балансе
	ECMD_EXREQUEST  // Уведомление о вопросе в поддержку биржи
	ECMD_EXANSWER   // Уведомление об ответе из поддержки биржи
	ECMD_REFREADY	// Уведомление о готовности ключа для реферала
	ECMD_SENDKEY    // Отправка ключей на email
	ECMD_FORKBLOCK  // Уведомление, что имеются разные ветки блокчейна
	
	EXCHANGE_USER = 0xefffffff
)

type Answer struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

type JsonEmail struct {
	Email  string `json:"email"`
	UserId int64  `json:"user_id"`
	Cmd    uint   `json:"cmd"`
	Params *map[string]string
}

type TypeNfyCame struct {
	TypeTx     string  
	FromUserId   int64
	Amount     float64
    CurrencyId int64 
	Comment    string
	CommentStatus string
}

type TypeNfySent struct {
	TypeTx     string  
	ToUserId   int64
	Amount     float64
	Commission float64 
    CurrencyId int64 
	Comment    string
	CommentStatus string
}

type TypeNfyCashRequest struct {
	FromUserId int64    `json:"from_user"`
	Amount     float64  `json:"amount"`
    CurrencyId int64    `json:"currency_id"`
}

type TypeNfyStatus struct {
	Status     string  `json:"status"`
}

type TypeNfyRefReady struct {
	RefId     string  `json:"refid"`
}

type TypeNfySendKey struct {
	UserId    int64     `json:"user_id"`
	Subject   string    `json:"subject"`
	Text      string    `json:"text"`
    TxtKey    string    `json:"txt_key"`
    PngKey    string    `json:"png_key"`
	RefId     int64     `json:"refid"`
}


func SendEmail(email string, userId int64, cmd uint, params *map[string]string) (err error) {
	var (
//		community         []int64
//		private, myPrefix string
		data, signature   []byte
//		privateKey        *rsa.PrivateKey
		answer            Answer
		req               *http.Request
		res               *http.Response
	)
	if len( email ) == 0 {
		return
	}
/*	community, err = DB.GetCommunityUsers()
	if len(community) > 0 {
		myPrefix = Int64ToStr(userId) + "_"
	}
	if private, err = DB.GetPrivateKey(myPrefix); err != nil {
		return
	}
	if privateKey, err = MakePrivateKey(private); err != nil {
		return
	}*/
	jsonEmail := &JsonEmail{Email: email, UserId: userId, Cmd: cmd, Params: params}

	if data, err = json.Marshal(jsonEmail); err != nil {
		return
	}
/*	signature, err = rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA1, HashSha1(string(data)))
	if err != nil {
		return
	}*/

	Client := &http.Client{
		Transport: http.DefaultTransport,
		Timeout:   20 * time.Second,
	}
	values := url.Values{}
	values.Set("data", string(data))
	values.Set("sign", base64.StdEncoding.EncodeToString(signature))
/*	if cmd == ECMD_TEST || cmd == ECMD_NEW {
		// В случае подключения уведомлений таблица users еще может не иметь данного пользователя
		// поэтому вместе с данными отправляем публичный ключ
		if public, err := DB.GetMyPublicKey(myPrefix); err == nil {
			values.Set("public", base64.StdEncoding.EncodeToString(public))
		}
	}*/
	emailServer := EMAIL_SERVER
	if val,ok := DB.ConfigIni["localemail"]; ok {
		emailServer = val
	}
	if req, err = http.NewRequest("POST", emailServer,
		strings.NewReader(values.Encode())); err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if res, err = Client.Do(req); err != nil {
		return
	}

	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	if err = json.Unmarshal(body, &answer); err != nil {
		return
	}
	if !answer.Success {
		return fmt.Errorf(answer.Error)
	}

	return
}

func ExchangeEmail(email, exchange string, cmd uint ) error {
	return SendEmail(email, EXCHANGE_USER, cmd, &map[string]string{`exchange`: exchange})
}