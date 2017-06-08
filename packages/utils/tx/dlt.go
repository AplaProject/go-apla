package tx

import "fmt"

type DLTTransfer struct {
	Header
	WalletAddress string
	Amount        string
	Commission    string
	Comment       string
}

type DLTChangeHostVote struct {
	Header
	Host        string
	AddressVote string
	FuelRate    string
}

type DLTChangeNodeKey struct {
	Header
	NewNodePublicKey []byte
}

func (c DLTChangeNodeKey) ForSign() string {
	return fmt.Sprintf("%s,%s,%s,%s", c.Header.Type, c.Header.Time, c.Header.UserID, c.NewNodePublicKey)
}
