package storage_test

import (
	ioutil "io/ioutil"
	"math/rand"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/GenesisCommunity/go-genesis/tools/update_server/model"
	"github.com/GenesisCommunity/go-genesis/tools/update_server/storage"
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

//TODO make this whole tests for all storage implementations, not just for boltDb
func TestBoltStorage_AddBinary(t *testing.T) {
	cases := []struct {
		version  string
		os       string
		arch     string
		binary   []byte
		expError bool
	}{
		{expError: true},
		{binary: []byte{1}, expError: true},
		{version: "0.0.1", expError: true},
		{version: "0.0.2", binary: []byte{1}, expError: true},
		{version: "0.0.3", os: "linux", arch: "amd64", binary: []byte{1}, expError: false},
	}

	bs, spath := newTempBoltStorage(t)
	var wr []model.Version // added versions
	for _, c := range cases {
		bu := model.Build{Version: model.Version{Number: c.version, OS: c.os, Arch: c.arch}, Body: c.binary}
		err := bs.Add(bu)
		if c.expError {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			wr = append(wr, bu.Version)
		}
	}

	rw, err := bs.GetVersionsList()
	require.NoError(t, err)
	assert.Equal(t, wr, rw)

	cleanUpFile(t, spath)

}

func TestBoltStorage_GetVersionsList(t *testing.T) {
	cases := []struct {
		versionsList []model.Version
	}{
		{},
		{versionsList: []model.Version{
			{Number: "1.0", OS: "darwin", Arch: "amd64"},
		}},
		{versionsList: []model.Version{
			{Number: "0.2", OS: "darwin", Arch: "amd64"},
			{Number: "1.0", OS: "darwin", Arch: "amd64"},
			{Number: "2.3.5", OS: "linux", Arch: "amd64"},
		}},
	}

	for _, c := range cases {
		bs, spath := newTempBoltStorage(t)
		for _, v := range c.versionsList {
			bu := model.Build{Version: v, Body: []byte{1}}
			require.NoError(t, bs.Add(bu))
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
		bu := model.Build{Version: model.Version{Number: c.version, OS: "linux", Arch: "amd64"}, Body: c.binary}
		require.NoError(t, bs.Add(bu))
	}

	for _, c := range cases {
		cb := model.Build{Version: model.Version{Number: c.version, OS: "linux", Arch: "amd64"}}
		b, err := bs.Get(cb)
		if c.expError {
			require.Error(t, err)
		} else {
			assert.NoError(t, err)
		}

		assert.Equal(t, c.binary, b.Body)
	}
	cleanUpFile(t, spath)
}

func TestBoltStorage_DeleteBinary(t *testing.T) {
	cases := []struct {
		id       int
		version  string
		os       string
		arch     string
		binary   []byte
		expError bool
	}{
		{id: 1, version: "0.0.1", os: "darwin", arch: "amd64", binary: []byte{1}},
		{id: 2, version: "0.1.0", os: "linux", arch: "amd64", binary: []byte{2}},
		{id: 3, version: "0.1.1", os: "linux", arch: "amd64", binary: []byte{3}},
	}

	// after removing one binary we need to check that all other still present in storage
	checkOtherExists := func(st storage.BoltStorage, versions map[int]bool) {
		for v, e := range versions {
			if e {
				var f bool
				for _, c := range cases {
					if c.id == v {
						cv := model.Build{Version: model.Version{Number: c.version, Arch: c.arch, OS: c.os}}
						b, err := st.Get(cv)
						require.NoError(t, err)
						assert.Equal(t, c.binary, b.Body)
						f = true
					}
				}
				assert.True(t, f)
			}
		}
	}

	bs, spath := newTempBoltStorage(t)
	wr := make(map[int]bool) // status of added versions
	for _, c := range cases {
		bu := model.Build{Version: model.Version{Number: c.version, OS: c.os, Arch: c.arch}, Body: c.binary}
		require.NoError(t, bs.Add(bu))
		wr[c.id] = true
	}

	for _, c := range cases {
		cb := model.Build{Version: model.Version{Number: c.version, OS: c.os, Arch: c.arch}}
		require.NoError(t, bs.Delete(cb))
		wr[c.id] = false
		checkOtherExists(bs, wr)
	}

	cleanUpFile(t, spath)
}
