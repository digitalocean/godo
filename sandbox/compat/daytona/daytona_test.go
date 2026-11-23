package daytona

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/digitalocean/godo/sandbox/compat/daytona/apiclient"
	dterrors "github.com/digitalocean/godo/sandbox/compat/daytona/errors"
	"github.com/digitalocean/godo/sandbox/compat/daytona/options"
	"github.com/digitalocean/godo/sandbox/compat/daytona/types"
	"github.com/digitalocean/godo/sandbox/compat/internal/fakeapi"
	"github.com/digitalocean/godo/sandbox/compat/internal/stubtest"
	"github.com/digitalocean/godo/sandbox/compat/unsupported"
)

func newTestClient(t *testing.T) (*fakeapi.API, *Client) {
	t.Helper()
	a := fakeapi.New(t)
	c, err := NewClientWithConfig(&types.DaytonaConfig{APIKey: "do-token", APIUrl: a.URL})
	if err != nil {
		t.Fatal(err)
	}
	return a, c
}

func running(id string) string { return fakeapi.Sandbox(id, "SANDBOX_STATE_RUNNING", "coding-base") }
func paused(id string) string  { return fakeapi.Sandbox(id, "SANDBOX_STATE_PAUSED", "coding-base") }

func wantNotImplemented(t *testing.T, err error, substr string) {
	t.Helper()
	if !errors.Is(err, unsupported.ErrNotImplemented) {
		t.Fatalf("error = %v, want a not-implemented error", err)
	}
	if !strings.Contains(err.Error(), substr) {
		t.Errorf("error %q does not mention %q", err, substr)
	}
}

func TestClientsNeverReadDaytonaCredentials(t *testing.T) {
	t.Setenv("DAYTONA_API_KEY", "daytona-secret")
	t.Setenv("DIGITALOCEAN_ACCESS_TOKEN", "")
	if _, err := NewClient(); err == nil {
		t.Fatal("NewClient used DAYTONA_API_KEY")
	}
	t.Setenv("DIGITALOCEAN_ACCESS_TOKEN", "do-token")
	if _, err := NewClient(); err != nil {
		t.Fatalf("NewClient with a DigitalOcean token: %v", err)
	}
	_, err := NewClientWithConfig(&types.DaytonaConfig{JWTToken: "jwt"})
	wantNotImplemented(t, err, "JWTToken")
	_, err = NewClientWithConfig(&types.DaytonaConfig{APIKey: "k", Target: "us"})
	wantNotImplemented(t, err, "Target")
}

func TestCreateFromSnapshotParams(t *testing.T) {
	a, c := newTestClient(t)
	a.Handlers["POST /v2/sandboxes"] = fakeapi.Reply(201, `{"sandbox":`+fakeapi.Sandbox("sb-1", "SANDBOX_STATE_RUNNING", "my-tpl")+`}`)

	sb, err := c.Create(context.Background(), types.SnapshotParams{
		Snapshot: "my-tpl",
		SandboxBaseParams: types.SandboxBaseParams{
			EnvVars: map[string]string{"A": "1"}, NetworkBlockAll: true, Language: types.CodeLanguagePython,
		},
	}, options.WithTimeout(time.Minute), options.WithWaitForStart(false))
	if err != nil {
		t.Fatal(err)
	}
	if sb.ID != "sb-1" || sb.State != apiclient.SANDBOXSTATE_STARTED || sb.Snapshot == nil || *sb.Snapshot != "my-tpl" {
		t.Errorf("sandbox = %+v", sb)
	}
	if sb.Cpu != 2 || sb.Memory != 4 || sb.CreatedAt == nil {
		t.Errorf("size or time not mapped: cpu=%v mem=%v created=%v", sb.Cpu, sb.Memory, sb.CreatedAt)
	}
	body := a.Body("POST /v2/sandboxes")
	env, _ := body["env_vars"].(map[string]any)
	if body["template"] != "my-tpl" || body["allow_internet_access"] != false || env["A"] != "1" || body["size_slug"] != "mars-2vcpu-4gb" {
		t.Errorf("body = %+v", body)
	}
	if a.Auth() != "Bearer do-token" {
		t.Errorf("Authorization = %q", a.Auth())
	}

	if _, err := c.Create(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	if got := a.Body("POST /v2/sandboxes")["template"]; got != "coding-base" {
		t.Errorf("default template = %v", got)
	}
}

func TestCreateFromAnImageBuildsATemplateOnce(t *testing.T) {
	a, c := newTestClient(t)
	a.Handlers["GET /v2/sandboxes/templates"] = fakeapi.Reply(200, `{"templates":[],"next_page_token":""}`)
	a.Handlers["POST /v2/sandboxes/templates"] = fakeapi.Reply(201, `{"template":`+fakeapi.Template("t-1", "daytona-x", "TEMPLATE_STATUS_READY")+`}`)
	a.Handlers["POST /v2/sandboxes"] = fakeapi.Reply(201, `{"sandbox":`+running("sb-1")+`}`)

	logs := make(chan string)
	if _, err := c.Create(context.Background(), types.ImageParams{
		Image: "python:3.12", Resources: &types.Resources{CPU: 4, Memory: 8},
	}, options.WithLogChannel(logs)); err != nil {
		t.Fatal(err)
	}
	if _, open := <-logs; open {
		t.Error("log channel was not closed")
	}
	tpl := a.Body("POST /v2/sandboxes/templates")
	if tpl["source_oci_ref"] != "python:3.12" || !strings.HasPrefix(tpl["name"].(string), "daytona-") {
		t.Errorf("template body = %+v", tpl)
	}
	body := a.Body("POST /v2/sandboxes")
	if body["template"] != tpl["name"] || body["size_slug"] != "mars-4vcpu-8gb" {
		t.Errorf("sandbox body = %+v", body)
	}
}

func TestCreateRejectsWhatDigitalOceanCannotDo(t *testing.T) {
	_, c := newTestClient(t)
	zero, stop := 0, 30
	cases := map[string]struct {
		params any
		want   string
	}{
		"Name":                {types.SnapshotParams{SandboxBaseParams: types.SandboxBaseParams{Name: "n"}}, "Name"},
		"Labels":              {types.SnapshotParams{SandboxBaseParams: types.SandboxBaseParams{Labels: map[string]string{"a": "b"}}}, "Labels"},
		"AutoStopInterval":    {types.SnapshotParams{SandboxBaseParams: types.SandboxBaseParams{AutoStopInterval: &stop}}, "AutoStopInterval"},
		"AutoDeleteInterval":  {types.SnapshotParams{SandboxBaseParams: types.SandboxBaseParams{AutoDeleteInterval: &zero}}, "AutoDeleteInterval"},
		"Volumes":             {types.SnapshotParams{SandboxBaseParams: types.SandboxBaseParams{Volumes: []types.VolumeMount{{VolumeID: "v"}}}}, "Volumes"},
		"Ephemeral":           {types.SnapshotParams{SandboxBaseParams: types.SandboxBaseParams{Ephemeral: true}}, "Ephemeral"},
		"GPU":                 {types.ImageParams{Image: "x", Resources: &types.Resources{GPU: 1}}, "GPU"},
		"Disk":                {types.ImageParams{Image: "x", Resources: &types.Resources{Disk: 10}}, "Disk"},
		"DockerImage builder": {types.ImageParams{Image: Base("python:3.12").PipInstall([]string{"numpy"})}, "Image"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := c.Create(context.Background(), tc.params)
			wantNotImplemented(t, err, tc.want)
		})
	}
	if _, err := c.Create(context.Background(), 42); err == nil {
		t.Error("an unsupported params type returned no error")
	}
}

func TestGetStartStopDeleteAndRefresh(t *testing.T) {
	a, c := newTestClient(t)
	ctx := context.Background()
	a.Handlers["GET /v2/sandboxes/sb-1"] = fakeapi.Reply(200, `{"sandbox":`+running("sb-1")+`}`)
	a.Handlers["POST /v2/sandboxes/sb-1/pause"] = fakeapi.Reply(200, `{"sandbox":`+paused("sb-1")+`}`)
	a.Handlers["POST /v2/sandboxes/sb-1/resume"] = fakeapi.Reply(200, `{"sandbox":`+running("sb-1")+`}`)
	a.Handlers["DELETE /v2/sandboxes/sb-1"] = fakeapi.Reply(200, `{}`)

	sb, err := c.Get(ctx, "sb-1")
	if err != nil || sb.State != apiclient.SANDBOXSTATE_STARTED {
		t.Fatalf("Get = %+v, %v", sb, err)
	}
	if err := sb.Stop(ctx); err != nil || sb.State != apiclient.SANDBOXSTATE_STOPPED {
		t.Fatalf("Stop: state %v, err %v", sb.State, err)
	}
	if err := sb.Start(ctx); err != nil || sb.State != apiclient.SANDBOXSTATE_STARTED {
		t.Fatalf("Start: state %v, err %v", sb.State, err)
	}
	if err := sb.Pause(ctx); err != nil {
		t.Fatal(err)
	}
	a.Handlers["GET /v2/sandboxes/sb-1"] = fakeapi.Reply(200, `{"sandbox":`+paused("sb-1")+`}`)
	if err := sb.WaitForStop(ctx, 5*time.Second); err != nil {
		t.Errorf("WaitForStop: %v", err)
	}
	a.Handlers["GET /v2/sandboxes/sb-1"] = fakeapi.Reply(200, `{"sandbox":`+running("sb-1")+`}`)
	if err := sb.RefreshData(ctx); err != nil || sb.State != apiclient.SANDBOXSTATE_STARTED {
		t.Errorf("RefreshData: state %v, err %v", sb.State, err)
	}
	if err := sb.WaitForStart(ctx, time.Second); err != nil {
		t.Errorf("WaitForStart: %v", err)
	}
	if err := sb.Delete(ctx); err != nil || sb.State != apiclient.SANDBOXSTATE_DESTROYED {
		t.Errorf("Delete: state %v, err %v", sb.State, err)
	}
}

func TestErrorsBecomeDaytonaErrors(t *testing.T) {
	a, c := newTestClient(t)
	a.Handlers["GET /v2/sandboxes/gone"] = fakeapi.Reply(http.StatusNotFound, `{"id":"not_found","message":"no such sandbox"}`)
	a.Handlers["GET /v2/sandboxes/busy"] = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "0")
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"id":"forbidden","message":"no scope"}`))
	})

	_, err := c.Get(context.Background(), "gone")
	var nf *dterrors.DaytonaNotFoundError
	if !errors.As(err, &nf) || !strings.Contains(err.Error(), "no such sandbox") {
		t.Errorf("404 = %T %v", err, err)
	}
	_, err = c.Get(context.Background(), "busy")
	var fb *dterrors.DaytonaForbiddenError
	if !errors.As(err, &fb) {
		t.Errorf("403 = %T %v", err, err)
	}
}

func TestExecuteCommandRunsBashAndCombinesOutput(t *testing.T) {
	a, c := newTestClient(t)
	ctx := context.Background()
	a.Handlers["GET /v2/sandboxes/sb-1"] = fakeapi.Reply(200, `{"sandbox":`+running("sb-1")+`}`)
	a.Handlers["POST /v2/sandboxes/sb-1/commands"] = fakeapi.Reply(200, `{"pid":42}`)
	a.Handlers["GET /v2/sandboxes/sb-1/commands/42/stream"] = fakeapi.Reply(200, fakeapi.Stream(3, "hel", "lo\n"))

	sb, err := c.Get(ctx, "sb-1")
	if err != nil {
		t.Fatal(err)
	}
	res, err := sb.Process.ExecuteCommand(ctx, "echo hello", options.WithCwd("/w"),
		options.WithCommandEnv(map[string]string{"K": "v"}), options.WithExecuteTimeout(1500*time.Millisecond))
	if err != nil {
		t.Fatal(err)
	}
	if res.ExitCode != 3 || res.Result != "hello\n" || res.Artifacts != nil {
		t.Errorf("response = %+v", res)
	}
	body := a.Body("POST /v2/sandboxes/sb-1/commands")
	args, _ := body["args"].([]any)
	env, _ := body["env_vars"].(map[string]any)
	if body["cmd"] != "bash" || len(args) != 2 || args[1] != "echo hello" || body["cwd"] != "/w" || env["K"] != "v" || body["timeout_seconds"] != "2" {
		t.Errorf("start body = %+v", body)
	}
}

func TestWorkingAndHomeDirRunCommands(t *testing.T) {
	a, c := newTestClient(t)
	ctx := context.Background()
	a.Handlers["GET /v2/sandboxes/sb-1"] = fakeapi.Reply(200, `{"sandbox":`+running("sb-1")+`}`)
	a.Handlers["POST /v2/sandboxes/sb-1/commands"] = fakeapi.Reply(200, `{"pid":42}`)
	a.Handlers["GET /v2/sandboxes/sb-1/commands/42/stream"] = fakeapi.Reply(200, fakeapi.Stream(0, "/workspace\n"))
	sb, _ := c.Get(ctx, "sb-1")
	if dir, err := sb.GetWorkingDir(ctx); err != nil || dir != "/workspace" {
		t.Errorf("GetWorkingDir = %q, %v", dir, err)
	}
}

func TestListFiltersAndIterates(t *testing.T) {
	a, c := newTestClient(t)
	ctx := context.Background()
	a.Handlers["GET /v2/sandboxes"] = fakeapi.Reply(200, `{"sandboxes":[`+running("sb-1")+`,`+paused("sb-2")+`],"next_page_token":""}`)

	it := c.List(ctx, nil)
	var ids []string
	for it.Next() {
		ids = append(ids, it.Value().ID)
	}
	if it.Err() != nil || strings.Join(ids, ",") != "sb-1,sb-2" {
		t.Errorf("List = %v, %v", ids, it.Err())
	}

	stopped := apiclient.SANDBOXSTATE_STOPPED
	n := 0
	for sb, err := range c.ListSeq(ctx, &ListSandboxesQuery{States: []apiclient.SandboxState{stopped, apiclient.SANDBOXSTATE_STARTED}, Snapshots: []string{"coding-base"}}) {
		if err != nil {
			t.Fatal(err)
		}
		n++
		_ = sb
	}
	if n != 2 {
		t.Errorf("ListSeq with two states yielded %d, want 2", n)
	}

	bad := c.List(ctx, &ListSandboxesQuery{Labels: map[string]string{"a": "b"}})
	if bad.Next() {
		t.Error("a label filter returned a sandbox")
	}
	wantNotImplemented(t, bad.Err(), "Labels")
}

func TestSnapshotServiceMapsToTemplates(t *testing.T) {
	pollEvery = 5 * time.Millisecond
	t.Cleanup(func() { pollEvery = 2 * time.Second })
	a, c := newTestClient(t)
	ctx := context.Background()
	a.Handlers["POST /v2/sandboxes/templates"] = fakeapi.Reply(201, `{"template":`+fakeapi.Template("t-1", "my-env", "TEMPLATE_STATUS_PENDING")+`}`)
	a.Handlers["GET /v2/sandboxes/templates/t-1"] = fakeapi.Reply(200, `{"template":`+fakeapi.Template("t-1", "my-env", "TEMPLATE_STATUS_READY")+`}`)
	a.Handlers["GET /v2/sandboxes/templates"] = fakeapi.Reply(200, `{"templates":[`+fakeapi.Template("t-1", "my-env", "TEMPLATE_STATUS_READY")+`],"next_page_token":""}`)
	a.Handlers["DELETE /v2/sandboxes/templates/t-1"] = fakeapi.Reply(200, `{}`)

	snap, logs, err := c.Snapshot.Create(ctx, &types.CreateSnapshotParams{Name: "my-env", Image: "python:3.12"})
	if err != nil {
		t.Fatal(err)
	}
	if snap.ID != "t-1" || snap.State != "pending" || snap.ImageName != "docker.io/library/python:3.12" {
		t.Errorf("snapshot = %+v", snap)
	}
	var lines []string
	for l := range logs {
		lines = append(lines, l)
	}
	if len(lines) == 0 || !strings.Contains(lines[len(lines)-1], "READY") {
		t.Errorf("log lines = %q", lines)
	}
	if body := a.Body("POST /v2/sandboxes/templates"); body["name"] != "my-env" || body["source_oci_ref"] != "python:3.12" {
		t.Errorf("template body = %+v", body)
	}

	got, err := c.Snapshot.Get(ctx, "my-env")
	if err != nil || got.State != "active" {
		t.Errorf("Get = %+v, %v", got, err)
	}
	page, err := c.Snapshot.List(ctx, nil, nil)
	if err != nil || page.Total != 1 || page.Page != 1 || page.TotalPages != 1 || len(page.Items) != 1 {
		t.Errorf("List = %+v, %v", page, err)
	}
	if err := c.Snapshot.Delete(ctx, got); err != nil {
		t.Errorf("Delete: %v", err)
	}
	if _, err := c.Snapshot.Get(ctx, "nope"); err == nil {
		t.Error("Get of an unknown snapshot returned no error")
	} else if nf := new(dterrors.DaytonaNotFoundError); !errors.As(err, &nf) {
		t.Errorf("Get error = %T", err)
	}

	for name, p := range map[string]*types.CreateSnapshotParams{
		"Resources":  {Name: "n", Image: "i", Resources: &types.Resources{CPU: 1}},
		"Entrypoint": {Name: "n", Image: "i", Entrypoint: []string{"x"}},
		"Image":      {Name: "n", Image: Base("python")},
	} {
		_, _, err := c.Snapshot.Create(ctx, p)
		wantNotImplemented(t, err, name)
	}
}

func TestExperimentalCreateSnapshotCapturesTheSandbox(t *testing.T) {
	a, c := newTestClient(t)
	ctx := context.Background()
	a.Handlers["GET /v2/sandboxes/sb-1"] = fakeapi.Reply(200, `{"sandbox":`+running("sb-1")+`}`)
	a.Handlers["POST /v2/sandboxes/sb-1/snapshots"] = fakeapi.Reply(202, `{"snapshot":`+fakeapi.Snapshot("snap-1", "SNAPSHOT_STATE_COMPLETE")+`}`)
	sb, _ := c.Get(ctx, "sb-1")
	if err := sb.ExperimentalCreateSnapshot(ctx, "snap"); err != nil {
		t.Fatal(err)
	}
	if got := a.Body("POST /v2/sandboxes/sb-1/snapshots")["name"]; got != "snap" {
		t.Errorf("snapshot name = %v", got)
	}
	if err := sb.ExperimentalCreateSnapshot(ctx, "Bad Name"); err == nil || !strings.Contains(err.Error(), "Bad Name") {
		t.Errorf("an invalid snapshot name returned %v", err)
	}
}

func TestUnimplementedMethodsReturnNotImplemented(t *testing.T) {
	implemented := map[string]bool{
		"Start": true, "StartWithTimeout": true, "Stop": true, "StopWithTimeout": true, "Pause": true, "PauseWithTimeout": true,
		"Delete": true, "DeleteWithTimeout": true, "RefreshData": true, "WaitForStart": true, "WaitForStop": true,
		"ExperimentalCreateSnapshot": true, "ExperimentalCreateSnapshotWithTimeout": true, "GetUserHomeDir": true, "GetWorkingDir": true,
		"CreateLspServer": true, // returns a service whose methods are checked below
	}
	total := 0
	total += stubtest.Check(t, "Daytona", &Sandbox{}, implemented)
	total += stubtest.Check(t, "Daytona", &ProcessService{}, map[string]bool{"ExecuteCommand": true})
	for _, recv := range []any{
		&FileSystemService{}, &GitService{}, &CodeInterpreterService{}, &ComputerUseService{}, &MouseService{}, &KeyboardService{},
		&ScreenshotService{}, &DisplayService{}, &RecordingService{}, &AccessibilityService{}, &LspServerService{}, &VolumeService{},
	} {
		total += stubtest.Check(t, "Daytona", recv, nil)
	}
	if total < 100 {
		t.Errorf("checked only %d stub methods; the generated surface looks incomplete", total)
	}
	var sb Sandbox
	if sb.CreateLspServer(types.LspLanguagePython, "/w") == nil {
		t.Error("CreateLspServer returned nil")
	}
}
