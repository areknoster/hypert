package hypert

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
	"testing"
)

// RequestData is some data related to the request, that can be used to create filename in the NamingScheme's FileNames method implementations or during request validation.
// The fields are cloned from request's fields and their modification will not affect actual request's values.
type RequestData struct {
	Header    http.Header
	URL       *url.URL
	BodyBytes []byte
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

func requestDataFromRequest(t *testing.T, req *http.Request) RequestData {
	if req.Body == nil {
		req.Body = http.NoBody
	}
	var originalReqBody bytes.Buffer
	teeReader := io.TeeReader(req.Body, &originalReqBody)
	req.Body = io.NopCloser(&originalReqBody)
	gotBodyBytes, err := io.ReadAll(teeReader)
	if err != nil {
		t.Fatal("hypert: got error when reading request body")
	}

	return RequestData{
		Header:    req.Header.Clone(),
		URL:       cloneURL(req.URL),
		BodyBytes: gotBodyBytes,
	}
}
