package config_test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/GenesisKernel/go-genesis/tools/update_server/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDo(t *testing.T) {
	cases := []struct {
		filename           string
		configContent      string
		expectConfig       config.Config
		expectParsingError bool
	}{
		{expectParsingError: true},
		{filename: "config.ini", expectParsingError: true},
		{filename: "config.ini", configContent: "bad-config", expectParsingError: true},
		{
			filename:           "config.ini",
			configContent:      "[without=default]\nlogin=admin\npass=admin",
			expectParsingError: true,
		},
		{
			filename:      "config.ini",
			configContent: "[default]\nlogin=admin\npass=admin\nhost=localhost\nport=12345\nstorage=./resources/storage\npubkeypath=pubkey",
			expectConfig:  config.Config{Login: "admin", Pass: "admin", Host: "localhost", Port: "12345", StorageDir: "./resources/storage", PubkeyPath: "pubkey"},
		},
	}
	var rfs []string
	for _, c := range cases {
		fpath := ""
		if c.filename != "" {
			tmp, err := ioutil.TempFile(os.TempDir(), c.filename)
			fpath = tmp.Name()
			require.NoError(t, err)
			rfs = append(rfs, fpath)

			tmp.Write([]byte(c.configContent))
			tmp.Close()
		}

		p := config.NewParser(fpath)
		conf, err := p.Do()

		assert.Equal(t, conf, c.expectConfig)
		if c.expectParsingError {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
		}
	}

	for _, f := range rfs {
		require.NoError(t, os.Remove(f))
	}
}
