package crypto

import (
	"crypto/sha256"
	"fmt"
)

type hashProvider int

const (
	_SHA256 hashProvider = iota
)

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

func DoubleHash(msg []byte, version int) ([]byte, error) {
	if len(msg) == 0 {
		log.Debug(HashingEmpty.Error())
	}
	switch hashProv {
	case _SHA256:
		if version == 0 {
			return hashDoubleSHA256Old(msg), nil
		}
		return hashDoubleSHA256(msg), nil
	default:
		return nil, UnknownProviderError
	}
}

func hashSHA256(msg []byte) []byte {
	hash := sha256.Sum256(msg)
	return hash[:]
}

func hashDoubleSHA256Old(msg []byte) []byte {
	firstHash := sha256.Sum256(msg)
	secondHash := sha256.Sum256([]byte(fmt.Sprintf("%x", firstHash[:])))
	return secondHash[:]
}

func hashDoubleSHA256(msg []byte) []byte {
	firstHash := sha256.Sum256(msg)
	secondHash := sha256.Sum256(firstHash[:])
	return secondHash[:]

}
