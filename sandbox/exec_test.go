package sandbox_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/digitalocean/godo/sandbox"
)

const execPath = "POST /v2/sandboxes/sb-1/exec"

func execExit(code int, durationMS string) string {
	if code == 0 {
		return ev(fmt.Sprintf(`{"exit":{"duration_ms":%q}}`, durationMS))
	}
	return ev(fmt.Sprintf(`{"exit":{"exit_code":%d,"duration_ms":%q}}`, code, durationMS))
}

func TestExecRunCollectsOutputAndReturnsTheExitCode(t *testing.T) {
	a, sb := sandboxWithCommands(t)
	a.handlers[execPath] = reply(200,
		stderr([]byte("oops\n"))+ev(`{"keepalive":{}}`)+stdout([]byte("hel"))+stdout([]byte("lo\n"))+execExit(3, "39"))

	var out, errOut []string
	res, err := sb.Exec.Run(context.Background(), []string{"sh", "-c", "x"}, &sandbox.ExecOptions{
		Cwd:            "/w",
		EnvVars:        map[string]string{"A": "1"},
		TimeoutSeconds: 30,
		Stdin:          []byte("in"),
		OnStdout:       func(s string) { out = append(out, s) },
		OnStderr:       func(s string) { errOut = append(errOut, s) },
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.ExitCode != 3 || res.Stdout != "hello\n" || res.Stderr != "oops\n" || res.DurationMS != 39 {
		t.Errorf("result = %+v", res)
	}
	if strings.Join(out, "|") != "hel|lo\n" || strings.Join(errOut, "|") != "oops\n" {
		t.Errorf("callbacks: out=%q err=%q", out, errOut)
	}
	body := a.bodies[len(a.bodies)-1]
	argv, _ := body["argv"].([]interface{})
	if len(argv) != 3 || argv[0] != "sh" || body["workdir"] != "/w" || body["timeout_seconds"] != "30" ||
		body["stdin"] != enc([]byte("in")) || body["keep_stdin_open"] != nil || body["pty"] != nil {
		t.Errorf("body = %+v", body)
	}
}

func TestExecRunShellUsesBashAndAMissingExitCodeIsZero(t *testing.T) {
	a, sb := sandboxWithCommands(t)
	// The envelopes arrive without newlines between them.
	a.handlers[execPath] = reply(200,
		strings.ReplaceAll(stdout([]byte("hi\n"))+execExit(0, "5"), "\n", ""))

	res, err := sb.Exec.RunShell(context.Background(), `echo hi`, nil)
	if err != nil || res.ExitCode != 0 || res.Stdout != "hi\n" || res.DurationMS != 5 {
		t.Fatalf("RunShell = %+v, %v", res, err)
	}
	argv, _ := a.bodies[len(a.bodies)-1]["argv"].([]interface{})
	if len(argv) != 3 || argv[0] != "bash" || argv[1] != "-lc" || argv[2] != "echo hi" {
		t.Errorf("argv = %v", argv)
	}
	if _, err := sb.Exec.RunShell(context.Background(), " ", nil); err == nil {
		t.Error("a blank command returned no error")
	}
}

func TestExecRunKeepsAMultiByteCharacterTogether(t *testing.T) {
	a, sb := sandboxWithCommands(t)
	b := []byte("héllo")
	a.handlers[execPath] = reply(200, stdout(b[:2])+stdout(b[2:])+execExit(0, "1"))
	var chunks []string
	res, err := sb.Exec.Run(context.Background(), []string{"cat"}, &sandbox.ExecOptions{
		OnStdout: func(s string) { chunks = append(chunks, s) },
	})
	if err != nil || res.Stdout != "héllo" {
		t.Fatalf("Run = %+v, %v", res, err)
	}
	for _, c := range chunks {
		if strings.ContainsRune(c, '�') {
			t.Errorf("a chunk split a character: %q", chunks)
		}
	}
}

func TestExecStreamErrorsAndEarlyEnd(t *testing.T) {
	a, sb := sandboxWithCommands(t)
	ctx := context.Background()

	a.handlers[execPath] = reply(200, stdout([]byte("a"))+`{"error":{"http_code":500,"message":"gateway failed"}}`+"\n")
	if _, err := sb.Exec.Run(ctx, []string{"ls"}, nil); err == nil || err.Error() != "gateway failed" {
		t.Errorf("error envelope: err = %v", err)
	}

	a.handlers[execPath] = reply(200, stdout([]byte("partial")))
	if _, err := sb.Exec.Run(ctx, []string{"ls"}, nil); err == nil || !strings.Contains(err.Error(), "ended before the process exited") {
		t.Errorf("early end: err = %v", err)
	}

	a.handlers[execPath] = reply(400, `{"id":"bad_request","message":"this sandbox's template image does not support interactive stdin or PTY exec; create a new sandbox","request_id":"r"}`)
	if _, err := sb.Exec.Start(ctx, nil, &sandbox.ExecStartOptions{PTY: &sandbox.PTYOptions{}}); err == nil || !sandbox.IsInvalidRequest(err) {
		t.Errorf("pre-stream 400: err = %v", err)
	}
}

func TestExecValidatesBeforeSending(t *testing.T) {
	a, sb := sandboxWithCommands(t)
	ctx := context.Background()
	before := len(a.calls)
	cases := map[string]func() error{
		"empty argv": func() error { _, err := sb.Exec.Run(ctx, nil, nil); return err },
		"blank argv0": func() error {
			_, err := sb.Exec.Run(ctx, []string{" "}, nil)
			return err
		},
		"empty env key": func() error {
			_, err := sb.Exec.Run(ctx, []string{"ls"}, &sandbox.ExecOptions{EnvVars: map[string]string{"": "x"}})
			return err
		},
		"negative timeout": func() error {
			_, err := sb.Exec.Run(ctx, []string{"ls"}, &sandbox.ExecOptions{TimeoutSeconds: -1})
			return err
		},
		"pty rows": func() error {
			_, err := sb.Exec.Start(ctx, nil, &sandbox.ExecStartOptions{PTY: &sandbox.PTYOptions{Rows: 1001}})
			return err
		},
		"empty argv start": func() error { _, err := sb.Exec.Start(ctx, nil, nil); return err },
	}
	for name, f := range cases {
		if f() == nil {
			t.Errorf("%s returned no error", name)
		}
	}
	if len(a.calls) != before {
		t.Errorf("validation sent requests: %v", a.calls[before:])
	}
}

func TestExecStartReadsStartedAndTracksTheStdinOffset(t *testing.T) {
	a, sb := sandboxWithCommands(t)
	a.handlers[execPath] = reply(200,
		ev(`{"keepalive":{}}`)+ev(`{"started":{"exec_id":"ex-1"}}`)+stdout([]byte("hi\n"))+execExit(0, "3892"))
	var mu sync.Mutex
	offsets := []string{`{"offset":"3"}`, `{"offset":"3"}`}
	a.handlers["POST /v2/sandboxes/sb-1/exec/ex-1/stdin"] = func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		fmt.Fprint(w, offsets[0])
		offsets = offsets[1:]
	}
	ctx := context.Background()

	s, err := sb.Exec.Start(ctx, []string{"cat"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if s.ExecID != "ex-1" {
		t.Fatalf("ExecID = %q", s.ExecID)
	}
	if a.bodies[len(a.bodies)-1]["keep_stdin_open"] != true {
		t.Errorf("start body = %+v", a.bodies[len(a.bodies)-1])
	}
	if err := s.WriteStdin(ctx, []byte("hi\n"), false); err != nil {
		t.Fatal(err)
	}
	if err := s.CloseStdin(ctx); err != nil {
		t.Fatal(err)
	}
	bodies := a.bodies[len(a.bodies)-2:]
	if bodies[0]["data"] != enc([]byte("hi\n")) || bodies[0]["offset"] != "0" || bodies[0]["eof"] != nil ||
		bodies[1]["offset"] != "3" || bodies[1]["eof"] != true || bodies[1]["data"] != nil {
		t.Errorf("stdin bodies = %+v", bodies)
	}

	var types []sandbox.EventType
	for e, err := range s.Events(ctx) {
		if err != nil {
			t.Fatal(err)
		}
		types = append(types, e.Type)
	}
	if fmt.Sprint(types) != "[started stdout exit]" {
		t.Errorf("types = %v", types)
	}
	for _, err := range s.Events(ctx) {
		if err == nil || !strings.Contains(err.Error(), "already read") {
			t.Errorf("second Events: err = %v", err)
		}
	}
}

func TestExecStartWithPTYAndResize(t *testing.T) {
	a, sb := sandboxWithCommands(t)
	a.handlers[execPath] = reply(200, ev(`{"started":{"exec_id":"ex-1"}}`)+execExit(0, "1"))
	a.handlers["POST /v2/sandboxes/sb-1/exec/ex-1/resize"] = reply(200, `{}`)
	ctx := context.Background()

	no := false
	s, err := sb.Exec.Start(ctx, nil, &sandbox.ExecStartOptions{
		KeepStdinOpen: &no, PTY: &sandbox.PTYOptions{Rows: 40, Cols: 120, Term: "xterm"},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	body := a.bodies[len(a.bodies)-1]
	pty, _ := body["pty"].(map[string]interface{})
	if body["argv"] != nil || body["keep_stdin_open"] != nil ||
		pty["rows"] != float64(40) || pty["cols"] != float64(120) || pty["term"] != "xterm" {
		t.Errorf("start body = %+v", body)
	}
	if err := s.Resize(ctx, 50, 200); err != nil {
		t.Fatal(err)
	}
	if got := a.bodies[len(a.bodies)-1]; got["rows"] != float64(50) || got["cols"] != float64(200) {
		t.Errorf("resize body = %+v", got)
	}
	n := len(a.calls)
	if s.Resize(ctx, 0, 10) == nil || s.Resize(ctx, 10, 1001) == nil {
		t.Error("an out-of-range size returned no error")
	}
	if len(a.calls) != n {
		t.Error("an invalid resize sent a request")
	}
	res, err := s.Wait(ctx, nil)
	if err != nil || res.ExitCode != 0 || res.DurationMS != 1 {
		t.Errorf("Wait = %+v, %v", res, err)
	}
}

func TestExecPlainStartHasNoIDAndRejectsStdin(t *testing.T) {
	a, sb := sandboxWithCommands(t)
	a.handlers[execPath] = reply(200, stdout([]byte("x"))+execExit(1, "2"))
	ctx := context.Background()
	no := false
	s, err := sb.Exec.Start(ctx, []string{"ls"}, &sandbox.ExecStartOptions{KeepStdinOpen: &no})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if s.ExecID != "" {
		t.Errorf("ExecID = %q", s.ExecID)
	}
	if s.WriteStdin(ctx, []byte("a"), false) == nil || s.Resize(ctx, 10, 10) == nil {
		t.Error("stdin or resize without an exec id returned no error")
	}
	res, err := s.Wait(ctx, nil)
	if err != nil || res.ExitCode != 1 || res.Stdout != "x" {
		t.Errorf("Wait = %+v, %v", res, err)
	}
}

func TestExecOKAndCheck(t *testing.T) {
	a, sb := sandboxWithCommands(t)
	ctx := context.Background()

	a.handlers[execPath] = reply(200, stdout([]byte("fine\n"))+execExit(0, "1"))
	res, err := sb.Exec.RunShell(ctx, "true", &sandbox.ExecOptions{Check: true})
	if err != nil || !res.OK {
		t.Errorf("zero exit with Check: res=%+v err=%v", res, err)
	}

	a.handlers[execPath] = reply(200, stderr([]byte("bad\n"))+execExit(4, "1"))
	res, err = sb.Exec.RunShell(ctx, "exit 4", nil)
	if err != nil || res.OK || res.ExitCode != 4 {
		t.Errorf("non-zero without Check: res=%+v err=%v", res, err)
	}

	res, err = sb.Exec.RunShell(ctx, "exit 4", &sandbox.ExecOptions{Check: true})
	var ce *sandbox.CommandError
	if !errors.As(err, &ce) {
		t.Fatalf("err = %v, want *CommandError", err)
	}
	if ce.Result == nil || ce.Result != res || res.ExitCode != 4 || res.OK || res.Stderr != "bad\n" {
		t.Errorf("result = %+v, error result = %+v", res, ce.Result)
	}
	if err.Error() != "sandbox: command exited with code 4" {
		t.Errorf("message = %q", err.Error())
	}
}

func TestExecTimeout(t *testing.T) {
	a, sb := sandboxWithCommands(t)
	ctx := context.Background()
	a.handlers[execPath] = reply(200, execExit(0, "1"))

	for d, want := range map[time.Duration]string{90 * time.Second: "90", 1500 * time.Millisecond: "2", time.Millisecond: "1"} {
		if _, err := sb.Exec.Run(ctx, []string{"ls"}, &sandbox.ExecOptions{Timeout: d}); err != nil {
			t.Fatal(err)
		}
		if got := a.bodies[len(a.bodies)-1]["timeout_seconds"]; got != want {
			t.Errorf("Timeout %v: timeout_seconds = %v, want %s", d, got, want)
		}
	}

	s, err := sb.Exec.Start(ctx, []string{"ls"}, &sandbox.ExecStartOptions{Timeout: 1500 * time.Millisecond, KeepStdinOpen: new(bool)})
	if err != nil {
		t.Fatal(err)
	}
	s.Close()
	if got := a.bodies[len(a.bodies)-1]["timeout_seconds"]; got != "2" {
		t.Errorf("Start: timeout_seconds = %v", got)
	}

	before := len(a.calls)
	for name, f := range map[string]func() error{
		"negative": func() error {
			_, err := sb.Exec.Run(ctx, []string{"ls"}, &sandbox.ExecOptions{Timeout: -time.Second})
			return err
		},
		"both": func() error {
			_, err := sb.Exec.Run(ctx, []string{"ls"}, &sandbox.ExecOptions{Timeout: time.Second, TimeoutSeconds: 5})
			return err
		},
		"negative start": func() error {
			_, err := sb.Exec.Start(ctx, []string{"ls"}, &sandbox.ExecStartOptions{Timeout: -time.Second})
			return err
		},
		"both start": func() error {
			_, err := sb.Exec.Start(ctx, []string{"ls"}, &sandbox.ExecStartOptions{Timeout: time.Second, TimeoutSeconds: 5})
			return err
		},
	} {
		if f() == nil {
			t.Errorf("%s: no error", name)
		}
	}
	if len(a.calls) != before {
		t.Errorf("validation sent requests: %v", a.calls[before:])
	}
}
