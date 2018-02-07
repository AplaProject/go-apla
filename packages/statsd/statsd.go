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

package statsd

import (
	"fmt"
	"strings"

	"github.com/cactus/go-statsd-client/statsd"
)

const (
	Count = ".count"
	Time  = ".time"
)

var Client statsd.Statter

func Init(host string, port int, name string) error {
	var err error
	Client, err = statsd.NewClient(fmt.Sprintf("%s:%d", host, port), name)
	if err != nil {
		return err
	}
	return nil
}

func Close() {
	if Client != nil {
		Client.Close()
	}
}

func APIRouteCounterName(method, pattern string) string {
	routeCounterName := strings.Replace(strings.Replace(pattern, ":", "", -1), "/", ".", -1)
	return "api." + strings.ToLower(method) + "." + routeCounterName
}

func DaemonCounterName(daemonName string) string {
	return "daemon." + daemonName
}
