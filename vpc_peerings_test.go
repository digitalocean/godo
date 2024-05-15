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

var vPeerTestObj = &VPCPeering{
	ID:        "f5a0c5e4-7537-47de-bb8d-46c766f89ffb",
	Name:      "my-new-vpc-peering",
	VPCIDs:    []string{"880b7f98-f062-404d-b33c-458d545696f6"},
	CreatedAt: time.Date(2019, 2, 4, 21, 48, 40, 995304079, time.UTC),
	Status:    "ACTIVE",
}

var vPeerTestJSON = `
    {
      "id":"f5a0c5e4-7537-47de-bb8d-46c766f89ffb",
      "name":"my-new-vpc-peering",
      "vpc_ids":["880b7f98-f062-404d-b33c-458d545696f6"],
      "created_at":"2019-02-04T21:48:40.995304079Z",
      "status":"ACTIVE"
    }
`

func TestVPCPeering_Create(t *testing.T) {
	setup()
	defer teardown()

	svc := client.VPCs
	path := "/v2/vpc_peerings"
	want := vPeerTestObj
	req := &VPCPeeringCreateRequest{
		Name:   "my-new-vpc-peering",
		VPCIDs: []string{"880b7f98-f062-404d-b33c-458d545696f6"},
	}
	jsonBlob := `
{
  "vpc_peering":
` + vPeerTestJSON + `
}
`

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		c := new(VPCPeeringCreateRequest)
		err := json.NewDecoder(r.Body).Decode(c)
		if err != nil {
			t.Fatal(err)
		}

		testMethod(t, r, http.MethodPost)
		require.Equal(t, c, req)
		fmt.Fprint(w, jsonBlob)
	})

	got, _, err := svc.CreateVPCPeering(ctx, req)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestVPCPeering_Get(t *testing.T) {
	setup()
	defer teardown()

	svc := client.VPCs
	path := "/v2/vpc_peerings"
	want := vPeerTestObj
	id := "880b7f98-f062-404d-b33c-458d545696f6"
	jsonBlob := `
{
  "vpc_peering":
` + vPeerTestJSON + `
}
`

	mux.HandleFunc(path+"/"+id, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, jsonBlob)
	})

	got, _, err := svc.GetVPCPeering(ctx, id)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestVPCPeering_List(t *testing.T) {
	setup()
	defer teardown()

	svc := client.VPCs
	path := "/v2/vpc_peerings"
	want := []*VPCPeering{
		vPeerTestObj,
	}
	links := &Links{
		Pages: &Pages{
			Last: "http://localhost/v2/vpcs?page=3&per_page=1",
			Next: "http://localhost/v2/vpcs?page=2&per_page=1",
		},
	}

	meta := &Meta{
		Total: 3,
	}
	jsonBlob := `
{
  "vpc_peerings": [
` + vPeerTestJSON + `
  ],
  "links": {
    "pages": {
      "last": "http://localhost/v2/vpcs?page=3&per_page=1",
      "next": "http://localhost/v2/vpcs?page=2&per_page=1"
    }
  },
  "meta": {"total": 3}
}
`
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, jsonBlob)
	})

	got, resp, err := svc.ListVPCPeerings(ctx, nil)
	require.NoError(t, err)
	assert.Equal(t, want, got)
	assert.Equal(t, resp.Links, links)
	assert.Equal(t, resp.Meta, meta)
}

func TestVPCPeering_Update(t *testing.T) {
	setup()
	defer teardown()

	svc := client.VPCs
	path := "/v2/vpc_peerings"
	want := vPeerTestObj
	id := "880b7f98-f062-404d-b33c-458d545696f6"
	req := &VPCPeeringUpdateRequest{
		Name: "my-new-vpc-peering",
	}
	jsonBlob := `
{
  "vpc_peering":
` + vPeerTestJSON + `
}
`

	mux.HandleFunc(path+"/"+id, func(w http.ResponseWriter, r *http.Request) {
		u := new(VPCPeeringUpdateRequest)
		err := json.NewDecoder(r.Body).Decode(u)
		if err != nil {
			t.Fatal(err)
		}

		testMethod(t, r, http.MethodPatch)
		require.Equal(t, u, req)
		fmt.Fprint(w, jsonBlob)
	})

	got, _, err := svc.UpdateVPCPeering(ctx, id, req)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestVPCPeering_Delete(t *testing.T) {
	setup()
	defer teardown()

	svc := client.VPCs
	path := "/v2/vpc_peerings"
	id := "f5a0c5e4-7537-47de-bb8d-46c766f89ffb"

	mux.HandleFunc(path+"/"+id, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
	})

	_, err := svc.DeleteVPCPeering(ctx, id)
	require.NoError(t, err)
}

func TestVPCPeering_CreateByVPCID(t *testing.T) {
	setup()
	defer teardown()

	svc := client.VPCs
	path := "/v2/vpcs/%s/peerings"
	want := vPeerTestObj
	id := "880b7f98-f062-404d-b33c-458d545696f6"
	req := &VPCPeeringCreateRequestByVPCID{
		Name:  "my-new-vpc-peering",
		VPCID: id,
	}
	jsonBlob := `
{
  "vpc_peering":
` + vPeerTestJSON + `
}
`

	mux.HandleFunc(fmt.Sprintf(path, id), func(w http.ResponseWriter, r *http.Request) {
		c := new(VPCPeeringCreateRequestByVPCID)
		err := json.NewDecoder(r.Body).Decode(c)
		if err != nil {
			t.Fatal(err)
		}

		testMethod(t, r, http.MethodPost)
		require.Equal(t, c, req)
		fmt.Fprint(w, jsonBlob)
	})

	got, _, err := svc.CreateVPCPeeringByVPCID(ctx, id, req)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestVPCPeering_ListByVPCID(t *testing.T) {
	setup()
	defer teardown()

	svc := client.VPCs
	path := "/v2/vpcs/880b7f98-f062-404d-b33c-458d545696f6/peerings"
	want := []*VPCPeering{
		vPeerTestObj,
	}
	links := &Links{
		Pages: &Pages{
			Last: "http://localhost/v2/vpcs?page=3&per_page=1",
			Next: "http://localhost/v2/vpcs?page=2&per_page=1",
		},
	}

	meta := &Meta{
		Total: 3,
	}
	jsonBlob := `
{
  "vpc_peerings": [
` + vPeerTestJSON + `
  ],
  "links": {
    "pages": {
      "last": "http://localhost/v2/vpcs?page=3&per_page=1",
      "next": "http://localhost/v2/vpcs?page=2&per_page=1"
    }
  },
  "meta": {"total": 3}
}
`
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, jsonBlob)
	})

	got, resp, err := svc.ListVPCPeeringsByVPCID(ctx, "880b7f98-f062-404d-b33c-458d545696f6", nil)
	require.NoError(t, err)
	assert.Equal(t, want, got)
	assert.Equal(t, resp.Links, links)
	assert.Equal(t, resp.Meta, meta)
}

func TestVPCPeering_UpdateByVPCID(t *testing.T) {
	setup()
	defer teardown()

	svc := client.VPCs
	path := "/v2/vpcs/%s/peerings/%s"
	want := vPeerTestObj
	vpcID := "880b7f98-f062-404d-b33c-458d545696f6"
	peerID := "peer-id"
	req := &VPCPeeringUpdateRequest{
		Name: "my-new-vpc-peering",
	}
	jsonBlob := `
{
  "vpc_peering":
` + vPeerTestJSON + `
}
`

	mux.HandleFunc(fmt.Sprintf(path, vpcID, peerID), func(w http.ResponseWriter, r *http.Request) {
		u := new(VPCPeeringUpdateRequest)
		err := json.NewDecoder(r.Body).Decode(u)
		if err != nil {
			t.Fatal(err)
		}

		testMethod(t, r, http.MethodPatch)
		require.Equal(t, u, req)
		fmt.Fprint(w, jsonBlob)
	})

	got, _, err := svc.UpdateVPCPeeringByVPCID(ctx, vpcID, peerID, req)
	require.NoError(t, err)
	require.Equal(t, want, got)
}
