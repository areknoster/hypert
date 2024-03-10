package htttest

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
)

type recordTransport struct {
	httpTransport http.RoundTripper
	namingScheme  NamingScheme
}

func newRecordTransport(httpTransport http.RoundTripper, namingScheme NamingScheme, sanitizer RequestSanitizer) *recordTransport {
	return &recordTransport{httpTransport: httpTransport, namingScheme: namingScheme}
}

func (d *recordTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if d.httpTransport == nil {
		d.httpTransport = http.DefaultTransport
	}

	reqFile, respFile := d.namingScheme.FileNames(requestMetaFromRequest(req))
	req, err := d.dumpReqToFile(reqFile, req)
	if err != nil {
		return nil, err
	}

	resp, err := d.httpTransport.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	resp, err = d.dumpRespToFile(respFile, req, resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (d *recordTransport) dumpReqToFile(name string, req *http.Request) (*http.Request, error) {
	var buf bytes.Buffer
	err := req.WriteProxy(&buf)
	if err != nil {
		return nil, err
	}

	f, err := os.OpenFile(name, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	reqBytes := buf.Bytes()
	_, err = io.Copy(f, bytes.NewReader(reqBytes))
	if err != nil {
		return nil, err
	}

	req, err = http.ReadRequest(bufio.NewReader(bytes.NewReader(reqBytes)))
	if err != nil {
		return nil, err
	}

	return req, nil
}

func (d *recordTransport) dumpRespToFile(name string, req *http.Request, resp *http.Response) (*http.Response, error) {
	var buf bytes.Buffer
	err := resp.Write(&buf)
	if err != nil {
		return nil, err
	}

	f, err := os.OpenFile(name, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		return nil, fmt.Errorf("error opening file %s: %w", name, err)
	}
	respBytes := buf.Bytes()

	_, err = io.Copy(f, bytes.NewReader(respBytes))
	if err != nil {
		return nil, err
	}
	if err := f.Close(); err != nil {
		return nil, err
	}

	resp, err = http.ReadResponse(bufio.NewReader(bytes.NewReader(respBytes)), req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
