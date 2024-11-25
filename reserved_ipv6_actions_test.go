package godo

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

func TestReservedIPV6sActions_Assign(t *testing.T) {
	setup()
	defer teardown()
	dropletID := 12345
	assignRequest := &ActionRequest{
		"droplet_id": float64(dropletID),
		"type":       "assign",
	}

	mux.HandleFunc("/v2/reserved_ipv6/2604:a880:800:14::42c3:d000/actions", func(w http.ResponseWriter, r *http.Request) {
		v := new(ActionRequest)
		err := json.NewDecoder(r.Body).Decode(v)
		if err != nil {
			t.Fatalf("decode json: %v", err)
		}

		testMethod(t, r, http.MethodPost)
		if !reflect.DeepEqual(v, assignRequest) {
			t.Errorf("Request body = %#v, expected %#v", v, assignRequest)
		}

		fmt.Fprintf(w, `{"action":{"status":"in-progress","id":1,"type":"assign_ip","resource_type":"reserved_ipv6"}}`)

	})

	assign, _, err := client.ReservedIPV6Actions.Assign(ctx, "2604:a880:800:14::42c3:d000", 12345)
	if err != nil {
		t.Errorf("ReservedIPV6sActions.Assign returned error: %v", err)
	}

	expected := &Action{Status: "in-progress", ID: 1, Type: "assign_ip", ResourceType: "reserved_ipv6"}
	if !reflect.DeepEqual(assign, expected) {
		t.Errorf("ReservedIPV6sActions.Assign returned %+v, expected %+v", assign, expected)
	}
}

func TestReservedIPV6sActions_Unassign(t *testing.T) {
	setup()
	defer teardown()

	unassignRequest := &ActionRequest{
		"type": "unassign",
	}

	mux.HandleFunc("/v2/reserved_ipv6/2604:a880:800:14::42c3:d000/actions", func(w http.ResponseWriter, r *http.Request) {
		v := new(ActionRequest)
		err := json.NewDecoder(r.Body).Decode(v)
		if err != nil {
			t.Fatalf("decode json: %v", err)
		}

		testMethod(t, r, http.MethodPost)
		if !reflect.DeepEqual(v, unassignRequest) {
			t.Errorf("Request body = %+v, expected %+v", v, unassignRequest)
		}

		fmt.Fprintf(w, `{"action":{"status":"in-progress","id":1,"type":"unassign_ip","resource_type":"reserved_ipv6"}}`)
	})

	action, _, err := client.ReservedIPV6Actions.Unassign(ctx, "2604:a880:800:14::42c3:d000")
	if err != nil {
		t.Errorf("ReservedIPV6sActions.Unassign returned error: %v", err)
	}

	expected := &Action{Status: "in-progress", ID: 1, Type: "unassign_ip", ResourceType: "reserved_ipv6"}
	if !reflect.DeepEqual(action, expected) {
		t.Errorf("ReservedIPV6sActions.Unassign returned %+v, expected %+v", action, expected)
	}
}
