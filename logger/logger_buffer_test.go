package logger

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"
)

func testConfig(t *testing.T, bufferSize int) *Config {
	t.Helper()
	return &Config{
		FileName:      filepath.Join(t.TempDir(), "app.log"),
		RotateRule:    "1hour",
		MaxFileNum:    2,
		BufferSize:    bufferSize,
		FlushDuration: 50,
	}
}

func TestGetWriter_BufferSizeNegative_UsesSyncPath(t *testing.T) {
	conf := testConfig(t, -1)

	w, err := conf.getWriter()
	if err != nil {
		t.Fatalf("getWriter failed: %v", err)
	}
	t.Cleanup(func() { _ = w.Close() })

	if _, err = w.Write([]byte("sync-path\n")); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	if strings.Contains(strings.ToLower(strings.TrimSpace(typeName(w))), "asyncwriter") {
		t.Fatalf("expected sync writer for BufferSize=-1, got %T", w)
	}
}

func TestGetWriter_BufferSizeZero_UsesDefaultAsyncQueue(t *testing.T) {
	conf := testConfig(t, 0)
	conf.SetDefaults()

	if conf.BufferSize != 4096 {
		t.Fatalf("expected default BufferSize=4096, got %d", conf.BufferSize)
	}

	w, err := conf.getWriter()
	if err != nil {
		t.Fatalf("getWriter failed: %v", err)
	}
	t.Cleanup(func() { _ = w.Close() })

	if !strings.Contains(strings.ToLower(typeName(w)), "asyncwriter") {
		t.Fatalf("expected async writer for BufferSize=0(default), got %T", w)
	}
}

func TestGetWriter_BufferSizePositive_KeepsAsyncBehavior(t *testing.T) {
	conf := testConfig(t, 8)

	w, err := conf.getWriter()
	if err != nil {
		t.Fatalf("getWriter failed: %v", err)
	}
	t.Cleanup(func() { _ = w.Close() })

	if !strings.Contains(strings.ToLower(typeName(w)), "asyncwriter") {
		t.Fatalf("expected async writer for BufferSize>0, got %T", w)
	}
}

func typeName(v any) string {
	return fmt.Sprintf("%T", v)
}
