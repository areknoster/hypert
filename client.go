package htttest

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"runtime"
	"testing"
)

const RecordModeEnv = "HTTTEST_RECORD_MODE"

type config struct {
	isRecordMode     bool
	namingScheme     NamingScheme
	requestSanitizer RequestSanitizer
	parentHTTPClient *http.Client
}

type Option func(*config)

// WithDumpMode allows user to set the record mode explicitly to chosen value.
// You might prefer to use this option instead of setting the environment variable.
func WithDumpMode() Option {
	return func(c *config) {
		c.isRecordMode = true
	}
}

// WithNamingScheme allows user to set the naming scheme for the recorded requests.
// By default, the naming scheme is set to SequentialNamingScheme.
func WithNamingScheme(n NamingScheme) Option {
	return func(c *config) {
		c.namingScheme = n
	}
}

// WithParentHTTPClient allows user to set the parent http client.
func WithParentHTTPClient(c *http.Client) Option {
	return func(cfg *config) {
		cfg.parentHTTPClient = c
	}
}

// returns caller directory, assuming the caller is 2 levels up
func callerDir() string {
	_, filename, _, _ := runtime.Caller(2)
	return path.Dir(filename)
}

// NewDefaultTestClient returns a new http.Client that will either dump requests to the given dir or read them from it.
// It can be used to record and replay http requests.
func NewDefaultTestClient(t *testing.T, opts ...Option) *http.Client {
	t.Helper()
	scheme, err := NewSequentialNamingScheme(
		path.Join(callerDir(), "testdata", t.Name()),
	)
	if err != nil {
		t.Fatal(fmt.Errorf("failed to create naming scheme: %w", err))
	}

	cfg := &config{
		namingScheme:     scheme,
		requestSanitizer: DefaultRequestSanitizer(),
		isRecordMode:     os.Getenv(RecordModeEnv) != "",
		parentHTTPClient: &http.Client{},
	}
	for _, opt := range opts {
		opt(cfg)
	}

	var transport http.RoundTripper
	if cfg.isRecordMode {
		t.Log("test record request mode is on - all requests will be recorded")
		transport = newRecordTransport(cfg.parentHTTPClient.Transport, cfg.namingScheme, cfg.requestSanitizer)
	} else {
		t.Log("test record request mode is off - requests will be read from a directory if available, otherwise they will fail")
		transport = newReplayTransport(cfg.namingScheme)
	}
	cfg.parentHTTPClient.Transport = transport
	return cfg.parentHTTPClient
}
