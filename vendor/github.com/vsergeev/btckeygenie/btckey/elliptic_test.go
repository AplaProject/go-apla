/* btckeygenie v1.0.0
 * https://github.com/vsergeev/btckeygenie
 * License: MIT
 */

package btckey

import (
	"math/big"
	"testing"
)

var curve EllipticCurve

func init() {
	/* See SEC2 pg.9 http://www.secg.org/collateral/sec2_final.pdf */
	/* secp256k1 elliptic curve parameters */
	curve.P, _ = new(big.Int).SetString("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEFFFFFC2F", 16)
	curve.A, _ = new(big.Int).SetString("0000000000000000000000000000000000000000000000000000000000000000", 16)
	curve.B, _ = new(big.Int).SetString("0000000000000000000000000000000000000000000000000000000000000007", 16)
	curve.G.X, _ = new(big.Int).SetString("79BE667EF9DCBBAC55A06295CE870B07029BFCDB2DCE28D959F2815B16F81798", 16)
	curve.G.Y, _ = new(big.Int).SetString("483ADA7726A3C4655DA4FBFC0E1108A8FD17B448A68554199C47D08FFB10D4B8", 16)
	curve.N, _ = new(big.Int).SetString("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEBAAEDCE6AF48A03BBFD25E8CD0364141", 16)
	curve.H, _ = new(big.Int).SetString("01", 16)
}

func hex2int(hexstring string) (v *big.Int) {
	v, _ = new(big.Int).SetString(hexstring, 16)
	return v
}

func TestOnCurve(t *testing.T) {
	if !curve.IsOnCurve(curve.G) {
		t.Fatal("failure G on curve")
	}

	t.Log("G on curve")
}

func TestInfinity(t *testing.T) {
	O := Point{nil, nil}

	/* O not on curve */
	if curve.IsOnCurve(O) {
		t.Fatal("failure O on curve")
	}

	/* O is infinity */
	if !curve.IsInfinity(O) {
		t.Fatal("failure O not infinity on curve")
	}

	t.Log("O is not on curve and is infinity")
}

func TestPointAdd(t *testing.T) {
	X := "50863ad64a87ae8a2fe83c1af1a8403cb53f53e486d8511dad8a04887e5b2352"
	Y := "2cd470243453a299fa9e77237716103abc11a1df38855ed6f2ee187e9c582ba6"

	P := Point{hex2int(X), hex2int(Y)}
	O := Point{nil, nil}

	/* R = O + O = O */
	{
		R := curve.Add(O, O)
		if !curve.IsInfinity(R) {
			t.Fatal("failure O + O = O")
		}
		t.Log("success O + O = O")
	}

	/* R = P + O = P */
	{
		R := curve.Add(P, O)
		if R.X.Cmp(P.X) != 0 || R.Y.Cmp(P.Y) != 0 {
			t.Fatal("failure P + O = P")
		}
		t.Log("success P + O = P")
	}

	/* R = O + Q = Q */
	{
		R := curve.Add(O, P)
		if R.X.Cmp(P.X) != 0 || R.Y.Cmp(P.Y) != 0 {
			t.Fatal("failure O + Q = Q")
		}
		t.Log("success O + Q = Q")
	}

	/* R = (x,y) + (x,-y) = O */
	{
		Q := Point{P.X, subMod(big.NewInt(0), P.Y, curve.P)}

		R := curve.Add(P, Q)
		if !curve.IsInfinity(R) {
			t.Fatal("failure (x,y) + (x,-y) = O")
		}
		t.Log("success (x,y) + (x,-y) = O")
	}

	/* R = P + P */
	{
		PP := Point{hex2int("5dbcd5dfea550eb4fd3b5333f533f086bb5267c776e2a1a9d8e84c16a6743d82"), hex2int("8dde3986b6cbe395da64b6e95fb81f8af73f6e0cf1100555005bb4ba2a6a4a07")}

		R := curve.Add(P, P)
		if R.X.Cmp(PP.X) != 0 || R.Y.Cmp(PP.Y) != 0 {
			t.Fatal("failure P + P")
		}
		t.Log("success P + P")
	}

	Q := Point{hex2int("a83b8de893467d3a88d959c0eb4032d9ce3bf80f175d4d9e75892a3ebb8ab7e5"), hex2int("370f723328c24b7a97fe34063ba68f253fb08f8645d7c8b9a4ff98e3c29e7f0d")}
	PQ := Point{hex2int("fe7d540002e4355eb0ec36c217b4735495de7bd8634055ded3683b0e9da70ef1"), hex2int("fc033c1d74cb34e087a3495e505c0fc0e9e3e8297994878d89d882254ce8a9ef")}

	/* R = P + Q */
	{
		R := curve.Add(P, Q)
		if R.X.Cmp(PQ.X) != 0 || R.Y.Cmp(PQ.Y) != 0 {
			t.Fatal("failure P + Q")
		}
		t.Log("success P + Q")
	}

	/* R = Q + P */
	{
		R := curve.Add(Q, P)
		if R.X.Cmp(PQ.X) != 0 || R.Y.Cmp(PQ.Y) != 0 {
			t.Fatal("failure Q + P")
		}
		t.Log("success Q + P")
	}
}

func TestPointScalarMult(t *testing.T) {
	X := "50863ad64a87ae8a2fe83c1af1a8403cb53f53e486d8511dad8a04887e5b2352"
	Y := "2cd470243453a299fa9e77237716103abc11a1df38855ed6f2ee187e9c582ba6"
	P := Point{hex2int(X), hex2int(Y)}

	/* Q = k*P */
	{
		T := Point{hex2int("87d592bfdd24adb52147fea343db93e10d0585bc66d91e365c359973c0dc7067"), hex2int("a374e206cb7c8cd1074bdf9bf6ddea135f983aaa6475c9ab3bb4c38a0046541b")}
		Q := curve.ScalarMult(hex2int("14eb373700c3836404acd0820d9fa8dfa098d26177ca6e18b1c7f70c6af8fc18"), P)
		if Q.X.Cmp(T.X) != 0 || Q.Y.Cmp(T.Y) != 0 {
			t.Fatal("failure k*P")
		}
		t.Log("success k*P")
	}

	/* Q = n*G = O */
	{
		Q := curve.ScalarMult(curve.N, curve.G)
		if !curve.IsInfinity(Q) {
			t.Fatal("failure n*G = O")
		}
		t.Log("success n*G = O")
	}
}

func TestPointScalarBaseMult(t *testing.T) {
	/* Sample Private Key */
	D := "18e14a7b6a307f426a94f8114701e7c8e774e7f9a47e2c2035db29a206321725"
	/* Sample Corresponding Public Key */
	X := "50863ad64a87ae8a2fe83c1af1a8403cb53f53e486d8511dad8a04887e5b2352"
	Y := "2cd470243453a299fa9e77237716103abc11a1df38855ed6f2ee187e9c582ba6"

	P := Point{hex2int(X), hex2int(Y)}

	/* Q = d*G = P */
	Q := curve.ScalarBaseMult(hex2int(D))
	if P.X.Cmp(Q.X) != 0 || P.Y.Cmp(Q.Y) != 0 {
		t.Fatal("failure Q = d*G")
	}
	t.Log("success Q = d*G")

	/* Q on curve */
	if !curve.IsOnCurve(Q) {
		t.Fatal("failure Q on curve")
	}
	t.Log("success Q on curve")

	/* R = 0*G = O */
	R := curve.ScalarBaseMult(big.NewInt(0))
	if !curve.IsInfinity(R) {
		t.Fatal("failure 0*G = O")
	}
	t.Log("success 0*G = O")
}

func TestPointDecompress(t *testing.T) {
	/* Valid points */
	var validDecompressVectors = []Point{
		{hex2int("50863ad64a87ae8a2fe83c1af1a8403cb53f53e486d8511dad8a04887e5b2352"), hex2int("2cd470243453a299fa9e77237716103abc11a1df38855ed6f2ee187e9c582ba6")},
		{hex2int("a83b8de893467d3a88d959c0eb4032d9ce3bf80f175d4d9e75892a3ebb8ab7e5"), hex2int("370f723328c24b7a97fe34063ba68f253fb08f8645d7c8b9a4ff98e3c29e7f0d")},
		{hex2int("f680556678e25084a82fa39e1b1dfd0944f7e69fddaa4e03ce934bd6b291dca0"), hex2int("52c10b721d34447e173721fb0151c68de1106badb089fb661523b8302a9097f5")},
		{hex2int("241febb8e23cbd77d664a18f66ad6240aaec6ecdc813b088d5b901b2e285131f"), hex2int("513378d9ff94f8d3d6c420bd13981df8cd50fd0fbd0cb5afabb3e66f2750026d")},
	}

	for i := 0; i < len(validDecompressVectors); i++ {
		P, err := curve.Decompress(validDecompressVectors[i].X, validDecompressVectors[i].Y.Bit(0))
		if err != nil {
			t.Fatalf("failure decompress P, got error %v on index %d", err, i)
		}
		if P.X.Cmp(validDecompressVectors[i].X) != 0 || P.Y.Cmp(validDecompressVectors[i].Y) != 0 {
			t.Fatalf("failure decompress P, got mismatch on index %d", i)
		}
	}
	t.Log("success Decompress() on valid vectors")

	/* Invalid points */
	var invalidDecompressVectors = []struct {
		X    *big.Int
		YLsb uint
	}{
		{hex2int("c8e337cee51ae9af3c0ef923705a0cb1b76f7e8463b3d3060a1c8d795f9630fd"), 0},
		{hex2int("c8e337cee51ae9af3c0ef923705a0cb1b76f7e8463b3d3060a1c8d795f9630fd"), 1},
	}

	for i := 0; i < len(invalidDecompressVectors); i++ {
		_, err := curve.Decompress(invalidDecompressVectors[i].X, invalidDecompressVectors[i].YLsb)
		if err == nil {
			t.Fatalf("failure decompress invalid P, got decompressed point on index %d", i)
		}
	}
	t.Log("success Decompress() on invalid vectors")
}
