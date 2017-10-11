package structs

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"fmt"
	"math/big"
	"time"
)

type Binary struct {
	Version    string
	Body       []byte
	Date       time.Time
	Sign       []byte
	Name       string
	StartBlock int64
}

func (b *Binary) CheckSign(public []byte) (bool, error) {
	if len(b.Sign) != 64 {
		return false, fmt.Errorf("invalid parameters len(signature) == 0")
	}

	pubkeyCurve := elliptic.P256()

	data := b.Body
	data = append(data, []byte(b.Date.String())...)
	data = append(data, []byte(b.Version)...)
	hash := sha256.Sum256(data)
	pubkey := new(ecdsa.PublicKey)
	pubkey.Curve = pubkeyCurve
	pubkey.X = new(big.Int).SetBytes(public[0:32])
	pubkey.Y = new(big.Int).SetBytes(public[32:])

	r := new(big.Int).SetBytes(b.Sign[:32])
	s := new(big.Int).SetBytes(b.Sign[32:])

	verifyStatus := ecdsa.Verify(pubkey, hash[:], r, s)
	if !verifyStatus {
		return false, nil
	}
	return true, nil
}
