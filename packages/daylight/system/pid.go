// MIT License
//
// Copyright (c) 2016 GenesisKernel
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
package system

import (
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/GenesisCommunity/go-genesis/packages/conf"
	"github.com/GenesisCommunity/go-genesis/packages/consts"

	log "github.com/sirupsen/logrus"
)

// CreatePidFile creats pid file
func CreatePidFile() error {
	pid := os.Getpid()
	data := []byte(strconv.Itoa(pid))
	return ioutil.WriteFile(conf.Config.GetPidPath(), data, 0644)
}

// RemovePidFile removes pid file
func RemovePidFile() error {
	return os.Remove(conf.Config.GetPidPath())
}

// ReadPidFile reads pid file
func ReadPidFile() (int, error) {
	pidPath := conf.Config.GetPidPath()
	if _, err := os.Stat(pidPath); err != nil {
		return 0, nil
	}

	data, err := ioutil.ReadFile(pidPath)
	if err != nil {
		log.WithFields(log.Fields{"path": pidPath, "error": err, "type": consts.IOError}).Error("reading pid file")
		return 0, err
	}

	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		log.WithFields(log.Fields{"data": data, "error": err, "type": consts.ConversionError}).Error("pid file data to int")
	}
	return pid, err
}
