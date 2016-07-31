package main

import (
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"tests_utils"
)

func main() {

	f:=tests_utils.InitLog()
	defer f.Close()

	txType := "NewMiner.";
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
	txSlice = append(txSlice, []byte("91573"))
	//race
	txSlice = append(txSlice, []byte("1"))
	//country
	txSlice = append(txSlice, []byte("1"))
	//latitude
	txSlice = append(txSlice, []byte("55"))
	//longitude
	txSlice = append(txSlice, []byte("55"))
	//host
	txSlice = append(txSlice, []byte("http://55.55.55.55/"))
	//face_coords
	txSlice = append(txSlice, []byte("[[118,275],[241,274],[39,274],[316,276],[180,364],[182,430],[181,490],[93,441],[259,433]]"))
	//profile_coords
	txSlice = append(txSlice, []byte("[[289,224],[148,216],[172,304],[123,239],[328,261],[305,349]]"))
	//face_hash
	txSlice = append(txSlice, []byte("face_hash"))
	//profile_hash
	txSlice = append(txSlice, []byte("profile_hash"))
	//video_type
	txSlice = append(txSlice, []byte("youtube"))
	//video_url_id
	txSlice = append(txSlice, []byte("video_url_id"))
	//node_public_key
	txSlice = append(txSlice, []byte("node_public_key"))
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
