package daytona

import (
	"context"
	"strings"
	"time"

	"github.com/digitalocean/godo"
	"github.com/digitalocean/godo/sandbox"
	"github.com/digitalocean/godo/sandbox/compat/daytona/apiclient"
	"github.com/digitalocean/godo/sandbox/compat/internal/shared"
)

// Sandbox mirrors daytona.Sandbox. It wraps a DigitalOcean sandbox; the
// fields that describe Daytona-only features stay at their zero values.
type Sandbox struct {
	ID             string
	Name           string
	OrganizationId string
	Snapshot       *string
	User           string
	Labels         map[string]string
	Public         bool
	Target         string
	Cpu            float32
	Gpu            float32
	Memory         float32
	Disk           float32
	State          apiclient.SandboxState
	ErrorReason    *string
	Recoverable    *bool
	BackupState    *string

	AutoStopInterval    int
	AutoArchiveInterval int
	AutoDeleteInterval  int

	CreatedAt      *string
	UpdatedAt      *string
	LastActivityAt *string

	Env             map[string]string
	BackupCreatedAt *string
	Volumes         []apiclient.SandboxVolume
	BuildInfo       *apiclient.BuildInfo

	NetworkBlockAll  *bool
	NetworkAllowList *string
	DomainAllowList  *string

	FileSystem      *FileSystemService
	Git             *GitService
	Process         *ProcessService
	CodeInterpreter *CodeInterpreterService
	ComputerUse     *ComputerUseService

	client *Client
	sb     *sandbox.Sandbox
}

func newSandbox(c *Client, sb *sandbox.Sandbox) *Sandbox {
	s := &Sandbox{
		client:          c,
		sb:              sb,
		Labels:          map[string]string{},
		FileSystem:      &FileSystemService{},
		Git:             &GitService{},
		CodeInterpreter: &CodeInterpreterService{},
		ComputerUse:     &ComputerUseService{},
		Process:         &ProcessService{sb: sb},
	}
	s.fill()
	return s
}

// fill copies the DigitalOcean record into the Daytona-shaped fields.
func (s *Sandbox) fill() {
	info := s.sb.Info
	s.ID = info.ID
	tpl := info.Template
	s.Snapshot = &tpl
	s.State = mapState(info.State)
	s.CreatedAt = timeString(info.CreatedAt)
	if size, ok := shared.SizeInfo(info.SizeSlug); ok {
		s.Cpu = float32(size.VCPUs)
		s.Memory = float32(size.MemoryMiB) / 1024
	}
}

func (s *Sandbox) bounded(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout <= 0 {
		return ctx, func() {}
	}
	return context.WithTimeout(ctx, timeout)
}

// Start mirrors Sandbox.Start: it resumes the sandbox and waits until it runs.
func (s *Sandbox) Start(ctx context.Context) error {
	return s.StartWithTimeout(ctx, defaultTimeout)
}

// StartWithTimeout mirrors Sandbox.StartWithTimeout.
func (s *Sandbox) StartWithTimeout(ctx context.Context, timeout time.Duration) error {
	ctx, cancel := s.bounded(ctx, timeout)
	defer cancel()
	if err := s.sb.Resume(ctx); err != nil {
		return convertError(err)
	}
	if err := s.sb.WaitUntilRunning(ctx, nil); err != nil {
		return convertError(err)
	}
	s.fill()
	return nil
}

// Stop mirrors Sandbox.Stop: it pauses the sandbox.
func (s *Sandbox) Stop(ctx context.Context) error {
	return s.StopWithTimeout(ctx, defaultTimeout, false)
}

// StopWithTimeout mirrors Sandbox.StopWithTimeout. force has no effect: a
// DigitalOcean sandbox has one way to pause.
func (s *Sandbox) StopWithTimeout(ctx context.Context, timeout time.Duration, force bool) error {
	ctx, cancel := s.bounded(ctx, timeout)
	defer cancel()
	if err := s.sb.Pause(ctx); err != nil {
		return convertError(err)
	}
	s.fill()
	return nil
}

// Pause mirrors Sandbox.Pause.
func (s *Sandbox) Pause(ctx context.Context) error {
	return s.StopWithTimeout(ctx, defaultTimeout, false)
}

// PauseWithTimeout mirrors Sandbox.PauseWithTimeout.
func (s *Sandbox) PauseWithTimeout(ctx context.Context, timeout time.Duration) error {
	return s.StopWithTimeout(ctx, timeout, false)
}

// Delete mirrors Sandbox.Delete.
func (s *Sandbox) Delete(ctx context.Context) error { return s.DeleteWithTimeout(ctx, defaultTimeout) }

// DeleteWithTimeout mirrors Sandbox.DeleteWithTimeout. Deleting a sandbox that
// is already gone is not an error.
func (s *Sandbox) DeleteWithTimeout(ctx context.Context, timeout time.Duration) error {
	ctx, cancel := s.bounded(ctx, timeout)
	defer cancel()
	if _, err := s.sb.Kill(ctx); err != nil {
		return convertError(err)
	}
	s.State = apiclient.SANDBOXSTATE_DESTROYED
	return nil
}

// RefreshData mirrors Sandbox.RefreshData.
func (s *Sandbox) RefreshData(ctx context.Context) error {
	if err := s.sb.Refresh(ctx); err != nil {
		return convertError(err)
	}
	s.fill()
	return nil
}

// WaitForStart mirrors Sandbox.WaitForStart.
func (s *Sandbox) WaitForStart(ctx context.Context, timeout time.Duration) error {
	if err := s.sb.WaitUntilRunning(ctx, &sandbox.WaitOptions{Timeout: timeout}); err != nil {
		return convertError(err)
	}
	s.fill()
	return nil
}

// WaitForStop mirrors Sandbox.WaitForStop: it polls until the sandbox is paused.
func (s *Sandbox) WaitForStop(ctx context.Context, timeout time.Duration) error {
	ctx, cancel := s.bounded(ctx, timeout)
	defer cancel()
	for {
		if err := s.sb.Refresh(ctx); err != nil {
			return convertError(err)
		}
		s.fill()
		if s.sb.State() == godo.SandboxStatePaused {
			return nil
		}
		select {
		case <-ctx.Done():
			return convertError(ctx.Err())
		case <-time.After(time.Second):
		}
	}
}

// ExperimentalCreateSnapshot mirrors Sandbox.ExperimentalCreateSnapshot: it
// captures the sandbox as a DigitalOcean snapshot named name and waits for it.
func (s *Sandbox) ExperimentalCreateSnapshot(ctx context.Context, name string) error {
	return s.ExperimentalCreateSnapshotWithTimeout(ctx, name, 5*time.Minute)
}

// ExperimentalCreateSnapshotWithTimeout mirrors the method of the same name.
func (s *Sandbox) ExperimentalCreateSnapshotWithTimeout(ctx context.Context, name string, timeout time.Duration) error {
	ctx, cancel := s.bounded(ctx, timeout)
	defer cancel()
	snap, err := s.sb.CreateSnapshot(ctx, &sandbox.SnapshotParams{Name: name})
	if err != nil {
		return convertError(err)
	}
	return convertError(snap.WaitUntilReady(ctx, nil))
}

// GetWorkingDir mirrors Sandbox.GetWorkingDir. It runs pwd in the sandbox.
func (s *Sandbox) GetWorkingDir(ctx context.Context) (string, error) {
	return s.outputOf(ctx, "pwd")
}

// GetUserHomeDir mirrors Sandbox.GetUserHomeDir. It reads $HOME in the sandbox.
func (s *Sandbox) GetUserHomeDir(ctx context.Context) (string, error) {
	return s.outputOf(ctx, "printf %s \"$HOME\"")
}

func (s *Sandbox) outputOf(ctx context.Context, command string) (string, error) {
	res, err := s.sb.Commands.Run(ctx, command, nil)
	if err != nil {
		return "", convertError(err)
	}
	return strings.TrimSpace(shared.TrimOutput(res.Stdout)), nil
}
