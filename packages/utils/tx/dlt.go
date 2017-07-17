package tx

import "fmt"

type DLTTransfer struct {
	Header
	WalletAddress string
	Amount        string
	Commission    string
	Comment       string
}

func (s DLTTransfer) ForSign() string {
	return fmt.Sprintf("%d,%d,%d,%s,%s,%s,%s", s.Header.Type, s.Header.Time, s.Header.UserID, s.WalletAddress, s.Amount, s.Commission, s.Comment)
}

type DLTChangeHostVote struct {
	Header
	Host        string
	AddressVote string
	FuelRate    string
}

func (s DLTChangeHostVote) ForSign() string {
	return fmt.Sprintf("%d,%d,%d,%s,%s,%s", s.Header.Type, s.Header.Time, s.Header.UserID, s.Host, s.AddressVote, s.FuelRate)
}

type DLTChangeNodeKey struct {
	Header
	NewNodePublicKey []byte
}

func (s DLTChangeNodeKey) ForSign() string {
	return fmt.Sprintf("%d,%d,%d,%s", s.Header.Type, s.Header.Time, s.Header.UserID, s.NewNodePublicKey)
}
