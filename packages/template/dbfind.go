package template

import (
	"encoding/json"
	"fmt"

	"github.com/AplaProject/go-apla/packages/consts"
	log "github.com/sirupsen/logrus"
)

const (
	columnTypeText     = "text"
	columnTypeLongText = "long_text"
	columnTypeBlob     = "blob"

	substringLength = 32
)

func dbfindExpressionBlob(column string) string {
	return fmt.Sprintf(`md5(%s) "%[1]s"`, column)
}

func dbfindExpressionLongText(column string) string {
	return fmt.Sprintf(`json_build_array(
		substr(%s, 1, %d),
		CASE WHEN length(%[1]s)>%[2]d THEN md5(%[1]s) END) "%[1]s"`, column, substringLength)
}

type valueLink struct {
	title string

	id     string
	table  string
	column string
	hash   string
}

func (vl *valueLink) link() string {
	if len(vl.hash) > 0 {
		return fmt.Sprintf("/data/%s/%s/%s/%s", vl.table, vl.id, vl.column, vl.hash)
	}
	return ""
}

func (vl *valueLink) marshal() (string, error) {
	b, err := json.Marshal(map[string]string{
		"title": vl.title,
		"link":  vl.link(),
	})
	if err != nil {
		log.WithFields(log.Fields{"type": consts.JSONMarshallError, "error": err}).Error("marshalling valueLink to JSON")
		return "", err
	}
	return string(b), nil
}
