package hypert

import "net/http"

// RequestSanitizer ensures, that no sensitive data is written to the request records.
// The sanitized version would be stored, whilst the original one would be sent in the record mode.
// It is allowed to mutate the request in place, becuase it is copied before invoking the RoundTrip method.
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

type RequestSanitizerFunc func(req *http.Request) *http.Request

func (f RequestSanitizerFunc) SanitizeRequest(req *http.Request) *http.Request {
	return f(req)
}

func ComposedRequestSanitizer(s ...RequestSanitizer) RequestSanitizer {
	return RequestSanitizerFunc(func(req *http.Request) *http.Request {
		for _, s := range s {
			req = s.SanitizeRequest(req)
		}
		return req
	})
}

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

func QueryParamsSanitizer(params ...string) RequestSanitizer {
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

func DefaultQueryParamsSanitizer() RequestSanitizer {
	return QueryParamsSanitizer(
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
