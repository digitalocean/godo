package godo

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

func TestMicroDroplets_List(t *testing.T) {
	setup()
	defer teardown()

	jBlob := `{
		"micro_droplets": [
			{"id": "aaa-111", "name": "one", "region": "nyc3", "state": "running"},
			{"id": "bbb-222", "name": "two", "region": "nyc3", "state": "paused"}
		],
		"links": {
			"pages": {
				"next": "http://example.com/v2/microdroplets/instances?page=2"
			}
		},
		"meta": {"total": 2}
	}`

	mux.HandleFunc("/v2/microdroplets/instances", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, jBlob)
	})

	microDroplets, resp, err := client.MicroDroplets.List(ctx, nil)
	if err != nil {
		t.Fatalf("MicroDroplets.List returned error: %v", err)
	}

	expected := []MicroDroplet{
		{ID: "aaa-111", Name: "one", Region: "nyc3", State: MicroDropletStateRunning},
		{ID: "bbb-222", Name: "two", Region: "nyc3", State: MicroDropletStatePaused},
	}
	if !reflect.DeepEqual(microDroplets, expected) {
		t.Errorf("MicroDroplets.List returned %+v, expected %+v", microDroplets, expected)
	}

	if resp.Meta == nil || resp.Meta.Total != 2 {
		t.Errorf("MicroDroplets.List Meta not propagated: %+v", resp.Meta)
	}
	checkCurrentPage(t, resp, 1)
}

func TestMicroDroplets_List_Paginated(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/microdroplets/instances", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		if got, want := r.URL.Query().Get("page"), "2"; got != want {
			t.Errorf("page query = %q, expected %q", got, want)
		}
		if got, want := r.URL.Query().Get("per_page"), "50"; got != want {
			t.Errorf("per_page query = %q, expected %q", got, want)
		}
		fmt.Fprint(w, `{"micro_droplets": [], "meta": {"total": 0}}`)
	})

	_, _, err := client.MicroDroplets.List(ctx, &ListOptions{Page: 2, PerPage: 50})
	if err != nil {
		t.Fatalf("MicroDroplets.List returned error: %v", err)
	}
}

func TestMicroDroplets_ListByRegion(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/microdroplets/instances", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		if got, want := r.URL.Query().Get("region"), "sfo3"; got != want {
			t.Errorf("region query = %q, expected %q", got, want)
		}
		fmt.Fprint(w, `{"micro_droplets": [{"id": "aaa-111", "region": "sfo3"}]}`)
	})

	microDroplets, _, err := client.MicroDroplets.ListByRegion(ctx, "sfo3", nil)
	if err != nil {
		t.Fatalf("MicroDroplets.ListByRegion returned error: %v", err)
	}

	expected := []MicroDroplet{{ID: "aaa-111", Region: "sfo3"}}
	if !reflect.DeepEqual(microDroplets, expected) {
		t.Errorf("MicroDroplets.ListByRegion returned %+v, expected %+v", microDroplets, expected)
	}
}

func TestMicroDroplets_ListByRegion_EmptyRegion(t *testing.T) {
	_, _, err := (&MicroDropletsServiceOp{}).ListByRegion(ctx, "", nil)
	if err == nil {
		t.Fatal("expected error for empty region")
	}
	if _, ok := err.(*ArgError); !ok {
		t.Errorf("expected *ArgError, got %T: %v", err, err)
	}
}

func TestMicroDroplets_ListByName(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/microdroplets/instances", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		if got, want := r.URL.Query().Get("name"), "agent-sandbox-1"; got != want {
			t.Errorf("name query = %q, expected %q", got, want)
		}
		fmt.Fprint(w, `{"micro_droplets": [{"id": "aaa-111", "name": "agent-sandbox-1"}]}`)
	})

	microDroplets, _, err := client.MicroDroplets.ListByName(ctx, "agent-sandbox-1", nil)
	if err != nil {
		t.Fatalf("MicroDroplets.ListByName returned error: %v", err)
	}

	expected := []MicroDroplet{{ID: "aaa-111", Name: "agent-sandbox-1"}}
	if !reflect.DeepEqual(microDroplets, expected) {
		t.Errorf("MicroDroplets.ListByName returned %+v, expected %+v", microDroplets, expected)
	}
}

func TestMicroDroplets_ListByName_EmptyName(t *testing.T) {
	_, _, err := (&MicroDropletsServiceOp{}).ListByName(ctx, "", nil)
	if err == nil {
		t.Fatal("expected error for empty name")
	}
	if _, ok := err.(*ArgError); !ok {
		t.Errorf("expected *ArgError, got %T: %v", err, err)
	}
}

func TestMicroDroplets_Get(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/microdroplets/instances/aaa-111", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"micro_droplet": {
				"id": "aaa-111",
				"name": "sandbox",
				"region": "nyc3",
				"state": "running",
				"size": "sm-1vcpu-1gb",
				"networking": "public",
				"image": "do:microdroplet_image:img-uuid",
				"endpoint": "https://sandbox.example.com",
				"created_at": "2026-07-16T10:00:00Z"
			}
		}`)
	})

	microDroplet, _, err := client.MicroDroplets.Get(ctx, "aaa-111")
	if err != nil {
		t.Fatalf("MicroDroplets.Get returned error: %v", err)
	}

	expected := &MicroDroplet{
		ID:         "aaa-111",
		Name:       "sandbox",
		Region:     "nyc3",
		State:      MicroDropletStateRunning,
		Size:       "sm-1vcpu-1gb",
		Networking: MicroDropletNetworkingPublic,
		Image:      "do:microdroplet_image:img-uuid",
		Endpoint:   "https://sandbox.example.com",
		Created:    "2026-07-16T10:00:00Z",
	}
	if !reflect.DeepEqual(microDroplet, expected) {
		t.Errorf("MicroDroplets.Get returned %+v, expected %+v", microDroplet, expected)
	}
}

func TestMicroDroplets_Get_EmptyID(t *testing.T) {
	_, _, err := (&MicroDropletsServiceOp{}).Get(ctx, "")
	if err == nil {
		t.Fatal("expected error for empty id")
	}
	if _, ok := err.(*ArgError); !ok {
		t.Errorf("expected *ArgError, got %T: %v", err, err)
	}
}

func TestMicroDroplets_Create(t *testing.T) {
	setup()
	defer teardown()

	autoResume := true
	autoPauseEnabled := true
	createRequest := &MicroDropletCreateRequest{
		Name:         "sandbox",
		Region:       "nyc3",
		Size:         "sm-1vcpu-1gb",
		Image:        "do:microdroplet_image:img-uuid",
		Networking:   MicroDropletNetworkingVPC,
		VPCUUID:      "vpc-uuid",
		AutoPause:    &AutoPauseConfig{Enabled: &autoPauseEnabled, IdleTimeout: "5m"},
		AutoResume:   &autoResume,
		HTTPPort:     8080,
		HTTPProtocol: MicroDropletHTTPProtocolHTTPS,
		Environment:  map[string]string{"FOO": "bar"},
		Tags:         []string{"env:dev", "team:agents"},
	}

	mux.HandleFunc("/v2/microdroplets/instances", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)

		expected := map[string]interface{}{
			"name":       "sandbox",
			"region":     "nyc3",
			"size":       "sm-1vcpu-1gb",
			"image":      "do:microdroplet_image:img-uuid",
			"networking": "vpc",
			"vpc_uuid":   "vpc-uuid",
			"auto_pause": map[string]interface{}{
				"enabled":      true,
				"idle_timeout": "5m",
			},
			"auto_resume":   true,
			"http_port":     float64(8080),
			"http_protocol": "https",
			"environment":   map[string]interface{}{"FOO": "bar"},
			"tags":          []interface{}{"env:dev", "team:agents"},
		}

		var got map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if !reflect.DeepEqual(got, expected) {
			t.Errorf("Create body\n got=%#v\nwant=%#v", got, expected)
		}

		fmt.Fprint(w, `{"micro_droplet": {"id": "aaa-111", "name": "sandbox", "state": "creating"}}`)
	})

	microDroplet, _, err := client.MicroDroplets.Create(ctx, createRequest)
	if err != nil {
		t.Fatalf("MicroDroplets.Create returned error: %v", err)
	}

	if microDroplet.ID != "aaa-111" {
		t.Errorf("MicroDroplets.Create returned ID %q, expected %q", microDroplet.ID, "aaa-111")
	}
	if microDroplet.State != MicroDropletStateCreating {
		t.Errorf("MicroDroplets.Create returned State %q, expected %q", microDroplet.State, MicroDropletStateCreating)
	}
}

func TestMicroDroplets_Create_Minimal(t *testing.T) {
	setup()
	defer teardown()

	createRequest := &MicroDropletCreateRequest{
		Name:   "sandbox",
		Region: "nyc3",
		Size:   "sm-1vcpu-1gb",
		Image:  "do:microdroplet_image:img-uuid",
	}

	mux.HandleFunc("/v2/microdroplets/instances", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)

		expected := map[string]interface{}{
			"name":   "sandbox",
			"region": "nyc3",
			"size":   "sm-1vcpu-1gb",
			"image":  "do:microdroplet_image:img-uuid",
		}

		var got map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if !reflect.DeepEqual(got, expected) {
			t.Errorf("Create (minimal) body\n got=%#v\nwant=%#v", got, expected)
		}

		fmt.Fprint(w, `{"micro_droplet": {"id": "aaa-111"}}`)
	})

	if _, _, err := client.MicroDroplets.Create(ctx, createRequest); err != nil {
		t.Fatalf("MicroDroplets.Create returned error: %v", err)
	}
}

func TestMicroDroplets_Create_NilRequest(t *testing.T) {
	_, _, err := (&MicroDropletsServiceOp{}).Create(ctx, nil)
	if err == nil {
		t.Fatal("expected error for nil createRequest")
	}
	if _, ok := err.(*ArgError); !ok {
		t.Errorf("expected *ArgError, got %T: %v", err, err)
	}
}

func TestMicroDroplets_Update_Pause(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/microdroplets/instances/aaa-111", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPatch)

		expected := map[string]interface{}{"state": "paused"}
		var got map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if !reflect.DeepEqual(got, expected) {
			t.Errorf("Update body\n got=%#v\nwant=%#v", got, expected)
		}

		fmt.Fprint(w, `{"micro_droplet": {"id": "aaa-111", "state": "paused"}}`)
	})

	microDroplet, _, err := client.MicroDroplets.Update(ctx, "aaa-111", &MicroDropletUpdateRequest{
		State: MicroDropletStatePaused,
	})
	if err != nil {
		t.Fatalf("MicroDroplets.Update returned error: %v", err)
	}

	if microDroplet.State != MicroDropletStatePaused {
		t.Errorf("MicroDroplets.Update returned State %q, expected %q", microDroplet.State, MicroDropletStatePaused)
	}
}

func TestMicroDroplets_Update_Resume(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/microdroplets/instances/aaa-111", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPatch)

		expected := map[string]interface{}{"state": "running"}
		var got map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if !reflect.DeepEqual(got, expected) {
			t.Errorf("Update body\n got=%#v\nwant=%#v", got, expected)
		}

		fmt.Fprint(w, `{"micro_droplet": {"id": "aaa-111", "state": "running"}}`)
	})

	microDroplet, _, err := client.MicroDroplets.Update(ctx, "aaa-111", &MicroDropletUpdateRequest{
		State: MicroDropletStateRunning,
	})
	if err != nil {
		t.Fatalf("MicroDroplets.Update returned error: %v", err)
	}

	if microDroplet.State != MicroDropletStateRunning {
		t.Errorf("MicroDroplets.Update returned State %q, expected %q", microDroplet.State, MicroDropletStateRunning)
	}
}

func TestMicroDroplets_Update_EmptyID(t *testing.T) {
	_, _, err := (&MicroDropletsServiceOp{}).Update(ctx, "", &MicroDropletUpdateRequest{State: MicroDropletStatePaused})
	if err == nil {
		t.Fatal("expected error for empty id")
	}
	if _, ok := err.(*ArgError); !ok {
		t.Errorf("expected *ArgError, got %T: %v", err, err)
	}
}

func TestMicroDroplets_Update_NilRequest(t *testing.T) {
	_, _, err := (&MicroDropletsServiceOp{}).Update(ctx, "aaa-111", nil)
	if err == nil {
		t.Fatal("expected error for nil updateRequest")
	}
	if _, ok := err.(*ArgError); !ok {
		t.Errorf("expected *ArgError, got %T: %v", err, err)
	}
}

func TestMicroDroplets_Delete(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/microdroplets/instances/aaa-111", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		w.WriteHeader(http.StatusNoContent)
	})

	if _, err := client.MicroDroplets.Delete(ctx, "aaa-111"); err != nil {
		t.Fatalf("MicroDroplets.Delete returned error: %v", err)
	}
}

func TestMicroDroplets_Delete_EmptyID(t *testing.T) {
	_, err := (&MicroDropletsServiceOp{}).Delete(ctx, "")
	if err == nil {
		t.Fatal("expected error for empty id")
	}
	if _, ok := err.(*ArgError); !ok {
		t.Errorf("expected *ArgError, got %T: %v", err, err)
	}
}

func TestMicroDroplets_ListSnapshots(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/microdroplets/instances/aaa-111/snapshots", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"snapshots": [
				{"id": "snap-1", "micro_droplet_id": "aaa-111", "status": "SNAPSHOT_AVAILABLE", "memory_bytes": 1024, "disk_bytes": 2048},
				{"id": "snap-2", "micro_droplet_id": "aaa-111", "status": "SNAPSHOT_CREATING"}
			],
			"meta": {"total": 2}
		}`)
	})

	snapshots, resp, err := client.MicroDroplets.ListSnapshots(ctx, "aaa-111", nil)
	if err != nil {
		t.Fatalf("MicroDroplets.ListSnapshots returned error: %v", err)
	}

	expected := []MicroDropletSnapshot{
		{ID: "snap-1", MicroDropletID: "aaa-111", Status: MicroDropletSnapshotStatusAvailable, MemoryBytes: 1024, DiskBytes: 2048},
		{ID: "snap-2", MicroDropletID: "aaa-111", Status: MicroDropletSnapshotStatusCreating},
	}
	if !reflect.DeepEqual(snapshots, expected) {
		t.Errorf("MicroDroplets.ListSnapshots returned %+v, expected %+v", snapshots, expected)
	}

	if resp.Meta == nil || resp.Meta.Total != 2 {
		t.Errorf("MicroDroplets.ListSnapshots Meta not propagated: %+v", resp.Meta)
	}
}

func TestMicroDroplets_ListSnapshots_EmptyID(t *testing.T) {
	_, _, err := (&MicroDropletsServiceOp{}).ListSnapshots(ctx, "", nil)
	if err == nil {
		t.Fatal("expected error for empty id")
	}
	if _, ok := err.(*ArgError); !ok {
		t.Errorf("expected *ArgError, got %T: %v", err, err)
	}
}

func TestMicroDroplet_URN(t *testing.T) {
	md := MicroDroplet{ID: "aaa-111"}
	want := "do:microdroplet:aaa-111"
	if got := md.URN(); got != want {
		t.Errorf("MicroDroplet.URN = %q, expected %q", got, want)
	}
}
