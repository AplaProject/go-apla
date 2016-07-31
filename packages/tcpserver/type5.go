package tcpserver

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
)

/*
Запросы от домена connector
*/

func (t *TcpServer) Type5() {
	// данные присылает демон connector
	buf := make([]byte, 5)
	_, err := t.Conn.Read(buf)
	if err != nil {
		log.Error("%v", utils.ErrInfo(err))
		return
	}
	userId := utils.BinToDec(buf)
	log.Debug("userId: %d", userId)
	// если работаем в режиме пула, то нужно проверить, верный ли у юзера нодовский ключ
	community, err := t.GetCommunityUsers()
	if err != nil {
		log.Error("%v", utils.ErrInfo(err))
		t.Conn.Write(utils.DecToBin(0, 1))
		return
	}
	if len(community) > 0 {
		allTables, err := t.GetAllTables()
		if err != nil {
			log.Error("%v", utils.ErrInfo(err))
			t.Conn.Write(utils.DecToBin(0, 1))
			return
		}
		keyTable := utils.Int64ToStr(userId) + "_my_node_keys"
		if !utils.InSliceString(keyTable, allTables) {
			log.Error("incorrect user_id %d", userId)
			t.Conn.Write(utils.DecToBin(0, 1))
			return
		}
		myBlockId, err := t.GetMyBlockId()
		if err != nil {
			log.Error("%v", utils.ErrInfo(err))
			t.Conn.Write(utils.DecToBin(0, 1))
			return
		}
		myNodeKey, err := t.Single(`
				SELECT public_key
				FROM `+keyTable+`
				WHERE block_id = (SELECT max(block_id) FROM  `+keyTable+`) AND
							 block_id < ?
				`, myBlockId).String()
		if err != nil {
			log.Error("%v", utils.ErrInfo(err))
			t.Conn.Write(utils.DecToBin(0, 1))
			return
		}
		if len(myNodeKey) == 0 {
			log.Error("len(myNodeKey) userId %d",  userId)
			t.Conn.Write(utils.DecToBin(0, 1))
			return
		}
		nodePublicKey, err := t.GetNodePublicKey(userId)
		if err != nil {
			log.Error("%v", utils.ErrInfo(err))
			t.Conn.Write(utils.DecToBin(0, 1))
			return
		}
		if myNodeKey != string(nodePublicKey) {
			log.Error("%v", utils.ErrInfo("myNodeKey != nodePublicKey"))
			t.Conn.Write(utils.DecToBin(0, 1))
			return
		}
		// всё норм, шлем 1
		_, err = t.Conn.Write(utils.DecToBin(1, 1))
		if err != nil {
			log.Error("%v", utils.ErrInfo(err))
			return
		}
	} else {
		// всё норм, шлем 1
		_, err = t.Conn.Write(utils.DecToBin(1, 1))
		if err != nil {
			log.Error("%v", utils.ErrInfo(err))
			return
		}
	}
}
