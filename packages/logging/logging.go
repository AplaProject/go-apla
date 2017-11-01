package logging

import (
	"fmt"
	"os"
	"time"

	"github.com/AplaProject/go-apla/packages/model"
	"github.com/AplaProject/go-apla/packages/utils"
)

func WriteSelectiveLog(text interface{}) error {
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
			return err
		}
		for _, data := range *allTransactions {
			allTransactionsStr += fmt.Sprintf("%+v", data)
		}
		t := time.Now()
		data := allTransactionsStr + utils.GetParent() + " ### " + t.Format(time.StampMicro) + " ### " + stext + "\n\n"
		f, err := os.OpenFile(*utils.Dir+"/SelectiveLog.txt", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
		if err != nil {
			return err
		}
		defer f.Close()

		if _, err := f.WriteString(data); err != nil {
			return err
		}
	}
	return nil
}
