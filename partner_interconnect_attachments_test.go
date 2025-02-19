package godo

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var vInterconnectTestObj = &PartnerInterconnectAttachment{
	ID:                        "880b7f98-f062-404d-b33c-458d545696f6",
	Name:                      "my-new-partner-interconnect",
	State:                     "ACTIVE",
	ConnectionBandwidthInMbps: 50,
	Region:                    "NYC",
	NaaSProvider:              "MEGAPORT",
	VPCIDs:                    []string{"f5a0c5e4-7537-47de-bb8d-46c766f89ffb"},
	BGP: BGP{
		LocalASN:      64532,
		LocalRouterIP: "169.250.0.1",
		PeerASN:       133937,
		PeerRouterIP:  "169.250.0.6",
	},
	CreatedAt: time.Date(2024, 12, 26, 21, 48, 40, 995304079, time.UTC),
}

var vInterconnectNoBGPTestObj = &PartnerInterconnectAttachment{
	ID:                        "880b7f98-f062-404d-b33c-458d545696f6",
	Name:                      "my-new-partner-interconnect",
	State:                     "ACTIVE",
	ConnectionBandwidthInMbps: 50,
	Region:                    "NYC",
	NaaSProvider:              "MEGAPORT",
	VPCIDs:                    []string{"f5a0c5e4-7537-47de-bb8d-46c766f89ffb"},
	CreatedAt:                 time.Date(2024, 12, 26, 21, 48, 40, 995304079, time.UTC),
}

var vInterconnectTestJSON = `
	{
		"id":"880b7f98-f062-404d-b33c-458d545696f6",
		"name":"my-new-partner-interconnect",
		"state":"ACTIVE",
		"connection_bandwidth_in_mbps":50,
		"region":"NYC",
		"naas_provider":"MEGAPORT",
		"vpc_ids":["f5a0c5e4-7537-47de-bb8d-46c766f89ffb"],
		"bgp":{
			"local_asn":64532,
			"local_router_ip":"169.250.0.1",
			"peer_router_asn":133937,
			"peer_router_ip":"169.250.0.6"
			},
		"created_at":"2024-12-26T21:48:40.995304079Z"
	}
`

var vInterconnectNoBGPTestJSON = `
	{
		"id":"880b7f98-f062-404d-b33c-458d545696f6",
		"name":"my-new-partner-interconnect",
		"state":"ACTIVE",
		"connection_bandwidth_in_mbps":50,
		"region":"NYC",
		"naas_provider":"MEGAPORT",
		"vpc_ids":["f5a0c5e4-7537-47de-bb8d-46c766f89ffb"],
		"created_at":"2024-12-26T21:48:40.995304079Z"
	}
`

const expectedCreateBodyNoBGP = `{"name":"my-new-partner-interconnect","connection_bandwidth_in_mbps":50,"region":"NYC","naas_provider":"MEGAPORT","vpc_ids":["f5a0c5e4-7537-47de-bb8d-46c766f89ffb"]}
`

func TestPartnerInterconnectAttachments_List(t *testing.T) {
	setup()
	defer teardown()

	svc := client.PartnerInterconnectAttachments
	path := "/v2/partner_interconnect/attachments"
	want := []*PartnerInterconnectAttachment{
		vInterconnectTestObj,
	}
	links := &Links{
		Pages: &Pages{
			Last: "http://localhost/v2/partner_interconnect/attachments?page=3&per_page=1",
			Next: "http://localhost/v2/partner_interconnect/attachments?page=2&per_page=1",
		},
	}
	meta := &Meta{
		Total: 3,
	}
	jsonBlob := `
{
  "partner_interconnect_attachments": [
` + vInterconnectTestJSON + `
  ],
  "links": {
    "pages": {
      "last": "http://localhost/v2/partner_interconnect/attachments?page=3&per_page=1",
      "next": "http://localhost/v2/partner_interconnect/attachments?page=2&per_page=1"
    }
  },
  "meta": {"total": 3}
}
`
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Write([]byte(jsonBlob))
	})

	got, resp, err := svc.List(ctx, nil)
	require.NoError(t, err)
	assert.Equal(t, want, got)
	assert.Equal(t, resp.Links, links)
	assert.Equal(t, resp.Meta, meta)
}

func TestPartnerInterconnectAttachments_Create(t *testing.T) {
	setup()
	defer teardown()

	svc := client.PartnerInterconnectAttachments
	path := "/v2/partner_interconnect/attachments"
	want := vInterconnectTestObj
	req := &PartnerInterconnectAttachmentCreateRequest{
		Name:                      "my-new-partner-interconnect",
		ConnectionBandwidthInMbps: 50,
		Region:                    "NYC",
		NaaSProvider:              "MEGAPORT",
		VPCIDs:                    []string{"f5a0c5e4-7537-47de-bb8d-46c766f89ffb"},
		BGP: BGP{
			LocalASN:      64532,
			LocalRouterIP: "169.250.0.1",
			PeerASN:       133937,
			PeerRouterIP:  "169.250.0.6",
		},
	}
	jsonBlob := `
{
	"partner_interconnect_attachment":
` + vInterconnectTestJSON + `
}
`

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		c := new(PartnerInterconnectAttachmentCreateRequest)
		err := json.NewDecoder(r.Body).Decode(c)
		if err != nil {
			t.Fatal(err)
		}

		testMethod(t, r, http.MethodPost)
		require.Equal(t, c, req)
		w.Write([]byte(jsonBlob))
	})

	got, _, err := svc.Create(ctx, req)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestPartnerInterconnectAttachments_CreateNoBGP(t *testing.T) {
	setup()
	defer teardown()

	svc := client.PartnerInterconnectAttachments
	path := "/v2/partner_interconnect/attachments"
	want := vInterconnectNoBGPTestObj
	req := &PartnerInterconnectAttachmentCreateRequest{
		Name:                      "my-new-partner-interconnect",
		ConnectionBandwidthInMbps: 50,
		Region:                    "NYC",
		NaaSProvider:              "MEGAPORT",
		VPCIDs:                    []string{"f5a0c5e4-7537-47de-bb8d-46c766f89ffb"},
	}
	jsonBlob := `
{
	"partner_interconnect_attachment":
` + vInterconnectNoBGPTestJSON + `
}
`

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		defer r.Body.Close()

		require.Equal(t, expectedCreateBodyNoBGP, string(body))

		c := new(PartnerInterconnectAttachmentCreateRequest)
		err = json.Unmarshal(body, c)
		if err != nil {
			t.Fatal(err)
		}

		testMethod(t, r, http.MethodPost)
		require.Equal(t, c, req)
		w.Write([]byte(jsonBlob))
	})

	got, _, err := svc.Create(ctx, req)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestPartnerInterconnectAttachments_Get(t *testing.T) {
	setup()
	defer teardown()

	svc := client.PartnerInterconnectAttachments
	path := "/v2/partner_interconnect/attachments"
	want := vInterconnectTestObj
	id := "880b7f98-f062-404d-b33c-458d545696f6"
	jsonBlob := `
{
	"partner_interconnect_attachment":
` + vInterconnectTestJSON + `
}
`

	mux.HandleFunc(path+"/"+id, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Write([]byte(jsonBlob))
	})

	got, _, err := svc.Get(ctx, id)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestPartnerInterconnectAttachments_Update(t *testing.T) {
	setup()
	defer teardown()

	svc := client.PartnerInterconnectAttachments
	path := "/v2/partner_interconnect/attachments"
	want := vInterconnectTestObj
	id := "880b7f98-f062-404d-b33c-458d545696f6"
	req := &PartnerInterconnectAttachmentUpdateRequest{
		Name:   "my-renamed-partner-interconnect",
		VPCIDs: []string{"g5a0c5e4-7537-47de-bb8d-46c766f89ffb"},
	}
	jsonBlob := `
{
	"partner_interconnect_attachment":
` + vInterconnectTestJSON + `
}
`

	mux.HandleFunc(path+"/"+id, func(w http.ResponseWriter, r *http.Request) {
		v := new(PartnerInterconnectAttachmentUpdateRequest)
		err := json.NewDecoder(r.Body).Decode(v)
		if err != nil {
			t.Fatal(err)
		}

		testMethod(t, r, http.MethodPatch)
		require.Equal(t, v, req)
		w.Write([]byte(jsonBlob))
	})

	got, _, err := svc.Update(ctx, id, req)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestPartnerInterconnectAttachment_Delete(t *testing.T) {
	setup()
	defer teardown()

	svc := client.PartnerInterconnectAttachments
	path := "/v2/partner_interconnect/attachments"
	id := "880b7f98-f062-404d-b33c-458d545696f6"

	mux.HandleFunc(path+"/"+id, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
	})

	_, err := svc.Delete(ctx, id)
	require.NoError(t, err)
}

func TestPartnerInterconnectAttachment_GetServiceKey(t *testing.T) {
	setup()
	defer teardown()

	svc := client.PartnerInterconnectAttachments
	path := "/v2/partner_interconnect/attachments"
	want := &ServiceKey{
		Value:     "my-service-key",
		State:     "ACTIVE",
		CreatedAt: time.Date(2024, 12, 26, 21, 48, 40, 995304079, time.UTC),
	}
	id := "880b7f98-f062-404d-b33c-458d545696f6"
	jsonBlob := `
{
	"service_key": {
		"value": "my-service-key",
		"state": "ACTIVE",
		"created_at": "2024-12-26T21:48:40.995304079Z"
	}
}
`

	mux.HandleFunc(path+"/"+id+"/service_key", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Write([]byte(jsonBlob))
	})

	got, _, err := svc.GetServiceKey(ctx, id)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestPartnerInterconnectAttachment_ListRoutes(t *testing.T) {
	setup()
	defer teardown()

	svc := client.PartnerInterconnectAttachments
	path := "/v2/partner_interconnect/attachments"
	want := []*RemoteRoute{
		{
			ID:   "a0eb6eb0-fa38-41a8-a5de-1a75524667fe",
			Cidr: "169.250.0.0/29",
		},
	}
	links := &Links{
		Pages: &Pages{
			Last: "http://localhost/v2/partner_interconnect/attachments?page=2&per_page=1",
			Next: "http://localhost/v2/partner_interconnect/attachments?page=2&per_page=1",
		},
	}
	meta := &Meta{
		Total: 1,
	}
	id := "880b7f98-f062-404d-b33c-458d545696f6"
	jsonBlob := `
{
  "remote_routes": [
	{"id": "a0eb6eb0-fa38-41a8-a5de-1a75524667fe", "cidr": "169.250.0.0/29"}
  ],
  "links": {
    "pages": {
      "last": "http://localhost/v2/partner_interconnect/attachments?page=2&per_page=1",
      "next": "http://localhost/v2/partner_interconnect/attachments?page=2&per_page=1"
    }
  },
  "meta": {"total": 1}
}
`

	mux.HandleFunc(path+"/"+id+"/remote_routes", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Write([]byte(jsonBlob))
	})

	got, resp, err := svc.ListRoutes(ctx, id, nil)
	require.NoError(t, err)
	require.Equal(t, want, got)
	assert.Equal(t, resp.Links, links)
	assert.Equal(t, resp.Meta, meta)
}

func TestPartnerInterconnectAttachment_Set(t *testing.T) {
	tests := []struct {
		desc                        string
		id                          string
		req                         *PartnerInterconnectAttachmentSetRoutesRequest
		mockResponse                string
		expectedRequestBody         string
		expectedUpdatedInterconnect *PartnerInterconnectAttachment
	}{
		{
			desc: "set remote routes",
			id:   "880b7f98-f062-404d-b33c-458d545696f6",
			req: &PartnerInterconnectAttachmentSetRoutesRequest{
				Routes: []string{"169.250.0.1/29", "169.250.0.6/29"},
			},
			mockResponse: `
{
	"partner_interconnect_attachment":
` + vInterconnectTestJSON + `
}
			`,
			expectedRequestBody:         `{"routes":["169.250.0.1/29", "169.250.0.6/29"]}`,
			expectedUpdatedInterconnect: vInterconnectTestObj,
		},
	}

	for _, tt := range tests {
		setup()

		mux.HandleFunc("/v2/partner_interconnect/attachments/"+tt.id+"/remote_routes", func(w http.ResponseWriter, r *http.Request) {
			v := new(PartnerInterconnectAttachmentSetRoutesRequest)
			err := json.NewDecoder(r.Body).Decode(v)
			if err != nil {
				t.Fatal(err)
			}

			testMethod(t, r, http.MethodPut)
			require.Equal(t, v, tt.req)
			w.Write([]byte(tt.mockResponse))
		})

		got, _, err := client.PartnerInterconnectAttachments.SetRoutes(ctx, tt.id, tt.req)

		teardown()

		require.NoError(t, err)
		require.Equal(t, tt.expectedUpdatedInterconnect, got)
	}
}

func TestPartnerInterconnectAttachment_GetBGPAuthKey(t *testing.T) {
	setup()
	defer teardown()

	svc := client.PartnerInterconnectAttachments
	path := "/v2/partner_interconnect/attachments"
	want := &BgpAuthKey{
		Value: "bgp-auth-secret",
	}
	id := "880b7f98-f062-404d-b33c-458d545696f6"
	jsonBlob := `
{
  "bgp_auth_key": {
    "value": "bgp-auth-secret"
  }
}
`

	mux.HandleFunc(path+"/"+id+"/bgp_auth_key", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Write([]byte(jsonBlob))
	})

	got, _, err := svc.GetBGPAuthKey(ctx, id)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestPartnerInterconnectAttachment_RegenerateServiceKey(t *testing.T) {
	setup()
	defer teardown()

	svc := client.PartnerInterconnectAttachments
	path := "/v2/partner_interconnect/attachments"
	id := "880b7f98-f062-404d-b33c-458d545696f6"
	jsonBlob := `{}`

	mux.HandleFunc(path+"/"+id+"/service_key", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		w.Write([]byte(jsonBlob))
	})

	got, _, err := svc.RegenerateServiceKey(ctx, id)
	require.NoError(t, err)

	expectedResponse := regenerateServiceKeyRoot{}
	require.Equal(t, expectedResponse.RegenerateServiceKey, got)
}
