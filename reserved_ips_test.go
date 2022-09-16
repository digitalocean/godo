package godo

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

func TestReservedIPs_ListReservedIPs(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/reserved_ips", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{"reserved_ips": [
			{"region":{"slug":"nyc3"},"droplet":{"id":1},"ip":"192.168.0.1","project_id":"46d8977a-35cd-11ed-909f-43c99bbf6032", "locked":false},
			{"region":{"slug":"nyc3"},"droplet":{"id":2},"ip":"192.168.0.2","project_id":"46d8977a-35cd-11ed-909f-43c99bbf6032", "locked":false}],
			"meta":{"total":2}
		}`)
	})

	reservedIPs, resp, err := client.ReservedIPs.List(ctx, nil)
	if err != nil {
		t.Errorf("ReservedIPs.List returned error: %v", err)
	}

	expectedReservedIPs := []ReservedIP{
		{Region: &Region{Slug: "nyc3"}, Droplet: &Droplet{ID: 1}, IP: "192.168.0.1", Locked: false, ProjectID: "46d8977a-35cd-11ed-909f-43c99bbf6032"},
		{Region: &Region{Slug: "nyc3"}, Droplet: &Droplet{ID: 2}, IP: "192.168.0.2", Locked: false, ProjectID: "46d8977a-35cd-11ed-909f-43c99bbf6032"},
	}
	if !reflect.DeepEqual(reservedIPs, expectedReservedIPs) {
		t.Errorf("ReservedIPs.List returned reserved IPs %+v, expected %+v", reservedIPs, expectedReservedIPs)
	}

	expectedMeta := &Meta{
		Total: 2,
	}
	if !reflect.DeepEqual(resp.Meta, expectedMeta) {
		t.Errorf("ReservedIPs.List returned meta %+v, expected %+v", resp.Meta, expectedMeta)
	}
}

func TestReservedIPs_ListReservedIPsMultiplePages(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/reserved_ips", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{"reserved_ips": [
			{"region":{"slug":"nyc3"},"droplet":{"id":1},"ip":"192.168.0.1","project_id":"46d8977a-35cd-11ed-909f-43c99bbf6032", "locked":false},
			{"region":{"slug":"nyc3"},"droplet":{"id":2},"ip":"192.168.0.2","project_id":"46d8977a-35cd-11ed-909f-43c99bbf6032", "locked":false}],
			"links":{"pages":{"next":"http://example.com/v2/reserved_ips/?page=2"}}}
		`)
	})

	_, resp, err := client.ReservedIPs.List(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}

	checkCurrentPage(t, resp, 1)
}

func TestReservedIPs_RetrievePageByNumber(t *testing.T) {
	setup()
	defer teardown()

	jBlob := `
	{
		"reserved_ips": [
			{"region":{"slug":"nyc3"},"droplet":{"id":1},"ip":"192.168.0.1","project_id":"46d8977a-35cd-11ed-909f-43c99bbf6032", "locked":false},
			{"region":{"slug":"nyc3"},"droplet":{"id":2},"ip":"192.168.0.2","project_id":"46d8977a-35cd-11ed-909f-43c99bbf6032", "locked":false}],
		"links":{
			"pages":{
				"next":"http://example.com/v2/reserved_ips/?page=3",
				"prev":"http://example.com/v2/reserved_ips/?page=1",
				"last":"http://example.com/v2/reserved_ips/?page=3",
				"first":"http://example.com/v2/reserved_ips/?page=1"
			}
		}
	}`

	mux.HandleFunc("/v2/reserved_ips", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, jBlob)
	})

	opt := &ListOptions{Page: 2}
	_, resp, err := client.ReservedIPs.List(ctx, opt)
	if err != nil {
		t.Fatal(err)
	}

	checkCurrentPage(t, resp, 2)
}

func TestReservedIPs_Get(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/reserved_ips/192.168.0.1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{"reserved_ip":{"region":{"slug":"nyc3"},"droplet":{"id":1},"ip":"192.168.0.1","project_id":"46d8977a-35cd-11ed-909f-43c99bbf6032", "locked":false}}`)
	})

	reservedIP, _, err := client.ReservedIPs.Get(ctx, "192.168.0.1")
	if err != nil {
		t.Errorf("domain.Get returned error: %v", err)
	}

	expected := &ReservedIP{Region: &Region{Slug: "nyc3"}, Droplet: &Droplet{ID: 1}, IP: "192.168.0.1", Locked: false, ProjectID: "46d8977a-35cd-11ed-909f-43c99bbf6032"}
	if !reflect.DeepEqual(reservedIP, expected) {
		t.Errorf("ReservedIPs.Get returned %+v, expected %+v", reservedIP, expected)
	}
}

func TestReservedIPs_Create(t *testing.T) {
	setup()
	defer teardown()

	createRequest := &ReservedIPCreateRequest{
		Region:    "nyc3",
		DropletID: 1,
		ProjectID: "46d8977a-35cd-11ed-909f-43c99bbf6032",
	}

	mux.HandleFunc("/v2/reserved_ips", func(w http.ResponseWriter, r *http.Request) {
		v := new(ReservedIPCreateRequest)
		err := json.NewDecoder(r.Body).Decode(v)
		if err != nil {
			t.Fatal(err)
		}

		testMethod(t, r, http.MethodPost)
		if !reflect.DeepEqual(v, createRequest) {
			t.Errorf("Request body = %+v, expected %+v", v, createRequest)
		}

		fmt.Fprint(w, `{"reserved_ip":{"region":{"slug":"nyc3"},"droplet":{"id":1},"ip":"192.168.0.1","project_id":"46d8977a-35cd-11ed-909f-43c99bbf6032", "locked":false}}`)
	})

	reservedIP, _, err := client.ReservedIPs.Create(ctx, createRequest)
	if err != nil {
		t.Errorf("ReservedIPs.Create returned error: %v", err)
	}

	expected := &ReservedIP{Region: &Region{Slug: "nyc3"}, Droplet: &Droplet{ID: 1}, IP: "192.168.0.1", Locked: false, ProjectID: "46d8977a-35cd-11ed-909f-43c99bbf6032"}
	if !reflect.DeepEqual(reservedIP, expected) {
		t.Errorf("ReservedIPs.Create returned %+v, expected %+v", reservedIP, expected)
	}
}

func TestReservedIPs_Destroy(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/reserved_ips/192.168.0.1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
	})

	_, err := client.ReservedIPs.Delete(ctx, "192.168.0.1")
	if err != nil {
		t.Errorf("ReservedIPs.Delete returned error: %v", err)
	}
}
