package htttest

import (
	"bufio"
	"bytes"
	"io"
	"net/http"
	"os"
)

type replayTransport struct {
	scheme NamingScheme
}

func newReplayTransport(scheme NamingScheme) *replayTransport {
	return &replayTransport{scheme: scheme}
}

func (d *replayTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	_, respFile := d.scheme.FileNames(requestMetaFromRequest(req))
	return d.readRespFromFile(respFile, req)
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
