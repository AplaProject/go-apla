package syspar

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	log "github.com/sirupsen/logrus"
)

const publicKeyLength = 64

var (
	errFullNodeInvalidValues = errors.New("Invalid values of the full_node parameter")
)

//because of PublicKey is byte
type fullNodeJSON struct {
	TCPAddress string      `json:"tcp_address"`
	APIAddress string      `json:"api_address"`
	KeyID      json.Number `json:"key_id"`
	PublicKey  string      `json:"public_key"`
	UnbanTime  json.Number `json:"unban_time,er"`
}

// FullNode is storing full node data
type FullNode struct {
	TCPAddress string
	APIAddress string
	KeyID      int64
	PublicKey  []byte
	UnbanTime  time.Time
}

// UnmarshalJSON is custom json unmarshaller
func (fn *FullNode) UnmarshalJSON(b []byte) (err error) {
	data := fullNodeJSON{}
	if err = json.Unmarshal(b, &data); err != nil {
		log.WithFields(log.Fields{"type": consts.JSONMarshallError, "error": err, "value": string(b)}).Error("Unmarshalling full nodes to json")
		return err
	}

	fn.TCPAddress = data.TCPAddress
	fn.APIAddress = data.APIAddress
	fn.KeyID = converter.StrToInt64(data.KeyID.String())

	if fn.PublicKey, err = hex.DecodeString(data.PublicKey); err != nil {
		log.WithFields(log.Fields{"type": consts.ConversionError, "error": err, "value": data.PublicKey}).Error("converting full nodes public key from hex")
		return err
	}
	fn.UnbanTime = time.Unix(converter.StrToInt64(data.UnbanTime.String()), 0)

	if err = fn.Validate(); err != nil {
		return err
	}

	return nil
}

func (fn *FullNode) MarshalJSON() ([]byte, error) {
	jfn := fullNodeJSON{
		TCPAddress: fn.TCPAddress,
		APIAddress: fn.APIAddress,
		KeyID:      json.Number(strconv.FormatInt(fn.KeyID, 10)),
		PublicKey:  hex.EncodeToString(fn.PublicKey),
		UnbanTime:  json.Number(strconv.FormatInt(fn.UnbanTime.Unix(), 10)),
	}

	data, err := json.Marshal(jfn)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.JSONUnmarshallError, "error": err}).Error("Marshalling full nodes to json")
		return nil, err
	}

	return data, nil
}

// ValidateURL returns error if the URL is invalid
func validateURL(rawurl string) error {
	u, err := url.ParseRequestURI(rawurl)
	if err != nil {
		return err
	}

	if len(u.Scheme) == 0 {
		return fmt.Errorf("Invalid scheme: %s", rawurl)
	}

	if len(u.Host) == 0 {
		return fmt.Errorf("Invalid host: %s", rawurl)
	}

	return nil
}

// Validate checks values
func (fn *FullNode) Validate() error {
	if fn.KeyID == 0 || len(fn.PublicKey) != publicKeyLength || len(fn.TCPAddress) == 0 {
		return errFullNodeInvalidValues
	}

	if err := validateURL(fn.APIAddress); err != nil {
		return err
	}

	return nil
}
