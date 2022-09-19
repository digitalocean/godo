package godo

import (
	"context"
	"errors"
	"fmt"
	"net/http"
)

const (
	appsBasePath = "/v2/apps"
)

// AppLogType is the type of app logs.
type AppLogType string

const (
	// AppLogTypeBuild represents build logs.
	AppLogTypeBuild AppLogType = "BUILD"
	// AppLogTypeDeploy represents deploy logs.
	AppLogTypeDeploy AppLogType = "DEPLOY"
	// AppLogTypeRun represents run logs.
	AppLogTypeRun AppLogType = "RUN"
)

// AppsService is an interface for interfacing with the App Platform endpoints
// of the DigitalOcean API.
type AppsService interface {
	Create(ctx context.Context, create *AppCreateRequest) (*App, *Response, error)
	Get(ctx context.Context, appID string) (*App, *Response, error)
	List(ctx context.Context, opts *ListOptions) ([]*App, *Response, error)
	Update(ctx context.Context, appID string, update *AppUpdateRequest) (*App, *Response, error)
	Delete(ctx context.Context, appID string) (*Response, error)
	Propose(ctx context.Context, propose *AppProposeRequest) (*AppProposeResponse, *Response, error)

	GetDeployment(ctx context.Context, appID, deploymentID string) (*Deployment, *Response, error)
	ListDeployments(ctx context.Context, appID string, opts *ListOptions) ([]*Deployment, *Response, error)
	CreateDeployment(ctx context.Context, appID string, create ...*DeploymentCreateRequest) (*Deployment, *Response, error)

	GetLogs(ctx context.Context, appID, deploymentID, component string, logType AppLogType, follow bool, tailLines int) (*AppLogs, *Response, error)

	ListRegions(ctx context.Context) ([]*AppRegion, *Response, error)

	ListTiers(ctx context.Context) ([]*AppTier, *Response, error)
	GetTier(ctx context.Context, slug string) (*AppTier, *Response, error)

	ListInstanceSizes(ctx context.Context) ([]*AppInstanceSize, *Response, error)
	GetInstanceSize(ctx context.Context, slug string) (*AppInstanceSize, *Response, error)

	ListAlerts(ctx context.Context, appID string) ([]*AppAlert, *Response, error)
	UpdateAlertDestinations(ctx context.Context, appID, alertID string, update *AlertDestinationUpdateRequest) (*AppAlert, *Response, error)

	Detect(ctx context.Context, detect *DetectRequest) (*DetectResponse, *Response, error)
}

// AppLogs represent app logs.
type AppLogs struct {
	LiveURL      string   `json:"live_url"`
	HistoricURLs []string `json:"historic_urls"`
}

// AppUpdateRequest represents a request to update an app.
type AppUpdateRequest struct {
	Spec *AppSpec `json:"spec"`
}

// DeploymentCreateRequest represents a request to create a deployment.
type DeploymentCreateRequest struct {
	ForceBuild bool `json:"force_build"`
}

// AlertDestinationUpdateRequest represents a request to update alert destinations.
type AlertDestinationUpdateRequest struct {
	Emails        []string                `json:"emails"`
	SlackWebhooks []*AppAlertSlackWebhook `json:"slack_webhooks"`
}

type appRoot struct {
	App *App `json:"app"`
}

type appsRoot struct {
	Apps  []*App `json:"apps"`
	Links *Links `json:"links"`
	Meta  *Meta  `json:"meta"`
}

type deploymentRoot struct {
	Deployment *Deployment `json:"deployment"`
}

type deploymentsRoot struct {
	Deployments []*Deployment `json:"deployments"`
	Links       *Links        `json:"links"`
	Meta        *Meta         `json:"meta"`
}

type appTierRoot struct {
	Tier *AppTier `json:"tier"`
}

type appTiersRoot struct {
	Tiers []*AppTier `json:"tiers"`
}

type instanceSizeRoot struct {
	InstanceSize *AppInstanceSize `json:"instance_size"`
}

type instanceSizesRoot struct {
	InstanceSizes []*AppInstanceSize `json:"instance_sizes"`
}

type appRegionsRoot struct {
	Regions []*AppRegion `json:"regions"`
}

type appAlertsRoot struct {
	Alerts []*AppAlert `json:"alerts"`
}

type appAlertRoot struct {
	Alert *AppAlert `json:"alert"`
}

// AppsServiceOp handles communication with Apps methods of the DigitalOcean API.
type AppsServiceOp struct {
	client *Client
}

// URN returns a URN identifier for the app
func (a App) URN() string {
	return ToURN("app", a.ID)
}

// Create an app.
func (s *AppsServiceOp) Create(ctx context.Context, create *AppCreateRequest) (*App, *Response, error) {
	path := appsBasePath
	req, err := s.client.NewRequest(ctx, http.MethodPost, path, create)
	if err != nil {
		return nil, nil, err
	}

	root := new(appRoot)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}
	return root.App, resp, nil
}

// Get an app.
func (s *AppsServiceOp) Get(ctx context.Context, appID string) (*App, *Response, error) {
	path := fmt.Sprintf("%s/%s", appsBasePath, appID)
	req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}
	root := new(appRoot)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}
	return root.App, resp, nil
}

// List apps.
func (s *AppsServiceOp) List(ctx context.Context, opts *ListOptions) ([]*App, *Response, error) {
	path := appsBasePath
	path, err := addOptions(path, opts)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}
	root := new(appsRoot)
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

	return root.Apps, resp, nil
}

// Update an app.
func (s *AppsServiceOp) Update(ctx context.Context, appID string, update *AppUpdateRequest) (*App, *Response, error) {
	path := fmt.Sprintf("%s/%s", appsBasePath, appID)
	req, err := s.client.NewRequest(ctx, http.MethodPut, path, update)
	if err != nil {
		return nil, nil, err
	}

	root := new(appRoot)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}
	return root.App, resp, nil
}

// Delete an app.
func (s *AppsServiceOp) Delete(ctx context.Context, appID string) (*Response, error) {
	path := fmt.Sprintf("%s/%s", appsBasePath, appID)
	req, err := s.client.NewRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return nil, err
	}
	resp, err := s.client.Do(ctx, req, nil)
	if err != nil {
		return resp, err
	}
	return resp, nil
}

// Propose an app.
func (s *AppsServiceOp) Propose(ctx context.Context, propose *AppProposeRequest) (*AppProposeResponse, *Response, error) {
	path := fmt.Sprintf("%s/propose", appsBasePath)
	req, err := s.client.NewRequest(ctx, http.MethodPost, path, propose)
	if err != nil {
		return nil, nil, err
	}

	res := &AppProposeResponse{}
	resp, err := s.client.Do(ctx, req, res)
	if err != nil {
		return nil, resp, err
	}
	return res, resp, nil
}

// GetDeployment gets an app deployment.
func (s *AppsServiceOp) GetDeployment(ctx context.Context, appID, deploymentID string) (*Deployment, *Response, error) {
	path := fmt.Sprintf("%s/%s/deployments/%s", appsBasePath, appID, deploymentID)
	req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}
	root := new(deploymentRoot)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}
	return root.Deployment, resp, nil
}

// ListDeployments lists an app deployments.
func (s *AppsServiceOp) ListDeployments(ctx context.Context, appID string, opts *ListOptions) ([]*Deployment, *Response, error) {
	path := fmt.Sprintf("%s/%s/deployments", appsBasePath, appID)
	path, err := addOptions(path, opts)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}
	root := new(deploymentsRoot)
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

	return root.Deployments, resp, nil
}

// CreateDeployment creates an app deployment.
func (s *AppsServiceOp) CreateDeployment(ctx context.Context, appID string, create ...*DeploymentCreateRequest) (*Deployment, *Response, error) {
	path := fmt.Sprintf("%s/%s/deployments", appsBasePath, appID)

	var createReq *DeploymentCreateRequest
	for _, c := range create {
		createReq = c
	}

	req, err := s.client.NewRequest(ctx, http.MethodPost, path, createReq)
	if err != nil {
		return nil, nil, err
	}
	root := new(deploymentRoot)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}
	return root.Deployment, resp, nil
}

// GetLogs retrieves app logs.
func (s *AppsServiceOp) GetLogs(ctx context.Context, appID, deploymentID, component string, logType AppLogType, follow bool, tailLines int) (*AppLogs, *Response, error) {
	url := fmt.Sprintf("%s/%s/deployments/%s/logs?type=%s&follow=%t&tail_lines=%d", appsBasePath, appID, deploymentID, logType, follow, tailLines)
	if component != "" {
		url = fmt.Sprintf("%s&component_name=%s", url, component)
	}

	req, err := s.client.NewRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, nil, err
	}
	logs := new(AppLogs)
	resp, err := s.client.Do(ctx, req, logs)
	if err != nil {
		return nil, resp, err
	}
	return logs, resp, nil
}

// ListRegions lists all regions supported by App Platform.
func (s *AppsServiceOp) ListRegions(ctx context.Context) ([]*AppRegion, *Response, error) {
	path := fmt.Sprintf("%s/regions", appsBasePath)
	req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}
	root := new(appRegionsRoot)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}
	return root.Regions, resp, nil
}

// ListTiers lists available app tiers.
func (s *AppsServiceOp) ListTiers(ctx context.Context) ([]*AppTier, *Response, error) {
	path := fmt.Sprintf("%s/tiers", appsBasePath)
	req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}
	root := new(appTiersRoot)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}
	return root.Tiers, resp, nil
}

// GetTier retrieves information about a specific app tier.
func (s *AppsServiceOp) GetTier(ctx context.Context, slug string) (*AppTier, *Response, error) {
	path := fmt.Sprintf("%s/tiers/%s", appsBasePath, slug)
	req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}
	root := new(appTierRoot)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}
	return root.Tier, resp, nil
}

// ListInstanceSizes lists available instance sizes for service, worker, and job components.
func (s *AppsServiceOp) ListInstanceSizes(ctx context.Context) ([]*AppInstanceSize, *Response, error) {
	path := fmt.Sprintf("%s/tiers/instance_sizes", appsBasePath)
	req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}
	root := new(instanceSizesRoot)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}
	return root.InstanceSizes, resp, nil
}

// GetInstanceSize retrieves information about a specific instance size for service, worker, and job components.
func (s *AppsServiceOp) GetInstanceSize(ctx context.Context, slug string) (*AppInstanceSize, *Response, error) {
	path := fmt.Sprintf("%s/tiers/instance_sizes/%s", appsBasePath, slug)
	req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}
	root := new(instanceSizeRoot)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}
	return root.InstanceSize, resp, nil
}

// ListAlerts retrieves a list of alerts on an app
func (s *AppsServiceOp) ListAlerts(ctx context.Context, appID string) ([]*AppAlert, *Response, error) {
	path := fmt.Sprintf("%s/%s/alerts", appsBasePath, appID)
	req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}
	root := new(appAlertsRoot)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}
	return root.Alerts, resp, nil
}

// UpdateAlertDestinations updates the alert destinations of an app's alert
func (s *AppsServiceOp) UpdateAlertDestinations(ctx context.Context, appID, alertID string, update *AlertDestinationUpdateRequest) (*AppAlert, *Response, error) {
	path := fmt.Sprintf("%s/%s/alerts/%s/destinations", appsBasePath, appID, alertID)
	req, err := s.client.NewRequest(ctx, http.MethodPost, path, update)
	if err != nil {
		return nil, nil, err
	}
	root := new(appAlertRoot)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}
	return root.Alert, resp, nil
}

// Detect an app.
func (s *AppsServiceOp) Detect(ctx context.Context, detect *DetectRequest) (*DetectResponse, *Response, error) {
	path := fmt.Sprintf("%s/detect", appsBasePath)
	req, err := s.client.NewRequest(ctx, http.MethodPost, path, detect)
	if err != nil {
		return nil, nil, err
	}

	res := &DetectResponse{}
	resp, err := s.client.Do(ctx, req, res)
	if err != nil {
		return nil, resp, err
	}
	return res, resp, nil
}

// AppComponentType is an app component type.
type AppComponentType string

const (
	// AppComponentTypeService is the type for a service component.
	AppComponentTypeService AppComponentType = "service"
	// AppComponentTypeWorker is the type for a worker component.
	AppComponentTypeWorker AppComponentType = "worker"
	// AppComponentTypeJob is the type for a job component.
	AppComponentTypeJob AppComponentType = "job"
	// AppComponentTypeStaticSite is the type for a static site component.
	AppComponentTypeStaticSite AppComponentType = "static_site"
	// AppComponentTypeDatabase is the type for a database component.
	AppComponentTypeDatabase AppComponentType = "database"
	// AppComponentTypeFunctions is the type for a functions component.
	AppComponentTypeFunctions AppComponentType = "functions"
)

// GetType returns the Service component type.
func (s *AppServiceSpec) GetType() AppComponentType {
	return AppComponentTypeService
}

// GetType returns the Worker component type.
func (s *AppWorkerSpec) GetType() AppComponentType {
	return AppComponentTypeWorker
}

// GetType returns the Job component type.
func (s *AppJobSpec) GetType() AppComponentType {
	return AppComponentTypeJob
}

// GetType returns the StaticSite component type.
func (s *AppStaticSiteSpec) GetType() AppComponentType {
	return AppComponentTypeStaticSite
}

// GetType returns the Database component type.
func (s *AppDatabaseSpec) GetType() AppComponentType {
	return AppComponentTypeDatabase
}

// GetType returns the Functions component type.
func (s *AppFunctionsSpec) GetType() AppComponentType {
	return AppComponentTypeFunctions
}

// AppComponentSpec represents a component's spec.
type AppComponentSpec interface {
	GetName() string
	GetType() AppComponentType
}

// AppBuildableComponentSpec is a component that needs to be built.
type AppBuildableComponentSpec interface {
	AppComponentSpec

	GetGit() *GitSourceSpec
	GetGitHub() *GitHubSourceSpec
	GetGitLab() *GitLabSourceSpec

	GetSourceDir() string
	GetDockerfilePath() string
	GetBuildCommand() string
	GetEnvironmentSlug() string

	GetEnvs() []*AppVariableDefinition
}

// AppContainerComponentSpec is a component that runs in a cluster.
type AppContainerComponentSpec interface {
	AppBuildableComponentSpec

	GetImage() *ImageSourceSpec

	GetRunCommand() string

	GetInstanceSizeSlug() string
	GetInstanceCount() int64
}

// AppRoutableComponentSpec is a component that defines routes.
type AppRoutableComponentSpec interface {
	AppComponentSpec

	GetRoutes() []*AppRouteSpec
	GetCORS() *AppCORSPolicy
}

// AppSourceType is an app source type.
type AppSourceType string

const (
	AppSourceTypeGitHub AppSourceType = "github"
	AppSourceTypeGitLab AppSourceType = "gitlab"
	AppSourceTypeGit    AppSourceType = "git"
	AppSourceTypeImage  AppSourceType = "image"
)

// SourceSpec represents a source.
type SourceSpec interface {
	GetType() AppSourceType
}

// GetType returns the GitHub source type.
func (s *GitHubSourceSpec) GetType() AppSourceType {
	return AppSourceTypeGitHub
}

// GetType returns the GitLab source type.
func (s *GitLabSourceSpec) GetType() AppSourceType {
	return AppSourceTypeGitLab
}

// GetType returns the Git source type.
func (s *GitSourceSpec) GetType() AppSourceType {
	return AppSourceTypeGit
}

// GetType returns the Image source type.
func (s *ImageSourceSpec) GetType() AppSourceType {
	return AppSourceTypeImage
}

// VCSSourceSpec represents a VCS source.
type VCSSourceSpec interface {
	SourceSpec
	GetRepo() string
	GetBranch() string
}

// GetRepo allows GitSourceSpec to implement the SourceSpec interface.
func (s *GitSourceSpec) GetRepo() string {
	return s.RepoCloneURL
}

// ForEachAppComponentSpec iterates over each component spec in an app.
func (s *AppSpec) ForEachAppComponentSpec(fn func(component AppComponentSpec) error) error {
	if s == nil {
		return nil
	}
	for _, c := range s.Services {
		if err := fn(c); err != nil {
			return err
		}
	}
	for _, c := range s.Workers {
		if err := fn(c); err != nil {
			return err
		}
	}
	for _, c := range s.Jobs {
		if err := fn(c); err != nil {
			return err
		}
	}
	for _, c := range s.StaticSites {
		if err := fn(c); err != nil {
			return err
		}
	}
	for _, c := range s.Databases {
		if err := fn(c); err != nil {
			return err
		}
	}
	for _, c := range s.Functions {
		if err := fn(c); err != nil {
			return err
		}
	}
	return nil
}

// ForEachAppSpecComponent loops over each component spec that matches the provided interface type.
func ForEachAppSpecComponent[T any](s *AppSpec, fn func(component T) error) error {
	return s.ForEachAppComponentSpec(func(component AppComponentSpec) error {
		if c, ok := component.(T); ok {
			if err := fn(c); err != nil {
				return err
			}
		}
		return nil
	})
}

// GetAppSpecComponent returns an app spec component by type and name.
// The ComponentSpec type can be used for no restriction on component type.
func GetAppSpecComponent[T interface {
	GetName() string
}](s *AppSpec, name string) (T, error) {
	var c T
	errStop := errors.New("stop")
	err := ForEachAppSpecComponent(s, func(component T) error {
		if component.GetName() == name {
			c = component
			return errStop
		}
		return nil
	})
	if err == errStop {
		return c, nil
	}
	return c, fmt.Errorf("component %s not found", name)
}
