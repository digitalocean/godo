package godo

import (
	"context"
	"fmt"
	"net/http"
)

const agentConnectBasePath = "/v2/gen-ai/agents"

type AgentService interface {
	List(context.Context, *AgentListOptions) ([]*Agent, *Response, error)
	Create(context.Context, *AgentCreateRequest) (*Agent, *Response, error)
	Get(context.Context, string) (*Agent, *Response, error)
	Update(context.Context, string, *AgentUpdateRequest) (*Agent, *Response, error)
	Delete(context.Context, string) (*Agent, *Response, error)
	ListVersions(context.Context, string, *ListOptions) ([]*AgentVersions, *Response, error)
	UpdateVersion(context.Context, string, *AgentVersionUpdateRequest) (*AgentVersionUpdateResponse, *Response, error)
	UpdateVisibility(context.Context, string, *AgentVisibilityUpdateRequest) (*Agent, *Response, error)
}

var _ AgentService = &AgentServiceOp{}

type AgentServiceOp struct {
	client *Client
}

type genAIAgentsRoot struct {
	Agents []*Agent `json:"agents"`
	Links  *Links   `json:"links"`
	Meta   *Meta    `json:"meta"`
}

type genAIAgentRoot struct {
	Agent *Agent `json:"agents"`
}

type genAIAgentAuditRoot struct {
	AgentVersion *AgentVersionUpdateResponse `json:"agent_version"`
}

type genAIAgentsVersionRoot struct {
	AgentVersions []*AgentVersions `json:"agent_versions"`
	Links         *Links           `json:"links"`
	Meta          *Meta            `json:"meta"`
}

type Agent struct {
	AnthropicApiKey    *Info                `json:"anthropic_api_key,omitempty"`
	ApiKeyInfos        *Info                `json:"api_key_infos,omitempty"`
	ApiKeys            []*ApiKeys           `json:"api_keys,omitempty"`
	ChatBot            *ChatBot             `json:"chatbot,omitempty"`
	ChatbotIdentifiers []ChatbotIdentifiers `json:"chatbot_identifiers,omitempty"`
	Name               string               `json:"name,omitempty"`
	CreatedAt          string               `json:"created_at,omitempty"`
	UpdateAt           string               `json:"updated_at,omitempty"`
	Instruction        string               `json:"instruction,omitempty"`
	Descripton         string               `json:"description,omitempty"`
	IfCase             string               `json:"if_case,omitempty"`
	K                  int                  `json:"k,omitempty"`
	MaxToken           int                  `json:"max_tokens,omitempty"`
	ProjectId          string               `json:"project_id,omitempty"`
	Region             string               `json:"region,omitempty"`
	RetrievalMethod    string               `json:"retrieval_method,omitempty"`
	RouteCreatedAt     string               `json:"route_created_at,omitempty"`
	RouteCreatedBy     string               `json:"route_created_by,omitempty"`
	RouteUuid          string               `json:"route_uuid,omitempty"`
	RouteName          string               `json:"route_name,omitempty"`
	Model              *Model               `json:"model,omitempty"`
	Deployment         *AgentDeployment     `json:"deployment,omitempty"`
	Tags               []string             `json:"tags,omitempty"`
	Template           *AgentTemplate       `json:"template,omitempty"`
	Temperature        float64              `json:"temperature,omitempty"`
	TopP               float64              `json:"top_p,omitempty"`
	Url                string               `json:"url,omitempty"`
	UserId             string               `json:"user_id,omitempty"`
	Uuid               string               `json:"uuid,omitempty"`
}

type AgentVersions struct {
	AgentUuid              string                    `json:"agent_uuid,omitempty"`
	AttachedChildAgents    []*AttachedChildAgents    `json:"attached_child_agents,omitempty"`
	AttachedFunctions      []*AttachedFunctions      `json:"attached_functions,omitempty"`
	AttachedGuardRails     []*AttachedGuardRails     `json:"attached_guardrails,omitempty"`
	AttachedKnowledgebases []*AttachedKnowledgebases `json:"attached_knowledgebases,omitempty"`
	CreatedAt              string                    `json:"created_at,omitempty"`
	CreatingUserEmail      string                    `json:"creating_user_email,omitempty"`
	CurrentlyApplied       bool                      `json:"currently_applied,omitempty"`
	Id                     string                    `json:"id,omitempty"`
	Descripton             string                    `json:"description,omitempty"`
	Instruction            string                    `json:"instruction,omitempty"`
	K                      int                       `json:"k,omitempty"`
	MaxToken               int                       `json:"max_tokens,omitempty"`
	ModelName              string                    `json:"model_name,omitempty"`
	Name                   string                    `json:"name,omitempty"`
	RetrievalMethod        string                    `json:"retrieval_method,omitempty"`
	Tags                   []string                  `json:"tags,omitempty"`
	Temperature            float64                   `json:"temperature,omitempty"`
	TopP                   float64                   `json:"top_p,omitempty"`
	TriggerAction          string                    `json:"trigger_action,omitempty"`
	VersionHash            string                    `json:"version_hash,omitempty"`
}

type AttachedChildAgents struct {
	AgentName      string `json:"agent_name,omitempty"`
	ChildAgentUuid string `json:"child_agent_uuid,omitempty"`
	IfCase         string `json:"if_case,omitempty"`
	IsDeleted      bool   `json:"is_deleted,omitempty"`
	RouteName      string `json:"route_name,omitempty"`
}

type AttachedFunctions struct {
	Description   string `json:"description,omitempty"`
	FaasName      string `json:"faas_name,omitempty"`
	FaasNamespace string `json:"faas_namespace,omitempty"`
	IsDeleted     bool   `json:"is_deleted,omitempty"`
	Name          string `json:"name,omitempty"`
}

type AttachedGuardRails struct {
	IsDeleted bool   `json:"is_deleted,omitempty"`
	Name      string `json:"name,omitempty"`
	Priority  string `json:"priority,omitempty"`
	Uuid      string `json:"uuid,omitempty"`
}

type AttachedKnowledgebases struct {
	IsDeleted bool   `json:"is_deleted,omitempty"`
	Name      string `json:"name,omitempty"`
	Uuid      string `json:"uuid,omitempty"`
}

type ApiKeys struct {
	ApiKey string `json:"api_key,omitempty"`
}

type Info struct {
	CreatedAt string `json:"created_at,omitempty"`
	CreatedBy string `json:"created_by,omitempty"`
	DeletedAt string `json:"deleted_at,omitempty"`
	Name      string `json:"name,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
	Uuid      string `json:"uuid,omitempty"`
}

type AgentVisibilityUpdateRequest struct {
	Uuid       string `json:"uuid,omitempty"`
	Visibility string `json:"visibility,omitempty"`
}

type AgentTemplate struct {
	CreatedAt      string            `json:"created_at,omitempty"`
	Instruction    string            `json:"instruction,omitempty"`
	Descripton     string            `json:"description,omitempty"`
	K              int               `json:"k,omitempty"`
	KnowledgeBases []*KnowledgeBases `json:"knowledge_bases,omitempty"`
	MaxToken       int               `json:"max_tokens,omitempty"`
	Model          *Model            `json:"model,omitempty"`
	Name           string            `json:"name,omitempty"`
	Temperature    float64           `json:"temperature,omitempty"`
	TopP           float64           `json:"top_p,omitempty"`
	UpdateAt       string            `json:"updated_at,omitempty"`
	Uuid           string            `json:"uuid,omitempty"`
}

type KnowledgeBases struct {
	AddedToAgentAt     string           `json:"added_to_agent_at,omitempty"`
	CreatedAt          string           `json:"created_at,omitempty"`
	DatabaseId         string           `json:"database_id,omitempty"`
	EmbeddingModelUuid string           `json:"embedding_model_uuid,omitempty"`
	IsPublic           bool             `json:"is_public,omitempty"`
	LastIndexingJob    *LastIndexingJob `json:"last_indexing_job,omitempty"`
	Name               string           `json:"name,omitempty"`
	ProjectId          string           `json:"project_id,omitempty"`
	Region             string           `json:"region,omitempty"`
	Tags               []string         `json:"tags,omitempty"`
	UpdateAt           string           `json:"updated_at,omitempty"`
	UserId             string           `json:"user_id,omitempty"`
	Uuid               string           `json:"uuid,omitempty"`
}

type LastIndexingJob struct {
	CompletedDatasources int      `json:"completed_datasources,omitempty"`
	CreatedAt            string   `json:"created_at,omitempty"`
	DataSourceUuids      []string `json:"data_source_uuids,omitempty"`
	FinishedAt           string   `json:"finished_at,omitempty"`
	KnowledgeBaseUuid    string   `json:"knowledge_base_uuid,omitempty"`
	Phase                string   `json:"phase,omitempty"`
	StartedAt            string   `json:"started_at,omitempty"`
	Tokens               int      `json:"tokens,omitempty"`
	TotalDatasources     string   `json:"total_datasources,omitempty"`
	UpdatedAt            string   `json:"updated_at,omitempty"`
	Uuid                 string   `json:"uuid,omitempty"`
}

type ChatbotIdentifiers struct {
	AgentChatbotIdentifier string `json:"agent_chatbot_identifier,omitempty"`
}

type AgentDeployment struct {
	CreatedAt  string `json:"created_at,omitempty"`
	Name       string `json:"name,omitempty"`
	Status     string `json:"status,omitempty"`
	UpdatedAt  string `json:"updated_at,omitempty"`
	Url        string `json:"url,omitempty"`
	Uuid       string `json:"uuid,omitempty"`
	Visibility string `json:"visibility,omitempty"`
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
	Agreement        *Agreement `json:"agreement,omitempty"`
	CreatedAt        string     `json:"created_at,omitempty"`
	InferenceName    string     `json:"inference_name,omitempty"`
	InferenceVersion string     `json:"inference_version,omitempty"`
	IsFoundational   bool       `json:"is_foundational,omitempty"`
	// Metadata         string    `json:"metadata,omitempty"` doubt
	Name           string   `json:"name,omitempty"`
	ParentUuid     string   `json:"parent_uuid,omitempty"`
	Provider       string   `json:"provider,omitempty"`
	UpdatedAt      string   `json:"updated_at,omitempty"`
	UploadComplete bool     `json:"upload_complete,omitempty"`
	Url            string   `json:"url,omitempty"`
	Usecases       []string `json:"usecases,omitempty"`
	Uuid           string   `json:"uuid,omitempty"`
	Version        *Version `json:"version,omitempty"`
}

type Agreement struct {
	Description string `json:"description,omitempty"`
	Name        string `json:"name,omitempty"`
	Url         string `json:"url,omitempty"`
	Uuid        string `json:"uuid,omitempty"`
}

type Version struct {
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
	RetrievalMethod   string   `json:"retrieval_method,omitempty"`
	Tags              []string `json:"tags,omitempty"`
}

type AgentUpdateRequest struct {
	AnthropicKeyUuid string   `json:"anthropic_key_uuid,omitempty"`
	Description      string   `json:"description,omitempty"`
	Instruction      string   `json:"instruction,omitempty"`
	K                int      `json:"k,omitempty"`
	MaxToken         int      `json:"max_tokens,omitempty"`
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

type AgentVersionUpdateRequest struct {
	Uuid        string `json:"uuid,omitempty"`
	VersionHash string `json:"version_hash,omitempty"`
}

type AgentVersionUpdateResponse struct {
	AuditHeader *AuditHeader `json:"audit_header,omitempty"`
	VersionHash string       `json:"version_hash,omitempty"`
}

type AuditHeader struct {
	ActorId           string `json:"actor_id,omitempty"`
	ActorIp           string `json:"actor_ip,omitempty"`
	ActorUuid         string `json:"actor_uuid,omitempty"`
	ContextUrn        string `json:"context_urn,omitempty"`
	OriginApplication string `json:"origin_application,omitempty"`
	UserId            string `json:"user_id,omitempty"`
	UserUuid          string `json:"user_uuid,omitempty"`
}

func (s *AgentServiceOp) List(ctx context.Context, opt *AgentListOptions) ([]*Agent, *Response, error) {
	fmt.Println("Added options")
	path, err := addOptions(agentConnectBasePath, opt)
	if err != nil {
		return nil, nil, err
	}
	fmt.Println("Created a new request")
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
	fmt.Println("Response :- ")
	return root.Agents, resp, nil
}

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

// Get returns the details of a Gen AI Agent.
func (s *AgentServiceOp) Get(ctx context.Context, id string) (*Agent, *Response, error) {
	path := fmt.Sprintf("%s/%s", agentConnectBasePath, id)
	fmt.Println(path)
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

// Update updates a Gen AI Agent properties.
func (s *AgentServiceOp) Update(ctx context.Context, id string, update *AgentUpdateRequest) (*Agent, *Response, error) {
	path := fmt.Sprintf("%s/%s/", agentConnectBasePath, id)
	req, err := s.client.NewRequest(ctx, http.MethodPatch, path, update)
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

// Delete deletes a Gen AI Agent.
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

func (s *AgentServiceOp) ListVersions(ctx context.Context, id string, opt *ListOptions) ([]*AgentVersions, *Response, error) {
	path := fmt.Sprintf("%s/%s/versions", agentConnectBasePath, id)
	fmt.Println(path)
	path, err := addOptions(agentConnectBasePath, opt)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}
	root := new(genAIAgentsVersionRoot)
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
	return root.AgentVersions, resp, nil
}

func (s *AgentServiceOp) UpdateVersion(ctx context.Context, id string, update *AgentVersionUpdateRequest) (*AgentVersionUpdateResponse, *Response, error) {
	path := fmt.Sprintf("%s/%s/versions", agentConnectBasePath, id)
	req, err := s.client.NewRequest(ctx, http.MethodPut, path, update)
	if err != nil {
		return nil, nil, err
	}

	root := new(genAIAgentAuditRoot)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}

	return root.AgentVersion, resp, nil
}

// Update updates a Gen AI Agent status by changing visibility to public or private.
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
