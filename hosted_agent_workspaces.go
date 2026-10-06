package godo

import (
	"context"
	"errors"
	"fmt"
	"net/http"
)

// hostedAgentIdempotencyKeyHeader carries the optional client-chosen key that
// makes a retried CreateWorkspace safe.
const hostedAgentIdempotencyKeyHeader = "Idempotency-Key"

// HostedAgentWorkspaceState is the lifecycle state of a persistent workspace.
type HostedAgentWorkspaceState string

const (
	// HostedAgentWorkspaceStateAvailable means no session holds the workspace.
	HostedAgentWorkspaceStateAvailable HostedAgentWorkspaceState = "AVAILABLE"
	// HostedAgentWorkspaceStateAttaching means a session is claiming the workspace.
	HostedAgentWorkspaceStateAttaching HostedAgentWorkspaceState = "ATTACHING"
	// HostedAgentWorkspaceStateAttached means a session holds the workspace.
	HostedAgentWorkspaceStateAttached HostedAgentWorkspaceState = "ATTACHED"
	// HostedAgentWorkspaceStateReleasing means the workspace is being released
	// by the session that held it.
	HostedAgentWorkspaceStateReleasing HostedAgentWorkspaceState = "RELEASING"
	// HostedAgentWorkspaceStateFailed means the workspace is not usable.
	HostedAgentWorkspaceStateFailed HostedAgentWorkspaceState = "FAILED"
)

// HostedAgentWorkspace is a persistent workspace: a separate disk that outlives
// sessions and is attached to one session at a time. It is not the session's
// /workspace file transfer API (see HostedAgentWorkspaceTransfer and
// UploadWorkspace / DownloadWorkspace), which moves files in and out of a
// running session.
type HostedAgentWorkspace struct {
	// WorkspaceID is the opaque workspace identifier.
	WorkspaceID string `json:"workspace_id"`
	// Name is an optional label. It is not unique.
	Name  string                    `json:"name,omitempty"`
	State HostedAgentWorkspaceState `json:"state"`
	// AttachedSessionID is the session that holds the workspace; omitted when no
	// session does.
	AttachedSessionID string `json:"attached_session_id,omitempty"`
	SizeGibibytes     int32  `json:"size_gibibytes"`
	BytesUsed         int64  `json:"bytes_used"`
	// LastSavedAt is when the workspace contents were last saved; omitted until
	// the first save.
	LastSavedAt *Timestamp `json:"last_saved_at,omitempty"`
	CreatedAt   Timestamp  `json:"created_at"`
	UpdatedAt   Timestamp  `json:"updated_at"`
}

// HostedAgentWorkspaceCreateRequest is the body for CreateWorkspace.
type HostedAgentWorkspaceCreateRequest struct {
	// SizeGibibytes is the disk size, from 1 to 100. The server validates it, and
	// a team limit can lower the maximum.
	SizeGibibytes int32 `json:"size_gibibytes"`
	// Name is an optional label. It is not unique.
	Name string `json:"name,omitempty"`
	// IdempotencyKey makes a retried create safe. It is sent as the
	// Idempotency-Key header, not in the body. Within 24 hours the same key and
	// body return the first workspace, while the same key with a different body
	// is rejected (HTTP 422). Keys are scoped to the team. Omitted when empty.
	IdempotencyKey string `json:"-"`
}

// HostedAgentWorkspaceListOptions paginates ListWorkspaces (newest-first).
type HostedAgentWorkspaceListOptions struct {
	PageToken string `url:"page_token,omitempty"`
	// PageSize defaults to 50 on the server and is clamped to 200.
	PageSize int `url:"page_size,omitempty"`
}

// HostedAgentWorkspacesListResponse is returned by ListWorkspaces.
type HostedAgentWorkspacesListResponse struct {
	Workspaces []HostedAgentWorkspace `json:"workspaces"`
	// NextPageToken is empty when there are no more results.
	NextPageToken string `json:"next_page_token"`
}

type hostedAgentWorkspaceRoot struct {
	Workspace *HostedAgentWorkspace `json:"workspace"`
}

// CreateWorkspace creates a persistent workspace. The API answers 201 for a new
// workspace and 200 when an IdempotencyKey replays an earlier create; both are
// returned as success.
func (s *HostedAgentsServiceOp) CreateWorkspace(ctx context.Context, create *HostedAgentWorkspaceCreateRequest) (*HostedAgentWorkspace, *Response, error) {
	if create == nil {
		return nil, nil, errors.New("hosted agents: create request is required")
	}
	req, err := s.client.NewRequest(ctx, http.MethodPost, hostedAgentWorkspacesBasePath, create)
	if err != nil {
		return nil, nil, err
	}
	if create.IdempotencyKey != "" {
		req.Header.Set(hostedAgentIdempotencyKeyHeader, create.IdempotencyKey)
	}
	root := new(hostedAgentWorkspaceRoot)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}
	if root.Workspace == nil {
		return nil, resp, errors.New("hosted agents: create workspace returned no workspace")
	}
	return root.Workspace, resp, nil
}

// ListWorkspaces returns the team's persistent workspaces, newest first.
func (s *HostedAgentsServiceOp) ListWorkspaces(ctx context.Context, opt *HostedAgentWorkspaceListOptions) (*HostedAgentWorkspacesListResponse, *Response, error) {
	path, err := addOptions(hostedAgentWorkspacesBasePath, opt)
	if err != nil {
		return nil, nil, err
	}
	req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}
	root := new(HostedAgentWorkspacesListResponse)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}
	return root, resp, nil
}

// GetWorkspace returns a single persistent workspace by ID.
func (s *HostedAgentsServiceOp) GetWorkspace(ctx context.Context, workspaceID string) (*HostedAgentWorkspace, *Response, error) {
	if workspaceID == "" {
		return nil, nil, errors.New("hosted agents: workspace id is required")
	}
	path := fmt.Sprintf(hostedAgentWorkspaceByIDPath, workspaceID)
	req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}
	root := new(hostedAgentWorkspaceRoot)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}
	if root.Workspace == nil {
		return nil, resp, errors.New("hosted agents: get workspace returned no workspace")
	}
	return root.Workspace, resp, nil
}

// DeleteWorkspace deletes a persistent workspace. The API returns HTTP 204, or
// 409 when a session is still using the workspace.
func (s *HostedAgentsServiceOp) DeleteWorkspace(ctx context.Context, workspaceID string) (*Response, error) {
	if workspaceID == "" {
		return nil, errors.New("hosted agents: workspace id is required")
	}
	path := fmt.Sprintf(hostedAgentWorkspaceByIDPath, workspaceID)
	req, err := s.client.NewRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return nil, err
	}
	return s.client.Do(ctx, req, nil)
}
