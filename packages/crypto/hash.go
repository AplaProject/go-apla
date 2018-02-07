// MIT License
//
// Copyright (c) 2016-2018 GenesisKernel
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package crypto

import (
	"crypto/hmac"
	"crypto/sha256"

	"github.com/GenesisKernel/go-genesis/packages/consts"
	log "github.com/sirupsen/logrus"
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
		return nil, ErrUnknownProvider
	}
}

func Hash(msg []byte) ([]byte, error) {
	if len(msg) == 0 {
		log.WithFields(log.Fields{"type": consts.CryptoError, "error": ErrHashingEmpty.Error()}).Debug(ErrHashingEmpty.Error())
	}
	switch hashProv {
	case _SHA256:
		return hashSHA256(msg), nil
	default:
		return nil, ErrUnknownProvider
	}
}

func DoubleHash(msg []byte) ([]byte, error) {
	if len(msg) == 0 {
		log.WithFields(log.Fields{"type": consts.CryptoError, "error": ErrHashingEmpty.Error()}).Debug(ErrHashingEmpty.Error())
	}
	switch hashProv {
	case _SHA256:
		return hashDoubleSHA256(msg), nil
	default:
		return nil, ErrUnknownProvider
	}
}

func hashSHA256(msg []byte) []byte {
	if len(msg) == 0 {
		log.Debug(ErrHashingEmpty.Error())
	}
	hash := sha256.Sum256(msg)
	return hash[:]
}

//TODO Replace hashDoubleSHA256 with this method
func hashDoubleSHA3(msg []byte) ([]byte, error) {
	if len(msg) == 0 {
		log.Debug(ErrHashingEmpty.Error())
	}
	return hashSHA3256(msg), nil
}

//In the previous version of this function (api v 1.0) this func worked in another way.
//First, hash has been calculated from input data
//Second, obtained hash has been converted to hex
//Third, hex value has been hashed once more time
//In this variant second step is omitted.
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
