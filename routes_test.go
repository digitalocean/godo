package godo

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	testVPCUUID    = "997615ce-132d-4bae-9270-9ee21b395e5d"
	testSubnetUUID = "6b5c619c-359c-44ca-87e2-47e98170c01d"
	testRouteUUID  = "f0e1d2c3-b4a5-6789-0fed-cba987654321"
)

var routeTestObj = &Route{
	ID:              testRouteUUID,
	Type:            RouteTypeStatic,
	DestinationCIDR: "0.0.0.0/0",
	TargetURNs:      []string{"do:nat_gateway:14aa1d1b-e6ab-4ccb-bb10-dade56fcb8ec"},
	Modifiable:      false,
	CreatedAt:       time.Date(2026, 1, 14, 10, 0, 0, 0, time.UTC),
}

var subnetRouteTestObj = &Route{
	ID:              testRouteUUID,
	Type:            RouteTypeStatic,
	DestinationCIDR: "0.0.0.0/0",
	TargetURNs:      []string{"do:droplet:14aa1d1b-e6ab-4ccb-bb10-dade56fcb8ec"},
	Modifiable:      true,
	CreatedAt:       time.Date(2026, 1, 14, 10, 5, 0, 0, time.UTC),
}

var routeTestJSON = `
    {
      "id": "f0e1d2c3-b4a5-6789-0fed-cba987654321",
      "type": "STATIC",
      "destination_cidr": "0.0.0.0/0",
      "target_urns": [
        "do:nat_gateway:14aa1d1b-e6ab-4ccb-bb10-dade56fcb8ec"
      ],
      "modifiable": false,
      "created_at": "2026-01-14T10:00:00Z"
    }
`

var subnetRouteTestJSON = `
    {
      "id": "f0e1d2c3-b4a5-6789-0fed-cba987654321",
      "type": "STATIC",
      "destination_cidr": "0.0.0.0/0",
      "target_urns": [
        "do:droplet:14aa1d1b-e6ab-4ccb-bb10-dade56fcb8ec"
      ],
      "modifiable": true,
      "created_at": "2026-01-14T10:05:00Z"
    }
`

func TestRoutes_ListVPCRoutes(t *testing.T) {
	setup()
	defer teardown()

	path := "/v2/vpcs/" + testVPCUUID + "/routes"
	want := []*Route{routeTestObj}
	links := &Links{
		Pages: &Pages{
			Next: "https://api.digitalocean.com/v2/vpcs/" + testVPCUUID + "/routes?page=2&per_page=20",
		},
	}
	meta := &Meta{Total: 1}
	jsonBlob := `
{
  "routes": [
` + routeTestJSON + `
  ],
  "links": {
    "pages": {
      "next": "https://api.digitalocean.com/v2/vpcs/` + testVPCUUID + `/routes?page=2&per_page=20"
    }
  },
  "meta": {"total": 1}
}
`

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		testFormValues(t, r, values{"page": "2", "per_page": "20"})
		fmt.Fprint(w, jsonBlob)
	})

	got, resp, err := client.Routes.ListVPCRoutes(ctx, testVPCUUID, &ListOptions{Page: 2, PerPage: 20})
	require.NoError(t, err)
	assert.Equal(t, want, got)
	assert.Equal(t, links, resp.Links)
	assert.Equal(t, meta, resp.Meta)
}

func TestRoutes_ListSubnetRoutes(t *testing.T) {
	setup()
	defer teardown()

	path := "/v2/vpcs/" + testVPCUUID + "/subnets/" + testSubnetUUID + "/routes"
	want := []*Route{subnetRouteTestObj}
	links := &Links{
		Pages: &Pages{
			Next: "https://api.digitalocean.com/v2/vpcs/" + testVPCUUID + "/subnets/" + testSubnetUUID + "/routes?page=2&per_page=20",
		},
	}
	meta := &Meta{Total: 1}
	jsonBlob := `
{
  "routes": [
` + subnetRouteTestJSON + `
  ],
  "links": {
    "pages": {
      "next": "https://api.digitalocean.com/v2/vpcs/` + testVPCUUID + `/subnets/` + testSubnetUUID + `/routes?page=2&per_page=20"
    }
  },
  "meta": {"total": 1}
}
`

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, jsonBlob)
	})

	got, resp, err := client.Routes.ListSubnetRoutes(ctx, testVPCUUID, testSubnetUUID, nil)
	require.NoError(t, err)
	assert.Equal(t, want, got)
	assert.Equal(t, links, resp.Links)
	assert.Equal(t, meta, resp.Meta)
}

func TestRoutes_CreateSubnetRoute(t *testing.T) {
	setup()
	defer teardown()

	path := "/v2/vpcs/" + testVPCUUID + "/subnets/" + testSubnetUUID + "/routes"
	want := subnetRouteTestObj
	createReq := &RouteCreateRequest{
		DestinationCIDR: "0.0.0.0/0",
		TargetURNs:      []string{"do:droplet:14aa1d1b-e6ab-4ccb-bb10-dade56fcb8ec"},
	}
	jsonBlob := `
{
  "route":
` + subnetRouteTestJSON + `
}
`

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		c := new(RouteCreateRequest)
		err := json.NewDecoder(r.Body).Decode(c)
		if err != nil {
			t.Fatal(err)
		}

		testMethod(t, r, http.MethodPost)
		require.Equal(t, createReq, c)
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, jsonBlob)
	})

	got, _, err := client.Routes.CreateSubnetRoute(ctx, testVPCUUID, testSubnetUUID, createReq)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestRoutes_UpdateSubnetRoute(t *testing.T) {
	setup()
	defer teardown()

	path := "/v2/vpcs/" + testVPCUUID + "/subnets/" + testSubnetUUID + "/routes/" + testRouteUUID
	want := subnetRouteTestObj
	updateReq := &RouteUpdateRequest{
		TargetURNs: []string{"do:droplet:14aa1d1b-e6ab-4ccb-bb10-dade56fcb8ec"},
	}
	jsonBlob := `
{
  "route":
` + subnetRouteTestJSON + `
}
`

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		u := new(RouteUpdateRequest)
		err := json.NewDecoder(r.Body).Decode(u)
		if err != nil {
			t.Fatal(err)
		}

		testMethod(t, r, http.MethodPatch)
		require.Equal(t, updateReq, u)
		fmt.Fprint(w, jsonBlob)
	})

	got, _, err := client.Routes.UpdateSubnetRoute(ctx, testVPCUUID, testSubnetUUID, testRouteUUID, updateReq)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestRoutes_DeleteSubnetRoute(t *testing.T) {
	setup()
	defer teardown()

	path := "/v2/vpcs/" + testVPCUUID + "/subnets/" + testSubnetUUID + "/routes/" + testRouteUUID

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		w.WriteHeader(http.StatusNoContent)
	})

	resp, err := client.Routes.DeleteSubnetRoute(ctx, testVPCUUID, testSubnetUUID, testRouteUUID)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}
