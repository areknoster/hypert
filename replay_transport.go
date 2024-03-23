package hypert

import (
	"bufio"
	"bytes"
	"io"
	"net/http"
	"os"
	"testing"
)

type replayTransport struct {
	t         *testing.T
	scheme    NamingScheme
	validator RequestValidator
}

func newReplayTransport(t *testing.T, scheme NamingScheme, validator RequestValidator) *replayTransport {
	return &replayTransport{
		t:         t,
		scheme:    scheme,
		validator: validator,
	}
}

func (d *replayTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	reqFile, respFile := d.scheme.FileNames(requestDataFromRequest(d.t, req))
	d.readReqFromFile(reqFile)

	return d.readRespFromFile(respFile, req)
}

func (d *replayTransport) readReqFromFile(name string) (*http.Response, error) {
	return nil, nil
}

func (d *replayTransport) readRespFromFile(name string, req *http.Request) (*http.Response, error) {
	f, err := os.OpenFile(name, os.O_RDONLY, 000)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var buf bytes.Buffer
	_, err = io.Copy(&buf, f)
	if err != nil {
		return nil, err
	}
	return http.ReadResponse(bufio.NewReader(&buf), req)
}
