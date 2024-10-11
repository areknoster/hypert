package hypert

import (
	"net/http"
)

type config struct {
	isRecordMode     bool
	transformMode    TransformRespMode
	transform        ResponseTransform
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

type TransformRespMode int

const (
	// TransformRespModeNone. No transformations are applied to the response. Default value.
	TransformRespModeNone TransformRespMode = iota
	// TransformRespModeOnRecord will apply transform only in record mode, so the transformed response would be visible in stored files.
	// In replay mode, whatever is stored in the file will be used, without any transformations.
	TransformRespModeOnRecord
	// TransformRespModeAlways will apply transformation to the response in both record and replay modes.
	// When in replay mode, the file is not modified, so the response is not transformed.
	TransformRespModeAlways
	// TransformModeRuntime will apply transformation only in runtime. This means, that the files would always contain untransformed responses,
	// but the response will be transformed on the fly during the test execution.
	TransformRespModeRuntime

	// TransformRespModeOnReplay will apply transformation only in replay mode.
	// This is useful when there is some other action (e.g. oauth flow) that needs to be performed in record mode,
	// but then the response is not feasible in replay mode. (e.g. we want to override the redirect url in oauth responses)
	TransformRespModeOnReplay
)

func WithResponseTransform(mode TransformRespMode, transform ResponseTransform) Option {
	return func(cfg *config) {
		cfg.transformMode = mode
		cfg.transform = transform
	}
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
		transport = &recordTransport{
			httpTransport: cfg.parentHTTPClient.Transport,
			namingScheme:  cfg.namingScheme,
			sanitizer:     cfg.requestSanitizer,
			transformMode: cfg.transformMode,
			transform:     cfg.transform,
		}
	} else {
		t.Log("hypert: replay request mode - requests will be read from previously stored files.")
		transport = &replayTransport{
			t:             t,
			scheme:        cfg.namingScheme,
			validator:     cfg.requestValidator,
			sanitizer:     cfg.requestSanitizer,
			transform:     cfg.transform,
			transformMode: cfg.transformMode,
		}
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
		requestsDir := DefaultTestDataDir(t)
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
