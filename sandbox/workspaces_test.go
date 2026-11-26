package sandbox_test

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/digitalocean/godo"
	"github.com/digitalocean/godo/sandbox"
)

func wsJSON(id string, state godo.SandboxWorkspaceState) string {
	return fmt.Sprintf(`{"workspace_id":%q,"name":"my-workspace","size_limit_gib":10,"state":%q,"bytes_used":"42","created_at":"2026-11-20T12:00:00Z"}`, id, state)
}

func TestCreateWorkspaceSendsSizeAndName(t *testing.T) {
	a, c := newAPI(t)
	a.handlers["POST /v2/sandboxes/workspaces"] = reply(201, `{"workspace":`+wsJSON("ws-1", godo.SandboxWorkspaceStateCreating)+`}`)

	ws, err := c.CreateWorkspace(context.Background(), &sandbox.WorkspaceParams{Name: "my-workspace", SizeLimitGiB: 10})
	if err != nil {
		t.Fatal(err)
	}
	if ws.ID() != "ws-1" || ws.Name() != "my-workspace" || ws.State() != godo.SandboxWorkspaceStateCreating || ws.Info.Used() != 42 {
		t.Errorf("workspace = %+v", ws.Info)
	}
	if body := a.bodies[0]; body["name"] != "my-workspace" || body["size_limit_gib"] != float64(10) {
		t.Errorf("body = %+v", body)
	}

	for _, p := range []*sandbox.WorkspaceParams{nil, {}, {SizeLimitGiB: -1}, {SizeLimitGiB: 101}} {
		if _, err := c.CreateWorkspace(context.Background(), p); err == nil {
			t.Errorf("%+v returned no error", p)
		}
	}
}

func TestGetRefreshAndDeleteWorkspace(t *testing.T) {
	a, c := newAPI(t)
	ctx := context.Background()
	state := godo.SandboxWorkspaceStateCreating
	a.handlers["GET /v2/sandboxes/workspaces/ws-1"] = func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"workspace":`+wsJSON("ws-1", state)+`}`)
	}
	a.handlers["DELETE /v2/sandboxes/workspaces/ws-1"] = reply(200, `{}`)
	a.handlers["DELETE /v2/sandboxes/workspaces/gone"] = reply(404, `{"id":"not_found","message":"no such workspace"}`)

	ws, err := c.GetWorkspace(ctx, "ws-1")
	if err != nil || ws.State() != godo.SandboxWorkspaceStateCreating {
		t.Fatalf("GetWorkspace = %+v, %v", ws, err)
	}
	state = godo.SandboxWorkspaceStateAvailable
	if err := ws.Refresh(ctx); err != nil || ws.State() != godo.SandboxWorkspaceStateAvailable {
		t.Errorf("Refresh = %v, state %s", err, ws.State())
	}
	if ok, err := ws.Delete(ctx); err != nil || !ok {
		t.Errorf("Delete = %v, %v", ok, err)
	}
	if ok, err := c.DeleteWorkspace(ctx, "gone"); err != nil || ok {
		t.Errorf("DeleteWorkspace of a missing workspace = %v, %v; expected false, nil", ok, err)
	}
}

func TestListWorkspacesSendsFiltersAndPages(t *testing.T) {
	a, c := newAPI(t)
	ctx := context.Background()
	var queries []string
	a.handlers["GET /v2/sandboxes/workspaces"] = func(w http.ResponseWriter, r *http.Request) {
		queries = append(queries, r.URL.RawQuery)
		next := ""
		if r.URL.Query().Get("page_token") == "" {
			next = "p2"
		}
		fmt.Fprintf(w, `{"workspaces":[%s],"next_page_token":%q}`, wsJSON("ws-"+fmt.Sprint(len(queries)), godo.SandboxWorkspaceStateAvailable), next)
	}

	var ids []string
	pages := c.ListWorkspaces(&sandbox.ListWorkspacesParams{State: godo.SandboxWorkspaceStateAvailable, Limit: 1})
	for ws, err := range pages.All(ctx) {
		if err != nil {
			t.Fatal(err)
		}
		ids = append(ids, ws.ID())
	}
	if strings.Join(ids, ",") != "ws-1,ws-2" {
		t.Errorf("ids = %v", ids)
	}
	if len(queries) != 2 || !strings.Contains(queries[0], "state=WORKSPACE_STATE_AVAILABLE") || !strings.Contains(queries[0], "page_size=1") ||
		!strings.Contains(queries[1], "page_token=p2") {
		t.Errorf("queries = %v", queries)
	}
}

func TestCreateAttachesAWorkspace(t *testing.T) {
	a, c := newAPI(t)
	a.handlers["POST /v2/sandboxes"] = reply(200, `{"sandbox":`+raw("sb-1", godo.SandboxStateRunning)+`}`)

	if _, err := c.Create(context.Background(), &sandbox.CreateParams{WorkspaceID: "ws-1"}); err != nil {
		t.Fatal(err)
	}
	if a.bodies[0]["workspace_id"] != "ws-1" {
		t.Errorf("body = %+v", a.bodies[0])
	}
}

func TestWaitUntilAvailablePollsACreatingWorkspace(t *testing.T) {
	a, c := newAPI(t)
	n := 0
	a.handlers["GET /v2/sandboxes/workspaces/ws-1"] = func(w http.ResponseWriter, r *http.Request) {
		n++
		state := godo.SandboxWorkspaceStateCreating
		if n >= 3 {
			state = godo.SandboxWorkspaceStateAvailable
		}
		fmt.Fprint(w, `{"workspace":`+wsJSON("ws-1", state)+`}`)
	}
	ctx := context.Background()
	ws, err := c.GetWorkspace(ctx, "ws-1")
	if err != nil {
		t.Fatal(err)
	}
	if err := ws.WaitUntilAvailable(ctx, &sandbox.WaitOptions{PollInterval: time.Millisecond}); err != nil {
		t.Fatal(err)
	}
	if ws.State() != godo.SandboxWorkspaceStateAvailable || n != 3 {
		t.Errorf("state %s after %d reads", ws.State(), n)
	}
}

func TestWaitUntilAvailableFailsOnAFailedWorkspace(t *testing.T) {
	a, c := newAPI(t)
	a.handlers["GET /v2/sandboxes/workspaces/ws-1"] = reply(200, `{"workspace":`+wsJSON("ws-1", godo.SandboxWorkspaceStateFailed)+`}`)
	ctx := context.Background()
	ws, _ := c.GetWorkspace(ctx, "ws-1")
	err := ws.WaitUntilAvailable(ctx, nil)
	if err == nil || !strings.Contains(err.Error(), "will not become available") {
		t.Errorf("err = %v", err)
	}
}

func TestWaitUntilAvailableTimesOut(t *testing.T) {
	a, c := newAPI(t)
	a.handlers["GET /v2/sandboxes/workspaces/ws-1"] = reply(200, `{"workspace":`+wsJSON("ws-1", godo.SandboxWorkspaceStateCreating)+`}`)
	ctx := context.Background()
	ws, _ := c.GetWorkspace(ctx, "ws-1")
	err := ws.WaitUntilAvailable(ctx, &sandbox.WaitOptions{Timeout: 30 * time.Millisecond, PollInterval: time.Millisecond})
	if err == nil || !strings.Contains(err.Error(), "not available in time") {
		t.Errorf("err = %v", err)
	}
}
