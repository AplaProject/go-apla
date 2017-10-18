package client

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"
	"time"

	"github.com/AplaProject/go-apla/packages/utils"
	"github.com/AplaProject/go-apla/tools/update_client/structs"
	version "github.com/hashicorp/go-version"
)

type UpdateClient struct {
}

func (uc *UpdateClient) GenerateKeys(privatePath string, publicPath string) error {
	curve := elliptic.P256()
	priv, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		return errors.New("can't generate private key")
	}
	pub := priv.PublicKey
	privKey, err := os.OpenFile(privatePath, os.O_CREATE, 0600)
	if err != nil {
		return errors.New("can't open private key file")
	}
	_, err = privKey.Write(priv.D.Bytes())
	if err != nil {
		return errors.New("can't write private key")
	}

	pubKey, err := os.OpenFile(publicPath, os.O_CREATE, 0600)
	if err != nil {
		return errors.New("can't open public key file")
	}
	var key []byte
	key = append(key, pub.X.Bytes()...)
	key = append(key, pub.Y.Bytes()...)
	_, err = pubKey.Write(key)
	if err != nil {
		return errors.New("can't write public key")
	}
	return nil
}

func (uc *UpdateClient) AddBinary(publicPath string, privatePath string, binaryPath string, vers string,
	login string, pass string, updateAddr string, startBlock int64) error {
	priv, err := os.Open(privatePath)
	if err != nil {
		return errors.New("can't open private key: " + privatePath)
	}
	data, err := ioutil.ReadAll(priv)
	if err != nil {
		return errors.New("can't read private key: " + privatePath)
	}

	file, err := os.Open(binaryPath)
	if err != nil {
		fmt.Println(binaryPath)
		return errors.New("can't open binary: " + binaryPath)
	}

	binaryData, err := ioutil.ReadAll(file)
	if err != nil {
		return errors.New("can't read binary: " + binaryPath)
	}

	v, err := version.NewVersion(vers)
	if err != nil {
		return errors.New("incorrect version number: " + vers)
	}
	binary := structs.Binary{Body: binaryData, Date: time.Now().UTC(), Version: v.String(),
		Name: path.Base(binaryPath), StartBlock: startBlock}
	err = binary.MakeSign(data)
	if err != nil {
		return errors.New("can't create sign")
	}

	request := structs.Request{Login: login, Pass: pass, B: binary}
	jsonData, err := json.Marshal(request)
	if err != nil {
		return errors.New("can't marshal to json: " + err.Error())
	}
	var buf bytes.Buffer
	buf.Write(jsonData)
	resp, err := http.Post(updateAddr+"/v1/binary", "application/json", &buf)
	if err != nil {
		return errors.New("can't send request: " + err.Error())
	}
	if resp.StatusCode != http.StatusOK {
		return errors.New("request error: " + resp.Status)
	}
	return nil
}

func (uc *UpdateClient) GetBinary(updateAddr string, publicPath string, version string) error {
	resp, err := http.Get(updateAddr + "/v1/binary/" + version + "/" + runtime.GOOS + "/" + runtime.GOARCH)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return errors.New("request error: " + resp.Status)
	}
	binary := &structs.Binary{}
	err = json.Unmarshal(data, binary)
	if err != nil {
		return errors.New("unmarshaling error: " + err.Error())
	}

	pub, err := os.Open(publicPath)
	if err != nil {
		return nil
	}
	defer pub.Close()

	keyData, err := ioutil.ReadAll(pub)
	if err != nil {
		return err
	}

	verified, err := binary.CheckSign(keyData)
	if err != nil {
		return err
	}

	if !verified {
		return errors.New("binary not verified")
	}

	err = ioutil.WriteFile("update_"+binary.Version, data, 0600)
	if err != nil {
		return err
	}
	return nil
}

func (uc *UpdateClient) RemoveBinary(version string, login string, pass string, updateAddr string) error {
	binary := structs.Binary{Version: version}
	request := structs.Request{Login: login, Pass: pass, B: binary}
	jsonData, err := json.Marshal(request)
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	buf.Write(jsonData)
	client := &http.Client{}
	req, _ := http.NewRequest("DELETE", updateAddr+"/v1/binary/"+version, &buf)
	req.Header.Add("Content-type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	fmt.Println(resp.Status)
	return nil
}

func (uc *UpdateClient) GetVersionList(updateAddr string) ([]string, error) {
	resp, err := http.Get(updateAddr + "/v1/version")
	if err != nil {
		return nil, err
	}
	data, _ := ioutil.ReadAll(resp.Body)
	versions := strings.Split(string(data), "|")
	defer resp.Body.Close()
	return versions[:len(versions)-1], nil
}

func (uc *UpdateClient) UpdateFile(newVersion string, oldFilePath string, publicKeyPath string) error {
	newFile, err := os.Open("update_" + newVersion)
	if err != nil {
		return err
	}
	newData, err := ioutil.ReadAll(newFile)
	if err != nil {
		return err
	}

	err = os.Remove(oldFilePath)
	if err != nil {
		return err
	}

	var binary structs.Binary
	err = json.Unmarshal(newData, &binary)
	if err != nil {
		return err
	}

	pubkey, err := os.Open(publicKeyPath)
	if err != nil {
		return err
	}

	pubData, err := ioutil.ReadAll(pubkey)
	if err != nil {
		return err
	}

	verified, err := binary.CheckSign(pubData)
	if err != nil {
		return err
	}

	if !verified {
		return errors.New("binary not verified")
	}

	err = ioutil.WriteFile(binary.Name, binary.Body, 0600)
	if err != nil {
		return err
	}
	exec.Command(*utils.Dir+"/"+binary.Name, "")
	return nil
}
