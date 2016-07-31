package utils

import (
	"fmt"
	"net"
	"sync"
	"time"
)

var ChatMinSignTime int64

// сигнал горутине, которая мониторит таблу chat, что есть новые данные
var ChatNewTx = make(chan int64, 1000)

//var ChatJoinConn = make(chan net.Conn)
//var ChatPoolConn []net.Conn
//var ChatDelConn = make(chan net.Conn)

var ChatMutex = &sync.Mutex{}

type ChatData struct {
	Hashes     []byte
	HashesArr  [][]byte
	LastMessId int64
}
type ChatOutConnectionsType struct {
	MessIds        []int64
	ConnectionChan chan *ChatData
}

//var ChatDataChan chan *ChatData = make(chan *ChatData, 1000)
// исходящие соединения протоколируем тут для исключения создания повторных
// исходящих соединений. []MessIds - id сообщений, которые отправили
var ChatOutConnections map[int64]*ChatOutConnectionsType = make(map[int64]*ChatOutConnectionsType)
var ChatInConnections map[int64]int = make(map[int64]int)

// Ждет входящие данные
func ChatInput(conn net.Conn, userId int64) {

	fmt.Println("ChatInput start. wait data from ", conn.RemoteAddr().String(), Time())

	for {

		conn.SetReadDeadline(time.Now().Add(120 * time.Second))

		// тут ждем, пока нам пришлют данные
		fmt.Println("ChatInput for", conn.RemoteAddr().String(), Time())
		binaryData, err := TCPGetSizeAndData(conn, 1048576)
		if err != nil {
			fmt.Println("ChatInput ERROR", err, conn.RemoteAddr().String(), Time())
			log.Error("ChatInput ERROR", err, conn.RemoteAddr().String(), Time())
			safeDeleteFromChatMapIn(ChatInConnections, userId)
			safeDeleteFromChatMap(ChatOutConnections, userId)
			return
		}
		conn.SetReadDeadline(time.Time{})
		fmt.Printf("binaryData %x\n", binaryData)

		// каждые 30 сек шлется сигнал, что канал еще жив
		if len(binaryData) < 16 {
			fmt.Println(">> Get test data from ", conn.RemoteAddr().String(), Time())
			continue
		}

		var hash []byte
		addsql := ""
		var hashes []map[string]int
		for {
			hash = BytesShift(&binaryData, 16)
			if DB.ConfigIni["db_type"] == "postgresql" {
				addsql += "decode('" + string(BinToHex(hash)) + "', 'hex'),"
			} else {
				addsql += "x'" + string(BinToHex(hash)) + "',"
			}
			hashes = append(hashes, map[string]int{string(hash): 1})
			if len(binaryData) < 16 {
				break
			}
		}

		if len(addsql) == 0 {
			fmt.Println("empty hashes")
			log.Error("empty hashes")
			safeDeleteFromChatMapIn(ChatInConnections, userId)
			safeDeleteFromChatMap(ChatOutConnections, userId)
			return
		}
		addsql = addsql[:len(addsql)-1]
		fmt.Println("addsql", addsql)

		// смотрим в табле chat, чего у нас уже есть
		fmt.Println(`SELECT hash FROM chat WHERE hash IN (` + addsql + `)`)
		rows, err := DB.Query(`SELECT hash FROM chat WHERE hash IN (` + addsql + `)`)
		if err != nil {
			fmt.Println(ErrInfo(err))
			log.Error("%v", ErrInfo(err))
			safeDeleteFromChatMapIn(ChatInConnections, userId)
			safeDeleteFromChatMap(ChatOutConnections, userId)
			return
		}
		defer rows.Close()
		for rows.Next() {
			var hash string
			err = rows.Scan(&hash)
			if err != nil {
				fmt.Println(ErrInfo(err))
				log.Error("%v", ErrInfo(err))
				safeDeleteFromChatMapIn(ChatInConnections, userId)
				safeDeleteFromChatMap(ChatOutConnections, userId)
				return
			}
			// отмечаем 0 то, что у нас уже есть
			for k, v := range hashes {
				if _, ok := v[hash]; ok {
					hashes[k][hash] = 0
				}
			}
		}

		var needTx bool // есть ли что слать
		binHash := ""
		// преобразуем хэши в набор бит, где 0 означет, что такой хэш есть и его слать не надо, а 1 - надо
		for _, hashmap := range hashes {
			for _, result := range hashmap {
				binHash = binHash + IntToStr(result)
				if result == 1 {
					needTx = true
				}
			}
		}

		fmt.Println("binHash", binHash)
		// шлем набор байт, который содержит метки, чего надо качать или "0" - значит ничего качать не будем
		err = WriteSizeAndData([]byte(binHash), conn)
		if err != nil {
			fmt.Println(ErrInfo(err))
			log.Error("%v", ErrInfo(err))
			safeDeleteFromChatMapIn(ChatInConnections, userId)
			safeDeleteFromChatMap(ChatOutConnections, userId)
			return
		}
		if !needTx {
			fmt.Println("continue")
			continue
		}

		// время последнего сообщения
		lastMessTime, err := DB.Single(`SELECT max(sign_time) FROM chat`).Int64()
		fmt.Println("lastMessTime===", lastMessTime)
		if err != nil {
			log.Error("%v", ErrInfo(err))
		}
		if lastMessTime > Time() {
			lastMessTime = Time() - 1800
		} else if lastMessTime == 0 {
			lastMessTime = Time() - 86400*7
		} else {
			lastMessTime = lastMessTime - 1800
		}

		// получаем тр-ии, которых у нас нету
		binaryData, err = TCPGetSizeAndData(conn, 10485760)
		if err != nil {
			fmt.Println(ErrInfo(err))
			log.Error("%v", ErrInfo(err))
			safeDeleteFromChatMapIn(ChatInConnections, userId)
			safeDeleteFromChatMap(ChatOutConnections, userId)
			return
		}
		var sendToChan int64
		for {
			length := DecodeLength(&binaryData)
			fmt.Println("length: ", length)
			if int(length) > len(binaryData) {
				fmt.Println("break length > len(binaryData)", length, len(binaryData))
				log.Error("break length > len(binaryData)", length, len(binaryData))
				safeDeleteFromChatMapIn(ChatInConnections, userId)
				safeDeleteFromChatMap(ChatOutConnections, userId)
				return
			}
			if length > 0 {
				txData := BytesShift(&binaryData, length)
				//fmt.Printf("txData %x\n", txData)
				lang := BinToDecBytesShift(&txData, 1)
				room := BinToDecBytesShift(&txData, 4)
				receiver := BinToDecBytesShift(&txData, 4)
				sender := BinToDecBytesShift(&txData, 4)
				status := BinToDecBytesShift(&txData, 1)
				message := BytesShift(&txData, DecodeLength(&txData))
				signTime := BinToDecBytesShift(&txData, 4)
				signature := BinToHex(BytesShift(&txData, DecodeLength(&txData)))

				// нам не нужны старые сообщения, которые мы могли уже удалить
				// поэтому смотрим время последнего сообщения или нашего времени,
				// вычитаем 30 минут и не берем всё что меньше
				if signTime < lastMessTime {
					continue
				}
				fmt.Println("signTime", signTime)
				fmt.Println("lastMessTime", lastMessTime)

				// проверяем даннные из тр-ий
				err := DB.CheckChatMessage(string(message), sender, receiver, lang, room, status, signTime, signature)
				if err != nil {
					fmt.Println(ErrInfo(err))
					log.Error("%v", ErrInfo(err))
					//safeDeleteFromChatMapIn(ChatInConnections, userId)
					//safeDeleteFromChatMap(ChatOutConnections, userId)
					//return
					continue
				}

				data := Int64ToByte(lang)
				data = append(data, Int64ToByte(room)...)
				data = append(data, Int64ToByte(receiver)...)
				data = append(data, Int64ToByte(sender)...)
				data = append(data, Int64ToByte(status)...)
				data = append(data, []byte(message)...)
				data = append(data, Int64ToByte(signTime)...)
				data = append(data, []byte(signature)...)
				hash = Md5(data)
				// заносим в таблу
				chatId, err := DB.ExecSqlGetLastInsertId(`INSERT INTO chat (hash, time, lang, room, receiver, sender, status, message, sign_time, signature) VALUES ([hex], ?, ?, ?, ?, ?, ?, ?, ?, [hex])`, "id", hash, Time(), lang, room, receiver, sender, status, message, signTime, signature)
				if err != nil {
					fmt.Println(ErrInfo(err))
					log.Error("%v", ErrInfo(err))
					//return
				}
				sendToChan = chatId

			}
			if length == 0 {
				break
			}
		}

		// шлем макс. ID, т.к. ChatOutput отправит все предыдущие проврив какие ID есть в ChatOutConnections
		if sendToChan > 0 {
			ChatNewTx <- sendToChan
		}
	}
}

// каждый 30 сек шлет данные в канал, чтобы держать его живым
func ChatOutputTesting() {
	for {
		// шлем всем горутинам ChatTxDisseminator, чтобы они разослали по серверам,
		// которые ранее к нам подключились или к которым мы подключались
		//fmt.Println("ChatOutConnections:", ChatOutConnections)
		for _, data := range ChatOutConnections {
			data.ConnectionChan <- nil
		}
		Sleep(30)
	}
}

// ожидает появления свежих записей в чате, затем ждет появления коннектов
// (заносятся из демеона connections и от тех, кто сам подключился к ноде)
func ChatOutput(newTx chan int64) {

	// держим канал в активном состоянии
	go ChatOutputTesting()

	for {
		fmt.Println("ChatOutput wait newTx")
		// просто так тр-ии в chat не появятся, их кто-то должен туда запихать, ждем тут
		chatId := <-newTx
		fmt.Println("ChatOutput newTx")

		// готовим ID, которые выберем из БД
		ids := make(map[int64]int)
		ChatMutex.Lock()
		for _, data := range ChatOutConnections {
			var lastMessId int64
			if len(data.MessIds) > 0 {
				lastMessId = data.MessIds[len(data.MessIds)-1]
			}
			if lastMessId < chatId {
				for i := lastMessId + 1; i <= chatId; i++ {
					ids[i] = 1
				}
			}
		}
		ChatMutex.Unlock()
		fmt.Println("chat ids for send", ids)
		log.Debug("%v", ids)

		if len(ids) == 0 {
			Sleep(10)
			continue
		}
		// смотрим, есть ли в табле неотправленные тр-ии
		rows, err := DB.Query("SELECT id, hash, lang, room, receiver, sender, status, message, enc_message, sign_time, signature FROM chat WHERE id IN (" + JoinInts64(ids, ",") + ")")
		if err != nil {
			fmt.Println(ErrInfo(err))
		}
		defer rows.Close()
		messages := make(map[int64][][]byte)
		for rows.Next() {
			var id, lang, room, receiver, sender, status, signTime int64
			var message, enc_message string
			var signature, hash []byte
			err = rows.Scan(&id, &hash, &lang, &room, &receiver, &sender, &status, &message, &enc_message, &signTime, &signature)
			if err != nil {
				fmt.Println(ErrInfo(err))
				continue
			}
			if status == 2 {
				message = enc_message
				status = 1
			}
			/*fmt.Println(`UPDATE chat SET sent = 1 WHERE hex(hash) = ?`, string(BinToHex(hash)))
			err = DB.ExecSql(`UPDATE chat SET sent = 1 WHERE hex(hash) = ?`, string(BinToHex(hash)))
			if err != nil {
				fmt.Println(ErrInfo(err))
				continue
			}*/
			data := DecToBin(lang, 1)
			data = append(data, DecToBin(room, 4)...)
			data = append(data, DecToBin(receiver, 4)...)
			data = append(data, DecToBin(sender, 4)...)
			data = append(data, DecToBin(status, 1)...)
			data = append(data, EncodeLengthPlusData(message)...)
			data = append(data, DecToBin(signTime, 4)...)
			data = append(data, EncodeLengthPlusData(signature)...)
			//allTx = append(allTx, utils.EncodeLengthPlusData(data))

			//hashes = append(hashes, hash...)
			//hashesArr = append(hashesArr, data)
			messages[id] = [][]byte{hash, data}
		}
		if len(messages) == 0 {
			fmt.Println("len(messages) == 0")
			log.Debug("len(messages) == 0")
			continue
		}

		var hashes []byte
		var hashesArr [][]byte
		// шлем всем горутинам ChatTxDisseminator, чтобы они разослали по серверам,
		// которые ранее к нам подключились или к которым мы подключались
		ChatMutex.Lock()
		for _, data := range ChatOutConnections {
			var lastMessId int64
			if len(data.MessIds) > 0 {
				lastMessId = data.MessIds[len(data.MessIds)-1]
			}
			if lastMessId < chatId {
				for i := lastMessId + 1; i <= chatId; i++ {
					if message, ok := messages[i]; ok {
						hashes = append(hashes, message[0]...)
						hashesArr = append(hashesArr, message[1])
					}
				}
				lastMessId = chatId
				if len(hashesArr) > 0 {
					data.ConnectionChan <- &ChatData{Hashes: hashes, HashesArr: hashesArr, LastMessId: lastMessId}
				}
			}
		}
		ChatMutex.Unlock()
		/*for i:=0; i < len(ChatOutConnections); i++ {
			fmt.Println("ChatData", i, hashes, hashesArr)
			ChatDataChan <- &ChatData{Hashes: hashes, HashesArr: hashesArr}
		}*/
	}
}

// когда подклюаемся к кому-то или когда кто-то подключается к нам,
// то создается горутина, которая будет ждать, пока появятся свежие
// данные в табле chat, чтобы послать их

// create a go routine on connect. It waits for fresh data in table chat
func ChatTxDisseminator(conn net.Conn, userId int64, connectionChan chan *ChatData) {
	chatId, err := DB.Single(`SELECT max(id) FROM chat`).Int64()
	if err != nil {
		log.Error("%v", ErrInfo(err))
		return
	}
	// даем команду рассыльщику, чтобы отправил всем хэш тр-ии сообщения
	ChatNewTx <- chatId

	for {
		fmt.Println("wait ChatDataChan send TO->", conn.RemoteAddr().String(), Time())
		data := <-connectionChan
		if data == nil {
			fmt.Println("> send test data to ", conn.RemoteAddr().String(), Time())
			err := WriteSizeAndData(EncodeLengthPlusData([]byte{0}), conn)
			if err != nil {
				fmt.Println(ErrInfo(err))
				log.Error("%v", ErrInfo(err))
				safeDeleteFromChatMap(ChatOutConnections, userId)
				break
			}
			Sleep(1)
			continue
		} else {
			fmt.Println("data", len(data.Hashes), "TO->", conn.RemoteAddr().String(), Time())
			// шлем хэши
			err := WriteSizeAndData(data.Hashes, conn)
			if err != nil {
				fmt.Println(ErrInfo(err))
				log.Error("%v", ErrInfo(err))
				safeDeleteFromChatMap(ChatOutConnections, userId)
				break
			}
		}
		fmt.Println("WriteSizeAndData ok", conn.RemoteAddr().String(), Time())

		// получаем номера хэшей, тр-ии которых пошлем далее
		hashesBin, err := TCPGetSizeAndData(conn, 10485760)
		if err != nil {
			fmt.Println(ErrInfo(err))
			log.Error("%v", ErrInfo(err))
			safeDeleteFromChatMap(ChatOutConnections, userId)
			break
		}
		fmt.Println("TCPGetSizeAndData ok")

		var TxForSend []byte
		for i := 0; i < len(hashesBin); i++ {
			hashMark := hashesBin[i : i+1]
			if string(hashMark) == "1" {
				TxForSend = append(TxForSend, EncodeLengthPlusData(data.HashesArr[i])...)
			}
		}

		fmt.Printf("TxForSend: %x\n", TxForSend)
		// шлем тр-ии
		if len(TxForSend) > 0 {
			err = WriteSizeAndData(TxForSend, conn)
			if err != nil {
				fmt.Println(ErrInfo(err))
				log.Error("%v", ErrInfo(err))
				safeDeleteFromChatMap(ChatOutConnections, userId)
				break
			}
		}

		if ChatOutConnections[userId] == nil {
			fmt.Println("ChatOutConnections[userId] == nil", userId)
			log.Error("ChatOutConnections[userId] == nil", userId)
			break
		}

		// добавляем ID сообщения, чтобы больше его не слать
		ChatMutex.Lock()
		var lastMessId int64
		if len(ChatOutConnections[userId].MessIds) > 0 {
			lastMessId = ChatOutConnections[userId].MessIds[len(ChatOutConnections[userId].MessIds)-1]
		}
		if lastMessId < data.LastMessId {
			for i := lastMessId + 1; i <= data.LastMessId; i++ {
				ChatOutConnections[userId].MessIds = append(ChatOutConnections[userId].MessIds, i)
			}
		}
		ChatMutex.Unlock()

		fmt.Println("WriteSizeAndData 2  ok")
		time.Sleep(10 * time.Millisecond)
	}
}

func safeDeleteFromChatMap(delMap map[int64]*ChatOutConnectionsType, userId int64) {
	ChatMutex.Lock()
	delete(delMap, userId)
	ChatMutex.Unlock()
}

func safeDeleteFromChatMapIn(delMap map[int64]int, userId int64) {
	log.Debug("safeDeleteFromChatMapIn %v %d", safeDeleteFromChatMapIn, userId)
	ChatMutex.Lock()
	delete(delMap, userId)
	ChatMutex.Unlock()
}
