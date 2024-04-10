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

func (m *staticNamingScheme) FileNames(_ RequestData) (string, string) {
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

func TestRecordTransport_RoundTripHappyPath(t *testing.T) {
	staticNS := &staticNamingScheme{
		reqFile:  path.Join(t.TempDir(), "request.txt"),
		respFile: path.Join(t.TempDir(), "response.txt"),
	}

	sampleReq := func() *http.Request {
		req, err := http.NewRequest("GET", "http://example.com/", nil)
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}
		return req
	}

	sampleResp := func() *http.Response {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString("response body")),
		}
	}
	sanitizer := RequestSanitizerFunc(func(req *http.Request) *http.Request {
		req.Header.Set("Sanitizer", "was run")
		return req
	})

	mockRT := &mockRoundTripper{
		recordedReq: nil,
		resp:        sampleResp(),
	}

	rt := newRecordTransport(mockRT, staticNS, sanitizer)

	gotResp, err := rt.RoundTrip(sampleReq())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotResp.StatusCode != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, gotResp.StatusCode)
	}
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
	if storedReq.Header.Get("Sanitizer") != "was run" {
		t.Errorf("expected header %q, got %q", "Sanitizer", storedReq.Header.Get("Sanitizer"))
	}

	respContent, err := os.ReadFile(staticNS.respFile)
	if err != nil {
		t.Fatalf("error when reading response file: %v", err)
	}
	storedResp, err := http.ReadResponse(bufio.NewReader(bytes.NewReader(respContent)), sampleReq())
	if err != nil {
		t.Fatalf("error when reading response from file: %v", err)
	}
	if storedResp.StatusCode != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, storedResp.StatusCode)
	}
}

func TestRecordTransport_NilBody(t *testing.T) {
	staticNS := &staticNamingScheme{
		reqFile:  path.Join(t.TempDir(), "request.txt"),
		respFile: path.Join(t.TempDir(), "response.txt"),
	}

	sampleReq := func() *http.Request {
		req, err := http.NewRequest("GET", "http://example.com/", nil)
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}
		return req
	}

	sampleResp := func() *http.Response {
		return &http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       io.NopCloser(bytes.NewBufferString("response body")),
		}
	}
	sanitizer := RequestSanitizerFunc(func(req *http.Request) *http.Request {
		req.Header.Set("Sanitizer", "was run")
		return req
	})

	mockRT := &mockRoundTripper{
		resp: sampleResp(),
	}

	rt := newRecordTransport(mockRT, staticNS, sanitizer)

	_, err := rt.RoundTrip(sampleReq())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}
