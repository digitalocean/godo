package godo

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"testing"
	"time"
)

const sandboxJSON = `{"sandbox_id":"sb-1","template":"coding-base","state":"SANDBOX_STATE_RUNNING","size_slug":"mars-2vcpu-4gb","created_at":"2026-11-20T12:00:00Z"}`

func sandboxWant(state SandboxState) *Sandbox {
	return &Sandbox{
		ID:        "sb-1",
		Template:  "coding-base",
		State:     state,
		SizeSlug:  "mars-2vcpu-4gb",
		CreatedAt: Timestamp{time.Date(2026, 11, 20, 12, 0, 0, 0, time.UTC)},
	}
}

func TestSandboxes_Create(t *testing.T) {
	setup()
	defer teardown()

	allow := false
	mux.HandleFunc("/v2/sandboxes", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var got map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatalf("decode: %v", err)
		}
		want := map[string]interface{}{
			"template":              "coding-base",
			"size_slug":             "mars-2vcpu-4gb",
			"allow_internet_access": false,
			"env_vars":              map[string]interface{}{"A": "1"},
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("request body = %+v, expected %+v", got, want)
		}
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, `{"sandbox":%s}`, sandboxJSON)
	})

	sb, _, err := client.Sandboxes.Create(ctx, &SandboxCreateRequest{
		Template:            "coding-base",
		SizeSlug:            "mars-2vcpu-4gb",
		AllowInternetAccess: &allow,
		EnvVars:             map[string]string{"A": "1"},
	})
	if err != nil {
		t.Fatalf("Sandboxes.Create returned error: %v", err)
	}
	if !reflect.DeepEqual(sb, sandboxWant(SandboxStateRunning)) {
		t.Errorf("Sandboxes.Create returned %+v, expected %+v", sb, sandboxWant(SandboxStateRunning))
	}
}

func TestSandboxes_Create_omitsUnsetOptions(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/sandboxes", func(w http.ResponseWriter, r *http.Request) {
		var got map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&got)
		if _, ok := got["allow_internet_access"]; ok {
			t.Errorf("allow_internet_access was sent: %+v", got)
		}
		if _, ok := got["env_vars"]; ok {
			t.Errorf("env_vars was sent: %+v", got)
		}
		fmt.Fprintf(w, `{"sandbox":%s}`, sandboxJSON)
	})

	if _, _, err := client.Sandboxes.Create(ctx, &SandboxCreateRequest{Template: "t", SizeSlug: "s"}); err != nil {
		t.Fatalf("Sandboxes.Create returned error: %v", err)
	}
}

func TestSandboxes_Get(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/sandboxes/sb-1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprintf(w, `{"sandbox":%s}`, sandboxJSON)
	})

	sb, _, err := client.Sandboxes.Get(ctx, "sb-1")
	if err != nil {
		t.Fatalf("Sandboxes.Get returned error: %v", err)
	}
	if !reflect.DeepEqual(sb, sandboxWant(SandboxStateRunning)) {
		t.Errorf("Sandboxes.Get returned %+v, expected %+v", sb, sandboxWant(SandboxStateRunning))
	}
}

func TestSandboxes_List(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/sandboxes", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		testFormValues(t, r, values{
			"state":      "SANDBOX_STATE_RUNNING",
			"page_size":  "1",
			"page_token": "t1",
		})
		fmt.Fprintf(w, `{"sandboxes":[%s],"next_page_token":"t2"}`, sandboxJSON)
	})

	list, _, err := client.Sandboxes.List(ctx, &SandboxListOptions{
		State:     SandboxStateRunning,
		PageSize:  1,
		PageToken: "t1",
	})
	if err != nil {
		t.Fatalf("Sandboxes.List returned error: %v", err)
	}
	want := &SandboxList{Sandboxes: []Sandbox{*sandboxWant(SandboxStateRunning)}, NextPageToken: "t2"}
	if !reflect.DeepEqual(list, want) {
		t.Errorf("Sandboxes.List returned %+v, expected %+v", list, want)
	}
}

func TestSandboxes_PauseResumeDelete(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/sandboxes/sb-1/pause", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `{"sandbox":{"sandbox_id":"sb-1","state":"SANDBOX_STATE_PAUSED"}}`)
	})
	mux.HandleFunc("/v2/sandboxes/sb-1/resume", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `{"sandbox":{"sandbox_id":"sb-1","state":"SANDBOX_STATE_RUNNING"}}`)
	})
	mux.HandleFunc("/v2/sandboxes/sb-1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		fmt.Fprint(w, `{}`)
	})

	paused, _, err := client.Sandboxes.Pause(ctx, "sb-1")
	if err != nil || paused.State != SandboxStatePaused {
		t.Fatalf("Sandboxes.Pause returned %+v, %v", paused, err)
	}
	resumed, _, err := client.Sandboxes.Resume(ctx, "sb-1")
	if err != nil || resumed.State != SandboxStateRunning {
		t.Fatalf("Sandboxes.Resume returned %+v, %v", resumed, err)
	}
	if _, err := client.Sandboxes.Delete(ctx, "sb-1"); err != nil {
		t.Fatalf("Sandboxes.Delete returned error: %v", err)
	}
}

func TestSandboxes_ErrorsAndValidation(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/sandboxes/gone", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"id":"not_found","message":"Sandbox not found for your team.","request_id":"req-1"}`)
	})

	_, _, err := client.Sandboxes.Get(ctx, "gone")
	apiErr, ok := err.(*ErrorResponse)
	if !ok || apiErr.Response.StatusCode != http.StatusNotFound || apiErr.RequestID != "req-1" {
		t.Fatalf("Sandboxes.Get returned %#v, expected a 404 ErrorResponse", err)
	}

	if _, _, err := client.Sandboxes.Get(ctx, ""); err == nil {
		t.Error("Sandboxes.Get with an empty id returned no error")
	}
	if _, _, err := client.Sandboxes.Create(ctx, nil); err == nil {
		t.Error("Sandboxes.Create with a nil request returned no error")
	}
	if _, err := client.Sandboxes.Delete(ctx, ""); err == nil {
		t.Error("Sandboxes.Delete with an empty id returned no error")
	}
}
