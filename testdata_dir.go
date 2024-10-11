package hypert

import (
	"path/filepath"
	"runtime"
	"strings"
)

// DefaultTestdataDir returns fully qualified directory name following <your package directory>/testdata/<name of the test> convention.
//
// Note, that it relies on runtime.Caller function with given skip stack number.
// Because of that, usually you'd want to call this function directly in a file that belongs to a directory
// that the test data directory should be placed in.
func DefaultTestDataDir(t T) string {
	t.Helper()
	for i := 0; i < 8; i++ {
		_, file, _, ok := runtime.Caller(i)
		if !ok {
			t.Fatalf("failed to get caller")
		}
		if strings.HasSuffix(file, "_test.go") {
			return filepath.Join(filepath.Dir(file), "testdata", t.Name())
		}
	}
	t.Fatalf("failed to get testdata path")
	return ""
}
