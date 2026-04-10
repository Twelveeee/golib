package writer

import (
	"os"
	"path/filepath"
	"testing"
)

type staticRotateProducer struct {
	info RotateInfo
}

func (s *staticRotateProducer) Get() RotateInfo {
	return s.info
}

func (s *staticRotateProducer) RegisterCallBack(func(info RotateInfo)) {}

func (s *staticRotateProducer) Stop() error {
	return nil
}

func TestRotateWriter_PreExistingFileOnStartup(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "app.log")

	if err := os.WriteFile(logPath, []byte("old\n"), 0644); err != nil {
		t.Fatalf("prepare pre-existing log file failed: %v", err)
	}

	producer := &staticRotateProducer{
		info: RotateInfo{
			RawName:  logPath,
			FilePath: logPath,
		},
	}

	w, err := NewRotate(&RotateOption{FileProducer: producer})
	if err != nil {
		t.Fatalf("NewRotate failed: %v", err)
	}
	defer func() {
		_ = w.Close()
	}()

	if _, err = w.Write([]byte("new\n")); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	if err = w.Close(); err != nil {
		t.Fatalf("close failed: %v", err)
	}

	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read log file failed: %v", err)
	}

	if string(content) != "old\nnew\n" {
		t.Fatalf("unexpected log content: %q", string(content))
	}
}
