package godo

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

func TestMicroVMs_List(t *testing.T) {
	setup()
	defer teardown()

	jBlob := `{
		"microvms": [
			{"id": "aaa-111", "name": "one", "region": "nyc3", "state": "running"},
			{"id": "bbb-222", "name": "two", "region": "nyc3", "state": "paused"}
		],
		"links": {
			"pages": {
				"next": "http://example.com/v2/microvms?page=2"
			}
		},
		"meta": {"total": 2}
	}`

	mux.HandleFunc("/v2/microvms", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, jBlob)
	})

	microVMs, resp, err := client.MicroVMs.List(ctx, nil)
	if err != nil {
		t.Fatalf("MicroVMs.List returned error: %v", err)
	}

	expected := []MicroVM{
		{ID: "aaa-111", Name: "one", Region: "nyc3", State: MicroVMStateRunning},
		{ID: "bbb-222", Name: "two", Region: "nyc3", State: MicroVMStatePaused},
	}
	if !reflect.DeepEqual(microVMs, expected) {
		t.Errorf("MicroVMs.List returned %+v, expected %+v", microVMs, expected)
	}

	if resp.Meta == nil || resp.Meta.Total != 2 {
		t.Errorf("MicroVMs.List Meta not propagated: %+v", resp.Meta)
	}
	checkCurrentPage(t, resp, 1)
}

func TestMicroVMs_List_Paginated(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/microvms", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		if got, want := r.URL.Query().Get("page"), "2"; got != want {
			t.Errorf("page query = %q, expected %q", got, want)
		}
		if got, want := r.URL.Query().Get("per_page"), "50"; got != want {
			t.Errorf("per_page query = %q, expected %q", got, want)
		}
		fmt.Fprint(w, `{"microvms": [], "meta": {"total": 0}}`)
	})

	_, _, err := client.MicroVMs.List(ctx, &ListOptions{Page: 2, PerPage: 50})
	if err != nil {
		t.Fatalf("MicroVMs.List returned error: %v", err)
	}
}

func TestMicroVMs_ListByRegion(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/microvms", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		if got, want := r.URL.Query().Get("region"), "sfo3"; got != want {
			t.Errorf("region query = %q, expected %q", got, want)
		}
		fmt.Fprint(w, `{"microvms": [{"id": "aaa-111", "region": "sfo3"}]}`)
	})

	microVMs, _, err := client.MicroVMs.ListByRegion(ctx, "sfo3", nil)
	if err != nil {
		t.Fatalf("MicroVMs.ListByRegion returned error: %v", err)
	}

	expected := []MicroVM{{ID: "aaa-111", Region: "sfo3"}}
	if !reflect.DeepEqual(microVMs, expected) {
		t.Errorf("MicroVMs.ListByRegion returned %+v, expected %+v", microVMs, expected)
	}
}

func TestMicroVMs_ListByRegion_EmptyRegion(t *testing.T) {
	_, _, err := (&MicroVMsServiceOp{}).ListByRegion(ctx, "", nil)
	if err == nil {
		t.Fatal("expected error for empty region")
	}
	if _, ok := err.(*ArgError); !ok {
		t.Errorf("expected *ArgError, got %T: %v", err, err)
	}
}

func TestMicroVMs_ListByName(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/microvms", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		if got, want := r.URL.Query().Get("name"), "agent-sandbox-1"; got != want {
			t.Errorf("name query = %q, expected %q", got, want)
		}
		fmt.Fprint(w, `{"microvms": [{"id": "aaa-111", "name": "agent-sandbox-1"}]}`)
	})

	microVMs, _, err := client.MicroVMs.ListByName(ctx, "agent-sandbox-1", nil)
	if err != nil {
		t.Fatalf("MicroVMs.ListByName returned error: %v", err)
	}

	expected := []MicroVM{{ID: "aaa-111", Name: "agent-sandbox-1"}}
	if !reflect.DeepEqual(microVMs, expected) {
		t.Errorf("MicroVMs.ListByName returned %+v, expected %+v", microVMs, expected)
	}
}

func TestMicroVMs_ListByName_EmptyName(t *testing.T) {
	_, _, err := (&MicroVMsServiceOp{}).ListByName(ctx, "", nil)
	if err == nil {
		t.Fatal("expected error for empty name")
	}
	if _, ok := err.(*ArgError); !ok {
		t.Errorf("expected *ArgError, got %T: %v", err, err)
	}
}

func TestMicroVMs_Get(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/microvms/aaa-111", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"microvm": {
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

	microVM, _, err := client.MicroVMs.Get(ctx, "aaa-111")
	if err != nil {
		t.Fatalf("MicroVMs.Get returned error: %v", err)
	}

	expected := &MicroVM{
		ID:         "aaa-111",
		Name:       "sandbox",
		Region:     "nyc3",
		State:      MicroVMStateRunning,
		Size:       &MicroVMSize{CPU: 2, Memory: 4096, Disk: 80},
		Networking: MicroVMNetworkingPublic,
		Source:     &MicroVMSource{OCIRef: "docker.io/library/nginx:1.27"},
		URLs: []MicroVMURL{
			{Hostname: "sandbox.example.com", Port: 8080, Default: true, Status: MicroVMURLStatusActive},
		},
		Ports:   []uint32{8080},
		Tags:    []string{"env:dev"},
		Created: "2026-07-16T10:00:00Z",
	}
	if !reflect.DeepEqual(microVM, expected) {
		t.Errorf("MicroVMs.Get returned %+v, expected %+v", microVM, expected)
	}
}

func TestMicroVMs_Get_EmptyID(t *testing.T) {
	_, _, err := (&MicroVMsServiceOp{}).Get(ctx, "")
	if err == nil {
		t.Fatal("expected error for empty id")
	}
	if _, ok := err.(*ArgError); !ok {
		t.Errorf("expected *ArgError, got %T: %v", err, err)
	}
}

func TestMicroVMs_Create(t *testing.T) {
	setup()
	defer teardown()

	autoResume := true
	autoPauseEnabled := true
	createRequest := &MicroVMCreateRequest{
		Name:         "sandbox",
		Region:       "nyc3",
		Size:         &MicroVMSizeRequest{CPU: 2, Memory: 4096},
		Source:       &MicroVMSource{OCIRef: "docker.io/library/nginx:1.27"},
		Networking:   MicroVMNetworkingVPC,
		VPCUUID:      "vpc-uuid",
		AutoPause:    &AutoPauseConfig{Enabled: &autoPauseEnabled, IdleTimeout: "5m"},
		AutoResume:   &autoResume,
		HTTPPort:     8080,
		HTTPProtocol: MicroVMHTTPProtocolHTTP2,
		Ports:        []uint32{80, 8080},
		Environment:  map[string]string{"FOO": "bar"},
		Tags:         []string{"env:dev", "team:agents"},
	}

	mux.HandleFunc("/v2/microvms", func(w http.ResponseWriter, r *http.Request) {
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

		fmt.Fprint(w, `{"microvm": {"id": "aaa-111", "name": "sandbox", "state": "creating"}}`)
	})

	microVM, _, err := client.MicroVMs.Create(ctx, createRequest)
	if err != nil {
		t.Fatalf("MicroVMs.Create returned error: %v", err)
	}

	if microVM.ID != "aaa-111" {
		t.Errorf("MicroVMs.Create returned ID %q, expected %q", microVM.ID, "aaa-111")
	}
	if microVM.State != MicroVMStateCreating {
		t.Errorf("MicroVMs.Create returned State %q, expected %q", microVM.State, MicroVMStateCreating)
	}
}

func TestMicroVMs_Create_FromCheckpoint(t *testing.T) {
	setup()
	defer teardown()

	createRequest := &MicroVMCreateRequest{
		Name:   "sandbox-clone",
		Source: &MicroVMSource{CheckpointID: "chk-1"},
	}

	mux.HandleFunc("/v2/microvms", func(w http.ResponseWriter, r *http.Request) {
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

		fmt.Fprint(w, `{"microvm": {"id": "bbb-222"}}`)
	})

	if _, _, err := client.MicroVMs.Create(ctx, createRequest); err != nil {
		t.Fatalf("MicroVMs.Create returned error: %v", err)
	}
}

func TestMicroVMs_Create_NilRequest(t *testing.T) {
	_, _, err := (&MicroVMsServiceOp{}).Create(ctx, nil)
	if err == nil {
		t.Fatal("expected error for nil createRequest")
	}
	if _, ok := err.(*ArgError); !ok {
		t.Errorf("expected *ArgError, got %T: %v", err, err)
	}
}

func TestMicroVMs_Pause(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/microvms/aaa-111/pause", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `{"microvm": {"id": "aaa-111", "state": "paused"}}`)
	})

	microVM, _, err := client.MicroVMs.Pause(ctx, "aaa-111")
	if err != nil {
		t.Fatalf("MicroVMs.Pause returned error: %v", err)
	}

	if microVM.State != MicroVMStatePaused {
		t.Errorf("MicroVMs.Pause returned State %q, expected %q", microVM.State, MicroVMStatePaused)
	}
}

func TestMicroVMs_Pause_EmptyID(t *testing.T) {
	_, _, err := (&MicroVMsServiceOp{}).Pause(ctx, "")
	if err == nil {
		t.Fatal("expected error for empty id")
	}
	if _, ok := err.(*ArgError); !ok {
		t.Errorf("expected *ArgError, got %T: %v", err, err)
	}
}

func TestMicroVMs_Resume(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/microvms/aaa-111/resume", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `{"microvm": {"id": "aaa-111", "state": "running"}}`)
	})

	microVM, _, err := client.MicroVMs.Resume(ctx, "aaa-111")
	if err != nil {
		t.Fatalf("MicroVMs.Resume returned error: %v", err)
	}

	if microVM.State != MicroVMStateRunning {
		t.Errorf("MicroVMs.Resume returned State %q, expected %q", microVM.State, MicroVMStateRunning)
	}
}

func TestMicroVMs_Resume_EmptyID(t *testing.T) {
	_, _, err := (&MicroVMsServiceOp{}).Resume(ctx, "")
	if err == nil {
		t.Fatal("expected error for empty id")
	}
	if _, ok := err.(*ArgError); !ok {
		t.Errorf("expected *ArgError, got %T: %v", err, err)
	}
}

func TestMicroVMs_Delete(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/microvms/aaa-111", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		w.WriteHeader(http.StatusNoContent)
	})

	if _, err := client.MicroVMs.Delete(ctx, "aaa-111"); err != nil {
		t.Fatalf("MicroVMs.Delete returned error: %v", err)
	}
}

func TestMicroVMs_Delete_EmptyID(t *testing.T) {
	_, err := (&MicroVMsServiceOp{}).Delete(ctx, "")
	if err == nil {
		t.Fatal("expected error for empty id")
	}
	if _, ok := err.(*ArgError); !ok {
		t.Errorf("expected *ArgError, got %T: %v", err, err)
	}
}

func TestMicroVMs_ListCheckpoints(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/microvms/checkpoints", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		if got, want := r.URL.Query().Get("microvm_id"), "aaa-111"; got != want {
			t.Errorf("microvm_id query = %q, expected %q", got, want)
		}
		fmt.Fprint(w, `{
			"checkpoints": [
				{
					"id": "chk-1",
					"microvm_id": "aaa-111",
					"microvm_name": "sandbox",
					"region": "nyc3",
					"status": "CHECKPOINT_AVAILABLE",
					"memory_bytes": 1024,
					"disk_bytes": 2048
				},
				{"id": "chk-2", "microvm_id": "aaa-111", "status": "CHECKPOINT_CREATING"}
			],
			"meta": {"total": 2}
		}`)
	})

	checkpoints, resp, err := client.MicroVMs.ListCheckpoints(ctx, &ListMicroVMCheckpointsOptions{
		MicroVMID: "aaa-111",
	})
	if err != nil {
		t.Fatalf("MicroVMs.ListCheckpoints returned error: %v", err)
	}

	expected := []MicroVMCheckpoint{
		{
			ID:          "chk-1",
			MicroVMID:   "aaa-111",
			MicroVMName: "sandbox",
			Region:      "nyc3",
			Status:      MicroVMCheckpointStatusAvailable,
			MemoryBytes: 1024,
			DiskBytes:   2048,
		},
		{ID: "chk-2", MicroVMID: "aaa-111", Status: MicroVMCheckpointStatusCreating},
	}
	if !reflect.DeepEqual(checkpoints, expected) {
		t.Errorf("MicroVMs.ListCheckpoints returned %+v, expected %+v", checkpoints, expected)
	}

	if resp.Meta == nil || resp.Meta.Total != 2 {
		t.Errorf("MicroVMs.ListCheckpoints Meta not propagated: %+v", resp.Meta)
	}
}

func TestMicroVMs_ListCheckpoints_All(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/microvms/checkpoints", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		if got := r.URL.Query().Get("microvm_id"); got != "" {
			t.Errorf("unexpected microvm_id query = %q", got)
		}
		fmt.Fprint(w, `{"checkpoints": [{"id": "chk-1"}], "meta": {"total": 1}}`)
	})

	checkpoints, _, err := client.MicroVMs.ListCheckpoints(ctx, nil)
	if err != nil {
		t.Fatalf("MicroVMs.ListCheckpoints returned error: %v", err)
	}
	if len(checkpoints) != 1 || checkpoints[0].ID != "chk-1" {
		t.Errorf("MicroVMs.ListCheckpoints returned %+v", checkpoints)
	}
}

func TestMicroVMs_CreateCheckpoint(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/microvms/aaa-111/checkpoints", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)

		var got map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		expected := map[string]interface{}{"name": "chk-named"}
		if !reflect.DeepEqual(got, expected) {
			t.Errorf("CreateCheckpoint body\n got=%#v\nwant=%#v", got, expected)
		}

		fmt.Fprint(w, `{"checkpoint": {"id": "chk-1", "microvm_id": "aaa-111", "status": "CHECKPOINT_CREATING"}}`)
	})

	checkpoint, _, err := client.MicroVMs.CreateCheckpoint(ctx, "aaa-111", &MicroVMCheckpointCreateRequest{Name: "chk-named"})
	if err != nil {
		t.Fatalf("MicroVMs.CreateCheckpoint returned error: %v", err)
	}
	if checkpoint.ID != "chk-1" || checkpoint.Status != MicroVMCheckpointStatusCreating {
		t.Errorf("MicroVMs.CreateCheckpoint returned %+v", checkpoint)
	}
}

func TestMicroVMs_CreateCheckpoint_EmptyID(t *testing.T) {
	_, _, err := (&MicroVMsServiceOp{}).CreateCheckpoint(ctx, "", nil)
	if err == nil {
		t.Fatal("expected error for empty microVMID")
	}
	if _, ok := err.(*ArgError); !ok {
		t.Errorf("expected *ArgError, got %T: %v", err, err)
	}
}

func TestMicroVMs_GetCheckpoint(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/microvms/checkpoints/chk-1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{"checkpoint": {"id": "chk-1", "status": "CHECKPOINT_AVAILABLE", "region": "nyc3"}}`)
	})

	checkpoint, _, err := client.MicroVMs.GetCheckpoint(ctx, "chk-1")
	if err != nil {
		t.Fatalf("MicroVMs.GetCheckpoint returned error: %v", err)
	}
	expected := &MicroVMCheckpoint{
		ID:     "chk-1",
		Status: MicroVMCheckpointStatusAvailable,
		Region: "nyc3",
	}
	if !reflect.DeepEqual(checkpoint, expected) {
		t.Errorf("MicroVMs.GetCheckpoint returned %+v, expected %+v", checkpoint, expected)
	}
}

func TestMicroVMs_GetCheckpoint_EmptyID(t *testing.T) {
	_, _, err := (&MicroVMsServiceOp{}).GetCheckpoint(ctx, "")
	if err == nil {
		t.Fatal("expected error for empty id")
	}
	if _, ok := err.(*ArgError); !ok {
		t.Errorf("expected *ArgError, got %T: %v", err, err)
	}
}

func TestMicroVMs_DeleteCheckpoint(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/microvms/checkpoints/chk-1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		w.WriteHeader(http.StatusNoContent)
	})

	if _, err := client.MicroVMs.DeleteCheckpoint(ctx, "chk-1"); err != nil {
		t.Fatalf("MicroVMs.DeleteCheckpoint returned error: %v", err)
	}
}

func TestMicroVMs_DeleteCheckpoint_EmptyID(t *testing.T) {
	_, err := (&MicroVMsServiceOp{}).DeleteCheckpoint(ctx, "")
	if err == nil {
		t.Fatal("expected error for empty id")
	}
	if _, ok := err.(*ArgError); !ok {
		t.Errorf("expected *ArgError, got %T: %v", err, err)
	}
}

func TestMicroVMs_GetCreateOptions(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/microvms/options", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"default_region": "nyc1",
			"sizes": [{
				"cpu": 2,
				"memory": 4096,
				"disk": 80,
				"available": true,
				"regions": ["nyc1", "sfo3"],
				"pricing": {"price_per_hour": 0.0119, "price_per_month": 8.0}
			}],
			"features": [{"name": "microvm", "enabled": true}],
			"account_limits": {"max_concurrent_running": 10, "max_total_count": 25}
		}`)
	})

	opts, _, err := client.MicroVMs.GetCreateOptions(ctx)
	if err != nil {
		t.Fatalf("MicroVMs.GetCreateOptions returned error: %v", err)
	}

	expected := &MicroVMCreateOptions{
		DefaultRegion: "nyc1",
		Sizes: []MicroVMSizeOption{{
			CPU:       2,
			Memory:    4096,
			Disk:      80,
			Available: true,
			Regions:   []string{"nyc1", "sfo3"},
			Pricing:   &MicroVMSizePricing{PricePerHour: 0.0119, PricePerMonth: 8.0},
		}},
		Features: []MicroVMFeatureOption{{Name: "microvm", Enabled: true}},
		AccountLimits: &MicroVMAccountLimits{
			MaxConcurrentRunning: 10,
			MaxTotalCount:        25,
		},
	}
	if !reflect.DeepEqual(opts, expected) {
		t.Errorf("MicroVMs.GetCreateOptions returned %+v, expected %+v", opts, expected)
	}
}

func TestMicroVM_URN(t *testing.T) {
	md := MicroVM{ID: "aaa-111"}
	want := "do:microdroplet:aaa-111"
	if got := md.URN(); got != want {
		t.Errorf("MicroVM.URN = %q, expected %q", got, want)
	}
}
