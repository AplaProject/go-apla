package main

import (
	"testing"

	"github.com/AplaProject/go-apla/tools/update_server/config"
)

func TestConfig(t *testing.T) {
	c := config.Config{}
	err := c.Read()
	if err != nil {
		t.Error("config read error")
	}
	if c.Host != "localhost" || c.Port != "8090" || c.Login != "admin" || c.Pass != "admin" || c.DBPath != "update.db" {
		t.Error("config data error")
	}
}

func TestVersions(t *testing.T) {
	versions := []string{"1.0.0", "1.1.0"}
	version, _ := getLast(versions)
	if version != "1.1.0" {
		t.Error("want: 1.1.0, have: ", version)
	}

	versions = []string{"1.0.0", "1.1.0", "2.1.0"}
	version, _ = getLast(versions)
	if version != "2.1.0" {
		t.Error("want: 2.1.0, have: ", version)
	}

	versions = []string{"1.0.0", "1.1.0", "1.11.0"}
	version, _ = getLast(versions)
	if version != "1.11.0" {
		t.Error("want: 1.11.0, have: ", version)
	}

	versions = []string{"1.0.0", "1.1.0", "1.1.1"}
	version, _ = getLast(versions)
	if version != "1.1.1" {
		t.Error("want: 1.1.1, have: ", version)
	}

	versions = []string{"1.0.0", "1.1.0", "1.1a.0"}
	version, _ = getLast(versions)
	if version != "1.1.0" {
		t.Error("want: 1.1a.0, have: ", version)
	}

	versions = []string{"12.0.0", "1.1.0", "1.1a.0"}
	version, _ = getLast(versions)
	if version != "12.0.0" {
		t.Error("want: 1.1a.0, have: ", version)
	}

	versions = []string{"12.0", "11a.1", "2.1"}
	version, _ = getLast(versions)
	if version != "12.0.0" {
		t.Error("want: 1.1a.0, have: ", version)
	}
}
