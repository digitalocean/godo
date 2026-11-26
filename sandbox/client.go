package sandbox

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"net/url"
	"os"
	"strings"

	"github.com/digitalocean/godo"
)

const (
	// DefaultTemplate is the template used when CreateParams.Template is empty.
	DefaultTemplate = "coding-base"
	// DefaultSize is the size slug used when CreateParams.Size is empty. It
	// must be an enabled size.
	DefaultSize = "mars-2vcpu-4gb"

	defaultAPIURL = "https://api.digitalocean.com"
)

// Config configures a Client. The zero value reads everything from the
// environment.
type Config struct {
	// Token is a DigitalOcean API token. Defaults to DIGITALOCEAN_ACCESS_TOKEN.
	Token string
	// APIURL is the API origin. Defaults to DIGITALOCEAN_API_URL, then
	// https://api.digitalocean.com.
	APIURL string
}

// Client creates and finds sandboxes.
type Client struct {
	// API is the underlying godo client, for anything this package does not cover.
	API *godo.Client
}

// NewClient returns a Client. A nil cfg reads the environment.
func NewClient(cfg *Config) (*Client, error) {
	if cfg == nil {
		cfg = &Config{}
	}
	token := strings.TrimSpace(cfg.Token)
	if token == "" {
		token = strings.TrimSpace(os.Getenv("DIGITALOCEAN_ACCESS_TOKEN"))
	}
	if token == "" {
		return nil, ErrNoToken
	}
	apiURL := cfg.APIURL
	if apiURL == "" {
		apiURL = os.Getenv("DIGITALOCEAN_API_URL")
	}
	api := godo.NewFromToken(token)
	if apiURL != "" && apiURL != defaultAPIURL {
		u, err := url.Parse(strings.TrimRight(apiURL, "/") + "/")
		if err != nil {
			return nil, fmt.Errorf("sandbox: invalid API URL: %w", err)
		}
		api.BaseURL = u
	}
	api.UserAgent = "godo-sandbox " + api.UserAgent
	return &Client{API: api}, nil
}

// CreateParams are the options for Create. Every field is optional.
type CreateParams struct {
	// Template to create the sandbox from. Defaults to DefaultTemplate.
	Template string
	// Size slug. Defaults to DefaultSize.
	Size string
	// AllowInternetAccess false limits egress to the platform file-transfer
	// host. Nil leaves it unrestricted.
	AllowInternetAccess *bool
	// EnvVars are environment variables for the sandbox. The values are secrets.
	EnvVars map[string]string
}

// Create creates a sandbox. It returns once the sandbox is running. A nil
// params uses the defaults.
func (c *Client) Create(ctx context.Context, params *CreateParams) (*Sandbox, error) {
	if params == nil {
		params = &CreateParams{}
	}
	template, size := params.Template, params.Size
	if template == "" {
		template = DefaultTemplate
	}
	if strings.TrimSpace(template) == "" {
		return nil, errors.New("sandbox: Template must not be blank")
	}
	if size == "" {
		size = DefaultSize
	}
	info, _, err := c.API.Sandboxes.Create(ctx, &godo.SandboxCreateRequest{
		Template:            template,
		SizeSlug:            size,
		AllowInternetAccess: params.AllowInternetAccess,
		EnvVars:             params.EnvVars,
	})
	if err != nil {
		return nil, err
	}
	return newSandbox(c, *info), nil
}

// Get fetches a sandbox by ID.
func (c *Client) Get(ctx context.Context, id string) (*Sandbox, error) {
	info, _, err := c.API.Sandboxes.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return newSandbox(c, *info), nil
}

// Connect fetches a sandbox by ID and resumes it if it is paused.
func (c *Client) Connect(ctx context.Context, id string) (*Sandbox, error) {
	sb, err := c.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if sb.Info.State == godo.SandboxStatePaused {
		if err := sb.Resume(ctx); err != nil {
			return nil, err
		}
	}
	return sb, nil
}

// Kill deletes a sandbox by ID. It reports false, with no error, when the
// sandbox does not exist (HTTP 404). When the public API returns success for an
// already-deleted id, Kill reports true — callers that need "was it still
// there?" should Get after Kill or rely on a NotFound delete response.
func (c *Client) Kill(ctx context.Context, id string) (bool, error) {
	if _, err := c.API.Sandboxes.Delete(ctx, id); err != nil {
		if IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// Pause pauses a sandbox by ID.
func (c *Client) Pause(ctx context.Context, id string) (*Sandbox, error) {
	info, _, err := c.API.Sandboxes.Pause(ctx, id)
	if err != nil {
		return nil, err
	}
	return newSandbox(c, *info), nil
}

// ListParams are the options for List. Every field is optional.
type ListParams struct {
	// State returns only sandboxes in this state.
	State godo.SandboxState
	// Limit is the page size.
	Limit int
	// NextToken resumes after the NextToken of an earlier Paginator.
	NextToken string
}

// List lists the team's sandboxes. Nothing is fetched until you read a page.
func (c *Client) List(params *ListParams) *Paginator {
	if params == nil {
		params = &ListParams{}
	}
	return &Paginator{client: c, params: *params, token: params.NextToken}
}

// Paginator reads sandboxes one page at a time with NextItems, or all of them
// with All.
type Paginator struct {
	client  *Client
	params  ListParams
	token   string
	started bool
}

// HasNext reports whether another page may exist: true until a page came back
// without a next-page token.
func (p *Paginator) HasNext() bool { return !p.started || p.token != "" }

// NextToken is the token that resumes after the last page read. Empty on the
// last page.
func (p *Paginator) NextToken() string { return p.token }

// NextItems reads the next page.
func (p *Paginator) NextItems(ctx context.Context) ([]*Sandbox, error) {
	if !p.HasNext() {
		return nil, errors.New("sandbox: no more sandboxes to list")
	}
	list, _, err := p.client.API.Sandboxes.List(ctx, &godo.SandboxListOptions{
		State:     p.params.State,
		PageSize:  p.params.Limit,
		PageToken: p.token,
	})
	if err != nil {
		return nil, err
	}
	p.started = true
	p.token = list.NextPageToken
	out := make([]*Sandbox, len(list.Sandboxes))
	for i := range list.Sandboxes {
		out[i] = newSandbox(p.client, list.Sandboxes[i])
	}
	return out, nil
}

// All yields every remaining sandbox, fetching pages as it goes. A failed page
// is yielded as the error, and the sequence ends.
//
//	for sb, err := range c.List(nil).All(ctx) { ... }
func (p *Paginator) All(ctx context.Context) iter.Seq2[*Sandbox, error] {
	return func(yield func(*Sandbox, error) bool) {
		for p.HasNext() {
			page, err := p.NextItems(ctx)
			if err != nil {
				yield(nil, err)
				return
			}
			for _, sb := range page {
				if !yield(sb, nil) {
					return
				}
			}
		}
	}
}
