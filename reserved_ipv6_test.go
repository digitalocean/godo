package godo

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"testing"
	"time"
)

func TestReservedIPV6s_Create(t *testing.T) {
	setup()
	defer teardown()

	reserveRequest := &ReservedIPV6CreateRequest{
		Region: "nyc3",
	}
	nowTime := time.Now()

	mux.HandleFunc("/v2/reserved_ipv6", func(w http.ResponseWriter, r *http.Request) {
		v := new(ReservedIPV6CreateRequest)
		err := json.NewDecoder(r.Body).Decode(v)
		if err != nil {
			t.Fatal(err)
		}

		testMethod(t, r, http.MethodPost)
		if !reflect.DeepEqual(v, reserveRequest) {
			t.Errorf("Request body = %+v, expected %+v", v, reserveRequest)
		}

		fmt.Fprint(w, `{"reserved_ipv6":{"ip":"2604:a880:800:14::42c3:d000","region_slug":"nyc3","reserved_at":"`+nowTime.Format(time.RFC3339Nano)+`"}}`)
	})

	reservedIP, _, err := client.ReservedIPV6s.Create(ctx, reserveRequest)
	if err != nil {
		t.Errorf("ReservedIPV6s.Create returned error: %v", err)
	}

	expected := &ReservedIPV6Resp{ReservedIPV6: &ReservedIPV6{RegionSlug: "nyc3", IP: "2604:a880:800:14::42c3:d000", ReservedAt: nowTime}}

	if !equalReserveIPv6Objects(reservedIP.ReservedIPV6, expected.ReservedIPV6) {
		t.Errorf("ReservedIPV6s.Create returned %+v, expected %+v", reservedIP, expected)
	}
}

func equalReserveIPv6Objects(a, b *ReservedIPV6) bool {
	return a.IP == b.IP &&
		a.RegionSlug == b.RegionSlug &&
		a.ReservedAt.Equal(b.ReservedAt) &&
		reflect.DeepEqual(a.Droplet, b.Droplet)
}

func TestReservedIPV6s_List(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/reserved_ipv6", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{"reserved_ipv6s": [
			{"region_slug":"nyc3","droplet":{"id":1},"ip":"2604:a880:800:14::42c3:d000"},
			{"region_slug":"nyc3","droplet":{"id":2},"ip":"2604:a880:800:14::42c3:d001"}
			],
			"meta": {"total": 2}
		}`)
	})

	reservedIPs, resp, err := client.ReservedIPV6s.List(ctx, nil)
	if err != nil {
		t.Errorf("ReservedIPs.List returned error: %v", err)
	}

	expectedReservedIPs := []ReservedIPV6{
		{RegionSlug: "nyc3", Droplet: &Droplet{ID: 1}, IP: "2604:a880:800:14::42c3:d000"},
		{RegionSlug: "nyc3", Droplet: &Droplet{ID: 2}, IP: "2604:a880:800:14::42c3:d001"},
	}

	if !reflect.DeepEqual(reservedIPs, expectedReservedIPs) {
		t.Errorf("ReservedIPV6s.List returned reserved IPs %+v, expected %+v", reservedIPs, expectedReservedIPs)
	}

	expectedMeta := &Meta{
		Total: 2,
	}
	if !reflect.DeepEqual(resp.Meta, expectedMeta) {
		t.Errorf("ReservedIPV6s.List returned meta %+v, expected %+v", resp.Meta, expectedMeta)
	}
}

func TestReservedIPV6s_ListReservedIPsMultiplePages(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/reserved_ipv6", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{"reserved_ipv6s": [
			{"region_slug":"nyc3","droplet":{"id":1},"ip":"2604:a880:800:14::42c3:d001"},
			{"region":{"slug":"nyc3"},"droplet":{"id":2},"ip":"2604:a880:800:14::42c3:d002"}],
			"links":{"pages":{"next":"http://example.com/v2/reserved_ipv6/?page=2"}}}
		`)
	})

	_, resp, err := client.ReservedIPV6s.List(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}

	checkCurrentPage(t, resp, 1)
}

func TestReservedIPV6s_Get(t *testing.T) {
	setup()
	defer teardown()
	nowTime := time.Now()
	mux.HandleFunc("/v2/reserved_ipv6/2604:a880:800:14::42c3:d001", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{"reserved_ipv6":{"region_slug":"nyc3","droplet":{"id":1},"ip":"2604:a880:800:14::42c3:d001", "reserved_at":"`+nowTime.Format(time.RFC3339Nano)+`"}}`)
	})

	reservedIP, _, err := client.ReservedIPV6s.Get(ctx, "2604:a880:800:14::42c3:d001")
	if err != nil {
		t.Errorf("ReservedIPV6s.Get returned error: %v", err)
	}

	expected := &ReservedIPV6Resp{ReservedIPV6: &ReservedIPV6{RegionSlug: "nyc3", Droplet: &Droplet{ID: 1}, IP: "2604:a880:800:14::42c3:d001", ReservedAt: nowTime}}
	if !equalReserveIPv6Objects(reservedIP.ReservedIPV6, expected.ReservedIPV6) {
		t.Errorf("ReservedIPV6s.Get returned %+v, expected %+v", reservedIP, expected)
	}
}

func TestReservedIPV6s_Delete(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/reserved_ipv6/2604:a880:800:14::42c3:d001", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
	})

	_, err := client.ReservedIPV6s.Delete(ctx, "2604:a880:800:14::42c3:d001")
	if err != nil {
		t.Errorf("ReservedIPV6s.Release returned error: %v", err)
	}
}
