package crypto

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"

	log "github.com/sirupsen/logrus"
)

type signProvider int

const (
	_ECDSA signProvider = iota
)

// Sign in signing data with private key
func Sign(privateKey string, data string) ([]byte, error) {
	if len(data) == 0 {
		log.WithFields(log.Fields{"type": consts.CryptoError}).Debug(ErrSigningEmpty.Error())
	}
	switch signProv {
	case _ECDSA:
		return signECDSA(privateKey, data)
	default:
		return nil, ErrUnknownProvider
	}
}

// CheckSign is checking sign
func CheckSign(public []byte, data string, signature []byte) (bool, error) {
	if len(public) == 0 {
		log.WithFields(log.Fields{"type": consts.CryptoError}).Debug(ErrCheckingSignEmpty.Error())
	}
	switch signProv {
	case _ECDSA:
		return checkECDSA(public, data, signature)
	default:
		return false, ErrUnknownProvider
	}
}

// JSSignToBytes converts hex signature which has got from the browser to []byte
func JSSignToBytes(in string) ([]byte, error) {
	r, s, err := parseSign(in)
	if err != nil {
		return nil, err
	}
	return append(converter.FillLeft(r.Bytes()), converter.FillLeft(s.Bytes())...), nil
}

func signECDSA(privateKey string, data string) (ret []byte, err error) {
	var pubkeyCurve elliptic.Curve

	switch ellipticSize {
	case elliptic256:
		pubkeyCurve = elliptic.P256()
	default:
		log.WithFields(log.Fields{"type": consts.CryptoError}).Fatal(ErrUnsupportedCurveSize.Error())
	}

	b, err := hex.DecodeString(privateKey)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.ConversionError, "error": err}).Error("decoding private key from hex")
		return
	}
	bi := new(big.Int).SetBytes(b)
	priv := new(ecdsa.PrivateKey)
	priv.PublicKey.Curve = pubkeyCurve
	priv.D = bi

	signhash, err := Hash([]byte(data))
	if err != nil {
		log.WithFields(log.Fields{"type": consts.CryptoError}).Fatal(ErrHashing.Error())
	}
	r, s, err := ecdsa.Sign(crand.Reader, priv, signhash)
	if err != nil {
		return
	}
	ret = append(converter.FillLeft(r.Bytes()), converter.FillLeft(s.Bytes())...)
	return
}

// CheckECDSA checks if forSign has been signed with corresponding to public the private key
func checkECDSA(public []byte, data string, signature []byte) (bool, error) {
	if len(data) == 0 {
		log.WithFields(log.Fields{"type": consts.CryptoError}).Error("data is empty")
		return false, fmt.Errorf("invalid parameters len(data) == 0")
	}
	if len(public) != consts.PubkeySizeLength {
		log.WithFields(log.Fields{"size": len(public), "size_match": consts.PubkeySizeLength, "type": consts.SizeDoesNotMatch}).Error("invalid public key")
		return false, fmt.Errorf("invalid parameters len(public) = %d", len(public))
	}
	if len(signature) == 0 {
		log.WithFields(log.Fields{"type": consts.CryptoError}).Error("invalid signature")
		return false, fmt.Errorf("invalid parameters len(signature) == 0")
	}

	var pubkeyCurve elliptic.Curve
	switch ellipticSize {
	case elliptic256:
		pubkeyCurve = elliptic.P256()
	default:
		log.WithFields(log.Fields{"type": consts.CryptoError}).Error(ErrUnsupportedCurveSize.Error())
	}

	hash, err := Hash([]byte(data))
	if err != nil {
		log.WithFields(log.Fields{"type": consts.CryptoError}).Error(ErrHashing.Error())
	}

	pubkey := new(ecdsa.PublicKey)
	pubkey.Curve = pubkeyCurve
	pubkey.X = new(big.Int).SetBytes(public[0:consts.PrivkeyLength])
	pubkey.Y = new(big.Int).SetBytes(public[consts.PrivkeyLength:])
	r, s, err := parseSign(hex.EncodeToString(signature))
	if err != nil {
		return false, err
	}
	verifystatus := ecdsa.Verify(pubkey, hash, r, s)
	if !verifystatus {
		return false, ErrIncorrectSign
	}
	return true, nil
}

// parseSign converts the hex signature to r and s big number
func parseSign(sign string) (*big.Int, *big.Int, error) {
	var (
		binSign []byte
		err     error
	)
	//	var off int
	parse := func(bsign []byte) []byte {
		blen := int(bsign[1])
		if blen > len(bsign)-2 {
			return nil
		}
		ret := bsign[2 : 2+blen]
		if len(ret) > 32 {
			ret = ret[len(ret)-32:]
		} else if len(ret) < 32 {
			ret = append(bytes.Repeat([]byte{0}, 32-len(ret)), ret...)
		}
		return ret
	}
	if len(sign) > 128 {
		binSign, err = hex.DecodeString(sign)
		if err != nil {
			log.WithFields(log.Fields{"type": consts.ConversionError, "error": err}).Error("decoding sign from string")
			return nil, nil, err
		}
		left := parse(binSign[2:])
		if left == nil || int(binSign[3])+6 > len(binSign) {
			log.WithFields(log.Fields{"type": consts.CryptoError}).Error("wrong left parsing")
			return nil, nil, fmt.Errorf(`wrong left parsing`)
		}
		right := parse(binSign[4+binSign[3]:])
		if right == nil {
			log.WithFields(log.Fields{"type": consts.CryptoError}).Error("wrong right parsing")
			return nil, nil, fmt.Errorf(`wrong right parsing`)
		}
		sign = hex.EncodeToString(append(left, right...))
	} else if len(sign) < 128 {
		log.WithFields(log.Fields{"size": len(sign), "size_match": 128}).Error("wrong signature size")
		return nil, nil, fmt.Errorf(`wrong len of signature %d`, len(sign))
	}
	all, err := hex.DecodeString(sign[:])
	if err != nil {
		log.WithFields(log.Fields{"size": len(sign), "size_match": 128}).Error("wrong signature size")
		return nil, nil, err
	}
	return new(big.Int).SetBytes(all[:32]), new(big.Int).SetBytes(all[len(all)-32:]), nil
}
