package hypert

import (
	"bytes"
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
	validator := RequestValidatorFunc(func(sanitizerT T, recorded RequestData, got RequestData) error {
		if recorded.Method != http.MethodGet {
			t.Errorf("expected method GET, got %s", recorded.Method)
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
		if got.Method != http.MethodPut {
			t.Errorf("expected method to be PUT, got %s", got.Method)
		}

		sanitizerT.Errorf("this should fail the mocked T")
		return nil
	})

	transport := replayTransport{
		t:         mockedT,
		scheme:    namingScheme,
		validator: validator,
		sanitizer: sanitizer,
	}
	req, err := http.NewRequest(http.MethodPut, "https://example.com", http.NoBody)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	resp, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatalf("failed to round trip: %v", err)
	}
	defer resp.Body.Close()
	if resp.Header.Get("Samplerespheader") != "SampleRespHeaderValue" {
		t.Fatalf("expected Samplerespheader to be set")
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
	validator := RequestValidatorFunc(func(_ T, _ RequestData, _ RequestData) error {
		return nil
	})
	sampleReq := func(t *testing.T) *http.Request {
		req, err := http.NewRequest(http.MethodPut, "https://example.com", http.NoBody)
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}
		return req
	}

	t.Run("should fail with help message if request file doesn't exist", func(t *testing.T) {
		mockedT := &mockT{}
		transport := replayTransport{
			t:         mockedT,
			scheme:    namingScheme,
			validator: validator,
			sanitizer: sanitizer,
		}

		resp, err := transport.RoundTrip(sampleReq(t))
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
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
		transport := replayTransport{
			t:         mockedT,
			scheme:    namingScheme,
			validator: validator,
			sanitizer: sanitizer,
		}
		resp, err := transport.RoundTrip(sampleReq(t))
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
		if !mockedT.failed || !mockedT.fatal {
			t.Errorf("expected mocked T to fail fatally ")
		}
		if !strings.Contains(mockedT.msg, helpMsgReplayFileDoesntExist) {
			t.Errorf("expected error message to contain helper error message, got %q", err.Error())
		}
	})
}

func TestReplayTransport_TransformModes(t *testing.T) {
	namingScheme := &staticNamingScheme{
		reqFile:  "testdata/0.req.http",
		respFile: "testdata/0.resp.http",
	}

	sanitizer := NoOpRequestSanitizer{}
	validator := RequestValidatorFunc(func(_ T, _ RequestData, _ RequestData) error {
		return nil
	})

	transform := ResponseTransformFunc(func(r *http.Response) *http.Response {
		r.Body = io.NopCloser(bytes.NewBufferString("transformed response body"))
		return r
	})

	testCases := []struct {
		name           string
		transformMode  TransformRespMode
		expectedBody   string
		applyTransform bool
	}{
		{
			name:           "TransformRespModeNone",
			transformMode:  TransformRespModeNone,
			expectedBody:   "Wassup, world?",
			applyTransform: true, // Changed to true to test that transform is not applied
		},
		{
			name:           "TransformRespModeOnRecord",
			transformMode:  TransformRespModeOnRecord,
			expectedBody:   "Wassup, world?",
			applyTransform: true, // Changed to true to test that transform is not applied
		},
		{
			name:           "TransformRespModeAlways",
			transformMode:  TransformRespModeAlways,
			expectedBody:   "transformed response body",
			applyTransform: true,
		},
		{
			name:           "TransformRespModeRuntime",
			transformMode:  TransformRespModeRuntime,
			expectedBody:   "transformed response body",
			applyTransform: true,
		},
		{
			name:           "TransformRespModeOnReplay",
			transformMode:  TransformRespModeOnReplay,
			expectedBody:   "transformed response body",
			applyTransform: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockedT := &mockT{}
			transport := replayTransport{
				t:             mockedT,
				scheme:        namingScheme,
				validator:     validator,
				sanitizer:     sanitizer,
				transformMode: tc.transformMode,
			}

			if tc.applyTransform {
				transport.transform = transform
			}

			req, err := http.NewRequest(http.MethodGet, "https://example.com", http.NoBody)
			if err != nil {
				t.Fatalf("failed to create request: %v", err)
			}

			resp, err := transport.RoundTrip(req)
			if err != nil {
				t.Fatalf("failed to round trip: %v", err)
			}
			defer resp.Body.Close()

			bodyBytes, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("failed to read response body: %v", err)
			}

			if string(bodyBytes) != tc.expectedBody {
				t.Errorf("expected response body to be %q, got %q", tc.expectedBody, string(bodyBytes))
				t.Errorf("expected response body to be %q, got %q", tc.expectedBody, string(bodyBytes))
			}
		})
	}
}
