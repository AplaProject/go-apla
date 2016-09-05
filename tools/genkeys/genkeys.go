// genkeys
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"github.com/DayLightProject/go-daylight/packages/lib"
	"io/ioutil"
)

func main() {
	var (
		private, public []byte
		privKey, pubKey string
		err             error
	)
	ext := `.txt`
	hexf := flag.Bool("hex", false, "Keys are stored as hex-text files.")
	seed := flag.String("seed", ``, "Initial seed text file.")

	flag.Parse()

	if len(*seed) > 0 {
		if seedText, err := ioutil.ReadFile(seed); err != nil {
			privkey = err.Error()
		} else if len(seedText) == 0 {
			privkey = `Seed file is empty`
		} else {
			bi := new(big.Int).SetBytes(sha256.Sum256(seedText))
			priv := new(ecdsa.PrivateKey)
			priv.PublicKey.Curve = elliptic.P256()
			priv.D = bi
			priv.PublicKey.X, priv.PublicKey.Y = pubkeyCurve.ScalarBaseMult(bi.Bytes())
			privKey = hex.EncodeToString(priv.D.Bytes())
			pubKey = hex.EncodeToString(append(lib.FillLeft(priv.PublicKey.X.Bytes()),
				lib.FillLeft(priv.PublicKey.Y.Bytes())...))
		}
	} else {
		privKey, pubKey = lib.GenKeys()
	}
	if !*hexf {
		ext = `.key`
		if private, err = hex.DecodeString(privKey); err != nil {
			private = []byte(err.Error())
		}
		if public, err = hex.DecodeString(pubKey); err != nil {
			public = []byte(err.Error())
		}
	} else {
		private = []byte(privKey)
		public = []byte(pubKey)
	}
	ioutil.WriteFile(`private`+ext, private, 0644)
	ioutil.WriteFile(`public`+ext, public, 0644)
}
