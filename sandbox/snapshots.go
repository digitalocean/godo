package sandbox

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/digitalocean/godo"
)

var snapshotNameRE = regexp.MustCompile(`^[a-z0-9-]{1,63}$`)

// SnapshotParams are the options for Sandbox.CreateSnapshot. Every field is
// optional.
type SnapshotParams struct {
	// Name is 1 to 63 characters: lowercase letters, digits and hyphens.
	Name string
	// TTL is how long to keep the snapshot, as a Go duration such as "168h".
	// When empty, the platform default applies.
	TTL string
}

// ListSnapshotsParams are the options for ListSnapshots. Every field is optional.
type ListSnapshotsParams struct {
	// SandboxID lists only the snapshots of this sandbox. When empty, every
	// snapshot of your team is listed.
	SandboxID string
	// Limit is the page size.
	Limit int
	// NextToken resumes after the NextToken of an earlier Pages.
	NextToken string
}

// Snapshot is a snapshot of a sandbox. Create one with Sandbox.CreateSnapshot
// and wait for it with WaitUntilReady.
type Snapshot struct {
	client *Client

	// Info is the record as last returned by the API. Refresh updates it.
	Info godo.SandboxSnapshot
}

// ID is the snapshot ID.
func (s *Snapshot) ID() string { return s.Info.ID }

// SandboxID is the ID of the sandbox the snapshot was taken from.
func (s *Snapshot) SandboxID() string { return s.Info.SandboxID }

// Name is the name given at creation, or empty.
func (s *Snapshot) Name() string { return s.Info.Name }

// State is the state as last returned by the API. It can be stale: Refresh
// reads it again.
func (s *Snapshot) State() godo.SandboxSnapshotState { return s.Info.State }

// CreateSnapshot starts a snapshot of the sandbox, which must be running or
// paused. It returns as soon as the capture starts: call WaitUntilReady before
// relying on the snapshot. A nil params uses the platform defaults.
func (s *Sandbox) CreateSnapshot(ctx context.Context, params *SnapshotParams) (*Snapshot, error) {
	req := &godo.SandboxSnapshotCreateRequest{}
	if params != nil {
		if params.Name != "" && !snapshotNameRE.MatchString(params.Name) {
			return nil, fmt.Errorf("sandbox: snapshot Name %q must be 1 to 63 characters: lowercase letters, digits and hyphens", params.Name)
		}
		if params.TTL != "" {
			d, err := time.ParseDuration(params.TTL)
			if err != nil || d <= 0 {
				return nil, fmt.Errorf("sandbox: snapshot TTL %q must be a positive duration such as \"168h\"", params.TTL)
			}
		}
		req.Name, req.TTL = params.Name, params.TTL
	}
	info, _, err := s.client.API.Sandboxes.CreateSnapshot(ctx, s.Info.ID, req)
	if err != nil {
		return nil, err
	}
	return &Snapshot{client: s.client, Info: *info}, nil
}

// GetSnapshot fetches a snapshot by ID.
func (c *Client) GetSnapshot(ctx context.Context, id string) (*Snapshot, error) {
	info, _, err := c.API.Sandboxes.GetSnapshot(ctx, id)
	if err != nil {
		return nil, err
	}
	return &Snapshot{client: c, Info: *info}, nil
}

// DeleteSnapshot deletes a snapshot by ID. It reports false, with no error,
// when the snapshot does not exist.
func (c *Client) DeleteSnapshot(ctx context.Context, id string) (bool, error) {
	if _, err := c.API.Sandboxes.DeleteSnapshot(ctx, id); err != nil {
		if IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// ListSnapshots lists your team's snapshots, newest first. Nothing is fetched
// until you read a page.
func (c *Client) ListSnapshots(params *ListSnapshotsParams) *Pages[*Snapshot] {
	if params == nil {
		params = &ListSnapshotsParams{}
	}
	p := *params
	return &Pages[*Snapshot]{
		token: p.NextToken,
		fetch: func(ctx context.Context, token string) ([]*Snapshot, string, error) {
			list, _, err := c.API.Sandboxes.ListSnapshots(ctx, &godo.SandboxSnapshotListOptions{
				SandboxID: p.SandboxID, PageSize: p.Limit, PageToken: token,
			})
			if err != nil {
				return nil, "", err
			}
			out := make([]*Snapshot, len(list.Snapshots))
			for i := range list.Snapshots {
				out[i] = &Snapshot{client: c, Info: list.Snapshots[i]}
			}
			return out, list.NextPageToken, nil
		},
	}
}

// Refresh reads the snapshot again.
func (s *Snapshot) Refresh(ctx context.Context) error {
	info, _, err := s.client.API.Sandboxes.GetSnapshot(ctx, s.Info.ID)
	if err != nil {
		return err
	}
	s.Info = *info
	return nil
}

// Delete deletes the snapshot. It reports false, with no error, when it is
// already gone.
func (s *Snapshot) Delete(ctx context.Context) (bool, error) {
	return s.client.DeleteSnapshot(ctx, s.Info.ID)
}

// WaitUntilReady polls until the snapshot is complete. It fails, with the
// API's error message when there is one, if the snapshot failed, and when the
// wait runs out. Timeout defaults to 5 minutes and PollInterval to 2 seconds;
// a shorter ctx deadline wins.
func (s *Snapshot) WaitUntilReady(ctx context.Context, opts *WaitOptions) error {
	timeout, interval := 5*time.Minute, 2*time.Second
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
		case godo.SandboxSnapshotStateComplete:
			return nil
		case godo.SandboxSnapshotStateFailed:
			msg := fmt.Sprintf("sandbox: snapshot %s failed", s.Info.ID)
			if m := strings.TrimSpace(s.Info.ErrorMessage); m != "" {
				return fmt.Errorf("%s: %s", msg, m)
			}
			return errors.New(msg)
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("sandbox: snapshot %s was not ready in time (last state %s): %w", s.Info.ID, s.Info.State, ctx.Err())
		case <-time.After(interval):
		}
		if err := s.Refresh(ctx); err != nil {
			if ctx.Err() != nil {
				return fmt.Errorf("sandbox: snapshot %s was not ready in time (last state %s): %w", s.Info.ID, s.Info.State, ctx.Err())
			}
			return err
		}
	}
}
