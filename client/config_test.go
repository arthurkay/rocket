package client

import (
	"testing"
)

func TestNormalizeAddress(t *testing.T) {

	addr, err := normalizeAddress("12345", "test")

	if err != nil {
		t.Fatal(err)
	}

	if addr != "127.0.0.1:12345" {
		t.Errorf("Unexpected normalized address: %s", addr)
	}

}

func TestValidateProtocol(t *testing.T) {

	if err := validateProtocol("http", "protocol"); err != nil {
		t.Error("http should be a valid protocol")
	}

	if err := validateProtocol("https", "protocol"); err != nil {
		t.Error("http should be a valid protocol")
	}

	if err := validateProtocol("http+https", "protocol"); err != nil {
		t.Error("http should be a valid protocol")
	}

	if err := validateProtocol("tcp", "protocol"); err != nil {
		t.Error("http should be a valid protocol")
	}

	if err := validateProtocol("foo", "protocol"); err == nil {
		t.Error("foo should be an invalid protocol")
	}

}
