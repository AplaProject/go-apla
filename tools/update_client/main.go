package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/EGaaS/update_client/structs"
	version "github.com/hashicorp/go-version"
)

/*
var (
	command    *string
	binaryPath *string
	updateAddr *string
	login      *string
	pass       *string
	privKey    *string
	pubKey     *string
	version    *string
	genKeys    *bool
)
*/
func main() {
	command := flag.String("command", "", "")
	binaryPath := flag.String("binary", "", "")
	updateAddr := flag.String("updateAddr", "", "")
	login := flag.String("login", "", "")
	pass := flag.String("pass", "", "")
	privKey := flag.String("priv", "update.priv", "")
	pubKey := flag.String("pub", "update.pub", "")
	version := flag.String("version", "", "")
	startBlock := flag.Int64("startBlock", -1231, "")
	flag.Parse()

	var err error
	switch *command {
	case "a":
		if strings.Trim(*pubKey, " ") == "" ||
			strings.Trim(*privKey, " ") == "" ||
			strings.Trim(*binaryPath, " ") == "" ||
			strings.Trim(*version, " ") == "" ||
			strings.Trim(*login, " ") == "" ||
			strings.Trim(*pass, " ") == "" ||
			strings.Trim(*updateAddr, " ") == "" ||
			*startBlock == -1231 {
			fmt.Println(`Usage of a command: -command="a" 
-pubKey="file_path" [default:"update.pub"] 
-privKey="file_path" [default:"update.priv"] 
-binary="file_path" 
-version="X.X.X" 
-login="login" 
-pass="pass"
-updateAddr="http://XXX.xx"
-startBlock=0 or block number`)
			return
		}
		err = addBinary(*pubKey, *privKey, *binaryPath, *version, *login, *pass, *updateAddr, *startBlock)
	case "g":
		if strings.Trim(*pubKey, " ") == "" ||
			strings.Trim(*updateAddr, " ") == "" ||
			strings.Trim(*version, " ") == "" {
			fmt.Println(`Usage of a command: -command="g" -pubKey="file_path" [default:"update.pub"], -updateAddr="http://XXX.xx", -version="x.x.x"`)
			return
		}
		err = getBinary(*updateAddr, *pubKey, *version)
	case "r":
		if strings.Trim(*updateAddr, " ") == "" ||
			strings.Trim(*login, " ") == "" ||
			strings.Trim(*pass, " ") == "" ||
			strings.Trim(*version, " ") == "" {
			fmt.Println(`Usage of a command: -command="r" -updateAddr="http://XXX.xx" -login="login" -pass="pass" -version="x.x.x"`)
			return
		}
		err = removeBinary(*version, *login, *pass, *updateAddr)
	case "gv":
		if strings.Trim(*updateAddr, " ") == "" {
			fmt.Println(`Usage of a command: -command="gv" -updateAddr="http://XXX.xx"`)
			return
		}
		err = getVersionList(*updateAddr)
	case "gen":
		if strings.Trim(*pubKey, " ") == "" ||
			strings.Trim(*privKey, " ") == "" {
			fmt.Println(`Usage of a command: -command="g" -pubKey="file_path" [default:"update.pub"], -privKey="file_path"`)
			return
		}
		err = generateKeys(*privKey, *pubKey)
	default:
		fmt.Println(`available commands: "a" for adding binary to update server, 
"g" for getting binary from server, 
"r" for removing binary from update server, 
"gv" for getting available versions list, 
"gen" for keys generating`)
	}
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("ok")
	}
}

func generateKeys(privatePath string, publicPath string) error {
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

func addBinary(publicPath string, privatePath string, binaryPath string, vers string,
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

func getBinary(updateAddr string, publicPath string, version string) error {
	resp, err := http.Get(updateAddr + "/v1/binary/" + version)
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

func removeBinary(version string, login string, pass string, updateAddr string) error {
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

func getLastVersion() error {
	return nil
}

func getVersionList(updateAddr string) error {
	resp, err := http.Get(updateAddr + "/v1/version")
	if err != nil {
		return err
	}
	data, _ := ioutil.ReadAll(resp.Body)
	versions := strings.Split(string(data), "|")
	defer resp.Body.Close()
	fmt.Println(versions)
	return nil
}
