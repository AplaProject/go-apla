package tx

import "fmt"

type SmartContract struct {
	Header
	Data []byte
}

func (s SmartContract) ForSign() string {
	return fmt.Sprintf("%d,%d,%d,%d", s.Type, s.Time, s.UserID, s.StateID)
}
