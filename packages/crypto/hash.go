package crypto

import (
	"crypto/hmac"
	"crypto/sha256"

	"golang.org/x/crypto/sha3"
)

type hashProvider int

const (
	_SHA256 hashProvider = iota
)

func GetHMAC(secret string, message string) ([]byte, error) {
	switch hmacProv {
	case _SHA256:
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write([]byte(message))
		return mac.Sum(nil), nil
	default:
		return nil, UnknownProviderError
	}
}

func Hash(msg []byte) ([]byte, error) {
	if len(msg) == 0 {
		log.Debug(HashingEmpty.Error())
	}
	switch hashProv {
	case _SHA256:
		return hashSHA256(msg), nil
	default:
		return nil, UnknownProviderError
	}
}

func DoubleHash(msg []byte) ([]byte, error) {
	if len(msg) == 0 {
		log.Debug(HashingEmpty.Error())
	}
	switch hashProv {
	case _SHA256:
		return hashDoubleSHA256(msg), nil
	default:
		return nil, UnknownProviderError
	}
}

func hashSHA256(msg []byte) []byte {
	if len(msg) == 0 {
		log.Debug(HashingEmpty.Error())
	}
	hash := sha256.Sum256(msg)
	return hash[:]
}

//TODO Replace hashDoubleSHA256 with this method
func hashDoubleSHA3(msg []byte) ([]byte, error) {
	if len(msg) == 0 {
		log.Debug(HashingEmpty.Error())
	}
	return hashSHA3256(msg), nil
}

//In the previous version of this function (api v 1.0) this func worked in another way.
//First, hash has been calculated from input data
//Second, obtained hash has been converted to hex
//Third, hex value has been hashed once more time
//In this variant second step is omited.
func hashDoubleSHA256(msg []byte) []byte {
	firstHash := sha256.Sum256(msg)
	secondHash := sha256.Sum256(firstHash[:])
	return secondHash[:]
}

func hashSHA3256(msg []byte) []byte {
	hash := make([]byte, 64)
	sha3.ShakeSum256(hash, msg)
	return hash[:]
}
