package tx

import "fmt"

type UpdFullNodes struct {
	Header
}

func (s UpdFullNodes) ForSign() string {
	return fmt.Sprintf("%d,%d,%d,%d", s.Header.Type, s.Header.Time, s.Header.UserID, 0)
}
