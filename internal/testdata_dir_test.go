package internal

import (
	"fmt"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/areknoster/hypert"
)

type mockT struct {
	errMsg string
	hypert.T
}

func (m *mockT) Errorf(format string, args ...any) {
	m.errMsg = fmt.Sprintf(format, args...)
}

func (m *mockT) Helper() {}

func (m *mockT) Name() string {
	return "mockT"
}

func (m *mockT) Fatalf(format string, args ...any) {
	m.errMsg = fmt.Sprintf(format, args...)
}

func TestDefaultTestDataDir(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		path := hypert.DefaultTestDataDir(t)
		expected := filepath.Join(filepath.Dir(currentFilePath()), "testdata", t.Name())
		if path != expected {
			t.Fatalf("failed to get testdata path, got: %s, expected: %s", path, expected)
		}
	})

	t.Run("when wrapped a few times it works fine", func(t *testing.T) {
		mockT := &mockT{}
		path := WrapTestDataDir(mockT, 1)
		expected := filepath.Join(filepath.Dir(currentFilePath()), "testdata", mockT.Name())
		if path != expected {
			t.Fatalf("failed to get testdata path, got: %s, expected: %s", path, expected)
		}
	})
	t.Run("when wrapped a lot it finally fails", func(t *testing.T) {
		mockT := &mockT{}
		_ = WrapTestDataDir(mockT, 21)
		if mockT.errMsg == "" {
			t.Fatalf("expected error, got none")
		}
	})
}

func currentFilePath() string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		panic("failed to get current file path")
	}
	return file
}
