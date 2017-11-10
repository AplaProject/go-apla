package tx

import "fmt"

type SmartContract struct {
	Header
	TokenEcosystem int64
	MaxSum         string
	PayOver        string
	Data           []byte
}

func (s SmartContract) ForSign() string {
	return fmt.Sprintf("%d,%d,%d,%d,%s,%s", s.Type, s.Time, s.KeyID,
		s.TokenEcosystem, s.MaxSum, s.PayOver)
}
