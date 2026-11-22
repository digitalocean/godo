package sandbox

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/digitalocean/godo"
)

// TemplateParams are the options for CreateTemplate.
type TemplateParams struct {
	// Name of the template. Required. It must not collide with a global base
	// template name.
	Name string
	// BaseTemplate is the accepted base template to rebase the image onto.
	// Defaults to DefaultTemplate.
	BaseTemplate string
	// SourceOCIRef is the customer OCI image: registry/repository:tag or
	// registry/repository@digest. Required.
	SourceOCIRef string
}

// TemplateUpdate is the change UpdateTemplate makes. An empty field keeps its
// previous value.
type TemplateUpdate struct {
	BaseTemplate string
	SourceOCIRef string
}

// ListTemplatesParams are the options for ListTemplates. Every field is optional.
type ListTemplatesParams struct {
	// Limit is the page size.
	Limit int
	// NextToken resumes after the NextToken of an earlier Pages.
	NextToken string
}

// Template is a team custom template. Create one with Client.CreateTemplate and
// wait for it with WaitUntilReady; then pass its name to Client.Create.
type Template struct {
	client *Client

	// Info is the record as last returned by the API. Refresh updates it.
	Info godo.SandboxTemplate
}

// ID is the template ID.
func (t *Template) ID() string { return t.Info.ID }

// Name is the template name, which is what Client.Create takes.
func (t *Template) Name() string { return t.Info.Name }

// Status is the status as last returned by the API. It can be stale: Refresh
// reads it again.
func (t *Template) Status() godo.SandboxTemplateStatus { return t.Info.Status }

// CreateTemplate builds a custom template from an OCI image. It returns as soon
// as the build starts: call WaitUntilReady before using the template.
func (c *Client) CreateTemplate(ctx context.Context, params *TemplateParams) (*Template, error) {
	if params == nil {
		return nil, errors.New("sandbox: TemplateParams is required")
	}
	if strings.TrimSpace(params.Name) == "" {
		return nil, errors.New("sandbox: template Name is required")
	}
	if strings.TrimSpace(params.SourceOCIRef) == "" {
		return nil, errors.New("sandbox: template SourceOCIRef is required")
	}
	base := params.BaseTemplate
	if base == "" {
		base = DefaultTemplate
	}
	info, _, err := c.API.Sandboxes.CreateTemplate(ctx, &godo.SandboxTemplateCreateRequest{
		Name:         params.Name,
		BaseTemplate: base,
		SourceOCIRef: params.SourceOCIRef,
	})
	if err != nil {
		return nil, err
	}
	return &Template{client: c, Info: *info}, nil
}

// GetTemplate fetches a custom template by ID.
func (c *Client) GetTemplate(ctx context.Context, id string) (*Template, error) {
	info, _, err := c.API.Sandboxes.GetTemplate(ctx, id)
	if err != nil {
		return nil, err
	}
	return &Template{client: c, Info: *info}, nil
}

// DeleteTemplate deletes a custom template by ID. It reports false, with no
// error, when the template does not exist.
func (c *Client) DeleteTemplate(ctx context.Context, id string) (bool, error) {
	if _, err := c.API.Sandboxes.DeleteTemplate(ctx, id); err != nil {
		if IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// ListTemplates lists your team's custom templates. Global base templates are
// not included. Nothing is fetched until you read a page.
func (c *Client) ListTemplates(params *ListTemplatesParams) *Pages[*Template] {
	if params == nil {
		params = &ListTemplatesParams{}
	}
	limit := params.Limit
	return &Pages[*Template]{
		token: params.NextToken,
		fetch: func(ctx context.Context, token string) ([]*Template, string, error) {
			list, _, err := c.API.Sandboxes.ListTemplates(ctx, &godo.SandboxTemplateListOptions{PageSize: limit, PageToken: token})
			if err != nil {
				return nil, "", err
			}
			out := make([]*Template, len(list.Templates))
			for i := range list.Templates {
				out[i] = &Template{client: c, Info: list.Templates[i]}
			}
			return out, list.NextPageToken, nil
		},
	}
}

// Refresh reads the template again.
func (t *Template) Refresh(ctx context.Context) error {
	info, _, err := t.client.API.Sandboxes.GetTemplate(ctx, t.Info.ID)
	if err != nil {
		return err
	}
	t.Info = *info
	return nil
}

// Update changes the base template and/or source image and starts a new build.
// Call WaitUntilReady afterwards.
func (t *Template) Update(ctx context.Context, update *TemplateUpdate) error {
	if update == nil || (update.BaseTemplate == "" && update.SourceOCIRef == "") {
		return errors.New("sandbox: TemplateUpdate needs a BaseTemplate or a SourceOCIRef")
	}
	info, _, err := t.client.API.Sandboxes.UpdateTemplate(ctx, t.Info.ID, &godo.SandboxTemplateUpdateRequest{
		BaseTemplate: update.BaseTemplate,
		SourceOCIRef: update.SourceOCIRef,
	})
	if err != nil {
		return err
	}
	t.Info = *info
	return nil
}

// Delete deletes the template. It reports false, with no error, when it is
// already gone.
func (t *Template) Delete(ctx context.Context) (bool, error) {
	return t.client.DeleteTemplate(ctx, t.Info.ID)
}

// Builds lists the template's builds. limit is the page size; 0 uses the
// API's default.
func (t *Template) Builds(limit int) *Pages[godo.SandboxTemplateBuild] {
	return &Pages[godo.SandboxTemplateBuild]{
		fetch: func(ctx context.Context, token string) ([]godo.SandboxTemplateBuild, string, error) {
			list, _, err := t.client.API.Sandboxes.ListTemplateBuilds(ctx, t.Info.ID,
				&godo.SandboxTemplateListOptions{PageSize: limit, PageToken: token})
			if err != nil {
				return nil, "", err
			}
			return list.Builds, list.NextPageToken, nil
		},
	}
}

// Build fetches one build of the template.
func (t *Template) Build(ctx context.Context, buildID string) (*godo.SandboxTemplateBuild, error) {
	build, _, err := t.client.API.Sandboxes.GetTemplateBuild(ctx, t.Info.ID, buildID)
	return build, err
}

// WaitUntilReady polls until the template is ready, so it can create sandboxes.
// It fails, with the build's error when the API gives one, if the build failed,
// and when the wait runs out. Timeout defaults to 10 minutes and PollInterval to
// 2 seconds; a shorter ctx deadline wins.
func (t *Template) WaitUntilReady(ctx context.Context, opts *WaitOptions) error {
	timeout, interval := 10*time.Minute, 2*time.Second
	if opts != nil {
		if opts.Timeout > 0 {
			timeout = opts.Timeout
		}
		if opts.PollInterval > 0 {
			interval = opts.PollInterval
		}
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	for {
		switch t.Info.Status {
		case godo.SandboxTemplateStatusReady:
			return nil
		case godo.SandboxTemplateStatusFailed:
			return t.buildFailure(ctx)
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("sandbox: template %s was not ready in time (last status %s): %w", t.Info.ID, t.Info.Status, ctx.Err())
		case <-time.After(interval):
		}
		if err := t.Refresh(ctx); err != nil {
			if ctx.Err() != nil {
				return fmt.Errorf("sandbox: template %s was not ready in time (last status %s): %w", t.Info.ID, t.Info.Status, ctx.Err())
			}
			return err
		}
	}
}

// buildFailure explains a failed template with the newest failed build's error.
func (t *Template) buildFailure(ctx context.Context) error {
	msg := fmt.Sprintf("sandbox: template %s failed to build", t.Info.ID)
	page, err := t.Builds(0).NextItems(ctx)
	if err != nil {
		return errors.New(msg)
	}
	var newest *godo.SandboxTemplateBuild
	for i := range page {
		b := &page[i]
		if b.Status == godo.SandboxTemplateBuildStatusFailed && b.Error != "" &&
			(newest == nil || b.UpdatedAt.Time.After(newest.UpdatedAt.Time)) {
			newest = b
		}
	}
	if newest != nil {
		return fmt.Errorf("%s: %s", msg, newest.Error)
	}
	return errors.New(msg)
}
