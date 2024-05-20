package hypert

import (
	"net/http"
	"path"
	"runtime"
	"testing"
)

type config struct {
	isRecordMode     bool
	namingScheme     NamingScheme
	requestSanitizer RequestSanitizer
	requestValidator RequestValidator
	parentHTTPClient *http.Client
}

// Option can be used to customize TestClient behaviour. See With* functions to find customization options
type Option func(*config)

// WithNamingScheme allows user to set the naming scheme for the recorded requests.
// By default, the naming scheme is set to SequentialNamingScheme.
func WithNamingScheme(n NamingScheme) Option {
	return func(c *config) {
		c.namingScheme = n
	}
}

// WithParentHTTPClient allows user to set the custom parent http client.
// TestClient's will use passed client's transport in record mode to make actual HTTP calls.
func WithParentHTTPClient(c *http.Client) Option {
	return func(cfg *config) {
		cfg.parentHTTPClient = c
	}
}

// WithRequestSanitizer configures RequestSanitizer.
// You may consider using RequestSanitizerFunc, ComposedRequestSanitizer, NoopRequestSanitizer,
// SanitizerQueryParams, HeadersSanitizer helper functions to compose sanitization rules or implement your own, custom sanitizer.
func WithRequestSanitizer(sanitizer RequestSanitizer) Option {
	return func(cfg *config) {
		cfg.requestSanitizer = sanitizer
	}
}

// WithRequestValidator allows user to set the request validator.
func WithRequestValidator(v RequestValidator) Option {
	return func(cfg *config) {
		cfg.requestValidator = v
	}
}

// returns caller directory, assuming the caller is 2 levels up
func callerDir() string {
	_, filePath, _, _ := runtime.Caller(3)
	return path.Dir(filePath)
}

// DefaultTestdataDir returns fully qualified directory name following <your package directory>/testdata/<name of the test> convention.
//
// Note, that it relies on runtime.Caller function with given skip stack number.
// Because of that, usually you'd want to call this function directly in a file that belongs to a directory
// that the test data directory should be placed in.
func DefaultTestdataDir(t *testing.T) string {
	return path.Join(callerDir(), "testdata", t.Name())
}

// TestClient returns a new http.Client that will either dump requests to the given dir or read them from it.
// It is the main entrypoint for using the hypert's capabilities.
// It provides sane defaults, that can be overwritten using Option arguments. Option's can be initialized using With* functions.
//
// The returned *http.Client should be injected to given component before the tests are run.
//
// In most scenarios, you'd set recordModeOn to true during the development, when you have set up the authentication to the HTTP API you're using.
// This will result in the requests and response pairs being stored in <package name>/testdata/<test name>/<sequential number>.(req|resp).http
// Before the requests are stored, they are sanitized using DefaultRequestSanitizer. It can be adjusted using WithRequestSanitizer option.
// Ensure that sanitization works as expected, otherwise sensitive details might be committed
//
// recordModeOn should be false when given test is not actively worked on, so in most cases the committed value should be false.
// This mode will result in the requests and response pairs previously stored being replayed, mimicking interactions with actual HTTP APIs,
// but skipping making actual calls.
func TestClient(t T, recordModeOn bool, opts ...Option) *http.Client {
	t.Helper()
	cfg := configWithDefaults(t, recordModeOn, opts)

	var transport http.RoundTripper
	if cfg.isRecordMode {
		t.Log("hypert: record request mode - requests will be stored")
		transport = newRecordTransport(cfg.parentHTTPClient.Transport, cfg.namingScheme, cfg.requestSanitizer)
	} else {
		t.Log("hypert: replay request mode - requests will be read from previously stored files.")
		transport = newReplayTransport(t, cfg.namingScheme, cfg.requestValidator, cfg.requestSanitizer)
	}
	cfg.parentHTTPClient.Transport = transport
	return cfg.parentHTTPClient
}

func configWithDefaults(t T, recordModeOn bool, opts []Option) *config {
	cfg := &config{
		isRecordMode: recordModeOn,
	}
	for _, opt := range opts {
		opt(cfg)
	}
	if cfg.namingScheme == nil {
		requestsDir := path.Join(callerDir(), "testdata", t.Name())
		t.Logf("hypert: using sequential naming scheme in %s directory", requestsDir)
		scheme, err := NewSequentialNamingScheme(requestsDir)
		if err != nil {
			t.Fatalf("failed to create naming scheme: %s", err.Error())
		}
		cfg.namingScheme = scheme
	}
	if cfg.requestSanitizer == nil {
		cfg.requestSanitizer = DefaultRequestSanitizer()
	}
	if cfg.parentHTTPClient == nil {
		cfg.parentHTTPClient = &http.Client{}
	}
	if cfg.requestValidator == nil {
		cfg.requestValidator = DefaultRequestValidator()
	}
	return cfg
}

// T is a subset of testing.T interface that is used by hypert's functions.
// custom T's implementation can be used to e.g. make logs silent, stop failing on errors and others.
type T interface {
	Helper()
	Name() string
	Log(args ...any)
	Logf(format string, args ...any)
	Error(args ...any)
	Errorf(format string, args ...any)
	Fatal(args ...any)
	Fatalf(format string, args ...any)
}
