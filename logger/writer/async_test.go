package writer

import (
	"bytes"
	"testing"
)

type nopWriteCloser struct {
	bytes.Buffer
}

func (n *nopWriteCloser) Close() error { return nil }

func TestNewAsync_NegativeBufferSize_NoPanic(t *testing.T) {
	raw := &nopWriteCloser{}
	w := NewAsync(-1, 0, raw)

	if _, err := w.Write([]byte("hello")); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	if err := w.Close(); err != nil {
		t.Fatalf("close failed: %v", err)
	}

	if got := raw.String(); got != "hello" {
		t.Fatalf("unexpected written content: %q", got)
	}
}
