package godo

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

func TestFloatingIPs_ListFloatingIPs(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/floating_ips", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{"floating_ips": [
			{"region":{"slug":"nyc3"},"droplet":{"id":1},"ip":"192.168.0.1","project_id":"46d8977a-35cd-11ed-909f-43c99bbf6032", "locked":false},
			{"region":{"slug":"nyc3"},"droplet":{"id":2},"ip":"192.168.0.2","project_id":"46d8977a-35cd-11ed-909f-43c99bbf6032", "locked":false}],
			"meta":{"total":2}
		}`)
	})

	floatingIPs, resp, err := client.FloatingIPs.List(ctx, nil)
	if err != nil {
		t.Errorf("FloatingIPs.List returned error: %v", err)
	}

	expectedFloatingIPs := []FloatingIP{
		{Region: &Region{Slug: "nyc3"}, Droplet: &Droplet{ID: 1}, IP: "192.168.0.1", Locked: false, ProjectID: "46d8977a-35cd-11ed-909f-43c99bbf6032"},
		{Region: &Region{Slug: "nyc3"}, Droplet: &Droplet{ID: 2}, IP: "192.168.0.2", Locked: false, ProjectID: "46d8977a-35cd-11ed-909f-43c99bbf6032"},
	}
	if !reflect.DeepEqual(floatingIPs, expectedFloatingIPs) {
		t.Errorf("FloatingIPs.List returned floating IPs %+v, expected %+v", floatingIPs, expectedFloatingIPs)
	}

	expectedMeta := &Meta{
		Total: 2,
	}
	if !reflect.DeepEqual(resp.Meta, expectedMeta) {
		t.Errorf("FloatingIPs.List returned meta %+v, expected %+v", resp.Meta, expectedMeta)
	}
}

func TestFloatingIPs_ListFloatingIPsMultiplePages(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/floating_ips", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{"floating_ips": [
			{"region":{"slug":"nyc3"},"droplet":{"id":1},"ip":"192.168.0.1","project_id":"46d8977a-35cd-11ed-909f-43c99bbf6032", "locked":false},
			{"region":{"slug":"nyc3"},"droplet":{"id":2},"ip":"192.168.0.2","project_id":"46d8977a-35cd-11ed-909f-43c99bbf6032", "locked":false}],
			"links":{"pages":{"next":"http://example.com/v2/floating_ips/?page=2"}}}
		`)
	})

	_, resp, err := client.FloatingIPs.List(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}

	checkCurrentPage(t, resp, 1)
}

func TestFloatingIPs_RetrievePageByNumber(t *testing.T) {
	setup()
	defer teardown()

	jBlob := `
	{
		"floating_ips": [
			{"region":{"slug":"nyc3"},"droplet":{"id":1},"ip":"192.168.0.1","project_id":"46d8977a-35cd-11ed-909f-43c99bbf6032", "locked":false},
			{"region":{"slug":"nyc3"},"droplet":{"id":2},"ip":"192.168.0.2","project_id":"46d8977a-35cd-11ed-909f-43c99bbf6032", "locked":false}],
		"links":{
			"pages":{
				"next":"http://example.com/v2/floating_ips/?page=3",
				"prev":"http://example.com/v2/floating_ips/?page=1",
				"last":"http://example.com/v2/floating_ips/?page=3",
				"first":"http://example.com/v2/floating_ips/?page=1"
			}
		}
	}`

	mux.HandleFunc("/v2/floating_ips", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, jBlob)
	})

	opt := &ListOptions{Page: 2}
	_, resp, err := client.FloatingIPs.List(ctx, opt)
	if err != nil {
		t.Fatal(err)
	}

	checkCurrentPage(t, resp, 2)
}

func TestFloatingIPs_Get(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/floating_ips/192.168.0.1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{"floating_ip":{"region":{"slug":"nyc3"},"droplet":{"id":1},"ip":"192.168.0.1","project_id":"46d8977a-35cd-11ed-909f-43c99bbf6032", "locked":false}}`)
	})

	floatingIP, _, err := client.FloatingIPs.Get(ctx, "192.168.0.1")
	if err != nil {
		t.Errorf("domain.Get returned error: %v", err)
	}

	expected := &FloatingIP{Region: &Region{Slug: "nyc3"}, Droplet: &Droplet{ID: 1}, IP: "192.168.0.1", Locked: false, ProjectID: "46d8977a-35cd-11ed-909f-43c99bbf6032"}
	if !reflect.DeepEqual(floatingIP, expected) {
		t.Errorf("FloatingIPs.Get returned %+v, expected %+v", floatingIP, expected)
	}
}

func TestFloatingIPs_Create(t *testing.T) {
	setup()
	defer teardown()

	createRequest := &FloatingIPCreateRequest{
		Region:    "nyc3",
		DropletID: 1,
		ProjectID: "46d8977a-35cd-11ed-909f-43c99bbf6032",
	}

	mux.HandleFunc("/v2/floating_ips", func(w http.ResponseWriter, r *http.Request) {
		v := new(FloatingIPCreateRequest)
		err := json.NewDecoder(r.Body).Decode(v)
		if err != nil {
			t.Fatal(err)
		}

		testMethod(t, r, http.MethodPost)
		if !reflect.DeepEqual(v, createRequest) {
			t.Errorf("Request body = %+v, expected %+v", v, createRequest)
		}

		fmt.Fprint(w, `{"floating_ip":{"region":{"slug":"nyc3"},"droplet":{"id":1},"ip":"192.168.0.1","project_id":"46d8977a-35cd-11ed-909f-43c99bbf6032", "locked":false}}`)
	})

	floatingIP, _, err := client.FloatingIPs.Create(ctx, createRequest)
	if err != nil {
		t.Errorf("FloatingIPs.Create returned error: %v", err)
	}

	expected := &FloatingIP{Region: &Region{Slug: "nyc3"}, Droplet: &Droplet{ID: 1}, IP: "192.168.0.1", Locked: false, ProjectID: "46d8977a-35cd-11ed-909f-43c99bbf6032"}
	if !reflect.DeepEqual(floatingIP, expected) {
		t.Errorf("FloatingIPs.Create returned %+v, expected %+v", floatingIP, expected)
	}
}

func TestFloatingIPs_Destroy(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/floating_ips/192.168.0.1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
	})

	_, err := client.FloatingIPs.Delete(ctx, "192.168.0.1")
	if err != nil {
		t.Errorf("FloatingIPs.Delete returned error: %v", err)
	}
}
