package client

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/parnurzeal/gorequest"
	"github.com/pkg/errors"

	"github.com/GenesisKernel/go-genesis/packages/crypto"
	"github.com/GenesisKernel/go-genesis/tools/update_client/params"
	upd_crypto "github.com/GenesisKernel/go-genesis/tools/update_server/crypto"
	"github.com/GenesisKernel/go-genesis/tools/update_server/model"
)

type UpdateClient struct {
}

// GenerateKeys creates public/private key pair
func (uc *UpdateClient) GenerateKeys(keyParams params.KeyParams) error {
	priv, pub, err := crypto.GenBytesKeys()
	if err != nil {
		return errors.Wrapf(err, "can't generate keys")
	}

	err = ioutil.WriteFile(keyParams.PrivateKeyPath, priv, 0600)
	if err != nil {
		return errors.Wrapf(err, "can't write private key")
	}

	err = ioutil.WriteFile(keyParams.PublicKeyPath, pub, 0600)
	if err != nil {
		return errors.Wrapf(err, "can't write public key")
	}
	return nil
}

// AddBinary is adding build to update server (require auth credentials)
func (uc *UpdateClient) AddBinary(keyParams params.KeyParams, binaryParams params.BinaryParams, serverParams params.ServerParams) error {
	priv, err := os.Open(keyParams.PrivateKeyPath)
	if err != nil {
		return errors.Wrapf(err, "can't open private key path %s", keyParams.PrivateKeyPath)
	}
	data, err := ioutil.ReadAll(priv)
	if err != nil {
		return errors.Wrapf(err, "can't read private key path %s ", keyParams.PrivateKeyPath)
	}

	file, err := os.Open(binaryParams.Path)
	if err != nil {
		return errors.Wrapf(err, "can't open binary path %s", binaryParams.Path)
	}

	binaryData, err := ioutil.ReadAll(file)
	if err != nil {
		return errors.Wrapf(err, "can't read binary path %s ", binaryParams.Path)
	}

	pv, err := parseVersion(binaryParams.Version)
	if err != nil {
		return errors.Wrapf(err, "version parsing")
	}

	b := model.Build{
		Body:       binaryData,
		Time:       time.Now().UTC(),
		Version:    pv,
		Name:       path.Base(binaryParams.Path),
		StartBlock: uint64(binaryParams.StartBlock),
	}
	s := upd_crypto.NewBuildSigner(data)
	sign, err := s.MakeSign(b)
	if err != nil {
		return errors.Wrapf(err, "can't create sign")
	}
	b.Sign = sign

	r, bd, errs := gorequest.
		New().
		SetBasicAuth(serverParams.Login, serverParams.Password).
		Post(fmt.Sprintf("%s/api/v1/private/binary", serverParams.Server)).
		Send(b).
		End()

	if errs != nil {
		return errors.Errorf("creating request error: %v", errs)
	}

	if r.StatusCode != http.StatusOK {
		return errors.Errorf("adding binary response code: %d, body: %s", r.StatusCode, bd)
	}

	return nil
}

// GetBinary is retrieving full build model from update server
func (uc *UpdateClient) GetBinary(serverParams params.ServerParams, keyParams params.KeyParams, binaryParams params.BinaryParams) (model.Build, error) {
	var b model.Build

	pv, err := parseVersion(binaryParams.Version)
	if err != nil {
		return b, errors.Wrapf(err, "version parsing")
	}

	r, bd, errs := gorequest.
		New().
		Get(fmt.Sprintf("%s/api/v1/%s/%s/%s", serverParams.Server, pv.OS, pv.Arch, pv.Number)).
		EndStruct(&b)

	if errs != nil {
		return b, errors.Errorf("creating 1 request error: %v", errs)
	}

	if r.StatusCode != http.StatusOK {
		return b, errors.Errorf("getting binary response code: %d, body: %s", r.StatusCode, bd)
	}

	r, bdy, errs := gorequest.
		New().
		Get(fmt.Sprintf("%s/api/v1/%s/%s/%s/binary", serverParams.Server, pv.OS, pv.Arch, pv.Number)).
		End()

	if errs != nil {
		return b, errors.Errorf("creating 2 request error: %v", errs)
	}

	if r.StatusCode != http.StatusOK {
		return b, errors.Errorf("getting binary response code: %d, body: %s", r.StatusCode, bdy)
	}

	b.Body = []byte(bdy)

	pub, err := os.Open(keyParams.PublicKeyPath)
	if err != nil {
		return b, nil
	}
	defer pub.Close()

	keyData, err := ioutil.ReadAll(pub)
	if err != nil {
		return b, err
	}

	sn := upd_crypto.BuildSigner{}
	verified, err := sn.CheckSign(b, keyData)
	if err != nil {
		return b, errors.Wrapf(err, "verifying binary")
	}

	if !verified {
		return b, errors.New("binary not verified")
	}

	return b, nil
}

func (uc *UpdateClient) GetLastBinary(serverParams params.ServerParams,
	keyParams params.KeyParams,
	binaryParams params.BinaryParams) (model.Build, error) {
	var b model.Build

	pv, err := parseVersion(binaryParams.Version)
	if err != nil {
		return b, errors.Wrapf(err, "version parsing")
	}

	r, bd, errs := gorequest.
		New().
		Get(fmt.Sprintf("%s/api/v1/%s/%s/last", serverParams.Server, pv.OS, pv.Arch)).
		EndStruct(&b)

	if errs != nil {
		return b, errors.Errorf("creating 1 request error: %v", errs)
	}

	if r.StatusCode != http.StatusOK {
		return b, errors.Errorf("getting last version response code: %d, body: %s", r.StatusCode, bd)
	}

	binaryParams.Version = b.Version.String()
	return uc.GetBinary(serverParams, keyParams, binaryParams)
}

// RemoveBinary is removing build from update server (require auth credentials)
func (uc *UpdateClient) RemoveBinary(serverParams params.ServerParams, binaryParams params.BinaryParams) error {
	pv, err := parseVersion(binaryParams.Version)
	if err != nil {
		return errors.Wrapf(err, "version parsing")
	}

	r, bd, errs := gorequest.
		New().
		SetBasicAuth(serverParams.Login, serverParams.Password).
		Delete(fmt.Sprintf("%s/api/v1/private/binary/%s/%s/%s", serverParams.Server, pv.OS, pv.Arch, pv.Number)).
		End()

	if errs != nil {
		return errors.Errorf("creating request error: %v", errs)
	}

	if r.StatusCode != http.StatusOK {
		return errors.Errorf("removing binary response code: %d, body: %s", r.StatusCode, bd)
	}

	return nil
}

// GetVersionList is retrieving list of versions by os+arch parameters
func (uc *UpdateClient) GetVersionList(serverParams params.ServerParams, binaryParams params.BinaryParams) ([]model.Build, error) {
	var resp []model.Build

	pv, err := parseVersion(binaryParams.Version)
	if err != nil {
		return resp, errors.Wrapf(err, "version parsing")
	}

	r, bd, errs := gorequest.
		New().
		Get(fmt.Sprintf("%s/api/v1/%s/%s/versions", serverParams.Server, pv.OS, pv.Arch)).
		EndStruct(&resp)

	if errs != nil {
		return resp, errors.Errorf("creating request error: %v", errs)
	}

	if r.StatusCode != http.StatusOK {
		return resp, errors.Errorf("retrieving response list code: %d, body: %s", r.StatusCode, bd)
	}

	return resp, nil
}

func parseVersion(version string) (model.Version, error) {
	pv, err := model.NewVersion(version)
	if err != nil {
		tv := model.Version{Number: "%NUM%", OS: "%OS%", Arch: "%ARCH%"}
		return model.Version{}, errors.Wrapf(err, "wrong version format, expected: %s", tv.String())
	}

	return pv, nil
}
