package modal

import (
	"log/slog"
	"sync"

	"github.com/digitalocean/godo/sandbox"
	"github.com/digitalocean/godo/sandbox/compat/internal/shared"
	"github.com/digitalocean/godo/sandbox/compat/unsupported"
)

const platform = "Modal"

func notImplemented(method, needs string) error { return unsupported.New(platform, method, needs) }

func optionNotImplemented(method, option, needs string) error {
	return shared.Option(platform, method, option, needs)
}

// Client mirrors modal.Client. Its services create and find DigitalOcean sandboxes.
type Client struct {
	Apps              AppService
	CloudBucketMounts CloudBucketMountService
	Cls               ClsService
	Environments      *EnvironmentService
	Functions         FunctionService
	FunctionCalls     FunctionCallService
	Images            ImageService
	Proxies           ProxyService
	Queues            QueueService
	Sandboxes         SandboxService
	SandboxSnapshots  SandboxSnapshotService
	Secrets           SecretService
	Volumes           VolumeService

	do *sandbox.Client

	mu        sync.Mutex
	images    map[string]*Image
	templates map[string]string // OCI reference -> ready template name
}

// ClientParams mirrors modal.ClientParams without the gRPC fields
// (Config, ControlPlaneClient, ControlPlaneConn and the interceptors).
// TokenSecret is used as the DigitalOcean API token; TokenID and Environment
// have no effect.
type ClientParams struct {
	TokenID     string
	TokenSecret string
	Environment string
	Logger      *slog.Logger
}

// NewClient mirrors modal.NewClient. It reads DIGITALOCEAN_ACCESS_TOKEN and
// DIGITALOCEAN_API_URL, and never a MODAL_* variable.
func NewClient() (*Client, error) { return newClient("") }

// NewClientWithOptions mirrors modal.NewClientWithOptions.
func NewClientWithOptions(params *ClientParams) (*Client, error) {
	if params == nil {
		return newClient("")
	}
	return newClient(params.TokenSecret)
}

func newClient(token string) (*Client, error) {
	do, err := shared.NewClient(token, "")
	if err != nil {
		return nil, err
	}
	c := &Client{do: do, images: map[string]*Image{}, templates: map[string]string{}}
	c.Apps = appService{notImplementedAppService: notImplementedAppService{}, c: c}
	c.Images = imageService{notImplementedImageService: notImplementedImageService{}, c: c}
	c.Secrets = secretService{notImplementedSecretService: notImplementedSecretService{}}
	c.Sandboxes = sandboxService{notImplementedSandboxService: notImplementedSandboxService{}, c: c}
	c.SandboxSnapshots = snapshotService{c: c}
	c.CloudBucketMounts = notImplementedCloudBucketMountService{}
	c.Cls = notImplementedClsService{}
	c.Functions = notImplementedFunctionService{}
	c.FunctionCalls = notImplementedFunctionCallService{}
	c.Proxies = notImplementedProxyService{}
	c.Queues = notImplementedQueueService{}
	c.Volumes = notImplementedVolumeService{}
	c.Environments = &EnvironmentService{Objects: notImplementedEnvironmentObjectsService{}}
	return c, nil
}

// Close mirrors Client.Close. There is nothing to release.
func (c *Client) Close() {}

// Version mirrors Client.Version.
func (c *Client) Version() string { return "compat/modal-client-go v0.11.2" }

func (c *Client) remember(img *Image) {
	c.mu.Lock()
	c.images[img.ImageID] = img
	c.mu.Unlock()
}
