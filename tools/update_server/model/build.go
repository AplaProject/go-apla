package model

import (
	"fmt"
	"strings"
	"time"
)

// Build is storing build data with needed for update params
type Build struct {
	Version string    `json:"version"`
	OS      string    `json:"os"`
	Arch    string    `json:"arch"`
	Date    time.Time `json:"time"`
	Name    string    `json:"name"`
	Body    []byte    `json:"body"`
	Sign    []byte    `json:"sign"`

	StartBlock uint64 `json:"start_block"`
	IsCritical bool   `json:"is_critical"`

	Downloaded int  `json:"-"`
	Deprecated bool `json:"-"`
}

// TODO this is not completed list yet
var availableSystems = []struct {
	OS   string
	ARCH string
}{
	{OS: "linux", ARCH: "i386"},
	{OS: "linux", ARCH: "amd64"},
	{OS: "windows", ARCH: "i386"},
	{OS: "windows", ARCH: "amd64"},
}

// ValidateSystem is checking os+arch correctness
func (b *Build) ValidateSystem() bool {
	for _, av := range availableSystems {
		if av.OS == b.OS && av.ARCH == b.Arch {
			return true
		}
	}

	return false
}

func (b *Build) GetAvailableSystems() string {
	var sys []string
	for _, s := range availableSystems {
		sys = append(sys, fmt.Sprintf("%s_%s", s.OS, s.ARCH))
	}
	return strings.Join(sys, ",")
}

func (b *Build) GetSystem() string {
	if b.OS == "" || b.Arch == "" || b.Version == "" {
		return ""
	}

	return fmt.Sprintf("%s_%s_%s", b.OS, b.Arch, b.Version)
}

func GetLastVersion(versions []string) Build {
	// TODO
	return Build{}
}
