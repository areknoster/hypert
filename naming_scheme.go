package hypert

import (
	"fmt"
	"os"
	"path"
	"strconv"
	"sync"
)

// NamingScheme defines an interface that is used by hypert's test client to store or retrieve files with HTTP requests.
//
// FileNames returns a pair of filenames that request and response should be stored in,
// when Record Mode is active, and retrieved from when Replay Mode is active.
//
// File names should be fully qualified (not relative).
//
// Each invocation should yield a pair of file names that
// wasn't yielded before during lifetime of given hypert's http client.
//
// This method should be safe for concurrent use.
// This requirement can be skipped, if you are the user of the package,
// and know, that all invocations would  be sequential.
type NamingScheme interface {
	FileNames(RequestData) (reqFile, respFile string)
}

// SequentialNamingScheme should be initialized using NewSequentialNamingScheme function.
// It names the files following (<dir>/0.req.http, <dir>/1.resp.http), (<dir>/1.req.http, <dir>/1.resp.http) convention.
type SequentialNamingScheme struct {
	dir string

	requestIndex   int
	requestIndexMx sync.Mutex
}

// NewSequentialNamingScheme initializes SequentialNamingScheme naming scheme, that implements NamingScheme interface.
//
// 'dir' parameter indicates, in which directory should the sequential requests and responses be placed.
// The directory is created with 0760 permissions if doesn't exists.
// You can use DefaultTestdataDir function for a sane default directory name.
func NewSequentialNamingScheme(dir string) (*SequentialNamingScheme, error) {
	err := os.MkdirAll(dir, 0760)
	if err != nil {
		return nil, fmt.Errorf("error creating directory: %w", err)
	}

	return &SequentialNamingScheme{
		dir: dir,
	}, nil
}

func (s *SequentialNamingScheme) FileNames(_ RequestData) (reqFile, respFile string) {
	s.requestIndexMx.Lock()
	requestIndex := strconv.Itoa(s.requestIndex)
	defer func() {
		s.requestIndex++
		s.requestIndexMx.Unlock()
	}()

	withDir := func(name string) string {
		return path.Join(s.dir, name)
	}

	return withDir(requestIndex + ".req.http"), withDir(requestIndex + ".resp.http")
}
