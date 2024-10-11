package hypert

import "fmt"

// RequestValidator does assertions, that allows to make assertions on request that was caught by TestClient in the replay mode.
type RequestValidator interface {
	Validate(t T, recorded RequestData, got RequestData) error
}

type RequestValidatorFunc func(t T, recorded RequestData, got RequestData) error

func (f RequestValidatorFunc) Validate(t T, recorded, got RequestData) error {
	return f(t, recorded, got)
}

func ComposedRequestValidator(validators ...RequestValidator) RequestValidator {
	return RequestValidatorFunc(func(t T, recorded RequestData, got RequestData) error {
		for _, validator := range validators {
			if err := validator.Validate(t, recorded, got); err != nil {
				return fmt.Errorf("request validation failed: %w", err)
			}
		}
		return nil
	})
}

func DefaultRequestValidator() RequestValidator {
	return ComposedRequestValidator(
		PathValidator(),
		MethodValidator(),
		QueryParamsValidator(),
		HeadersValidator(),
		SchemeValidator(),
	)
}

func PathValidator() RequestValidator {
	return RequestValidatorFunc(func(t T, recorded RequestData, got RequestData) error {
		if recorded.URL.Path != got.URL.Path {
			t.Errorf("expected path '%s', got '%s'", recorded.URL.Path, got.URL.Path)
		}
		return nil
	})
}

// QueryParamsValidator validates query parameters of the request.
// It is not sensitive to the order of query parameters.
func QueryParamsValidator() RequestValidator {
	return RequestValidatorFunc(func(t T, recorded RequestData, got RequestData) error {
		recordedQ, gotQ := recorded.URL.Query(), got.URL.Query()

		for key := range recordedQ {
			recordedParam, gotParam := recordedQ.Get(key), gotQ.Get(key)
			if recordedParam != gotParam {
				t.Errorf("expected query parameter '%s' to be '%s', got '%s'", key, recordedParam, gotParam)
			}
			gotQ.Del(key)
		}
		for key := range gotQ {
			t.Errorf("unexpected query parameter '%s' with value '%s'", key, gotQ.Get(key))
		}
		return nil
	})
}

// MethodValidator validates the method of the request.
func MethodValidator() RequestValidator {
	return RequestValidatorFunc(func(t T, recorded RequestData, got RequestData) error {
		if recorded.Method != got.Method {
			t.Errorf("expected method '%s', got '%s'", recorded.Method, got.Method)
			return nil
		}
		return nil
	})
}

// SchemeValidator validates the scheme of the request.
func SchemeValidator() RequestValidator {
	return RequestValidatorFunc(func(t T, recorded RequestData, got RequestData) error {
		if recorded.URL.Scheme != got.URL.Scheme {
			t.Errorf("expected scheme '%s', got '%s'", recorded.URL.Scheme, got.URL.Scheme)
		}
		return nil
	})
}

// HeadersValidator validates headers of the request.
// It is not sensitive to the order of headers.
// User-Agent and Content-Length are removed from the comparison, because it is added deeper in the http client call.
func HeadersValidator() RequestValidator {
	return RequestValidatorFunc(func(t T, recorded RequestData, got RequestData) error {
		recordedHeaders := recorded.Headers.Clone()
		recordedHeaders.Del("User-Agent")
		recordedHeaders.Del("Content-Length")
		for key := range recordedHeaders {
			recordedHeader, gotHeader := recordedHeaders.Get(key), got.Headers.Get(key)
			if recordedHeader == "SANITIZED" {
				got.Headers.Del(key)
				continue
			}
			if recordedHeader != gotHeader {
				t.Errorf("expected header '%s' to be '%s', got '%s'", key, recordedHeader, gotHeader)
			}
			got.Headers.Del(key)
		}
		for key := range got.Headers {
			t.Errorf("unexpected header '%s' with value '%s'", key, got.Headers.Get(key))
		}
		return nil
	})
}
