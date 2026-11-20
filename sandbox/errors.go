package sandbox

import (
	"errors"
	"net/http"

	"github.com/digitalocean/godo"
)

// ErrNoToken is returned by NewClient when no token is given or set in
// DIGITALOCEAN_ACCESS_TOKEN.
var ErrNoToken = errors.New("sandbox: no API token: set Config.Token or DIGITALOCEAN_ACCESS_TOKEN")

// Failed API calls return a *godo.ErrorResponse, which carries the message, the
// request ID and the HTTP response. These helpers test its status.

// IsNotFound reports whether err is a 404: the sandbox or template does not
// exist, or belongs to another team.
func IsNotFound(err error) bool { return hasStatus(err, http.StatusNotFound) }

// IsUnauthorized reports whether err is a 401.
func IsUnauthorized(err error) bool { return hasStatus(err, http.StatusUnauthorized) }

// IsPermissionDenied reports whether err is a 403: the token lacks the
// sandbox scope, or the feature is not enabled for the team.
func IsPermissionDenied(err error) bool { return hasStatus(err, http.StatusForbidden) }

// IsInvalidRequest reports whether err is a 400 or 422: a malformed request,
// or a sandbox that cannot take it in its current state.
func IsInvalidRequest(err error) bool {
	return hasStatus(err, http.StatusBadRequest) || hasStatus(err, http.StatusUnprocessableEntity)
}

// IsConflict reports whether err is a 409.
func IsConflict(err error) bool { return hasStatus(err, http.StatusConflict) }

// IsRateLimited reports whether err is a 429.
func IsRateLimited(err error) bool { return hasStatus(err, http.StatusTooManyRequests) }

func hasStatus(err error, status int) bool {
	var apiErr *godo.ErrorResponse
	return errors.As(err, &apiErr) && apiErr.Response != nil && apiErr.Response.StatusCode == status
}
