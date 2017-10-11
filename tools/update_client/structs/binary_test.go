package structs

import (
	"io/ioutil"
	"os"
	"testing"
	"time"
)

func TestBinarySign(t *testing.T) {
	b := &Binary{Version: "1.1", Body: []byte("test"), Date: time.Now()}
	priv, err := os.Open("../update.priv")
	if err != nil {
		t.Error("private key not found")
	}
	privData, err := ioutil.ReadAll(priv)
	if err != nil {
		t.Error("erro reading private key")
	}

	pub, err := os.Open("../update.pub")
	if err != nil {
		t.Error("public key not found")
	}
	pubData, err := ioutil.ReadAll(pub)
	if err != nil {
		t.Error("error reading public key")
	}

	err = b.MakeSign(privData)
	if err != nil {
		t.Error("can't sign")
	}

	verify, err := b.CheckSign(pubData)
	if err != nil {
		t.Error("can't verify")
	}
	if !verify {
		t.Error("not verified")
	}
}
