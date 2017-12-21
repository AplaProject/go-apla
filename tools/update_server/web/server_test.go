package web_test

import (
	"net/http/httptest"

	"testing"

	"fmt"

	"os"

	"encoding/json"
	"net/http"

	"io/ioutil"

	"github.com/AplaProject/go-apla/tools/update_server/config"
	"github.com/AplaProject/go-apla/tools/update_server/crypto"
	"github.com/AplaProject/go-apla/tools/update_server/model"
	"github.com/AplaProject/go-apla/tools/update_server/storage"
	"github.com/AplaProject/go-apla/tools/update_server/web"
	"github.com/parnurzeal/gorequest"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	server     *httptest.Server
	sm         storage.MockEngine
	cm         crypto.MockSigner
	pubkey     []byte
	v1ApiRoute string
)

func init() {
	sm = storage.MockEngine{}
	cm = crypto.MockSigner{}
	pubkey = []byte("test")

	s := web.Server{
		Db:        &sm,
		Conf:      &config.Config{Login: "test", Pass: "test"},
		Signer:    &cm,
		PublicKey: pubkey,
	}

	server = httptest.NewServer(s.GetRoutes())
	v1ApiRoute = fmt.Sprintf("%s/api/v1", server.URL)
}

func TestGetLastVersion(t *testing.T) {
	cases := []struct {
		getVersionsList []model.Version
		getVersionsErr  error
		lastVersion     model.Build

		get    model.Build
		getErr error

		respBody []byte
		expCode  int
	}{
		{
			getVersionsList: []model.Version{
				{Number: "0.1", OS: "windows", Arch: "amd64"},
				{Number: "0.1.1", OS: "windows", Arch: "amd64"},
			},
			lastVersion: model.Build{Version: model.Version{Number: "0.1.1", OS: "windows", Arch: "amd64"}},
			get:         model.Build{Body: []byte{1, 2, 3, 4, 5}},
			respBody:    []byte{1, 2, 3, 4, 5},
			expCode:     http.StatusOK,
		},
	}

	for _, c := range cases {
		reloadMocks(t)
		sm.On("GetVersionsList").Return(c.getVersionsList, c.getVersionsErr)
		sm.On("Get", c.lastVersion).Return(c.get, c.getErr)

		r, b, errs := gorequest.New().Get(fmt.Sprintf("%s/%s/%s/last", v1ApiRoute, c.lastVersion.OS, c.lastVersion.Arch)).End()
		dumpErrors(t, errs)

		assert.Equal(t, c.expCode, r.StatusCode)
		assert.Equal(t, c.respBody, []byte(b))
	}
}

func TestGetVersion(t *testing.T) {
	cases := []struct {
		getVersionsList []model.Version
		getVersionsErr  error

		os   string
		arch string

		respBody string
		expCode  int
	}{
		{
			getVersionsList: []model.Version{
				{Number: "0.1", OS: "windows", Arch: "amd64"},
				{Number: "0.1.1", OS: "windows", Arch: "amd64"},
				{Number: "2.0.1", OS: "linux", Arch: "amd64"},
			},
			os:   "windows",
			arch: "amd64",
			respBody: toJson(t, []model.Version{
				{Number: "0.1", OS: "windows", Arch: "amd64"},
				{Number: "0.1.1", OS: "windows", Arch: "amd64"},
			}),
			expCode: http.StatusOK,
		},
	}

	for _, c := range cases {
		reloadMocks(t)
		sm.On("GetVersionsList").Return(c.getVersionsList, c.getVersionsErr)

		r, _, errs := gorequest.New().Get(fmt.Sprintf("%s/%s/%s/versions", v1ApiRoute, c.os, c.arch)).End()
		dumpErrors(t, errs)

		rb, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, c.expCode, r.StatusCode)
		assert.Equal(t, c.respBody, string(rb))
	}
}

func TestGetBinary(t *testing.T) {
	cases := []struct {
		get    model.Build
		getErr error

		os      string
		arch    string
		version string

		respBody []byte
		expCode  int
	}{
		{
			getErr:  errors.New("blah"),
			os:      "windows",
			arch:    "amd64",
			version: "1.0",
			expCode: http.StatusInternalServerError,
		},
		{
			get:      model.Build{Body: []byte{9, 5, 2, 7, 8, 0}},
			os:       "windows",
			arch:     "amd64",
			version:  "1.0",
			respBody: []byte{9, 5, 2, 7, 8, 0},
			expCode:  http.StatusOK,
		},
	}

	for _, c := range cases {
		reloadMocks(t)
		sm.On("Get", model.Build{Version: model.Version{Number: c.version, OS: c.os, Arch: c.arch}}).Return(c.get, c.getErr)

		r, _, errs := gorequest.New().Get(fmt.Sprintf("%s/%s/%s/%s", v1ApiRoute, c.os, c.arch, c.version)).End()
		dumpErrors(t, errs)

		rb, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, c.expCode, r.StatusCode)
		if c.expCode == http.StatusOK {
			assert.Equal(t, c.respBody, rb)
		}
	}
}

func TestAddBinaryAuthorize(t *testing.T) {
	r, _, errs := gorequest.
		New().
		Post(fmt.Sprintf("%s/private/binary", v1ApiRoute)).
		End()
	dumpErrors(t, errs)
	assert.Equal(t, http.StatusUnauthorized, r.StatusCode)

	r, _, errs = gorequest.
		New().
		Post(fmt.Sprintf("%s/private/binary", v1ApiRoute)).
		SetBasicAuth("wrong", "wrong").
		End()
	dumpErrors(t, errs)
	assert.Equal(t, http.StatusUnauthorized, r.StatusCode)

	r, _, errs = gorequest.
		New().
		Post(fmt.Sprintf("%s/private/binary", v1ApiRoute)).
		SetBasicAuth("test", "test").
		End()
	dumpErrors(t, errs)
	assert.Equal(t, http.StatusBadRequest, r.StatusCode)
}

func TestAddBinary(t *testing.T) {
	cases := []struct {
		binary model.Build

		checkSign    bool
		checkSignErr error

		add     error
		expCode int
	}{
		{
			binary:    model.Build{},
			checkSign: false,
			expCode:   http.StatusBadRequest,
		},
		{
			checkSign: true,
			expCode:   http.StatusBadRequest,
		},
		{
			checkSignErr: errors.New("blah"),
			expCode:      http.StatusBadRequest,
		},
		{
			checkSign: true,
			binary:    model.Build{Version: model.Version{Number: "blah-error", OS: "windows", Arch: "amd64"}},
			expCode:   http.StatusBadRequest,
		},
		{
			checkSign: true,
			binary:    model.Build{Version: model.Version{Number: "1.0", OS: "linux", Arch: "blah-not-exist"}},
			expCode:   http.StatusBadRequest,
		},
		{
			checkSign: true,
			binary:    model.Build{Version: model.Version{Number: "1.0", OS: "linux", Arch: "i386"}},
			add:       errors.New("blah"),
			expCode:   http.StatusInternalServerError,
		},
		{
			checkSign: true,
			binary:    model.Build{Version: model.Version{Number: "1.0", OS: "linux", Arch: "i386"}},
			expCode:   http.StatusOK,
		},
	}

	for _, c := range cases {
		reloadMocks(t)
		cm.On("CheckSign", c.binary, pubkey).Return(c.checkSign, c.checkSignErr)
		sm.On("Add", c.binary).Return(c.add)

		r, _, errs := gorequest.
			New().
			Post(fmt.Sprintf("%s/private/binary", v1ApiRoute)).
			SetBasicAuth("test", "test").
			Send(c.binary).
			End()
		dumpErrors(t, errs)

		assert.Equal(t, c.expCode, r.StatusCode)
	}
}

func TestRemoveBinary(t *testing.T) {
	cases := []struct {
		build     model.Build
		deleteErr error

		os      string
		arch    string
		version string

		expCode int
	}{
		{
			build:     model.Build{Version: model.Version{Number: "1.0", OS: "windows", Arch: "amd64"}},
			deleteErr: errors.New("blah"),
			os:        "windows",
			arch:      "amd64",
			version:   "1.0",
			expCode:   http.StatusInternalServerError,
		},
		{
			build:   model.Build{Version: model.Version{Number: "1.0", OS: "windows", Arch: "amd64"}},
			os:      "windows",
			arch:    "amd64",
			version: "1.0",
			expCode: http.StatusOK,
		},
	}

	for _, c := range cases {
		reloadMocks(t)
		sm.On("Delete", c.build).Return(c.deleteErr)

		r, _, errs := gorequest.
			New().
			Delete(fmt.Sprintf("%s/private/binary/%s/%s/%s", v1ApiRoute, c.os, c.arch, c.version)).
			SetBasicAuth("test", "test").
			End()
		dumpErrors(t, errs)

		assert.Equal(t, c.expCode, r.StatusCode)
	}
}

func reloadMocks(t *testing.T) {
	sm = storage.MockEngine{}
	cm = crypto.MockSigner{}
}

func toJson(t *testing.T, d interface{}) string {
	jsonString, err := json.Marshal(d)
	require.NoError(t, err)
	return string(jsonString)
}

func dumpErrors(t *testing.T, errs []error) {
	if errs != nil {
		for _, value := range errs {
			fmt.Println(value.Error())
		}
		os.Exit(1)
	}
}
