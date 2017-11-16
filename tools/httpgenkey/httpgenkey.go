// httpgenkey
package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/AplaProject/go-apla/packages/crypto"
)

type Settings struct {
	Port uint32 `json:"port"`
}

type Answer struct {
	Private string `json:"private"`
	Public  string `json:"public"`
	Key     string `json:"key"`
	Error   string `json:"error"`
}

var (
	GSettings Settings
)

func genHandler(w http.ResponseWriter, r *http.Request) {
	answer := Answer{}
	if r.URL.Path[1:] == `genkey` {
		priv, pub, err := crypto.GenHexKeys()
		if err == nil {
			answer.Public = pub
			answer.Private = priv
			apub, err := hex.DecodeString(pub)
			if err != nil {
				answer.Error = err.Error()
			}
			answer.Key = strconv.FormatInt(crypto.Address(apub), 10)
		} else {
			answer.Error = err.Error()
		}
	}
	ret, err := json.Marshal(answer)
	if err != nil {
		ret = []byte(fmt.Sprintf(`{"error":%q}`, err.Error()))
	}
	if r.URL.Path[1:] == `genkey` {
		fmt.Println(`GenKey: `, string(ret))
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(ret)
}

func main() {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		fmt.Println(`Dir`, err)
	}
	params, err := ioutil.ReadFile(filepath.Join(dir, `settings.json`))
	if err != nil {
		fmt.Println(dir, `settings.json`, err)
	}
	if err = json.Unmarshal(params, &GSettings); err != nil {
		fmt.Println(`Unmarshall`, err)
	}
	fmt.Println("Start")
	http.HandleFunc("/", genHandler)
	http.ListenAndServe(fmt.Sprintf(":%d", GSettings.Port), nil)
	fmt.Println("Finish")
}
