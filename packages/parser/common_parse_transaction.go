package parser

import (
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/consts"
	"github.com/DayLightProject/go-daylight/packages/utils"
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

		// первый байт - тип транзакции
		transSlice = append(transSlice, utils.Int64ToByte(utils.BinToDecBytesShift(transactionBinaryData, 1)))
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