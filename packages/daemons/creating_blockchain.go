// Copyright (C) 2017, 2018, 2019 EGAAS S.A.
//
// This program is free software; you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation; either version 2 of the License, or (at
// your option) any later version.
//
// This program is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
// General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program; if not, write to the Free Software
// Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301, USA.

package daemons

import (
	"io"
	"os"

	"github.com/AplaProject/go-apla/packages/conf/syspar"
	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/model"
	"github.com/AplaProject/go-apla/packages/utils"

	log "github.com/sirupsen/logrus"
)

func writeNextBlocks(fileName string, minToSave int64, logger *log.Entry) error {
	lastSavedBlockID, err := getLastBlockID(fileName, logger)
	if err != nil {
		return err
	}

	infoBlock := &model.InfoBlock{}
	_, err = infoBlock.Get()
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting info block")
		return err
	}

	curBlockID := infoBlock.BlockID

	if curBlockID-minToSave < lastSavedBlockID {
		// not enough blocks to save, just return
		return nil
	}

	// write the newest blocks to reserved blockchain
	// ??? curBlockID - COUNT_BLOCK_BEFORE_SAVE ???
	blocks, err := model.GetBlockchain(lastSavedBlockID, lastSavedBlockID+int64(minToSave), model.OrderASC)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting blockchain")
		return err
	}

	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("opening file, to write blocks")
		return err
	}
	defer file.Close()

	for _, b := range blocks {
		buff := marshallFileBlock(blockData{ID: b.ID, Data: b.Data})

		_, err := file.Write(buff)
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("writing block to file")
			return err
		}
	}

	return nil
}

/*
Block record format:
 block len - 5 bytes
 block id - 5 bytes
 block data len - variable
 block data - block data len bytes
 full len - 5 bytes (for read from end of file)
*/

const (
	// WordSize is size of word in file
	WordSize = 5
)

type blockData struct {
	ID   int64
	Data []byte
}

func readBlock(r io.Reader, logger *log.Entry) (*blockData, error) {
	var err error
	buf := make([]byte, WordSize)

	if _, err = io.ReadFull(r, buf); err != nil {
		logger.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("reading block from file")
		return nil, err
	}

	size := converter.BinToDec(buf)
	if size > syspar.GetMaxBlockSize() {
		logger.WithFields(log.Fields{"size": size, "max_size": syspar.GetMaxBlockSize(), "type": consts.ParameterExceeded}).Error("reading block from file")
		return nil, utils.ErrInfo("size > conts.MAX_BLOCK_SIZE")
	}

	if size == 0 {
		return nil, nil
	}

	dataBinary := make([]byte, size+WordSize)
	if _, err = r.Read(dataBinary); err != nil {
		logger.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("reading block from file")
		return nil, utils.ErrInfo(err)
	}

	// parse the block
	block, err := unmarshalBlockData(dataBinary, logger)
	if err != nil {
		return nil, utils.ErrInfo(err)
	}

	return &block, nil
}

// read last block from file
func getLastBlockID(fileName string, logger *log.Entry) (int64, error) {
	file, err := os.Open(fileName)
	if err != nil {
		logger.WithFields(log.Fields{"error": err, "type": consts.IOError}).Error("opening last block file")
		// if file doesn't exist create new one
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, utils.ErrInfo(err)
	}
	defer file.Close()

	fi, err := file.Stat()
	if err != nil {
		logger.WithFields(log.Fields{"error": err, "type": consts.IOError}).Fatal("stat last block file")
	}
	if fi.Size() == 0 {
		logger.WithFields(log.Fields{"error": err, "type": consts.IOError}).Error("last block file is empty")
		return 0, utils.ErrInfo("empty blockchain file")
	}

	// size of a block recorded into the last 5 bytes of blockchain file
	_, err = file.Seek(-WordSize, os.SEEK_END)
	if err != nil {
		logger.WithFields(log.Fields{"error": err, "type": consts.IOError}).Error("seek last block file")
		return 0, utils.ErrInfo(err)
	}

	buf := make([]byte, WordSize)
	_, err = file.Read(buf)
	if err != nil {
		logger.WithFields(log.Fields{"error": err, "type": consts.IOError}).Error("reading size from last block file")
		return 0, utils.ErrInfo(err)
	}
	size := converter.BinToDec(buf)
	if size > syspar.GetMaxBlockSize() {
		logger.WithFields(log.Fields{"size": size, "max_size": syspar.GetMaxBlockSize(), "type": consts.ParameterExceeded}).Error("block size is more than max size")
		return 0, utils.ErrInfo("size > conts.MAX_BLOCK_SIZE")
	}

	// read the block
	_, err = file.Seek(-(size + WordSize), os.SEEK_END)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("seeking the block from last block file")
		return 0, utils.ErrInfo(err)
	}

	block, err := readBlock(file, logger)
	if err != nil {
		return 0, utils.ErrInfo(err)
	}

	return block.ID, nil
}

func unmarshalBlockData(buff []byte, logger *log.Entry) (blockData, error) {

	blockID := converter.BinToDec(buff[:WordSize])
	buff = buff[WordSize:]

	// DecodeLength moves the pointer to the data field
	blockDataLen, err := converter.DecodeLength(&buff)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.UnmarshallingError, "error": err}).Error("decoding block length")
		return blockData{}, utils.ErrInfo(err)
	}

	if blockDataLen > int64(len(buff)) {
		logger.WithFields(log.Fields{"type": consts.UnmarshallingError, "length": blockDataLen, "real_length": int64(len(buff))}).Error("decoding block length")
		return blockData{}, utils.ErrInfo("bad length")
	}

	return blockData{
		ID:   blockID,
		Data: buff[:blockDataLen],
	}, nil
}

func marshallFileBlock(b blockData) []byte {
	data := append(converter.DecToBin(b.ID, WordSize), converter.EncodeLengthPlusData(b.Data)...)
	sizeAndData := append(converter.DecToBin(len(data), WordSize), data...)

	blockBin := append(sizeAndData, converter.DecToBin(len(sizeAndData), WordSize)...)
	return blockBin
}
