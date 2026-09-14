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
				"next": "http://example.com/v2/microdroplets?page=2"
			}
		},
		"meta": {"total": 2}
	}`

	mux.HandleFunc("/v2/microdroplets", func(w http.ResponseWriter, r *http.Request) {
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

	mux.HandleFunc("/v2/microdroplets", func(w http.ResponseWriter, r *http.Request) {
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

	mux.HandleFunc("/v2/microdroplets", func(w http.ResponseWriter, r *http.Request) {
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

	mux.HandleFunc("/v2/microdroplets", func(w http.ResponseWriter, r *http.Request) {
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

	mux.HandleFunc("/v2/microdroplets/aaa-111", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"micro_droplet": {
				"id": "aaa-111",
				"name": "sandbox",
				"region": "nyc3",
				"state": "running",
				"size": {"cpu": 2, "memory": 4096, "disk": 80},
				"networking": "public",
				"source": {"oci_ref": "docker.io/library/nginx:1.27"},
				"urls": [{"hostname": "sandbox.example.com", "port": 8080, "default": true, "status": "ACTIVE"}],
				"ports": [8080],
				"tags": ["env:dev"],
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
		Size:       &MicroDropletSize{CPU: 2, Memory: 4096, Disk: 80},
		Networking: MicroDropletNetworkingPublic,
		Source:     &MicroDropletSource{OCIRef: "docker.io/library/nginx:1.27"},
		URLs: []MicroDropletURL{
			{Hostname: "sandbox.example.com", Port: 8080, Default: true, Status: MicroDropletURLStatusActive},
		},
		Ports:   []uint32{8080},
		Tags:    []string{"env:dev"},
		Created: "2026-07-16T10:00:00Z",
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
		Size:         &MicroDropletSizeRequest{CPU: 2, Memory: 4096},
		Source:       &MicroDropletSource{OCIRef: "docker.io/library/nginx:1.27"},
		Networking:   MicroDropletNetworkingVPC,
		VPCUUID:      "vpc-uuid",
		AutoPause:    &AutoPauseConfig{Enabled: &autoPauseEnabled, IdleTimeout: "5m"},
		AutoResume:   &autoResume,
		HTTPPort:     8080,
		HTTPProtocol: MicroDropletHTTPProtocolHTTP2,
		Ports:        []uint32{80, 8080},
		Environment:  map[string]string{"FOO": "bar"},
		Tags:         []string{"env:dev", "team:agents"},
	}

	mux.HandleFunc("/v2/microdroplets", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)

		expected := map[string]interface{}{
			"name":   "sandbox",
			"region": "nyc3",
			"size": map[string]interface{}{
				"cpu":    float64(2),
				"memory": float64(4096),
			},
			"source": map[string]interface{}{
				"oci_ref": "docker.io/library/nginx:1.27",
			},
			"networking": "vpc",
			"vpc_uuid":   "vpc-uuid",
			"auto_pause": map[string]interface{}{
				"enabled":      true,
				"idle_timeout": "5m",
			},
			"auto_resume":   true,
			"http_port":     float64(8080),
			"http_protocol": "http2",
			"ports":         []interface{}{float64(80), float64(8080)},
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

func TestMicroDroplets_Create_FromCheckpoint(t *testing.T) {
	setup()
	defer teardown()

	createRequest := &MicroDropletCreateRequest{
		Name:   "sandbox-clone",
		Source: &MicroDropletSource{CheckpointID: "chk-1"},
	}

	mux.HandleFunc("/v2/microdroplets", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)

		expected := map[string]interface{}{
			"name": "sandbox-clone",
			"source": map[string]interface{}{
				"checkpoint_id": "chk-1",
			},
		}

		var got map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if !reflect.DeepEqual(got, expected) {
			t.Errorf("Create (checkpoint) body\n got=%#v\nwant=%#v", got, expected)
		}

		fmt.Fprint(w, `{"micro_droplet": {"id": "bbb-222"}}`)
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

func TestMicroDroplets_Pause(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/microdroplets/aaa-111/pause", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `{"micro_droplet": {"id": "aaa-111", "state": "paused"}}`)
	})

	microDroplet, _, err := client.MicroDroplets.Pause(ctx, "aaa-111")
	if err != nil {
		t.Fatalf("MicroDroplets.Pause returned error: %v", err)
	}

	if microDroplet.State != MicroDropletStatePaused {
		t.Errorf("MicroDroplets.Pause returned State %q, expected %q", microDroplet.State, MicroDropletStatePaused)
	}
}

func TestMicroDroplets_Pause_EmptyID(t *testing.T) {
	_, _, err := (&MicroDropletsServiceOp{}).Pause(ctx, "")
	if err == nil {
		t.Fatal("expected error for empty id")
	}
	if _, ok := err.(*ArgError); !ok {
		t.Errorf("expected *ArgError, got %T: %v", err, err)
	}
}

func TestMicroDroplets_Resume(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/microdroplets/aaa-111/resume", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `{"micro_droplet": {"id": "aaa-111", "state": "running"}}`)
	})

	microDroplet, _, err := client.MicroDroplets.Resume(ctx, "aaa-111")
	if err != nil {
		t.Fatalf("MicroDroplets.Resume returned error: %v", err)
	}

	if microDroplet.State != MicroDropletStateRunning {
		t.Errorf("MicroDroplets.Resume returned State %q, expected %q", microDroplet.State, MicroDropletStateRunning)
	}
}

func TestMicroDroplets_Resume_EmptyID(t *testing.T) {
	_, _, err := (&MicroDropletsServiceOp{}).Resume(ctx, "")
	if err == nil {
		t.Fatal("expected error for empty id")
	}
	if _, ok := err.(*ArgError); !ok {
		t.Errorf("expected *ArgError, got %T: %v", err, err)
	}
}

func TestMicroDroplets_Delete(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/microdroplets/aaa-111", func(w http.ResponseWriter, r *http.Request) {
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

func TestMicroDroplets_ListCheckpoints(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/microdroplets/checkpoints", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		if got, want := r.URL.Query().Get("micro_droplet_id"), "aaa-111"; got != want {
			t.Errorf("micro_droplet_id query = %q, expected %q", got, want)
		}
		fmt.Fprint(w, `{
			"checkpoints": [
				{
					"id": "chk-1",
					"micro_droplet_id": "aaa-111",
					"micro_droplet_name": "sandbox",
					"region": "nyc3",
					"status": "CHECKPOINT_AVAILABLE",
					"memory_bytes": 1024,
					"disk_bytes": 2048
				},
				{"id": "chk-2", "micro_droplet_id": "aaa-111", "status": "CHECKPOINT_CREATING"}
			],
			"meta": {"total": 2}
		}`)
	})

	checkpoints, resp, err := client.MicroDroplets.ListCheckpoints(ctx, &ListMicroDropletCheckpointsOptions{
		MicroDropletID: "aaa-111",
	})
	if err != nil {
		t.Fatalf("MicroDroplets.ListCheckpoints returned error: %v", err)
	}

	expected := []MicroDropletCheckpoint{
		{
			ID:               "chk-1",
			MicroDropletID:   "aaa-111",
			MicroDropletName: "sandbox",
			Region:           "nyc3",
			Status:           MicroDropletCheckpointStatusAvailable,
			MemoryBytes:      1024,
			DiskBytes:        2048,
		},
		{ID: "chk-2", MicroDropletID: "aaa-111", Status: MicroDropletCheckpointStatusCreating},
	}
	if !reflect.DeepEqual(checkpoints, expected) {
		t.Errorf("MicroDroplets.ListCheckpoints returned %+v, expected %+v", checkpoints, expected)
	}

	if resp.Meta == nil || resp.Meta.Total != 2 {
		t.Errorf("MicroDroplets.ListCheckpoints Meta not propagated: %+v", resp.Meta)
	}
}

func TestMicroDroplets_ListCheckpoints_All(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/microdroplets/checkpoints", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		if got := r.URL.Query().Get("micro_droplet_id"); got != "" {
			t.Errorf("unexpected micro_droplet_id query = %q", got)
		}
		fmt.Fprint(w, `{"checkpoints": [{"id": "chk-1"}], "meta": {"total": 1}}`)
	})

	checkpoints, _, err := client.MicroDroplets.ListCheckpoints(ctx, nil)
	if err != nil {
		t.Fatalf("MicroDroplets.ListCheckpoints returned error: %v", err)
	}
	if len(checkpoints) != 1 || checkpoints[0].ID != "chk-1" {
		t.Errorf("MicroDroplets.ListCheckpoints returned %+v", checkpoints)
	}
}

func TestMicroDroplets_CreateCheckpoint(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/microdroplets/aaa-111/checkpoints", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)

		var got map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		expected := map[string]interface{}{"name": "chk-named"}
		if !reflect.DeepEqual(got, expected) {
			t.Errorf("CreateCheckpoint body\n got=%#v\nwant=%#v", got, expected)
		}

		fmt.Fprint(w, `{"checkpoint": {"id": "chk-1", "micro_droplet_id": "aaa-111", "status": "CHECKPOINT_CREATING"}}`)
	})

	checkpoint, _, err := client.MicroDroplets.CreateCheckpoint(ctx, "aaa-111", &MicroDropletCheckpointCreateRequest{Name: "chk-named"})
	if err != nil {
		t.Fatalf("MicroDroplets.CreateCheckpoint returned error: %v", err)
	}
	if checkpoint.ID != "chk-1" || checkpoint.Status != MicroDropletCheckpointStatusCreating {
		t.Errorf("MicroDroplets.CreateCheckpoint returned %+v", checkpoint)
	}
}

func TestMicroDroplets_CreateCheckpoint_EmptyID(t *testing.T) {
	_, _, err := (&MicroDropletsServiceOp{}).CreateCheckpoint(ctx, "", nil)
	if err == nil {
		t.Fatal("expected error for empty microDropletID")
	}
	if _, ok := err.(*ArgError); !ok {
		t.Errorf("expected *ArgError, got %T: %v", err, err)
	}
}

func TestMicroDroplets_GetCheckpoint(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/microdroplets/checkpoints/chk-1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{"checkpoint": {"id": "chk-1", "status": "CHECKPOINT_AVAILABLE", "region": "nyc3"}}`)
	})

	checkpoint, _, err := client.MicroDroplets.GetCheckpoint(ctx, "chk-1")
	if err != nil {
		t.Fatalf("MicroDroplets.GetCheckpoint returned error: %v", err)
	}
	expected := &MicroDropletCheckpoint{
		ID:     "chk-1",
		Status: MicroDropletCheckpointStatusAvailable,
		Region: "nyc3",
	}
	if !reflect.DeepEqual(checkpoint, expected) {
		t.Errorf("MicroDroplets.GetCheckpoint returned %+v, expected %+v", checkpoint, expected)
	}
}

func TestMicroDroplets_GetCheckpoint_EmptyID(t *testing.T) {
	_, _, err := (&MicroDropletsServiceOp{}).GetCheckpoint(ctx, "")
	if err == nil {
		t.Fatal("expected error for empty id")
	}
	if _, ok := err.(*ArgError); !ok {
		t.Errorf("expected *ArgError, got %T: %v", err, err)
	}
}

func TestMicroDroplets_DeleteCheckpoint(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/microdroplets/checkpoints/chk-1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		w.WriteHeader(http.StatusNoContent)
	})

	if _, err := client.MicroDroplets.DeleteCheckpoint(ctx, "chk-1"); err != nil {
		t.Fatalf("MicroDroplets.DeleteCheckpoint returned error: %v", err)
	}
}

func TestMicroDroplets_DeleteCheckpoint_EmptyID(t *testing.T) {
	_, err := (&MicroDropletsServiceOp{}).DeleteCheckpoint(ctx, "")
	if err == nil {
		t.Fatal("expected error for empty id")
	}
	if _, ok := err.(*ArgError); !ok {
		t.Errorf("expected *ArgError, got %T: %v", err, err)
	}
}

func TestMicroDroplets_GetCreateOptions(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/microdroplets/options", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"regions": [{"slug": "nyc1", "available": true}],
			"default_region": "nyc1",
			"sizes": [{
				"size": {"cpu": 2, "memory": 4096, "disk": 80},
				"available": true,
				"pricing": {"price_per_hour": 0.0119, "price_per_month": 8.0}
			}],
			"features": [{"name": "microdroplet", "enabled": true}],
			"account_limits": {"max_concurrent_running": 10, "max_total_count": 25}
		}`)
	})

	opts, _, err := client.MicroDroplets.GetCreateOptions(ctx)
	if err != nil {
		t.Fatalf("MicroDroplets.GetCreateOptions returned error: %v", err)
	}

	expected := &MicroDropletCreateOptions{
		Regions:       []MicroDropletRegionOption{{Slug: "nyc1", Available: true}},
		DefaultRegion: "nyc1",
		Sizes: []MicroDropletSizeOption{{
			Size:      MicroDropletSize{CPU: 2, Memory: 4096, Disk: 80},
			Available: true,
			Pricing:   &MicroDropletSizePricing{PricePerHour: 0.0119, PricePerMonth: 8.0},
		}},
		Features: []MicroDropletFeatureOption{{Name: "microdroplet", Enabled: true}},
		AccountLimits: &MicroDropletAccountLimits{
			MaxConcurrentRunning: 10,
			MaxTotalCount:        25,
		},
	}
	if !reflect.DeepEqual(opts, expected) {
		t.Errorf("MicroDroplets.GetCreateOptions returned %+v, expected %+v", opts, expected)
	}
}

func TestMicroDroplet_URN(t *testing.T) {
	md := MicroDroplet{ID: "aaa-111"}
	want := "do:microdroplet:aaa-111"
	if got := md.URN(); got != want {
		t.Errorf("MicroDroplet.URN = %q, expected %q", got, want)
	}
}
