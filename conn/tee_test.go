package conn

import (
	"bytes"
	"io"
	"testing"
)

func TestTeeWrite(t *testing.T) {

    // Happy path
    var buf bytes.Buffer
    tee := Tee{wr: &buf}
    n, err := tee.Write([]byte("hello"))
    if n != 5 || err != nil {
        t.Errorf("Write did not write 5 bytes") 
    }

    // Validate write to pipe
    rd := tee.WriteBuffer()
    buf2 := make([]byte, 5)
    n2, err := rd.Read(buf2)
    if n2 != 5 || err != nil || string(buf2) != "hello" {
        t.Errorf("Write pipe did not contain written data")
    }

    // Validate close on error
/*     tee.wr = &ErrorWriter{}
    n, err = tee.Write([]byte("test"))
    if n != 0 || err == nil {
        t.Errorf("Expected error on write")
    } */
    if _, err := tee.WriteBuffer().Read(buf2); err != io.EOF {
        t.Errorf("Expected write pipe to be closed on error")
    }
}