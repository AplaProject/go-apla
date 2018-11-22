package api

import (
	"net/http"
	"strings"

	"github.com/GenesisKernel/go-genesis/packages/consts"
)

func getVersionHandler(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, strings.Join([]string{
		consts.VERSION, consts.BuildInfo}, " ",
	))
}
