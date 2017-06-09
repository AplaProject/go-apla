package tx

import "fmt"

type DLTTransfer struct {
	Header
	WalletAddress string
	Amount        string
	Commission    string
	Comment       string
}

func (d DLTTransfer) ForSign() string {
	return fmt.Sprintf("%s,%s,%d,%s,%s,%s,%s", d.Header.Type, d.Header.Time, d.Header.UserID, d.WalletAddress, d.Amount, d.Commission, d.Comment)
}

type DLTChangeHostVote struct {
	Header
	Host        string
	AddressVote string
	FuelRate    string
}

func (d DLTChangeHostVote) ForSign() string {
	return fmt.Sprintf("%s,%s,%d,%s,%s,%s", d.Header.Type, d.Header.Type, d.Header.UserID, d.Host, d.AddressVote, d.FuelRate)
}

type DLTChangeNodeKey struct {
	Header
	NewNodePublicKey []byte
}

func (c DLTChangeNodeKey) ForSign() string {
	return fmt.Sprintf("%s,%s,%s,%s", c.Header.Type, c.Header.Time, c.Header.UserID, c.NewNodePublicKey)
}
