package api

import (
	"net/http"
	"strings"

	"github.com/AplaProject/go-apla/packages/consts"
)

func getVersionHandler(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, strings.Join([]string{
		consts.VERSION, consts.BuildInfo}, " ",
	))
}
