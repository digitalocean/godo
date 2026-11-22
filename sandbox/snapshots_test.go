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

func snapJSON(id string, state godo.SandboxSnapshotState, errMsg string) string {
	return fmt.Sprintf(`{"snapshot_id":%q,"sandbox_id":"sb-1","name":"my-snapshot","state":%q,"error_message":%q,"created_at":"2026-11-20T12:00:00Z"}`, id, state, errMsg)
}

func TestCreateSnapshotSendsOptionsAndValidatesThem(t *testing.T) {
	a, c := newAPI(t)
	ctx := context.Background()
	a.handlers["GET /v2/sandboxes/sb-1"] = reply(200, `{"sandbox":`+raw("sb-1", godo.SandboxStateRunning)+`}`)
	a.handlers["POST /v2/sandboxes/sb-1/snapshots"] = reply(202, `{"snapshot":`+snapJSON("snap-1", godo.SandboxSnapshotStateCreating, "")+`}`)

	sb, err := c.Get(ctx, "sb-1")
	if err != nil {
		t.Fatal(err)
	}
	snap, err := sb.CreateSnapshot(ctx, &sandbox.SnapshotParams{Name: "my-snapshot", TTL: "168h"})
	if err != nil {
		t.Fatal(err)
	}
	if snap.ID() != "snap-1" || snap.SandboxID() != "sb-1" || snap.Name() != "my-snapshot" || snap.State() != godo.SandboxSnapshotStateCreating {
		t.Errorf("snapshot = %+v", snap.Info)
	}
	body := a.bodies[len(a.bodies)-1]
	if body["name"] != "my-snapshot" || body["ttl"] != "168h" {
		t.Errorf("body = %+v", body)
	}

	calls := len(a.calls)
	for _, p := range []*sandbox.SnapshotParams{{Name: "Bad_Name"}, {Name: strings.Repeat("a", 64)}, {TTL: "soon"}, {TTL: "-1h"}} {
		if _, err := sb.CreateSnapshot(ctx, p); err == nil {
			t.Errorf("%+v returned no error", p)
		}
	}
	if len(a.calls) != calls {
		t.Errorf("a request was sent for an invalid option: %v", a.calls[calls:])
	}
	if _, err := sb.CreateSnapshot(ctx, nil); err != nil {
		t.Errorf("nil params returned %v", err)
	}
}

func TestGetRefreshAndDeleteSnapshot(t *testing.T) {
	a, c := newAPI(t)
	ctx := context.Background()
	a.handlers["GET /v2/sandboxes/snapshots/snap-1"] = reply(200, `{"snapshot":`+snapJSON("snap-1", godo.SandboxSnapshotStateComplete, "")+`}`)
	a.handlers["DELETE /v2/sandboxes/snapshots/snap-1"] = reply(200, `{}`)
	a.handlers["DELETE /v2/sandboxes/snapshots/gone"] = reply(404, `{"id":"not_found","message":"gone"}`)

	snap, err := c.GetSnapshot(ctx, "snap-1")
	if err != nil || snap.State() != godo.SandboxSnapshotStateComplete {
		t.Fatalf("GetSnapshot = %+v, %v", snap, err)
	}
	if err := snap.Refresh(ctx); err != nil {
		t.Fatal(err)
	}
	if deleted, err := snap.Delete(ctx); err != nil || !deleted {
		t.Fatalf("Delete = %v, %v", deleted, err)
	}
	if deleted, err := c.DeleteSnapshot(ctx, "gone"); err != nil || deleted {
		t.Errorf("DeleteSnapshot = %v, %v; expected false, nil", deleted, err)
	}
}

func TestListSnapshotsPagesAndFilters(t *testing.T) {
	a, c := newAPI(t)
	var queries []string
	a.handlers["GET /v2/sandboxes/snapshots"] = func(w http.ResponseWriter, r *http.Request) {
		queries = append(queries, r.URL.RawQuery)
		if r.URL.Query().Get("page_token") == "" {
			fmt.Fprint(w, `{"snapshots":[`+snapJSON("a", godo.SandboxSnapshotStateComplete, "")+`],"next_page_token":"t2"}`)
			return
		}
		fmt.Fprint(w, `{"snapshots":[`+snapJSON("b", godo.SandboxSnapshotStateComplete, "")+`]}`)
	}
	pages := c.ListSnapshots(&sandbox.ListSnapshotsParams{SandboxID: "sb-1", Limit: 1})
	var ids []string
	for snap, err := range pages.All(context.Background()) {
		if err != nil {
			t.Fatal(err)
		}
		ids = append(ids, snap.ID())
	}
	if strings.Join(ids, ",") != "a,b" || pages.HasNext() {
		t.Errorf("ids = %v, HasNext = %v", ids, pages.HasNext())
	}
	if !strings.Contains(queries[0], "sandbox_id=sb-1") || !strings.Contains(queries[0], "page_size=1") || !strings.Contains(queries[1], "page_token=t2") {
		t.Errorf("queries = %v", queries)
	}
}

func TestSnapshotWaitUntilReady(t *testing.T) {
	a, c := newAPI(t)
	ctx := context.Background()
	opts := &sandbox.WaitOptions{Timeout: 5 * time.Second, PollInterval: 5 * time.Millisecond}

	states := []godo.SandboxSnapshotState{godo.SandboxSnapshotStateUploading, godo.SandboxSnapshotStateComplete}
	i := 0
	a.handlers["GET /v2/sandboxes/snapshots/snap-1"] = func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"snapshot":`+snapJSON("snap-1", states[min(i, len(states)-1)], "")+`}`)
		i++
	}
	a.handlers["POST /v2/sandboxes/sb-1/snapshots"] = reply(202, `{"snapshot":`+snapJSON("snap-1", godo.SandboxSnapshotStateCreating, "")+`}`)
	a.handlers["GET /v2/sandboxes/sb-1"] = reply(200, `{"sandbox":`+raw("sb-1", godo.SandboxStateRunning)+`}`)
	sb, _ := c.Get(ctx, "sb-1")
	snap, err := sb.CreateSnapshot(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := snap.WaitUntilReady(ctx, opts); err != nil || snap.State() != godo.SandboxSnapshotStateComplete {
		t.Fatalf("WaitUntilReady = %v, state %s", err, snap.State())
	}

	a.handlers["GET /v2/sandboxes/snapshots/bad"] = reply(200, `{"snapshot":`+snapJSON("bad", godo.SandboxSnapshotStateFailed, "disk full")+`}`)
	bad, err := c.GetSnapshot(ctx, "bad")
	if err != nil {
		t.Fatal(err)
	}
	if err := bad.WaitUntilReady(ctx, opts); err == nil || !strings.Contains(err.Error(), "disk full") {
		t.Errorf("failed snapshot returned %v; expected the error message", err)
	}

	a.handlers["GET /v2/sandboxes/snapshots/slow"] = reply(200, `{"snapshot":`+snapJSON("slow", godo.SandboxSnapshotStateCreating, "")+`}`)
	slow, _ := c.GetSnapshot(ctx, "slow")
	err = slow.WaitUntilReady(ctx, &sandbox.WaitOptions{Timeout: 50 * time.Millisecond, PollInterval: 10 * time.Millisecond})
	if err == nil || !strings.Contains(err.Error(), "was not ready in time") {
		t.Errorf("slow snapshot returned %v", err)
	}
}
