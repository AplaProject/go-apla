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

package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
)

func main() {
	pubkeyCurve := elliptic.P256()

	privhex := "f59d89d161b7ef26d499d5d896a06c03da6cea2bedbb078610fc4551c937ee9c"
	//	pubhex := "1a3a4db29ce0b6b1393b025e8f0e37733f79ff6c33bed36cf84e1b21ff40d3860f1456bfa1c43e14e8c4b692ee619de7bf599db5063663ae028d0e9c7c680baa"

	//	priv2 := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	pub2 := "131e2d69f16e0594173bad74b8bd2fefa6cd9d8a240fb9be58128549874ec3a99b42a6c192859ea7bee6214e8ff548479d3f047b500c571e2bb705f858bff80b"

	public, _ := hex.DecodeString(pub2)
	pub := new(ecdsa.PublicKey)
	pub.Curve = pubkeyCurve
	pub.X = new(big.Int).SetBytes(public[0:32])
	pub.Y = new(big.Int).SetBytes(public[32:])

	b, _ := hex.DecodeString(privhex)
	bi := new(big.Int).SetBytes(b)
	priv := new(ecdsa.PrivateKey)
	priv.PublicKey.Curve = pubkeyCurve
	priv.D = bi
	priv.PublicKey.X, priv.PublicKey.Y = pubkeyCurve.ScalarBaseMult(bi.Bytes())

	if priv.Curve != pub.Curve {
		fmt.Println(priv.Curve, " != ", pub.Curve)
	} else if !priv.Curve.IsOnCurve(pub.X, pub.Y) {
		fmt.Println("Not IsOnCurve")
	}
	x, _ := pub.Curve.ScalarMult(pub.X, pub.Y, priv.D.Bytes())
	shared := sha256.Sum256(x.Bytes()[:2])
	shared2 := sha256.Sum256([]byte(hex.EncodeToString(x.Bytes())))
	shared3 := sha256.Sum256([]byte(`message`))
	fmt.Println(`Bytes`, x.Bytes())
	fmt.Println("x=", hex.EncodeToString(x.Bytes()), hex.EncodeToString(shared[:]))
	fmt.Println(hex.EncodeToString(shared2[:]), hex.EncodeToString(shared3[:]))
}
