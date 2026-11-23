package daytona

import (
	"context"
	"time"

	"github.com/digitalocean/godo"
	"github.com/digitalocean/godo/sandbox"
	"github.com/digitalocean/godo/sandbox/compat/daytona/types"
	"github.com/digitalocean/godo/sandbox/compat/internal/shared"
)

// SnapshotService mirrors daytona.SnapshotService. A Daytona snapshot is a
// reusable environment built from an image, which is a DigitalOcean template.
type SnapshotService struct {
	client *Client
}

// pollEvery is how often Create's status channel reads the template. Tests shorten it.
var pollEvery = 2 * time.Second

// toSnapshot maps a template onto the Daytona snapshot shape. State is
// "pending", "building", "active" or "error".
func toSnapshot(ctx context.Context, t *sandbox.Template) *types.Snapshot {
	snap := &types.Snapshot{
		ID:        t.Info.ID,
		Name:      t.Info.Name,
		CreatedAt: t.Info.CreatedAt.Time,
		UpdatedAt: t.Info.UpdatedAt.Time,
	}
	if t.Info.Spec != nil {
		snap.ImageName = shared.ImageRef(t.Info.Spec.Image)
	}
	switch t.Info.Status {
	case godo.SandboxTemplateStatusReady:
		snap.State = "active"
	case godo.SandboxTemplateStatusBuilding:
		snap.State = "building"
	case godo.SandboxTemplateStatusFailed:
		snap.State = "error"
		if reason := failureOf(ctx, t); reason != "" {
			snap.ErrorReason = &reason
		}
	default:
		snap.State = "pending"
	}
	return snap
}

func failureOf(ctx context.Context, t *sandbox.Template) string {
	builds, err := t.Builds(0).NextItems(ctx)
	if err != nil {
		return ""
	}
	for _, b := range builds {
		if b.Status == godo.SandboxTemplateBuildStatusFailed && b.Error != "" {
			return b.Error
		}
	}
	return ""
}

// Create mirrors SnapshotService.Create. It starts a template build from the
// OCI image reference in params.Image and returns at once, as Daytona does; the
// channel reports status changes and closes when the build ends. Resources,
// Entrypoint, SkipValidation and SandboxClass are not supported.
func (s *SnapshotService) Create(ctx context.Context, params *types.CreateSnapshotParams) (*types.Snapshot, <-chan string, error) {
	const m = "SnapshotService.Create"
	if params == nil || params.Name == "" {
		return nil, nil, convertError(errString("snapshot name is required"))
	}
	ref, ok := params.Image.(string)
	if !ok || ref == "" {
		return nil, nil, optionNotImplemented(m, "Image", "an image builder API (only an OCI image reference string is supported)")
	}
	switch {
	case params.Resources != nil:
		return nil, nil, optionNotImplemented(m, "Resources", "a snapshot resource sizing API")
	case len(params.Entrypoint) > 0:
		return nil, nil, optionNotImplemented(m, "Entrypoint", "a snapshot entrypoint API")
	case params.SkipValidation != nil:
		return nil, nil, optionNotImplemented(m, "SkipValidation", "a snapshot validation API")
	case params.SandboxClass != nil:
		return nil, nil, optionNotImplemented(m, "SandboxClass", "a sandbox class API")
	}
	t, err := s.client.do.CreateTemplate(ctx, &sandbox.TemplateParams{Name: params.Name, SourceOCIRef: ref})
	if err != nil {
		return nil, nil, convertError(err)
	}
	logs := make(chan string, 16)
	go s.follow(ctx, t, logs)
	return toSnapshot(ctx, t), logs, nil
}

func (s *SnapshotService) follow(ctx context.Context, t *sandbox.Template, logs chan<- string) {
	defer close(logs)
	last := t.Info.Status
	send := func(msg string) {
		select {
		case logs <- msg:
		default:
		}
	}
	send("snapshot " + t.Info.Name + ": " + string(last))
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(pollEvery):
		}
		if err := t.Refresh(ctx); err != nil {
			send("snapshot " + t.Info.Name + ": " + err.Error())
			return
		}
		if t.Info.Status != last {
			last = t.Info.Status
			send("snapshot " + t.Info.Name + ": " + string(last))
		}
		if last == godo.SandboxTemplateStatusReady || last == godo.SandboxTemplateStatusFailed {
			return
		}
	}
}

// Get mirrors SnapshotService.Get. It matches a template by name or ID.
func (s *SnapshotService) Get(ctx context.Context, nameOrID string) (*types.Snapshot, error) {
	for t, err := range s.client.do.ListTemplates(nil).All(ctx) {
		if err != nil {
			return nil, convertError(err)
		}
		if t.Info.ID == nameOrID || t.Info.Name == nameOrID {
			return toSnapshot(ctx, t), nil
		}
	}
	return nil, notFound("snapshot " + nameOrID)
}

// List mirrors SnapshotService.List. It reads every template and slices the
// requested page (default page 1, limit 100).
func (s *SnapshotService) List(ctx context.Context, page *int, limit *int) (*types.PaginatedSnapshots, error) {
	pg, lim := 1, 100
	if page != nil && *page > 0 {
		pg = *page
	}
	if limit != nil && *limit > 0 {
		lim = *limit
	}
	var all []*types.Snapshot
	for t, err := range s.client.do.ListTemplates(nil).All(ctx) {
		if err != nil {
			return nil, convertError(err)
		}
		all = append(all, toSnapshot(ctx, t))
	}
	total := len(all)
	start := (pg - 1) * lim
	if start > total {
		start = total
	}
	end := start + lim
	if end > total {
		end = total
	}
	return &types.PaginatedSnapshots{Items: all[start:end], Total: total, Page: pg, TotalPages: (total + lim - 1) / lim}, nil
}

// Delete mirrors SnapshotService.Delete.
func (s *SnapshotService) Delete(ctx context.Context, snapshot *types.Snapshot) error {
	if snapshot == nil {
		return convertError(errString("snapshot is required"))
	}
	gone, err := s.client.do.DeleteTemplate(ctx, snapshot.ID)
	if err != nil {
		return convertError(err)
	}
	if !gone {
		return notFound("snapshot " + snapshot.ID)
	}
	return nil
}

type errString string

func (e errString) Error() string { return string(e) }
