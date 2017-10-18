package main

import (
	"flag"
	"fmt"
	"strings"

	"github.com/AplaProject/go-apla/tools/update_client/client"
)

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
	remove := flag.String("remove", "", "")
	flag.Parse()

	client := &client.UpdateClient{}

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
		err = client.AddBinary(*pubKey, *privKey, *binaryPath, *version, *login, *pass, *updateAddr, *startBlock)
	case "g":
		if strings.Trim(*pubKey, " ") == "" ||
			strings.Trim(*updateAddr, " ") == "" ||
			strings.Trim(*version, " ") == "" {
			fmt.Println(`Usage of a command: -command="g" -pubKey="file_path" [default:"update.pub"], -updateAddr="http://XXX.xx", -version="x.x.x"`)
			return
		}
		err = client.GetBinary(*updateAddr, *pubKey, *version)
	case "r":
		if strings.Trim(*updateAddr, " ") == "" ||
			strings.Trim(*login, " ") == "" ||
			strings.Trim(*pass, " ") == "" ||
			strings.Trim(*version, " ") == "" {
			fmt.Println(`Usage of a command: -command="r" -updateAddr="http://XXX.xx" -login="login" -pass="pass" -version="x.x.x"`)
			return
		}
		err = client.RemoveBinary(*version, *login, *pass, *updateAddr)
	case "gv":
		if strings.Trim(*updateAddr, " ") == "" {
			fmt.Println(`Usage of a command: -command="gv" -updateAddr="http://XXX.xx"`)
			return
		}
		var versions []string
		versions, err = client.GetVersionList(*updateAddr)
		fmt.Println(versions)
	case "gen":
		if strings.Trim(*pubKey, " ") == "" ||
			strings.Trim(*privKey, " ") == "" {
			fmt.Println(`Usage of a command: -command="g" -pubKey="file_path" [default:"update.pub"] -privKey="file_path"`)
			return
		}
		err = client.GenerateKeys(*privKey, *pubKey)
	case "u":
		if strings.Trim(*version, " ") == "" ||
			strings.Trim(*remove, " ") == "" {
			fmt.Println(`Usage of a command: -command="u" -version="x.x.x" -remove="old_binary_path -public="file_path" [default:"update.pub"]`)
			return
		}
		err = client.UpdateFile(*version, *remove, *pubKey)
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
