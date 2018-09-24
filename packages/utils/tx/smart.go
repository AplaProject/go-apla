package tx

// SmartContract is storing smart contract data
type SmartContract struct {
	Header
	RequestID      string
	TokenEcosystem int64
	MaxSum         string
	PayOver        string
	SignedBy       int64
	Params         map[string]interface{}
}
