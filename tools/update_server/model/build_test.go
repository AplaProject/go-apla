package model

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuild_GetSystem(t *testing.T) {
	cases := []struct {
		version string
		os      string
		arch    string
	}{
		{version: "0.0.1", os: "linux", arch: "i386"},
		{version: "0.1.0", os: "linux", arch: "amd64"},
		{version: "0.1.1", os: "windows", arch: "i386"},
		{version: "1.0.0", os: "windows", arch: "amd64"},
		{version: "1.0.1", os: "linux", arch: "i386"},
		{version: "1.1.0", os: "linux", arch: "amd64"},
		{version: "1.1.1", os: "windows", arch: "i386"},
	}

	for _, c := range cases {
		b := Build{Version: c.version, OS: c.os, Arch: c.arch}

		e := fmt.Sprintf("%s_%s_%s", c.os, c.arch, c.version)
		assert.Equal(t, e, b.GetSystem())
	}
}
