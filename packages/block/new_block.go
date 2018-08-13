package block

import (
	"bytes"
	"fmt"

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/crypto"
	"github.com/GenesisKernel/go-genesis/packages/transaction"
	"github.com/GenesisKernel/go-genesis/packages/utils"

	log "github.com/sirupsen/logrus"
	msgpack "gopkg.in/vmihailenco/msgpack.v2"
)

type NewBlock struct {
	Header       *utils.BlockData
	Transactions [][]byte
	MrklRoot     []byte
	PrevHash     []byte
	Sign         []byte
}

func (nb *NewBlock) GetMrklRoot() ([]byte, error) {
	var mrklArray [][]byte
	for _, tr := range nb.Transactions {
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

func (nb NewBlock) ForSign() string {
	return fmt.Sprintf("0,%d,%x,%d,%d,%d,%d,%s",
		nb.Header.BlockID, nb.PrevHash, nb.Header.Time, nb.Header.EcosystemID, nb.Header.KeyID, nb.Header.NodePosition, nb.MrklRoot)
}

func (nb *NewBlock) GetSign(key string) ([]byte, error) {
	forSign := nb.ForSign()
	signed, err := crypto.Sign(key, forSign)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.CryptoError, "error": err}).Error("signing block")
		return nil, err
	}
	return signed, nil
}

func (nb *NewBlock) Marshal(key string) ([]byte, error) {
	mrklRoot, err := nb.GetMrklRoot()
	sign, err := nb.GetSign(key)
	if err != nil {
		return nil, err
	}
	nb.MrklRoot = mrklRoot
	nb.Sign = sign
	if b, err := msgpack.Marshal(nb); err != nil {
		log.WithFields(log.Fields{"error": err, "type": consts.MarshallingError}).Error("marshalling block")
		return nil, err
	} else {
		return b, err
	}
}

func (nb *NewBlock) Unmarshal(b []byte) error {
	if err := msgpack.Unmarshal(b, &nb); err != nil {
		log.WithFields(log.Fields{"error": err, "type": consts.UnmarshallingError}).Error("unmarshalling block")
		return err
	}
	return nil
}

func (nb *NewBlock) ToBlock() (*Block, error) {
	transactions := make([]*transaction.Transaction, 0)
	for _, tx := range nb.Transactions {
		bufTransaction := bytes.NewBuffer(tx)
		t, err := transaction.UnmarshallTransaction(bufTransaction)
		if err != nil {
			if t != nil && t.TxHash != nil {
				transaction.MarkTransactionBad(t.DbTransaction, t.TxHash, err.Error())
			}
			return nil, fmt.Errorf("parse transaction error(%s)", err)
		}
		t.BlockData = nb.Header
		transactions = append(transactions, t)
	}
	return &Block{
		Header:       *nb.Header,
		Transactions: transactions,
		MrklRoot:     nb.MrklRoot,
	}, nil
}
