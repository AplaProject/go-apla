package logging

import (
	"fmt"
	"os"
	"time"

	"github.com/EGaaS/go-egaas-mvp/packages/model"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
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
		allTransactions, err := model.GetAllTransactions(100)
		if err != nil {
			return
		}
		for _, data := range allTransactions {
			allTransactionsStr += fmt.Sprintf("%+v", data)
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
