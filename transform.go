package hypert

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

// ResponseTransform is a type that can transform a response, in case the real one is not feasible for test.
// Use WithResponseTransform option to apply transformations to the response.
type ResponseTransform interface {
	TransformResponse(r *http.Response) *http.Response
}

// ResponseTransformFunc is a convenience type that implements ResponseTransform interface.
type ResponseTransformFunc func(r *http.Response) *http.Response

func (f ResponseTransformFunc) TransformResponse(r *http.Response) *http.Response {
	return f(r)
}

// ComposeTransforms composes multiple transforms into a single one.
func ComposeTransforms(transforms ...ResponseTransform) ResponseTransform {
	return ResponseTransformFunc(func(r *http.Response) *http.Response {
		for _, transform := range transforms {
			r = transform.TransformResponse(r)
		}
		return r
	})
}

// TransformResponseFormatJSON formats json so it's easier to read.
func TransformResponseFormatJSON() ResponseTransform {
	return ResponseTransformFunc(func(r *http.Response) *http.Response {
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			return r
		}
		var prettyJSON bytes.Buffer
		err = json.Indent(&prettyJSON, bodyBytes, "", "  ")
		if err != nil {
			return r
		}
		r.Body = io.NopCloser(&prettyJSON)
		return r
	})
}
