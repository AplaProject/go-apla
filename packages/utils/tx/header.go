package tx

// Header is contain header data
type Header struct {
	ID          int
	Time        int64
	EcosystemID int64
	KeyID       int64
	NetworkID   int64
	PublicKey   []byte
}
