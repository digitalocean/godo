package godo

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
)

const dataCloudPlatformPrefix = "/v2/data-cloud-platform"

// DataCloudPlatformService is the Agent Data Services client for workspace,
// token, secret, query-credential, and data-flow routes under
// /v2/data-cloud-platform.
type DataCloudPlatformService interface {
	// ListPlans returns the regions and query engine sizes a team can use
	// when creating a workspace. The list is not paginated.
	ListPlans(context.Context) (*DataCloudPlansResponse, *Response, error)
	// CreateWorkspace creates a workspace in region. Region is the path
	// segment. The body carries name and query_instance_size only.
	CreateWorkspace(context.Context, string, *DataCloudWorkspaceCreateRequest) (*DataCloudWorkspace, *Response, error)
	// ListWorkspaces returns the workspaces in region. The list is not paginated.
	ListWorkspaces(context.Context, string) (*DataCloudWorkspacesResponse, *Response, error)
	// GetWorkspace returns one workspace by region and uuid.
	GetWorkspace(context.Context, string, string) (*DataCloudWorkspace, *Response, error)
	// DeleteWorkspace deletes one workspace. The API returns HTTP 200 and an
	// empty object.
	DeleteWorkspace(context.Context, string, string) (*Response, error)
	// CreateToken creates a token in the region carried on the request body.
	// Region is also the path segment.
	CreateToken(context.Context, *DataCloudTokenCreateRequest) (*DataCloudTokenCreateResponse, *Response, error)
	// ListTokens returns tokens in the region carried on the request.
	// Region is the path segment. Optional filters are query parameters.
	ListTokens(context.Context, *DataCloudTokenListRequest) (*DataCloudListTokensResponse, *Response, error)
	// PutWorkspaceSecret writes a workspace secret. Region, UUID, and SecretKey
	// on the request are path segments and are also sent in the body. Data must
	// contain at least one string key-value pair.
	PutWorkspaceSecret(context.Context, *DataCloudPutWorkspaceSecretRequest) (*DataCloudPutWorkspaceSecretResponse, *Response, error)
	// GetWorkspaceSecret returns secret metadata and property names. Values are never returned.
	GetWorkspaceSecret(context.Context, string, string, string) (*DataCloudGetWorkspaceSecretResponse, *Response, error)
	// ListWorkspaceSecrets returns secret key names for a workspace. Values are never returned.
	ListWorkspaceSecrets(context.Context, string, string) (*DataCloudListWorkspaceSecretsResponse, *Response, error)
	// DeleteWorkspaceSecret permanently deletes a workspace secret.
	DeleteWorkspaceSecret(context.Context, string, string, string) (*Response, error)
	// GetWorkspaceQueryCredentials returns the query user and password for a workspace.
	// Do not log the response; it contains a password.
	GetWorkspaceQueryCredentials(context.Context, string, string) (*DataCloudQueryCredentials, *Response, error)

	// CreateDataFlow creates a data flow in a workspace.
	CreateDataFlow(context.Context, *DataCloudDataFlowCreateRequest) (*DataCloudDataFlowCreateResponse, *Response, error)
	// ListDataFlows lists data flows in a workspace.
	ListDataFlows(context.Context, *DataCloudDataFlowListRequest) (*DataCloudListDataFlowsResponse, *Response, error)
	// GetDataFlow returns one data flow by workspace and id.
	GetDataFlow(context.Context, string, string) (*DataCloudDataFlow, *Response, error)
	// UpdateDataFlow patches a data flow's replica settings, secret refs, or manifest.
	UpdateDataFlow(context.Context, *DataCloudDataFlowUpdateRequest) (*DataCloudDataFlowUpdateResponse, *Response, error)
	// DeleteDataFlow begins deleting a data flow.
	DeleteDataFlow(context.Context, string, string) (*Response, error)
	// PauseDataFlow pauses an active data flow.
	PauseDataFlow(context.Context, string, string) (*Response, error)
	// ResumeDataFlow resumes a paused data flow.
	ResumeDataFlow(context.Context, string, string) (*Response, error)
	// RestartDataFlow restarts a data flow.
	RestartDataFlow(context.Context, string, string) (*Response, error)
	// RewindDataFlow requests rewind to an earlier time. The server currently
	// reports rewind as unsupported; still call the endpoint and surface that response.
	RewindDataFlow(context.Context, *DataCloudDataFlowRewindRequest) (*Response, error)
	// CreateFchainUploadURL creates a temporary upload URL for a function-chain archive.
	CreateFchainUploadURL(context.Context, *DataCloudFchainUploadURLRequest) (*DataCloudFchainUploadURLResponse, *Response, error)
	// AttachDataFlowFchain attaches an uploaded function-chain archive to a data flow.
	AttachDataFlowFchain(context.Context, *DataCloudAttachDataFlowFchainRequest) (*Response, error)
	// GetWorkspaceMetrics queries data flow metrics for a workspace.
	GetWorkspaceMetrics(context.Context, *DataCloudWorkspaceMetricsRequest) (*DataCloudWorkspaceMetricsResponse, *Response, error)
}

// DataCloudPlatformServiceOp handles communication with Agent Data Services
// workspace and plan methods.
type DataCloudPlatformServiceOp struct {
	client *Client
}

var _ DataCloudPlatformService = &DataCloudPlatformServiceOp{}

// DataCloudQueryEngineSize is one selectable query engine size inside a plan.
// Size is a plain string. Advertised values are small, medium, and large.
type DataCloudQueryEngineSize struct {
	Size        string `json:"size,omitempty"`
	DropletSlug string `json:"droplet_slug,omitempty"`
	CPU         int32  `json:"cpu,omitempty"`
	MemoryGi    int32  `json:"memory_gi,omitempty"`
}

// DataCloudPlan is a region offering returned by ListPlans.
type DataCloudPlan struct {
	RegionSlug         string                     `json:"region_slug,omitempty"`
	AvailableForCreate bool                       `json:"available_for_create"`
	QueryEngineSizes   []DataCloudQueryEngineSize `json:"query_engine_sizes,omitempty"`
}

// DataCloudPlansResponse is the body of GET /v2/data-cloud-platform/plans.
type DataCloudPlansResponse struct {
	Plans []DataCloudPlan `json:"plans"`
}

// DataCloudWorkspace is a workspace in one region.
// Status is a plain string, for example pending, provisioning, active, or errored.
// QueryInstanceSize is a plain string. Advertised values are small, medium, and large.
type DataCloudWorkspace struct {
	UUID              string `json:"uuid"`
	OwnerUUID         string `json:"owner_uuid,omitempty"`
	Name              string `json:"name,omitempty"`
	Region            string `json:"region,omitempty"`
	Status            string `json:"status,omitempty"`
	QueryEndpoint     string `json:"query_endpoint,omitempty"`
	QueryInstanceSize string `json:"query_instance_size,omitempty"`
}

// DataCloudWorkspaceCreateRequest is the body for CreateWorkspace.
// Region is not a body field. Pass it as the region argument. Name and
// QueryInstanceSize are optional. An empty name leaves the server default.
// An empty QueryInstanceSize leaves the server default, small.
type DataCloudWorkspaceCreateRequest struct {
	Name              string `json:"name,omitempty"`
	QueryInstanceSize string `json:"query_instance_size,omitempty"`
}

// DataCloudTokenPurpose selects how a data-plane token can be used.
type DataCloudTokenPurpose string

const (
	DataCloudTokenPurposeUnknown DataCloudTokenPurpose = "TOKEN_PURPOSE_UNKNOWN"
	// DataCloudTokenPurposeQuery grants query access to the selected workspaces.
	DataCloudTokenPurposeQuery DataCloudTokenPurpose = "TOKEN_PURPOSE_QUERY"
)

// DataCloudTokenCreateRequest is the body for CreateToken.
// Region is required. The other fields are optional and omitted when empty.
// WorkspaceUUIDs is the workspace binding list. The legacy sandbox_uuids field is not sent.
// ExpiresInSeconds is written as a JSON string. Zero means the server default and is omitted.
type DataCloudTokenCreateRequest struct {
	Region           string                `json:"region"`
	Purpose          DataCloudTokenPurpose `json:"purpose,omitempty"`
	WorkspaceUUIDs   []string              `json:"workspace_uuids,omitempty"`
	Name             string                `json:"name,omitempty"`
	ExpiresInSeconds int64                 `json:"-"`
	Permissions      []string              `json:"permissions,omitempty"`
}

// MarshalJSON writes expires_in_seconds as a decimal string. Proto JSON represents int64 that way.
func (r DataCloudTokenCreateRequest) MarshalJSON() ([]byte, error) {
	payload := struct {
		Region           string                `json:"region"`
		Purpose          DataCloudTokenPurpose `json:"purpose,omitempty"`
		WorkspaceUUIDs   []string              `json:"workspace_uuids,omitempty"`
		Name             string                `json:"name,omitempty"`
		ExpiresInSeconds string                `json:"expires_in_seconds,omitempty"`
		Permissions      []string              `json:"permissions,omitempty"`
	}{
		Region:         r.Region,
		Purpose:        r.Purpose,
		WorkspaceUUIDs: r.WorkspaceUUIDs,
		Name:           r.Name,
		Permissions:    r.Permissions,
	}
	if r.ExpiresInSeconds != 0 {
		payload.ExpiresInSeconds = strconv.FormatInt(r.ExpiresInSeconds, 10)
	}
	return json.Marshal(payload)
}

// DataCloudTokenCreateResponse is the body returned by CreateToken.
// AccessToken is returned once, on create. Do not log this struct with %v or %+v.
type DataCloudTokenCreateResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name,omitempty"`
	AccessToken string `json:"access_token,omitempty"`
	ExpiresAt   string `json:"expires_at,omitempty"`
}

// DataCloudTokenListRequest holds path and query args for ListTokens.
// Region is required and used as the path segment. The other fields are
// optional query filters and omitted when empty. SandboxUUID is superseded
// by WorkspaceUUID; prefer WorkspaceUUID for new callers.
type DataCloudTokenListRequest struct {
	Region        string                `url:"-"`
	Purpose       DataCloudTokenPurpose `url:"purpose,omitempty"`
	SandboxUUID   string                `url:"sandbox_uuid,omitempty"`
	Page          int64                 `url:"page,omitempty"`
	PerPage       int64                 `url:"per_page,omitempty"`
	WorkspaceUUID string                `url:"workspace_uuid,omitempty"`
}

// DataCloudTokenSummary is one token in a ListTokens response.
// It never includes an access token. SandboxUUIDs is superseded by WorkspaceUUIDs.
type DataCloudTokenSummary struct {
	ID             string                `json:"id"`
	Name           string                `json:"name,omitempty"`
	CreatedAt      string                `json:"created_at,omitempty"`
	ExpiresAt      string                `json:"expires_at,omitempty"`
	Purpose        DataCloudTokenPurpose `json:"purpose,omitempty"`
	SandboxUUIDs   []string              `json:"sandbox_uuids,omitempty"`
	QueryHosts     []string              `json:"query_hosts,omitempty"`
	WorkspaceUUIDs []string              `json:"workspace_uuids,omitempty"`
}

// DataCloudListTokensResponse is the body of ListTokens.
// Tokens and Total are siblings. Total is a decimal string (proto JSON int64).
type DataCloudListTokensResponse struct {
	Tokens []DataCloudTokenSummary `json:"tokens"`
	Total  string                  `json:"total,omitempty"`
}

// DataCloudSecretOptions configures a secret write.
// CAS is currently ignored by the server because secret version checks are not supported.
type DataCloudSecretOptions struct {
	CAS int32 `json:"cas,omitempty"`
}

// DataCloudPutWorkspaceSecretRequest is the body for PutWorkspaceSecret.
// Region, UUID, and SecretKey are required path segments and are also sent in the body.
// Data must contain at least one string key-value pair.
type DataCloudPutWorkspaceSecretRequest struct {
	Region    string                  `json:"region"`
	UUID      string                  `json:"uuid"`
	SecretKey string                  `json:"secret_key"`
	Data      map[string]string       `json:"data,omitempty"`
	Options   *DataCloudSecretOptions `json:"options,omitempty"`
}

// DataCloudSecretMetadata describes a stored secret version.
type DataCloudSecretMetadata struct {
	CreatedTime    string         `json:"created_time,omitempty"`
	CustomMetadata map[string]any `json:"custom_metadata,omitempty"`
	DeletionTime   string         `json:"deletion_time,omitempty"`
	Destroyed      bool           `json:"destroyed"`
	Version        int32          `json:"version,omitempty"`
}

// DataCloudPutWorkspaceSecretResponse is the body of PutWorkspaceSecret.
type DataCloudPutWorkspaceSecretResponse struct {
	Data *DataCloudSecretMetadata `json:"data,omitempty"`
}

// DataCloudGetSecretData is the nested payload of GetWorkspaceSecret.
// Keys lists property names. Values are never returned.
type DataCloudGetSecretData struct {
	Data     map[string]any           `json:"data,omitempty"`
	Metadata *DataCloudSecretMetadata `json:"metadata,omitempty"`
	Keys     []string                 `json:"keys,omitempty"`
}

// DataCloudGetWorkspaceSecretResponse is the body of GetWorkspaceSecret.
type DataCloudGetWorkspaceSecretResponse struct {
	Data *DataCloudGetSecretData `json:"data,omitempty"`
}

// DataCloudListWorkspaceSecretsResponse is the body of ListWorkspaceSecrets.
type DataCloudListWorkspaceSecretsResponse struct {
	SecretKeys []string `json:"secret_keys"`
}

// DataCloudQueryCredentials is the body of GetWorkspaceQueryCredentials.
// Password is a secret. Do not log this struct with %v or %+v.
type DataCloudQueryCredentials struct {
	User     string `json:"user,omitempty"`
	Password string `json:"password,omitempty"`
}

// DataCloudWorkspacesResponse is the body of ListWorkspaces.
type DataCloudWorkspacesResponse struct {
	Workspaces []DataCloudWorkspace `json:"workspaces"`
}

type dataCloudWorkspaceRoot struct {
	Workspace *DataCloudWorkspace `json:"workspace"`
}

func dataCloudWorkspacesPath(region string) (string, error) {
	if region == "" {
		return "", errors.New("data cloud platform: region is required")
	}
	return dataCloudPlatformPrefix + "/" + url.PathEscape(region) + "/workspaces", nil
}

func dataCloudWorkspacePath(region, uuid string) (string, error) {
	if uuid == "" {
		return "", errors.New("data cloud platform: workspace uuid is required")
	}
	base, err := dataCloudWorkspacesPath(region)
	if err != nil {
		return "", err
	}
	return base + "/" + url.PathEscape(uuid), nil
}

func dataCloudTokensPath(region string) (string, error) {
	if region == "" {
		return "", errors.New("data cloud platform: region is required")
	}
	return dataCloudPlatformPrefix + "/" + url.PathEscape(region) + "/tokens", nil
}

func dataCloudWorkspaceSecretsPath(region, uuid string) (string, error) {
	base, err := dataCloudWorkspacePath(region, uuid)
	if err != nil {
		return "", err
	}
	return base + "/secrets", nil
}

func dataCloudWorkspaceSecretPath(region, uuid, secretKey string) (string, error) {
	if secretKey == "" {
		return "", errors.New("data cloud platform: secret key is required")
	}
	base, err := dataCloudWorkspaceSecretsPath(region, uuid)
	if err != nil {
		return "", err
	}
	return base + "/" + url.PathEscape(secretKey), nil
}

func dataCloudWorkspaceQueryCredentialsPath(region, uuid string) (string, error) {
	base, err := dataCloudWorkspacePath(region, uuid)
	if err != nil {
		return "", err
	}
	return base + "/query/credentials", nil
}

// ListPlans returns the regions and query engine sizes a team can use when
// creating a workspace.
func (s *DataCloudPlatformServiceOp) ListPlans(ctx context.Context) (*DataCloudPlansResponse, *Response, error) {
	req, err := s.client.NewRequest(ctx, http.MethodGet, dataCloudPlatformPrefix+"/plans", nil)
	if err != nil {
		return nil, nil, err
	}
	root := new(DataCloudPlansResponse)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}
	return root, resp, nil
}

// CreateWorkspace creates a workspace in region.
func (s *DataCloudPlatformServiceOp) CreateWorkspace(ctx context.Context, region string, create *DataCloudWorkspaceCreateRequest) (*DataCloudWorkspace, *Response, error) {
	if create == nil {
		return nil, nil, errors.New("data cloud platform: create request is required")
	}
	path, err := dataCloudWorkspacesPath(region)
	if err != nil {
		return nil, nil, err
	}
	req, err := s.client.NewRequest(ctx, http.MethodPost, path, create)
	if err != nil {
		return nil, nil, err
	}
	root := new(dataCloudWorkspaceRoot)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}
	if root.Workspace == nil {
		return nil, resp, errors.New("data cloud platform: create workspace returned no workspace")
	}
	return root.Workspace, resp, nil
}

// CreateToken creates a token. Region comes from create and is used as the path segment.
func (s *DataCloudPlatformServiceOp) CreateToken(ctx context.Context, create *DataCloudTokenCreateRequest) (*DataCloudTokenCreateResponse, *Response, error) {
	if create == nil {
		return nil, nil, errors.New("data cloud platform: create request is required")
	}
	path, err := dataCloudTokensPath(create.Region)
	if err != nil {
		return nil, nil, err
	}
	req, err := s.client.NewRequest(ctx, http.MethodPost, path, create)
	if err != nil {
		return nil, nil, err
	}
	root := new(DataCloudTokenCreateResponse)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}
	return root, resp, nil
}

// ListTokens returns tokens in the region on list. Optional filters become query parameters.
// An empty Purpose omits the purpose query param so the server default applies.
func (s *DataCloudPlatformServiceOp) ListTokens(ctx context.Context, list *DataCloudTokenListRequest) (*DataCloudListTokensResponse, *Response, error) {
	if list == nil {
		return nil, nil, errors.New("data cloud platform: list request is required")
	}
	path, err := dataCloudTokensPath(list.Region)
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
	root := new(DataCloudListTokensResponse)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}
	return root, resp, nil
}

// PutWorkspaceSecret writes a workspace secret at
// /v2/data-cloud-platform/{region}/workspaces/{uuid}/secrets/{secret_key}.
func (s *DataCloudPlatformServiceOp) PutWorkspaceSecret(ctx context.Context, put *DataCloudPutWorkspaceSecretRequest) (*DataCloudPutWorkspaceSecretResponse, *Response, error) {
	if put == nil {
		return nil, nil, errors.New("data cloud platform: put request is required")
	}
	if len(put.Data) == 0 {
		return nil, nil, errors.New("data cloud platform: secret data is required")
	}
	path, err := dataCloudWorkspaceSecretPath(put.Region, put.UUID, put.SecretKey)
	if err != nil {
		return nil, nil, err
	}
	req, err := s.client.NewRequest(ctx, http.MethodPut, path, put)
	if err != nil {
		return nil, nil, err
	}
	root := new(DataCloudPutWorkspaceSecretResponse)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}
	return root, resp, nil
}

// GetWorkspaceSecret returns metadata and property names for a workspace secret.
func (s *DataCloudPlatformServiceOp) GetWorkspaceSecret(ctx context.Context, region, uuid, secretKey string) (*DataCloudGetWorkspaceSecretResponse, *Response, error) {
	path, err := dataCloudWorkspaceSecretPath(region, uuid, secretKey)
	if err != nil {
		return nil, nil, err
	}
	req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}
	root := new(DataCloudGetWorkspaceSecretResponse)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}
	return root, resp, nil
}

// ListWorkspaceSecrets returns secret key names for a workspace.
func (s *DataCloudPlatformServiceOp) ListWorkspaceSecrets(ctx context.Context, region, uuid string) (*DataCloudListWorkspaceSecretsResponse, *Response, error) {
	path, err := dataCloudWorkspaceSecretsPath(region, uuid)
	if err != nil {
		return nil, nil, err
	}
	req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}
	root := new(DataCloudListWorkspaceSecretsResponse)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}
	return root, resp, nil
}

// DeleteWorkspaceSecret permanently deletes a workspace secret.
func (s *DataCloudPlatformServiceOp) DeleteWorkspaceSecret(ctx context.Context, region, uuid, secretKey string) (*Response, error) {
	path, err := dataCloudWorkspaceSecretPath(region, uuid, secretKey)
	if err != nil {
		return nil, err
	}
	req, err := s.client.NewRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return nil, err
	}
	return s.client.Do(ctx, req, nil)
}

// GetWorkspaceQueryCredentials returns query user and password for a workspace.
func (s *DataCloudPlatformServiceOp) GetWorkspaceQueryCredentials(ctx context.Context, region, uuid string) (*DataCloudQueryCredentials, *Response, error) {
	path, err := dataCloudWorkspaceQueryCredentialsPath(region, uuid)
	if err != nil {
		return nil, nil, err
	}
	req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}
	root := new(DataCloudQueryCredentials)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}
	return root, resp, nil
}

// ListWorkspaces returns the workspaces in region.
func (s *DataCloudPlatformServiceOp) ListWorkspaces(ctx context.Context, region string) (*DataCloudWorkspacesResponse, *Response, error) {
	path, err := dataCloudWorkspacesPath(region)
	if err != nil {
		return nil, nil, err
	}
	req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}
	root := new(DataCloudWorkspacesResponse)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}
	return root, resp, nil
}

// GetWorkspace returns one workspace by region and uuid.
func (s *DataCloudPlatformServiceOp) GetWorkspace(ctx context.Context, region, uuid string) (*DataCloudWorkspace, *Response, error) {
	path, err := dataCloudWorkspacePath(region, uuid)
	if err != nil {
		return nil, nil, err
	}
	req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}
	root := new(dataCloudWorkspaceRoot)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}
	if root.Workspace == nil {
		return nil, resp, errors.New("data cloud platform: get workspace returned no workspace")
	}
	return root.Workspace, resp, nil
}

// DeleteWorkspace deletes one workspace.
func (s *DataCloudPlatformServiceOp) DeleteWorkspace(ctx context.Context, region, uuid string) (*Response, error) {
	path, err := dataCloudWorkspacePath(region, uuid)
	if err != nil {
		return nil, err
	}
	req, err := s.client.NewRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return nil, err
	}
	return s.client.Do(ctx, req, nil)
}
