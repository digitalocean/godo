package modal

import (
	"context"
	stderrors "errors"
	"net/http"

	"github.com/digitalocean/godo"
	"github.com/digitalocean/godo/sandbox/compat/unsupported"
)

// convertError maps an error from the DigitalOcean client to the Modal error
// type that Modal code would test for.
func convertError(err error) error {
	if err == nil {
		return nil
	}
	var nie *unsupported.NotImplementedError
	if stderrors.As(err, &nie) {
		return err
	}
	if stderrors.Is(err, context.DeadlineExceeded) {
		return SandboxTimeoutError{Exception: err.Error()}
	}
	var er *godo.ErrorResponse
	if !stderrors.As(err, &er) {
		return err
	}
	msg := er.Message
	if msg == "" {
		msg = err.Error()
	}
	status := 0
	if er.Response != nil {
		status = er.Response.StatusCode
	}
	switch status {
	case http.StatusNotFound:
		return NotFoundError{Exception: msg}
	case http.StatusConflict:
		return ConflictError{Exception: msg}
	case http.StatusBadRequest, http.StatusUnprocessableEntity:
		return InvalidError{Exception: msg}
	case http.StatusTooManyRequests:
		return ResourceExhaustedError{Exception: msg}
	}
	return RemoteError{Exception: msg}
}
