package sandbox

import (
	"context"
	"errors"
	"fmt"
	"io"
	"iter"
	"strings"
	"sync"

	"github.com/digitalocean/godo"
)

// Exec runs processes in one sandbox and streams their output on the same
// request. Get it from Sandbox.Exec.
//
// Prefer Exec.Run over Commands.Run for short commands: the output arrives on
// the request that starts the process, so there is no attach race.
type Exec struct {
	client    *Client
	sandboxID string
}

// ExecOptions are the options for Exec.Run and Exec.RunShell. Every field is
// optional.
type ExecOptions struct {
	// Cwd is the working directory.
	Cwd string
	// EnvVars are environment variables for the process.
	EnvVars map[string]string
	// TimeoutSeconds bounds the run. Zero means the service default.
	TimeoutSeconds int64
	// Stdin is sent to the process with the start request.
	Stdin []byte
	// OnStdout and OnStderr are called with each chunk of output as it arrives.
	OnStdout func(text string)
	OnStderr func(text string)
}

// PTYOptions turn a terminal on for Exec.Start. Zero fields use the defaults:
// 24 rows, 80 columns and xterm-256color.
type PTYOptions struct {
	Rows int
	Cols int
	Term string
}

// ExecStartOptions are the options for Exec.Start. Every field is optional.
type ExecStartOptions struct {
	// Cwd is the working directory.
	Cwd string
	// EnvVars are environment variables for the process.
	EnvVars map[string]string
	// TimeoutSeconds bounds the run. Zero means the service default.
	TimeoutSeconds int64
	// Stdin is a first stdin chunk sent with the start request.
	Stdin []byte
	// KeepStdinOpen keeps stdin open for ExecSession.WriteStdin. Nil means
	// true.
	KeepStdinOpen *bool
	// PTY allocates a terminal when set. With a PTY, an empty argv starts the
	// guest's default shell.
	PTY *PTYOptions
}

// ExecResult is the outcome of a process that exited. A non-zero ExitCode is
// not an error.
type ExecResult struct {
	ExitCode   int32
	Stdout     string
	Stderr     string
	DurationMS int64
	// Status is a short status for the exit, such as "exited".
	Status string
	// Error is set when the process failed abnormally.
	Error string
}

// ExecEvent is one event of an exec's output. Type is EventStarted,
// EventStdout, EventStderr or EventExit.
type ExecEvent struct {
	Type EventType
	// ExecID is set on EventStarted.
	ExecID string
	// Data and Text are set on EventStdout and EventStderr. Text is the chunk
	// as is and can split a multi-byte character; ExecSession.Wait does not.
	Data []byte
	Text string
	// ExitCode, DurationMS, Status and Error are set on EventExit.
	ExitCode   int32
	DurationMS int64
	Status     string
	Error      string
}

const maxTerminalSize = 1000

// Run runs an executable and its arguments and waits for it to exit. A
// non-zero exit code is returned in the ExecResult, not as an error. A nil opts
// uses the defaults.
func (e *Exec) Run(ctx context.Context, argv []string, opts *ExecOptions) (*ExecResult, error) {
	if opts == nil {
		opts = &ExecOptions{}
	}
	req := &godo.SandboxExecStartRequest{
		Argv:           argv,
		Workdir:        opts.Cwd,
		EnvVars:        opts.EnvVars,
		TimeoutSeconds: opts.TimeoutSeconds,
		Stdin:          opts.Stdin,
	}
	if err := validateExec(req); err != nil {
		return nil, err
	}
	stream, _, err := e.client.API.Sandboxes.StartExec(ctx, e.sandboxID, req)
	if err != nil {
		return nil, err
	}
	s := &ExecSession{client: e.client, sandboxID: e.sandboxID, stream: stream}
	defer s.Close()
	return s.Wait(ctx, &AttachOptions{OnStdout: opts.OnStdout, OnStderr: opts.OnStderr})
}

// RunShell runs a shell command (bash -lc <command>) and waits for it to exit.
func (e *Exec) RunShell(ctx context.Context, command string, opts *ExecOptions) (*ExecResult, error) {
	if strings.TrimSpace(command) == "" {
		return nil, errors.New("sandbox: command must not be empty")
	}
	return e.Run(ctx, []string{"bash", "-lc", command}, opts)
}

// Start starts an executable for interactive use and returns once it runs.
// With KeepStdinOpen (the default) or PTY, it reads the started event first,
// so ExecSession.ExecID is set. ctx governs the stream for the whole session:
// when it ends, the stream disconnects. A nil opts uses the defaults.
func (e *Exec) Start(ctx context.Context, argv []string, opts *ExecStartOptions) (*ExecSession, error) {
	if opts == nil {
		opts = &ExecStartOptions{}
	}
	keepOpen := opts.KeepStdinOpen == nil || *opts.KeepStdinOpen
	req := &godo.SandboxExecStartRequest{
		Argv:           argv,
		Workdir:        opts.Cwd,
		EnvVars:        opts.EnvVars,
		TimeoutSeconds: opts.TimeoutSeconds,
		KeepStdinOpen:  keepOpen,
		Stdin:          opts.Stdin,
	}
	if opts.PTY != nil {
		req.PTY = &godo.SandboxExecPTY{Rows: int64(opts.PTY.Rows), Cols: int64(opts.PTY.Cols), Term: opts.PTY.Term}
	}
	if err := validateExec(req); err != nil {
		return nil, err
	}
	stream, _, err := e.client.API.Sandboxes.StartExec(ctx, e.sandboxID, req)
	if err != nil {
		return nil, err
	}
	s := &ExecSession{client: e.client, sandboxID: e.sandboxID, stream: stream}
	if !keepOpen && req.PTY == nil {
		return s, nil
	}
	// Read up to the started event. Keep any other event for the consumer.
	for {
		raw, err := stream.Next()
		if err == io.EOF {
			stream.Close()
			return nil, errors.New("sandbox: the exec stream ended before the process started")
		}
		if err != nil {
			stream.Close()
			if ctx.Err() != nil {
				err = ctx.Err()
			}
			return nil, err
		}
		ev := toExecEvent(raw)
		if ev == nil {
			continue
		}
		if ev.Type == EventStarted {
			s.ExecID = ev.ExecID
		}
		s.pending = append(s.pending, ev)
		if ev.Type == EventStarted || ev.Type == EventExit {
			return s, nil
		}
	}
}

func validateExec(req *godo.SandboxExecStartRequest) error {
	if len(req.Argv) == 0 && req.PTY == nil {
		return errors.New("sandbox: argv must not be empty unless a PTY is requested")
	}
	if len(req.Argv) > 0 && strings.TrimSpace(req.Argv[0]) == "" {
		return errors.New("sandbox: the executable (argv[0]) must not be empty")
	}
	for k := range req.EnvVars {
		if k == "" {
			return errors.New("sandbox: environment variable names must not be empty")
		}
	}
	if req.TimeoutSeconds < 0 {
		return errors.New("sandbox: timeout must not be negative")
	}
	if req.PTY != nil {
		if err := checkTerminalSize(req.PTY.Rows, req.PTY.Cols, true); err != nil {
			return err
		}
	}
	return nil
}

func checkTerminalSize(rows, cols int64, zeroOK bool) error {
	for _, v := range []struct {
		name string
		n    int64
	}{{"rows", rows}, {"cols", cols}} {
		if zeroOK && v.n == 0 {
			continue
		}
		if v.n < 1 || v.n > maxTerminalSize {
			return fmt.Errorf("sandbox: %s must be between 1 and %d, got %d", v.name, maxTerminalSize, v.n)
		}
	}
	return nil
}

// ExecSession is a running exec: read its output, write its stdin, resize its
// terminal. Close it when done.
type ExecSession struct {
	client    *Client
	sandboxID string
	stream    *godo.SandboxExecStream

	// ExecID identifies the exec. It is empty for a plain exec, which was
	// started without KeepStdinOpen or PTY.
	ExecID string

	pending []*ExecEvent

	mu       sync.Mutex // guards consumed
	consumed bool

	stdinMu sync.Mutex // serializes stdin writes and guards offset
	offset  int64
}

// Events yields the exec's events as they happen: EventStarted (when the exec
// has an ID), EventStdout, EventStderr and a final EventExit. Keepalives are
// dropped. Only one consumer can read the events: a second call yields an
// error. Stopping the loop or ending ctx disconnects; the process may keep
// running. A failure is yielded as the error, and the sequence ends. A stream
// that ends before an EventExit is an error.
func (s *ExecSession) Events(ctx context.Context) iter.Seq2[*ExecEvent, error] {
	return func(yield func(*ExecEvent, error) bool) {
		s.mu.Lock()
		consumed := s.consumed
		s.consumed = true
		s.mu.Unlock()
		if consumed {
			yield(nil, errors.New("sandbox: the exec's events were already read"))
			return
		}
		defer s.stream.Close()
		stop := context.AfterFunc(ctx, func() { s.stream.Close() })
		defer stop()

		pending := s.pending
		s.pending = nil
		for _, ev := range pending {
			if !yield(ev, nil) || ev.Type == EventExit {
				return
			}
		}
		for {
			raw, err := s.stream.Next()
			if err == io.EOF {
				yield(nil, errors.New("sandbox: the exec stream ended before the process exited; its outcome is unknown"))
				return
			}
			if err != nil {
				if ctx.Err() != nil {
					err = ctx.Err()
				}
				yield(nil, err)
				return
			}
			ev := toExecEvent(raw)
			if ev == nil {
				continue
			}
			if !yield(ev, nil) || ev.Type == EventExit {
				return
			}
		}
	}
}

func toExecEvent(raw *godo.SandboxExecEvent) *ExecEvent {
	switch {
	case raw.Started != nil:
		return &ExecEvent{Type: EventStarted, ExecID: raw.Started.ExecID}
	case len(raw.Stdout) > 0:
		return &ExecEvent{Type: EventStdout, Data: raw.Stdout, Text: string(raw.Stdout)}
	case len(raw.Stderr) > 0:
		return &ExecEvent{Type: EventStderr, Data: raw.Stderr, Text: string(raw.Stderr)}
	case raw.Exit != nil:
		return &ExecEvent{Type: EventExit, ExitCode: raw.Exit.ExitCode, DurationMS: raw.Exit.DurationMS,
			Status: raw.Exit.Status, Error: raw.Exit.Error}
	}
	return nil
}

// Wait reads the exec's events until it exits and returns its output and exit
// code. A nil opts only collects the output. Text is decoded so that a
// multi-byte character split across chunks stays whole.
func (s *ExecSession) Wait(ctx context.Context, opts *AttachOptions) (*ExecResult, error) {
	if opts == nil {
		opts = &AttachOptions{}
	}
	var stdout, stderr strings.Builder
	out, errDec := &textDecoder{}, &textDecoder{}
	for ev, err := range s.Events(ctx) {
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
			if text := out.flush(); text != "" {
				stdout.WriteString(text)
				if opts.OnStdout != nil {
					opts.OnStdout(text)
				}
			}
			if text := errDec.flush(); text != "" {
				stderr.WriteString(text)
				if opts.OnStderr != nil {
					opts.OnStderr(text)
				}
			}
			return &ExecResult{
				ExitCode:   ev.ExitCode,
				Stdout:     stdout.String(),
				Stderr:     stderr.String(),
				DurationMS: ev.DurationMS,
				Status:     ev.Status,
				Error:      ev.Error,
			}, nil
		}
	}
	return nil, errors.New("sandbox: stopped waiting for the exec before it exited")
}

// WriteStdin writes to the exec's stdin. Set eof to close stdin after writing
// (on a PTY it sends VEOF). The session tracks the stdin offset; concurrent
// writes are sent one at a time.
func (s *ExecSession) WriteStdin(ctx context.Context, data []byte, eof bool) error {
	if s.ExecID == "" {
		return errors.New("sandbox: this exec has no ID: start it with KeepStdinOpen or PTY to write stdin")
	}
	s.stdinMu.Lock()
	defer s.stdinMu.Unlock()
	offset, _, err := s.client.API.Sandboxes.WriteExecStdin(ctx, s.sandboxID, s.ExecID,
		&godo.SandboxExecStdinRequest{Data: data, Offset: s.offset, EOF: eof})
	if err != nil {
		return err
	}
	s.offset = offset
	return nil
}

// CloseStdin closes the exec's stdin.
func (s *ExecSession) CloseStdin(ctx context.Context) error {
	return s.WriteStdin(ctx, nil, true)
}

// Resize resizes the exec's terminal. The exec must have been started with
// PTY. Rows and cols must be between 1 and 1000.
func (s *ExecSession) Resize(ctx context.Context, rows, cols int) error {
	if err := checkTerminalSize(int64(rows), int64(cols), false); err != nil {
		return err
	}
	if s.ExecID == "" {
		return errors.New("sandbox: this exec has no ID: start it with PTY to resize it")
	}
	_, err := s.client.API.Sandboxes.ResizeExec(ctx, s.sandboxID, s.ExecID, int64(rows), int64(cols))
	return err
}

// Close disconnects the stream. The process may keep running.
func (s *ExecSession) Close() error {
	return s.stream.Close()
}
