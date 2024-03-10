package htttest

import (
	"os"
	"testing"
)

func Test_callerDir(t *testing.T) {
	call := func() string {
		// the callstack depth of the package user
		return callerDir()
	}
	cd := call()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if cd != cwd {
		t.Fatalf("expected %s, got %s", cwd, cd)
	}
}
