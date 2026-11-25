package sandbox

import (
	"context"
	"errors"
	"fmt"
	"io"
	"iter"
	"strings"
	"unicode/utf8"

	"github.com/digitalocean/godo"
)

// Commands starts commands in one sandbox and finds the ones already running.
// Get it from Sandbox.Commands.
type Commands struct {
	client    *Client
	sandboxID string
}

// CommandInfo is a command that is running.
type CommandInfo struct {
	PID int64
	Cmd string
	Cwd string
}

// StartOptions are the options for Start. Every field is optional.
type StartOptions struct {
	// Args are passed to the executable.
	Args []string
	// Cwd is the working directory.
	Cwd string
	// EnvVars are environment variables for the command.
	EnvVars map[string]string
	// Stdin leaves stdin open, so Command.WriteStdin can write to it.
	Stdin bool
	// TimeoutSeconds kills the command after this long. Zero means no timeout.
	TimeoutSeconds int64
}

// RunOptions are the options for Run. Every field is optional.
type RunOptions struct {
	Cwd            string
	EnvVars        map[string]string
	TimeoutSeconds int64
	// OnStdout and OnStderr are called with each chunk of output as it arrives.
	OnStdout func(text string)
	OnStderr func(text string)
}

// AttachOptions are the options for Command.Wait. Every field is optional.
type AttachOptions struct {
	// OnStdout and OnStderr are called with each chunk of output as it arrives.
	OnStdout func(text string)
	OnStderr func(text string)
}

// Result is the outcome of a command that exited. A non-zero ExitCode is not an
// error.
type Result struct {
	PID int64
	// OK is true when ExitCode is 0.
	OK       bool
	ExitCode int32
	Stdout   string
	Stderr   string
	// Status is a short status for the exit, such as "exited".
	Status string
	// Error is set when the process failed abnormally.
	Error string
}

// EventType tells the kind of an Event.
type EventType string

const (
	EventStarted EventType = "started"
	EventStdout  EventType = "stdout"
	EventStderr  EventType = "stderr"
	EventExit    EventType = "exit"
)

// Event is one event of a command's output.
type Event struct {
	Type EventType
	// PID is set on EventStarted.
	PID int64
	// Data and Text are set on EventStdout and EventStderr.
	Data []byte
	Text string
	// ExitCode, Status and Error are set on EventExit.
	ExitCode int32
	Status   string
	Error    string
}

// Start starts an executable and returns once it is running. It keeps running
// after Start returns. For a shell string use Run, or pass "bash" with Args
// {"-lc", "..."}. A nil opts uses the defaults.
func (c *Commands) Start(ctx context.Context, cmd string, opts *StartOptions) (*Command, error) {
	if strings.TrimSpace(cmd) == "" {
		return nil, errors.New("sandbox: command must not be empty")
	}
	if opts == nil {
		opts = &StartOptions{}
	}
	pid, _, err := c.client.API.Sandboxes.StartCommand(ctx, c.sandboxID, &godo.SandboxCommandStartRequest{
		Cmd:            cmd,
		Args:           opts.Args,
		Cwd:            opts.Cwd,
		EnvVars:        opts.EnvVars,
		Stdin:          opts.Stdin,
		TimeoutSeconds: opts.TimeoutSeconds,
	})
	if err != nil {
		return nil, err
	}
	return c.Connect(pid), nil
}

// Run runs a shell command (bash -lc <command>) and waits for it to exit. A
// non-zero exit code is returned in the Result, not as an error.
//
// Output can only be read while the command runs, so a command that exits
// before the stream attaches fails with an error for which IsNotFound is true.
func (c *Commands) Run(ctx context.Context, command string, opts *RunOptions) (*Result, error) {
	if opts == nil {
		opts = &RunOptions{}
	}
	cmd, err := c.Start(ctx, "bash", &StartOptions{
		Args:           []string{"-lc", command},
		Cwd:            opts.Cwd,
		EnvVars:        opts.EnvVars,
		TimeoutSeconds: opts.TimeoutSeconds,
	})
	if err != nil {
		return nil, err
	}
	return cmd.Wait(ctx, &AttachOptions{OnStdout: opts.OnStdout, OnStderr: opts.OnStderr})
}

// List lists the commands that are running.
func (c *Commands) List(ctx context.Context) ([]CommandInfo, error) {
	list, _, err := c.client.API.Sandboxes.ListCommands(ctx, c.sandboxID)
	if err != nil {
		return nil, err
	}
	out := make([]CommandInfo, len(list))
	for i, cmd := range list {
		out[i] = CommandInfo{PID: cmd.PID, Cmd: cmd.Cmd, Cwd: cmd.Cwd}
	}
	return out, nil
}

// Connect returns a handle on a command that is already running. It makes no
// request.
func (c *Commands) Connect(pid int64) *Command {
	return &Command{client: c.client, sandboxID: c.sandboxID, PID: pid}
}

// Command is a command in a sandbox: read its output, write its stdin, stop it.
type Command struct {
	client    *Client
	sandboxID string
	PID       int64
}

// Events yields the command's events as they happen: EventStarted, EventStdout,
// EventStderr and a final EventExit. Keepalives are dropped. Stopping the loop
// disconnects; the command keeps running. A failure is yielded as the error,
// and the sequence ends. A stream that ends before an EventExit is an error.
func (c *Command) Events(ctx context.Context) iter.Seq2[*Event, error] {
	return func(yield func(*Event, error) bool) {
		stream, _, err := c.client.API.Sandboxes.StreamCommand(ctx, c.sandboxID, c.PID)
		if err != nil {
			if IsNotFound(err) {
				err = fmt.Errorf("sandbox: command %d not found or already exited: output can only be read while it runs: %w", c.PID, err)
			}
			yield(nil, err)
			return
		}
		defer stream.Close()

		for {
			raw, err := stream.Next()
			if err == io.EOF {
				yield(nil, fmt.Errorf("sandbox: the stream for command %d ended before the command exited; its outcome is unknown", c.PID))
				return
			}
			if err != nil {
				if ctx.Err() != nil {
					err = ctx.Err()
				}
				yield(nil, err)
				return
			}
			ev := toEvent(raw)
			if ev == nil {
				continue
			}
			if !yield(ev, nil) {
				return
			}
			if ev.Type == EventExit {
				return
			}
		}
	}
}

func toEvent(raw *godo.SandboxCommandEvent) *Event {
	switch {
	case raw.Started != nil:
		return &Event{Type: EventStarted, PID: raw.Started.PID}
	case len(raw.Stdout) > 0:
		return &Event{Type: EventStdout, Data: raw.Stdout, Text: string(raw.Stdout)}
	case len(raw.Stderr) > 0:
		return &Event{Type: EventStderr, Data: raw.Stderr, Text: string(raw.Stderr)}
	case raw.Exit != nil:
		return &Event{Type: EventExit, ExitCode: raw.Exit.ExitCode, Status: raw.Exit.Status, Error: raw.Exit.Error}
	}
	return nil
}

// Wait waits for the command to exit and returns its output and exit code. A
// nil opts only collects the output.
func (c *Command) Wait(ctx context.Context, opts *AttachOptions) (*Result, error) {
	if opts == nil {
		opts = &AttachOptions{}
	}
	var stdout, stderr strings.Builder
	out, errDec := &textDecoder{}, &textDecoder{}
	for ev, err := range c.Events(ctx) {
		if err != nil {
			return nil, err
		}
		switch ev.Type {
		case EventStdout:
			if text := out.decode(ev.Data); text != "" {
				stdout.WriteString(text)
				if opts.OnStdout != nil {
					opts.OnStdout(text)
				}
			}
		case EventStderr:
			if text := errDec.decode(ev.Data); text != "" {
				stderr.WriteString(text)
				if opts.OnStderr != nil {
					opts.OnStderr(text)
				}
			}
		case EventExit:
			stdout.WriteString(out.flush())
			stderr.WriteString(errDec.flush())
			return &Result{
				PID:      c.PID,
				OK:       ev.ExitCode == 0,
				ExitCode: ev.ExitCode,
				Stdout:   stdout.String(),
				Stderr:   stderr.String(),
				Status:   ev.Status,
				Error:    ev.Error,
			}, nil
		}
	}
	return nil, fmt.Errorf("sandbox: stopped waiting for command %d before it exited", c.PID)
}

// Kill sends SIGTERM to the command.
func (c *Command) Kill(ctx context.Context) error {
	return c.Signal(ctx, godo.SandboxCommandSignalTerm)
}

// Signal sends a signal to the command: godo.SandboxCommandSignalTerm or
// godo.SandboxCommandSignalKill.
func (c *Command) Signal(ctx context.Context, signal godo.SandboxCommandSignal) error {
	switch signal {
	case godo.SandboxCommandSignalTerm, godo.SandboxCommandSignalKill:
	default:
		return fmt.Errorf("sandbox: unsupported signal %q: expected SIGTERM or SIGKILL", signal)
	}
	_, err := c.client.API.Sandboxes.SignalCommand(ctx, c.sandboxID, c.PID, signal)
	return err
}

// WriteStdin writes to the command's stdin. The command must have been started
// with StartOptions.Stdin. Set eof to close stdin after writing.
func (c *Command) WriteStdin(ctx context.Context, data []byte, eof bool) error {
	_, err := c.client.API.Sandboxes.WriteCommandStdin(ctx, c.sandboxID, c.PID,
		&godo.SandboxCommandStdinRequest{Data: data, EOF: eof})
	return err
}

// CloseStdin closes the command's stdin.
func (c *Command) CloseStdin(ctx context.Context) error {
	return c.WriteStdin(ctx, nil, true)
}

// textDecoder turns byte chunks into text without splitting a multi-byte
// character that arrives in two chunks.
type textDecoder struct{ pending []byte }

func (d *textDecoder) decode(chunk []byte) string {
	buf := append(d.pending, chunk...)
	end := len(buf)
	// Hold back an incomplete trailing character (at most 3 bytes).
	for i := 1; i <= 3 && i <= len(buf); i++ {
		b := buf[len(buf)-i]
		if utf8.RuneStart(b) {
			if !utf8.FullRune(buf[len(buf)-i:]) {
				end = len(buf) - i
			}
			break
		}
	}
	d.pending = append([]byte(nil), buf[end:]...)
	return string(buf[:end])
}

func (d *textDecoder) flush() string {
	s := string(d.pending)
	d.pending = nil
	return s
}
