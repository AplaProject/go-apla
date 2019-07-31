// Copyright (C) 2017, 2018, 2019 EGAAS S.A.
//
// This program is free software; you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation; either version 2 of the License, or (at
// your option) any later version.
//
// This program is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
// General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program; if not, write to the Free Software
// Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301, USA.

package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"math/big"
	"strconv"
	"strings"

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
)

// CutPub removes the first 04 byte
func CutPub(pubKey []byte) []byte {
	if len(pubKey) == 65 && pubKey[0] == 4 {
		pubKey = pubKey[1:]
	}
	return pubKey
}

// Address gets int64 EGGAS address from the public key
func Address(pubKey []byte) int64 {
	pubKey = CutPub(pubKey)
	h256 := sha256.Sum256(pubKey)
	h512 := sha512.Sum512(h256[:])
	crc := calcCRC64(h512[:])
	// replace the last digit by checksum
	num := strconv.FormatUint(crc, 10)
	val := []byte(strings.Repeat("0", consts.AddressLength-len(num)) + num)
	return int64(crc - (crc % 10) + uint64(checkSum(val[:len(val)-1])))
}

// PrivateToPublic returns the public key for the specified private key.
func PrivateToPublic(key []byte) ([]byte, error) {
	var pubkeyCurve elliptic.Curve
	switch ellipticSize {
	case elliptic256:
		pubkeyCurve = elliptic.P256()
	default:
		return nil, ErrUnsupportedCurveSize
	}

	bi := new(big.Int).SetBytes(key)
	priv := new(ecdsa.PrivateKey)
	priv.PublicKey.Curve = pubkeyCurve
	priv.D = bi
	priv.PublicKey.X, priv.PublicKey.Y = pubkeyCurve.ScalarBaseMult(key)
	return append(converter.FillLeft(priv.PublicKey.X.Bytes()), converter.FillLeft(priv.PublicKey.Y.Bytes())...), nil
}

// KeyToAddress converts a public key to apla address XXXX-...-XXXX.
func KeyToAddress(pubKey []byte) string {
	return converter.AddressToString(Address(pubKey))
}

// GetWalletIDByPublicKey converts public key to wallet id
func GetWalletIDByPublicKey(publicKey []byte) (int64, error) {
	key, _ := HexToPub(string(publicKey))
	return int64(Address(key)), nil
}

// HexToPub encodes hex string to []byte of pub key
func HexToPub(pub string) ([]byte, error) {
	key, err := hex.DecodeString(pub)
	if err != nil {
		return nil, err
	}
	return CutPub(key), nil
}

// PubToHex decodes []byte of pub key to hex string
func PubToHex(pub []byte) string {
	if len(pub) == 64 {
		pub = append([]byte{4}, pub...)
	}
	return hex.EncodeToString(pub)
}
