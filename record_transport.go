package hypert

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
)

// NoOpRequestSanitizer is a sanitizer that doesn't modify the request
type NoOpRequestSanitizer struct{}

func (NoOpRequestSanitizer) SanitizeRequest(req *http.Request) *http.Request {
	return req
}

type recordTransport struct {
	httpTransport http.RoundTripper
	namingScheme  NamingScheme
	sanitizer     RequestSanitizer
	transform     ResponseTransform
	transformMode TransformRespMode
}

func (d *recordTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if d.httpTransport == nil {
		d.httpTransport = http.DefaultTransport
	}

	reqData, err := requestDataFromRequest(req)
	if err != nil {
		return nil, fmt.Errorf("get request data: %w", err)
	}

	reqFile, respFile := d.namingScheme.FileNames(reqData)
	req, err = d.dumpReqToFile(reqFile, req)
	if err != nil {
		return nil, err
	}

	resp, err := d.httpTransport.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	if d.transformMode == TransformRespModeOnRecord || d.transformMode == TransformRespModeAlways {
		resp = d.transform.TransformResponse(resp)
	}
	resp, err = d.dumpRespToFile(respFile, req, resp)
	if err != nil {
		return nil, err
	}
	if d.transformMode == TransformRespModeRuntime {
		resp = d.transform.TransformResponse(resp)
	}

	return resp, nil
}

func (d *recordTransport) dumpReqToFile(name string, req *http.Request) (*http.Request, error) {
	if req.Body == nil {
		req.Body = http.NoBody
	}
	reqClone := req.Clone(req.Context())
	var originalReqBody bytes.Buffer
	teeReader := io.TeeReader(req.Body, &originalReqBody)
	reqClone.Body = io.NopCloser(teeReader)
	sanitizedReq := d.sanitizer.SanitizeRequest(reqClone)

	f, err := os.OpenFile(name, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0o644)
	if err != nil {
		return nil, err
	}
	err = sanitizedReq.WriteProxy(f)
	if err != nil {
		return nil, fmt.Errorf("write request to file %s: %w", name, err)
	}
	if err := f.Close(); err != nil {
		return nil, fmt.Errorf("close file %s: %w", name, err)
	}

	req.Body = io.NopCloser(&originalReqBody)
	return req, nil
}

func (d *recordTransport) dumpRespToFile(name string, req *http.Request, resp *http.Response) (*http.Response, error) {
	var buf bytes.Buffer
	err := resp.Write(&buf)
	if err != nil {
		return nil, err
	}

	f, err := os.OpenFile(name, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0o644)
	if err != nil {
		return nil, fmt.Errorf("open file %s: %w", name, err)
	}
	respBytes := buf.Bytes()

	_, err = io.Copy(f, bytes.NewReader(respBytes))
	if err != nil {
		return nil, err
	}
	if err := f.Close(); err != nil {
		return nil, fmt.Errorf("close file %s: %w", name, err)
	}

	resp, err = http.ReadResponse(bufio.NewReader(bytes.NewReader(respBytes)), req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
