package hypert

import (
	"io"
	"net/http"
	"strings"
)

// RequestSanitizer ensures, that no sensitive data is written to the request records.
// The sanitized version would be stored, whilst the original one would be sent in the record mode.
// It is allowed to mutate the request in place, because it is copied before invoking the RoundTrip method.
type RequestSanitizer interface {
	SanitizeRequest(req *http.Request) *http.Request
}

// DefaultRequestSanitizer returns a RequestSanitizer that sanitizes headers and query parameters.
func DefaultRequestSanitizer() RequestSanitizer {
	return ComposedRequestSanitizer(
		DefaultHeadersSanitizer(),
		DefaultQueryParamsSanitizer(),
	)
}

// RequestSanitizerFunc is a helper type for a function that implements RequestSanitizer interface.
type RequestSanitizerFunc func(req *http.Request) *http.Request

func (f RequestSanitizerFunc) SanitizeRequest(req *http.Request) *http.Request {
	return f(req)
}

// ComposedRequestSanitizer is a sanitizer that sequentially runs passed sanitizers.
func ComposedRequestSanitizer(s ...RequestSanitizer) RequestSanitizer {
	return RequestSanitizerFunc(func(req *http.Request) *http.Request {
		for _, s := range s {
			req = s.SanitizeRequest(req)
		}
		return req
	})
}

// HeadersSanitizer sets listed headers to "SANITIZED".
// Lookup DefaultHeadersSanitizer for a default value.
func HeadersSanitizer(headers ...string) RequestSanitizer {
	return RequestSanitizerFunc(func(req *http.Request) *http.Request {
		for _, header := range headers {
			if req.Header.Get(header) != "" {
				req.Header.Set(header, "SANITIZED")
			}
		}
		return req
	})
}

// DefaultHeadersSanitizer is HeadersSanitizer with the most common headers that should be sanitized in most cases.
func DefaultHeadersSanitizer() RequestSanitizer {
	return HeadersSanitizer(
		"Authorization",
		"Cookie",
		"X-Auth-Token",
		"X-API-Key",
		"Proxy-Authorization",
		"X-Forwarded-For",
		"Referrer",
		"X-Secret",
		"X-Access-Token",
		"X-Client-Secret",
		"X-Client-ID",
		"X-Auth",
		"X-Auth-Token",
	)
}

// SanitizerQueryParams sets listed query params in stored request URL to SANITIZED value.
// Lookup DefaultQueryParamsSanitizer for a default value.
func SanitizerQueryParams(params ...string) RequestSanitizer {
	return RequestSanitizerFunc(func(req *http.Request) *http.Request {
		q := req.URL.Query()
		for _, param := range params {
			if q.Has(param) {
				q.Set(param, "SANITIZED")
			}
		}
		req.URL.RawQuery = q.Encode()
		return req
	})
}

// DefaultQueryParamsSanitizer is SanitizerQueryParams with the most common query params that should be sanitized in most cases.
func DefaultQueryParamsSanitizer() RequestSanitizer {
	return SanitizerQueryParams(
		"access_token",
		"api_key",
		"auth",
		"key",
		"auth_token",
		"password",
		"secret",
		"token",
		"client_secret",
		"client_id",
		"signature",
		"sig",
		"session",
	)
}

// FormBodySanitizer sanitizes listed form fields in the request body for application/x-www-form-urlencoded content type.
// Lookup DefaultFormBodySanitizer for a default value.
func FormBodySanitizer(fields ...string) RequestSanitizer {
	fieldSet := make(map[string]bool)
	for _, field := range fields {
		fieldSet[field] = true
	}

	return RequestSanitizerFunc(func(req *http.Request) *http.Request {
		contentType := req.Header.Get("Content-Type")
		if !strings.HasPrefix(contentType, "application/x-www-form-urlencoded") {
			return req
		}

		if req.Body == nil {
			return req
		}

		// Parse form from URL-encoded body
		// Note: req.Body is a TeeReader in hypert's flow, so we can read it here
		err := req.ParseForm()
		if err != nil {
			return req
		}

		// Sanitize specified fields
		modified := false
		for key := range req.PostForm {
			if fieldSet[key] {
				req.PostForm.Set(key, "SANITIZED")
				modified = true
			}
		}

		// Reconstruct body if modified
		if modified {
			body := req.PostForm.Encode()
			req.Body = io.NopCloser(strings.NewReader(body))
			req.ContentLength = int64(len(body))
		}

		return req
	})
}

// DefaultFormBodySanitizer is FormBodySanitizer with the most common form fields that should be sanitized in most cases.
func DefaultFormBodySanitizer() RequestSanitizer {
	return FormBodySanitizer(
		"code", // OAuth authorization code
		"access_token",
		"refresh_token",
		"password",
		"secret",
		"client_secret",
		"client_id",
		"token",
		"api_key",
		"auth_token",
		"session",
	)
}
