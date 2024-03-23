package hypert

import (
	"testing"
)

// RequestValidator does assertions, that allows to make assertions on request that was caught by TestClient in the replay mode.
type RequestValidator interface {
	Validate(t *testing.T, recorded RequestData, got RequestData)
}

type RequestValidatorFunc func(t *testing.T, recorded RequestData, got RequestData)

func (f RequestValidatorFunc) Validate(t *testing.T, recorded RequestData, got RequestData) {
	f(t, recorded, got)
}

func ComposedRequestValidator(validators ...RequestValidator) RequestValidator {
	return RequestValidatorFunc(func(t *testing.T, recorded RequestData, got RequestData) {
		for _, validator := range validators {
			validator.Validate(t, recorded, got)
		}
	})
}
