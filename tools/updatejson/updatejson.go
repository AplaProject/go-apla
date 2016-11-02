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
	"encoding/hex"
	"encoding/json"
	"fmt"
	//"io"
	"io/ioutil"
	//	"net/http"
	"os"
	//	"os/exec"
	"path/filepath"
	"reflect"
	//	"runtime"
	//  "strings"
	"time"

	"github.com/EGaaS/go-egaas-mvp/packages/lib"
)

var (
	options Settings
)

type Settings struct {
	Version string
	Domain  string
	InPath  string
	OutPath string
	File    string
	ZipFile string
}

type Update struct {
	Version string
	Hash    string
	Sign    string
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

	out := make(map[string]Update)

	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		exit(err)
	}
	privateKey, err := ioutil.ReadFile(filepath.Join(dir, `PrivateKey`))
	if err != nil {
		exit(err)
	}
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
		var upd Update
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
		md5, err := lib.CalculateMd5(filepath.Join(set.InPath, set.File))
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

		out[key] = upd
	}
	//	fmt.Println(`Set`, out)
	if updjson, err := json.Marshal(out); err != nil {
		exit(err)
	} else if err = ioutil.WriteFile(filepath.Join(dir, `update.json`), updjson, 0644); err != nil {
		exit(err)
	}
	/*

		if err = os.Chdir(srcPath); err != nil {
			exit(err)
		}

		zipfile := `daylight.zip`
		switch runtime.GOOS {
		case `windows`:
			if runtime.GOARCH == `386` {
				zipfile = `daylight_win32.zip`
			} else {
				zipfile = `daylight_win64.zip`
			}
		}
		zipname := filepath.Join(filepath.Dir(filepath.Dir(options.OutFile)), zipfile)
		fmt.Println(`Compressing`, zipname)

		zipf, err := os.Create(zipname)
		if err != nil {
			exit(err)
		}
		z := zip.NewWriter(zipf)
		var out []byte
		if out, err = ioutil.ReadFile(options.OutFile); err != nil {
			exit(err)
		}
		header, _ := BytesInfoHeader(len(out), filepath.Base(options.OutFile))
		f, _ := z.CreateHeader(header)
		f.Write(out)
		z.Close()
		zipf.Close()

	*/
	exit(nil)
}
