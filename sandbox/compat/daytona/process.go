package daytona

import (
	"context"
	"strings"
	"sync"

	"github.com/digitalocean/godo/sandbox"
	"github.com/digitalocean/godo/sandbox/compat/daytona/options"
	"github.com/digitalocean/godo/sandbox/compat/daytona/types"
	"github.com/digitalocean/godo/sandbox/compat/internal/shared"
)

// ProcessService mirrors daytona.ProcessService. Only ExecuteCommand works.
type ProcessService struct {
	sb *sandbox.Sandbox
}

// ExecuteCommand mirrors ProcessService.ExecuteCommand. It runs the command
// with bash -lc and waits for it. Result holds stdout and stderr in the order
// they arrived. A non-zero exit code is in ExitCode, not an error.
func (p *ProcessService) ExecuteCommand(ctx context.Context, command string, opts ...func(*options.ExecuteCommand)) (*types.ExecuteResponse, error) {
	eo := &options.ExecuteCommand{}
	for _, o := range opts {
		o(eo)
	}
	run := &sandbox.RunOptions{EnvVars: eo.Env}
	if eo.Cwd != nil {
		run.Cwd = *eo.Cwd
	}
	if eo.Timeout != nil {
		run.TimeoutSeconds = shared.CeilSeconds(*eo.Timeout)
	}
	var mu sync.Mutex
	var out strings.Builder
	collect := func(text string) {
		mu.Lock()
		out.WriteString(text)
		mu.Unlock()
	}
	run.OnStdout, run.OnStderr = collect, collect
	res, err := p.sb.Commands.Run(ctx, command, run)
	if err != nil {
		return nil, convertError(err)
	}
	return &types.ExecuteResponse{ExitCode: int(res.ExitCode), Result: out.String()}, nil
}
