package tx

import "fmt"

// SmartContract is storing smart contract data
type SmartContract struct {
	Header
	TokenEcosystem int64
	MaxSum         string
	PayOver        string
	SignedBy       int64
	Data           []byte
}

// ForSign is converting SmartContract to string
func (s SmartContract) ForSign() string {
	return fmt.Sprintf("%d,%d,%d,%d,%d,%s,%s,%d", s.Type, s.Time, s.KeyID, s.EcosystemID,
		s.TokenEcosystem, s.MaxSum, s.PayOver, s.SignedBy)
}
