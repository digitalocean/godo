package godo

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"testing"
	"time"
)

const sandboxSnapshotJSON = `{"snapshot_id":"snap-1","sandbox_id":"sb-1","name":"my-snapshot","state":"SNAPSHOT_STATE_COMPLETE","size_bytes":"1073741824","created_at":"2026-11-20T12:00:00Z","ttl":"168h"}`

func sandboxSnapshotWant(state SandboxSnapshotState) *SandboxSnapshot {
	return &SandboxSnapshot{
		ID: "snap-1", SandboxID: "sb-1", Name: "my-snapshot", State: state,
		SizeBytes: "1073741824", TTL: "168h",
		CreatedAt: Timestamp{time.Date(2026, 11, 20, 12, 0, 0, 0, time.UTC)},
	}
}

func TestSandboxes_CreateSnapshot(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/sandboxes/sb-1/snapshots", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var got map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&got)
		want := map[string]interface{}{"name": "my-snapshot", "ttl": "168h"}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("request body = %+v, expected %+v", got, want)
		}
		w.WriteHeader(http.StatusAccepted)
		fmt.Fprintf(w, `{"snapshot":%s}`, sandboxSnapshotJSON)
	})

	snap, _, err := client.Sandboxes.CreateSnapshot(ctx, "sb-1", &SandboxSnapshotCreateRequest{Name: "my-snapshot", TTL: "168h"})
	if err != nil || !reflect.DeepEqual(snap, sandboxSnapshotWant(SandboxSnapshotStateComplete)) {
		t.Fatalf("CreateSnapshot returned %+v, %v", snap, err)
	}
	if snap.Size() != 1073741824 {
		t.Errorf("Size = %d", snap.Size())
	}
	if _, _, err := client.Sandboxes.CreateSnapshot(ctx, "", nil); err == nil {
		t.Error("an empty sandbox id returned no error")
	}
}

func TestSandboxes_CreateSnapshotWithoutBody(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/sandboxes/sb-1/snapshots", func(w http.ResponseWriter, r *http.Request) {
		if r.ContentLength > 0 {
			t.Errorf("a nil request sent a body of %d bytes", r.ContentLength)
		}
		fmt.Fprint(w, `{"snapshot":{"snapshot_id":"snap-1","sandbox_id":"sb-1","state":"SNAPSHOT_STATE_CREATING","created_at":"2026-11-20T12:00:00Z"}}`)
	})

	snap, _, err := client.Sandboxes.CreateSnapshot(ctx, "sb-1", nil)
	if err != nil || snap.State != SandboxSnapshotStateCreating || snap.Size() != 0 {
		t.Fatalf("CreateSnapshot returned %+v, %v", snap, err)
	}
}

func TestSandboxes_GetAndDeleteSnapshot(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/sandboxes/snapshots/snap-1", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			fmt.Fprintf(w, `{"snapshot":%s}`, sandboxSnapshotJSON)
		case http.MethodDelete:
			fmt.Fprint(w, `{}`)
		default:
			t.Errorf("unexpected method %s", r.Method)
		}
	})

	snap, _, err := client.Sandboxes.GetSnapshot(ctx, "snap-1")
	if err != nil || !reflect.DeepEqual(snap, sandboxSnapshotWant(SandboxSnapshotStateComplete)) {
		t.Fatalf("GetSnapshot returned %+v, %v", snap, err)
	}
	if _, err := client.Sandboxes.DeleteSnapshot(ctx, "snap-1"); err != nil {
		t.Fatalf("DeleteSnapshot returned %v", err)
	}
	if _, _, err := client.Sandboxes.GetSnapshot(ctx, ""); err == nil {
		t.Error("an empty snapshot id returned no error")
	}
	if _, err := client.Sandboxes.DeleteSnapshot(ctx, ""); err == nil {
		t.Error("an empty snapshot id returned no error")
	}
}

func TestSandboxes_ListSnapshots(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/sandboxes/snapshots", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		q := r.URL.Query()
		if q.Get("sandbox_id") != "sb-1" || q.Get("page_size") != "10" || q.Get("page_token") != "tok" {
			t.Errorf("query = %v", q)
		}
		fmt.Fprintf(w, `{"snapshots":[%s],"next_page_token":"next"}`, sandboxSnapshotJSON)
	})

	list, _, err := client.Sandboxes.ListSnapshots(ctx, &SandboxSnapshotListOptions{SandboxID: "sb-1", PageSize: 10, PageToken: "tok"})
	if err != nil || len(list.Snapshots) != 1 || list.NextPageToken != "next" ||
		!reflect.DeepEqual(&list.Snapshots[0], sandboxSnapshotWant(SandboxSnapshotStateComplete)) {
		t.Fatalf("ListSnapshots returned %+v, %v", list, err)
	}
}
