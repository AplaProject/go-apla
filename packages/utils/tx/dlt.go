package tx

// Трансфер с кошелька на кошелек
type DLTTransfer struct {
	Header
	WalletAddress string
	Amount        string
	Commission    string
	Comment       string
}

// Изменить голосующий хост
type DLTChangeHostVote struct {
	Header
	Host        string
	AddressVote string
	FuelRate    string
}

type DLTChangeNodeKey struct {
	Header
}
