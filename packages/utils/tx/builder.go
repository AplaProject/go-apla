package tx

import (
	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/crypto"
	"github.com/GenesisKernel/go-genesis/packages/model"
	log "github.com/sirupsen/logrus"
	"gopkg.in/vmihailenco/msgpack.v2"
)

// BuildTransaction creates transaction
func BuildTransaction(smartTx SmartContract, privateKey []byte) error {
	publicKey, err := crypto.PrivateToPublic(privateKey)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.CryptoError, "error": err}).Error("converting node private key to public")
		return err
	}
	smartTx.PublicKey = publicKey
	smartTx.SignedBy = crypto.Address(publicKey)

	data, err := msgpack.Marshal(smartTx)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.MarshallingError, "error": err}).Error("marshalling smart contract to msgpack")
		return err
	}

	hash, err := crypto.DoubleHash(data)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.CryptoError, "error": err}).Error("calculating hash of smart contract")
		return err
	}

	signature, err := crypto.Sign(privateKey, hash)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.CryptoError, "error": err}).Error("signing by node private key")
		return err
	}

	data = append(append([]byte{128}, converter.EncodeLengthPlusData(data)...), converter.EncodeLengthPlusData(signature)...)

	tx := &model.Transaction{
		Hash:     hash,
		Data:     data[:],
		Type:     1,
		KeyID:    smartTx.KeyID,
		HighRate: model.TransactionRateOnBlock,
	}
	if err = tx.Create(); err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("creating new transaction")
		return err
	}

	return nil
}
