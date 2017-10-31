// Copyright 2016 The go-daylight Authors
// This file is part of the go-daylight library.
//
// The go-daylight library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-daylight library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-daylight library. If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"archive/zip"
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"io/ioutil"
	//	"net/http"
	"os"
	//	"os/exec"
	"path/filepath"
	"reflect"
	//	"runtime"
	"strings"
	"time"

	"github.com/AplaProject/go-apla/packages/lib"
	"github.com/AplaProject/go-apla/packages/utils"
)

const (
	// Key is an encrypted private key
	Key = `17a984259355caeaad8837f3fc0af12527b801c349677effdcbcaceb1398c54463d64d6cc7081bc94f0479013bf24521`
	// CRC is a crc of the password
	CRC = 3362054855
)

var (
	options Settings
)

// Settings is a structure for saving in update.json file
type Settings struct {
	Version  string
	Domain   string
	InPath   string
	OutPath  string
	File     string
	ZipFile  string
	JSONPath string
}

func exit(err error) {
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(`Press Enter to exit...`)
	fmt.Scanln()
	if err != nil {
		os.Exit(1)
	}
}

// BytesInfoHeader returns file header for zip file
func BytesInfoHeader(size int, filename string) (*zip.FileHeader, error) {
	fh := &zip.FileHeader{
		Name:               filename,
		UncompressedSize64: uint64(size),
		UncompressedSize:   uint32(size),
		Method:             zip.Deflate,
	}
	fh.SetModTime(time.Now())
	//   fh.SetMode(fi.Mode())
	return fh, nil
}

func main() {
	var (
		settings map[string]Settings
	)

	out := make(map[string]utils.Update)

	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		exit(err)
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter password: ")
	psw, _ := reader.ReadString('\n')
	if crc32.Checksum([]byte(strings.TrimSpace(psw)), crc32.MakeTable(0xD5828281)) != CRC {
		exit(fmt.Errorf(`Wrong password`))
	}

	key, _ := hex.DecodeString(Key)
	pass := md5.Sum([]byte(psw))
	privKey, err := Decrypt(pass[:], key)
	if err != nil {
		exit(err)
	}
	privateKey := hex.EncodeToString(privKey)
	if len(privateKey) == 0 {
		exit(fmt.Errorf(`PrivateKey is unknown`))
	}
	params, err := ioutil.ReadFile(filepath.Join(dir, `updatejson.json`))
	if err != nil {
		exit(err)
	}
	if err = json.Unmarshal(params, &settings); err != nil {
		exit(err)
	}
	options = settings[`default`]
	//	fmt.Println(options)

	for key, opt := range settings {
		var upd lib.Update
		if key == `default` {
			continue
		}
		set := options
		r := reflect.ValueOf(opt)
		for i := 0; i < r.NumField(); i++ {
			val := r.Field(i).String()
			if len(val) > 0 {
				ro := reflect.ValueOf(&set)
				ro.Elem().Field(i).SetString(val)
			}
		}
		infile := filepath.Join(set.InPath, set.File)
		md5, err := lib.CalculateMd5(infile)
		if err != nil {
			exit(err)
		}
		upd.Version = set.Version
		upd.Hash = hex.EncodeToString(md5)
		sign, err := lib.SignECDSA(string(privateKey), upd.Hash)
		if err != nil {
			exit(err)
		}
		upd.Sign = hex.EncodeToString(sign)
		if err = os.MkdirAll(set.OutPath, 0755); err != nil {
			exit(err)
		}
		zipname := filepath.Join(set.OutPath, set.ZipFile)
		fmt.Println(`Compressing`, zipname)
		zipf, err := os.Create(zipname)
		if err != nil {
			exit(err)
		}
		z := zip.NewWriter(zipf)
		var outzip []byte
		if outzip, err = ioutil.ReadFile(infile); err != nil {
			exit(err)
		}
		header, _ := BytesInfoHeader(len(outzip), filepath.Base(set.ZipFile))
		f, _ := z.CreateHeader(header)
		f.Write(outzip)
		z.Close()
		zipf.Close()
		upd.URL = set.Domain + set.ZipFile
		out[key] = upd
	}

	if updjson, err := json.Marshal(out); err != nil {
		exit(err)
	} else if err = ioutil.WriteFile(filepath.Join(options.JSONPath, `update.json`), updjson, 0644); err != nil {
		exit(err)
	}

	exit(nil)
}
