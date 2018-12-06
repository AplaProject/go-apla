package api

import (
	"net/http"
	"strconv"

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/model"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

func getAvatarHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	logger := getLogger(r)

	memberID := converter.StrToInt64(params["member"])
	ecosystemID := converter.StrToInt64(params["ecosystem"])

	member := &model.Member{}
	member.SetTablePrefix(converter.Int64ToStr(ecosystemID))

	found, err := member.Get(memberID)
	if err != nil {
		logger.WithFields(log.Fields{
			"type":      consts.DBError,
			"error":     err,
			"ecosystem": ecosystemID,
			"member_id": memberID,
		}).Error("getting member")
		errorResponse(w, err)
		return
	}

	if !found {
		errorResponse(w, errNotFound)
		return
	}

	if member.ImageID == nil {
		errorResponse(w, errNotFound)
		return
	}

	bin := &model.Binary{}
	bin.SetTablePrefix(converter.Int64ToStr(ecosystemID))
	found, err = bin.GetByID(*member.ImageID)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err, "image_id": *member.ImageID}).Errorf("on getting binary by id")
		errorResponse(w, err)
		return
	}

	if !found {
		errorResponse(w, errNotFound)
		return
	}

	if len(bin.Data) == 0 {
		logger.WithFields(log.Fields{"type": consts.EmptyObject, "error": err, "image_id": *member.ImageID}).Errorf("on check avatar size")
		errorResponse(w, errNotFound)
		return
	}

	w.Header().Set("Content-Type", bin.MimeType)
	w.Header().Set("Content-Length", strconv.Itoa(len(bin.Data)))
	if _, err := w.Write(bin.Data); err != nil {
		logger.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("unable to write image")
	}
}
