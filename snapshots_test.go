package godo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

func TestSnapshots_List(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/snapshots", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{"snapshots":[{"id":"1"},{"id":"2", "size_gigabytes": 4.84}]}`)
	})
	ctx := context.Background()
	snapshots, _, err := client.Snapshots.List(ctx, nil)
	if err != nil {
		t.Errorf("Snapshots.List returned error: %v", err)
	}

	expected := []Snapshot{{ID: "1"}, {ID: "2", SizeGigaBytes: 4.84}}
	if !reflect.DeepEqual(snapshots, expected) {
		t.Errorf("Snapshots.List returned %+v, expected %+v", snapshots, expected)
	}
}

func TestSnapshots_ListVolume(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/snapshots", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		expected := "volume"
		actual := r.URL.Query().Get("resource_type")
		if actual != expected {
			t.Errorf("'type' query = %v, expected %v", actual, expected)
		}
		fmt.Fprint(w, `{"snapshots":[{"id":"1"},{"id":"2"}]}`)
	})

	ctx := context.Background()
	snapshots, _, err := client.Snapshots.ListVolume(ctx, nil)
	if err != nil {
		t.Errorf("Snapshots.ListVolume returned error: %v", err)
	}

	expected := []Snapshot{{ID: "1"}, {ID: "2"}}
	if !reflect.DeepEqual(snapshots, expected) {
		t.Errorf("Snapshots.ListVolume returned %+v, expected %+v", snapshots, expected)
	}
}

func TestSnapshots_ListVolume_Regional(t *testing.T) {
	setup()
	defer teardown()

	s2r1Snapshots := []Snapshot{
		{
			ID:      "1",
			Name:    "snapshot-1",
			Regions: []string{"s2r1"},
		},
		{
			ID:      "2",
			Name:    "snapshot-2",
			Regions: []string{"s2r1"},
		},
	}

	s2r6Snapshots := []Snapshot{
		{
			ID:      "3",
			Regions: []string{"s2r6"},
		},
	}

	mux.HandleFunc("/v2/snapshots", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		expected := "volume"
		actual := r.URL.Query().Get("resource_type")
		if actual != expected {
			t.Errorf("'type' query = %v, expected %v", actual, expected)
		}

		var resultSnapshot []Snapshot

		region := r.URL.Query().Get("region")
		if region == "" {
			resultSnapshot = append(resultSnapshot, s2r1Snapshots...)
			resultSnapshot = append(resultSnapshot, s2r6Snapshots...)
		} else if region == "s2r6" {
			resultSnapshot = append(resultSnapshot, s2r6Snapshots...)
		} else if region == "s2r1" {
			resultSnapshot = append(resultSnapshot, s2r1Snapshots...)
		}

		b, _ := json.Marshal(resultSnapshot)
		fmt.Fprint(w, fmt.Sprintf(`{"snapshots": %s}`, string(b)))
	})

	ctx := context.Background()
	snapshots, _, err := client.Snapshots.ListVolume(ctx, &ListOptions{Region: "s2r6"})
	if err != nil {
		t.Errorf("Snapshots.ListVolume returned error: %v", err)
	}

	if !reflect.DeepEqual(snapshots, s2r6Snapshots) {
		t.Errorf("Snapshots.ListVolume returned %+v, expected %+v", snapshots, s2r6Snapshots)
	}

	snapshots, _, err = client.Snapshots.ListVolume(ctx, &ListOptions{Region: "s2r1"})
	if err != nil {
		t.Errorf("Snapshots.ListVolume returned error: %v", err)
	}

	if !reflect.DeepEqual(snapshots, s2r1Snapshots) {
		t.Errorf("Snapshots.ListVolume returned %+v, expected %+v", snapshots, s2r1Snapshots)
	}

	snapshots, _, err = client.Snapshots.ListVolume(ctx, nil)
	if err != nil {
		t.Errorf("Snapshots.ListVolume returned error: %v", err)
	}

	// Without region all snapshots should be returned
	allSnapshots := append([]Snapshot{}, s2r1Snapshots...)
	allSnapshots = append(allSnapshots, s2r6Snapshots...)
	if !reflect.DeepEqual(snapshots, allSnapshots) {
		t.Errorf("Snapshots.ListVolume returned %+v, expected %+v", snapshots, allSnapshots)
	}
}

func TestSnapshots_ListDroplet(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/snapshots", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		expected := "droplet"
		actual := r.URL.Query().Get("resource_type")
		if actual != expected {
			t.Errorf("'resource_type' query = %v, expected %v", actual, expected)
		}

		fmt.Fprint(w, `{"snapshots":[{"id":"1"},{"id":"2", "size_gigabytes": 4.84}]}`)
	})

	ctx := context.Background()
	snapshots, _, err := client.Snapshots.ListDroplet(ctx, nil)
	if err != nil {
		t.Errorf("Snapshots.ListDroplet returned error: %v", err)
	}

	expected := []Snapshot{{ID: "1"}, {ID: "2", SizeGigaBytes: 4.84}}
	if !reflect.DeepEqual(snapshots, expected) {
		t.Errorf("Snapshots.ListDroplet returned %+v, expected %+v", snapshots, expected)
	}
}

func TestSnapshots_ListSnapshotsMultiplePages(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/snapshots", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{"snapshots": [{"id":"1"},{"id":"2"}], "links":{"pages":{"next":"http://example.com/v2/snapshots/?page=2"}}}`)
	})

	ctx := context.Background()
	_, resp, err := client.Snapshots.List(ctx, &ListOptions{Page: 2})
	if err != nil {
		t.Fatal(err)
	}
	checkCurrentPage(t, resp, 1)
}

func TestSnapshots_RetrievePageByNumber(t *testing.T) {
	setup()
	defer teardown()

	jBlob := `
    {
        "snapshots": [{"id":"1"},{"id":"2"}],
        "links":{
            "pages":{
                "next":"http://example.com/v2/snapshots/?page=3",
                "prev":"http://example.com/v2/snapshots/?page=1",
                "last":"http://example.com/v2/snapshots/?page=3",
                "first":"http://example.com/v2/snapshots/?page=1"
            }
        }
    }`

	mux.HandleFunc("/v2/snapshots", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, jBlob)
	})

	ctx := context.Background()
	opt := &ListOptions{Page: 2}
	_, resp, err := client.Snapshots.List(ctx, opt)
	if err != nil {
		t.Fatal(err)
	}

	checkCurrentPage(t, resp, 2)
}

func TestSnapshots_GetSnapshotByID(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/snapshots/12345", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{"snapshot":{"id":"12345"}}`)
	})

	ctx := context.Background()
	snapshots, _, err := client.Snapshots.Get(ctx, "12345")
	if err != nil {
		t.Errorf("Snapshot.GetByID returned error: %v", err)
	}

	expected := &Snapshot{ID: "12345"}
	if !reflect.DeepEqual(snapshots, expected) {
		t.Errorf("Snapshots.GetByID returned %+v, expected %+v", snapshots, expected)
	}
}

func TestSnapshots_Destroy(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/snapshots/12345", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
	})

	ctx := context.Background()
	_, err := client.Snapshots.Delete(ctx, "12345")
	if err != nil {
		t.Errorf("Snapshot.Delete returned error: %v", err)
	}
}

func TestSnapshot_String(t *testing.T) {
	snapshot := &Snapshot{
		ID:            "1",
		Name:          "Snapsh176ot",
		ResourceID:    "0",
		ResourceType:  "droplet",
		Regions:       []string{"one"},
		MinDiskSize:   20,
		SizeGigaBytes: 4.84,
		Created:       "2013-11-27T09:24:55Z",
		Tags:          []string{"one", "two"},
	}

	stringified := snapshot.String()
	expected := `godo.Snapshot{ID:"1", Name:"Snapsh176ot", ResourceID:"0", ResourceType:"droplet", Regions:["one"], MinDiskSize:20, SizeGigaBytes:4.84, Created:"2013-11-27T09:24:55Z", Tags:["one" "two"]}`
	if expected != stringified {
		t.Errorf("Snapshot.String returned %+v, expected %+v", stringified, expected)
	}
}
