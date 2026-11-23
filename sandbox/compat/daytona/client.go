package daytona

import (
	"context"
	"fmt"
	"iter"
	"time"

	"github.com/digitalocean/godo"
	"github.com/digitalocean/godo/sandbox"
	"github.com/digitalocean/godo/sandbox/compat/daytona/apiclient"
	dterrors "github.com/digitalocean/godo/sandbox/compat/daytona/errors"
	"github.com/digitalocean/godo/sandbox/compat/daytona/options"
	"github.com/digitalocean/godo/sandbox/compat/daytona/types"
	"github.com/digitalocean/godo/sandbox/compat/internal/shared"
)

// Client mirrors daytona.Client. It creates and finds DigitalOcean sandboxes.
type Client struct {
	// Volume mirrors Client.Volume. Every method returns a not-implemented error.
	Volume *VolumeService
	// Snapshot mirrors Client.Snapshot. A Daytona snapshot is a DigitalOcean template.
	Snapshot *SnapshotService

	do *sandbox.Client
}

// NewClient mirrors daytona.NewClient. It reads DIGITALOCEAN_ACCESS_TOKEN and
// DIGITALOCEAN_API_URL, and never a DAYTONA_* variable.
func NewClient() (*Client, error) {
	return newClient("", "")
}

// NewClientWithConfig mirrors daytona.NewClientWithConfig. APIKey is the
// DigitalOcean API token and APIUrl is the API origin; an empty field falls
// back to the DigitalOcean environment variables. OrganizationID, OtelEnabled
// and Experimental have no effect. JWTToken and Target are not supported.
func NewClientWithConfig(config *types.DaytonaConfig) (*Client, error) {
	if config == nil {
		return NewClient()
	}
	if config.JWTToken != "" && config.APIKey == "" {
		return nil, optionNotImplemented("NewClientWithConfig", "JWTToken", "JWT authentication")
	}
	if config.Target != "" {
		return nil, optionNotImplemented("NewClientWithConfig", "Target", "a region selection API")
	}
	return newClient(config.APIKey, config.APIUrl)
}

func newClient(token, apiURL string) (*Client, error) {
	do, err := shared.NewClient(token, apiURL)
	if err != nil {
		return nil, err
	}
	c := &Client{do: do, Volume: &VolumeService{}}
	c.Snapshot = &SnapshotService{client: c}
	return c, nil
}

// Close mirrors Client.Close. There is nothing to release.
func (c *Client) Close(ctx context.Context) error { return nil }

// baseCreate is what Create needs from its params argument.
type baseCreate struct {
	base     types.SandboxBaseParams
	snapshot string
	image    any
	res      *types.Resources
}

func parseCreateParams(params any) (baseCreate, error) {
	switch p := params.(type) {
	case nil:
		return baseCreate{}, nil
	case types.SnapshotParams:
		return baseCreate{base: p.SandboxBaseParams, snapshot: p.Snapshot}, nil
	case *types.SnapshotParams:
		if p == nil {
			return baseCreate{}, nil
		}
		return baseCreate{base: p.SandboxBaseParams, snapshot: p.Snapshot}, nil
	case types.ImageParams:
		return baseCreate{base: p.SandboxBaseParams, image: p.Image, res: p.Resources}, nil
	case *types.ImageParams:
		if p == nil {
			return baseCreate{}, nil
		}
		return baseCreate{base: p.SandboxBaseParams, image: p.Image, res: p.Resources}, nil
	}
	return baseCreate{}, fmt.Errorf("daytona: Create params must be types.SnapshotParams or types.ImageParams, not %T", params)
}

// checkBase fails for every SandboxBaseParams field that has no DigitalOcean
// equivalent. Language is a client-side hint for CodeRun and is ignored.
func checkBase(b types.SandboxBaseParams) error {
	const m = "Client.Create"
	switch {
	case b.Name != "":
		return optionNotImplemented(m, "Name", "a sandbox names API")
	case b.User != "":
		return optionNotImplemented(m, "User", "a sandbox user API")
	case len(b.Labels) > 0:
		return optionNotImplemented(m, "Labels", "a sandbox labels API")
	case b.Public:
		return optionNotImplemented(m, "Public", "a public sandbox API")
	case b.AutoStopInterval != nil:
		return optionNotImplemented(m, "AutoStopInterval", "a sandbox auto-stop lifetime API")
	case b.AutoArchiveInterval != nil:
		return optionNotImplemented(m, "AutoArchiveInterval", "a sandbox auto-archive API")
	case b.AutoDeleteInterval != nil:
		return optionNotImplemented(m, "AutoDeleteInterval", "a sandbox auto-delete lifetime API")
	case len(b.Volumes) > 0:
		return optionNotImplemented(m, "Volumes", "a volumes API")
	case b.NetworkAllowList != nil:
		return optionNotImplemented(m, "NetworkAllowList", "an outbound allow-list API")
	case b.DomainAllowList != nil:
		return optionNotImplemented(m, "DomainAllowList", "an outbound allow-list API")
	case b.Ephemeral:
		return optionNotImplemented(m, "Ephemeral", "an ephemeral sandbox API")
	case b.LinkedSandbox != "":
		return optionNotImplemented(m, "LinkedSandbox", "a linked sandboxes API")
	}
	return nil
}

func sizeFor(res *types.Resources) (string, error) {
	if res == nil {
		return "", nil
	}
	const m = "Client.Create"
	switch {
	case res.GPU != 0 || len(res.GpuType) > 0:
		return "", optionNotImplemented(m, "Resources.GPU", "a GPU sandbox API")
	case res.Disk != 0:
		return "", optionNotImplemented(m, "Resources.Disk", "a sandbox disk sizing API")
	}
	slug, err := shared.SizeFor(float64(res.CPU), res.Memory*1024)
	if err != nil {
		return "", dterrors.NewDaytonaValidationError(err.Error(), nil)
	}
	return slug, nil
}

// Create mirrors Client.Create. SnapshotParams starts a sandbox from the
// template named by Snapshot. ImageParams with an image reference builds (or
// reuses) a template from that OCI image first. A *DockerImage is not
// supported. Create always waits for the sandbox to run; WithTimeout bounds
// that wait, and WithWaitForStart(false) does not return earlier.
func (c *Client) Create(ctx context.Context, params any, opts ...func(*options.CreateSandbox)) (*Sandbox, error) {
	co := &options.CreateSandbox{WaitForStart: true}
	for _, o := range opts {
		o(co)
	}
	bc, err := parseCreateParams(params)
	if err != nil {
		return nil, err
	}
	if err := checkBase(bc.base); err != nil {
		return nil, err
	}
	size, err := sizeFor(bc.res)
	if err != nil {
		return nil, err
	}
	timeout := defaultTimeout
	if bc.image != nil {
		timeout = 10 * time.Minute // a template build takes longer than a start
	}
	if co.Timeout != nil {
		timeout = *co.Timeout
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	template := bc.snapshot
	if bc.image != nil {
		ref, ok := bc.image.(string)
		if !ok {
			return nil, optionNotImplemented("Client.Create", "Image", "an image builder API (only an OCI image reference string is supported)")
		}
		if template, err = c.templateForImage(ctx, ref); err != nil {
			return nil, convertError(err)
		}
		if co.LogChannel != nil {
			close(co.LogChannel)
		}
	}
	cp := &sandbox.CreateParams{Template: template, Size: size, EnvVars: bc.base.EnvVars}
	if bc.base.NetworkBlockAll {
		off := false
		cp.AllowInternetAccess = &off
	}
	sb, err := c.do.Create(ctx, cp)
	if err != nil {
		return nil, convertError(err)
	}
	return newSandbox(c, sb), nil
}

// templateForImage finds or builds the template for an OCI reference.
func (c *Client) templateForImage(ctx context.Context, ref string) (string, error) {
	name := shared.TemplateName("daytona", ref)
	for t, err := range c.do.ListTemplates(nil).All(ctx) {
		if err != nil {
			return "", err
		}
		if t.Name() == name {
			return name, t.WaitUntilReady(ctx, nil)
		}
	}
	t, err := c.do.CreateTemplate(ctx, &sandbox.TemplateParams{Name: name, SourceOCIRef: ref})
	if err != nil {
		return "", err
	}
	return name, t.WaitUntilReady(ctx, nil)
}

// Get mirrors Client.Get. The argument must be a sandbox ID: DigitalOcean
// sandboxes have no names.
func (c *Client) Get(ctx context.Context, sandboxIDOrName string) (*Sandbox, error) {
	sb, err := c.do.Get(ctx, sandboxIDOrName)
	if err != nil {
		return nil, convertError(err)
	}
	return newSandbox(c, sb), nil
}

// List mirrors Client.List.
func (c *Client) List(ctx context.Context, query *ListSandboxesQuery) *SandboxIterator {
	it := &SandboxIterator{}
	keep, params, err := queryFilter(query)
	if err != nil {
		it.err = err
		return it
	}
	pull, stop := iter.Pull2(c.do.List(params).All(ctx))
	it.pull, it.stop, it.keep, it.client = pull, stop, keep, c
	return it
}

// ListSeq mirrors Client.ListSeq.
func (c *Client) ListSeq(ctx context.Context, query *ListSandboxesQuery) iter.Seq2[*Sandbox, error] {
	return func(yield func(*Sandbox, error) bool) {
		it := c.List(ctx, query)
		for it.Next() {
			if !yield(it.Value(), nil) {
				it.close()
				return
			}
		}
		if err := it.Err(); err != nil {
			yield(nil, err)
		}
	}
}

// queryFilter turns a ListSandboxesQuery into a server-side filter, where the
// API has one, and a client-side filter for the rest. A field with no
// DigitalOcean equivalent fails the list.
func queryFilter(q *ListSandboxesQuery) (func(*sandbox.Sandbox) bool, *sandbox.ListParams, error) {
	params := &sandbox.ListParams{}
	if q == nil {
		return func(*sandbox.Sandbox) bool { return true }, params, nil
	}
	const m = "Client.List"
	switch {
	case q.Name != nil:
		return nil, nil, optionNotImplemented(m, "Name", "a sandbox names API")
	case len(q.Labels) > 0:
		return nil, nil, optionNotImplemented(m, "Labels", "a sandbox labels API")
	case len(q.Targets) > 0:
		return nil, nil, optionNotImplemented(m, "Targets", "a region selection API")
	case q.MinCpu != nil, q.MaxCpu != nil, q.MinMemoryGib != nil, q.MaxMemoryGib != nil, q.MinDiskGib != nil, q.MaxDiskGib != nil:
		return nil, nil, optionNotImplemented(m, "resource filters", "a sandbox resource filter API")
	case q.IsPublic != nil, q.IsRecoverable != nil:
		return nil, nil, optionNotImplemented(m, "IsPublic and IsRecoverable", "a sandbox visibility and recovery API")
	case q.LastActivityAfter != nil, q.LastActivityBefore != nil:
		return nil, nil, optionNotImplemented(m, "LastActivity filters", "a sandbox activity API")
	case q.Sort != nil, q.Order != nil:
		return nil, nil, optionNotImplemented(m, "Sort and Order", "a sandbox sort API")
	}
	if q.Limit != nil {
		params.Limit = *q.Limit
	}
	states := map[godo.SandboxState]bool{}
	for _, s := range q.States {
		switch s {
		case apiclient.SANDBOXSTATE_STARTED:
			states[godo.SandboxStateRunning] = true
		case apiclient.SANDBOXSTATE_STOPPED:
			states[godo.SandboxStatePaused] = true
		case apiclient.SANDBOXSTATE_CREATING:
			states[godo.SandboxStateCreating] = true
		}
	}
	if len(q.States) > 0 && len(states) == 0 {
		// No DigitalOcean sandbox is ever in the requested states.
		return func(*sandbox.Sandbox) bool { return false }, params, nil
	}
	if len(states) == 1 {
		for s := range states {
			params.State = s
		}
	}
	snaps := map[string]bool{}
	for _, s := range q.Snapshots {
		snaps[s] = true
	}
	return func(sb *sandbox.Sandbox) bool {
		if q.ID != nil && sb.ID() != *q.ID {
			return false
		}
		if len(states) > 1 && !states[sb.State()] {
			return false
		}
		if len(snaps) > 0 && !snaps[sb.Template()] {
			return false
		}
		created := sb.Info.CreatedAt.Time
		if q.CreatedAtAfter != nil && !created.After(*q.CreatedAtAfter) {
			return false
		}
		if q.CreatedAtBefore != nil && !created.Before(*q.CreatedAtBefore) {
			return false
		}
		return true
	}, params, nil
}
