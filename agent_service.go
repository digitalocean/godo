package godo

import (
	"context"
	"fmt"
	"net/http"
)

const agentConnectBasePath = "/v2/gen-ai/agents"

type AgentService interface {
	List(context.Context, *ListOptions) ([]*Agents, *Response, error)
	Create(context.Context, *AgentCreateRequest) (*Agent, *Response, error)
	Get(context.Context, string) (*Agent, *Response, error)
	Update(context.Context, string, *AgentUpdateRequest) (*Agent, *Response, error)
	Delete(context.Context, string) (*Agent, *Response, error)
	UpdateVisibility(context.Context, string, *AgentVisibilityUpdateRequest) (*Agent, *Response, error)
}

var _ AgentService = &AgentServiceOp{}

type AgentServiceOp struct {
	client *Client
}

type genAIAgentsRoot struct {
	Agents []*Agents `json:"agents"`
	Links  *Links    `json:"links"`
	Meta   *Meta     `json:"meta"`
}

type genAIAgentRoot struct {
	Agent *Agent `json:"agent"`
}

type Agent struct {
	AnthropicApiKey   *AnthropicApiKeyInfo     `json:"anthropic_api_key,omitempty"`
	ApiKeyInfos       []*ApiKeyInfo            `json:"api_key_infos,omitempty"`
	ApiKeys           []*ApiKey                `json:"api_keys,omitempty"`
	ChatBot           *ChatBot                 `json:"chatbot,omitempty"`
	ChatbotIdentifier []AgentChatbotIdentifier `json:"chatbot_identifiers,omitempty"`
	CreatedAt         *Timestamp               `json:"created_at,omitempty"`
	Deployment        *AgentDeployment         `json:"deployment,omitempty"`
	Description        string                   `json:"description,omitempty"`
	UpdatedAt         *Timestamp               `json:"updated_at,omitempty"`
	Functions         []*AgentFunction         `json:"functions,omitempty"`
	Guardrails        []*AgentGuardrail        `json:"guardrails,omitempty"`
	IfCase            string                   `json:"if_case,omitempty"`
	Instruction       string                   `json:"instruction,omitempty"`
	K                 int                      `json:"k,omitempty"`
	KnowledgeBases    []*KnowledgeBase         `json:"knowledge_bases,omitempty"`
	MaxToken          int                      `json:"max_tokens,omitempty"`
	Model             *Model                   `json:"model,omitempty"`
	Name              string                   `json:"name,omitempty"`
	OpenAiApiKey      *OpenAiApiKey            `json:"open_ai_api_key,omitempty"`
	ProjectId         string                   `json:"project_id,omitempty"`
	Region            string                   `json:"region,omitempty"`
	RetrievalMethod   string                   `json:"retrieval_method,omitempty"`
	RouteCreatedAt    *Timestamp               `json:"route_created_at,omitempty"`
	RouteCreatedBy    string                   `json:"route_created_by,omitempty"`
	RouteUuid         string                   `json:"route_uuid,omitempty"`
	RouteName         string                   `json:"route_name,omitempty"`
	Tags              []string                 `json:"tags,omitempty"`
	Template          *AgentTemplate           `json:"template,omitempty"`
	Temperature       float64                  `json:"temperature,omitempty"`
	TopP              float64                  `json:"top_p,omitempty"`
	Url               string                   `json:"url,omitempty"`
	UserId            string                   `json:"user_id,omitempty"`
	Uuid              string                   `json:"uuid,omitempty"`
}

type Agents struct {
	ChatBot            *ChatBot                 `json:"chatbot,omitempty"`
	ChatbotIdentifiers []AgentChatbotIdentifier `json:"chatbot_identifiers,omitempty"`
	Name               string                   `json:"name,omitempty"`
	CreatedAt          *Timestamp               `json:"created_at,omitempty"`
	UpdatedAt           *Timestamp               `json:"updated_at,omitempty"`
	Instruction        string                   `json:"instruction,omitempty"`
	Descripton         string                   `json:"description,omitempty"`
	IfCase             string                   `json:"if_case,omitempty"`
	K                  int                      `json:"k,omitempty"`
	MaxToken           int                      `json:"max_tokens,omitempty"`
	ProjectId          string                   `json:"project_id,omitempty"`
	Region             string                   `json:"region,omitempty"`
	RetrievalMethod    string                   `json:"retrieval_method,omitempty"`
	RouteCreatedAt     *Timestamp               `json:"route_created_at,omitempty"`
	RouteCreatedBy     string                   `json:"route_created_by,omitempty"`
	RouteUuid          string                   `json:"route_uuid,omitempty"`
	RouteName          string                   `json:"route_name,omitempty"`
	Model              *Model                   `json:"model,omitempty"`
	Deployment         *AgentDeployment         `json:"deployment,omitempty"`
	Tags               []string                 `json:"tags,omitempty"`
	Template           *AgentTemplate           `json:"template,omitempty"`
	Temperature        float64                  `json:"temperature,omitempty"`
	TopP               float64                  `json:"top_p,omitempty"`
	Url                string                   `json:"url,omitempty"`
	UserId             string                   `json:"user_id,omitempty"`
	Uuid               string                   `json:"uuid,omitempty"`
}

type AgentFunction struct {
	ApiKey        string     `json:"api_key,omitempty"`
	CreatedAt     *Timestamp `json:"created_at,omitempty"`
	Description   string     `json:"description,omitempty"`
	GuardrailUuid string     `json:"guardrail_uuid,omitempty"`
	FaasName      string     `json:"faas_name,omitempty"`
	FaasNamespace string     `json:"faas_namespace,omitempty"`
	Name          string     `json:"name,omitempty"`
	UpdatedAt     *Timestamp `json:"updated_at,omitempty"`
	Url           string     `json:"url,omitempty"`
	Uuid          string     `json:"uuid,omitempty"`
}

type AgentGuardrail struct {
	AgentUuid       string     `json:"agent_uuid,omitempty"`
	CreatedAt       *Timestamp `json:"created_at,omitempty"`
	DefaultResponse string     `json:"default_response,omitempty"`
	Description     string     `json:"description,omitempty"`
	GuardrailUuid   string     `json:"guardrail_uuid,omitempty"`
	IsAttached      bool       `json:"is_attached,omitempty"`
	IsDefault       bool       `json:"is_default,omitempty"`
	Name            string     `json:"name,omitempty"`
	Priority        int        `json:"priority,omitempty"`
	Type            string     `json:"type,omitempty"`
	UpdatedAt       *Timestamp `json:"updated_at,omitempty"`
	Uuid            string     `json:"uuid,omitempty"`
}

type ApiKey struct {
	ApiKey string `json:"api_key,omitempty"`
}

type AnthropicApiKeyInfo struct {
	CreatedAt *Timestamp `json:"created_at,omitempty"`
	CreatedBy string     `json:"created_by,omitempty"`
	DeletedAt *Timestamp `json:"deleted_at,omitempty"`
	Name      string     `json:"name,omitempty"`
	UpdatedAt *Timestamp `json:"updated_at,omitempty"`
	Uuid      string     `json:"uuid,omitempty"`
}

type ApiKeyInfo struct {
	CreatedAt *Timestamp `json:"created_at,omitempty"`
	CreatedBy string     `json:"created_by,omitempty"`
	DeletedAt *Timestamp `json:"deleted_at,omitempty"`
	Name      string     `json:"name,omitempty"`
	SecretKey string     `json:"secret_key,omitempty"`
	Uuid      string     `json:"uuid,omitempty"`
}

type OpenAiApiKey struct {
	CreatedAt *Timestamp `json:"created_at,omitempty"`
	CreatedBy string     `json:"created_by,omitempty"`
	DeletedAt *Timestamp `json:"deleted_at,omitempty"`
	Models    []*Model   `json:"models,omitempty"`
	Name      string     `json:"name,omitempty"`
	UpdatedAt *Timestamp `json:"updated_at,omitempty"`
	Uuid      string     `json:"uuid,omitempty"`
}

type AgentVisibilityUpdateRequest struct {
	Uuid       string `json:"uuid,omitempty"`
	Visibility string `json:"visibility,omitempty"`
}

type AgentTemplate struct {
	CreatedAt      *Timestamp       `json:"created_at,omitempty"`
	Instruction    string           `json:"instruction,omitempty"`
	Description    string           `json:"description,omitempty"`
	K              int              `json:"k,omitempty"`
	KnowledgeBases []*KnowledgeBase `json:"knowledge_bases,omitempty"`
	MaxToken       int              `json:"max_tokens,omitempty"`
	Model          *Model           `json:"model,omitempty"`
	Name           string           `json:"name,omitempty"`
	Temperature    float64          `json:"temperature,omitempty"`
	TopP           float64          `json:"top_p,omitempty"`
	UpdatedAt      *Timestamp       `json:"updated_at,omitempty"`
	Uuid           string           `json:"uuid,omitempty"`
}

type KnowledgeBase struct {
	AddedToAgentAt     *Timestamp       `json:"added_to_agent_at,omitempty"`
	CreatedAt          *Timestamp       `json:"created_at,omitempty"`
	DatabaseId         string           `json:"database_id,omitempty"`
	EmbeddingModelUuid string           `json:"embedding_model_uuid,omitempty"`
	IsPublic           bool             `json:"is_public,omitempty"`
	LastIndexingJob    *LastIndexingJob `json:"last_indexing_job,omitempty"`
	Name               string           `json:"name,omitempty"`
	ProjectId          string           `json:"project_id,omitempty"`
	Region             string           `json:"region,omitempty"`
	Tags               []string         `json:"tags,omitempty"`
	UpdateAt           *Timestamp       `json:"updated_at,omitempty"`
	UserId             string           `json:"user_id,omitempty"`
	Uuid               string           `json:"uuid,omitempty"`
}

type LastIndexingJob struct {
	CompletedDatasources int        `json:"completed_datasources,omitempty"`
	CreatedAt            *Timestamp `json:"created_at,omitempty"`
	DataSourceUuids      []string   `json:"data_source_uuids,omitempty"`
	FinishedAt           *Timestamp `json:"finished_at,omitempty"`
	KnowledgeBaseUuid    string     `json:"knowledge_base_uuid,omitempty"`
	Phase                string     `json:"phase,omitempty"`
	StartedAt            *Timestamp `json:"started_at,omitempty"`
	Tokens               int        `json:"tokens,omitempty"`
	TotalDatasources     int        `json:"total_datasources,omitempty"`
	UpdatedAt            *Timestamp `json:"updated_at,omitempty"`
	Uuid                 string     `json:"uuid,omitempty"`
}

type AgentChatbotIdentifier struct {
	AgentChatbotIdentifier string `json:"agent_chatbot_identifier,omitempty"`
}

type AgentDeployment struct {
	CreatedAt  *Timestamp `json:"created_at,omitempty"`
	Name       string     `json:"name,omitempty"`
	Status     string     `json:"status,omitempty"`
	UpdatedAt  *Timestamp `json:"updated_at,omitempty"`
	Url        string     `json:"url,omitempty"`
	Uuid       string     `json:"uuid,omitempty"`
	Visibility string     `json:"visibility,omitempty"`
}

type ChatBot struct {
	ButtonBackgroundColor string `json:"button_background_color,omitempty"`
	Logo                  string `json:"logo,omitempty"`
	Name                  string `json:"name,omitempty"`
	PrimaryColor          string `json:"primary_color,omitempty"`
	SecondaryColor        string `json:"secondary_color,omitempty"`
	StartingMessage       string `json:"starting_message,omitempty"`
}

type Model struct {
	Agreement        *Agreement    `json:"agreement,omitempty"`
	CreatedAt        *Timestamp    `json:"created_at,omitempty"`
	InferenceName    string        `json:"inference_name,omitempty"`
	InferenceVersion string        `json:"inference_version,omitempty"`
	IsFoundational   bool          `json:"is_foundational,omitempty"`
	Name             string        `json:"name,omitempty"`
	ParentUuid       string        `json:"parent_uuid,omitempty"`
	Provider         string        `json:"provider,omitempty"`
	UpdatedAt        *Timestamp    `json:"updated_at,omitempty"`
	UploadComplete   bool          `json:"upload_complete,omitempty"`
	Url              string        `json:"url,omitempty"`
	Usecases         []string      `json:"usecases,omitempty"`
	Uuid             string        `json:"uuid,omitempty"`
	Version          *ModelVersion `json:"version,omitempty"`
}

type Agreement struct {
	Description string `json:"description,omitempty"`
	Name        string `json:"name,omitempty"`
	Url         string `json:"url,omitempty"`
	Uuid        string `json:"uuid,omitempty"`
}

type ModelVersion struct {
	Major int `json:"major,omitempty"`
	Minor int `json:"minor,omitempty"`
	Patch int `json:"patch,omitempty"`
}

type AgentCreateRequest struct {
	AnthropicKeyUuid  string   `json:"anthropic_key_uuid,omitempty"`
	Description       string   `json:"description,omitempty"`
	Instruction       string   `json:"instruction,omitempty"`
	KnowledgeBaseUuid []string `json:"knowledge_base_uuid,omitempty"`
	ModelUuid         string   `json:"model_uuid,omitempty"`
	Name              string   `json:"name,omitempty"`
	OpenAiKeyUuid     string   `json:"open_ai_key_uuid,omitempty"`
	ProjectId         string   `json:"project_id,omitempty"`
	Region            string   `json:"region,omitempty"`
	Tags              []string `json:"tags,omitempty"`
}

type AgentUpdateRequest struct {
	AnthropicKeyUuid string   `json:"anthropic_key_uuid,omitempty"`
	Description      string   `json:"description,omitempty"`
	Instruction      string   `json:"instruction,omitempty"`
	K                int      `json:"k,omitempty"`
	MaxTokens         int      `json:"max_tokens,omitempty"`
	ModelUuid        string   `json:"model_uuid,omitempty"`
	Name             string   `json:"name,omitempty"`
	OpenAiKeyUuid    string   `json:"open_ai_key_uuid,omitempty"`
	ProjectId        string   `json:"project_id,omitempty"`
	Region           string   `json:"region,omitempty"`
	Tags             []string `json:"tags,omitempty"`
	Temperature      float64  `json:"temperature,omitempty"`
	TopP             float64  `json:"top_p,omitempty"`
	Uuid             string   `json:"uuid,omitempty"`
}

// List returns a list of Gen AI Agents
func (s *AgentServiceOp) List(ctx context.Context, opt *ListOptions) ([]*Agents, *Response, error) {
	path, err := addOptions(agentConnectBasePath, opt)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	root := new(genAIAgentsRoot)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}
	if l := root.Links; l != nil {
		resp.Links = l
	}
	if m := root.Meta; m != nil {
		resp.Meta = m
	}
	return root.Agents, resp, nil
}

// Create creates a new Gen AI Agent by providing the AgentCreateRequest object
func (s *AgentServiceOp) Create(ctx context.Context, create *AgentCreateRequest) (*Agent, *Response, error) {
	path := agentConnectBasePath

	req, err := s.client.NewRequest(ctx, http.MethodPost, path, create)
	if err != nil {
		return nil, nil, err
	}

	root := new(genAIAgentRoot)
	resp, err := s.client.Do(ctx, req, root)

	if err != nil {
		return nil, resp, err
	}

	return root.Agent, resp, nil
}

// Get returns the details of a Gen AI Agent based on the Agent UUID
func (s *AgentServiceOp) Get(ctx context.Context, id string) (*Agent, *Response, error) {
	path := fmt.Sprintf("%s/%s", agentConnectBasePath, id)

	req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	root := new(genAIAgentRoot)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}

	return root.Agent, resp, nil
}

// Update function updates a Gen AI Agent properties for the given UUID
func (s *AgentServiceOp) Update(ctx context.Context, id string, update *AgentUpdateRequest) (*Agent, *Response, error) {
	path := fmt.Sprintf("%s/%s", agentConnectBasePath, id)
	req, err := s.client.NewRequest(ctx, http.MethodPut, path, update)
	if err != nil {
		return nil, nil, err
	}

	root := new(genAIAgentRoot)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}

	return root.Agent, resp, nil
}

// Delete function deletes a Gen AI Agent by its corresponding UUID
func (s *AgentServiceOp) Delete(ctx context.Context, id string) (*Agent, *Response, error) {
	path := fmt.Sprintf("%s/%s", agentConnectBasePath, id)
	req, err := s.client.NewRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return nil, nil, err
	}

	root := new(genAIAgentRoot)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}

	return root.Agent, resp, nil
}

// Update function updates a Gen AI Agent status by changing visibility to public or private.
func (s *AgentServiceOp) UpdateVisibility(ctx context.Context, id string, update *AgentVisibilityUpdateRequest) (*Agent, *Response, error) {
	path := fmt.Sprintf("%s/%s/deployment_visibility", agentConnectBasePath, id)
	req, err := s.client.NewRequest(ctx, http.MethodPut, path, update)
	if err != nil {
		return nil, nil, err
	}

	root := new(genAIAgentRoot)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}

	return root.Agent, resp, nil
}

func (a Agents) String() string {
	return Stringify(a)
}

func (a Agent) String() string {
	return Stringify(a)
}
