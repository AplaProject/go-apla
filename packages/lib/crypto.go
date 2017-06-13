// Copyright 2016 The go-daylight Authors
// This file is part of the go-daylight library.
//
// The go-daylight library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-daylight library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-daylight library. If not, see <http://www.gnu.org/licenses/>.

package lib

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"math/big"
)

// GetSharedKey creates and returns the shared key = private * public.
// public must be the public key from the different private key.
func GetSharedKey(private, public []byte) (shared []byte, err error) {
	pubkeyCurve := elliptic.P256()

	private = FillLeft(private)
	public = FillLeft(public)
	pub := new(ecdsa.PublicKey)
	pub.Curve = pubkeyCurve
	pub.X = new(big.Int).SetBytes(public[0:32])
	pub.Y = new(big.Int).SetBytes(public[32:])

	bi := new(big.Int).SetBytes(private)
	priv := new(ecdsa.PrivateKey)
	priv.PublicKey.Curve = pubkeyCurve
	priv.D = bi
	priv.PublicKey.X, priv.PublicKey.Y = pubkeyCurve.ScalarBaseMult(bi.Bytes())

	if priv.Curve.IsOnCurve(pub.X, pub.Y) {
		x, _ := pub.Curve.ScalarMult(pub.X, pub.Y, priv.D.Bytes())
		key := sha256.Sum256([]byte(hex.EncodeToString(x.Bytes())))
		shared = key[:]
	} else {
		err = fmt.Errorf("Not IsOnCurve")
	}
	return
}

// GetSharedHex generates a shared key from private and public key. All keys are hex string.
func GetSharedHex(private, public string) (string, error) {
	var (
		err               error
		priv, pub, shared []byte
	)
	if priv, err = hex.DecodeString(private); err == nil {
		if pub, err = hex.DecodeString(public); err == nil {
			if shared, err = GetSharedKey(priv, pub); err == nil {
				return hex.EncodeToString(shared), nil
			}
		}
	}
	return ``, err
}

// GetShared returns the combined key for the specified public key. If the text is encrypted
// with this key then it can be decrypted with the shared key made from private key and the returned public key (pub).
// All keys are hex strings.
func GetShared(public string) (string, string, error) {
	priv, pub, err := GenHexKeys()
	if err != nil {
		return ``, ``, err
	}
	shared, err := GetSharedHex(priv, public)
	return shared, pub, err
}

// PKCS7Padding realizes PKCS#7 encoding which is described in RFC 5652.
func PKCS7Padding(src []byte, blockSize int) []byte {
	padding := blockSize - len(src)%blockSize
	return append(src, bytes.Repeat([]byte{byte(padding)}, padding)...)
}

// PKCS7UnPadding realizes PKCS#7 decoding.
func PKCS7UnPadding(src []byte) ([]byte, error) {
	length := len(src)
	if length < int(src[length-1]) {
		return nil, fmt.Errorf(`incorrect input of PKCS7UnPadding`)
	}
	return src[:length-int(src[length-1])], nil
}

// CBCEncrypt encrypts the text by using the key parameter. It uses CBC mode of AES.
func CBCEncrypt(key, text, iv []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	plaintext := PKCS7Padding(text, aes.BlockSize)
	if iv == nil {
		iv = make([]byte, aes.BlockSize, aes.BlockSize+len(plaintext))
		if _, err := io.ReadFull(rand.Reader, iv); err != nil {
			return nil, err
		}
	}
	if len(iv) < aes.BlockSize {
		return nil, fmt.Errorf(`wrong size of iv %d`, len(iv))
	}
	mode := cipher.NewCBCEncrypter(block, iv[:aes.BlockSize])
	encrypted := make([]byte, len(plaintext))
	mode.CryptBlocks(encrypted, plaintext)
	return append(iv, encrypted...), nil
}

// CBCDecrypt decrypts the text by using key. It uses CBC mode of AES.
func CBCDecrypt(key, ciphertext, iv []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	if len(ciphertext) < aes.BlockSize || len(ciphertext)%aes.BlockSize != 0 {
		return nil, fmt.Errorf(`Wrong size of cipher %d`, len(ciphertext))
	}
	if iv == nil {
		iv = ciphertext[:aes.BlockSize]
		ciphertext = ciphertext[aes.BlockSize:]
	}
	ret := make([]byte, len(ciphertext))
	cipher.NewCBCDecrypter(block, iv[:aes.BlockSize]).CryptBlocks(ret, ciphertext)
	if ret, err = PKCS7UnPadding(ret); err != nil {
		return nil, err
	}
	return ret, nil
	/*	cipher.NewCBCDecrypter(block, iv[:aes.BlockSize]).CryptBlocks(ciphertext, ciphertext)
		if ciphertext, err = PKCS7UnPadding(ciphertext); err != nil {
			return nil, err
		}
		return ciphertext, nil*/
}

// SharedEncrypt creates a shared key and encrypts text. The first 32 characters are the created public key.
// The cipher text can be only decrypted with the original private key.
func SharedEncrypt(public, text []byte) ([]byte, error) {
	priv, pub, err := GenBytesKeys()
	if err != nil {
		return nil, err
	}
	shared, err := GetSharedKey(priv, public)
	if err != nil {
		return nil, err
	}
	return CBCEncrypt(shared, text, pub)
}

// SharedDecrypt decrypts the ciphertext by using private key.
func SharedDecrypt(private, ciphertext []byte) ([]byte, error) {
	if len(ciphertext) <= 64 {
		return nil, fmt.Errorf(`too short cipher %d`, len(ciphertext))
	}
	shared, err := GetSharedKey(private, ciphertext[:64])
	if err != nil {
		return nil, err
	}
	return CBCDecrypt(shared, ciphertext[64:], ciphertext[:aes.BlockSize])
}
