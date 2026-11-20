package sandbox_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/digitalocean/godo"
	"github.com/digitalocean/godo/sandbox"
)

func raw(id string, state godo.SandboxState) string {
	return fmt.Sprintf(`{"sandbox_id":%q,"template":"coding-base","state":%q,"size_slug":"mars-2vcpu-4gb","created_at":"2026-11-20T12:00:00Z"}`, id, state)
}

// api serves handlers by "METHOD /path" and records each request.
type api struct {
	t        *testing.T
	handlers map[string]http.HandlerFunc
	calls    []string
	bodies   []map[string]interface{}
	auth     string
}

func newAPI(t *testing.T) (*api, *sandbox.Client) {
	t.Helper()
	a := &api{t: t, handlers: map[string]http.HandlerFunc{}}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Method + " " + r.URL.Path
		a.calls = append(a.calls, key)
		a.auth = r.Header.Get("Authorization")
		var body map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&body)
		a.bodies = append(a.bodies, body)
		h, ok := a.handlers[key]
		if !ok {
			t.Errorf("unexpected request %s", key)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		h(w, r)
	}))
	t.Cleanup(srv.Close)
	c, err := sandbox.NewClient(&sandbox.Config{Token: "test-token", APIURL: srv.URL})
	if err != nil {
		t.Fatal(err)
	}
	return a, c
}

func reply(status int, body string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		fmt.Fprint(w, body)
	}
}

func TestCreateDefaultsAndOptions(t *testing.T) {
	a, c := newAPI(t)
	a.handlers["POST /v2/sandboxes"] = reply(201, `{"sandbox":`+raw("sb-1", godo.SandboxStateRunning)+`}`)
	ctx := context.Background()

	sb, err := c.Create(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if sb.ID() != "sb-1" || sb.State() != godo.SandboxStateRunning || sb.Template() != "coding-base" {
		t.Errorf("unexpected sandbox %+v", sb.Info)
	}
	if a.bodies[0]["template"] != "coding-base" || a.bodies[0]["size_slug"] != "mars-2vcpu-4gb" {
		t.Errorf("defaults not applied: %+v", a.bodies[0])
	}
	if a.auth != "Bearer test-token" {
		t.Errorf("Authorization = %q", a.auth)
	}

	off := false
	if _, err := c.Create(ctx, &sandbox.CreateParams{
		Template: "other", Size: "s-1vcpu-2gb", AllowInternetAccess: &off, EnvVars: map[string]string{"A": "1"},
	}); err != nil {
		t.Fatal(err)
	}
	got := a.bodies[1]
	if got["template"] != "other" || got["size_slug"] != "s-1vcpu-2gb" || got["allow_internet_access"] != false {
		t.Errorf("options not sent: %+v", got)
	}
}

func TestCreateRejectsBlankTemplate(t *testing.T) {
	_, c := newAPI(t)
	if _, err := c.Create(context.Background(), &sandbox.CreateParams{Template: "  "}); err == nil {
		t.Error("blank template returned no error")
	}
}

func TestPauseResumeKill(t *testing.T) {
	a, c := newAPI(t)
	a.handlers["GET /v2/sandboxes/sb-1"] = reply(200, `{"sandbox":`+raw("sb-1", godo.SandboxStateRunning)+`}`)
	a.handlers["POST /v2/sandboxes/sb-1/pause"] = reply(200, `{"sandbox":`+raw("sb-1", godo.SandboxStatePaused)+`}`)
	a.handlers["POST /v2/sandboxes/sb-1/resume"] = reply(200, `{"sandbox":`+raw("sb-1", godo.SandboxStateRunning)+`}`)
	a.handlers["DELETE /v2/sandboxes/sb-1"] = reply(200, `{}`)
	ctx := context.Background()

	sb, err := c.Get(ctx, "sb-1")
	if err != nil {
		t.Fatal(err)
	}
	if err := sb.Pause(ctx); err != nil || sb.State() != godo.SandboxStatePaused {
		t.Fatalf("Pause: state %s, err %v", sb.State(), err)
	}
	if err := sb.Resume(ctx); err != nil || sb.State() != godo.SandboxStateRunning {
		t.Fatalf("Resume: state %s, err %v", sb.State(), err)
	}
	if running, err := sb.IsRunning(ctx); err != nil || !running {
		t.Fatalf("IsRunning = %v, %v", running, err)
	}
	if deleted, err := sb.Kill(ctx); err != nil || !deleted {
		t.Fatalf("Kill = %v, %v", deleted, err)
	}
}

func TestKillReportsFalseWhenGone(t *testing.T) {
	a, c := newAPI(t)
	a.handlers["DELETE /v2/sandboxes/gone"] = reply(404, `{"id":"not_found","message":"gone"}`)
	deleted, err := c.Kill(context.Background(), "gone")
	if err != nil || deleted {
		t.Errorf("Kill = %v, %v; expected false, nil", deleted, err)
	}
}

func TestPauseByID(t *testing.T) {
	a, c := newAPI(t)
	a.handlers["POST /v2/sandboxes/sb-2/pause"] = reply(200, `{"sandbox":`+raw("sb-2", godo.SandboxStatePaused)+`}`)
	sb, err := c.Pause(context.Background(), "sb-2")
	if err != nil || sb.State() != godo.SandboxStatePaused {
		t.Errorf("Pause = %+v, %v", sb, err)
	}
}

func TestConnectResumesAPausedSandbox(t *testing.T) {
	a, c := newAPI(t)
	a.handlers["GET /v2/sandboxes/sb-1"] = reply(200, `{"sandbox":`+raw("sb-1", godo.SandboxStatePaused)+`}`)
	a.handlers["POST /v2/sandboxes/sb-1/resume"] = reply(200, `{"sandbox":`+raw("sb-1", godo.SandboxStateRunning)+`}`)
	sb, err := c.Connect(context.Background(), "sb-1")
	if err != nil || sb.State() != godo.SandboxStateRunning {
		t.Fatalf("Connect = %+v, %v", sb, err)
	}
}

func TestListPagesAndAll(t *testing.T) {
	a, c := newAPI(t)
	var queries []string
	a.handlers["GET /v2/sandboxes"] = func(w http.ResponseWriter, r *http.Request) {
		queries = append(queries, r.URL.RawQuery)
		if r.URL.Query().Get("page_token") == "" {
			fmt.Fprint(w, `{"sandboxes":[`+raw("a", godo.SandboxStateRunning)+`],"next_page_token":"t2"}`)
			return
		}
		fmt.Fprint(w, `{"sandboxes":[`+raw("b", godo.SandboxStateRunning)+`]}`)
	}
	pages := c.List(&sandbox.ListParams{State: godo.SandboxStateRunning, Limit: 1})
	var ids []string
	for sb, err := range pages.All(context.Background()) {
		if err != nil {
			t.Fatal(err)
		}
		ids = append(ids, sb.ID())
	}
	if strings.Join(ids, ",") != "a,b" {
		t.Errorf("ids = %v", ids)
	}
	if pages.HasNext() {
		t.Error("HasNext is true after the last page")
	}
	if !strings.Contains(queries[0], "page_size=1") || !strings.Contains(queries[0], "state=SANDBOX_STATE_RUNNING") ||
		!strings.Contains(queries[1], "page_token=t2") {
		t.Errorf("queries = %v", queries)
	}
}

func TestListOnePageAtATime(t *testing.T) {
	a, c := newAPI(t)
	a.handlers["GET /v2/sandboxes"] = reply(200, `{"sandboxes":[`+raw("a", godo.SandboxStateRunning)+`],"next_page_token":"t2"}`)
	pages := c.List(nil)
	if !pages.HasNext() {
		t.Fatal("HasNext is false before reading")
	}
	items, err := pages.NextItems(context.Background())
	if err != nil || len(items) != 1 || pages.NextToken() != "t2" {
		t.Errorf("NextItems = %+v, %v, token %q", items, err, pages.NextToken())
	}
}

func TestErrorHelpers(t *testing.T) {
	a, c := newAPI(t)
	ctx := context.Background()
	for _, tc := range []struct {
		status int
		is     func(error) bool
	}{
		{404, sandbox.IsNotFound}, {401, sandbox.IsUnauthorized}, {403, sandbox.IsPermissionDenied},
		{400, sandbox.IsInvalidRequest}, {422, sandbox.IsInvalidRequest}, {409, sandbox.IsConflict},
	} {
		a.handlers["GET /v2/sandboxes/x"] = reply(tc.status, `{"id":"e","message":"m","request_id":"r"}`)
		_, err := c.Get(ctx, "x")
		if err == nil || !tc.is(err) {
			t.Errorf("status %d: helper did not match %v", tc.status, err)
		}
		if tc.status != 404 && sandbox.IsNotFound(err) {
			t.Errorf("status %d matched IsNotFound", tc.status)
		}
	}
}

func TestWaitUntilRunning(t *testing.T) {
	a, c := newAPI(t)
	n := 0
	a.handlers["GET /v2/sandboxes/sb-1"] = func(w http.ResponseWriter, r *http.Request) {
		n++
		state := godo.SandboxStateCreating
		if n > 1 {
			state = godo.SandboxStateRunning
		}
		fmt.Fprint(w, `{"sandbox":`+raw("sb-1", state)+`}`)
	}
	ctx := context.Background()
	sb, err := c.Get(ctx, "sb-1")
	if err != nil {
		t.Fatal(err)
	}
	if err := sb.WaitUntilRunning(ctx, &sandbox.WaitOptions{PollInterval: time.Millisecond}); err != nil {
		t.Fatal(err)
	}
	if sb.State() != godo.SandboxStateRunning {
		t.Errorf("state = %s", sb.State())
	}
}

func TestWaitUntilRunningFailsWhenPausedOrTimesOut(t *testing.T) {
	a, c := newAPI(t)
	a.handlers["GET /v2/sandboxes/p"] = reply(200, `{"sandbox":`+raw("p", godo.SandboxStatePaused)+`}`)
	a.handlers["GET /v2/sandboxes/c"] = reply(200, `{"sandbox":`+raw("c", godo.SandboxStateCreating)+`}`)
	ctx := context.Background()

	paused, _ := c.Get(ctx, "p")
	if err := paused.WaitUntilRunning(ctx, nil); err == nil {
		t.Error("a paused sandbox did not fail the wait")
	}
	creating, _ := c.Get(ctx, "c")
	err := creating.WaitUntilRunning(ctx, &sandbox.WaitOptions{Timeout: 20 * time.Millisecond, PollInterval: time.Millisecond})
	if err == nil || !strings.Contains(err.Error(), "not running in time") {
		t.Errorf("err = %v", err)
	}
}

func TestNewClientTokenAndURLFromEnvironment(t *testing.T) {
	t.Setenv("DIGITALOCEAN_ACCESS_TOKEN", "")
	if _, err := sandbox.NewClient(nil); err != sandbox.ErrNoToken {
		t.Errorf("err = %v, expected ErrNoToken", err)
	}

	var auth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth = r.Header.Get("Authorization")
		fmt.Fprint(w, `{"sandbox":`+raw("sb-1", godo.SandboxStateRunning)+`}`)
	}))
	defer srv.Close()
	t.Setenv("DIGITALOCEAN_ACCESS_TOKEN", "env-token")
	t.Setenv("DIGITALOCEAN_API_URL", srv.URL)
	c, err := sandbox.NewClient(nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.Get(context.Background(), "sb-1"); err != nil {
		t.Fatal(err)
	}
	if auth != "Bearer env-token" {
		t.Errorf("Authorization = %q", auth)
	}
}
