package hypert

import (
	"bytes"
	"io"
	"net/http"
	"testing"
)

func TestTransformResponseFormatJSON(t *testing.T) {
	type User struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	tests := []struct {
		name      string
		body      string
		want      string
		transform ResponseTransform
	}{
		{
			name: "Simple JSON",
			body: `{"name":"John","age":30}`,
			want: `{
  "name": "John",
  "age": 30
}`,
			transform: TransformResponseFormatJSON(),
		},
		{
			name: "JSON with nested object",
			body: `{"name":"John","age":30,"address":{"city":"New York","country":"USA"}}`,
			want: `{
  "name": "John",
  "age": 30,
  "address": {
    "city": "New York",
    "country": "USA"
  }
}`,
			transform: TransformResponseFormatJSON(),
		},
		{
			name: "composed",
			body: `"wassup`,
			want: `{
  "initial": "transformation"
}`,
			transform: ComposeTransforms(
				ResponseTransformFunc(func(r *http.Response) *http.Response {
					r.Body = io.NopCloser(bytes.NewBufferString(`{"initial":"transformation"}`))
					return r
				}),
				TransformResponseFormatJSON(),
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := &http.Response{Body: io.NopCloser(bytes.NewBufferString(tt.body))}
			tt.transform.TransformResponse(got)
			bodyBytes, err := io.ReadAll(got.Body)
			if err != nil {
				t.Fatalf("Failed to read response body: %v", err)
			}
			got.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			if string(bodyBytes) != tt.want {
				t.Errorf("Response body = %v, want %v", string(bodyBytes), tt.want)
			}
		})
	}
}
