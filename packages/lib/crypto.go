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

/*
// GetSharedKey creates and returns the shared key = private * public.
// public must be the public key from the different private key.
func GetSharedKey(private, public []byte) (shared []byte, err error) {
	pubkeyCurve := elliptic.P256()

	private = converter.FillLeft(private)
	public = converter.FillLeft(public)
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
	priv, pub, err := crypto.GenHexKeys()
	if err != nil {
		return ``, ``, err
	}
	shared, err := GetSharedHex(priv, public)
	return shared, pub, err
}
*/
