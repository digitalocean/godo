package modal

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/digitalocean/godo/sandbox/compat/internal/fakeapi"
	"github.com/digitalocean/godo/sandbox/compat/internal/stubtest"
	"github.com/digitalocean/godo/sandbox/compat/unsupported"
)

func newTestClient(t *testing.T) (*fakeapi.API, *Client) {
	t.Helper()
	a := fakeapi.New(t)
	t.Setenv("DIGITALOCEAN_API_URL", a.URL)
	c, err := NewClientWithOptions(&ClientParams{TokenSecret: "do-token"})
	if err != nil {
		t.Fatal(err)
	}
	return a, c
}

func running(id string) string { return fakeapi.Sandbox(id, "SANDBOX_STATE_RUNNING", "coding-base") }

func wantNotImplemented(t *testing.T, err error, substr string) {
	t.Helper()
	if !errors.Is(err, unsupported.ErrNotImplemented) {
		t.Fatalf("error = %v, want a not-implemented error", err)
	}
	if !strings.Contains(err.Error(), substr) {
		t.Errorf("error %q does not mention %q", err, substr)
	}
}

func TestClientsNeverReadModalCredentials(t *testing.T) {
	t.Setenv("MODAL_TOKEN_ID", "ak-modal")
	t.Setenv("MODAL_TOKEN_SECRET", "as-modal-secret")
	t.Setenv("DIGITALOCEAN_ACCESS_TOKEN", "")
	if _, err := NewClient(); err == nil {
		t.Fatal("NewClient used MODAL_TOKEN_SECRET")
	}
	t.Setenv("DIGITALOCEAN_ACCESS_TOKEN", "do-token")
	if _, err := NewClient(); err != nil {
		t.Fatalf("NewClient with a DigitalOcean token: %v", err)
	}
}

func TestAppsAndSecretsAreClientSide(t *testing.T) {
	_, c := newTestClient(t)
	ctx := context.Background()
	app, err := c.Apps.FromName(ctx, "my-app", &AppFromNameParams{CreateIfMissing: true})
	if err != nil || app.Name != "my-app" || !strings.HasPrefix(app.AppID, "ap-") {
		t.Fatalf("FromName = %+v, %v", app, err)
	}
	again, _ := c.Apps.FromName(ctx, "my-app", nil)
	if again.AppID != app.AppID {
		t.Error("an App ID is not stable")
	}
	sec, err := c.Secrets.FromMap(ctx, map[string]string{"K": "v"}, nil)
	if err != nil || !strings.HasPrefix(sec.SecretID, "st-") {
		t.Fatalf("FromMap = %+v, %v", sec, err)
	}
	_, err = c.Secrets.FromName(ctx, "named", nil)
	wantNotImplemented(t, err, "FromName")
}

// counter hands out command PIDs 1, 2, 3, ...
func counter() (*int32, http.HandlerFunc) {
	var n int32
	return &n, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"pid":%d}`, atomic.AddInt32(&n, 1))
	}
}

func TestCreateBuildsTemplateReplaysLayersAndStartsTheEntrypoint(t *testing.T) {
	a, c := newTestClient(t)
	ctx := context.Background()
	_, start := counter()
	a.Handlers["GET /v2/sandboxes/templates"] = fakeapi.Reply(200, `{"templates":[],"next_page_token":""}`)
	a.Handlers["POST /v2/sandboxes/templates"] = fakeapi.Reply(201, `{"template":`+fakeapi.Template("t-1", "modal-x", "TEMPLATE_STATUS_READY")+`}`)
	a.Handlers["POST /v2/sandboxes"] = fakeapi.Reply(201, `{"sandbox":`+running("sb-1")+`}`)
	a.Handlers["POST /v2/sandboxes/sb-1/commands"] = start
	a.Handlers["GET /v2/sandboxes/sb-1/commands/1/stream"] = fakeapi.Reply(200, fakeapi.Stream(0, "hi\n"))
	a.Handlers["GET /v2/sandboxes/sb-1/commands/2/stream"] = fakeapi.Reply(200, fakeapi.Stream(0))
	a.Handlers["GET /v2/sandboxes/sb-1/commands/3/stream"] = fakeapi.Reply(200, fakeapi.Stream(7, "ready\n"))

	app, _ := c.Apps.FromName(ctx, "app", nil)
	img := c.Images.FromRegistry("python:3.12-slim", nil).DockerfileCommands([]string{"RUN echo hi", "ENV A=B", "WORKDIR /app"}, nil)
	secret, _ := c.Secrets.FromMap(ctx, map[string]string{"S": "1"}, nil)

	sb, err := c.Sandboxes.Create(ctx, app, img, &SandboxCreateParams{
		CPU: 1, MemoryMiB: 1500, Env: map[string]string{"E": "2"}, Secrets: []*Secret{secret},
		BlockNetwork: true, Command: []string{"cat"}, ExperimentalPersistent: true, Verbose: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if sb.SandboxID != "sb-1" || sb.Stdin == nil || sb.Stdout == nil {
		t.Fatalf("sandbox = %+v", sb)
	}
	tpl := a.Body("POST /v2/sandboxes/templates")
	if tpl["source_oci_ref"] != "python:3.12-slim" || !strings.HasPrefix(tpl["name"].(string), "modal-") {
		t.Errorf("template body = %+v", tpl)
	}
	body := a.Body("POST /v2/sandboxes")
	env, _ := body["env_vars"].(map[string]any)
	if body["template"] != tpl["name"] || body["size_slug"] != "mars-2vcpu-2gb" || body["allow_internet_access"] != false ||
		env["A"] != "B" || env["E"] != "2" || env["S"] != "1" {
		t.Errorf("sandbox body = %+v", body)
	}
	out, err := io.ReadAll(sb.Stdout)
	if err != nil || string(out) != "ready\n" {
		t.Errorf("entrypoint stdout = %q, %v", out, err)
	}
	code, err := sb.Wait(ctx, nil)
	if err != nil || code != 7 {
		t.Errorf("Wait = %d, %v", code, err)
	}
	if p, _ := sb.Poll(ctx, nil); p == nil || *p != 7 {
		t.Errorf("Poll = %v", p)
	}
	// The WORKDIR layer ran mkdir, then the entrypoint started in /app.
	calls := a.Calls()
	if got := strings.Count(strings.Join(calls, "\n"), "POST /v2/sandboxes/sb-1/commands"); got != 3 {
		t.Errorf("started %d commands, want 3 (RUN, WORKDIR, entrypoint)", got)
	}
	if got := a.Body("POST /v2/sandboxes/sb-1/commands")["cwd"]; got != "/app" {
		t.Errorf("entrypoint cwd = %v", got)
	}
}

func TestExecStreamsStdinAndOutput(t *testing.T) {
	a, c := newTestClient(t)
	ctx := context.Background()
	a.Handlers["GET /v2/sandboxes/sb-1"] = fakeapi.Reply(200, `{"sandbox":`+running("sb-1")+`}`)
	a.Handlers["POST /v2/sandboxes/sb-1/commands"] = fakeapi.Reply(200, `{"pid":42}`)
	a.Handlers["GET /v2/sandboxes/sb-1/commands/42/stream"] = fakeapi.Reply(200, fakeapi.Stream(0, "he", "llo\n"))
	a.Handlers["POST /v2/sandboxes/sb-1/commands/42/stdin"] = fakeapi.Reply(200, `{}`)

	sb, err := c.Sandboxes.FromID(ctx, "sb-1", nil)
	if err != nil {
		t.Fatal(err)
	}
	proc, err := sb.Exec(ctx, []string{"cat", "-n"}, &SandboxExecParams{
		Workdir: "/w", Env: map[string]string{"K": "v"}, Timeout: 2 * time.Second,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := proc.Stdin.Write([]byte("in")); err != nil {
		t.Fatal(err)
	}
	if err := proc.Stdin.Close(); err != nil {
		t.Fatal(err)
	}
	out, _ := io.ReadAll(proc.Stdout)
	if string(out) != "hello\n" {
		t.Errorf("stdout = %q", out)
	}
	if code, err := proc.Wait(ctx, nil); err != nil || code != 0 {
		t.Errorf("Wait = %d, %v", code, err)
	}
	body := a.Body("POST /v2/sandboxes/sb-1/commands")
	args, _ := body["args"].([]any)
	env, _ := body["env_vars"].(map[string]any)
	if body["cmd"] != "cat" || len(args) != 1 || args[0] != "-n" || body["cwd"] != "/w" || env["K"] != "v" ||
		body["stdin"] != true || body["timeout_seconds"] != "2" {
		t.Errorf("start body = %+v", body)
	}
	stdin := a.Body("POST /v2/sandboxes/sb-1/commands/42/stdin")
	if stdin["eof"] != true {
		t.Errorf("last stdin write = %+v, want the end-of-input marker", stdin)
	}
	if _, err := sb.Exec(ctx, []string{"sh"}, &SandboxExecParams{PTY: true}); err == nil {
		t.Error("PTY exec returned no error")
	} else {
		wantNotImplemented(t, err, "PTY")
	}
	_ = base64.StdEncoding
}

func TestIgnoredStdoutIsEmpty(t *testing.T) {
	a, c := newTestClient(t)
	ctx := context.Background()
	a.Handlers["GET /v2/sandboxes/sb-1"] = fakeapi.Reply(200, `{"sandbox":`+running("sb-1")+`}`)
	a.Handlers["POST /v2/sandboxes/sb-1/commands"] = fakeapi.Reply(200, `{"pid":42}`)
	a.Handlers["GET /v2/sandboxes/sb-1/commands/42/stream"] = fakeapi.Reply(200, fakeapi.Stream(0, "noise"))
	sb, _ := c.Sandboxes.FromID(ctx, "sb-1", nil)
	proc, err := sb.Exec(ctx, []string{"echo"}, &SandboxExecParams{Stdout: Ignore})
	if err != nil {
		t.Fatal(err)
	}
	out, _ := io.ReadAll(proc.Stdout)
	if len(out) != 0 {
		t.Errorf("an ignored stdout held %q", out)
	}
}

func TestCreateRejectsWhatDigitalOceanCannotDo(t *testing.T) {
	_, c := newTestClient(t)
	cases := map[string]SandboxCreateParams{
		"Timeout":               {Timeout: time.Minute},
		"IdleTimeout":           {IdleTimeout: time.Minute},
		"GPU":                   {GPU: "A100"},
		"Volumes":               {Volumes: map[string]*Volume{"/data": {}}},
		"PTY":                   {PTY: true},
		"EncryptedPorts":        {EncryptedPorts: []int{8080}},
		"Regions":               {Regions: []string{"us-east"}},
		"Name":                  {Name: "n"},
		"Tags":                  {Tags: map[string]string{"a": "b"}},
		"OutboundCIDRAllowlist": {OutboundCIDRAllowlist: &Allowlist{}},
		"ReadinessProbe":        {ReadinessProbe: &Probe{}},
		"CPULimit":              {CPULimit: 2},
	}
	for name, p := range cases {
		t.Run(name, func(t *testing.T) {
			p := p
			_, err := c.Sandboxes.Create(context.Background(), nil, nil, &p)
			wantNotImplemented(t, err, name)
		})
	}
	_, err := c.Sandboxes.Create(context.Background(), nil, nil, &SandboxCreateParams{CPU: 64})
	var inv InvalidError
	if !errors.As(err, &inv) {
		t.Errorf("an oversized CPU request returned %T %v", err, err)
	}
	policy := NewExperimentalOutboundPolicy(ExperimentalOutboundPolicyHeaderReplacement{Domain: "x.com"})
	_, err = c.Sandboxes.Create(context.Background(), nil, nil, &SandboxCreateParams{ExperimentalOutboundPolicy: &policy})
	wantNotImplemented(t, err, "ExperimentalOutboundPolicy")
}

func TestImagesWithUnsupportedFeaturesFailAtBuild(t *testing.T) {
	_, c := newTestClient(t)
	ctx := context.Background()
	secret, _ := c.Secrets.FromMap(ctx, map[string]string{"K": "v"}, nil)
	for name, img := range map[string]*Image{
		"Secret":    c.Images.FromRegistry("private/img", &ImageFromRegistryParams{Secret: secret}),
		"AwsEcr":    c.Images.FromAwsEcr("acct.dkr.ecr/img", secret, nil),
		"COPY":      c.Images.FromRegistry("alpine", nil).DockerfileCommands([]string{"COPY a b"}, nil),
		"Secrets":   c.Images.FromRegistry("alpine", nil).DockerfileCommands(nil, &ImageDockerfileCommandsParams{Secrets: []*Secret{secret}}),
		"GPU layer": c.Images.FromRegistry("alpine", nil).DockerfileCommands(nil, &ImageDockerfileCommandsParams{GPU: "T4"}),
	} {
		t.Run(name, func(t *testing.T) {
			_, err := img.Build(ctx, nil, nil)
			if err == nil {
				t.Fatal("Build returned no error")
			}
			wantNotImplemented(t, err, "")
		})
	}
	if _, err := c.Images.FromID(ctx, "im-missing", nil); err == nil {
		t.Error("FromID of an unknown image returned no error")
	} else if nf := (NotFoundError{}); !errors.As(err, &nf) {
		t.Errorf("FromID error = %T", err)
	}
}

func TestPollWaitTerminateWithoutAnEntrypoint(t *testing.T) {
	a, c := newTestClient(t)
	ctx := context.Background()
	a.Handlers["GET /v2/sandboxes/sb-1"] = fakeapi.Reply(200, `{"sandbox":`+running("sb-1")+`}`)
	a.Handlers["DELETE /v2/sandboxes/sb-1"] = fakeapi.Reply(200, `{}`)
	sb, _ := c.Sandboxes.FromID(ctx, "sb-1", nil)

	if p, err := sb.Poll(ctx, nil); err != nil || p != nil {
		t.Errorf("Poll of a running sandbox = %v, %v", p, err)
	}
	if code, err := sb.Terminate(ctx, &SandboxTerminateParams{Wait: true}); err != nil || code != 0 {
		t.Errorf("Terminate = %d, %v", code, err)
	}
	a.Handlers["GET /v2/sandboxes/sb-1"] = fakeapi.Reply(404, `{"id":"not_found","message":"gone"}`)
	if p, err := sb.Poll(ctx, nil); err != nil || p == nil || *p != 0 {
		t.Errorf("Poll of a deleted sandbox = %v, %v", p, err)
	}
	if code, err := sb.Wait(ctx, nil); err != nil || code != 0 {
		t.Errorf("Wait = %d, %v", code, err)
	}
	if err := sb.Detach(); err != nil {
		t.Error(err)
	}
}

func TestPauseResumeReadyAndSnapshots(t *testing.T) {
	a, c := newTestClient(t)
	ctx := context.Background()
	a.Handlers["GET /v2/sandboxes/sb-1"] = fakeapi.Reply(200, `{"sandbox":`+running("sb-1")+`}`)
	a.Handlers["POST /v2/sandboxes/sb-1/pause"] = fakeapi.Reply(200, `{"sandbox":`+fakeapi.Sandbox("sb-1", "SANDBOX_STATE_PAUSED", "coding-base")+`}`)
	a.Handlers["POST /v2/sandboxes/sb-1/resume"] = fakeapi.Reply(200, `{"sandbox":`+running("sb-1")+`}`)
	a.Handlers["POST /v2/sandboxes/sb-1/snapshots"] = fakeapi.Reply(202, `{"snapshot":`+fakeapi.Snapshot("snap-1", "SNAPSHOT_STATE_COMPLETE")+`}`)
	a.Handlers["GET /v2/sandboxes/snapshots/snap-1"] = fakeapi.Reply(200, `{"snapshot":`+fakeapi.Snapshot("snap-1", "SNAPSHOT_STATE_COMPLETE")+`}`)
	sb, _ := c.Sandboxes.FromID(ctx, "sb-1", nil)

	if err := sb.ExperimentalPause(ctx, nil); err != nil {
		t.Fatal(err)
	}
	if err := sb.ExperimentalResume(ctx, nil); err != nil {
		t.Fatal(err)
	}
	if err := sb.WaitUntilReady(ctx, time.Second, nil); err != nil {
		t.Fatal(err)
	}
	snap, err := sb.ExperimentalSnapshot(ctx, nil)
	if err != nil || snap.SnapshotID != "snap-1" {
		t.Fatalf("ExperimentalSnapshot = %+v, %v", snap, err)
	}
	got, err := c.SandboxSnapshots.FromID(ctx, "snap-1", nil)
	if err != nil || got.SnapshotID != "snap-1" {
		t.Errorf("FromID = %+v, %v", got, err)
	}
	_, err = c.Sandboxes.ExperimentalFromSnapshot(ctx, got, nil)
	wantNotImplemented(t, err, "ExperimentalFromSnapshot")
	_, err = sb.SnapshotFilesystem(ctx, nil)
	wantNotImplemented(t, err, "SnapshotFilesystem")
}

func TestListSandboxes(t *testing.T) {
	a, c := newTestClient(t)
	ctx := context.Background()
	a.Handlers["GET /v2/sandboxes"] = fakeapi.Reply(200, `{"sandboxes":[`+running("sb-1")+`,`+running("sb-2")+`],"next_page_token":""}`)
	seq, err := c.Sandboxes.List(ctx, &SandboxListParams{})
	if err != nil {
		t.Fatal(err)
	}
	var ids []string
	for sb, err := range seq {
		if err != nil {
			t.Fatal(err)
		}
		ids = append(ids, sb.SandboxID)
	}
	if strings.Join(ids, ",") != "sb-1,sb-2" {
		t.Errorf("ids = %v", ids)
	}
	_, err = c.Sandboxes.List(ctx, &SandboxListParams{Tags: map[string]string{"a": "b"}})
	wantNotImplemented(t, err, "Tags")
	_, err = c.Sandboxes.ExperimentalList(ctx, &SandboxExperimentalListParams{AppID: "ap-1"})
	wantNotImplemented(t, err, "AppID")
}

func TestErrorsBecomeModalErrors(t *testing.T) {
	a, c := newTestClient(t)
	a.Handlers["GET /v2/sandboxes/gone"] = fakeapi.Reply(404, `{"id":"not_found","message":"no such sandbox"}`)
	_, err := c.Sandboxes.FromID(context.Background(), "gone", nil)
	var nf NotFoundError
	if !errors.As(err, &nf) || !strings.Contains(nf.Exception, "no such sandbox") {
		t.Errorf("404 = %T %v", err, err)
	}
}

func TestValidateExecArgs(t *testing.T) {
	if err := ValidateExecArgs([]string{"echo", "hi"}); err != nil {
		t.Error(err)
	}
	if err := ValidateExecArgs([]string{strings.Repeat("x", argMax+1)}); err == nil {
		t.Error("oversized arguments returned no error")
	}
}

func TestUnimplementedMethodsReturnNotImplemented(t *testing.T) {
	sandboxDone := map[string]bool{
		"Exec": true, "Poll": true, "Wait": true, "Terminate": true, "Detach": true, "WaitUntilReady": true,
		"ExperimentalPause": true, "ExperimentalResume": true, "ExperimentalSnapshot": true,
	}
	total := 0
	total += stubtest.Check(t, "Modal", &Sandbox{}, sandboxDone)
	total += stubtest.Check(t, "Modal", &Image{}, map[string]bool{"Build": true, "DockerfileCommands": true})
	total += stubtest.Check(t, "Modal", &App{}, nil)
	// Builders that return their receiver: the calls that use the result fail.
	builders := map[string]bool{"WithBatching": true, "WithConcurrency": true, "WithOptions": true, "WithMountOptions": true}
	for _, recv := range []any{&Function{}, &Cls{}, &Volume{}} {
		total += stubtest.Check(t, "Modal", recv, builders)
	}
	for _, recv := range []any{
		&FunctionCall{}, &ClsInstance{}, &Queue{}, &SandboxFilesystem{}, &SidecarContainer{},
		&Environment{}, &EnvironmentService{}, &SandboxLogsManager{}, &AppSandboxesManager{},
		notImplementedFunctionService{}, notImplementedFunctionCallService{}, notImplementedClsService{}, notImplementedQueueService{},
		notImplementedVolumeService{}, notImplementedProxyService{}, notImplementedCloudBucketMountService{}, notImplementedSidecarService{},
		notImplementedEnvironmentObjectsService{}, notImplementedEnvironmentAppsService{},
	} {
		total += stubtest.Check(t, "Modal", recv, nil)
	}
	total += stubtest.Check(t, "Modal", sandboxService{}, map[string]bool{
		"Create": true, "ExperimentalCreate": true, "FromID": true, "List": true, "ExperimentalList": true,
	})
	total += stubtest.Check(t, "Modal", secretService{}, map[string]bool{"FromMap": true})
	total += stubtest.Check(t, "Modal", imageService{}, map[string]bool{"FromRegistry": true, "FromAwsEcr": true, "FromGcpArtifactRegistry": true, "FromID": true})
	if total < 60 {
		t.Errorf("checked only %d stub methods; the generated surface looks incomplete", total)
	}
}
