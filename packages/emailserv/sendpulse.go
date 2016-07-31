// emailserv
package main

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Email struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type EmailClient struct {
	apiId       string
	apiSecret   string
	from        *Email
	timeExpired time.Time
	token       string
	Client      *http.Client
}

type jsonToken struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   uint32 `json:"expires_in"`
}

type emailJson struct {
	Html    string   `json:"html"`
	Text    string   `json:"text"`
	Subject string   `json:"subject"`
	From    *Email   `json:"from"`
	To      []*Email `json:"to"`
	Bcc     []*Email `json:"bcc"`
	Files   map[string][]byte  `json:"attachments"`
}

type EmailResult struct {
	Email  string `json:"recipient"`
	Code   string `json:"smtp_answer_code"`
}

var (
	waitBad bool
)

const (
	URL_SEND  = "https://api.sendpulse.com/smtp/emails"
	URL_TOKEN = "https://api.sendpulse.com/oauth/access_token"
	URL_LIST = "https://api.sendpulse.com/smtp/emails?limit=10"
)

func NewEmailClient(apiId, apiSecret string, from *Email) *EmailClient {
	Client := &EmailClient{
		apiId:     apiId,
		apiSecret: apiSecret,
		from:      from,
		Client: &http.Client{
			Transport: http.DefaultTransport,
			Timeout:   20 * time.Second,
		},
	}
	return Client
}

func (ec *EmailClient) GetToken() error {
	values := url.Values{}
	values.Set("grant_type", "client_credentials")
	values.Set("client_id", ec.apiId)
	values.Set("client_secret", ec.apiSecret)

	req, err := http.NewRequest("POST", URL_TOKEN,
		strings.NewReader(values.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res, e := ec.Client.Do(req)
	if e != nil {
		return e
	}

	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	var atoken jsonToken
	err = json.Unmarshal(body, &atoken)
	if err != nil {
		return err
	}
	log.Println(`GetToken Success`)
	ec.timeExpired = time.Now().Add(time.Duration(atoken.ExpiresIn) * time.Second)
	ec.token = atoken.TokenType + ` ` + atoken.AccessToken
	return nil
}

func (ec *EmailClient) CheckBad() {
	waitBad = false
	if time.Now().After(ec.timeExpired) {
		err := ec.GetToken()
		if err != nil {
			return 
		}
	}
	req, err := http.NewRequest("GET", URL_LIST, nil )
	if err != nil {
		return 
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", ec.token)
	res, e := ec.Client.Do(req)
	if e != nil {
		return 
	}
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	
	var answer []*EmailResult
	json.Unmarshal(body, &answer)

	for _,item := range answer {
		if ( item.Code != `250` ) {
			log.Println( `SendPulse code: `, item.Email, item.Code)
			if ( item.Code == `550` ) {
				AddToStopList( item.Email, 0 )
			}
		}
	}
	
	return 
}

func (ec *EmailClient) SendEmailAttach(html, text, subj string, toemail []*Email, files *map[string][]byte) error {
	to := make([]*Email, 0)
	for _,ito := range toemail {
		if len(GSettings.WhiteList) == 0 || utils.InSliceString( ito.Email, GSettings.WhiteList) {
			to = append(to, ito)
		}
	}
	if len(to) == 0 {
		return fmt.Errorf("White list conflict %s", toemail[0].Email )
	}
	
	if time.Now().After(ec.timeExpired) {
		err := ec.GetToken()
		if err != nil {
			return err
		}
	}
	
	values := url.Values{}

	edata := emailJson{
		Html:    base64.StdEncoding.EncodeToString([]byte(html)),
		Text:    text,
		Subject: subj,
		From:    ec.from,
		To:      to,
	}
	if len( GSettings.CopyTo ) > 0 {
		edata.Bcc = []*Email{ &Email{Email: GSettings.CopyTo}}
	}
	if files!=nil && len(*files) > 0 {
		edata.Files = make(map[string][]byte)
		for key,val := range *files {
			edata.Files[key] = val//base64.StdEncoding.EncodeToString(val)
		}
	}
	var serial []byte
	var err error
	if len(edata.Files) > 0 {
		data := make(map[interface{}]interface{})
		data[`html`] = edata.Html
		data[`subject`] = edata.Subject
		data[`text`] = edata.Text
		from := map[interface{}]interface{}{
			`name`: edata.From.Name, `email`: edata.From.Email}
		data[`from`] = from
		to := make( map[interface{}]interface{} )
		for i,val := range edata.To {
			to[i] = map[interface{}]interface{}{`name`: val.Name, `email`: val.Email}
		}
		data[`to`] = to
		attach := make( map[interface{}]interface{} )
		for name,ifile := range edata.Files {
			attach[name] = ifile
		}
		data[`attachments`] = attach
		serial, err = Encode( data )
	} else {
		serial, err = json.Marshal(edata)
	}
	if err != nil {
		return err
	}
	values.Set("email", string(serial))
	req, err := http.NewRequest("POST", URL_SEND,
		strings.NewReader(values.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", ec.token)
	res, e := ec.Client.Do(req)
	if e != nil {
		return e
	}
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	var ret map[string]bool
	json.Unmarshal(body, &ret)
	if ret[`result`] {
		if !waitBad {
			waitBad = true
			time.AfterFunc( 10*time.Second, ec.CheckBad )
		}
		return nil
	}
	return fmt.Errorf("%s", body)
}

func (ec *EmailClient) SendEmail(html, text, subj string, toemail []*Email) error {
	return ec.SendEmailAttach(html, text, subj, toemail, nil)
}
