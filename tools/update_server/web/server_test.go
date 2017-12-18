package web_test

import (
	"net/http/httptest"

	"testing"

	"fmt"

	"os"

	"encoding/json"
	"net/http"

	"github.com/AplaProject/go-apla/tools/update_server/storage"
	"github.com/AplaProject/go-apla/tools/update_server/web"
	"github.com/parnurzeal/gorequest"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

var (
	server     *httptest.Server
	sm         storage.MockEngine
	v1ApiRoute string
)

func init() {
	sm := storage.MockEngine{}

	s := web.Server{
		Db: &sm,
	}

	server = httptest.NewServer(s.GetRoutes())
	server.Start()

	v1ApiRoute = fmt.Sprintf("%s/api/v1", server.URL)
}

func TestGetLastVersion(t *testing.T) {
	cases := []struct {
		getVersionsList []string
		getVersionsErr  error

		getBinary    []byte
		getBinaryErr error

		respBody string
		expCode  int
	}{
		{respBody: "", expCode: http.StatusOK},
		{getVersionsErr: errors.New("blah"), expCode: http.StatusInternalServerError},
		{getVersionsList: []string{"0.1", "0.1.1"}, respBody: toJson(t, "0.1.1"), expCode: http.StatusOK},
		{getVersionsList: []string{"1.1", "2.0", "3.0"}, respBody: toJson(t, "3.0"), expCode: http.StatusOK},
		{getVersionsList: []string{"1.9", "2.0"}, respBody: toJson(t, "2.0"), expCode: http.StatusOK},
	}
	sm.On("GetVersionsList").Return([]string{"1.0.0", "0.1", "0.1.1", "2.0", "2.1"}, nil)

	r, b, errs := gorequest.New().Get(fmt.Sprintf("%s/last", v1ApiRoute)).End()
	dumpErrors(t, errs)

}

func reloadMocks(t *testing.T) {
	sm = storage.MockEngine{}
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
