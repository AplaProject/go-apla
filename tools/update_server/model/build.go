package model

import (
	"sort"
	"time"

	"fmt"

	"strings"

	version "github.com/hashicorp/go-version"
	"github.com/pkg/errors"
)

type Version struct {
	Number string `json:"number"`
	OS     string `json:"os"`
	Arch   string `json:"arch"`
}

func NewVersion(from string) (Version, error) {
	var v Version
	p := strings.Split(from, "_")

	if len(from) < 5 || len(p) != 3 {
		return v, fmt.Errorf("version parsing: %s", from)
	}

	v.OS = p[0]
	v.Arch = p[1]
	v.Number = p[2]

	var vf bool
	for _, av := range availableVersions {
		if v.Arch == av.Arch && v.OS == av.OS {
			vf = true
			break
		}
	}
	if !vf {
		return v, fmt.Errorf("cant find version with arch: %s, os %s", v.Arch, v.OS)
	}

	return v, nil
}

// String is formatting version to string
func (b *Version) String() string {
	if b.OS == "" || b.Arch == "" || b.Number == "" {
		return ""
	}

	return fmt.Sprintf("%s_%s_%s", b.OS, b.Arch, b.Number)
}

var availableVersions = []Version{
	{OS: "linux", Arch: "amd64"},
	{OS: "windows", Arch: "amd64"},
	{OS: "freebsd", Arch: "amd64"},
	{OS: "darwin", Arch: "amd64"},
}

// Build is storing build data with needed for update params
type Build struct {
	Time time.Time `json:"time"`
	Name string    `json:"name"`
	Body []byte    `json:"body"`
	Sign []byte    `json:"sign"`

	Version

	StartBlock uint64 `json:"start_block"`
	IsCritical bool   `json:"is_critical"`
}

// ValidateSystem is checking os+arch correctness
func (b *Version) Validate() bool {
	_, err := version.NewVersion(b.Number)
	if err != nil {
		return false
	}

	for _, av := range availableVersions {
		if av.OS == b.OS && av.Arch == b.Arch {
			return true
		}
	}

	return false
}

// GetAvailableVersions is returning available versions
func GetAvailableVersions() []Version {
	return availableVersions
}

// VersionFilter is filtering and leaves only the necessary versions and sorting by asc
func VersionFilter(versions []Version, os string, arch string) []Version {
	var fv []Version
	for _, v := range versions {
		if v.OS == os && v.Arch == arch {
			fv = append(fv, v)
		}
	}

	sortVersions(fv)
	return fv
}

// GetLastVersion returns the latest version for required os+arch from the version array
func GetLastVersion(versions []Version, os string, arch string) (Version, error) {
	lv := Version{Arch: arch, OS: os}

	if len(versions) == 0 {
		return Version{}, nil
	}

	vb := make([]Version, len(versions))
	copy(vb, versions)

	vb = VersionFilter(vb, os, arch)

	var bv []*version.Version
	for _, dv := range vb {
		v, err := version.NewVersion(dv.Number)
		if err != nil {
			return lv, errors.Wrapf(err, "creation version")
		}

		bv = append(bv, v)
	}
	if len(bv) == 0 {
		return Version{}, nil
	}

	sort.Sort(version.Collection(bv))
	last := bv[len(bv)-1]

	lv.Number = last.String()
	return lv, nil
}

func sortVersions(versions []Version) error {
	var bv []*version.Version
	for _, dv := range versions {
		v, err := version.NewVersion(dv.Number)
		if err != nil {
			return errors.Wrapf(err, "creation version")
		}

		bv = append(bv, v)
	}
	if len(bv) == 0 {
		return nil
	}

	sort.Sort(version.Collection(bv))
	return nil
}
