package hypert

import (
	"net/http"
	"net/url"
	"testing"
)

func TestRequestValidators(t *testing.T) {
	testCases := []struct {
		name      string
		validator RequestValidator
		recorded  RequestData
		got       RequestData
		expectErr bool
	}{
		{
			name:      "PathValidator_Match",
			validator: PathValidator(),
			recorded:  RequestData{URL: &url.URL{Path: "/foo"}},
			got:       RequestData{URL: &url.URL{Path: "/foo"}},
			expectErr: false,
		},
		{
			name:      "PathValidator_Mismatch",
			validator: PathValidator(),
			recorded:  RequestData{URL: &url.URL{Path: "/foo"}},
			got:       RequestData{URL: &url.URL{Path: "/bar"}},
			expectErr: true,
		},
		{
			name:      "QueryParamsValidator_Match",
			validator: QueryParamsValidator(),
			recorded:  RequestData{URL: &url.URL{RawQuery: "key1=value1&key2=value2"}},
			got:       RequestData{URL: &url.URL{RawQuery: "key2=value2&key1=value1"}},
			expectErr: false,
		},
		{
			name:      "QueryParamsValidator_Mismatch",
			validator: QueryParamsValidator(),
			recorded:  RequestData{URL: &url.URL{RawQuery: "key1=value1"}},
			got:       RequestData{URL: &url.URL{RawQuery: "key1=value2"}},
			expectErr: true,
		},
		{
			name:      "MethodValidator_Match",
			validator: MethodValidator(),
			recorded:  RequestData{Method: http.MethodGet},
			got:       RequestData{Method: http.MethodGet},
			expectErr: false,
		},
		{
			name:      "MethodValidator_Mismatch",
			validator: MethodValidator(),
			recorded:  RequestData{Method: http.MethodGet},
			got:       RequestData{Method: http.MethodPost},
			expectErr: true,
		},
		{
			name:      "SchemeValidator_Match",
			validator: SchemeValidator(),
			recorded:  RequestData{URL: &url.URL{Scheme: "https"}},
			got:       RequestData{URL: &url.URL{Scheme: "https"}},
			expectErr: false,
		},
		{
			name:      "SchemeValidator_Mismatch",
			validator: SchemeValidator(),
			recorded:  RequestData{URL: &url.URL{Scheme: "https"}},
			got:       RequestData{URL: &url.URL{Scheme: "http"}},
			expectErr: true,
		},
		{
			name:      "HeadersValidator_Match",
			validator: HeadersValidator(),
			recorded:  RequestData{Headers: http.Header{"Key1": []string{"Value1"}, "Key2": []string{"Value2"}}},
			got:       RequestData{Headers: http.Header{"Key2": []string{"Value2"}, "Key1": []string{"Value1"}}},
			expectErr: false,
		},
		{
			name:      "HeadersValidator_Mismatch",
			validator: HeadersValidator(),
			recorded:  RequestData{Headers: http.Header{"Key1": []string{"Value1"}}},
			got:       RequestData{Headers: http.Header{"Key1": []string{"Value2"}}},
			expectErr: true,
		},
		{
			name:      "HeadersValidator_Sanitized",
			validator: HeadersValidator(),
			recorded:  RequestData{Headers: http.Header{"Key1": []string{"SANITIZED"}}},
			got:       RequestData{Headers: http.Header{"Key1": []string{"Value1"}}},
			expectErr: false,
		},
		{
			name:      "HeadersValidator_Sanitized_Both",
			validator: HeadersValidator(),
			recorded:  RequestData{Headers: http.Header{"Key1": []string{"SANITIZED"}}},
			got:       RequestData{Headers: http.Header{"Key1": []string{"SANITIZED"}}},
			expectErr: false,
		},
		{
			name:      "HeadersValidator_SanitizedOneAnotherEmpty",
			validator: HeadersValidator(),
			recorded:  RequestData{Headers: http.Header{"Key1": []string{"SANITIZED"}}},
			got:       RequestData{Headers: http.Header{"Key1": []string{}}},
			expectErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mT := &mockT{}
			if err := tc.validator.Validate(mT, tc.recorded, tc.got); err != nil {
				t.Errorf("request validation failed: %v", err)
			}
			if tc.expectErr != mT.failed {
				t.Errorf("expected error value mismatch. Expected %v, got %v", tc.expectErr, mT.failed)
			}
		})
	}
}
