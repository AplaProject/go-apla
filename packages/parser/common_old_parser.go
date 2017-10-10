package parser

import (
	"bytes"
	"fmt"

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
)

func ParseOldTransaction(buffer *bytes.Buffer) ([][]byte, error) {
	var transSlice [][]byte

	transSlice = append(transSlice, []byte{})                                                  // hash placeholder
	transSlice = append(transSlice, []byte{})                                                  // type placeholder
	transSlice = append(transSlice, converter.Int64ToByte(converter.BinToDec(buffer.Next(4)))) // time

	if buffer.Len() == 0 {
		return transSlice, fmt.Errorf("incorrect tx")
	}

	for buffer.Len() > 0 {
		length, err := converter.DecodeLengthBuf(buffer)
		if err != nil {
			return nil, err
		}

		if length > buffer.Len() || length > consts.MAX_TX_SIZE {
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
			log.Debug("length == 0")
			break
		}
	}

	return transSlice, nil
}
