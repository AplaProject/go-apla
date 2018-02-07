//go:generate sh -c "mockery -inpkg -name Signer -print > file.tmp && mv file.tmp signer_mock.go"
package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"math/big"

	"github.com/GenesisCommunity/go-genesis/tools/update_server/model"
)

type Signer interface {
	MakeSign(build model.Build) ([]byte, error)
	CheckSign(build model.Build, public []byte) (bool, error)
}

type BuildSigner struct {
	privateKey []byte
}

func NewBuildSigner(privateKey []byte) BuildSigner {
	return BuildSigner{privateKey: privateKey}
}

// Make sign is signing build with private key
func (bs *BuildSigner) MakeSign(build model.Build) ([]byte, error) {
	var sign []byte
	var pubkeyCurve elliptic.Curve

	pubkeyCurve = elliptic.P256()

	bi := new(big.Int).SetBytes(bs.privateKey)
	priv := new(ecdsa.PrivateKey)
	priv.PublicKey.Curve = pubkeyCurve
	priv.D = bi

	data := build.Body
	data = append(data, []byte(build.Time.String())...)
	data = append(data, []byte(build.Version.String())...)

	signhash := sha256.Sum256(data)
	r, s, err := ecdsa.Sign(rand.Reader, priv, signhash[:])
	if err != nil {
		return sign, err
	}
	return append(r.Bytes(), s.Bytes()...), nil
}

// CheckSign is checking build sign with public key
func (bs *BuildSigner) CheckSign(build model.Build, public []byte) (bool, error) {
	if len(build.Sign) != 64 {
		return false, fmt.Errorf("invalid parameters len(signature) == 0")
	}

	pubkeyCurve := elliptic.P256()

	data := build.Body
	data = append(data, []byte(build.Time.String())...)
	data = append(data, []byte(build.Version.String())...)
	hash := sha256.Sum256(data)
	pubkey := new(ecdsa.PublicKey)
	pubkey.Curve = pubkeyCurve
	pubkey.X = new(big.Int).SetBytes(public[0:32])
	pubkey.Y = new(big.Int).SetBytes(public[32:])

	r := new(big.Int).SetBytes(build.Sign[:32])
	s := new(big.Int).SetBytes(build.Sign[32:])

	verifyStatus := ecdsa.Verify(pubkey, hash[:], r, s)
	if !verifyStatus {
		return false, nil
	}
	return true, nil
}
