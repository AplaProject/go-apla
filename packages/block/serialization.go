package block

/*
import (
	"github.com/GenesisKernel/go-genesis/packages/consts"
	log "github.com/sirupsen/logrus"
	"gopkg.in/vmihailenco/msgpack.v2"
)

// MarshallBlock is marshalling block
func MarshallBlock(header *utils.BlockData, trData [][]byte, prevHash []byte, key string) ([]byte, error) {
	var mrklArray [][]byte
	var blockDataTx []byte
	var signed []byte
	logger := log.WithFields(log.Fields{"block_id": header.BlockID, "block_hash": header.Hash, "block_time": header.Time, "block_version": header.Version, "block_wallet_id": header.KeyID, "block_state_id": header.EcosystemID})

	for _, tr := range trData {
		doubleHash, err := crypto.DoubleHash(tr)
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.CryptoError, "error": err}).Error("double hashing transaction")
			return nil, err
		}
		mrklArray = append(mrklArray, converter.BinToHex(doubleHash))
		blockDataTx = append(blockDataTx, converter.EncodeLengthPlusData(tr)...)
	}

	if key != "" {
		if len(mrklArray) == 0 {
			mrklArray = append(mrklArray, []byte("0"))
		}
		mrklRoot := utils.MerkleTreeRoot(mrklArray)

		forSign := fmt.Sprintf("0,%d,%x,%d,%d,%d,%d,%s",
			header.BlockID, prevHash, header.Time, header.EcosystemID, header.KeyID, header.NodePosition, mrklRoot)

		var err error
		signed, err = crypto.SignString(key, forSign)
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.CryptoError, "error": err}).Error("signing blocko")
			return nil, err
		}
	}

	var buf bytes.Buffer
	// fill header
	buf.Write(converter.DecToBin(header.Version, 2))
	buf.Write(converter.DecToBin(header.BlockID, 4))
	buf.Write(converter.DecToBin(header.Time, 4))
	buf.Write(converter.DecToBin(header.EcosystemID, 4))
	buf.Write(converter.EncodeLenInt64InPlace(header.KeyID))
	buf.Write(converter.DecToBin(header.NodePosition, 1))
	buf.Write(converter.EncodeLengthPlusData(signed))

	// data
	buf.Write(blockDataTx)

	return buf.Bytes(), nil
}

func UnmarshallBlock(blockBuffer *bytes.Buffer, firstBlock bool) (*Block, error) {
	header, err := utils.ParseBlockHeader(blockBuffer, !firstBlock)
	if err != nil {
		return nil, err
	}

	logger := log.WithFields(log.Fields{"block_id": header.BlockID, "block_time": header.Time, "block_wallet_id": header.KeyID,
		"block_state_id": header.EcosystemID, "block_hash": header.Hash, "block_version": header.Version})
	transactions := make([]*transaction.Transaction, 0)

	var mrklSlice [][]byte

	// parse transactions
	for blockBuffer.Len() > 0 {
		transactionSize, err := converter.DecodeLengthBuf(blockBuffer)
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.UnmarshallingError, "error": err}).Error("transaction size is 0")
			return nil, fmt.Errorf("bad block format (%s)", err)
		}
		if blockBuffer.Len() < int(transactionSize) {
			logger.WithFields(log.Fields{"size": blockBuffer.Len(), "match_size": int(transactionSize), "type": consts.SizeDoesNotMatch}).Error("transaction size does not matches encoded length")
			return nil, fmt.Errorf("bad block format (transaction len is too big: %d)", transactionSize)
		}

		if transactionSize == 0 {
			logger.WithFields(log.Fields{"type": consts.EmptyObject}).Error("transaction size is 0")
			return nil, fmt.Errorf("transaction size is 0")
		}

		bufTransaction := bytes.NewBuffer(blockBuffer.Next(int(transactionSize)))
		t, err := transaction.UnmarshallTransaction(bufTransaction)
		if err != nil {
			if t != nil && t.TxHash != nil {
				transaction.MarkTransactionBad(t.DbTransaction, t.TxHash, err.Error())
			}
			return nil, fmt.Errorf("parse transaction error(%s)", err)
		}
		t.BlockData = &header

		transactions = append(transactions, t)

		// build merkle tree
		if len(t.TxFullData) > 0 {
			dSha256Hash, err := crypto.DoubleHash(t.TxFullData)
			if err != nil {
				logger.WithFields(log.Fields{"type": consts.CryptoError, "error": err}).Error("double hashing tx full data")
				return nil, err
			}
			dSha256Hash = converter.BinToHex(dSha256Hash)
			mrklSlice = append(mrklSlice, dSha256Hash)
		}
	}

	if len(mrklSlice) == 0 {
		mrklSlice = append(mrklSlice, []byte("0"))
	}

	return &Block{
		Header:       header,
		Transactions: transactions,
		MrklRoot:     utils.MerkleTreeRoot(mrklSlice),
	}, nil
}
*/
