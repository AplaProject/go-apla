package storage_test

import (
	ioutil "io/ioutil"
	"os"
	"testing"

	"math/rand"

	"github.com/AplaProject/go-apla/tools/update_server/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTempBoltStorage(t *testing.T) (storage.BoltStorage, string) {
	tmp, err := ioutil.TempFile(os.TempDir(), "bolt.db")
	require.NoError(t, err)

	b, err := storage.NewBoltStorage(tmp.Name())
	require.NoError(t, err)

	return b, tmp.Name()
}

func cleanUpFile(t *testing.T, storagePath string) {
	require.NoError(t, os.Remove(storagePath))
}

func TestBoltStorage_AddBinary(t *testing.T) {
	cases := []struct {
		version  string
		binary   []byte
		expError bool
	}{
		{expError: true},
		{binary: []byte{1}, expError: true},
		{version: "0.0.1", expError: true},
		{version: "0.0.1", binary: []byte{1}, expError: false},
	}

	bs, spath := newTempBoltStorage(t)
	var wr []string // added versions
	for _, c := range cases {
		err := bs.AddBinary(c.binary, c.version)
		if c.expError {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			wr = append(wr, c.version)
		}
	}

	rw, err := bs.GetVersionsList()
	require.NoError(t, err)
	assert.Equal(t, wr, rw)

	cleanUpFile(t, spath)

}

func TestBoltStorage_GetVersionsList(t *testing.T) {
	cases := []struct {
		versionsList []string
	}{
		{versionsList: []string{"0.0.1"}},
		{versionsList: []string{"0.9", "0.9.5", "0.9.9", "1", "1.0.5", "2.0"}},
	}

	for _, c := range cases {
		bs, spath := newTempBoltStorage(t)
		for _, v := range c.versionsList {
			require.NoError(t, bs.AddBinary([]byte{1}, v))
		}

		rs, err := bs.GetVersionsList()
		require.NoError(t, err)
		assert.Equal(t, c.versionsList, rs)

		cleanUpFile(t, spath)
	}
}

func TestBoltStorage_GetBinary(t *testing.T) {
	// generating 0 - 4 mb byte slice
	genRandBytes := func(seed int64) []byte {
		b := make([]byte, rand.Intn(2<<21))
		rand.Seed(seed)
		_, err := rand.Read(b)
		require.NoError(t, err)
		return b
	}

	cases := []struct {
		version  string
		binary   []byte
		expError bool
	}{
		{version: "0.0.1", binary: genRandBytes(1)},
		{version: "0.1.0", binary: genRandBytes(2)},
		{version: "0.1.1", binary: genRandBytes(3)},
		{version: "1.0.0", binary: genRandBytes(4)},
		{version: "1.0.1", binary: genRandBytes(5)},
		{version: "1.1.0", binary: genRandBytes(6)},
		{version: "1.1.1", binary: genRandBytes(7)},
	}

	bs, spath := newTempBoltStorage(t)
	for _, c := range cases {
		require.NoError(t, bs.AddBinary(c.binary, c.version))
	}

	for _, c := range cases {
		b, err := bs.GetBinary(c.version)
		if c.expError {
			require.Error(t, err)
		} else {
			assert.NoError(t, err)
		}

		assert.Equal(t, c.binary, b)
	}
	cleanUpFile(t, spath)
}

func TestBoltStorage_DeleteBinary(t *testing.T) {
	cases := []struct {
		version  string
		binary   []byte
		expError bool
	}{
		{version: "0.0.1", binary: []byte{1}},
		{version: "0.1.0", binary: []byte{2}},
		{version: "0.1.1", binary: []byte{3}},
	}

	// after removing one binary we need to check that all other still present in storage
	checkOtherExists := func(st storage.BoltStorage, versions map[string]bool) {
		for v, e := range versions {
			if e {
				b, err := st.GetBinary(v)
				require.NoError(t, err)

				for _, c := range cases {
					if c.version == v {
						assert.Equal(t, c.binary, b)
					}
				}
			}
		}
	}

	bs, spath := newTempBoltStorage(t)
	wr := make(map[string]bool) // status of added versions
	for _, c := range cases {
		require.NoError(t, bs.AddBinary(c.binary, c.version))
		wr[c.version] = true
	}

	for _, c := range cases {
		require.NoError(t, bs.DeleteBinary(c.version))
		wr[c.version] = false
		checkOtherExists(bs, wr)
	}

	cleanUpFile(t, spath)
}
