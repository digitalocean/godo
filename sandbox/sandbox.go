package sandbox

import (
	"context"
	"fmt"
	"time"

	"github.com/digitalocean/godo"
)

// Sandbox is a DigitalOcean Sandbox. Get one from Client.Create, Get,
// Connect or List.
type Sandbox struct {
	client *Client

	// Info is the record as last returned by the API. Refresh updates it.
	Info godo.Sandbox

	// Commands starts, lists and streams the commands running in this sandbox.
	Commands *Commands
}

func newSandbox(c *Client, info godo.Sandbox) *Sandbox {
	sb := &Sandbox{client: c, Info: info}
	sb.Commands = &Commands{client: c, sandboxID: info.ID}
	return sb
}

// ID is the sandbox ID.
func (s *Sandbox) ID() string { return s.Info.ID }

// State is the state as last returned by the API. It can be stale: Refresh
// reads it again.
func (s *Sandbox) State() godo.SandboxState { return s.Info.State }

// Template is the name of the template the sandbox was created from.
func (s *Sandbox) Template() string { return s.Info.Template }

// Refresh reads the sandbox again.
func (s *Sandbox) Refresh(ctx context.Context) error {
	info, _, err := s.client.API.Sandboxes.Get(ctx, s.Info.ID)
	if err != nil {
		return err
	}
	s.Info = *info
	return nil
}

// IsRunning reports whether the sandbox is running right now. It reads the API.
func (s *Sandbox) IsRunning(ctx context.Context) (bool, error) {
	if err := s.Refresh(ctx); err != nil {
		return false, err
	}
	return s.Info.State == godo.SandboxStateRunning, nil
}

// Pause pauses the sandbox. Resume it to use it again.
func (s *Sandbox) Pause(ctx context.Context) error {
	info, _, err := s.client.API.Sandboxes.Pause(ctx, s.Info.ID)
	if err != nil {
		return err
	}
	s.Info = *info
	return nil
}

// Resume resumes a paused sandbox.
func (s *Sandbox) Resume(ctx context.Context) error {
	info, _, err := s.client.API.Sandboxes.Resume(ctx, s.Info.ID)
	if err != nil {
		return err
	}
	s.Info = *info
	return nil
}

// Kill deletes the sandbox. It reports false, with no error, when the sandbox
// is already gone.
func (s *Sandbox) Kill(ctx context.Context) (bool, error) {
	return s.client.Kill(ctx, s.Info.ID)
}

// WaitOptions are the options for WaitUntilRunning.
type WaitOptions struct {
	// Timeout defaults to 2 minutes. A shorter ctx deadline wins.
	Timeout time.Duration
	// PollInterval defaults to 1 second.
	PollInterval time.Duration
}

// WaitUntilRunning polls until the sandbox is running. It fails if the sandbox
// is paused or the wait runs out.
func (s *Sandbox) WaitUntilRunning(ctx context.Context, opts *WaitOptions) error {
	timeout, interval := 2*time.Minute, time.Second
	if opts != nil {
		if opts.Timeout > 0 {
			timeout = opts.Timeout
		}
		if opts.PollInterval > 0 {
			interval = opts.PollInterval
		}
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	for {
		switch s.Info.State {
		case godo.SandboxStateRunning:
			return nil
		case godo.SandboxStateCreating:
		default:
			return fmt.Errorf("sandbox: %s will not become running: %s", s.Info.ID, s.Info.State)
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("sandbox: %s was not running in time (last state %s): %w", s.Info.ID, s.Info.State, ctx.Err())
		case <-time.After(interval):
		}
		if err := s.Refresh(ctx); err != nil {
			if ctx.Err() != nil {
				return fmt.Errorf("sandbox: %s was not running in time (last state %s): %w", s.Info.ID, s.Info.State, ctx.Err())
			}
			return err
		}
	}
}
