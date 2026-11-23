// Package compat holds migration aids for code written against other sandbox
// providers. Each sub-package mirrors a vendor's Go SDK, so moving to
// DigitalOcean Sandboxes mostly means changing an import path; underneath, the
// calls use the DigitalOcean Sandbox package (github.com/digitalocean/godo/sandbox).
//
//   - github.com/digitalocean/godo/sandbox/compat/daytona mirrors the Daytona Go
//     SDK (github.com/daytonaio/daytona/libs/sdk-go v0.190.0).
//   - github.com/digitalocean/godo/sandbox/compat/modal mirrors the Modal Go SDK
//     (github.com/modal-labs/modal-client/go v0.11.2).
//
// There is no E2B package: E2B has no official Go SDK.
//
// These packages are for migration only. New code should use the sandbox
// package. Nothing in godo imports them, so they cost nothing unless you do.
//
// Every vendor method that DigitalOcean Sandboxes cannot do yet keeps its
// signature and returns a *NotImplementedError. Test for it with
//
//	errors.Is(err, compat.ErrNotImplemented)
//
// As DigitalOcean adds the missing APIs (file transfer, PTY, ports, volumes,
// and creating a sandbox from a snapshot), the stubs are replaced with real calls.
package compat

import (
	"github.com/digitalocean/godo/sandbox/compat/daytona"
	"github.com/digitalocean/godo/sandbox/compat/daytona/types"
	"github.com/digitalocean/godo/sandbox/compat/modal"
	"github.com/digitalocean/godo/sandbox/compat/unsupported"
)

// NotImplementedError is returned for a vendor method, or a method option,
// that has no DigitalOcean Sandboxes API behind it yet.
type NotImplementedError = unsupported.NotImplementedError

// ErrNotImplemented is the sentinel that every *NotImplementedError wraps.
var ErrNotImplemented = unsupported.ErrNotImplemented

// DaytonaClient is the Daytona client, running on DigitalOcean Sandboxes.
type DaytonaClient = daytona.Client

// ModalClient is the Modal client, running on DigitalOcean Sandboxes.
type ModalClient = modal.Client

// NewDaytonaClient returns a DaytonaClient. A nil config reads
// DIGITALOCEAN_ACCESS_TOKEN. DaytonaConfig.APIKey is the DigitalOcean token.
func NewDaytonaClient(config *types.DaytonaConfig) (*DaytonaClient, error) {
	return daytona.NewClientWithConfig(config)
}

// NewModalClient returns a ModalClient. A nil params reads
// DIGITALOCEAN_ACCESS_TOKEN. ClientParams.TokenSecret is the DigitalOcean token.
func NewModalClient(params *modal.ClientParams) (*ModalClient, error) {
	return modal.NewClientWithOptions(params)
}
