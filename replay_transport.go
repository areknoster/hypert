package hypert

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
)

type replayTransport struct {
	t             T
	scheme        NamingScheme
	validator     RequestValidator
	sanitizer     RequestSanitizer
	transform     ResponseTransform
	transformMode TransformRespMode
}

func (d *replayTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	sanitizedReq := d.sanitizer.SanitizeRequest(req)
	requestData, err := requestDataFromRequest(sanitizedReq)
	if err != nil {
		return nil, err
	}
	reqFile, respFile := d.scheme.FileNames(requestData)
	recordedReq, err := d.readReqFromFile(reqFile)
	if err != nil {
		d.t.Fatalf("read request %s from file: %v", requestData, err)
		return nil, err
	}

	err = d.validator.Validate(d.t, recordedReq, requestData)
	if err != nil {
		return nil, fmt.Errorf("request validation failed: %w", err)
	}

	respFromFile, err := d.readRespFromFile(respFile, req)
	if err != nil {
		return nil, err
	}

	// Apply transformation based on the transform mode
	if d.transform != nil {
		switch d.transformMode {
		case TransformRespModeAlways, TransformRespModeRuntime, TransformRespModeOnReplay:
			respFromFile = d.transform.TransformResponse(respFromFile)
		case TransformRespModeNone, TransformRespModeOnRecord:
			// No transformation applied during replay for these modes
		}
	}

	return respFromFile, nil
}

const helpMsgReplayFileDoesntExist = `make sure, to record the request first using recordModeOn parameter in the TestClient.`

func (d *replayTransport) readReqFromFile(name string) (RequestData, error) {
	f, err := os.OpenFile(name, os.O_RDONLY, 0o000)
	if errors.Is(err, os.ErrNotExist) {
		return RequestData{}, fmt.Errorf("file %s does not exist -  %s", name, helpMsgReplayFileDoesntExist)
	}
	if err != nil {
		return RequestData{}, fmt.Errorf("open file %s: %w", name, err)
	}
	gotReq, err := http.ReadRequest(bufio.NewReader(f))
	if err != nil {
		return RequestData{}, fmt.Errorf("read request from file %s: %w", name, err)
	}
	reqData, err := requestDataFromRequest(gotReq)
	if err != nil {
		return RequestData{}, fmt.Errorf("get request data: %w", err)
	}
	return reqData, nil
}

func (d *replayTransport) readRespFromFile(name string, req *http.Request) (*http.Response, error) {
	f, err := os.OpenFile(name, os.O_RDONLY, 0o000)
	if errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("file %s does not exist -  %s", name, helpMsgReplayFileDoesntExist)
	}
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
