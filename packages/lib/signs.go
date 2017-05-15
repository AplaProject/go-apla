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
	"crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"math/big"
)

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
