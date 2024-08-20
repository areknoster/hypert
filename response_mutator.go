package hypert

import (
	"net/http"
)

// ResponseMutator is an interface for types that can modify HTTP responses.
type ResponseMutator interface {
	MutateResponse(resp *http.Response) (*http.Response, error)
}

// ResponseMutatorFunc is a helper type for a function that implements ResponseMutator interface.
type ResponseMutatorFunc func(resp *http.Response) (*http.Response, error)

func (f ResponseMutatorFunc) MutateResponse(resp *http.Response) (*http.Response, error) {
	return f(resp)
}
