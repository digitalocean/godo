package godo

import (
	"context"
	"fmt"
	"net/http"
)

const (
	agentConnectBasePath         = "/v2/gen-ai/agents"
	agentModelBasePath           = "/v2/gen-ai/models"
	GenAIConnectBasePath         = "/v2/gen-ai"
	KnowledgeBasePath            = GenAIConnectBasePath + "/knowledge_bases"
	KnowledgeBaseDataSourcesPath = KnowledgeBasePath + "/%s/data_sources"
	GetKnowledgeBaseByIDPath     = KnowledgeBasePath + "/%s"
	UpdateKnowledgeBaseByIDPath  = KnowledgeBasePath + "/%s"
	DeleteKnowledgeBaseByIDPath  = KnowledgeBasePath + "/%s"
	AgentKnowledgBasePath        = GenAIConnectBasePath + "/agents" + "/%s/knowledge_bases/%s"
	DeleteDataSourcePath         = KnowledgeBasePath + "/%s/data_sources/%s"
)

// GenAIService is an interface for interfacing with the Gen AI Agent endpoints
// of the DigitalOcean API.
// See https://docs.digitalocean.com/reference/api/digitalocean/#tag/GenAI-Platform-(Public-Preview) for more details.
type GenAIService interface {
	ListAgents(context.Context, *ListOptions) ([]*Agent, *Response, error)
	CreateAgent(context.Context, *AgentCreateRequest) (*Agent, *Response, error)
	GetAgent(context.Context, string) (*Agent, *Response, error)
	UpdateAgent(context.Context, string, *AgentUpdateRequest) (*Agent, *Response, error)
	DeleteAgent(context.Context, string) (*Agent, *Response, error)
	UpdateAgentVisibility(context.Context, string, *AgentVisibilityUpdateRequest) (*Agent, *Response, error)
	ListModels(context.Context, *ListOptions) ([]*Model, *Response, error)

	ListKnowledgeBases(ctx context.Context, opt *ListOptions) ([]KnowledgeBase, *Response, error)
	CreateKnowledgeBase(ctx context.Context, KnowledgeBaseCreate *KnowledgeBaseCreateRequest) (*KnowledgeBase, *Response, error)
	ListDataSources(ctx context.Context, knowledgeBaseID string, opt *ListOptions) ([]KnowledgeBaseDataSource, *Response, error)
	AddDataSource(ctx context.Context, knowledgeBaseID string, addDataSource *AddDataSourceRequest) (*KnowledgeBaseDataSource, *Response, error)
	DeleteDataSource(ctx context.Context, knowledgeBaseID string, DataSourceID string) (string, string, *Response, error)
	GetKnowledgeBase(ctx context.Context, knowledgeBaseID string) (*KnowledgeBase, *Response, error)
	UpdateKnowledgeBase(ctx context.Context, knowledgeBaseID string, update *UpdateKnowledgeBaseRequest) (*KnowledgeBase, *Response, error)
	DeleteKnowledgeBase(ctx context.Context, knowledgeBaseID string) (string, *Response, error)
	// AttachKnowledgBases(ctx context.Context, AgentID string, knowledgeBaseID string) (*Agent, *Response, error)
	AttachKnowledgBase(ctx context.Context, AgentID string, knowledgeBaseID string) (*Agent, *Response, error)
	DetachKnowledgBase(ctx context.Context, AgentID string, knowledgeBaseID string) (*Agent, *Response, error)
}

var _ GenAIService = &GenAIServiceOp{}

// GenAIServiceOp interfaces with the Agent Service endpoints in the DigitalOcean API.
type GenAIServiceOp struct {
	client *Client
}

type genAIAgentsRoot struct {
	Agents []*Agent `json:"agents"`
	Links  *Links   `json:"links"`
	Meta   *Meta    `json:"meta"`
}

type genAIAgentRoot struct {
	Agent *Agent `json:"agent"`
}

type genAiModelsRoot struct {
	Models []*Model `json:"models"`
	Links  *Links   `json:"links"`
	Meta   *Meta    `json:"meta"`
}

// Agent represents a Gen AI Agent
type Agent struct {
	AnthropicApiKey    *AnthropicApiKeyInfo      `json:"anthropic_api_key,omitempty"`
	ApiKeyInfos        []*ApiKeyInfo             `json:"api_key_infos,omitempty"`
	ApiKeys            []*ApiKey                 `json:"api_keys,omitempty"`
	ChatBot            *ChatBot                  `json:"chatbot,omitempty"`
	ChatbotIdentifiers []*AgentChatbotIdentifier `json:"chatbot_identifiers,omitempty"`
	CreatedAt          *Timestamp                `json:"created_at,omitempty"`
	ChildAgents        []*Agent                  `json:"child_agents,omitempty"`
	Deployment         *AgentDeployment          `json:"deployment,omitempty"`
	Description        string                    `json:"description,omitempty"`
	UpdatedAt          *Timestamp                `json:"updated_at,omitempty"`
	Functions          []*AgentFunction          `json:"functions,omitempty"`
	Guardrails         []*AgentGuardrail         `json:"guardrails,omitempty"`
	IfCase             string                    `json:"if_case,omitempty"`
	Instruction        string                    `json:"instruction,omitempty"`
	K                  int                       `json:"k,omitempty"`
	KnowledgeBases     []*KnowledgeBase          `json:"knowledge_bases,omitempty"`
	MaxTokens          int                       `json:"max_tokens,omitempty"`
	Model              *Model                    `json:"model,omitempty"`
	Name               string                    `json:"name,omitempty"`
	OpenAiApiKey       *OpenAiApiKey             `json:"open_ai_api_key,omitempty"`
	ParentAgents       []*Agent                  `json:"parent_agents,omitempty"`
	ProjectId          string                    `json:"project_id,omitempty"`
	Region             string                    `json:"region,omitempty"`
	RetrievalMethod    string                    `json:"retrieval_method,omitempty"`
	RouteCreatedAt     *Timestamp                `json:"route_created_at,omitempty"`
	RouteCreatedBy     string                    `json:"route_created_by,omitempty"`
	RouteUuid          string                    `json:"route_uuid,omitempty"`
	RouteName          string                    `json:"route_name,omitempty"`
	Tags               []string                  `json:"tags,omitempty"`
	Template           *AgentTemplate            `json:"template,omitempty"`
	Temperature        float64                   `json:"temperature,omitempty"`
	TopP               float64                   `json:"top_p,omitempty"`
	Url                string                    `json:"url,omitempty"`
	UserId             string                    `json:"user_id,omitempty"`
	Uuid               string                    `json:"uuid,omitempty"`
}

// AgentFunction represents a Gen AI Agent Function
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

// AgentGuardrail represents a Guardrail attached to Gen AI Agent
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

// AnthropicApiKeyInfo represents the Anthropic API Key information
type AnthropicApiKeyInfo struct {
	CreatedAt *Timestamp `json:"created_at,omitempty"`
	CreatedBy string     `json:"created_by,omitempty"`
	DeletedAt *Timestamp `json:"deleted_at,omitempty"`
	Name      string     `json:"name,omitempty"`
	UpdatedAt *Timestamp `json:"updated_at,omitempty"`
	Uuid      string     `json:"uuid,omitempty"`
}

// ApiKeyInfo represents the information of an API key
type ApiKeyInfo struct {
	CreatedAt *Timestamp `json:"created_at,omitempty"`
	CreatedBy string     `json:"created_by,omitempty"`
	DeletedAt *Timestamp `json:"deleted_at,omitempty"`
	Name      string     `json:"name,omitempty"`
	SecretKey string     `json:"secret_key,omitempty"`
	Uuid      string     `json:"uuid,omitempty"`
}

// OpenAiApiKey represents the OpenAI API Key information
type OpenAiApiKey struct {
	CreatedAt *Timestamp `json:"created_at,omitempty"`
	CreatedBy string     `json:"created_by,omitempty"`
	DeletedAt *Timestamp `json:"deleted_at,omitempty"`
	Models    []*Model   `json:"models,omitempty"`
	Name      string     `json:"name,omitempty"`
	UpdatedAt *Timestamp `json:"updated_at,omitempty"`
	Uuid      string     `json:"uuid,omitempty"`
}

// AgentVersionUpdateRequest represents the request to update the version of an agent
type AgentVisibilityUpdateRequest struct {
	Uuid       string `json:"uuid,omitempty"`
	Visibility string `json:"visibility,omitempty"`
}

// AgentTemplate represents the template of a Gen AI Agent
type AgentTemplate struct {
	CreatedAt      *Timestamp       `json:"created_at,omitempty"`
	Instruction    string           `json:"instruction,omitempty"`
	Description    string           `json:"description,omitempty"`
	K              int              `json:"k,omitempty"`
	KnowledgeBases []*KnowledgeBase `json:"knowledge_bases,omitempty"`
	MaxTokens      int              `json:"max_tokens,omitempty"`
	Model          *Model           `json:"model,omitempty"`
	Name           string           `json:"name,omitempty"`
	Temperature    float64          `json:"temperature,omitempty"`
	TopP           float64          `json:"top_p,omitempty"`
	UpdatedAt      *Timestamp       `json:"updated_at,omitempty"`
	Uuid           string           `json:"uuid,omitempty"`
}

type AgentChatbotIdentifier struct {
	AgentChatbotIdentifier string `json:"agent_chatbot_identifier,omitempty"`
}

// AgentDeployment represents the deployment information of a Gen AI Agent
type AgentDeployment struct {
	CreatedAt  *Timestamp `json:"created_at,omitempty"`
	Name       string     `json:"name,omitempty"`
	Status     string     `json:"status,omitempty"`
	UpdatedAt  *Timestamp `json:"updated_at,omitempty"`
	Url        string     `json:"url,omitempty"`
	Uuid       string     `json:"uuid,omitempty"`
	Visibility string     `json:"visibility,omitempty"`
}

// ChatBot represents the chatbot information of a Gen AI Agent
type ChatBot struct {
	ButtonBackgroundColor string `json:"button_background_color,omitempty"`
	Logo                  string `json:"logo,omitempty"`
	Name                  string `json:"name,omitempty"`
	PrimaryColor          string `json:"primary_color,omitempty"`
	SecondaryColor        string `json:"secondary_color,omitempty"`
	StartingMessage       string `json:"starting_message,omitempty"`
}

// Model represents a Gen AI Model
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

// Agreement represents the agreement information of a Gen AI Model
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

// AgentCreateRequest represents the request to create a new Gen AI Agent
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

// AgentUpdateRequest represents the request to update an existing Gen AI Agent
type AgentUpdateRequest struct {
	AnthropicKeyUuid string   `json:"anthropic_key_uuid,omitempty"`
	Description      string   `json:"description,omitempty"`
	Instruction      string   `json:"instruction,omitempty"`
	K                int      `json:"k,omitempty"`
	MaxTokens        int      `json:"max_tokens,omitempty"`
	ModelUuid        string   `json:"model_uuid,omitempty"`
	Name             string   `json:"name,omitempty"`
	OpenAiKeyUuid    string   `json:"open_ai_key_uuid,omitempty"`
	ProjectId        string   `json:"project_id,omitempty"`
	RetrievalMethod  string   `json:"retrieval_method,omitempty"`
	Region           string   `json:"region,omitempty"`
	Tags             []string `json:"tags,omitempty"`
	Temperature      float64  `json:"temperature,omitempty"`
	TopP             float64  `json:"top_p,omitempty"`
	Uuid             string   `json:"uuid,omitempty"`
}

type genAIAgentKBRoot struct {
	Agent *Agent `json:"agent"`
}

type genAIAgentAuditRoot struct {
	AgentVersion *AgentVersionUpdateResponse `json:"agent_version"`
}

type genAIAgentsVersionRoot struct {
	AgentVersions []*AgentVersions `json:"agent_versions"`
	Links         *Links           `json:"links"`
	Meta          *Meta            `json:"meta"`
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

// KnowledgeBase represents a Gen AI Knowledge Base
type KnowledgeBase struct {
	AddedToAgentAt     *Timestamp       `json:"added_to_agent_at,omitempty"`
	CreatedAt          *Timestamp       `json:"created_at,omitempty"`
	DatabaseId         string           `json:"database_id,omitempty"`
	EmbeddingModelUUID string           `json:"embedding_model_uuid,omitempty"`
	IsPublic           bool             `json:"is_public,omitempty"`
	LastIndexingJob    *LastIndexingJob `json:"last_indexing_job,omitempty"`
	Name               string           `json:"name,omitempty"`
	ProjectId          string           `json:"project_id,omitempty"`
	Region             string           `json:"region,omitempty"`
	Tags               []string         `json:"tags,omitempty"`
	UpdatedAt          *Timestamp       `json:"updated_at,omitempty"`
	UserId             string           `json:"user_id,omitempty"`
	UUID               string           `json:"uuid,omitempty"`
}

// LastIndexingJob represents the last indexing job description of a Gen AI Knowledge Base
type LastIndexingJob struct {
	CompletedDatasources int        `json:"completed_datasources,omitempty"`
	CreatedAt            *Timestamp `json:"created_at,omitempty"`
	DataSourceUUIDs      []string   `json:"data_source_uuids,omitempty"`
	FinishedAt           *Timestamp `json:"finished_at,omitempty"`
	KnowledgeBaseUUID    string     `json:"knowledge_base_uuid,omitempty"`
	Phase                string     `json:"phase,omitempty"`
	StartedAt            *Timestamp `json:"started_at,omitempty"`
	Tokens               int        `json:"tokens,omitempty"`
	TotalDatasources     int        `json:"total_datasources,omitempty"`
	UpdatedAt            *Timestamp `json:"updated_at,omitempty"`
	UUID                 string     `json:"uuid,omitempty"`
}

type KnowledgeBaseDataSource struct {
	BucketName           string                `json:"bucket_name,omitempty"`
	CreatedAt            *Timestamp            `json:"created_at,omitempty"`
	FileUploadDataSource *FileUploadDataSource `json:"file_upload_data_source,omitempty"`
	ItemPath             string                `json:"item_path,omitempty"`
	LastIndexingJob      *LastIndexingJob      `json:"last_indexing_job,omitempty"`
	Region               string                `json:"region,omitempty"`
	SpacesDataSource     *SpacesDataSource     `json:"spaces_data_source,omitempty"`
	UpdatedAt            *Timestamp            `json:"updated_at,omitempty"`
	UUID                 string                `json:"uuid,omitempty"`
	WebCrawlerDataSource *WebCrawlerDataSource `json:"web_crawler_data_source,omitempty"`
}

type WebCrawlerDataSource struct {
	BaseUrl        string `json:"base_url"`
	CrawlingOption string `json:"crawling_option"`
	EmbedMedia     bool   `json:"embed_media"`
}

type SpacesDataSource struct {
	BucketName string `json:"bucket_name"`
	ItemPath   string `json:"item_path"`
	Region     string `json:"region"`
}

type FileUploadDataSource struct {
	OriginalFileName string `json:"original_file_name"`
	Size             string `json:"size_in_bytes"`
	StoredObjectKey  string `json:"stored_object_key"`
}

type KnowledgeBaseDataSourcesRoot struct {
	KnowledgeBaseDatasources []KnowledgeBaseDataSource `json:"knowledge_base_data_sources"`
	Links                    *Links                    `json:"links"`
	Meta                     *Meta                     `json:"meta"`
}
type SingleKnowledgeBaseDataSourceRoot struct {
	KnowledgeBaseDatasource *KnowledgeBaseDataSource `json:"knowledge_base_data_source"`
	Links                   *Links                   `json:"links"`
	Meta                    *Meta                    `json:"meta"`
}
type knowledgebasesRoot struct {
	KnowledgeBases []KnowledgeBase `json:"knowledge_bases"`
	Links          *Links          `json:"links"`
	Meta           *Meta           `json:"meta"`
}

type knowledgebaseRoot struct {
	KnowledgeBase  *KnowledgeBase `json:"knowledge_base"`
	DatabaseStatus string         `json:"database_status,omitempty"`
}

// - updated by adding omitempty above
// type knowledgebaseRoots struct {
// 	DatabaseStatus string         `json:"database_status"`
// 	KnowledgeBase  *KnowledgeBase `json:"knowledge_base"`
// }

type KnowledgeBaseCreateRequest struct {
	DatabaseID         string                    `json:"database_id"`
	DataSources        []KnowledgeBaseDataSource `json:"datasources"`
	EmbeddingModelUUID string                    `json:"embedding_model_uuid"`
	Name               string                    `json:"name"`
	ProjectID          string                    `json:"project_id"`
	Region             string                    `json:"region"`
	Tags               []string                  `json:"tags"`
	VPCUUIUD           string                    `json:"vpc_uuid"`
}

type DeleteDataSourceRoot struct {
	DataSourceUUID    string `json:"data_source_uuid"`
	KnowledgeBaseUUID string `json:"knowledge_base_uuid"`
}

type DeleteKnowledgeBaseRoot struct {
	KnowledgeBaseUUID string `json:"uuid"`
}

type DeletedKnowledgeBaseResponse struct {
	DataSourceUUID    string `json:"data_source_uuid"`
	KnowledgeBaseUUID string `json:"knowledge_base_uuid"`
}

type AddDataSourceRequest struct {
	KnowledgeBaseUUID    string                `json:"knowledge_base_uuid"`
	SpacesDataSource     *SpacesDataSource     `json:"spaces_data_source"`
	WebCrawlerDataSource *WebCrawlerDataSource `json:"web_crawler_data_source"`
}

type UpdateKnowledgeBaseRequest struct {
	DatabaseID         string   `json:"database_id"`
	EmbeddingModelUUID string   `json:"embedding_model_uuid"`
	Name               string   `json:"name"`
	ProjectID          string   `json:"project_id"`
	Tags               []string `json:"tags"`
	UUID               string   `json:"uuid"`
}

type ChatbotIdentifiers struct {
	AgentChatbotIdentifier string `json:"agent_chatbot_identifier,omitempty"`
}

type Version struct {
	Major int `json:"major,omitempty"`
	Minor int `json:"minor,omitempty"`
	Patch int `json:"patch,omitempty"`
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

// List returns a list of Gen AI Agents
func (s *GenAIServiceOp) ListAgents(ctx context.Context, opt *ListOptions) ([]*Agent, *Response, error) {
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
func (s *GenAIServiceOp) CreateAgent(ctx context.Context, create *AgentCreateRequest) (*Agent, *Response, error) {
	path := agentConnectBasePath
	if create.ProjectId == "" {
		return nil, nil, fmt.Errorf("Project ID is required")
	}
	if create.Region == "" {
		return nil, nil, fmt.Errorf("Region is required")
	}
	if create.Instruction == "" {
		return nil, nil, fmt.Errorf("Instruction is required")
	}
	if create.ModelUuid == "" {
		return nil, nil, fmt.Errorf("ModelUuid is required")
	}
	if create.Name == "" {
		return nil, nil, fmt.Errorf("Name is required")
	}
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
func (s *GenAIServiceOp) GetAgent(ctx context.Context, id string) (*Agent, *Response, error) {
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
func (s *GenAIServiceOp) UpdateAgent(ctx context.Context, id string, update *AgentUpdateRequest) (*Agent, *Response, error) {
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
func (s *GenAIServiceOp) DeleteAgent(ctx context.Context, id string) (*Agent, *Response, error) {
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
func (s *GenAIServiceOp) UpdateAgentVisibility(ctx context.Context, id string, update *AgentVisibilityUpdateRequest) (*Agent, *Response, error) {
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

// ListModels function returns a list of Gen AI Models
func (s *GenAIServiceOp) ListModels(ctx context.Context, opt *ListOptions) ([]*Model, *Response, error) {
	path, err := addOptions(agentModelBasePath, opt)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}
	root := new(genAiModelsRoot)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}
	if l := root.Links; l != nil {
		resp.Links = l
	}

	return root.Models, resp, nil
}

func (s *GenAIServiceOp) List(ctx context.Context, opt *AgentListOptions) ([]*Agent, *Response, error) {
	fmt.Println("Added options")
	path, err := addOptions(GenAIConnectBasePath, opt)
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

func (s *GenAIServiceOp) Create(ctx context.Context, create *AgentCreateRequest) (*Agent, *Response, error) {
	path := GenAIConnectBasePath

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
func (s *GenAIServiceOp) Get(ctx context.Context, id string) (*Agent, *Response, error) {
	path := fmt.Sprintf("%s/%s", GenAIConnectBasePath, id)
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
func (s *GenAIServiceOp) Update(ctx context.Context, id string, update *AgentUpdateRequest) (*Agent, *Response, error) {
	path := fmt.Sprintf("%s/%s/", GenAIConnectBasePath, id)
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
func (s *GenAIServiceOp) Delete(ctx context.Context, id string) (*Agent, *Response, error) {
	path := fmt.Sprintf("%s/%s", GenAIConnectBasePath, id)
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

func (s *GenAIServiceOp) ListVersions(ctx context.Context, id string, opt *ListOptions) ([]*AgentVersions, *Response, error) {
	path := fmt.Sprintf("%s/%s/versions", GenAIConnectBasePath, id)
	fmt.Println(path)
	path, err := addOptions(GenAIConnectBasePath, opt)
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

func (s *GenAIServiceOp) UpdateVersion(ctx context.Context, id string, update *AgentVersionUpdateRequest) (*AgentVersionUpdateResponse, *Response, error) {
	path := fmt.Sprintf("%s/%s/versions", GenAIConnectBasePath, id)
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
func (s *GenAIServiceOp) UpdateVisibility(ctx context.Context, id string, update *AgentVisibilityUpdateRequest) (*Agent, *Response, error) {
	path := fmt.Sprintf("%s/%s/deployment_visibility", GenAIConnectBasePath, id)
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

// List all knowledge bases
func (s *GenAIServiceOp) ListKnowledgeBases(ctx context.Context, opt *ListOptions) ([]KnowledgeBase, *Response, error) {

	path := KnowledgeBasePath
	path, err := addOptions(path, opt)
	if err != nil {
		return nil, nil, err
	}
	req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	root := new(knowledgebasesRoot)
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
	return root.KnowledgeBases, resp, err
}

// Create a knowledge base
func (s *GenAIServiceOp) CreateKnowledgeBase(ctx context.Context, KnowledgeBaseCreate *KnowledgeBaseCreateRequest) (*KnowledgeBase, *Response, error) {
	///v2/gen-ai/knowledge_bases.
	path := KnowledgeBasePath
	req, err := s.client.NewRequest(ctx, http.MethodPost, path, KnowledgeBaseCreate)
	if err != nil {
		return nil, nil, err
	}
	root := new(knowledgebaseRoot)
	resp, err := s.client.Do(ctx, req, root)

	if err != nil {
		return nil, resp, err
	}

	return root.KnowledgeBase, resp, err
}

// List Data Sources for a Knowledge Base
func (s *GenAIServiceOp) ListDataSources(ctx context.Context, knowledgeBaseID string, opt *ListOptions) ([]KnowledgeBaseDataSource, *Response, error) {

	path := fmt.Sprintf(KnowledgeBaseDataSourcesPath, knowledgeBaseID)
	path, err := addOptions(path, opt)
	if err != nil {
		return nil, nil, err
	}
	req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}
	root := new(KnowledgeBaseDataSourcesRoot)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, nil, err
	}
	if l := root.Links; l != nil {
		resp.Links = l
	}
	if m := root.Meta; m != nil {
		resp.Meta = m
	}
	return root.KnowledgeBaseDatasources, resp, err
}

// Add Data Source to a Knowledge Base
func (s *GenAIServiceOp) AddDataSource(ctx context.Context, knowledgeBaseID string, addDataSource *AddDataSourceRequest) (*KnowledgeBaseDataSource, *Response, error) {
	path := fmt.Sprintf(KnowledgeBaseDataSourcesPath, knowledgeBaseID)
	req, err := s.client.NewRequest(ctx, http.MethodPost, path, addDataSource)
	if err != nil {
		return nil, nil, err
	}
	root := new(SingleKnowledgeBaseDataSourceRoot)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}
	return root.KnowledgeBaseDatasource, resp, err
}

// Delete data source from a knowledge base
// confirm if strings are required in response arguments
func (s *GenAIServiceOp) DeleteDataSource(ctx context.Context, knowledgeBaseID string, DataSourceID string) (string, string, *Response, error) {

	path := fmt.Sprintf(DeleteDataSourcePath, knowledgeBaseID, DataSourceID)
	req, err := s.client.NewRequest(ctx, http.MethodDelete, path, nil)

	if err != nil {
		return "", "", nil, err
	}

	root := new(DeleteDataSourceRoot)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return "", "", resp, err

	}
	return root.KnowledgeBaseUUID, root.DataSourceUUID, resp, nil
}

// Get a KnowledgeBase
func (s *GenAIServiceOp) GetKnowledgeBase(ctx context.Context, knowledgeBaseID string) (*KnowledgeBase, *Response, error) {
	path := fmt.Sprintf(GetKnowledgeBaseByIDPath, knowledgeBaseID)
	req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)

	if err != nil {
		return nil, nil, err
	}
	root := new(knowledgebaseRoot)
	resp, err := s.client.Do(ctx, req, root)

	if err != nil {
		return nil, resp, err
	}
	return root.KnowledgeBase, resp, nil
}

// Update a knowledge base
func (s *GenAIServiceOp) UpdateKnowledgeBase(ctx context.Context, knowledgeBaseID string, update *UpdateKnowledgeBaseRequest) (*KnowledgeBase, *Response, error) {
	path := fmt.Sprintf(UpdateKnowledgeBaseByIDPath, knowledgeBaseID)
	req, err := s.client.NewRequest(ctx, http.MethodPut, path, update)
	if err != nil {
		return nil, nil, err
	}

	root := new(knowledgebaseRoot)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}

	return root.KnowledgeBase, resp, nil
}

// Delete a knowledge base
func (s *GenAIServiceOp) DeleteKnowledgeBase(ctx context.Context, knowledgeBaseID string) (string, *Response, error) {

	path := fmt.Sprintf(DeleteKnowledgeBaseByIDPath, knowledgeBaseID)
	req, err := s.client.NewRequest(ctx, http.MethodDelete, path, nil)
	fmt.Print(path)
	if err != nil {
		return "", nil, err
	}
	root := new(DeleteKnowledgeBaseRoot)
	resp, err := s.client.Do(ctx, req, root)

	if err != nil {
		return "", resp, err
	}
	return root.KnowledgeBaseUUID, resp, nil
}

// Attach a knowledge base to an agent
func (s *GenAIServiceOp) AttachKnowledgBase(ctx context.Context, AgentID string, knowledgeBaseID string) (*Agent, *Response, error) {
	path := fmt.Sprintf(AgentKnowledgBasePath, AgentID, knowledgeBaseID)
	fmt.Println(path)
	req, err := s.client.NewRequest(ctx, http.MethodPost, path, nil)
	if err != nil {
		return nil, nil, err
	}

	root := new(genAIAgentKBRoot)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}

	return root.Agent, resp, nil
}

// Detach a knowledge base from an agent
func (s *GenAIServiceOp) DetachKnowledgBase(ctx context.Context, AgentID string, knowledgeBaseID string) (*Agent, *Response, error) {
	path := fmt.Sprintf(AgentKnowledgBasePath, AgentID, knowledgeBaseID)
	fmt.Println("Constructed Path:", path)
	req, err := s.client.NewRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return nil, nil, err
	}
	root := new(genAIAgentKBRoot)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}
	return root.Agent, resp, nil
}

func (a Agent) String() string {
	return Stringify(a)
}

func (m Model) String() string {
	return Stringify(m)
}

func (a KnowledgeBase) String() string {
	return Stringify(a)
}

func (a KnowledgeBaseDataSource) String() string {
	return Stringify(a)
}

// func (s *AgentServiceOp) AttachKnowledgBases(ctx context.Context, AgentID string, knowledgeBaseID string) (*Agent, *Response, error) {
// 	path := fmt.Sprintf("s") // to do
// 	fmt.Println(path)
// 	req, err := s.client.NewRequest(ctx, http.MethodPost, path, nil)
// 	if err != nil {
// 		return nil, nil, err
// 	}

// 	root := new(genAIAgentsRoot)
// 	resp, err := s.client.Do(ctx, req, root)
// 	if err != nil {
// 		return nil, resp, err
// 	}

// 	return root.Agent, resp, nil
// }
