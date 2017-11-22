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

func APIRouteToCounterName(method, pattern string) string {
	routeCounterName := strings.Replace(strings.Replace(pattern, ":", "", -1), "/", ".", -1)
	return "api" + "." + strings.ToLower(method) + "." + routeCounterName
}
