package syspar

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"

	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/consts"
	log "github.com/sirupsen/logrus"
)

const publicKeyLength = 64

var (
	errFullNodeInvalidValues = errors.New("Invalid values of the full_node parameter")
)

//because of PublicKey is byte
type fullNodeJSON struct {
	TCPAddress string `json:"tcp_address"`
	APIAddress string `json:"api_address"`
	KeyID      string `json:"key_id"`
	PublicKey  string `json:"public_key"`
}

// FullNode is storing full node data
type FullNode struct {
	TCPAddress string
	APIAddress string
	KeyID      int64
	PublicKey  []byte
}

func (fn *FullNode) UnmarshalJSON(b []byte) (err error) {
	data := fullNodeJSON{}
	if err = json.Unmarshal(b, &data); err != nil {
		log.WithFields(log.Fields{"type": consts.JSONMarshallError, "error": err, "value": string(b)}).Error("Unmarshalling full nodes to json")
		return err
	}

	fn.TCPAddress = data.TCPAddress
	fn.APIAddress = data.APIAddress
	fn.KeyID = converter.StrToInt64(data.KeyID)

	if fn.PublicKey, err = hex.DecodeString(data.PublicKey); err != nil {
		log.WithFields(log.Fields{"type": consts.ConversionError, "error": err, "value": data.PublicKey}).Error("converting full nodes public key from hex")
		return err
	}

	if err = fn.Validate(); err != nil {
		return err
	}

	return nil
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

func (fn *FullNode) Validate() error {
	if fn.KeyID == 0 || len(fn.PublicKey) != publicKeyLength || len(fn.TCPAddress) == 0 {
		return errFullNodeInvalidValues
	}

	if err := validateURL(fn.APIAddress); err != nil {
		return err
	}

	return nil
}
