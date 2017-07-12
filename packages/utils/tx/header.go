package tx

type Header struct {
	Type          int
	Time          int64
	UserID        int64
	StateID       int64
	PublicKey     []byte
	BinSignatures []byte
}
