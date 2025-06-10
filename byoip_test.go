package godo

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"
	"time"
)

func TestBYOIPs_List(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/byoip_prefixes", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{"byoips": [
			{"uuid":"139efe95-c8fc-42a7-8faa-bd3afc2b0985","cidr":"192.168.0.0/24","region":"nyc3", "status": "active", "failure_reason": "", "validations": [{"name": "validation","status": "PASSED"}]},
			{"uuid":"e164034d-deaa-4288-b72e-dbff38103eb1","cidr":"127.0.0.0/24", "region":"nyc1", "status": "declined", "failure_reason": "not allowed local IP range", "validations": []}
			],
			"meta": {"total": 2}
		}`)
	})

	byoips, resp, err := client.BYOIPs.List(ctx, nil)
	if err != nil {
		t.Errorf("BYOIPs.List returned error: %v", err)
	}

	expectedBYOIPs := []BYOIP{
		{UUID: "139efe95-c8fc-42a7-8faa-bd3afc2b0985", Cidr: "192.168.0.0/24", Status: "active", RegionSlug: "nyc3", FailureReason: "", Validations: []map[string]interface{}{{"name": "validation", "status": "PASSED"}}},
		{UUID: "e164034d-deaa-4288-b72e-dbff38103eb1", Cidr: "127.0.0.0/24", Status: "declined", RegionSlug: "nyc1", FailureReason: "not allowed local IP range", Validations: []map[string]interface{}{}},
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

func TestBYOIPs_Get(t *testing.T) {
	setup()
	defer teardown()
	mux.HandleFunc("/v2/byoip_prefixes/1de94988-5102-4aae-b17d-f71b98707b88", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{"byoip": {"uuid":"1de94988-5102-4aae-b17d-f71b98707b88","cidr":"192.168.0.0/24","region":"nyc3", "status": "active", "failure_reason": "", "validations": [{"name": "validation","status": "PASSED"}]}}`)
	})

	byoip, _, err := client.BYOIPs.Get(ctx, "1de94988-5102-4aae-b17d-f71b98707b88")
	if err != nil {
		t.Errorf("BYOIPs.Get returned error: %v", err)
	}

	expected := &byoipRoot{BYOIP: &BYOIP{UUID: "1de94988-5102-4aae-b17d-f71b98707b88", Cidr: "192.168.0.0/24", Status: "active", RegionSlug: "nyc3", FailureReason: "", Validations: []map[string]interface{}{{"name": "validation", "status": "PASSED"}}}}

	if !reflect.DeepEqual(byoip, expected.BYOIP) {
		t.Errorf("BYOIPs.Get returned %+v, expected %+v", byoip, expected)
	}
}

func TestBYOIPs_GetResources(t *testing.T) {
	setup()
	defer teardown()
	mux.HandleFunc("/v2/byoip_prefixes/1de94988-5102-4aae-b17d-f71b98707b88/ips", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{"ips": [
			{"id":1,"byoip":"192.168.0.1","region":"nyc3","resource": "do:droplet:a3ec41f4-84f4-44d2-a4ff-27165b957cdc", "assigned_at": "2025-03-14T00:00:01.000Z"},
			{"id":4,"byoip":"192.168.0.10","region":"nyc3","resource": "do:droplet:81725c6d-97f7-4ffe-9129-d2e4890a0800", "assigned_at": "2025-03-15T00:00:02.000Z"}
			]}`)
	})

	resources, _, err := client.BYOIPs.GetResources(ctx, "1de94988-5102-4aae-b17d-f71b98707b88")
	if err != nil {
		t.Errorf("BYOIPs.GetResources returned error: %v", err)
	}

	expectedResources := []BYOIPResource{
		{ID: 1, BYOIP: "192.168.0.1", RegionSlug: "nyc3", Resource: "do:droplet:a3ec41f4-84f4-44d2-a4ff-27165b957cdc", AssignedAt: time.Date(2025, 3, 14, 0, 0, 1, 0, time.UTC)},
		{ID: 4, BYOIP: "192.168.0.10", RegionSlug: "nyc3", Resource: "do:droplet:81725c6d-97f7-4ffe-9129-d2e4890a0800", AssignedAt: time.Date(2025, 3, 15, 0, 0, 2, 0, time.UTC)},
	}

	if !reflect.DeepEqual(resources, expectedResources) {
		t.Errorf("BYOIPs.GetResources returned %+v, expected %+v", resources, expectedResources)
	}
}
