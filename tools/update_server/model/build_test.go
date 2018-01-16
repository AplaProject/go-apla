package model

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var versionCases = []struct {
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

func TestBuild_GetSystem(t *testing.T) {
	for _, c := range versionCases {
		b := Version{Number: c.version, OS: c.os, Arch: c.arch}

		e := fmt.Sprintf("%s_%s_%s", c.os, c.arch, c.version)
		assert.Equal(t, e, b.String())
	}
}

func TestGetLastVersion(t *testing.T) {
	sv := []Version{}
	for _, c := range versionCases {
		sv = append(sv, Version{Number: c.version, OS: c.os, Arch: c.arch})
	}

	glv := func(t *testing.T, versions []Version, os string, arch string) Version {
		v, err := GetLastVersion(versions, os, arch)
		require.NoError(t, err)
		return v
	}

	assert.Equal(t, Version{Number: "1.1.1", OS: "windows", Arch: "i386"}, glv(t, sv, "windows", "i386"))
	assert.Equal(t, Version{Number: "1.0.1", OS: "linux", Arch: "i386"}, glv(t, sv, "linux", "i386"))
	assert.Equal(t, Version{Number: "1.0.0", OS: "windows", Arch: "amd64"}, glv(t, sv, "windows", "amd64"))
	assert.Equal(t, Version{Number: "1.1.0", OS: "linux", Arch: "amd64"}, glv(t, sv, "linux", "amd64"))
}

func TestGetLastVersionEmpty(t *testing.T) {
	var sv []Version
	var eb Version

	v, err := GetLastVersion(sv, "windows", "i386")
	require.NoError(t, err)
	assert.Equal(t, eb, v)
}
