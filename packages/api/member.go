package api

import (
	"bytes"
	"net/http"
	"strconv"

	"github.com/GenesisKernel/go-genesis/packages/consts"

	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/model"
	log "github.com/sirupsen/logrus"
)

func getAvatar(w http.ResponseWriter, r *http.Request, data *apiData, logger *log.Entry) error {
	parMember := data.params["member"].(string)
	memberID := converter.StrToInt64(parMember)
	member := &model.Member{}
	member.SetTablePrefix(converter.Int64ToStr(data.ecosystemId))
	if err := member.Get(memberID); err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).
			Errorf("getting member with ecosystem: %d member_id: %d", data.ecosystemId, memberID)
		return err
	}

	log.Info("avatar:", member.Avatar)
	buf := bytes.NewBufferString(member.Avatar)
	w.Header().Set("Content-Type", http.DetectContentType(buf.Bytes()))
	w.Header().Set("Content-Length", strconv.Itoa(len(buf.Bytes())))
	if _, err := w.Write(buf.Bytes()); err != nil {
		log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("unable to write image")
		return err
	}

	return nil
}
