package htttest

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"strconv"
	"sync"
)

// RequestMeta is some data related to the request, that can be used to create filename in the naming scheme
type RequestMeta struct{}

func requestMetaFromRequest(req *http.Request) RequestMeta {
	return RequestMeta{}
}

type NamingScheme interface {
	FileNames(RequestMeta) (reqFile, respFile string)
}

type SequentialNamingScheme struct {
	dir string

	requestIndex   int
	requestIndexMx sync.Mutex
}

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
