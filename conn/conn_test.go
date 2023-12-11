package conn

import (
	"net"
	"testing"
)

func TestId(t *testing.T) {
	// Happy path
	c := &loggedConn{typ: "test", id: 1}

	id := c.Id()

	if id != "test:1" {
		t.Errorf("Expected test:1, got %s", id)
	}

	// Edge case - empty type
	c.typ = ""
	id = c.Id()

	if id != ":1" {
		t.Errorf("Expected :1, got %s", id)
	}
}

func TestDial(t *testing.T) {
	listener, _ := net.Listen("tcp", "localhost:0")
	defer listener.Close()

	// Happy path
	conn, err := Dial(listener.Addr().String(), "test", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	// Validate connection type
	if conn.typ != "test" {
		t.Error("unexpected connection type")
	}
}