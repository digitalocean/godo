// Package unsupported holds the error that the migration compat clients
// return for an API that DigitalOcean Sandboxes does not offer yet.
//
// Import it as github.com/digitalocean/godo/sandbox/compat, which re-exports
// the names; this leaf package exists so the per-vendor packages can use the
// error without importing each other.
package unsupported

import (
	"errors"
	"fmt"
)

// ErrNotImplemented is the sentinel that every *NotImplementedError wraps.
// Test for it with errors.Is.
var ErrNotImplemented = errors.New("not implemented on DigitalOcean Sandboxes")

// NotImplementedError says that a method, or an option of a method, of a
// vendor SDK has no DigitalOcean Sandboxes API behind it yet.
type NotImplementedError struct {
	// Platform is the vendor SDK being mirrored, such as "Daytona" or "Modal".
	Platform string
	// Method is the vendor method, such as "Sandbox.Archive". An option that
	// cannot be honored is written "Client.Create (Labels)".
	Method string
	// Needs names the missing DigitalOcean API.
	Needs string
}

// New returns a *NotImplementedError.
func New(platform, method, needs string) *NotImplementedError {
	return &NotImplementedError{Platform: platform, Method: method, Needs: needs}
}

func (e *NotImplementedError) Error() string {
	return fmt.Sprintf("%s.%s is not implemented on DigitalOcean Sandboxes yet: needs %s", e.Platform, e.Method, e.Needs)
}

// Unwrap lets errors.Is(err, ErrNotImplemented) match.
func (e *NotImplementedError) Unwrap() error { return ErrNotImplemented }
