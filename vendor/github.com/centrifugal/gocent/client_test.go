package gocent

import "testing"

func TestNewClient(t *testing.T) {
	c := New(Config{})
	if c == nil {
		t.Errorf("New returned nil client")
	}
}
