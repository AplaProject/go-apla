package logging

import (
	"fmt"
	"os"
	"time"

	"github.com/EGaaS/go-egaas-mvp/packages/consts"
	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
	"github.com/EGaaS/go-egaas-mvp/packages/utils/sql"
)

func WriteSelectiveLog(text interface{}) {
	if *utils.LogLevel == "DEBUG" {
		var stext string
		switch text.(type) {
		case string:
			stext = text.(string)
		case []byte:
			stext = string(text.([]byte))
		case error:
			stext = fmt.Sprintf("%v", text)
		}
		allTransactionsStr := ""
		allTransactions, _ := sql.DB.GetAll("SELECT hex(hash) as hex_hash, verified, used, high_rate, for_self_use, user_id, third_var, counter, sent FROM transactions", 100)
		for _, data := range allTransactions {
			allTransactionsStr += data["hex_hash"] + "|" + data["verified"] + "|" + data["used"] + "|" + data["high_rate"] + "|" + data["for_self_use"] + "|" + consts.TxTypes[converter.StrToInt(data["type"])] + "|" + data["user_id"] + "|" + data["third_var"] + "|" + data["counter"] + "|" + data["sent"] + "\n"
		}
		t := time.Now()
		data := allTransactionsStr + utils.GetParent() + " ### " + t.Format(time.StampMicro) + " ### " + stext + "\n\n"
		//ioutil.WriteFile(*Dir+"/SelectiveLog.txt", []byte(data), 0644)
		f, err := os.OpenFile(*utils.Dir+"/SelectiveLog.txt", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
		if err != nil {
			panic(err)
		}

		defer f.Close()

		if _, err = f.WriteString(data); err != nil {
			panic(err)
		}
	}
}
