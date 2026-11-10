package godo

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
)

// DataCloudTarballUpload is an upload grant returned when a named tarball object is absent.
type DataCloudTarballUpload struct {
	URL       string `json:"url,omitempty"`
	ObjectKey string `json:"object_key,omitempty"`
	ExpiresAt string `json:"expires_at,omitempty"`
}

// DataCloudTarballStatus is the tarball digests on a data flow.
type DataCloudTarballStatus struct {
	Requested string `json:"requested,omitempty"`
	Attached  string `json:"attached,omitempty"`
}

// DataCloudDataFlow is one data flow resource.
type DataCloudDataFlow struct {
	UUID            string                  `json:"uuid,omitempty"`
	Name            string                  `json:"name,omitempty"`
	State           string                  `json:"state,omitempty"`
	Workspace       string                  `json:"workspace,omitempty"`
	FchainID        string                  `json:"fchain_id,omitempty"`
	ConfigJSON      string                  `json:"config_json,omitempty"`
	DesiredReplicas int32                   `json:"desired_replicas,omitempty"`
	MaxReplicas     int32                   `json:"max_replicas,omitempty"`
	CreatedAt       string                  `json:"created_at,omitempty"`
	UpdatedAt       string                  `json:"updated_at,omitempty"`
	ImageRef        string                  `json:"image_ref,omitempty"`
	Spec            json.RawMessage         `json:"spec,omitempty"`
	Tarball         *DataCloudTarballStatus `json:"tarball,omitempty"`
	DesiredState    string                  `json:"desired_state,omitempty"`
	LastError       string                  `json:"last_error,omitempty"`
}

// DataCloudDataFlowCreateRequest is the body for CreateDataFlow.
// Workspace is required and used as the path segment.
type DataCloudDataFlowCreateRequest struct {
	Name            string          `json:"name,omitempty"`
	DesiredReplicas int32           `json:"desired_replicas,omitempty"`
	MaxReplicas     int32           `json:"max_replicas,omitempty"`
	Workspace       string          `json:"workspace"`
	ImageRef        string          `json:"image_ref,omitempty"`
	Manifest        json.RawMessage `json:"manifest,omitempty"`
}

// DataCloudDataFlowCreateResponse is the body of CreateDataFlow.
type DataCloudDataFlowCreateResponse struct {
	UUID   string                  `json:"uuid,omitempty"`
	State  string                  `json:"state,omitempty"`
	Upload *DataCloudTarballUpload `json:"upload,omitempty"`
}

// DataCloudDataFlowListRequest holds path and query args for ListDataFlows.
type DataCloudDataFlowListRequest struct {
	Workspace string `url:"-"`
	PageSize  int    `url:"page_size,omitempty"`
	PageToken string `url:"page_token,omitempty"`
}

// DataCloudListDataFlowsResponse is the body of ListDataFlows.
type DataCloudListDataFlowsResponse struct {
	DataFlows     []DataCloudDataFlow `json:"data_flows"`
	NextPageToken string              `json:"next_page_token,omitempty"`
}

// DataCloudDataFlowUpdateRequest is the body for UpdateDataFlow.
// Workspace and ID are required path segments and are also sent in the body.
type DataCloudDataFlowUpdateRequest struct {
	ID              string          `json:"id"`
	Workspace       string          `json:"workspace"`
	DesiredReplicas int32           `json:"desired_replicas,omitempty"`
	MaxReplicas     int32           `json:"max_replicas,omitempty"`
	SecretRefs      []string        `json:"secret_refs,omitempty"`
	ClearSecretRefs bool            `json:"clear_secret_refs,omitempty"`
	Manifest        json.RawMessage `json:"manifest,omitempty"`
}

// DataCloudDataFlowUpdateResponse is the body of UpdateDataFlow.
type DataCloudDataFlowUpdateResponse struct {
	UUID   string                  `json:"uuid,omitempty"`
	State  string                  `json:"state,omitempty"`
	Upload *DataCloudTarballUpload `json:"upload,omitempty"`
}

// DataCloudDataFlowRewindRequest is the body for RewindDataFlow.
type DataCloudDataFlowRewindRequest struct {
	ID        string `json:"id"`
	Workspace string `json:"workspace"`
	RewindTo  string `json:"rewind_to,omitempty"`
}

// DataCloudFchainUploadURLRequest is the body for CreateFchainUploadURL.
type DataCloudFchainUploadURLRequest struct {
	SHA256    string `json:"sha256,omitempty"`
	Workspace string `json:"workspace"`
}

// DataCloudFchainUploadURLResponse is the body of CreateFchainUploadURL.
type DataCloudFchainUploadURLResponse struct {
	UploadURL string `json:"upload_url,omitempty"`
	ObjectKey string `json:"object_key,omitempty"`
	ExpiresAt string `json:"expires_at,omitempty"`
}

// DataCloudAttachDataFlowFchainRequest is the body for AttachDataFlowFchain.
type DataCloudAttachDataFlowFchainRequest struct {
	ID        string `json:"id"`
	Workspace string `json:"workspace"`
	ObjectKey string `json:"object_key,omitempty"`
}

// DataCloudWorkspaceMetricsRequest is the body for GetWorkspaceMetrics.
type DataCloudWorkspaceMetricsRequest struct {
	Workspace   string   `json:"workspace"`
	MetricNames []string `json:"metric_names,omitempty"`
	DataFlowID  string   `json:"data_flow_id,omitempty"`
	From        string   `json:"from,omitempty"`
	To          string   `json:"to,omitempty"`
	GroupBy     []string `json:"group_by,omitempty"`
	Limit       int32    `json:"limit,omitempty"`
	MetricType  string   `json:"metric_type,omitempty"`
}

// DataCloudWorkspaceMetricsResponse is the body of GetWorkspaceMetrics.
type DataCloudWorkspaceMetricsResponse struct {
	Count   int32            `json:"count,omitempty"`
	Columns []string         `json:"columns,omitempty"`
	Rows    []map[string]any `json:"rows,omitempty"`
}

func dataCloudDataFlowsPath(workspace string) (string, error) {
	if workspace == "" {
		return "", errors.New("data cloud platform: workspace is required")
	}
	return dataCloudPlatformPrefix + "/workspaces/" + url.PathEscape(workspace) + "/data-flows", nil
}

func dataCloudDataFlowPath(workspace, id string) (string, error) {
	if id == "" {
		return "", errors.New("data cloud platform: data flow id is required")
	}
	base, err := dataCloudDataFlowsPath(workspace)
	if err != nil {
		return "", err
	}
	return base + "/" + url.PathEscape(id), nil
}

func dataCloudDataFlowActionPath(workspace, id, action string) (string, error) {
	base, err := dataCloudDataFlowPath(workspace, id)
	if err != nil {
		return "", err
	}
	return base + "/" + action, nil
}

func dataCloudFchainUploadURLPath(workspace string) (string, error) {
	if workspace == "" {
		return "", errors.New("data cloud platform: workspace is required")
	}
	return dataCloudPlatformPrefix + "/workspaces/" + url.PathEscape(workspace) + "/fchain/upload-url", nil
}

func dataCloudWorkspaceMetricsPath(workspace string) (string, error) {
	if workspace == "" {
		return "", errors.New("data cloud platform: workspace is required")
	}
	return dataCloudPlatformPrefix + "/workspaces/" + url.PathEscape(workspace) + "/metrics", nil
}

// CreateDataFlow creates a data flow in the workspace on create.
func (s *DataCloudPlatformServiceOp) CreateDataFlow(ctx context.Context, create *DataCloudDataFlowCreateRequest) (*DataCloudDataFlowCreateResponse, *Response, error) {
	if create == nil {
		return nil, nil, errors.New("data cloud platform: create request is required")
	}
	path, err := dataCloudDataFlowsPath(create.Workspace)
	if err != nil {
		return nil, nil, err
	}
	req, err := s.client.NewRequest(ctx, http.MethodPost, path, create)
	if err != nil {
		return nil, nil, err
	}
	root := new(DataCloudDataFlowCreateResponse)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}
	return root, resp, nil
}

// ListDataFlows lists data flows in the workspace on list.
func (s *DataCloudPlatformServiceOp) ListDataFlows(ctx context.Context, list *DataCloudDataFlowListRequest) (*DataCloudListDataFlowsResponse, *Response, error) {
	if list == nil {
		return nil, nil, errors.New("data cloud platform: list request is required")
	}
	path, err := dataCloudDataFlowsPath(list.Workspace)
	if err != nil {
		return nil, nil, err
	}
	path, err = addOptions(path, list)
	if err != nil {
		return nil, nil, err
	}
	req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}
	root := new(DataCloudListDataFlowsResponse)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}
	return root, resp, nil
}

// GetDataFlow returns one data flow by workspace and id.
func (s *DataCloudPlatformServiceOp) GetDataFlow(ctx context.Context, workspace, id string) (*DataCloudDataFlow, *Response, error) {
	path, err := dataCloudDataFlowPath(workspace, id)
	if err != nil {
		return nil, nil, err
	}
	req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}
	root := new(DataCloudDataFlow)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}
	return root, resp, nil
}

// UpdateDataFlow patches a data flow.
func (s *DataCloudPlatformServiceOp) UpdateDataFlow(ctx context.Context, update *DataCloudDataFlowUpdateRequest) (*DataCloudDataFlowUpdateResponse, *Response, error) {
	if update == nil {
		return nil, nil, errors.New("data cloud platform: update request is required")
	}
	path, err := dataCloudDataFlowPath(update.Workspace, update.ID)
	if err != nil {
		return nil, nil, err
	}
	req, err := s.client.NewRequest(ctx, http.MethodPatch, path, update)
	if err != nil {
		return nil, nil, err
	}
	root := new(DataCloudDataFlowUpdateResponse)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}
	return root, resp, nil
}

// DeleteDataFlow begins deleting a data flow.
func (s *DataCloudPlatformServiceOp) DeleteDataFlow(ctx context.Context, workspace, id string) (*Response, error) {
	path, err := dataCloudDataFlowPath(workspace, id)
	if err != nil {
		return nil, err
	}
	req, err := s.client.NewRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return nil, err
	}
	return s.client.Do(ctx, req, nil)
}

// PauseDataFlow pauses an active data flow.
func (s *DataCloudPlatformServiceOp) PauseDataFlow(ctx context.Context, workspace, id string) (*Response, error) {
	path, err := dataCloudDataFlowActionPath(workspace, id, "pause")
	if err != nil {
		return nil, err
	}
	req, err := s.client.NewRequest(ctx, http.MethodPost, path, nil)
	if err != nil {
		return nil, err
	}
	return s.client.Do(ctx, req, nil)
}

// ResumeDataFlow resumes a paused data flow.
func (s *DataCloudPlatformServiceOp) ResumeDataFlow(ctx context.Context, workspace, id string) (*Response, error) {
	path, err := dataCloudDataFlowActionPath(workspace, id, "resume")
	if err != nil {
		return nil, err
	}
	req, err := s.client.NewRequest(ctx, http.MethodPost, path, nil)
	if err != nil {
		return nil, err
	}
	return s.client.Do(ctx, req, nil)
}

// RestartDataFlow restarts a data flow.
func (s *DataCloudPlatformServiceOp) RestartDataFlow(ctx context.Context, workspace, id string) (*Response, error) {
	path, err := dataCloudDataFlowActionPath(workspace, id, "restart")
	if err != nil {
		return nil, err
	}
	req, err := s.client.NewRequest(ctx, http.MethodPost, path, nil)
	if err != nil {
		return nil, err
	}
	return s.client.Do(ctx, req, nil)
}

// RewindDataFlow requests rewind. The server may respond that rewind is unsupported.
func (s *DataCloudPlatformServiceOp) RewindDataFlow(ctx context.Context, rewind *DataCloudDataFlowRewindRequest) (*Response, error) {
	if rewind == nil {
		return nil, errors.New("data cloud platform: rewind request is required")
	}
	path, err := dataCloudDataFlowActionPath(rewind.Workspace, rewind.ID, "rewind")
	if err != nil {
		return nil, err
	}
	req, err := s.client.NewRequest(ctx, http.MethodPost, path, rewind)
	if err != nil {
		return nil, err
	}
	return s.client.Do(ctx, req, nil)
}

// CreateFchainUploadURL creates a temporary upload URL for a function-chain archive.
func (s *DataCloudPlatformServiceOp) CreateFchainUploadURL(ctx context.Context, create *DataCloudFchainUploadURLRequest) (*DataCloudFchainUploadURLResponse, *Response, error) {
	if create == nil {
		return nil, nil, errors.New("data cloud platform: create request is required")
	}
	path, err := dataCloudFchainUploadURLPath(create.Workspace)
	if err != nil {
		return nil, nil, err
	}
	req, err := s.client.NewRequest(ctx, http.MethodPost, path, create)
	if err != nil {
		return nil, nil, err
	}
	root := new(DataCloudFchainUploadURLResponse)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}
	return root, resp, nil
}

// AttachDataFlowFchain attaches an uploaded function-chain archive to a data flow.
func (s *DataCloudPlatformServiceOp) AttachDataFlowFchain(ctx context.Context, attach *DataCloudAttachDataFlowFchainRequest) (*Response, error) {
	if attach == nil {
		return nil, errors.New("data cloud platform: attach request is required")
	}
	path, err := dataCloudDataFlowActionPath(attach.Workspace, attach.ID, "fchain")
	if err != nil {
		return nil, err
	}
	req, err := s.client.NewRequest(ctx, http.MethodPost, path, attach)
	if err != nil {
		return nil, err
	}
	return s.client.Do(ctx, req, nil)
}

// GetWorkspaceMetrics queries data flow metrics for a workspace.
func (s *DataCloudPlatformServiceOp) GetWorkspaceMetrics(ctx context.Context, metrics *DataCloudWorkspaceMetricsRequest) (*DataCloudWorkspaceMetricsResponse, *Response, error) {
	if metrics == nil {
		return nil, nil, errors.New("data cloud platform: metrics request is required")
	}
	path, err := dataCloudWorkspaceMetricsPath(metrics.Workspace)
	if err != nil {
		return nil, nil, err
	}
	req, err := s.client.NewRequest(ctx, http.MethodPost, path, metrics)
	if err != nil {
		return nil, nil, err
	}
	root := new(DataCloudWorkspaceMetricsResponse)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}
	return root, resp, nil
}
