package sandbox_test

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/digitalocean/godo"
	"github.com/digitalocean/godo/sandbox"
)

func enc(b []byte) string { return base64.StdEncoding.EncodeToString(b) }

func ev(result string) string { return `{"result":` + result + `}` + "\n" }

func stdout(b []byte) string { return ev(fmt.Sprintf(`{"stdout":%q}`, enc(b))) }
func stderr(b []byte) string { return ev(fmt.Sprintf(`{"stderr":%q}`, enc(b))) }
func exit(code int, status string) string {
	return ev(fmt.Sprintf(`{"exit":{"exit_code":%d,"status":%q}}`, code, status))
}

// sandboxWithCommands returns a sandbox whose GET is served and a handler map.
func sandboxWithCommands(t *testing.T) (*api, *sandbox.Sandbox) {
	t.Helper()
	a, c := newAPI(t)
	a.handlers["GET /v2/sandboxes/sb-1"] = reply(200, `{"sandbox":`+raw("sb-1", godo.SandboxStateRunning)+`}`)
	sb, err := c.Get(context.Background(), "sb-1")
	if err != nil {
		t.Fatal(err)
	}
	return a, sb
}

func TestStartSendsTheExecutableAndReturnsAHandle(t *testing.T) {
	a, sb := sandboxWithCommands(t)
	a.handlers["POST /v2/sandboxes/sb-1/commands"] = reply(200, `{"pid":42}`)

	cmd, err := sb.Commands.Start(context.Background(), "sleep", &sandbox.StartOptions{
		Args: []string{"60"}, Cwd: "/workspace", EnvVars: map[string]string{"A": "1"}, Stdin: true, TimeoutSeconds: 30,
	})
	if err != nil || cmd.PID != 42 {
		t.Fatalf("Start = %+v, %v", cmd, err)
	}
	got := a.bodies[len(a.bodies)-1]
	if got["cmd"] != "sleep" || got["cwd"] != "/workspace" || got["stdin"] != true || got["timeout_seconds"] != "30" {
		t.Errorf("body = %+v", got)
	}
	if _, err := sb.Commands.Start(context.Background(), " ", nil); err == nil {
		t.Error("a blank command returned no error")
	}
}

func TestRunStartsBashAndCollectsOutput(t *testing.T) {
	a, sb := sandboxWithCommands(t)
	a.handlers["POST /v2/sandboxes/sb-1/commands"] = reply(200, `{"pid":42}`)
	a.handlers["GET /v2/sandboxes/sb-1/commands/42/stream"] = reply(200,
		ev(`{"started":{"pid":42}}`)+stdout([]byte("Hel"))+ev(`{"keepalive":{}}`)+stdout([]byte("lo\n"))+stderr([]byte("warn\n"))+exit(3, "exited"))

	var out, errOut []string
	res, err := sb.Commands.Run(context.Background(), `echo "Hello"`, &sandbox.RunOptions{
		Cwd:      "/w",
		OnStdout: func(s string) { out = append(out, s) },
		OnStderr: func(s string) { errOut = append(errOut, s) },
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.ExitCode != 3 || res.Stdout != "Hello\n" || res.Stderr != "warn\n" || res.Status != "exited" || res.PID != 42 {
		t.Errorf("result = %+v", res)
	}
	if strings.Join(out, "|") != "Hel|lo\n" || strings.Join(errOut, "|") != "warn\n" {
		t.Errorf("callbacks: out=%q err=%q", out, errOut)
	}
	body := a.bodies[1]
	args, _ := body["args"].([]interface{})
	if body["cmd"] != "bash" || len(args) != 2 || args[0] != "-lc" || args[1] != `echo "Hello"` || body["cwd"] != "/w" {
		t.Errorf("start body = %+v", body)
	}
}

func TestWaitKeepsAMultiByteCharacterTogether(t *testing.T) {
	a, sb := sandboxWithCommands(t)
	bytes := []byte("héllo") // é is two bytes
	a.handlers["GET /v2/sandboxes/sb-1/commands/42/stream"] = reply(200,
		stdout(bytes[:2])+stdout(bytes[2:])+exit(0, ""))

	var chunks []string
	res, err := sb.Commands.Connect(42).Wait(context.Background(), &sandbox.AttachOptions{
		OnStdout: func(s string) { chunks = append(chunks, s) },
	})
	if err != nil || res.Stdout != "héllo" {
		t.Fatalf("Wait = %+v, %v", res, err)
	}
	for _, c := range chunks {
		if strings.ContainsRune(c, '�') {
			t.Errorf("a chunk split a character: %q", chunks)
		}
	}
}

func TestEventsYieldsTypedEventsAndDropsKeepalives(t *testing.T) {
	a, sb := sandboxWithCommands(t)
	a.handlers["GET /v2/sandboxes/sb-1/commands/42/stream"] = reply(200,
		ev(`{"started":{"pid":42}}`)+ev(`{"keepalive":{}}`)+stdout([]byte("a"))+exit(0, "exited"))

	var types []sandbox.EventType
	for e, err := range sb.Commands.Connect(42).Events(context.Background()) {
		if err != nil {
			t.Fatal(err)
		}
		types = append(types, e.Type)
	}
	if fmt.Sprint(types) != "[started stdout exit]" {
		t.Errorf("types = %v", types)
	}
}

func TestStreamEndingBeforeExitIsAnError(t *testing.T) {
	a, sb := sandboxWithCommands(t)
	a.handlers["GET /v2/sandboxes/sb-1/commands/42/stream"] = reply(200, stdout([]byte("partial")))
	_, err := sb.Commands.Connect(42).Wait(context.Background(), nil)
	if err == nil || !strings.Contains(err.Error(), "ended before the command exited") {
		t.Errorf("err = %v", err)
	}
}

func TestExitWithAnErrorIsReturnedInTheResult(t *testing.T) {
	a, sb := sandboxWithCommands(t)
	a.handlers["GET /v2/sandboxes/sb-1/commands/42/stream"] = reply(200,
		ev(`{"exit":{"exit_code":137,"status":"killed","error":"out of memory"}}`))
	res, err := sb.Commands.Connect(42).Wait(context.Background(), nil)
	if err != nil || res.ExitCode != 137 || res.Error != "out of memory" {
		t.Errorf("Wait = %+v, %v", res, err)
	}
}

func TestUnknownPidIsNotFoundAndSaysWhy(t *testing.T) {
	a, sb := sandboxWithCommands(t)
	a.handlers["GET /v2/sandboxes/sb-1/commands/99/stream"] = reply(404, `{"id":"not_found","message":"not found","request_id":"r"}`)
	_, err := sb.Commands.Connect(99).Wait(context.Background(), nil)
	if err == nil || !sandbox.IsNotFound(err) || !strings.Contains(err.Error(), "already exited") {
		t.Errorf("err = %v; IsNotFound = %v", err, sandbox.IsNotFound(err))
	}
}

func TestStreamErrorEnvelopeIsReturned(t *testing.T) {
	a, sb := sandboxWithCommands(t)
	a.handlers["GET /v2/sandboxes/sb-1/commands/42/stream"] = reply(200,
		ev(`{"started":{"pid":42}}`)+`{"error":{"http_code":500,"message":"gateway failed"}}`+"\n")
	_, err := sb.Commands.Connect(42).Wait(context.Background(), nil)
	if err == nil || err.Error() != "gateway failed" {
		t.Errorf("err = %v", err)
	}
}

func TestListCommands(t *testing.T) {
	a, sb := sandboxWithCommands(t)
	a.handlers["GET /v2/sandboxes/sb-1/commands"] = reply(200, `{"commands":[{"pid":42,"cmd":"bash","cwd":"/workspace"},{"pid":43,"cmd":"sleep"}]}`)
	list, err := sb.Commands.List(context.Background())
	if err != nil || len(list) != 2 || list[0] != (sandbox.CommandInfo{PID: 42, Cmd: "bash", Cwd: "/workspace"}) || list[1].PID != 43 {
		t.Errorf("List = %+v, %v", list, err)
	}
}

func TestStdinAndSignals(t *testing.T) {
	a, sb := sandboxWithCommands(t)
	ok := func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, `{}`) }
	a.handlers["POST /v2/sandboxes/sb-1/commands/42/stdin"] = ok
	a.handlers["POST /v2/sandboxes/sb-1/commands/42/signal"] = ok
	ctx := context.Background()
	cmd := sb.Commands.Connect(42)

	if err := cmd.WriteStdin(ctx, []byte("hello\n"), true); err != nil {
		t.Fatal(err)
	}
	if err := cmd.CloseStdin(ctx); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Kill(ctx); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Signal(ctx, godo.SandboxCommandSignalKill); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Signal(ctx, "SIGHUP"); err == nil {
		t.Error("an unsupported signal returned no error")
	}

	bodies := a.bodies[len(a.bodies)-4:]
	if bodies[0]["data"] != enc([]byte("hello\n")) || bodies[0]["eof"] != true ||
		bodies[1]["eof"] != true || bodies[1]["data"] != nil ||
		bodies[2]["signal"] != "SANDBOX_COMMAND_SIGNAL_SIGTERM" || bodies[3]["signal"] != "SANDBOX_COMMAND_SIGNAL_SIGKILL" {
		t.Errorf("bodies = %+v", bodies)
	}
}
