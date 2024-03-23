package hypert

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
	// windows compatibility. It is not handled in the function itself,
	// because this form is expected by other os.* function that leverage dir name in this package.
	cd := filepath.ToSlash(call())
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if cd != filepath.ToSlash(cwd) {
		t.Fatalf("expected %s, got %s", cwd, cd)
	}
}
