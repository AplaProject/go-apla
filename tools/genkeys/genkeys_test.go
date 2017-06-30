package main

import (
	"bytes"
	"encoding/hex"
	"io/ioutil"
	"os"
	"os/exec"
	"testing"
)

func TestGenKeys(t *testing.T) {
	os.Remove(`private.key`)
	os.Remove(`public.key`)
	cmd := exec.Command("genkeys", "-seed=seed.txt")
	err := cmd.Run()
	if err != nil {
		t.Error(err)
	}
	if privKey, err := ioutil.ReadFile(`private.key`); err != nil {
		t.Error(err.Error())
	} else {
		want, _ := hex.DecodeString(`b29c04d999140afab5dfdaa437533927af3981651425d6d77b49ecb9c8d7e60b`)
		if bytes.Compare(privKey, want) != 0 {
			t.Error(`different private key`)
		}
	}
	if pubKey, err := ioutil.ReadFile(`public.key`); err != nil {
		t.Error(err.Error())
	} else {
		want, _ := hex.DecodeString(`0a3234f3e25c57358316ae47dcce7cd59fa1d5501d85523d6706d712b1daf602` +
			`b37f9e3ebd7309d9a2c0669848898c71410679b62e3f9a8fee6a29a3013005cf`)
		if bytes.Compare(pubKey, want) != 0 {
			t.Error(`different public key`)
		}
	}
}
