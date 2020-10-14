// Code generated by golang.org/x/tools/cmd/bundle. DO NOT EDIT.
//   $ bundle -pkg godo -prefix  ./dev/dist/godo

package godo

import (
	"time"
)

// App An application's configuration and status.
type App struct {
	ID                      string      `json:"id,omitempty"`
	OwnerUUID               string      `json:"owner_uuid,omitempty"`
	Spec                    *AppSpec    `json:"spec"`
	DefaultIngress          string      `json:"default_ingress,omitempty"`
	CreatedAt               time.Time   `json:"created_at,omitempty"`
	UpdatedAt               time.Time   `json:"updated_at,omitempty"`
	ActiveDeployment        *Deployment `json:"active_deployment,omitempty"`
	InProgressDeployment    *Deployment `json:"in_progress_deployment,omitempty"`
	LastDeploymentCreatedAt time.Time   `json:"last_deployment_created_at,omitempty"`
	LiveURL                 string      `json:"live_url,omitempty"`
	Region                  *AppRegion  `json:"region,omitempty"`
	TierSlug                string      `json:"tier_slug,omitempty"`
	LiveURLBase             string      `json:"live_url_base,omitempty"`
	LiveDomain              string      `json:"live_domain,omitempty"`
}

// AppDatabaseSpec struct for AppDatabaseSpec
type AppDatabaseSpec struct {
	// The name. Must be unique across all components within the same app.
	Name    string                `json:"name"`
	Engine  AppDatabaseSpecEngine `json:"engine,omitempty"`
	Version string                `json:"version,omitempty"`
	// Deprecated.
	Size string `json:"size,omitempty"`
	// Deprecated.
	NumNodes int64 `json:"num_nodes,omitempty"`
	// Whether this is a production or dev database.
	Production bool `json:"production,omitempty"`
	// The name of the underlying DigitalOcean DBaaS cluster. This is required for production databases. For dev databases, if cluster_name is not set, a new cluster will be provisioned.
	ClusterName string `json:"cluster_name,omitempty"`
	// The name of the MySQL or PostgreSQL database to configure.
	DBName string `json:"db_name,omitempty"`
	// The name of the MySQL or PostgreSQL user to configure.
	DBUser string `json:"db_user,omitempty"`
}

// AppDatabaseSpecEngine the model 'AppDatabaseSpecEngine'
type AppDatabaseSpecEngine string

// List of AppDatabaseSpecEngine
const (
	AppDatabaseSpecEngine_Unset AppDatabaseSpecEngine = "UNSET"
	AppDatabaseSpecEngine_MySQL AppDatabaseSpecEngine = "MYSQL"
	AppDatabaseSpecEngine_PG    AppDatabaseSpecEngine = "PG"
	AppDatabaseSpecEngine_Redis AppDatabaseSpecEngine = "REDIS"
)

// AppDomainSpec struct for AppDomainSpec
type AppDomainSpec struct {
	// The hostname.
	Domain string            `json:"domain"`
	Type   AppDomainSpecType `json:"type,omitempty"`
	// Whether the domain includes all sub-domains, in addition to the given domain.
	Wildcard bool `json:"wildcard,omitempty"`
}

// AppDomainSpecType  - DEFAULT: The default .ondigitalocean.app domain assigned to this app.  - PRIMARY: The primary domain for this app. This is the domain that is displayed as the default in the control panel, used in bindable environment variables, and any other places that reference an app's live URL. Only one domain may be set as primary.  - ALIAS: A non-primary domain.
type AppDomainSpecType string

// List of AppDomainSpecType
const (
	AppDomainSpecType_Unspecified AppDomainSpecType = "UNSPECIFIED"
	AppDomainSpecType_Default     AppDomainSpecType = "DEFAULT"
	AppDomainSpecType_Primary     AppDomainSpecType = "PRIMARY"
	AppDomainSpecType_Alias       AppDomainSpecType = "ALIAS"
)

// AppJobSpec struct for AppJobSpec
type AppJobSpec struct {
	// The name. Must be unique across all components within the same app.
	Name   string            `json:"name"`
	Git    *GitSourceSpec    `json:"git,omitempty"`
	GitHub *GitHubSourceSpec `json:"github,omitempty"`
	// The path to the Dockerfile relative to the root of the repo. If set, it will be used to build this component. Otherwise, App Platform will attempt to build it using buildpacks.
	DockerfilePath string `json:"dockerfile_path,omitempty"`
	// An optional build command to run while building this component from source.
	BuildCommand string `json:"build_command,omitempty"`
	// An optional run command to override the component's default.
	RunCommand string `json:"run_command,omitempty"`
	// An optional path to the working directory to use for the build. For Dockerfile builds, this will be used as the build context. Must be relative to the root of the repo.
	SourceDir string `json:"source_dir,omitempty"`
	// An environment slug describing the type of this app. For a full list, please refer to [the product documentation](https://www.digitalocean.com/docs/app-platform/).
	EnvironmentSlug string `json:"environment_slug,omitempty"`
	// A list of environment variables made available to the component.
	Envs []*AppVariableDefinition `json:"envs,omitempty"`
	// The instance size to use for this component.
	InstanceSizeSlug string         `json:"instance_size_slug,omitempty"`
	InstanceCount    int64          `json:"instance_count,omitempty"`
	Kind             AppJobSpecKind `json:"kind,omitempty"`
}

// AppJobSpecKind  - UNSPECIFIED: Default job type, will auto-complete to POST_DEPLOY kind.  - PRE_DEPLOY: Indicates a job that runs before an app deployment.  - POST_DEPLOY: Indicates a job that runs after an app deployment.
type AppJobSpecKind string

// List of AppJobSpecKind
const (
	AppJobSpecKind_Unspecified AppJobSpecKind = "UNSPECIFIED"
	AppJobSpecKind_PreDeploy   AppJobSpecKind = "PRE_DEPLOY"
	AppJobSpecKind_PostDeploy  AppJobSpecKind = "POST_DEPLOY"
)

// AppRouteSpec struct for AppRouteSpec
type AppRouteSpec struct {
	// An HTTP path prefix. Paths must start with / and must be unique across all components within an app.
	Path string `json:"path,omitempty"`
}

// AppServiceSpec struct for AppServiceSpec
type AppServiceSpec struct {
	// The name. Must be unique across all components within the same app.
	Name   string            `json:"name"`
	Git    *GitSourceSpec    `json:"git,omitempty"`
	GitHub *GitHubSourceSpec `json:"github,omitempty"`
	// The path to the Dockerfile relative to the root of the repo. If set, it will be used to build this component. Otherwise, App Platform will attempt to build it using buildpacks.
	DockerfilePath string `json:"dockerfile_path,omitempty"`
	// An optional build command to run while building this component from source.
	BuildCommand string `json:"build_command,omitempty"`
	// An optional run command to override the component's default.
	RunCommand string `json:"run_command,omitempty"`
	// An optional path to the working directory to use for the build. For Dockerfile builds, this will be used as the build context. Must be relative to the root of the repo.
	SourceDir string `json:"source_dir,omitempty"`
	// An environment slug describing the type of this app. For a full list, please refer to [the product documentation](https://www.digitalocean.com/docs/app-platform/).
	EnvironmentSlug string `json:"environment_slug,omitempty"`
	// A list of environment variables made available to the component.
	Envs []*AppVariableDefinition `json:"envs,omitempty"`
	// The instance size to use for this component.
	InstanceSizeSlug string `json:"instance_size_slug,omitempty"`
	InstanceCount    int64  `json:"instance_count,omitempty"`
	// The internal port on which this service's run command will listen. Default: 8080 If there is not an environment variable with the name `PORT`, one will be automatically added with its value set to the value of this field.
	HTTPPort int64 `json:"http_port,omitempty"`
	// A list of HTTP routes that should be routed to this component.
	Routes      []*AppRouteSpec            `json:"routes,omitempty"`
	HealthCheck *AppServiceSpecHealthCheck `json:"health_check,omitempty"`
}

// AppServiceSpecHealthCheck struct for AppServiceSpecHealthCheck
type AppServiceSpecHealthCheck struct {
	// Deprecated. Use http_path instead.
	Path string `json:"path,omitempty"`
	// The number of seconds to wait before beginning health checks.
	InitialDelaySeconds int32 `json:"initial_delay_seconds,omitempty"`
	// The number of seconds to wait between health checks.
	PeriodSeconds int32 `json:"period_seconds,omitempty"`
	// The number of seconds after which the check times out.
	TimeoutSeconds int32 `json:"timeout_seconds,omitempty"`
	// The number of successful health checks before considered healthy.
	SuccessThreshold int32 `json:"success_threshold,omitempty"`
	// The number of failed health checks before considered unhealthy.
	FailureThreshold int32 `json:"failure_threshold,omitempty"`
	// The route path used for the HTTP health check ping. If not set, the HTTP health check will be disabled and a TCP health check used instead.
	HTTPPath string `json:"http_path,omitempty"`
}

// AppSpec The desired configuration of an application.
type AppSpec struct {
	// The name of the app. Must be unique across all  in the same account.
	Name string `json:"name"`
	// Workloads which expose publicy-accessible HTTP services.
	Services []*AppServiceSpec `json:"services,omitempty"`
	// Content which can be rendered to static web assets.
	StaticSites []*AppStaticSiteSpec `json:"static_sites,omitempty"`
	// Workloads which do not expose publicly-accessible HTTP services.
	Workers []*AppWorkerSpec `json:"workers,omitempty"`
	// Pre and post deployment workloads which do not expose publicly-accessible HTTP routes.
	Jobs []*AppJobSpec `json:"jobs,omitempty"`
	// Database instances which can provide persistence to workloads within the application.
	Databases []*AppDatabaseSpec `json:"databases,omitempty"`
	// A set of hostnames where the application will be available.
	Domains []*AppDomainSpec `json:"domains,omitempty"`
	// The slug form of the geographical origin of the app.
	Region string `json:"region,omitempty"`
}

// AppStaticSiteSpec struct for AppStaticSiteSpec
type AppStaticSiteSpec struct {
	// The name. Must be unique across all components within the same app.
	Name   string            `json:"name"`
	Git    *GitSourceSpec    `json:"git,omitempty"`
	GitHub *GitHubSourceSpec `json:"github,omitempty"`
	// The path to the Dockerfile relative to the root of the repo. If set, it will be used to build this component. Otherwise, App Platform will attempt to build it using buildpacks.
	DockerfilePath string `json:"dockerfile_path,omitempty"`
	// An optional build command to run while building this component from source.
	BuildCommand string `json:"build_command,omitempty"`
	// An optional path to the working directory to use for the build. For Dockerfile builds, this will be used as the build context. Must be relative to the root of the repo.
	SourceDir string `json:"source_dir,omitempty"`
	// An environment slug describing the type of this app. For a full list, please refer to [the product documentation](https://www.digitalocean.com/docs/app-platform/).
	EnvironmentSlug string `json:"environment_slug,omitempty"`
	// An optional path to where the built assets will be located, relative to the build context. If not set, App Platform will automatically scan for these directory names: `_static`, `dist`, `public`.
	OutputDir     string `json:"output_dir,omitempty"`
	IndexDocument string `json:"index_document,omitempty"`
	// The name of the error document to use when serving this static site. Default: 404.html. If no such file exists within the built assets, App Platform will supply one.
	ErrorDocument string `json:"error_document,omitempty"`
	// A list of environment variables made available to the component.
	Envs []*AppVariableDefinition `json:"envs,omitempty"`
	// A list of HTTP routes that should be routed to this component.
	Routes []*AppRouteSpec `json:"routes,omitempty"`
}

// AppVariableDefinition struct for AppVariableDefinition
type AppVariableDefinition struct {
	// The name
	Key string `json:"key"`
	// The value. If the type is `SECRET`, the value will be encrypted on first submission. On following submissions, the encrypted value should be used.
	Value string           `json:"value,omitempty"`
	Scope AppVariableScope `json:"scope,omitempty"`
	Type  AppVariableType  `json:"type,omitempty"`
}

// AppWorkerSpec struct for AppWorkerSpec
type AppWorkerSpec struct {
	// The name. Must be unique across all components within the same app.
	Name   string            `json:"name"`
	Git    *GitSourceSpec    `json:"git,omitempty"`
	GitHub *GitHubSourceSpec `json:"github,omitempty"`
	// The path to the Dockerfile relative to the root of the repo. If set, it will be used to build this component. Otherwise, App Platform will attempt to build it using buildpacks.
	DockerfilePath string `json:"dockerfile_path,omitempty"`
	// An optional build command to run while building this component from source.
	BuildCommand string `json:"build_command,omitempty"`
	// An optional run command to override the component's default.
	RunCommand string `json:"run_command,omitempty"`
	// An optional path to the working directory to use for the build. For Dockerfile builds, this will be used as the build context. Must be relative to the root of the repo.
	SourceDir string `json:"source_dir,omitempty"`
	// An environment slug describing the type of this app. For a full list, please refer to [the product documentation](https://www.digitalocean.com/docs/app-platform/).
	EnvironmentSlug string `json:"environment_slug,omitempty"`
	// A list of environment variables made available to the component.
	Envs []*AppVariableDefinition `json:"envs,omitempty"`
	// The instance size to use for this component.
	InstanceSizeSlug string `json:"instance_size_slug,omitempty"`
	InstanceCount    int64  `json:"instance_count,omitempty"`
}

// Deployment struct for Deployment
type Deployment struct {
	ID                 string                  `json:"id,omitempty"`
	Spec               *AppSpec                `json:"spec,omitempty"`
	Services           []*DeploymentService    `json:"services,omitempty"`
	StaticSites        []*DeploymentStaticSite `json:"static_sites,omitempty"`
	Workers            []*DeploymentWorker     `json:"workers,omitempty"`
	Jobs               []*DeploymentJob        `json:"jobs,omitempty"`
	PhaseLastUpdatedAt time.Time               `json:"phase_last_updated_at,omitempty"`
	CreatedAt          time.Time               `json:"created_at,omitempty"`
	UpdatedAt          time.Time               `json:"updated_at,omitempty"`
	Cause              string                  `json:"cause,omitempty"`
	ClonedFrom         string                  `json:"cloned_from,omitempty"`
	Progress           *DeploymentProgress     `json:"progress,omitempty"`
	Phase              DeploymentPhase         `json:"phase,omitempty"`
	TierSlug           string                  `json:"tier_slug,omitempty"`
}

// DeploymentJob struct for DeploymentJob
type DeploymentJob struct {
	Name             string `json:"name,omitempty"`
	SourceCommitHash string `json:"source_commit_hash,omitempty"`
}

// DeploymentPhase the model 'DeploymentPhase'
type DeploymentPhase string

// List of DeploymentPhase
const (
	DeploymentPhase_Unknown       DeploymentPhase = "UNKNOWN"
	DeploymentPhase_PendingBuild  DeploymentPhase = "PENDING_BUILD"
	DeploymentPhase_Building      DeploymentPhase = "BUILDING"
	DeploymentPhase_PendingDeploy DeploymentPhase = "PENDING_DEPLOY"
	DeploymentPhase_Deploying     DeploymentPhase = "DEPLOYING"
	DeploymentPhase_Active        DeploymentPhase = "ACTIVE"
	DeploymentPhase_Superseded    DeploymentPhase = "SUPERSEDED"
	DeploymentPhase_Error         DeploymentPhase = "ERROR"
	DeploymentPhase_Canceled      DeploymentPhase = "CANCELED"
)

// DeploymentProgress struct for DeploymentProgress
type DeploymentProgress struct {
	PendingSteps int32                     `json:"pending_steps,omitempty"`
	RunningSteps int32                     `json:"running_steps,omitempty"`
	SuccessSteps int32                     `json:"success_steps,omitempty"`
	ErrorSteps   int32                     `json:"error_steps,omitempty"`
	TotalSteps   int32                     `json:"total_steps,omitempty"`
	Steps        []*DeploymentProgressStep `json:"steps,omitempty"`
	SummarySteps []*DeploymentProgressStep `json:"summary_steps,omitempty"`
}

// DeploymentService struct for DeploymentService
type DeploymentService struct {
	Name             string `json:"name,omitempty"`
	SourceCommitHash string `json:"source_commit_hash,omitempty"`
}

// DeploymentStaticSite struct for DeploymentStaticSite
type DeploymentStaticSite struct {
	Name             string `json:"name,omitempty"`
	SourceCommitHash string `json:"source_commit_hash,omitempty"`
}

// DeploymentWorker struct for DeploymentWorker
type DeploymentWorker struct {
	Name             string `json:"name,omitempty"`
	SourceCommitHash string `json:"source_commit_hash,omitempty"`
}

// GitHubSourceSpec struct for GitHubSourceSpec
type GitHubSourceSpec struct {
	Repo         string `json:"repo,omitempty"`
	Branch       string `json:"branch,omitempty"`
	DeployOnPush bool   `json:"deploy_on_push,omitempty"`
}

// GitSourceSpec struct for GitSourceSpec
type GitSourceSpec struct {
	RepoCloneURL string `json:"repo_clone_url,omitempty"`
	Branch       string `json:"branch,omitempty"`
}

// InstanceSize struct for InstanceSize
type InstanceSize struct {
	Name            string              `json:"name,omitempty"`
	Slug            string              `json:"slug,omitempty"`
	CPUType         InstanceSizeCPUType `json:"cpu_type,omitempty"`
	CPUs            string              `json:"cpus,omitempty"`
	MemoryBytes     string              `json:"memory_bytes,omitempty"`
	USDPerMonth     string              `json:"usd_per_month,omitempty"`
	USDPerSecond    string              `json:"usd_per_second,omitempty"`
	TierSlug        string              `json:"tier_slug,omitempty"`
	TierUpgradeTo   string              `json:"tier_upgrade_to,omitempty"`
	TierDowngradeTo string              `json:"tier_downgrade_to,omitempty"`
}

// InstanceSizeCPUType the model 'InstanceSizeCPUType'
type InstanceSizeCPUType string

// List of InstanceSizeCPUType
const (
	InstanceSizeCPUType_Unspecified InstanceSizeCPUType = "UNSPECIFIED"
	InstanceSizeCPUType_Shared      InstanceSizeCPUType = "SHARED"
	InstanceSizeCPUType_Dedicated   InstanceSizeCPUType = "DEDICATED"
)

// DeploymentProgressStep struct for DeploymentProgressStep
type DeploymentProgressStep struct {
	Name          string                        `json:"name,omitempty"`
	Status        DeploymentProgressStepStatus  `json:"status,omitempty"`
	Steps         []*DeploymentProgressStep     `json:"steps,omitempty"`
	StartedAt     time.Time                     `json:"started_at,omitempty"`
	EndedAt       time.Time                     `json:"ended_at,omitempty"`
	Reason        *DeploymentProgressStepReason `json:"reason,omitempty"`
	ComponentName string                        `json:"component_name,omitempty"`
	// The base of a human-readable description of the step intended to be combined with the component name for presentation. For example:  `message_base` = \"Building service\" `component_name` = \"api\"
	MessageBase string `json:"message_base,omitempty"`
}

// DeploymentProgressStepStatus the model 'DeploymentProgressStepStatus'
type DeploymentProgressStepStatus string

// List of DeploymentProgressStepStatus
const (
	DeploymentProgressStepStatus_Unknown DeploymentProgressStepStatus = "UNKNOWN"
	DeploymentProgressStepStatus_Pending DeploymentProgressStepStatus = "PENDING"
	DeploymentProgressStepStatus_Running DeploymentProgressStepStatus = "RUNNING"
	DeploymentProgressStepStatus_Error   DeploymentProgressStepStatus = "ERROR"
	DeploymentProgressStepStatus_Success DeploymentProgressStepStatus = "SUCCESS"
)

// AppRegion struct for AppRegion
type AppRegion struct {
	Slug        string   `json:"slug,omitempty"`
	Label       string   `json:"label,omitempty"`
	Flag        string   `json:"flag,omitempty"`
	Continent   string   `json:"continent,omitempty"`
	Disabled    bool     `json:"disabled,omitempty"`
	DataCenters []string `json:"data_centers,omitempty"`
	Reason      string   `json:"reason,omitempty"`
	// Whether or not the region is presented as the default.
	Default bool `json:"default,omitempty"`
}

// DeploymentProgressStepReason struct for DeploymentProgressStepReason
type DeploymentProgressStepReason struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

// AppTier struct for AppTier
type AppTier struct {
	Name                 string `json:"name,omitempty"`
	Slug                 string `json:"slug,omitempty"`
	StorageBytes         string `json:"storage_bytes,omitempty"`
	EgressBandwidthBytes string `json:"egress_bandwidth_bytes,omitempty"`
	BuildSeconds         string `json:"build_seconds,omitempty"`
}

// AppVariableScope the model 'AppVariableScope'
type AppVariableScope string

// List of AppVariableScope
const (
	AppVariableScope_Unset           AppVariableScope = "UNSET"
	AppVariableScope_RunTime         AppVariableScope = "RUN_TIME"
	AppVariableScope_BuildTime       AppVariableScope = "BUILD_TIME"
	AppVariableScope_RunAndBuildTime AppVariableScope = "RUN_AND_BUILD_TIME"
)

// AppVariableType the model 'AppVariableType'
type AppVariableType string

// List of AppVariableType
const (
	AppVariableType_General AppVariableType = "GENERAL"
	AppVariableType_Secret  AppVariableType = "SECRET"
)
