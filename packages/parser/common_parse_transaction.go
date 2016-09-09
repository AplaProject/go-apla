package parser

import (
	"fmt"
	"reflect"
	"github.com/DayLightProject/go-daylight/packages/consts"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"github.com/DayLightProject/go-daylight/packages/lib"
)

func (p *Parser) ParseTransaction(transactionBinaryData *[]byte) ([][]byte, error) {

	var returnSlice [][]byte
	var transSlice [][]byte
	var merkleSlice [][]byte
	log.Debug("transactionBinaryData: %x", *transactionBinaryData)
	log.Debug("transactionBinaryData: %s", *transactionBinaryData)
	if len(*transactionBinaryData) > 0 {

		// хэш транзакции
		transSlice = append(transSlice, utils.DSha256(*transactionBinaryData))
		input := (*transactionBinaryData)[:]
		// первый байт - тип транзакции
		txType := utils.BinToDecBytesShift(transactionBinaryData, 1)
		isStruct := consts.IsStruct(int(txType))
		if isStruct {
			p.TxPtr = consts.MakeStruct(consts.TxTypes[int(txType)])
			if err := lib.BinUnmarshal(&input, p.TxPtr); err != nil {
				return nil, err
			}
			p.TxVars = make(map[string]string)
			head := consts.Header(p.TxPtr)
			p.TxCitizenID = head.CitizenId
			p.TxWalletID = head.WalletId
			p.TxTime = int64(head.Time)
			
			fmt.Println(`PARSED STRUCT %v`, p.TxPtr)
		} 
		transSlice = append(transSlice, utils.Int64ToByte(txType))
		if len(*transactionBinaryData) == 0 {
			return transSlice, utils.ErrInfo(fmt.Errorf("incorrect tx"))
		}
		// следующие 4 байта - время транзакции
		transSlice = append(transSlice, utils.Int64ToByte(utils.BinToDecBytesShift(transactionBinaryData, 4)))
		if len(*transactionBinaryData) == 0 {
			return transSlice, utils.ErrInfo(fmt.Errorf("incorrect tx"))
		}
		log.Debug("%s", transSlice)
		// преобразуем бинарные данные транзакции в массив
		if isStruct {
			t := reflect.ValueOf(p.TxPtr).Elem()
			
			//walletId & citizenId
			for i:=2; i<4; i++ {
				data := lib.FieldToBytes( t.Field(0).Interface(), i)	
				returnSlice = append(returnSlice, data) 
				merkleSlice = append(merkleSlice, utils.DSha256(data))
			}	
			for i:= 1; i < t.NumField(); i++ {
				data := lib.FieldToBytes( t.Interface(), i )
				returnSlice = append(returnSlice, data)
				merkleSlice = append(merkleSlice, utils.DSha256(data))
			}
		} else {
			i := 0
			for {
				length := utils.DecodeLength(transactionBinaryData)
				log.Debug("length: %d\n", length)
				if length > 0 && length < consts.MAX_TX_SIZE {
					data := utils.BytesShift(transactionBinaryData, length)
					returnSlice = append(returnSlice, data)
					merkleSlice = append(merkleSlice, utils.DSha256(data))
					log.Debug("%x", data)
					log.Debug("%s", data)
				}
				i++
				if length == 0 || i >= 20 { // у нас нет тр-ий с более чем 20 элементами
					break
				}
			}
		}
		if isStruct {
			*transactionBinaryData = (*transactionBinaryData)[len(*transactionBinaryData):]
		}
		if len(*transactionBinaryData) > 0 {
			return transSlice, utils.ErrInfo(fmt.Errorf("incorrect transactionBinaryData %x", transactionBinaryData))
		}
	} else {
		merkleSlice = append(merkleSlice, []byte("0"))
	}
	log.Debug("merkleSlice", merkleSlice)
	if len(merkleSlice) == 0 {
		merkleSlice = append(merkleSlice, []byte("0"))
	}
	p.MerkleRoot = utils.MerkleTreeRoot(merkleSlice)
	log.Debug("MerkleRoot %s\n", p.MerkleRoot)
	return append(transSlice, returnSlice...), nil
}