package hypert

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// RequestData is some data related to the request, that can be used to create filename in the NamingScheme's FileNames method implementations or during request validation.
// The fields are cloned from request's fields and their modification will not affect actual request's values.
type RequestData struct {
	Headers   http.Header
	URL       *url.URL
	Method    string
	BodyBytes []byte
}

func (r RequestData) String() string {
	return fmt.Sprintf("%s %s", r.Method, r.URL)
}

func cloneURL(u *url.URL) *url.URL {
	if u == nil { // this shouldn't actually happen, unless there is very weird injected clients' transport setup
		return nil
	}
	var userInfo *url.Userinfo
	if u.User != nil {
		userInfoCopy := *u.User
		userInfo = &userInfoCopy
	}
	uCopy := *u
	uCopy.User = userInfo
	return &uCopy
}

func requestDataFromRequest(req *http.Request) (RequestData, error) {
	if req.Body == nil {
		req.Body = http.NoBody
	}
	var originalReqBody bytes.Buffer
	teeReader := io.TeeReader(req.Body, &originalReqBody)
	req.Body = io.NopCloser(&originalReqBody)
	gotBodyBytes, err := io.ReadAll(teeReader)
	if err != nil {
		return RequestData{}, fmt.Errorf("read request body: %w", err)
	}

	return RequestData{
		Headers:   req.Header.Clone(),
		URL:       cloneURL(req.URL),
		Method:    req.Method,
		BodyBytes: gotBodyBytes,
	}, nil
}
