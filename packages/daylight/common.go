package daylight

import (
	"fmt"
	"github.com/astaxie/beego/session"
	"github.com/DayLightProject/go-daylight/packages/consts"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"github.com/op/go-logging"
	_ "image/png"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"encoding/hex"
)

var (
	log            = logging.MustGetLogger("daylight")
	format         = logging.MustStringFormatter("%{color}%{time:15:04:05.000} %{shortfile} %{shortfunc} [%{level:.4s}] %{color:reset} %{message}[" + consts.VERSION + "]" + string(byte(0)))
	configIni      map[string]string
	globalSessions *session.Manager
)

func firstBlock() {
	if *utils.GenerateFirstBlock == 1 {
		PublicKey, _ := ioutil.ReadFile(*utils.Dir + "/PublicKey")
//		PublicKeyBytes, _ := base64.StdEncoding.DecodeString(string(PublicKey))
		PublicKeyBytes,_ := hex.DecodeString(string(PublicKey))

		NodePublicKey, _ := ioutil.ReadFile(*utils.Dir + "/NodePublicKey")
//		NodePublicKeyBytes, _ := base64.StdEncoding.DecodeString(string(NodePublicKey))
		NodePublicKeyBytes,_ := hex.DecodeString(string(NodePublicKey))
		Host, _ := ioutil.ReadFile(*utils.Dir + "/Host")

		tx := utils.DecToBin(1, 1)
		tx = append(tx, utils.DecToBin(utils.Time(), 4)...)
		tx = append(tx, utils.EncodeLengthPlusData("1")...) // wallet_id
		tx = append(tx, utils.EncodeLengthPlusData("0")...) // citizen_id
		tx = append(tx, utils.EncodeLengthPlusData(PublicKeyBytes)...)
		tx = append(tx, utils.EncodeLengthPlusData(NodePublicKeyBytes)...)
		tx = append(tx, utils.EncodeLengthPlusData(Host)...)

		block := utils.DecToBin(0, 1)
		block = append(block, utils.DecToBin(1, 4)...)
		block = append(block, utils.DecToBin(utils.Time(), 4)...)
		block = append(block, utils.EncodeLengthPlusData("1")...) // wallet_id
		block = append(block, utils.DecToBin(0, 1)...) // cb_id
		block = append(block, utils.EncodeLengthPlusData(tx)...)

		static := filepath.Join("", "static")
		if _, err := os.Stat(static); os.IsNotExist(err) {
			if err = os.Mkdir(static, 0755); err != nil {
				log.Error("%v", utils.ErrInfo(err))
			} 
		}
		ioutil.WriteFile( filepath.Join( static, "1block"), block, 0644)
		os.Exit(0)
	}

}

// http://grokbase.com/t/gg/golang-nuts/12a9yhgr64/go-nuts-disable-directory-listing-with-http-fileserver#201210093cnylxyosmdfuf3wh5xqnwiut4
func noDirListing(h http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/") {
			http.NotFound(w, r)
			return
		}
		h.ServeHTTP(w, r)
	})
}

func openBrowser(BrowserHttpHost string) {
	log.Debug("runtime.GOOS: %v", runtime.GOOS)
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", BrowserHttpHost).Start()
	case "windows", "darwin":
		err = exec.Command("open", BrowserHttpHost).Start()
		if err != nil {
			exec.Command("cmd", "/c", "start", BrowserHttpHost).Start()
		}
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Error("%v", err)
	}
}

func GetHttpHost() (string, string, string) {
	BrowserHttpHost := "http://localhost:"+*utils.ListenHttpPort
	HandleHttpHost := ""
	ListenHttpHost := ":"+*utils.ListenHttpPort
	if len(*utils.TcpHost) > 0 {
		ListenHttpHost = *utils.TcpHost+":"+*utils.ListenHttpPort
	}
	return BrowserHttpHost, HandleHttpHost, ListenHttpHost
}
