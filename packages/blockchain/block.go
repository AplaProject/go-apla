package blockchain

import (
	"fmt"

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/crypto"
	"github.com/GenesisKernel/go-genesis/packages/utils"

	log "github.com/sirupsen/logrus"
	msgpack "gopkg.in/vmihailenco/msgpack.v2"
)

const blockPrefix = "block-"

type Block struct {
	Header       *utils.BlockData
	Transactions [][]byte
	MrklRoot     []byte
	PrevHash     []byte
	Sign         []byte
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
	return utils.MerkleTreeRoot(mrklArray), nil
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
