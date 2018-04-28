package api

import (
	"net/http"

	"github.com/GenesisKernel/go-genesis/packages/consts"
)

func versionHandler(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, consts.VERSION)
}
