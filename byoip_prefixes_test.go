package godo

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"testing"
	"time"
)

func TestBYOIPPrefixes_List(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/byoip_prefixes", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{"byoip_prefixes": [
			{"uuid":"139efe95-c8fc-42a7-8faa-bd3afc2b0985","prefix":"192.168.0.0/24","region":"nyc3", "status": "active", "failure_reason": "", "validations": [{"name": "validation","status": "PASSED"}], "project_id": "project-123", "advertised": true, "locked": false},
			{"uuid":"e164034d-deaa-4288-b72e-dbff38103eb1","prefix":"127.0.0.0/24", "region":"nyc1", "status": "declined", "failure_reason": "not allowed local IP range", "validations": [], "project_id": "project-456", "advertised": false, "locked": false}
			],
			"meta": {"total": 2}
		}`)
	})

	byoips, resp, err := client.BYOIPPrefixes.List(ctx, nil)
	if err != nil {
		t.Errorf("BYOIPs.List returned error: %v", err)
	}

	expectedBYOIPs := []*BYOIPPrefix{
		{UUID: "139efe95-c8fc-42a7-8faa-bd3afc2b0985", Prefix: "192.168.0.0/24", Status: "active", Region: "nyc3", FailureReason: "", Validations: []any{map[string]interface{}{"name": "validation", "status": "PASSED"}}, ProjectID: "project-123", Advertised: true, Locked: false},
		{UUID: "e164034d-deaa-4288-b72e-dbff38103eb1", Prefix: "127.0.0.0/24", Status: "declined", Region: "nyc1", FailureReason: "not allowed local IP range", Validations: []any{}, ProjectID: "project-456", Advertised: false, Locked: false},
	}

	if !reflect.DeepEqual(byoips, expectedBYOIPs) {
		t.Errorf("BYOIPs.List returned %+v, expected %+v", byoips, expectedBYOIPs)
	}

	expectedMeta := &Meta{
		Total: 2,
	}

	if !reflect.DeepEqual(resp.Meta, expectedMeta) {
		t.Errorf("BYOIPs.List returned meta %+v, expected %+v", resp.Meta, expectedMeta)
	}
}

func TestBYOIPPrefixes_Get(t *testing.T) {
	setup()
	defer teardown()
	mux.HandleFunc("/v2/byoip_prefixes/1de94988-5102-4aae-b17d-f71b98707b88", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{"byoip_prefix": {"uuid":"1de94988-5102-4aae-b17d-f71b98707b88","prefix":"192.168.0.0/24","region":"nyc3", "status": "active", "failure_reason": "", "validations": [{"name": "validation","status": "PASSED"}], "project_id": "project-abc", "advertised": true, "locked": false}}`)
	})

	byoipPrefix, _, err := client.BYOIPPrefixes.Get(ctx, "1de94988-5102-4aae-b17d-f71b98707b88")
	if err != nil {
		t.Errorf("BYOIPs.Get returned error: %v", err)
	}

	expected := &byoipPrefixRoot{BYOIPPrefix: &BYOIPPrefix{UUID: "1de94988-5102-4aae-b17d-f71b98707b88", Prefix: "192.168.0.0/24", Status: "active", Region: "nyc3", FailureReason: "", Validations: []any{map[string]any{"name": "validation", "status": "PASSED"}}, ProjectID: "project-abc", Advertised: true, Locked: false}}

	if !reflect.DeepEqual(byoipPrefix, expected.BYOIPPrefix) {
		t.Errorf("BYOIPs.Get returned %+v, expected %+v", byoipPrefix, expected)
	}
}

func TestBYOIPPrefixes_GetResources(t *testing.T) {
	setup()
	defer teardown()
	mux.HandleFunc("/v2/byoip_prefixes/1de94988-5102-4aae-b17d-f71b98707b88/ips", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{"ips": [
			{"id":1,"byoip":"192.168.0.1","region":"nyc3","resource": "do:droplet:a3ec41f4-84f4-44d2-a4ff-27165b957cdc", "assigned_at": "2025-03-14T00:00:01.000Z"},
			{"id":4,"byoip":"192.168.0.10","region":"nyc3","resource": "do:droplet:81725c6d-97f7-4ffe-9129-d2e4890a0800", "assigned_at": "2025-03-15T00:00:02.000Z"}
			],
			"meta": {"total": 2}}`)
	})

	resources, resp, err := client.BYOIPPrefixes.GetResources(ctx, "1de94988-5102-4aae-b17d-f71b98707b88", nil)
	if err != nil {
		t.Errorf("BYOIPs.GetResources returned error: %v", err)
	}

	expectedResources := []BYOIPPrefixResource{
		{ID: 1, BYOIP: "192.168.0.1", Region: "nyc3", Resource: "do:droplet:a3ec41f4-84f4-44d2-a4ff-27165b957cdc", AssignedAt: time.Date(2025, 3, 14, 0, 0, 1, 0, time.UTC)},
		{ID: 4, BYOIP: "192.168.0.10", Region: "nyc3", Resource: "do:droplet:81725c6d-97f7-4ffe-9129-d2e4890a0800", AssignedAt: time.Date(2025, 3, 15, 0, 0, 2, 0, time.UTC)},
	}

	if !reflect.DeepEqual(resources, expectedResources) {
		t.Errorf("BYOIPs.GetResources returned %+v, expected %+v", resources, expectedResources)
	}

	expectedMeta := &Meta{
		Total: 2,
	}

	if !reflect.DeepEqual(resp.Meta, expectedMeta) {
		t.Errorf("BYOIPs.GetResources returned meta %+v, expected %+v", resp.Meta, expectedMeta)
	}
}

func TestBYOIPPrefixes_Create(t *testing.T) {
	setup()
	defer teardown()

	byoipCR := &BYOIPPrefixCreateReq{
		Prefix:    "10.10.10.10/24",
		Signature: "signature",
		Region:    "nyc3",
	}

	mux.HandleFunc("/v2/byoip_prefixes", func(w http.ResponseWriter, r *http.Request) {

		v := new(BYOIPPrefixCreateReq)
		err := json.NewDecoder(r.Body).Decode(v)
		if err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(v, byoipCR) {
			t.Errorf("Request body = %+v, expected %+v", v, byoipCR)
		}

		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `{"byoip_prefix": {"uuid": "byoip-prefix-uuid", "region": "nyc3", "status": "pending"}}`)
	})

	byoipCreated, _, err := client.BYOIPPrefixes.Create(ctx, byoipCR)
	if err != nil {
		t.Errorf("BYOIPs.Create returned error: %v", err)
	}

	expectedBYOIP := &BYOIPPrefixCreateResp{
		UUID:   "byoip-prefix-uuid",
		Region: "nyc3",
		Status: "pending",
	}

	if !reflect.DeepEqual(byoipCreated, expectedBYOIP) {
		t.Errorf("BYOIPs.Create returned %+v, expected %+v", byoipCreated, expectedBYOIP)
	}

}

func TestBYOIPPrefixes_Delete(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/byoip_prefixes/1de94988-5102-4aae-b17d-f71b98707b88", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		w.WriteHeader(http.StatusAccepted)
	})

	_, err := client.BYOIPPrefixes.Delete(ctx, "1de94988-5102-4aae-b17d-f71b98707b88")
	if err != nil {
		t.Errorf("BYOIPs.Delete returned error: %v", err)
	}

}

func TestBYOIPPrefixes_Update(t *testing.T) {
	setup()
	defer teardown()

	updateReq := &BYOIPPrefixUpdateReq{
		Advertise: PtrTo(true),
	}

	mux.HandleFunc("/v2/byoip_prefixes/test-uuid-123", func(w http.ResponseWriter, r *http.Request) {
		v := new(BYOIPPrefixUpdateReq)
		err := json.NewDecoder(r.Body).Decode(v)
		if err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(v, updateReq) {
			t.Errorf("Request body = %+v, expected %+v", v, updateReq)
		}

		testMethod(t, r, http.MethodPatch)
		fmt.Fprint(w, `{"byoip_prefix": {
			"uuid":"test-uuid-123",
			"prefix":"192.168.1.0/24",
			"region":"nyc3", 
			"status": "active", 
			"failure_reason": "", 
			"validations": [],
			"project_id": "project-update-123", 
			"advertised": true, 
			"locked": false
		}}`)
	})

	byoipPrefix, _, err := client.BYOIPPrefixes.Update(ctx, "test-uuid-123", updateReq)
	if err != nil {
		t.Errorf("BYOIPs.Update returned error: %v", err)
	}

	expected := &BYOIPPrefix{
		UUID:          "test-uuid-123",
		Prefix:        "192.168.1.0/24",
		Status:        "active",
		Region:        "nyc3",
		FailureReason: "",
		Validations:   []any{},
		ProjectID:     "project-update-123",
		Advertised:    true,
		Locked:        false,
	}

	if !reflect.DeepEqual(byoipPrefix, expected) {
		t.Errorf("BYOIPs.Update returned %+v, expected %+v", byoipPrefix, expected)
	}

	// Verify the advertised field was updated
	if !byoipPrefix.Advertised {
		t.Error("Expected Advertised to be true after update, got false")
	}
}
