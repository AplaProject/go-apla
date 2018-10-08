package blockchain

import (
	"errors"
	"fmt"

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/crypto"
	"github.com/GenesisKernel/go-genesis/packages/utils"
	"github.com/syndtr/goleveldb/leveldb"

	log "github.com/sirupsen/logrus"
	msgpack "gopkg.in/vmihailenco/msgpack.v2"
)

const firstBlockKey = "first_block"
const lastBlockKey = "last_block"

func prefixFunc(prefix string) func([]byte) []byte {
	return func(hash []byte) []byte {
		return []byte(prefix + string(hash))
	}
}

var blockPrefix func([]byte) []byte = prefixFunc("block-")
var nextBlockHashPrefix func([]byte) []byte = prefixFunc("next_block_hash-")

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

func GetNextBlockHash(tx *leveldb.Transaction, hash []byte) ([]byte, bool, error) {
	val, err := GetDB(tx).Get(blockPrefix(hash), nil)
	if err == leveldb.ErrNotFound {
		return nil, false, nil
	}
	if err != nil {
		log.WithFields(log.Fields{"type": consts.LevelDBError, "error": err}).Error("getting next block hash")
		return nil, false, err
	}
	return val, true, nil
}

func SetNextBlockHash(tx *leveldb.Transaction, hash, nextBlockHash []byte) error {
	return GetDB(tx).Put(nextBlockHashPrefix(hash), nextBlockHash, nil)
}

func DelNextBlockHash(tx *leveldb.Transaction, hash []byte) error {
	return GetDB(tx).Delete(nextBlockHashPrefix(hash), nil)
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
	Transactions  []*Transaction
	MrklRoot      []byte
	PrevHash      []byte
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
		doubleHash, err := tr.Hash()
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

func (b *Block) Marshal() ([]byte, error) {
	mrklRoot, err := b.GetMrklRoot()
	NodePrivateKey, _, err := utils.GetNodeKeys()
	if err != nil {
		return nil, err
	}
	b.MrklRoot = mrklRoot
	sign, err := b.GetSign(NodePrivateKey)
	if err != nil {
		return nil, err
	}
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

func (b *Block) Get(tx *leveldb.Transaction, hash []byte) (bool, error) {
	val, err := GetDB(tx).Get(blockPrefix(hash), nil)
	if err == leveldb.ErrNotFound {
		return false, nil
	}
	if err != nil {
		log.WithFields(log.Fields{"type": consts.LevelDBError, "error": err}).Error("getting block")
		return false, err
	}
	if err := b.Unmarshal(val); err != nil {
		log.WithFields(log.Fields{"type": consts.UnmarshallingError, "error": err}).Error("unmarshalling transaction")
		return true, err
	}
	return true, nil
}

func GetNextBlock(tx *leveldb.Transaction, hash []byte) (*BlockWithHash, bool, error) {
	nextHash, found, err := GetNextBlockHash(tx, hash)
	if err != nil {
		return nil, false, err
	}
	if !found {
		return nil, false, nil
	}
	b := &Block{}
	found, err = b.Get(tx, nextHash)
	if err != nil {
		return nil, false, err
	}
	if !found {
		return nil, false, nil
	}
	return &BlockWithHash{Hash: nextHash, Block: b}, true, nil
}

func (b *Block) Hash() ([]byte, error) {
	bBlock, err := b.Marshal()
	if err != nil {
		return nil, err
	}
	blockHash, err := crypto.DoubleHash(bBlock)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.CryptoError, "error": err}).Error("double hashing block data")
		return nil, err
	}
	return blockHash, nil
}

func (b *Block) Insert(tx *leveldb.Transaction) error {
	prevHash := b.PrevHash
	hash, err := b.Hash()
	if err != nil {
		return err
	}
	val, err := b.Marshal()
	if err != nil {
		return err
	}
	err = GetDB(tx).Put(blockPrefix(hash), val, nil)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.LevelDBError, "error": err}).Error("inserting block")
		return err
	}
	if b.Header.BlockID == 1 {
		if err := GetDB(tx).Put([]byte(firstBlockKey), hash, nil); err != nil {
			log.WithFields(log.Fields{"type": consts.LevelDBError, "error": err}).Error("updating first block key")
			return err
		}
	}
	if err := GetDB(tx).Put([]byte(lastBlockKey), hash, nil); err != nil {
		log.WithFields(log.Fields{"type": consts.LevelDBError, "error": err}).Error("updating last block key")
		return err
	}
	if b.Header.BlockID != 1 {
		if err := SetNextBlockHash(tx, prevHash, hash); err != nil {
			return err
		}
	}
	for _, tr := range b.Transactions {
		txHash, err := tr.Hash()
		if err != nil {
			return err
		}
		txStatus := TxStatus{BlockID: b.Header.BlockID}
		if err := tr.Insert(tx); err != nil {
			return err
		}
		if err := txStatus.Insert(tx, txHash); err != nil {
			return err
		}
	}

	return nil
}

func DeleteBlocksFrom(tx *leveldb.Transaction, blockHash []byte) ([]*BlockWithHash, error) {
	blocks, err := GetNBlocksFrom(tx, blockHash, -1, -1)
	if err != nil {
		return nil, err
	}
	for _, b := range blocks {
		if err := GetDB(tx).Delete(blockPrefix(b.Hash), nil); err != nil {
			log.WithFields(log.Fields{"type": consts.LevelDBError, "error": err}).Error("deleting block")
			return nil, err
		}
		if err := GetDB(tx).Delete(nextBlockHashPrefix(b.Hash), nil); err != nil {
			log.WithFields(log.Fields{"type": consts.LevelDBError, "error": err}).Error("deleting next block hash prefix")
			return nil, err
		}
	}
	if err := GetDB(tx).Put([]byte(lastBlockKey), blockHash, nil); err != nil {
		log.WithFields(log.Fields{"type": consts.LevelDBError, "error": err}).Error("updating last block key")
		return nil, err
	}
	return blocks, nil
}

func GetFirstBlock(tx *leveldb.Transaction) (*BlockWithHash, bool, error) {
	hash, err := GetDB(tx).Get([]byte(firstBlockKey), nil)
	if err == leveldb.ErrNotFound {
		return nil, false, nil
	}
	if err != nil {
		log.WithFields(log.Fields{"type": consts.LevelDBError, "error": err}).Error("getting first block key")
		return nil, false, err
	}
	block := &Block{}
	found, err := block.Get(tx, hash)
	return &BlockWithHash{Block: block, Hash: hash}, found, err
}

func GetLastBlock(tx *leveldb.Transaction) (*Block, []byte, bool, error) {
	hash, err := GetDB(tx).Get([]byte(lastBlockKey), nil)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.LevelDBError, "error": err}).Error("getting last block key")
		return nil, nil, false, err
	}
	block := &Block{}
	found, err := block.Get(tx, hash)
	return block, hash, found, err
}

func GetNBlocksFrom(tx *leveldb.Transaction, hash []byte, n, ordering int) ([]*BlockWithHash, error) {
	if ordering == 0 {
		return nil, errors.New("ordering must be positive or negative, not 0")
	}
	result := []*BlockWithHash{}
	if n < 1 {
		return result, nil
	}
	lastBlock := &Block{}
	found, err := lastBlock.Get(tx, hash)
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
		nextHash, found, err = GetNextBlockHash(tx, hash)
		if err != nil {
			return nil, err
		}
	}
	if !found {
		return result, nil
	}
	if n > 0 {
		for i := n - 1; i > 0; i-- {
			block := &Block{}
			found, err := block.Get(tx, nextHash)
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
				nextHash, found, err = GetNextBlockHash(tx, nextHash)
			}
			if !found {
				return result, nil
			}
		}
	} else {
		for len(nextHash) != 0 {
			block := &Block{}
			found, err := block.Get(tx, nextHash)
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
				nextHash, found, err = GetNextBlockHash(tx, nextHash)
			}
			if !found {
				return result, nil
			}
		}
	}
	return result, nil
}

func GetFirstNBlocks(tx *leveldb.Transaction, n int) ([]*BlockWithHash, error) {
	return GetNBlocksFrom(tx, []byte(firstBlockKey), n, 1)
}

func GetLastNBlocks(tx *leveldb.Transaction, n int) ([]*BlockWithHash, error) {
	return GetNBlocksFrom(tx, []byte(lastBlockKey), n, -1)
}

func GetMaxForeignBlock(tx *leveldb.Transaction, keyID int64) (*Block, bool, error) {
	blocks, err := GetLastNBlocks(tx, lastNBlocksCount)
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
