package hypert

import (
	"bufio"
	"bytes"
	"io"
	"net/http"
	"os"
	"path"
	"testing"
)

type staticNamingScheme struct {
	reqFile  string
	respFile string
}

func (m *staticNamingScheme) FileNames(_ RequestData) (reqFile, respFile string) {
	return m.reqFile, m.respFile
}

type mockRoundTripper struct {
	recordedReq *http.Request
	resp        *http.Response
	err         error
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	m.recordedReq = req
	return m.resp, m.err
}

func TestRecordTransport_RoundTrip(t *testing.T) {
	testCases := []struct {
		name           string
		transformMode  TransformRespMode
		applyTransform bool
		expectedBody   string
	}{
		{
			name:           "TransformRespModeNone",
			transformMode:  TransformRespModeNone,
			applyTransform: true,
			expectedBody:   "original response body",
		},
		{
			name:           "TransformRespModeOnRecord",
			transformMode:  TransformRespModeOnRecord,
			applyTransform: true,
			expectedBody:   "transformed response body",
		},
		{
			name:           "TransformRespModeAlways",
			transformMode:  TransformRespModeAlways,
			applyTransform: true,
			expectedBody:   "transformed response body",
		},
		{
			name:           "TransformRespModeRuntime",
			transformMode:  TransformRespModeRuntime,
			applyTransform: true,
			expectedBody:   "transformed response body",
		},
		{
			name:           "TransformRespModeOnReplay",
			transformMode:  TransformRespModeOnReplay,
			applyTransform: true,
			expectedBody:   "original response body",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			staticNS := &staticNamingScheme{
				reqFile:  path.Join(t.TempDir(), "request.txt"),
				respFile: path.Join(t.TempDir(), "response.txt"),
			}

			sampleReq := func() *http.Request {
				// Replace nil with http.NoBody
				req, err := http.NewRequest("GET", "http://example.com/", http.NoBody)
				if err != nil {
					t.Fatalf("failed to create request: %v", err)
				}
				return req
			}

			sampleResp := func() *http.Response {
				resp := &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewBufferString("original response body")),
				}
				defer resp.Body.Close()
				return resp
			}

			transform := ResponseTransformFunc(func(r *http.Response) *http.Response {
				r.Body = io.NopCloser(bytes.NewBufferString("transformed response body"))
				return r
			})

			mockRT := &mockRoundTripper{
				// Add explanation for nolint directive
				resp: sampleResp(), //nolint:bodyclose // Response body is closed in the test cleanup
			}

			rt := recordTransport{
				httpTransport: mockRT,
				namingScheme:  staticNS,
				sanitizer:     NoOpRequestSanitizer{},
				transformMode: tc.transformMode,
			}

			if tc.applyTransform {
				rt.transform = transform
			}

			gotResp, err := rt.RoundTrip(sampleReq())
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			defer gotResp.Body.Close()

			if gotResp.StatusCode != http.StatusOK {
				t.Errorf("expected status code %d, got %d", http.StatusOK, gotResp.StatusCode)
			}

			// Check the response body
			bodyBytes, err := io.ReadAll(gotResp.Body)
			if err != nil {
				t.Fatalf("failed to read response body: %v", err)
			}

			if string(bodyBytes) != tc.expectedBody {
				t.Errorf("expected response body %q, got %q", tc.expectedBody, string(bodyBytes))
			}

			// Check the recorded request
			reqContent, err := os.ReadFile(staticNS.reqFile)
			if err != nil {
				t.Fatalf("error when reading request file: %v", err)
			}
			storedReq, err := http.ReadRequest(bufio.NewReader(bytes.NewReader(reqContent)))
			if err != nil {
				t.Fatalf("error when reading request from file: %v", err)
			}
			if storedReq.Method != "GET" {
				t.Errorf("expected method %q, got %q", "GET", storedReq.Method)
			}
			if storedReq.URL.String() != sampleReq().URL.String() {
				t.Errorf("expected URL %q, got %q", sampleReq().URL.String(), storedReq.URL.String())
			}

			// Check the recorded response
			respContent, err := os.ReadFile(staticNS.respFile)
			if err != nil {
				t.Fatalf("error when reading response file: %v", err)
			}
			storedResp, err := http.ReadResponse(bufio.NewReader(bytes.NewReader(respContent)), sampleReq())
			if err != nil {
				t.Fatalf("error when reading response from file: %v", err)
			}
			defer storedResp.Body.Close()
			if storedResp.StatusCode != http.StatusOK {
				t.Errorf("expected status code %d, got %d", http.StatusOK, storedResp.StatusCode)
			}

			storedRespBody, err := io.ReadAll(storedResp.Body)
			if err != nil {
				t.Fatalf("failed to read stored response body: %v", err)
			}

			expectedStoredBody := "original response body"
			if tc.transformMode == TransformRespModeOnRecord || tc.transformMode == TransformRespModeAlways {
				expectedStoredBody = "transformed response body"
			}

			if string(storedRespBody) != expectedStoredBody {
				t.Errorf("expected stored response body %q, got %q", expectedStoredBody, string(storedRespBody))
			}
		})
	}
}

func TestRecordTransport_NilBody(t *testing.T) {
	staticNS := &staticNamingScheme{
		reqFile:  path.Join(t.TempDir(), "request.txt"),
		respFile: path.Join(t.TempDir(), "response.txt"),
	}

	sampleReq := func() *http.Request {
		// Replace nil with http.NoBody
		req, err := http.NewRequest("GET", "http://example.com/", http.NoBody)
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}
		return req
	}

	sampleResp := func() *http.Response {
		resp := &http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       io.NopCloser(bytes.NewBufferString("response body")),
		}
		defer resp.Body.Close()
		return resp
	}

	mockRT := &mockRoundTripper{
		// Add explanation for nolint directive
		resp: sampleResp(), //nolint:bodyclose // Response body is closed in the test cleanup
	}

	rt := recordTransport{
		httpTransport: mockRT,
		namingScheme:  staticNS,
		sanitizer:     NoOpRequestSanitizer{},
	}

	resp, err := rt.RoundTrip(sampleReq())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer resp.Body.Close()

	// Check if the response file was created and contains the expected body
	respContent, err := os.ReadFile(staticNS.respFile)
	if err != nil {
		t.Fatalf("error when reading response file: %v", err)
	}

	if !bytes.Contains(respContent, []byte("response body")) {
		t.Errorf("expected response file to contain 'response body', got %s", string(respContent))
	}
}
