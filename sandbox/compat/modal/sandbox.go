package modal

import (
	"context"
	"fmt"
	"io"
	"iter"
	"sync"
	"time"

	"github.com/digitalocean/godo/sandbox"
	"github.com/digitalocean/godo/sandbox/compat/internal/shared"
)

// Sandbox mirrors modal.Sandbox. It wraps a DigitalOcean sandbox.
type Sandbox struct {
	SandboxID string
	Stdin     io.WriteCloser
	Stdout    io.ReadCloser
	Stderr    io.ReadCloser

	// Filesystem mirrors Sandbox.Filesystem. Its methods are not implemented.
	Filesystem *SandboxFilesystem
	// ExperimentalSidecars mirrors Sandbox.ExperimentalSidecars. Its methods are not implemented.
	ExperimentalSidecars SidecarService
	// Logs mirrors Sandbox.Logs. Its methods are not implemented.
	Logs *SandboxLogsManager

	client  *Client
	sb      *sandbox.Sandbox
	workdir string

	mu     sync.Mutex
	entry  *ContainerProcess
	procs  []*ContainerProcess
	killed bool
}

// sandboxService implements SandboxService. FromName, ExperimentalFromName and
// ExperimentalFromSnapshot are not implemented.
type sandboxService struct {
	notImplementedSandboxService
	c *Client
}

func newSandbox(c *Client, sb *sandbox.Sandbox, workdir string) *Sandbox {
	return &Sandbox{
		SandboxID:            sb.ID(),
		client:               c,
		sb:                   sb,
		workdir:              workdir,
		Filesystem:           &SandboxFilesystem{},
		ExperimentalSidecars: notImplementedSidecarService{},
		Logs:                 &SandboxLogsManager{},
	}
}

// checkCreate fails for every SandboxCreateParams field with no DigitalOcean
// equivalent. Verbose has no effect, and ExperimentalEnableSnapshot and
// ExperimentalPersistent are always on.
func checkCreate(m string, p *SandboxCreateParams) error {
	switch {
	case p.CPULimit != 0 || p.MemoryLimitMiB != 0:
		return optionNotImplemented(m, "CPULimit and MemoryLimitMiB", "a hard resource limit API")
	case p.Runtime != "":
		return optionNotImplemented(m, "Runtime", "a sandbox runtime selection API")
	case p.GPU != "":
		return optionNotImplemented(m, "GPU", "a GPU sandbox API")
	case p.Timeout != 0:
		return optionNotImplemented(m, "Timeout", "a sandbox lifetime API")
	case p.IdleTimeout != 0:
		return optionNotImplemented(m, "IdleTimeout", "a sandbox idle timeout API")
	case len(p.Volumes) > 0:
		return optionNotImplemented(m, "Volumes", "a volumes API")
	case len(p.CloudBucketMounts) > 0:
		return optionNotImplemented(m, "CloudBucketMounts", "a cloud bucket mounts API")
	case p.PTY:
		return optionNotImplemented(m, "PTY", "a PTY API")
	case len(p.EncryptedPorts) > 0 || len(p.H2Ports) > 0 || len(p.UnencryptedPorts) > 0:
		return optionNotImplemented(m, "EncryptedPorts, H2Ports and UnencryptedPorts", "a ports and tunnels API")
	case p.OutboundCIDRAllowlist != nil || p.OutboundDomainAllowlist != nil:
		return optionNotImplemented(m, "OutboundCIDRAllowlist and OutboundDomainAllowlist", "an outbound allow-list API")
	case len(p.InboundCIDRAllowlist) > 0:
		return optionNotImplemented(m, "InboundCIDRAllowlist", "an inbound allow-list API")
	case p.I6PN:
		return optionNotImplemented(m, "I6PN", "a private IPv6 networking API")
	case p.Cloud != "" || len(p.Regions) > 0:
		return optionNotImplemented(m, "Cloud and Regions", "a region selection API")
	case p.Proxy != nil:
		return optionNotImplemented(m, "Proxy", "a sandbox proxy API")
	case p.ReadinessProbe != nil:
		return optionNotImplemented(m, "ReadinessProbe", "a readiness probe API")
	case p.Name != "":
		return optionNotImplemented(m, "Name", "a sandbox names API")
	case len(p.Tags) > 0:
		return optionNotImplemented(m, "Tags", "a sandbox tags API")
	case len(p.ExperimentalOptions) > 0:
		return optionNotImplemented(m, "ExperimentalOptions", "experimental sandbox options")
	case p.CustomDomain != "":
		return optionNotImplemented(m, "CustomDomain", "a custom domain API")
	case p.IncludeOidcIdentityToken:
		return optionNotImplemented(m, "IncludeOidcIdentityToken", "an OIDC identity token API")
	case p.ExperimentalOutboundPolicy != nil:
		return optionNotImplemented(m, "ExperimentalOutboundPolicy", "an outbound header policy API")
	}
	return nil
}

// Create mirrors SandboxService.Create. The app argument is only a label. A nil
// image uses the default template. Command runs as the sandbox's entrypoint,
// with Stdin, Stdout and Stderr connected to it.
func (s sandboxService) Create(ctx context.Context, app *App, image *Image, params *SandboxCreateParams) (*Sandbox, error) {
	return s.create(ctx, image, params, "Sandboxes.Create")
}

// ExperimentalCreate mirrors SandboxService.ExperimentalCreate. It behaves like Create.
func (s sandboxService) ExperimentalCreate(ctx context.Context, app *App, image *Image, params *SandboxCreateParams) (*Sandbox, error) {
	return s.create(ctx, image, params, "Sandboxes.ExperimentalCreate")
}

func (s sandboxService) create(ctx context.Context, image *Image, params *SandboxCreateParams, method string) (*Sandbox, error) {
	if params == nil {
		params = &SandboxCreateParams{}
	}
	if err := checkCreate(method, params); err != nil {
		return nil, err
	}
	size, err := shared.SizeFor(params.CPU, params.MemoryMiB)
	if err != nil {
		return nil, InvalidError{Exception: err.Error()}
	}
	secrets, err := secretEnv(params.Secrets)
	if err != nil {
		return nil, err
	}
	plan := &imagePlan{env: map[string]string{}}
	if image != nil {
		if plan, err = image.plan(ctx); err != nil {
			return nil, convertError(err)
		}
	}
	env := joinEnv(plan.env, secrets, params.Env)
	cp := &sandbox.CreateParams{Template: plan.template, Size: size, EnvVars: env}
	if params.BlockNetwork {
		off := false
		cp.AllowInternetAccess = &off
	}
	sb, err := s.c.do.Create(ctx, cp)
	if err != nil {
		return nil, convertError(err)
	}
	workdir := params.Workdir
	if workdir == "" {
		workdir = plan.workdir
	}
	for _, step := range plan.steps {
		res, err := sb.Commands.Run(ctx, step, &sandbox.RunOptions{Cwd: workdir, EnvVars: env})
		if err == nil && res.ExitCode != 0 {
			err = fmt.Errorf("modal: image step %q exited with code %d: %s", step, res.ExitCode, res.Stderr)
		}
		if err != nil {
			_, _ = sb.Kill(context.WithoutCancel(ctx))
			return nil, convertError(err)
		}
	}
	out := newSandbox(s.c, sb, workdir)
	if len(params.Command) > 0 {
		entry, err := out.start(ctx, params.Command, workdir, nil, 0, true, true)
		if err != nil {
			_, _ = sb.Kill(context.WithoutCancel(ctx))
			return nil, convertError(err)
		}
		out.entry = entry
		out.Stdin, out.Stdout, out.Stderr = entry.Stdin, entry.Stdout, entry.Stderr
	}
	return out, nil
}

// FromID mirrors SandboxService.FromID.
func (s sandboxService) FromID(ctx context.Context, sandboxID string, params *SandboxFromIDParams) (*Sandbox, error) {
	sb, err := s.c.do.Get(ctx, sandboxID)
	if err != nil {
		return nil, convertError(err)
	}
	return newSandbox(s.c, sb, ""), nil
}

// List mirrors SandboxService.List. Filtering by AppID or Tags is not supported.
func (s sandboxService) List(ctx context.Context, params *SandboxListParams) (iter.Seq2[*Sandbox, error], error) {
	if params != nil {
		if err := checkList("Sandboxes.List", params.AppID, params.Tags); err != nil {
			return nil, err
		}
	}
	return s.list(ctx), nil
}

// ExperimentalList mirrors SandboxService.ExperimentalList.
func (s sandboxService) ExperimentalList(ctx context.Context, params *SandboxExperimentalListParams) (iter.Seq2[*Sandbox, error], error) {
	if params != nil {
		if err := checkList("Sandboxes.ExperimentalList", params.AppID, params.Tags); err != nil {
			return nil, err
		}
	}
	return s.list(ctx), nil
}

func checkList(method, appID string, tags map[string]string) error {
	switch {
	case appID != "":
		return optionNotImplemented(method, "AppID", "a sandbox app grouping API")
	case len(tags) > 0:
		return optionNotImplemented(method, "Tags", "a sandbox tags API")
	}
	return nil
}

func (s sandboxService) list(ctx context.Context) iter.Seq2[*Sandbox, error] {
	return func(yield func(*Sandbox, error) bool) {
		for sb, err := range s.c.do.List(nil).All(ctx) {
			if err != nil {
				yield(nil, convertError(err))
				return
			}
			if !yield(newSandbox(s.c, sb, ""), nil) {
				return
			}
		}
	}
}

// Exec mirrors Sandbox.Exec. It starts the command and returns at once with
// its stdin, stdout and stderr streams. PTY is not supported.
func (sb *Sandbox) Exec(ctx context.Context, command []string, params *SandboxExecParams) (*ContainerProcess, error) {
	if len(command) == 0 {
		return nil, InvalidError{Exception: "command must not be empty"}
	}
	if params == nil {
		params = &SandboxExecParams{}
	}
	if params.PTY {
		return nil, optionNotImplemented("Sandbox.Exec", "PTY", "a PTY API")
	}
	secrets, err := secretEnv(params.Secrets)
	if err != nil {
		return nil, err
	}
	cwd := params.Workdir
	if cwd == "" {
		cwd = sb.workdir
	}
	cp, err := sb.start(ctx, command, cwd, joinEnv(secrets, params.Env), params.Timeout,
		params.Stdout != Ignore, params.Stderr != Ignore)
	return cp, convertError(err)
}

func (sb *Sandbox) start(ctx context.Context, argv []string, cwd string, env map[string]string, timeout time.Duration, pipeOut, pipeErr bool) (*ContainerProcess, error) {
	cmd, err := sb.sb.Commands.Start(ctx, argv[0], &sandbox.StartOptions{
		Args: argv[1:], Cwd: cwd, EnvVars: env, Stdin: true, TimeoutSeconds: shared.CeilSeconds(timeout),
	})
	if err != nil {
		return nil, err
	}
	cp := newProcess(ctx, cmd, pipeOut, pipeErr)
	sb.mu.Lock()
	sb.procs = append(sb.procs, cp)
	sb.mu.Unlock()
	return cp, nil
}

// Poll mirrors Sandbox.Poll. It returns nil while the sandbox is alive. With a
// Command it returns that command's exit code once it ends; without one it
// returns 0 once the sandbox is gone.
func (sb *Sandbox) Poll(ctx context.Context, params *SandboxPollParams) (*int, error) {
	if sb.entry != nil {
		select {
		case <-sb.entry.done:
			code := sb.entry.code
			return &code, nil
		default:
			return nil, nil
		}
	}
	if err := sb.sb.Refresh(ctx); err != nil {
		if sandbox.IsNotFound(err) {
			zero := 0
			return &zero, nil
		}
		return nil, convertError(err)
	}
	return nil, nil
}

// Wait mirrors Sandbox.Wait. With a Command it waits for it to exit; without
// one it waits until the sandbox is gone and returns 0.
func (sb *Sandbox) Wait(ctx context.Context, params *SandboxWaitParams) (int, error) {
	if sb.entry != nil {
		return sb.entry.Wait(ctx, nil)
	}
	for {
		code, err := sb.Poll(ctx, nil)
		if err != nil {
			return 0, err
		}
		if code != nil {
			return *code, nil
		}
		select {
		case <-ctx.Done():
			return 0, convertError(ctx.Err())
		case <-time.After(2 * time.Second):
		}
	}
}

// Terminate mirrors Sandbox.Terminate. It deletes the sandbox. The returned
// code is the Command's exit code when it ended, else 0.
func (sb *Sandbox) Terminate(ctx context.Context, params *SandboxTerminateParams) (int, error) {
	if _, err := sb.sb.Kill(ctx); err != nil {
		return 0, convertError(err)
	}
	sb.mu.Lock()
	sb.killed = true
	sb.mu.Unlock()
	sb.release()
	if sb.entry != nil {
		select {
		case <-sb.entry.done:
			return sb.entry.code, nil
		default:
		}
	}
	return 0, nil
}

// Detach mirrors Sandbox.Detach. It stops the local streams and leaves the
// sandbox running.
func (sb *Sandbox) Detach() error {
	sb.release()
	return nil
}

func (sb *Sandbox) release() {
	sb.mu.Lock()
	procs := append([]*ContainerProcess{}, sb.procs...)
	sb.mu.Unlock()
	for _, p := range procs {
		p.cancel()
	}
}

// WaitUntilReady mirrors Sandbox.WaitUntilReady: a DigitalOcean sandbox is
// ready when it runs.
func (sb *Sandbox) WaitUntilReady(ctx context.Context, timeout time.Duration, params *SandboxWaitUntilReadyParams) error {
	return convertError(sb.sb.WaitUntilRunning(ctx, &sandbox.WaitOptions{Timeout: timeout}))
}

// ExperimentalPause mirrors Sandbox.ExperimentalPause: it pauses the sandbox.
func (sb *Sandbox) ExperimentalPause(ctx context.Context, params *SandboxExperimentalPauseParams) error {
	return convertError(sb.sb.Pause(ctx))
}

// ExperimentalResume mirrors Sandbox.ExperimentalResume: it resumes the sandbox.
func (sb *Sandbox) ExperimentalResume(ctx context.Context, params *SandboxExperimentalResumeParams) error {
	return convertError(sb.sb.Resume(ctx))
}

// SandboxSnapshot mirrors modal.SandboxSnapshot: a DigitalOcean snapshot.
type SandboxSnapshot struct {
	SnapshotID string
}

// ExperimentalSnapshot mirrors Sandbox.ExperimentalSnapshot. It captures the
// sandbox as a DigitalOcean snapshot and waits for it. A sandbox cannot be
// created from the snapshot yet.
func (sb *Sandbox) ExperimentalSnapshot(ctx context.Context, params *SandboxExperimentalSnapshotParams) (*SandboxSnapshot, error) {
	snap, err := sb.sb.CreateSnapshot(ctx, nil)
	if err != nil {
		return nil, convertError(err)
	}
	if err := snap.WaitUntilReady(ctx, nil); err != nil {
		return nil, convertError(err)
	}
	return &SandboxSnapshot{SnapshotID: snap.ID()}, nil
}

// snapshotService implements SandboxSnapshotService.
type snapshotService struct{ c *Client }

// FromID mirrors SandboxSnapshotService.FromID.
func (s snapshotService) FromID(ctx context.Context, snapshotID string, params *SandboxSnapshotFromIDParams) (*SandboxSnapshot, error) {
	snap, err := s.c.do.GetSnapshot(ctx, snapshotID)
	if err != nil {
		return nil, convertError(err)
	}
	return &SandboxSnapshot{SnapshotID: snap.ID()}, nil
}
