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

var (
	copyright []byte
)

// ProcessDir proceeds the specified directory and inserts the copyright at the beginning of the files
func ProcessDir(dir string, recurse bool) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		fmt.Println(err)
	}

	for _, file := range files {
		fname := file.Name()
		fullName := filepath.Join(dir, fname)
		if strings.HasSuffix(fname, `.go`) {
			fmt.Printf(fullName)
			if fdata, err := ioutil.ReadFile(fullName); err != nil {
				fmt.Println(err)
				os.Exit(1)
			} else {
				if len(fdata) <= len(copyright) || bytes.Compare(fdata[:len(copyright)], copyright) != 0 {
					off := bytes.Index(fdata, []byte(`package`))
					if off == -1 {
						fmt.Println(`...package has not been found`)
					} else {
						out := copyright
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
		ProcessDir("../../packages", true)
	} else {
		fmt.Println(err)
	}
}
