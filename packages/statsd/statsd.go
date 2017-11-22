package statsd

import (
	"fmt"

	"github.com/cactus/go-statsd-client/statsd"
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
