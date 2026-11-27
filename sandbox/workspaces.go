package sandbox

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/digitalocean/godo"
)

// MaxWorkspaceSizeGiB is the largest size the API accepts for a workspace.
const MaxWorkspaceSizeGiB = 100

// WorkspaceParams are the options for CreateWorkspace.
type WorkspaceParams struct {
	// Name is an optional label for the workspace.
	Name string
	// SizeLimitGiB is the size limit in GiB. Required; 1 to MaxWorkspaceSizeGiB.
	SizeLimitGiB int64
}

// ListWorkspacesParams are the options for ListWorkspaces. Every field is optional.
type ListWorkspacesParams struct {
	// State returns only workspaces in this state. Empty returns any state.
	State godo.SandboxWorkspaceState
	// Limit is the page size.
	Limit int
	// NextToken resumes after the NextToken of an earlier Pages.
	NextToken string
}

// Workspace is a persistent disk that outlives any one sandbox. Create one with
// Client.CreateWorkspace and wait for it with WaitUntilAvailable; then pass its
// ID as CreateParams.WorkspaceID to Client.Create.
type Workspace struct {
	client *Client

	// Info is the record as last returned by the API. Refresh updates it.
	Info godo.SandboxWorkspace
}

// ID is the workspace ID, which is what CreateParams.WorkspaceID takes.
func (w *Workspace) ID() string { return w.Info.ID }

// Name is the workspace name. Empty when none was given.
func (w *Workspace) Name() string { return w.Info.Name }

// State is the state as last returned by the API. It can be stale: Refresh
// reads it again.
func (w *Workspace) State() godo.SandboxWorkspaceState { return w.Info.State }

// CreateWorkspace creates a persistent workspace. It returns as soon as the API
// answers, when the workspace may still be creating: call WaitUntilAvailable
// before attaching it.
func (c *Client) CreateWorkspace(ctx context.Context, params *WorkspaceParams) (*Workspace, error) {
	if params == nil {
		return nil, errors.New("sandbox: WorkspaceParams is required")
	}
	if params.SizeLimitGiB < 1 || params.SizeLimitGiB > MaxWorkspaceSizeGiB {
		return nil, fmt.Errorf("sandbox: workspace SizeLimitGiB must be 1 to %d", MaxWorkspaceSizeGiB)
	}
	info, _, err := c.API.Sandboxes.CreateWorkspace(ctx, &godo.SandboxWorkspaceCreateRequest{
		Name:         params.Name,
		SizeLimitGiB: params.SizeLimitGiB,
	})
	if err != nil {
		return nil, err
	}
	return &Workspace{client: c, Info: *info}, nil
}

// GetWorkspace fetches a workspace by ID.
func (c *Client) GetWorkspace(ctx context.Context, id string) (*Workspace, error) {
	info, _, err := c.API.Sandboxes.GetWorkspace(ctx, id)
	if err != nil {
		return nil, err
	}
	return &Workspace{client: c, Info: *info}, nil
}

// DeleteWorkspace deletes a workspace by ID. It reports false, with no error,
// when the workspace does not exist.
func (c *Client) DeleteWorkspace(ctx context.Context, id string) (bool, error) {
	if _, err := c.API.Sandboxes.DeleteWorkspace(ctx, id); err != nil {
		if IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// ListWorkspaces lists your team's workspaces. Nothing is fetched until you
// read a page.
func (c *Client) ListWorkspaces(params *ListWorkspacesParams) *Pages[*Workspace] {
	if params == nil {
		params = &ListWorkspacesParams{}
	}
	state, limit := params.State, params.Limit
	return &Pages[*Workspace]{
		token: params.NextToken,
		fetch: func(ctx context.Context, token string) ([]*Workspace, string, error) {
			list, _, err := c.API.Sandboxes.ListWorkspaces(ctx, &godo.SandboxWorkspaceListOptions{State: state, PageSize: limit, PageToken: token})
			if err != nil {
				return nil, "", err
			}
			out := make([]*Workspace, len(list.Workspaces))
			for i := range list.Workspaces {
				out[i] = &Workspace{client: c, Info: list.Workspaces[i]}
			}
			return out, list.NextPageToken, nil
		},
	}
}

// Refresh reads the workspace again.
func (w *Workspace) Refresh(ctx context.Context) error {
	info, _, err := w.client.API.Sandboxes.GetWorkspace(ctx, w.Info.ID)
	if err != nil {
		return err
	}
	w.Info = *info
	return nil
}

// Delete deletes the workspace and its data. It reports false, with no error,
// when it is already gone.
func (w *Workspace) Delete(ctx context.Context) (bool, error) {
	return w.client.DeleteWorkspace(ctx, w.Info.ID)
}

// WaitUntilAvailable polls until the workspace can be attached to a sandbox. It
// fails if the workspace failed or is being deleted, and when the wait runs
// out. Timeout defaults to 5 minutes and PollInterval to 2 seconds; a shorter
// ctx deadline wins.
func (w *Workspace) WaitUntilAvailable(ctx context.Context, opts *WaitOptions) error {
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
		switch w.Info.State {
		case godo.SandboxWorkspaceStateAvailable:
			return nil
		case godo.SandboxWorkspaceStateFailed, godo.SandboxWorkspaceStateDeleting:
			return fmt.Errorf("sandbox: workspace %s will not become available (state %s)", w.Info.ID, w.Info.State)
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("sandbox: workspace %s was not available in time (last state %s): %w", w.Info.ID, w.Info.State, ctx.Err())
		case <-time.After(interval):
		}
		if err := w.Refresh(ctx); err != nil {
			if ctx.Err() != nil {
				return fmt.Errorf("sandbox: workspace %s was not available in time (last state %s): %w", w.Info.ID, w.Info.State, ctx.Err())
			}
			return err
		}
	}
}
