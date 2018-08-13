package tx

// Header is contain header data
type Header struct {
	Type          int
	Time          int64
	EcosystemID   int64
	KeyID         int64
	RoleID        int64
	NetworkID     int64
	NodePosition  int64
	BlockID       int64
	Attempts      int64
	Error         string
	PublicKey     []byte
	BinSignatures []byte
}
