package hypert

import (
	"net/http"
	"testing"
)

func TestNewComposedRequestSanitizer(t *testing.T) {
	s := ComposedRequestSanitizer(
		RequestSanitizerFunc(func(r *http.Request) *http.Request {
			r.Header.Set("X-Request-Sanitizer-Test1", "1")
			return r
		}),
		RequestSanitizerFunc(func(r *http.Request) *http.Request {
			r.Header.Set("X-Request-Sanitizer-Test2", "2")
			return r
		}),
	)

	req, _ := http.NewRequest("GET", "http://example.com", nil)
	req = s.SanitizeRequest(req)
	if req.Header.Get("X-Request-Sanitizer-Test1") != "1" {
		t.Errorf("expected 1, got %s", req.Header.Get("X-Request-Sanitizer-Test1"))
	}
	if req.Header.Get("X-Request-Sanitizer-Test2") != "2" {
		t.Errorf("expected 2, got %s", req.Header.Get("X-Request-Sanitizer-Test2"))
	}
}

func TestHeadersSanitizer(t *testing.T) {
	s := HeadersSanitizer("X-Request-Header-Sanitizer")

	req, _ := http.NewRequest("GET", "http://example.com", nil)
	req.Header.Set("X-Request-Header-Sanitizer", "1")

	originalHeadersCount := len(req.Header)
	sanitizedReq := s.SanitizeRequest(req)
	if sanitizedReq.Header.Get("X-Request-Header-Sanitizer") != "SANITIZED" {
		t.Errorf("expected SANITIZED, got %s", sanitizedReq.Header.Get("X-Request-Header-Sanitizer"))
	}
	if len(sanitizedReq.Header) != originalHeadersCount {
		t.Errorf("original headers count doesn't equal headers count after sanitization. Expected %d, got %d", originalHeadersCount, len(req.Header))
	}
}

func TestQueryParamsSanitizer(t *testing.T) {
	s := QueryParamsSanitizer("param1", "param2")

	req, _ := http.NewRequest("GET", "http://example.com?param1=1&param2=2&param3=3", nil)
	sanitizedReq := s.SanitizeRequest(req)
	if sanitizedReq.URL.RawQuery != "param1=SANITIZED&param2=SANITIZED&param3=3" {
		t.Errorf("expected param1=SANITIZED&param2=SANITIZED, got %s", sanitizedReq.URL.RawQuery)
	}
}
