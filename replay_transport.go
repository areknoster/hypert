package hypert

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
)

type replayTransport struct {
	t         T
	scheme    NamingScheme
	validator RequestValidator
	sanitizer RequestSanitizer
}

func newReplayTransport(t T, scheme NamingScheme, validator RequestValidator, sanitizer RequestSanitizer) *replayTransport {
	return &replayTransport{
		t:         t,
		scheme:    scheme,
		validator: validator,
		sanitizer: sanitizer,
	}
}

func (d *replayTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	sanitizedReq := d.sanitizer.SanitizeRequest(req)
	requestData, err := requestDataFromRequest(sanitizedReq)
	if err != nil {
		return nil, fmt.Errorf("error getting request data: %w", err)
	}
	reqFile, respFile := d.scheme.FileNames(requestData)
	recordedReq, err := d.readReqFromFile(reqFile)
	if err != nil {
		return nil, fmt.Errorf("error reading request %s from file: %w", requestData, err)
	}

	d.validator.Validate(d.t, recordedReq, requestData)

	respFromFile, err := d.readRespFromFile(respFile, req)
	if err != nil {
		return nil, fmt.Errorf("error reading response from file %s: %w", respFile, err)
	}
	return respFromFile, nil
}

func (d *replayTransport) readReqFromFile(name string) (RequestData, error) {
	f, err := os.OpenFile(name, os.O_RDONLY, 000)
	if err != nil {
		return RequestData{}, fmt.Errorf("error opening file %s: %w", name, err)
	}
	gotReq, err := http.ReadRequest(bufio.NewReader(f))
	if err != nil {
		return RequestData{}, fmt.Errorf("error reading request from file %s: %w", name, err)
	}
	reqData, err := requestDataFromRequest(gotReq)
	if err != nil {
		return RequestData{}, fmt.Errorf("error getting request data: %w", err)
	}
	return reqData, nil
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
