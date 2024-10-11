package hypert

import (
	"path/filepath"
	"runtime"
	"testing"
)

func TestDefaultTestDataDir(t *testing.T) {
	path := DefaultTestDataDir(t)
	expected := filepath.Join(filepath.Dir(currentFilePath()), "testdata", t.Name())
	if path != expected {
		t.Fatalf("failed to get testdata path, got: %s, expected: %s", path, expected)
	}
}

func currentFilePath() string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		panic("failed to get current file path")
	}
	return file
}
