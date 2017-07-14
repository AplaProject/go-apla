package tx

import "fmt"

type RestoreAccess struct {
	Header
	StateID int64
}

func (s RestoreAccess) ForSign() string {
	return fmt.Sprintf("%d,%d,%d,%d,%d", s.Header.Type, s.Header.Time, s.Header.UserID, s.Header.StateID, s.StateID)
}

type RestoreAccessActive struct {
	Header
	Secret []byte
}

func (s RestoreAccessActive) ForSign() string {
	return fmt.Sprintf("%d,%d,%d,%d,%s", s.Header.Type, s.Header.Time, s.Header.UserID, s.Header.StateID, s.Secret)
}

type RestoreAccessClose struct {
	Header
}

func (s RestoreAccessClose) ForSign() string {
	return fmt.Sprintf("%d,%d,%d,%d", s.Header.Type, s.Header.Time, s.Header.UserID, s.Header.StateID)
}

type RestoreAccessRequest struct {
	Header
}

func (s RestoreAccessRequest) ForSign() string {
	return fmt.Sprintf("%d,%d,%d,%d", s.Header.Type, s.Header.Time, s.Header.UserID, s.Header.StateID)
}
