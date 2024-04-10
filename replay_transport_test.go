package hypert

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
)

type mockT struct {
	T
	failed bool
	fatal  bool
	msg    string
}

func (m *mockT) Errorf(format string, args ...interface{}) {
	m.failed = true
	m.msg = fmt.Sprintf(format, args...)
}

func (m *mockT) Fatalf(format string, args ...interface{}) {
	m.failed = true
	m.fatal = true
	m.msg = fmt.Sprintf(format, args...)
}

func TestReplayTransport_HappyPath(t *testing.T) {
	namingScheme := &staticNamingScheme{
		reqFile:  "testdata/0.req.http",
		respFile: "testdata/0.resp.http",
	}
	const reqURL = "https://example.com"

	sanitizer := RequestSanitizerFunc(func(req *http.Request) *http.Request {
		req.Header.Set("Sanitizer", "was run")
		return req
	})
	mockedT := &mockT{}
	validator := RequestValidatorFunc(func(sanitizerT T, recorded RequestData, got RequestData) {
		if recorded.Method != "GET" {
			t.Errorf("expected read from file method to be GET, got %s", recorded.Method)
		}
		if recorded.URL.String() != reqURL {
			t.Errorf("expected URL read from file to be %s, got %s", reqURL, recorded.URL.String())
		}
		if recorded.Headers.Get("Sample-Header") != "sample-value" {
			t.Errorf("expected Sample-Header read from file to be set to sample-value, got %s", recorded.Headers.Get("Sample-Header"))
		}

		if got.Headers.Get("Sanitizer") != "was run" {
			t.Errorf("expected Sanitizer header to be set in the request passed to the validator")
		}
		if got.Method != "PUT" {
			t.Errorf("expected method to be PUT, got %s", got.Method)
		}

		sanitizerT.Errorf("this should fail the mocked T")
	})

	transport := newReplayTransport(mockedT, namingScheme, validator, sanitizer)
	req, err := http.NewRequest("PUT", "https://example.com", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	resp, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatalf("failed to round trip: %v", err)
	}
	if resp.Header.Get("SampleRespHeader") != "SampleRespHeaderValue" {
		t.Fatalf("expected SampleRespHeader to be set")
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status code to be 200, got %d", resp.StatusCode)
	}
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}
	const expectedBodyBytes = `Wassup, world?`
	if string(bodyBytes) != expectedBodyBytes {
		t.Errorf("expected response body to be %q, got %q", expectedBodyBytes, string(bodyBytes))
	}
	if mockedT.failed == false {
		t.Errorf("expected mocked T to fail")
	}
}

func TestReplayTransport_FilesDontExist(t *testing.T) {
	namingScheme := &staticNamingScheme{
		reqFile:  "testdata/doesnt_exist.req.http",
		respFile: "testdata/doesnt_exist.resp.http",
	}
	sanitizer := noopRequestSanitizer{}
	validator := noopRequestValidator{}
	sampleReq := func(t *testing.T) *http.Request {
		req, err := http.NewRequest("PUT", "https://example.com", nil)
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}
		return req
	}

	t.Run("should fail with help message if request file doesn't exist", func(t *testing.T) {
		mockedT := &mockT{}
		transport := newReplayTransport(mockedT, namingScheme, validator, sanitizer)

		_, err := transport.RoundTrip(sampleReq(t))
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		if !mockedT.failed || !mockedT.fatal {
			t.Errorf("expected mocked T to fail fatally ")
		}
		if !strings.Contains(mockedT.msg, helpMsgReplayFileDoesntExist) {
			t.Errorf("expected error message to contain helper error message, got %q", err.Error())
		}
	})

	t.Run("should fail with help message if response file doesn't exist", func(t *testing.T) {
		mockedT := &mockT{}
		transport := newReplayTransport(mockedT, namingScheme, validator, sanitizer)
		_, err := transport.RoundTrip(sampleReq(t))
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		if !mockedT.failed || !mockedT.fatal {
			t.Errorf("expected mocked T to fail fatally ")
		}
		if !strings.Contains(mockedT.msg, helpMsgReplayFileDoesntExist) {
			t.Errorf("expected error message to contain helper error message, got %q", err.Error())
		}
	})
}
