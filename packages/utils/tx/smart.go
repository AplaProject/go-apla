package tx

// SmartContract is storing smart contract data
type SmartContract struct {
	Header
	TokenEcosystem int64
	MaxSum         string
	PayOver        string
	SignedBy       int64
	Params         map[string]interface{}
}
