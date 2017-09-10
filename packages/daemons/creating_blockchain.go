package daemons

import (
	"github.com/EGaaS/go-egaas-mvp/packages/consts"
	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/model"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"

	"context"
	"io"
	"os"
	"time"

	"github.com/EGaaS/go-egaas-mvp/packages/config/syspar"
	logger "github.com/EGaaS/go-egaas-mvp/packages/log"
)

func CreatingBlockchain(d *daemon, ctx context.Context) error {
	logger.LogDebug(consts.FuncStarted, "")
	d.sleepTime = 10 * time.Second
	return writeNextBlocks(*utils.Dir+"/public/blockchain", consts.COUNT_BLOCK_BEFORE_SAVE)
}

func writeNextBlocks(fileName string, minToSave int) error {
	logger.LogDebug(consts.FuncStarted, "")
	lastSavedBlockID, err := getLastBlockID(fileName)
	if err != nil {
		logger.LogError(consts.DBError, err)
		return err
	}

	infoBlock := &model.InfoBlock{}
	err = infoBlock.GetInfoBlock()
	if err != nil {
		logger.LogError(consts.DBError, err)
		return err
	}

	curBlockID := infoBlock.BlockID

	if curBlockID-int64(minToSave) < lastSavedBlockID {
		// not enough blocks to save, just return
		return nil
	}

	// write the newest blocks to reserved blockchain
	// ??? curBlockID - COUNT_BLOCK_BEFORE_SAVE ???
	blocks, err := model.GetBlockchain(lastSavedBlockID, lastSavedBlockID+int64(minToSave))
	if err != nil {
		logger.LogError(consts.DBError, err)
		return err
	}

	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		logger.LogError(consts.IOError, err)
		return err
	}
	defer file.Close()

	for _, b := range blocks {
		buff := marshallFileBlock(blockData{ID: b.ID, Data: b.Data})

		_, err := file.Write(buff)
		if err != nil {
			logger.LogError(consts.IOError, err)
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
	WordSize = 5 // size of word in file
)

type blockData struct {
	ID   int64
	Data []byte
}

func readBlock(r io.Reader) (*blockData, error) {
	logger.LogDebug(consts.FuncStarted, "")
	var err error
	buf := make([]byte, WordSize)

	if _, err = io.ReadFull(r, buf); err != nil {
		logger.LogError(consts.IOError, err)
		return nil, err
	}

	size := converter.BinToDec(buf)
	if size > syspar.GetMaxBlockSize() {
		logger.LogError(consts.BlockError, "size > conts.MAX_BLOCK_SIZE")
		return nil, utils.ErrInfo("size > conts.MAX_BLOCK_SIZE")
	}

	if size == 0 {
		return nil, nil
	}

	dataBinary := make([]byte, size+WordSize)
	if _, err = r.Read(dataBinary); err != nil {
		logger.LogError(consts.IOError, err)
		return nil, utils.ErrInfo(err)
	}

	// parse the block
	block, err := unmarshalBlockData(dataBinary)
	if err != nil {
		logger.LogError(consts.BlockError, err)
		return nil, utils.ErrInfo(err)
	}

	return &block, nil
}

// read last block from file
func getLastBlockID(fileName string) (int64, error) {
	logger.LogDebug(consts.FuncStarted, "")
	file, err := os.Open(fileName)
	if err != nil {
		// if file doesn't exist create new one
		if os.IsNotExist(err) {
			return 0, nil
		}
		logger.LogError(consts.IOError, err)
		return 0, utils.ErrInfo(err)
	}
	defer file.Close()

	fi, err := file.Stat()
	if err != nil {
		logger.LogError(consts.IOError, err)
	}
	if fi.Size() == 0 {
		logger.LogError(consts.BlockchainLoadError, "empty blockchain file")
		return 0, utils.ErrInfo("empty blockchain file")
	}

	// size of a block recorded into the last 5 bytes of blockchain file
	_, err = file.Seek(-WordSize, os.SEEK_END)
	if err != nil {
		logger.LogError(consts.BlockchainLoadError, err)
		return 0, utils.ErrInfo(err)
	}

	buf := make([]byte, WordSize)
	_, err = file.Read(buf)
	if err != nil {
		logger.LogError(consts.IOError, err)
		return 0, utils.ErrInfo(err)
	}
	size := converter.BinToDec(buf)
	if size > syspar.GetMaxBlockSize() {
		logger.LogError(consts.BlockchainLoadError, "size > conts.MAX_BLOCK_SIZE")
		return 0, utils.ErrInfo("size > conts.MAX_BLOCK_SIZE")
	}

	// read the block
	_, err = file.Seek(-(size + WordSize), os.SEEK_END)
	if err != nil {
		logger.LogError(consts.BlockchainLoadError, err)
		return 0, utils.ErrInfo(err)
	}

	block, err := readBlock(file)
	if err != nil {
		logger.LogError(consts.IOError, err)
		return 0, utils.ErrInfo(err)
	}

	return block.ID, nil
}

func unmarshalBlockData(buff []byte) (blockData, error) {
	logger.LogDebug(consts.FuncStarted, "")
	blockID := converter.BinToDec(buff[:WordSize])
	buff = buff[WordSize:]

	// DecodeLength moves the pointer to the data field
	blockDataLen, err := converter.DecodeLength(&buff)
	if err != nil {
		logger.LogError(consts.BlockError, err)
		return blockData{}, utils.ErrInfo(err)
	}

	if blockDataLen > int64(len(buff)) {
		logger.LogError(consts.BlockError, "bad length")
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
