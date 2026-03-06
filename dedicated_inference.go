package godo

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

const dedicatedInferenceBasePath = "/v2/dedicated-inferences"

// DedicatedInferenceService is an interface for managing Dedicated Inference with the DigitalOcean API.
type DedicatedInferenceService interface {
	Create(context.Context, *DedicatedInferenceCreateRequest) (*DedicatedInference, *DedicatedInferenceToken, *Response, error)
	Get(context.Context, string) (*DedicatedInference, *Response, error)
	Update(context.Context, string, *DedicatedInferenceUpdateRequest) (*DedicatedInference, *Response, error)
}

// DedicatedInferenceServiceOp handles communication with Dedicated Inference methods of the DigitalOcean API.
type DedicatedInferenceServiceOp struct {
	client *Client
}

var _ DedicatedInferenceService = &DedicatedInferenceServiceOp{}

// DedicatedInferenceCreateRequest represents a request to create a Dedicated Inference.
type DedicatedInferenceCreateRequest struct {
	Spec    *DedicatedInferenceSpecRequest `json:"spec"`
	Secrets *DedicatedInferenceSecrets     `json:"secrets,omitempty"`
}

// DedicatedInferenceSpecRequest represents the deployment specification in a create/update request.
type DedicatedInferenceSpecRequest struct {
	Version              int                               `json:"version"`
	Name                 string                            `json:"name"`
	Region               string                            `json:"region"`
	EnablePublicEndpoint bool                              `json:"enable_public_endpoint"`
	VPC                  *DedicatedInferenceVPCRequest     `json:"vpc"`
	ModelDeployments     []*DedicatedInferenceModelRequest `json:"model_deployments"`
}

// DedicatedInferenceVPCRequest represents the VPC configuration in a request.
type DedicatedInferenceVPCRequest struct {
	UUID string `json:"uuid"`
}

// DedicatedInferenceModelRequest represents a model deployment in a request.
type DedicatedInferenceModelRequest struct {
	ModelID        string                                  `json:"model_id,omitempty"`
	ModelSlug      string                                  `json:"model_slug"`
	ModelProvider  string                                  `json:"model_provider"`
	WorkloadConfig *DedicatedInferenceWorkloadConfig       `json:"workload_config,omitempty"`
	Accelerators   []*DedicatedInferenceAcceleratorRequest `json:"accelerators"`
}

// DedicatedInferenceWorkloadConfig represents workload-specific configuration.
type DedicatedInferenceWorkloadConfig struct{}

// DedicatedInferenceAcceleratorRequest represents an accelerator in a request.
type DedicatedInferenceAcceleratorRequest struct {
	AcceleratorSlug string `json:"accelerator_slug"`
	Scale           uint64 `json:"scale"`
	Type            string `json:"type"`
}

// DedicatedInferenceSecrets represents secrets for external model providers.
type DedicatedInferenceSecrets struct {
	HuggingFaceToken string `json:"hugging_face_token,omitempty"`
}

// DedicatedInferenceUpdateRequest represents a request to update a Dedicated Inference.
type DedicatedInferenceUpdateRequest struct {
	Spec    *DedicatedInferenceSpecRequest `json:"spec"`
	Secrets *DedicatedInferenceSecrets     `json:"secrets,omitempty"`
}

// -- Response types (what the API returns) --

// DedicatedInference represents a Dedicated Inference resource returned by the API.
type DedicatedInference struct {
	ID                    string                        `json:"id"`
	Name                  string                        `json:"name"`
	Region                string                        `json:"region"`
	Status                string                        `json:"status"`
	VPCUUID               string                        `json:"vpc_uuid"`
	Endpoints             *DedicatedInferenceEndpoints  `json:"endpoints,omitempty"`
	DeploymentSpec        *DedicatedInferenceDeployment `json:"spec,omitempty"`
	PendingDeploymentSpec *DedicatedInferenceDeployment `json:"pending_deployment_spec,omitempty"`
	CreatedAt             time.Time                     `json:"created_at,omitempty"`
	UpdatedAt             time.Time                     `json:"updated_at,omitempty"`
}

func (d DedicatedInference) String() string {
	return Stringify(d)
}

// DedicatedInferenceEndpoints represents the endpoints for a Dedicated Inference.
type DedicatedInferenceEndpoints struct {
	PublicEndpointFQDN  string `json:"public_endpoint_fqdn,omitempty"`
	PrivateEndpointFQDN string `json:"private_endpoint_fqdn,omitempty"`
}

// DedicatedInferenceDeployment represents a deployment spec in the API response.
type DedicatedInferenceDeployment struct {
	Version              uint64                               `json:"version"`
	ID                   string                               `json:"id"`
	DedicatedInferenceID string                               `json:"dedicated_inference_id"`
	State                string                               `json:"state"`
	EnablePublicEndpoint bool                                 `json:"enable_public_endpoint"`
	VPCConfig            *DedicatedInferenceVPCConfig         `json:"vpc_config,omitempty"`
	ModelDeployments     []*DedicatedInferenceModelDeployment `json:"model_deployments"`
	CreatedAt            time.Time                            `json:"created_at,omitempty"`
	UpdatedAt            time.Time                            `json:"updated_at,omitempty"`
}

// DedicatedInferenceVPCConfig represents the VPC config in an API response.
type DedicatedInferenceVPCConfig struct {
	VPCUUID string `json:"vpc_uuid"`
}

// DedicatedInferenceModelDeployment represents a model deployment in an API response.
type DedicatedInferenceModelDeployment struct {
	ModelID       string                           `json:"model_id"`
	ModelSlug     string                           `json:"model_slug"`
	ModelProvider string                           `json:"model_provider"`
	Accelerators  []*DedicatedInferenceAccelerator `json:"accelerators"`
}

// DedicatedInferenceAccelerator represents an accelerator in an API response.
type DedicatedInferenceAccelerator struct {
	AcceleratorID   string `json:"accelerator_id"`
	AcceleratorSlug string `json:"accelerator_slug"`
	State           string `json:"state"`
	Type            string `json:"type"`
	Scale           uint64 `json:"scale"`
}

// DedicatedInferenceToken represents an auth token returned on create.
type DedicatedInferenceToken struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Value     string    `json:"value,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

func (t DedicatedInferenceToken) String() string {
	return Stringify(t)
}

// -- Root types for JSON deserialization --

type dedicatedInferenceRoot struct {
	DedicatedInference *DedicatedInference      `json:"dedicated_inference"`
	Token              *DedicatedInferenceToken `json:"token,omitempty"`
}

// -- Service methods --

// Create a new Dedicated Inference with the given configuration.
func (s *DedicatedInferenceServiceOp) Create(ctx context.Context, createRequest *DedicatedInferenceCreateRequest) (*DedicatedInference, *DedicatedInferenceToken, *Response, error) {
	req, err := s.client.NewRequest(ctx, http.MethodPost, dedicatedInferenceBasePath, createRequest)
	if err != nil {
		return nil, nil, nil, err
	}

	root := new(dedicatedInferenceRoot)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, nil, resp, err
	}

	return root.DedicatedInference, root.Token, resp, nil
}

// Get an existing Dedicated Inference by its UUID.
func (s *DedicatedInferenceServiceOp) Get(ctx context.Context, id string) (*DedicatedInference, *Response, error) {
	path := fmt.Sprintf("%s/%s", dedicatedInferenceBasePath, id)

	req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	root := new(dedicatedInferenceRoot)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}

	return root.DedicatedInference, resp, nil
}

// Update an existing Dedicated Inference.
func (s *DedicatedInferenceServiceOp) Update(ctx context.Context, id string, updateRequest *DedicatedInferenceUpdateRequest) (*DedicatedInference, *Response, error) {
	path := fmt.Sprintf("%s/%s", dedicatedInferenceBasePath, id)

	req, err := s.client.NewRequest(ctx, http.MethodPatch, path, updateRequest)
	if err != nil {
		return nil, nil, err
	}

	root := new(dedicatedInferenceRoot)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}

	return root.DedicatedInference, resp, nil
}
