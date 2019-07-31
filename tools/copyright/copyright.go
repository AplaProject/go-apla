package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// The program inserts copyright notice at the beginning of .go files.

const (
	root = `https://github.com/AplaProject/go-apla/blob/master`
)

var (
	copyright []byte
)

// ProcessDir proceeds the specified directory and inserts the copyright at the beginning of the files
func ProcessDir(dir string, recurse bool) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		fmt.Println(err)
	}
	path, err := filepath.Abs(dir)
	if err != nil {
		fmt.Println(err)
		return
	}
	path = path[strings.Index(path, `go-apla`)+7:]
	for _, file := range files {
		fname := file.Name()
		fullName := filepath.Join(dir, fname)
		if strings.HasSuffix(fname, `.go`) {
			fmt.Printf(fullName)
			if fdata, err := ioutil.ReadFile(fullName); err != nil {
				fmt.Println(err)
				os.Exit(1)
			} else {
				var prefix string
				if bytes.Equal(fdata[:4], []byte(`// +`)) {
					lines := strings.Split(string(fdata[:256]), "\n")
					for _, line := range lines {
						if strings.HasPrefix(line, `// +`) {
							prefix += line + "\r\n"
						} else {
							break
						}
					}
				}
				if len(prefix) > 0 {
					prefix += "\r\n"
				}
				off := bytes.IndexByte(fdata, 0xa)
				if len(fdata) <= len(copyright) || !bytes.Equal(fdata[off+1:off+1+len(copyright)], copyright) {
					off := bytes.Index(fdata, []byte(`package`))
					if off == -1 {
						fmt.Println(`...package has not been found`)
					} else {
						out := append([]byte(fmt.Sprintf("%s// %s%s/%s\r\n", prefix, root, path,
							fname)), copyright...)
						out = append(out, fdata[off:]...)
						if err := ioutil.WriteFile(fullName, out, 0644); err == nil {
							fmt.Println(`...Overwrited`)
						} else {
							fmt.Println(`...` + err.Error())
						}
					}
				} else {
					fmt.Println(`...Skipped`)
				}
			}
		}
		if recurse && file.IsDir() {
			ProcessDir(fullName, recurse)
		}
	}
}

func main() {
	var err error
	if copyright, err = ioutil.ReadFile(`copyright.txt`); err == nil && len(copyright) > 0 {
		if copyright[len(copyright)-1] != 0xa {
			copyright = append(copyright, 0xa)
		}
		ProcessDir("../..", false)
		ProcessDir("../../cmd", true)
		ProcessDir("../../packages", true)
		ProcessDir("../../tools", true)
	} else {
		fmt.Println(err)
	}
}
