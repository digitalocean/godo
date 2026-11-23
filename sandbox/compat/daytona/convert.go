package daytona

import (
	"context"
	stderrors "errors"
	"net/http"
	"time"

	"github.com/digitalocean/godo"
	"github.com/digitalocean/godo/sandbox/compat/daytona/apiclient"
	dterrors "github.com/digitalocean/godo/sandbox/compat/daytona/errors"
	"github.com/digitalocean/godo/sandbox/compat/internal/shared"
	"github.com/digitalocean/godo/sandbox/compat/unsupported"
)

const platform = "Daytona"

// defaultTimeout is the wait Daytona's Client.Create applies when no
// options.WithTimeout is given.
const defaultTimeout = 60 * time.Second

func notImplemented(method, needs string) error { return unsupported.New(platform, method, needs) }

func optionNotImplemented(method, option, needs string) error {
	return shared.Option(platform, method, option, needs)
}

// convertError maps an error from the DigitalOcean client to the Daytona error
// type that Daytona code would test for.
func convertError(err error) error {
	if err == nil {
		return nil
	}
	var nie *unsupported.NotImplementedError
	if stderrors.As(err, &nie) {
		return err
	}
	if stderrors.Is(err, context.DeadlineExceeded) {
		return dterrors.NewDaytonaTimeoutError(err.Error())
	}
	var er *godo.ErrorResponse
	if !stderrors.As(err, &er) {
		return dterrors.NewDaytonaError(err.Error(), 0, nil)
	}
	status := 0
	var headers http.Header
	if er.Response != nil {
		status = er.Response.StatusCode
		headers = er.Response.Header
	}
	msg := er.Message
	if msg == "" {
		msg = err.Error()
	}
	switch {
	case status == http.StatusBadRequest, status == http.StatusUnprocessableEntity:
		return dterrors.NewDaytonaValidationError(msg, headers)
	case status == http.StatusUnauthorized:
		return dterrors.NewDaytonaAuthenticationError(msg, headers)
	case status == http.StatusForbidden:
		return dterrors.NewDaytonaForbiddenError(msg, headers)
	case status == http.StatusNotFound:
		return dterrors.NewDaytonaNotFoundError(msg, headers)
	case status == http.StatusConflict:
		return dterrors.NewDaytonaConflictError(msg, headers)
	case status == http.StatusTooManyRequests:
		return dterrors.NewDaytonaRateLimitError(msg, headers)
	case status >= 500 && status <= 599:
		return dterrors.NewDaytonaServerError(msg, status, headers)
	}
	return dterrors.NewDaytonaError(msg, status, headers)
}

func notFound(what string) error { return dterrors.NewDaytonaNotFoundError(what+" not found", nil) }

func mapState(s godo.SandboxState) apiclient.SandboxState {
	switch s {
	case godo.SandboxStateRunning:
		return apiclient.SANDBOXSTATE_STARTED
	case godo.SandboxStatePaused:
		return apiclient.SANDBOXSTATE_STOPPED
	case godo.SandboxStateCreating:
		return apiclient.SANDBOXSTATE_CREATING
	}
	return apiclient.SANDBOXSTATE_UNKNOWN
}

func timeString(t godo.Timestamp) *string {
	if t.IsZero() {
		return nil
	}
	s := t.Time.UTC().Format(time.RFC3339)
	return &s
}
