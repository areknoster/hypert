package htttest

import (
	"os"
	"testing"
)

import (
	"path/filepath"
)

func Test_callerDir(t *testing.T) {
	call := func() string {
		// the callstack depth of the package user
		return callerDir()
	}
	cd := filepath.ToSlash(call())
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if cd != filepath.ToSlash(cwd) {
		t.Fatalf("expected %s, got %s", cwd, cd)
	}
}
