package crypto_test

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/GenesisCommunity/go-genesis/tools/update_server/crypto"
	"github.com/GenesisCommunity/go-genesis/tools/update_server/model"
)

func TestBinarySign(t *testing.T) {
	b := model.Build{Version: model.Version{Number: "1.1"}, Body: []byte("test"), Time: time.Now()}
	priv, err := os.Open("../testdata/key")
	if err != nil {
		t.Error("private key not found")
	}
	privData, err := ioutil.ReadAll(priv)
	if err != nil {
		t.Error("erro reading private key")
	}

	pub, err := os.Open("../testdata/key.pub")
	if err != nil {
		t.Error("public key not found")
	}
	pubData, err := ioutil.ReadAll(pub)
	if err != nil {
		t.Error("error reading public key")
	}

	bs := crypto.NewBuildSigner(privData)
	sign, err := bs.MakeSign(b)
	if err != nil {
		t.Error("can't sign")
	}
	b.Sign = sign

	verify, err := bs.CheckSign(b, pubData)
	if err != nil {
		t.Error("can't verify")
	}
	if !verify {
		t.Error("not verified")
	}
}
