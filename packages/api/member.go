package api

import (
	"bytes"
	"encoding/base64"
	"net/http"
	"strconv"

	"github.com/GenesisKernel/go-genesis/packages/consts"

	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/model"
	log "github.com/sirupsen/logrus"
)

func getAvatar(w http.ResponseWriter, r *http.Request, data *apiData, logger *log.Entry) error {
	parMember := data.params["member"].(string)
	parEcosystem := data.params["ecosystem"].(string)

	memberID := converter.StrToInt64(parMember)
	ecosystemID := converter.StrToInt64(parEcosystem)

	member := &model.Member{}
	member.SetTablePrefix(converter.Int64ToStr(ecosystemID))

	found, err := member.Get(memberID)
	if err != nil {
		log.WithFields(log.Fields{
			"type":      consts.DBError,
			"error":     err,
			"ecosystem": ecosystemID,
			"member_id": memberID,
		}).Error("getting member")
		return errorAPI(w, "E_SERVER", http.StatusInternalServerError)
	}

	if !found {
		return errorAPI(w, "E_SERVER", http.StatusNotFound)
	}

	if member.ImageID == nil {
		return errorAPI(w, "E_SERVER", http.StatusNotFound)
	}

	bin := &model.Binary{}
	bin.SetTablePrefix(converter.Int64ToStr(ecosystemID))
	found, err = bin.GetByID(*member.ImageID)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err, "image_id": *member.ImageID}).Errorf("on getting binary by id")
		return errorAPI(w, "E_SERVER", http.StatusInternalServerError)
	}

	if !found {
		return errorAPI(w, "E_SERVER", http.StatusNotFound)
	}

	if len(bin.Data) == 0 {
		log.WithFields(log.Fields{"type": consts.EmptyObject, "error": err, "image_id": *member.ImageID}).Errorf("on check avatar size")
		return errorAPI(w, "E_SERVER", http.StatusNotFound)
	}

	// cut the prefix like a 'data:blah-blah;base64,'
	b64data := bin.Data[bytes.IndexByte(bin.Data, ',')+1:]
	buf, err := base64.StdEncoding.DecodeString(string(b64data))
	if err != nil {
		log.WithFields(log.Fields{"type": consts.ConversionError, "error": err}).Error("on decoding avatar")
		return errorAPI(w, "E_SERVER", http.StatusInternalServerError)
	}

	mime := http.DetectContentType(buf)
	w.Header().Set("Content-Type", mime)
	w.Header().Set("Content-Length", strconv.Itoa(len(buf)))
	if _, err := w.Write(buf); err != nil {
		log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("unable to write image")
		return err
	}

	return nil
}
