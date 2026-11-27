package godo

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"testing"
	"time"
)

const sandboxWorkspaceJSON = `{"workspace_id":"ws-1","name":"my-workspace","size_limit_gib":10,"state":"WORKSPACE_STATE_ATTACHED","attached_sandbox_id":"sb-1","last_saved_at":"2026-11-20T12:30:00Z","created_at":"2026-11-20T12:00:00Z"}`

func sandboxWorkspaceWant() *SandboxWorkspace {
	return &SandboxWorkspace{
		ID:                "ws-1",
		Name:              "my-workspace",
		SizeLimitGiB:      10,
		State:             SandboxWorkspaceStateAttached,
		AttachedSandboxID: "sb-1",
		LastSavedAt:       Timestamp{time.Date(2026, 11, 20, 12, 30, 0, 0, time.UTC)},
		CreatedAt:         Timestamp{time.Date(2026, 11, 20, 12, 0, 0, 0, time.UTC)},
	}
}

func TestSandboxes_CreateWorkspace(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/sandboxes/workspaces", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var got map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&got)
		want := map[string]interface{}{"name": "my-workspace", "size_limit_gib": float64(10)}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("request body = %+v, expected %+v", got, want)
		}
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, `{"workspace":%s}`, sandboxWorkspaceJSON)
	})

	ws, _, err := client.Sandboxes.CreateWorkspace(ctx, &SandboxWorkspaceCreateRequest{Name: "my-workspace", SizeLimitGiB: 10})
	if err != nil || !reflect.DeepEqual(ws, sandboxWorkspaceWant()) {
		t.Fatalf("CreateWorkspace returned %+v, %v", ws, err)
	}
	if _, _, err := client.Sandboxes.CreateWorkspace(ctx, nil); err == nil {
		t.Error("a nil request returned no error")
	}
}

func TestSandboxes_CreateWorkspaceOmitsEmptyName(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/sandboxes/workspaces", func(w http.ResponseWriter, r *http.Request) {
		var got map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&got)
		if !reflect.DeepEqual(got, map[string]interface{}{"size_limit_gib": float64(5)}) {
			t.Errorf("request body = %+v, expected only size_limit_gib", got)
		}
		fmt.Fprintf(w, `{"workspace":%s}`, sandboxWorkspaceJSON)
	})
	if _, _, err := client.Sandboxes.CreateWorkspace(ctx, &SandboxWorkspaceCreateRequest{SizeLimitGiB: 5}); err != nil {
		t.Fatalf("CreateWorkspace returned error: %v", err)
	}
}

func TestSandboxes_GetDeleteWorkspace(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/sandboxes/workspaces/ws-1", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			fmt.Fprintf(w, `{"workspace":%s}`, sandboxWorkspaceJSON)
		case http.MethodDelete:
			fmt.Fprint(w, `{}`)
		default:
			t.Errorf("unexpected method %s", r.Method)
		}
	})

	got, _, err := client.Sandboxes.GetWorkspace(ctx, "ws-1")
	if err != nil || !reflect.DeepEqual(got, sandboxWorkspaceWant()) {
		t.Fatalf("GetWorkspace returned %+v, %v", got, err)
	}
	if _, err := client.Sandboxes.DeleteWorkspace(ctx, "ws-1"); err != nil {
		t.Fatalf("DeleteWorkspace returned error: %v", err)
	}
	if _, _, err := client.Sandboxes.GetWorkspace(ctx, ""); err == nil {
		t.Error("get: no error for missing id")
	}
	if _, err := client.Sandboxes.DeleteWorkspace(ctx, ""); err == nil {
		t.Error("delete: no error for missing id")
	}
}

func TestSandboxes_ListWorkspaces(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/sandboxes/workspaces", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		testFormValues(t, r, values{"state": "WORKSPACE_STATE_ATTACHED", "page_size": "1", "page_token": "t1"})
		fmt.Fprintf(w, `{"workspaces":[%s],"next_page_token":"t2"}`, sandboxWorkspaceJSON)
	})
	list, _, err := client.Sandboxes.ListWorkspaces(ctx, &SandboxWorkspaceListOptions{State: SandboxWorkspaceStateAttached, PageSize: 1, PageToken: "t1"})
	want := &SandboxWorkspaceList{Workspaces: []SandboxWorkspace{*sandboxWorkspaceWant()}, NextPageToken: "t2"}
	if err != nil || !reflect.DeepEqual(list, want) {
		t.Fatalf("ListWorkspaces returned %+v, %v; expected %+v", list, err, want)
	}
}
