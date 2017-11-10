package parser

import (
	"bytes"
	"fmt"

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"

	log "github.com/sirupsen/logrus"
)

func ParseOldTransaction(buffer *bytes.Buffer) ([][]byte, error) {
	var transSlice [][]byte

	transSlice = append(transSlice, []byte{})                                                  // hash placeholder
	transSlice = append(transSlice, []byte{})                                                  // type placeholder
	transSlice = append(transSlice, converter.Int64ToByte(converter.BinToDec(buffer.Next(4)))) // time

	if buffer.Len() == 0 {
		log.WithFields(log.Fields{"type": consts.EmptyObject}).Error("buffer is empty, while parsing old transaction")
		return transSlice, fmt.Errorf("incorrect tx")
	}

	for buffer.Len() > 0 {
		log.WithFields(log.Fields{"type": consts.EmptyObject}).Error("buffer is empty, while parsing old transaction")
		length, err := converter.DecodeLengthBuf(buffer)
		if err != nil {
			return nil, err
		}

		if length > buffer.Len() || length > consts.MAX_TX_SIZE {
			log.WithFields(log.Fields{"size": buffer.Len(), "max_size": consts.MAX_TX_SIZE, "decoded_size": length, "type": consts.ParameterExceeded}).Error("bad transaction")
			return nil, fmt.Errorf("bad transaction")
		}

		if length > 0 {
			transSlice = append(transSlice, buffer.Next(length))
			continue
		}

		if length == 0 && buffer.Len() > 0 {
			transSlice = append(transSlice, []byte{})
			continue
		}

		if length == 0 {
			log.WithFields(log.Fields{"type": consts.EmptyObject}).Error("bad transaction, length is 0")
			break
		}
	}

	return transSlice, nil
}
