package tx

type Header struct {
	Type          int
	Time          int64
	EcosystemID   int64
	KeyID         int64
	NodePosition  int64
	PublicKey     []byte
	BinSignatures []byte
}
