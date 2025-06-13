package hypert

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path"
	"path/filepath"
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
	err := os.MkdirAll(dir, 0o760)
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

// PathBasedNamingScheme creates filenames based on the request path
type PathBasedNamingScheme struct {
	dir     string
	mu      sync.Mutex
	counter map[string]int
}

// FileNames returns filenames based on the request path
func (s *PathBasedNamingScheme) FileNames(data RequestData) (reqFile, respFile string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.counter == nil {
		s.counter = make(map[string]int)
	}

	// Get full URL including query parameters
	fullURL := data.URL.String()

	// Create a hash of the full URL
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(fullURL)))

	// Add counter if this URL has been seen before
	count := s.counter[fullURL]
	s.counter[fullURL]++

	filename := hash[:16] // Use first 16 characters of hash for uniqueness
	if count > 0 {
		filename = fmt.Sprintf("%s-%d", filename, count)
	}

	return filepath.Join(s.dir, filename+".req.http"), filepath.Join(s.dir, filename+".resp.http")
}

// NewPathBasedNamingScheme creates a new PathBasedNamingScheme
func NewPathBasedNamingScheme(dir string) (*PathBasedNamingScheme, error) {
	err := os.MkdirAll(dir, 0o760)
	if err != nil {
		return nil, fmt.Errorf("error creating directory: %w", err)
	}

	return &PathBasedNamingScheme{
		dir: dir,
	}, nil
}

// ContentHashNamingScheme creates filenames based on the request path and content hash
type ContentHashNamingScheme struct {
	dir string
}

// FileNames returns filenames based on the request path and content hash
func (s *ContentHashNamingScheme) FileNames(data RequestData) (reqFile, respFile string) {
	// Get path from URL and sanitize it for filename use
	path := "root"
	if data.URL != nil {
		path = data.URL.Path
		if path == "" {
			path = "root"
		}
	}

	// Create a hash of path and content
	content := data.BodyBytes
	if content == nil {
		content = []byte{}
	}

	// Combine path and content for hashing
	hashInput := append([]byte(path), content...)
	hash := fmt.Sprintf("%x", sha256.Sum256(hashInput))

	// Use first 16 characters of hash for uniqueness
	filename := hash[:16]

	return filepath.Join(s.dir, filename+".req.http"), filepath.Join(s.dir, filename+".resp.http")
}

// NewContentHashNamingScheme creates a new ContentHashNamingScheme
func NewContentHashNamingScheme(dir string) (*ContentHashNamingScheme, error) {
	err := os.MkdirAll(dir, 0o760)
	if err != nil {
		return nil, fmt.Errorf("error creating directory: %w", err)
	}

	return &ContentHashNamingScheme{
		dir: dir,
	}, nil
}
