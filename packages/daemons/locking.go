// MIT License
//
// Copyright (c) 2016-2018 GenesisKernel
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

package daemons

import (
	"context"
	"sync"
	"time"

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/model"

	log "github.com/sirupsen/logrus"
)

var mutex = sync.Mutex{}

// WaitDB waits for the end of the installation
func WaitDB(ctx context.Context) error {
	// There is could be the situation when installation is not over yet.
	// Database could be created but tables are not inserted yet

	if model.DBConn != nil && CheckDB() {
		return nil
	}

	// poll a base with period
	tick := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-tick.C:
			if model.DBConn != nil && CheckDB() {
				return nil
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// CheckDB check if installation complete or not
func CheckDB() bool {
	install := &model.Install{}

	err := install.Get()
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting install")
	}

	if install.Progress == model.ProgressComplete {
		return true
	}

	return false
}

// DBLock locks daemons
func DBLock() {
	mutex.Lock()
}

// DBUnlock unlocks database
func DBUnlock() {
	mutex.Unlock()
}
