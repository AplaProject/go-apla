package daemons

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type netAddressTest struct {
	Address string
	Result  string
	IsError bool
}

func TestNormalizeHostAddress(t *testing.T) {
	defaultPort := 7078

	tests := []netAddressTest{
		netAddressTest{
			Address: "127.0.0.1",
			Result:  "127.0.0.1:7078",
		},
		netAddressTest{
			Address: "127.0.0.1:8080",
			Result:  "127.0.0.1:8080",
		},
		netAddressTest{
			Address: "2001:4860:0:2001::68",
			Result:  "[2001:4860:0:2001::68]:7078",
		},
		netAddressTest{
			Address: "[2001:4860:0:2001::68]:8081",
			Result:  "[2001:4860:0:2001::68]:8081",
		},
		netAddressTest{
			Address: "127.0.1",
			IsError: true,
		},
	}

	for _, test := range tests {
		addr, err := NormalizeHostAddress(test.Address, int64(defaultPort))
		if err != nil && !test.IsError {
			t.Error(err)
			return
		}

		assert.Equal(t, addr, test.Result, "should be equal")
	}
}
