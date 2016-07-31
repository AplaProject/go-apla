package main

import (
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"tests_utils"
)

func main() {

	f:=tests_utils.InitLog()
	defer f.Close()

	txType := "CfProjectData";
	txTime := "1427383713";
	userId := []byte("2")
	var blockId int64 = 128008

	var txSlice [][]byte
	// hash
	txSlice = append(txSlice, []byte("1111111111"))
	// type
	txSlice = append(txSlice,  utils.Int64ToByte(utils.TypeInt(txType)))
	// time
	txSlice = append(txSlice, []byte(txTime))
	// user_id
	txSlice = append(txSlice, []byte("4"))
	//project_id
	txSlice = append(txSlice, []byte("1"))
	//lang_id
	txSlice = append(txSlice, []byte("45"))
	//blurb_img
	txSlice = append(txSlice, []byte("http://i.imgur.com/YRCoVnc.jpg"))
	//head_img
	txSlice = append(txSlice, []byte("http://i.imgur.com/YRCoVnc.jpg"))
	//description_img
	txSlice = append(txSlice, []byte("http://i.imgur.com/YRCoVnc.jpg"))
	//picture
	txSlice = append(txSlice, []byte("http://i.imgur.com/YRCoVnc.jpg"))
	//video_type
	txSlice = append(txSlice, []byte("youtube"))
	//video_url_id
	txSlice = append(txSlice, []byte("X-_fg47G5yf-_f"))
	//news_img
	txSlice = append(txSlice, []byte("http://i.imgur.com/YRCoVnc.jpg"))
	//links
	txSlice = append(txSlice, []byte(`[["http:\/\/goo.gl\/fnfh1Dg",1,532,234,0],["http:\/\/goo.gl\/28Fh4h",1355,1344,2222,66]]`))
	//hide
	txSlice = append(txSlice, []byte("1"))
	// sign
	txSlice = append(txSlice, []byte("11111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111"))

	blockData := new(utils.BlockData)
	blockData.BlockId = blockId
	blockData.Time = utils.StrToInt64(txTime)
	blockData.UserId = utils.BytesToInt64(userId)

	err := tests_utils.MakeTest(txSlice, blockData, txType, "work_and_rollback");
	if err != nil {
		fmt.Println(err)
	}

}
