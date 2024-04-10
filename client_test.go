package hypert

import (
	"net/http"
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

type noopRequestSanitizer struct{}

func (n noopRequestSanitizer) SanitizeRequest(req *http.Request) *http.Request {
	return req
}

type noopRequestValidator struct{}

func (n noopRequestValidator) Validate(t T, recorded RequestData, got RequestData) {}

func Test_configWithDefaults(t *testing.T) {
	t.Run("should return default config", func(t *testing.T) {
		cfg := configWithDefaults(t, false, nil)
		if cfg.isRecordMode {
			t.Error("expected isRecordMode to be false")
		}
		if cfg.namingScheme == nil {
			t.Error("expected namingScheme to be set")
		}
		if cfg.requestSanitizer == nil {
			t.Error("expected requestSanitizer to be set")
		}
		if cfg.parentHTTPClient == nil {
			t.Error("expected parentHTTPClient to be set")
		}
		if cfg.requestValidator == nil {
			t.Error("expected requestValidator to be set")
		}
	})
	t.Run("should return config with options", func(t *testing.T) {
		sanitizer := &noopRequestSanitizer{}
		validator := &noopRequestValidator{}
		namingScheme := &staticNamingScheme{}
		parentHTTPClient := &http.Client{}

		cfg := configWithDefaults(t, false, []Option{
			WithRequestValidator(validator),
			WithRequestSanitizer(sanitizer),
			WithNamingScheme(namingScheme),
			WithParentHTTPClient(parentHTTPClient),
		})
		if cfg.isRecordMode {
			t.Error("expected isRecordMode to be false")
		}
		if cfg.namingScheme != namingScheme {
			t.Error("expected namingScheme to be set")
		}
		if cfg.requestSanitizer != sanitizer {
			t.Error("expected requestSanitizer to be set")
		}
		if cfg.parentHTTPClient != parentHTTPClient {
			t.Error("expected parentHTTPClient to be set")
		}
		if cfg.requestValidator != validator {
			t.Error("expected requestValidator to be set")
		}
	})
}

func Test_TestClient(t *testing.T) {
	t.Run("when record mode is on, it should use record transport", func(t *testing.T) {
		c := TestClient(t, true)
		if c.Transport == nil {
			t.Fatal("expected transport to be set")
		}
		if _, ok := c.Transport.(*recordTransport); !ok {
			t.Error("expected transport to be recordTransport")
		}
	})
	t.Run("when record mode is off, it should use replay transport", func(t *testing.T) {
		c := TestClient(t, false)
		if c.Transport == nil {
			t.Fatal("expected transport to be set")
		}
		if _, ok := c.Transport.(*replayTransport); !ok {
			t.Error("expected transport to be replayTransport")
		}
	})
}
