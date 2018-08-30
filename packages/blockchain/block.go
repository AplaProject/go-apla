package blockchain

import (
	"errors"
	"fmt"

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/crypto"
	"github.com/syndtr/goleveldb/leveldb"

	log "github.com/sirupsen/logrus"
	msgpack "gopkg.in/vmihailenco/msgpack.v2"
)

const blockPrefix = "block-"
const firstBlockKey = "first_block"
const lastBlockKey = "last_block"
const lastNBlocksCount = 5

// BlockData is a structure of the block's header
type BlockHeader struct {
	BlockID       int64
	Time          int64
	EcosystemID   int64
	KeyID         int64
	NodePosition  int64
	Sign          []byte
	RollbacksHash []byte
	Version       int
}

func (b BlockHeader) String() string {
	return fmt.Sprintf("BlockID:%d, Time:%d, NodePosition %d", b.BlockID, b.Time, b.NodePosition)
}

// MerkleTreeRoot rertun Merkle value
func MerkleTreeRoot(dataArray [][]byte) []byte {
	log.Debug("dataArray: %s", dataArray)
	result := make(map[int32][][]byte)
	for _, v := range dataArray {
		hash, err := crypto.DoubleHash(v)
		if err != nil {
			log.WithFields(log.Fields{"error": err, "type": consts.CryptoError}).Fatal("double hasing value, while calculating merkle tree root")
		}
		hash = converter.BinToHex(hash)
		result[0] = append(result[0], hash)
	}
	var j int32
	for len(result[j]) > 1 {
		for i := 0; i < len(result[j]); i = i + 2 {
			if len(result[j]) <= (i + 1) {
				if _, ok := result[j+1]; !ok {
					result[j+1] = [][]byte{result[j][i]}
				} else {
					result[j+1] = append(result[j+1], result[j][i])
				}
			} else {
				if _, ok := result[j+1]; !ok {
					hash, err := crypto.DoubleHash(append(result[j][i], result[j][i+1]...))
					if err != nil {
						log.WithFields(log.Fields{"error": err, "type": consts.CryptoError}).Fatal("double hasing value, while calculating merkle tree root")
					}
					hash = converter.BinToHex(hash)
					result[j+1] = [][]byte{hash}
				} else {
					hash, err := crypto.DoubleHash([]byte(append(result[j][i], result[j][i+1]...)))
					if err != nil {
						log.WithFields(log.Fields{"error": err, "type": consts.CryptoError}).Fatal("double hasing value, while calculating merkle tree root")
					}
					hash = converter.BinToHex(hash)
					result[j+1] = append(result[j+1], hash)
				}
			}
		}
		j++
	}

	ret := result[int32(len(result)-1)]
	return []byte(ret[0])
}

type Block struct {
	Header        *BlockHeader
	Transactions  [][]byte
	MrklRoot      []byte
	PrevHash      []byte
	NextHash      []byte
	RollbacksHash []byte
	Sign          []byte
}

type BlockWithHash struct {
	*Block
	Hash []byte
}

func (b *Block) GetMrklRoot() ([]byte, error) {
	var mrklArray [][]byte
	for _, tr := range b.Transactions {
		doubleHash, err := crypto.DoubleHash(tr)
		if err != nil {
			log.WithFields(log.Fields{"type": consts.CryptoError, "error": err}).Error("double hashing transaction")
			return nil, err
		}
		mrklArray = append(mrklArray, converter.BinToHex(doubleHash))
	}
	if len(mrklArray) == 0 {
		mrklArray = append(mrklArray, []byte("0"))
	}
	return MerkleTreeRoot(mrklArray), nil
}

func (b Block) ForSign() string {
	return fmt.Sprintf("0,%d,%x,%d,%d,%d,%d,%s",
		b.Header.BlockID, b.PrevHash, b.Header.Time, b.Header.EcosystemID, b.Header.KeyID, b.Header.NodePosition, b.MrklRoot)
}

func (b *Block) GetSign(key string) ([]byte, error) {
	forSign := b.ForSign()
	signed, err := crypto.Sign(key, forSign)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.CryptoError, "error": err}).Error("signing block")
		return nil, err
	}
	return signed, nil
}

func (b *Block) Marshal(key string) ([]byte, error) {
	mrklRoot, err := b.GetMrklRoot()
	sign, err := b.GetSign(key)
	if err != nil {
		return nil, err
	}
	b.MrklRoot = mrklRoot
	b.Sign = sign
	if res, err := msgpack.Marshal(b); err != nil {
		log.WithFields(log.Fields{"error": err, "type": consts.MarshallingError}).Error("marshalling block")
		return nil, err
	} else {
		return res, err
	}
}

func (b *Block) Unmarshal(bt []byte) error {
	if err := msgpack.Unmarshal(bt, &b); err != nil {
		log.WithFields(log.Fields{"error": err, "type": consts.UnmarshallingError}).Error("unmarshalling block")
		return err
	}
	return nil
}

func GetBlock(hash []byte) (*Block, bool, error) {
	val, err := db.Get([]byte(blockPrefix+string(hash)), nil)
	if err == leveldb.ErrNotFound {
		return nil, false, nil
	}
	if err != nil {
		log.WithFields(log.Fields{"type": consts.LevelDBError, "error": err}).Error("getting block")
		return nil, false, err
	}
	block := &Block{}
	if err := block.Unmarshal(val); err != nil {
		log.WithFields(log.Fields{"type": consts.UnmarshallingError, "error": err}).Error("unmarshalling transaction")
		return nil, true, err
	}
	return block, true, nil
}

func InsertBlock(hash []byte, block *Block, key string) error {
	prevHash := block.PrevHash
	val, err := block.Marshal(key)
	if err != nil {
		return err
	}
	err = db.Put([]byte(blockPrefix+string(hash)), val, nil)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.LevelDBError, "error": err}).Error("inserting block")
		return err
	}
	if block.Header.BlockID == 1 {
		if err := db.Put([]byte(firstBlockKey), []byte(blockPrefix+string(hash)), nil); err != nil {
			log.WithFields(log.Fields{"type": consts.LevelDBError, "error": err}).Error("updating first block key")
			return err
		}
	}
	if err := db.Put([]byte(lastBlockKey), []byte(blockPrefix+string(hash)), nil); err != nil {
		log.WithFields(log.Fields{"type": consts.LevelDBError, "error": err}).Error("updating last block key")
		return err
	}
	prevBlock, found, err := GetBlock(prevHash)
	if err != nil {
		return err
	}
	if !found {
		return nil
	}
	prevBlock.NextHash = hash
	prevBlockVal, err := prevBlock.Marshal(key)
	if err != nil {
		return err
	}
	if err := db.Put([]byte(blockPrefix+string(prevHash)), prevBlockVal, nil); err != nil {
		log.WithFields(log.Fields{"type": consts.LevelDBError, "error": err}).Error("inserting block")
		return err
	}

	return nil
}

func DeleteBlock(blockHash []byte) error {
	return db.Delete([]byte(blockPrefix+string(blockHash)), nil)
}

func GetFirstBlock() (*Block, []byte, bool, error) {
	hash, err := db.Get([]byte(firstBlockKey), nil)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.LevelDBError, "error": err}).Error("getting first block key")
		return nil, nil, false, err
	}
	block, found, err := GetBlock([]byte(blockPrefix + string(hash)))
	return block, hash, found, err
}

func GetLastBlock() (*Block, []byte, bool, error) {
	hash, err := db.Get([]byte(lastBlockKey), nil)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.LevelDBError, "error": err}).Error("getting last block key")
		return nil, nil, false, err
	}
	block, found, err := GetBlock([]byte(blockPrefix + string(hash)))
	return block, hash, found, err
}

func GetNBlocksFrom(hash []byte, n, ordering int) ([]*BlockWithHash, error) {
	if ordering == 0 {
		return nil, errors.New("ordering must be positive or negative, not 0")
	}
	result := []*BlockWithHash{}
	if n < 1 {
		return result, nil
	}
	lastBlock, found, err := GetBlock(hash)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, err
	}
	result = append(result, &BlockWithHash{Block: lastBlock, Hash: hash})
	var nextHash []byte
	if ordering < 0 {
		nextHash = lastBlock.PrevHash
	}
	if ordering > 0 {
		nextHash = lastBlock.NextHash
	}
	if len(nextHash) == 0 {
		return result, nil
	}
	for i := n - 1; i > 0; i-- {
		block, found, err := GetBlock(nextHash)
		if err != nil {
			return nil, err
		}
		if !found {
			return result, nil
		}
		result = append(result, &BlockWithHash{Block: block, Hash: nextHash})
		if ordering < 0 {
			nextHash = block.PrevHash
		} else if ordering > 0 {
			nextHash = block.NextHash
		}
		if len(nextHash) == 0 {
			return result, nil
		}
	}
	return result, nil
}

func GetFirstNBlocks(n int) ([]*BlockWithHash, error) {
	return GetNBlocksFrom([]byte(firstBlockKey), n, 1)
}

func GetLastNBlocks(n int) ([]*BlockWithHash, error) {
	return GetNBlocksFrom([]byte(lastBlockKey), n, -1)
}

func GetMaxForeignBlock(keyID int64) (*Block, bool, error) {
	blocks, err := GetLastNBlocks(lastNBlocksCount)
	if err != nil {
		return nil, false, err
	}
	for _, b := range blocks {
		if b.Block.Header.KeyID != keyID {
			return b.Block, true, nil
		}
	}
	return nil, false, nil
}
