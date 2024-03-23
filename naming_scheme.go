package hypert

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"sync"
)

// RequestMeta is some data related to the request, that can be used to create filename in the NamingScheme's FileNames method implementations.
// The fields are cloned from request's URL and their modification will not affect actual request's values.
type RequestMeta struct {
	Header http.Header
	URL    *url.URL
}

func cloneURL(u *url.URL) *url.URL {
	if u == nil { // this shouldn't actually happen, unless there is very weird injected clients' transport setup
		return nil
	}
	var userInfo *url.Userinfo
	if u.User != nil {
		userInfoCopy := *u.User
		userInfo = &userInfoCopy
	}
	uCopy := *u
	uCopy.User = userInfo
	return &uCopy
}

func requestMetaFromRequest(req *http.Request) RequestMeta {
	return RequestMeta{
		Header: req.Header.Clone(),
		URL:    cloneURL(req.URL),
	}
}

// NamingScheme defines an interface that is used by hypert's test client to store or retrieve files with HTTP requests.
//
// FileNames returns a pair of filenames that request and response should be stored in, when Record Mode is active, and retrieved from when Replay Mode is active.
//
// File names should be fully qualified (not relative).
//
// Each invocation should yield a pair of file names that wasn't yielded before during lifetime of given hypert's http client.
//
// This method should be safe for concurrent use. This requirement can be skipped, if you are the user of the package, and know, that all invocations would  be sequential.
type NamingScheme interface {
	FileNames(RequestMeta) (reqFile, respFile string)
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

func (s *SequentialNamingScheme) FileNames(_ RequestMeta) (reqFile, respFile string) {
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
