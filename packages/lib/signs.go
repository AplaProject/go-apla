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
	"crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
)

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
			return nil, nil, err
		}
		left := parse(binSign[2:])
		if left == nil || int(binSign[3])+6 > len(binSign) {
			return nil, nil, fmt.Errorf(`wrong left parsing`)
		}
		right := parse(binSign[4+binSign[3]:])
		if right == nil {
			return nil, nil, fmt.Errorf(`wrong right parsing`)
		}
		sign = hex.EncodeToString(append(left, right...))
	} else if len(sign) < 128 {
		return nil, nil, fmt.Errorf(`wrong len of signature %d`, len(sign))
	}
	all, err := hex.DecodeString(sign[:])
	if err != nil {
		return nil, nil, err
	}
	return new(big.Int).SetBytes(all[:32]), new(big.Int).SetBytes(all[len(all)-32:]), nil
}

// CheckECDSA checks if forSign has been signed with corresponding to public the private key
func CheckECDSA(public []byte, forSign string, signature []byte) (bool, error) {
	if len(forSign) == 0 || len(public) != 64 || len(signature) == 0 {
		return false, fmt.Errorf("invalid parameters")
	}
	pubkeyCurve := elliptic.P256()
	signhash := sha256.Sum256([]byte(forSign))

	pubkey := new(ecdsa.PublicKey)
	pubkey.Curve = pubkeyCurve
	pubkey.X = new(big.Int).SetBytes(public[0:32])
	pubkey.Y = new(big.Int).SetBytes(public[32:])
	r, s, err := parseSign(hex.EncodeToString(signature))
	if err != nil {
		return false, err
	}
	verifystatus := ecdsa.Verify(pubkey, signhash[:], r, s)
	if !verifystatus {
		return false, fmt.Errorf("incorrect sign:  hash = %x; forSign = %v, publicKey = %x, sign = %x",
			signhash, forSign, public, signature)
	}
	return true, nil
}

// SignECDSA returns the signature of forSign made with privateKey.
func SignECDSA(privateKey string, forSign string) (ret []byte, err error) {
	pubkeyCurve := elliptic.P256()

	b, err := hex.DecodeString(privateKey)
	if err != nil {
		return
	}
	bi := new(big.Int).SetBytes(b)
	priv := new(ecdsa.PrivateKey)
	priv.PublicKey.Curve = pubkeyCurve
	priv.D = bi
	priv.PublicKey.X, priv.PublicKey.Y = pubkeyCurve.ScalarBaseMult(bi.Bytes())

	signhash := sha256.Sum256([]byte(forSign))
	r, s, err := ecdsa.Sign(crand.Reader, priv, signhash[:])
	if err != nil {
		return
	}
	ret = append(FillLeft(r.Bytes()), FillLeft(s.Bytes())...)
	return
}

// JSSignToBytes converts hex signature which has got from the browser to []byte
func JSSignToBytes(in string) ([]byte, error) {
	r, s, err := parseSign(in)
	if err != nil {
		return nil, err
	}
	return append(FillLeft(r.Bytes()), FillLeft(s.Bytes())...), nil
}
