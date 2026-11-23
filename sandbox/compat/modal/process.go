package modal

import (
	"context"
	"io"
	"sync"

	"github.com/digitalocean/godo/sandbox"
	"github.com/digitalocean/godo/sandbox/compat/internal/shared"
)

// ContainerProcess mirrors modal.ContainerProcess: a command in a sandbox, with
// its standard streams.
type ContainerProcess struct {
	Stdin  io.WriteCloser
	Stdout io.ReadCloser
	Stderr io.ReadCloser

	done   chan struct{}
	code   int
	err    error
	cancel context.CancelFunc
}

type stdinWriter struct {
	cmd  *sandbox.Command
	once sync.Once
	err  error
}

func (w *stdinWriter) Write(p []byte) (int, error) {
	if err := w.cmd.WriteStdin(context.Background(), p, false); err != nil {
		return 0, convertError(err)
	}
	return len(p), nil
}

func (w *stdinWriter) Close() error {
	w.once.Do(func() { w.err = convertError(w.cmd.CloseStdin(context.Background())) })
	return w.err
}

// newProcess connects a started command to Modal-shaped streams. It reads the
// command's events in the background, until the command exits. The streams
// outlive ctx, which only bounded the start.
func newProcess(ctx context.Context, cmd *sandbox.Command, pipeOut, pipeErr bool) *ContainerProcess {
	streamCtx, cancel := context.WithCancel(context.WithoutCancel(ctx))
	outP, errP := shared.NewPipe(), shared.NewPipe()
	cp := &ContainerProcess{
		Stdin:  &stdinWriter{cmd: cmd},
		Stdout: outP,
		Stderr: errP,
		done:   make(chan struct{}),
		cancel: cancel,
	}
	go func() {
		defer func() {
			outP.CloseWrite()
			errP.CloseWrite()
			cancel()
			close(cp.done)
		}()
		for ev, err := range cmd.Events(streamCtx) {
			if err != nil {
				cp.err = convertError(err)
				return
			}
			switch ev.Type {
			case sandbox.EventStdout:
				if pipeOut {
					_, _ = outP.Write(ev.Data)
				}
			case sandbox.EventStderr:
				if pipeErr {
					_, _ = errP.Write(ev.Data)
				}
			case sandbox.EventExit:
				cp.code = int(ev.ExitCode)
			}
		}
	}()
	return cp
}

// Wait mirrors ContainerProcess.Wait: it blocks until the process exits and
// returns its exit code.
func (cp *ContainerProcess) Wait(ctx context.Context, params *ContainerProcessWaitParams) (int, error) {
	select {
	case <-cp.done:
		return cp.code, cp.err
	case <-ctx.Done():
		return 0, convertError(ctx.Err())
	}
}
