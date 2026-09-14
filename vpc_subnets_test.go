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

var subnetTestObj = &VPCSubnet{
	ID:        "3bb1d9f0-e7d4-4527-8018-ec7f6886f943",
	VPCID:     "e427bce8-5ffa-438c-8ae9-8d043d871525",
	Name:      "my-subnet",
	IPRange:   "10.116.0.0/24",
	CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
	Default:   false,
	Type:      "REGULAR",
}

var subnetTestJSON = `
{
  "id": "3bb1d9f0-e7d4-4527-8018-ec7f6886f943",
  "vpc_id": "e427bce8-5ffa-438c-8ae9-8d043d871525",
  "name": "my-subnet",
  "ip_range": "10.116.0.0/24",
  "created_at": "2024-01-15T10:30:00Z",
  "default": false,
  "type": "REGULAR"
}
`

func TestVPCSubnets_Create(t *testing.T) {
	setup()
	defer teardown()

	vpcID := "e427bce8-5ffa-438c-8ae9-8d043d871525"
	path := "/v2/vpcs/" + vpcID + "/subnets"
	want := subnetTestObj

	req := &VPCSubnetCreateRequest{
		Name:    "my-subnet",
		IPRange: "10.116.0.0/24",
	}

	jsonBlob := `
{
  "vpc_subnet":
` + subnetTestJSON + `
}
`

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		c := new(VPCSubnetCreateRequest)
		err := json.NewDecoder(r.Body).Decode(c)
		if err != nil {
			t.Fatal(err)
		}

		testMethod(t, r, http.MethodPost)
		require.Equal(t, req, c)
		fmt.Fprint(w, jsonBlob)
	})

	got, _, err := client.VPCs.CreateSubnet(ctx, vpcID, req)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestVPCSubnets_List(t *testing.T) {
	setup()
	defer teardown()

	vpcID := "e427bce8-5ffa-438c-8ae9-8d043d871525"
	path := "/v2/vpcs/" + vpcID + "/subnets"
	want := []*VPCSubnet{subnetTestObj}

	links := &Links{
		Pages: &Pages{
			Last: "http://localhost/v2/vpcs/" + vpcID + "/subnets?page=3&per_page=1",
			Next: "http://localhost/v2/vpcs/" + vpcID + "/subnets?page=2&per_page=1",
		},
	}
	meta := &Meta{
		Total: 3,
	}

	jsonBlob := `
{
  "vpc_subnets": [
` + subnetTestJSON + `
  ],
  "links": {
    "pages": {
      "last": "http://localhost/v2/vpcs/` + vpcID + `/subnets?page=3&per_page=1",
      "next": "http://localhost/v2/vpcs/` + vpcID + `/subnets?page=2&per_page=1"
    }
  },
  "meta": {"total": 3}
}
`

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, jsonBlob)
	})

	got, resp, err := client.VPCs.ListSubnets(ctx, vpcID, nil)
	require.NoError(t, err)
	assert.Equal(t, want, got)
	assert.Equal(t, links, resp.Links)
	assert.Equal(t, meta, resp.Meta)
}

func TestVPCSubnets_List_WithPagination(t *testing.T) {
	setup()
	defer teardown()

	vpcID := "e427bce8-5ffa-438c-8ae9-8d043d871525"
	path := "/v2/vpcs/" + vpcID + "/subnets"

	jsonBlob := `{"vpc_subnets":[],"links":{},"meta":{}}`

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		testFormValues(t, r, values{"page": "2", "per_page": "20"})
		fmt.Fprint(w, jsonBlob)
	})

	_, _, err := client.VPCs.ListSubnets(ctx, vpcID, &ListOptions{Page: 2, PerPage: 20})
	require.NoError(t, err)
}

func TestVPCSubnets_Get(t *testing.T) {
	setup()
	defer teardown()

	vpcID := "e427bce8-5ffa-438c-8ae9-8d043d871525"
	subnetID := "3bb1d9f0-e7d4-4527-8018-ec7f6886f943"
	path := "/v2/vpcs/" + vpcID + "/subnets"
	want := subnetTestObj

	jsonBlob := `
{
  "vpc_subnet":
` + subnetTestJSON + `
}
`

	mux.HandleFunc(path+"/"+subnetID, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, jsonBlob)
	})

	got, _, err := client.VPCs.GetSubnet(ctx, vpcID, subnetID)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestVPCSubnets_Update(t *testing.T) {
	setup()
	defer teardown()

	vpcID := "e427bce8-5ffa-438c-8ae9-8d043d871525"
	subnetID := "3bb1d9f0-e7d4-4527-8018-ec7f6886f943"
	path := "/v2/vpcs/" + vpcID + "/subnets/" + subnetID

	req := &VPCSubnetUpdateRequest{
		Name: "my-subnet-updated",
	}

	want := &VPCSubnet{
		ID:        subnetID,
		VPCID:     vpcID,
		Name:      "my-subnet-updated",
		IPRange:   "10.116.0.0/24",
		CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		Default:   false,
		Type:      "REGULAR",
	}

	jsonBlob := `
{
  "vpc_subnet": {
    "id": "3bb1d9f0-e7d4-4527-8018-ec7f6886f943",
    "vpc_id": "e427bce8-5ffa-438c-8ae9-8d043d871525",
    "name": "my-subnet-updated",
    "ip_range": "10.116.0.0/24",
    "created_at": "2024-01-15T10:30:00Z",
    "default": false,
    "type": "REGULAR"
  }
}
`

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		c := new(VPCSubnetUpdateRequest)
		err := json.NewDecoder(r.Body).Decode(c)
		if err != nil {
			t.Fatal(err)
		}

		testMethod(t, r, http.MethodPatch)
		require.Equal(t, req, c)
		fmt.Fprint(w, jsonBlob)
	})

	got, _, err := client.VPCs.UpdateSubnet(ctx, vpcID, subnetID, req)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestVPCSubnets_Delete(t *testing.T) {
	setup()
	defer teardown()

	vpcID := "e427bce8-5ffa-438c-8ae9-8d043d871525"
	subnetID := "3bb1d9f0-e7d4-4527-8018-ec7f6886f943"
	path := "/v2/vpcs/" + vpcID + "/subnets/" + subnetID

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		w.WriteHeader(http.StatusNoContent)
	})

	resp, err := client.VPCs.DeleteSubnet(ctx, vpcID, subnetID)
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestVPCSubnets_Create_AutoGenerateIPRange(t *testing.T) {
	setup()
	defer teardown()

	vpcID := "e427bce8-5ffa-438c-8ae9-8d043d871525"
	path := "/v2/vpcs/" + vpcID + "/subnets"

	req := &VPCSubnetCreateRequest{
		Name: "my-subnet",
		// IPRange omitted - should auto-generate
	}

	want := &VPCSubnet{
		ID:        "3bb1d9f0-e7d4-4527-8018-ec7f6886f943",
		VPCID:     vpcID,
		Name:      "my-subnet",
		IPRange:   "10.116.1.0/24", // Auto-generated by API
		CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		Default:   false,
		Type:      "REGULAR",
	}

	jsonBlob := `
{
  "vpc_subnet": {
    "id": "3bb1d9f0-e7d4-4527-8018-ec7f6886f943",
    "vpc_id": "e427bce8-5ffa-438c-8ae9-8d043d871525",
    "name": "my-subnet",
    "ip_range": "10.116.1.0/24",
    "created_at": "2024-01-15T10:30:00Z",
    "default": false,
    "type": "REGULAR"
  }
}
`

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		c := new(VPCSubnetCreateRequest)
		err := json.NewDecoder(r.Body).Decode(c)
		if err != nil {
			t.Fatal(err)
		}

		testMethod(t, r, http.MethodPost)
		// Verify IPRange was omitted in request
		require.Empty(t, c.IPRange)
		fmt.Fprint(w, jsonBlob)
	})

	got, _, err := client.VPCs.CreateSubnet(ctx, vpcID, req)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestVPCSubnet_UnmarshalJSON(t *testing.T) {
	var subnet VPCSubnet
	err := json.Unmarshal([]byte(subnetTestJSON), &subnet)
	require.NoError(t, err)
	require.Equal(t, subnetTestObj.ID, subnet.ID)
	require.Equal(t, subnetTestObj.VPCID, subnet.VPCID)
	require.Equal(t, subnetTestObj.Name, subnet.Name)
	require.Equal(t, subnetTestObj.IPRange, subnet.IPRange)
	require.Equal(t, subnetTestObj.Default, subnet.Default)
	require.Equal(t, subnetTestObj.Type, subnet.Type)
}
