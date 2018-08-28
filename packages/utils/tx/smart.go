package tx

import (
	"fmt"

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/crypto"

	log "github.com/sirupsen/logrus"
	msgpack "gopkg.in/vmihailenco/msgpack.v2"
)

// SmartContract is storing smart contract data
type SmartContract struct {
	Header
	RequestID      string
	TokenEcosystem int64
	MaxSum         string
	PayOver        string
	SignedBy       int64
	Params         map[string]string
}

// ForSign is converting SmartContract to string
func (s SmartContract) ForSign() string {
	return fmt.Sprintf("%s,%d,%d,%d,%d,%d,%s,%s,%d", s.RequestID, s.Type, s.Time, s.KeyID, s.EcosystemID,
		s.TokenEcosystem, s.MaxSum, s.PayOver, s.SignedBy)
}

func (s SmartContract) Marshal() ([]byte, error) {
	var b []byte
	var err error
	if b, err = msgpack.Marshal(s); err != nil {
		log.WithFields(log.Fields{"type": consts.MarshallingError, "error": err}).Error("marshalling tx")
		return nil, err
	}
	return b, err
}

func (s *SmartContract) Unmarshal(b []byte) error {
	if err := msgpack.Unmarshal(b, s); err != nil {
		log.WithFields(log.Fields{"type": consts.UnmarshallingError, "error": err}).Error("unmarshalling tx")
		return err
	}
	return nil
}

func (s SmartContract) Hash() ([]byte, error) {
	b, err := s.Marshal()
	if err != nil {
		return nil, err
	}
	return crypto.DoubleHash(b)
}
