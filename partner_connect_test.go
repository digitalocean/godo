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

var vPartnerConnectTestObj = &PartnerConnect{
	ID:                        "880b7f98-f062-404d-b33c-458d545696f6",
	Name:                      "my-new-partner-connect",
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
		AuthKey:       "my-auth-key",
	},
	CreatedAt: time.Date(2024, 12, 26, 21, 48, 40, 995304079, time.UTC),
}

var vPartnerConnectNoBGPTestObj = &PartnerConnect{
	ID:                        "880b7f98-f062-404d-b33c-458d545696f6",
	Name:                      "my-new-partner-connect",
	State:                     "ACTIVE",
	ConnectionBandwidthInMbps: 50,
	Region:                    "NYC",
	NaaSProvider:              "MEGAPORT",
	VPCIDs:                    []string{"f5a0c5e4-7537-47de-bb8d-46c766f89ffb"},
	CreatedAt:                 time.Date(2024, 12, 26, 21, 48, 40, 995304079, time.UTC),
}

var vPartnerConnectTestJSON = `
	{
		"id":"880b7f98-f062-404d-b33c-458d545696f6",
		"name":"my-new-partner-connect",
		"state":"ACTIVE",
		"connection_bandwidth_in_mbps":50,
		"region":"NYC",
		"naas_provider":"MEGAPORT",
		"vpc_ids":["f5a0c5e4-7537-47de-bb8d-46c766f89ffb"],
		"bgp":{
			"local_asn":64532,
			"local_router_ip":"169.250.0.1",
			"peer_asn":133937,
			"peer_router_ip":"169.250.0.6",
			"auth_key":"my-auth-key"
			},
		"created_at":"2024-12-26T21:48:40.995304079Z"
	}
`

var vPartnerConnectNoBGPTestJSON = `
	{
		"id":"880b7f98-f062-404d-b33c-458d545696f6",
		"name":"my-new-partner-connect",
		"state":"ACTIVE",
		"connection_bandwidth_in_mbps":50,
		"region":"NYC",
		"naas_provider":"MEGAPORT",
		"vpc_ids":["f5a0c5e4-7537-47de-bb8d-46c766f89ffb"],
		"created_at":"2024-12-26T21:48:40.995304079Z"
	}
`

const expectedPartnerConnectCreateBodyNoBGP = `{"name":"my-new-partner-connect","connection_bandwidth_in_mbps":50,"region":"NYC","naas_provider":"MEGAPORT","vpc_ids":["f5a0c5e4-7537-47de-bb8d-46c766f89ffb"]}
`

func TestPartnerConnects_List(t *testing.T) {
	setup()
	defer teardown()

	svc := client.PartnerConnect
	path := "/v2/partner_connect/attachments"
	want := []*PartnerConnect{
		vPartnerConnectTestObj,
	}
	links := &Links{
		Pages: &Pages{
			Last: "http://localhost/v2/partner_connect/attachments?page=3&per_page=1",
			Next: "http://localhost/v2/partner_connect/attachments?page=2&per_page=1",
		},
	}
	meta := &Meta{
		Total: 3,
	}
	jsonBlob := `
{
  "partner_connects": [
` + vPartnerConnectTestJSON + `
  ],
  "links": {
    "pages": {
      "last": "http://localhost/v2/partner_connect/attachments?page=3&per_page=1",
      "next": "http://localhost/v2/partner_connect/attachments?page=2&per_page=1"
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

func TestPartnerConnects_Create(t *testing.T) {
	setup()
	defer teardown()

	svc := client.PartnerConnect
	path := "/v2/partner_connect/attachments"
	want := vPartnerConnectTestObj
	req := &PartnerConnectCreateRequest{
		Name:                      "my-new-partner-connect",
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
	"partner_connect":
` + vPartnerConnectTestJSON + `
}
`

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		c := new(PartnerConnectCreateRequest)
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

func TestPartnerConnects_CreateNoBGP(t *testing.T) {
	setup()
	defer teardown()

	svc := client.PartnerConnect
	path := "/v2/partner_connect/attachments"
	want := vPartnerConnectNoBGPTestObj
	req := &PartnerConnectCreateRequest{
		Name:                      "my-new-partner-connect",
		ConnectionBandwidthInMbps: 50,
		Region:                    "NYC",
		NaaSProvider:              "MEGAPORT",
		VPCIDs:                    []string{"f5a0c5e4-7537-47de-bb8d-46c766f89ffb"},
	}
	jsonBlob := `
{
	"partner_connect":
` + vPartnerConnectNoBGPTestJSON + `
}
`

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		defer r.Body.Close()

		require.Equal(t, expectedPartnerConnectCreateBodyNoBGP, string(body))

		c := new(PartnerConnectCreateRequest)
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

func TestPartnerConnects_Get(t *testing.T) {
	setup()
	defer teardown()

	svc := client.PartnerConnect
	path := "/v2/partner_connect/attachments"
	want := vPartnerConnectTestObj
	id := "880b7f98-f062-404d-b33c-458d545696f6"
	jsonBlob := `
{
	"partner_connect":
` + vPartnerConnectTestJSON + `
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

func TestPartnerConnects_Update(t *testing.T) {
	setup()
	defer teardown()

	svc := client.PartnerConnect
	path := "/v2/partner_connect/attachments"
	want := vPartnerConnectTestObj
	id := "880b7f98-f062-404d-b33c-458d545696f6"
	req := &PartnerConnectUpdateRequest{
		Name:   "my-renamed-partner-connect",
		VPCIDs: []string{"g5a0c5e4-7537-47de-bb8d-46c766f89ffb"},
	}
	jsonBlob := `
{
	"partner_connect":
` + vPartnerConnectTestJSON + `
}
`

	mux.HandleFunc(path+"/"+id, func(w http.ResponseWriter, r *http.Request) {
		v := new(PartnerConnectUpdateRequest)
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

func TestPartnerConnect_Delete(t *testing.T) {
	setup()
	defer teardown()

	svc := client.PartnerConnect
	path := "/v2/partner_connect/attachments"
	id := "880b7f98-f062-404d-b33c-458d545696f6"

	mux.HandleFunc(path+"/"+id, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
	})

	_, err := svc.Delete(ctx, id)
	require.NoError(t, err)
}

func TestPartnerConnect_GetServiceKey(t *testing.T) {
	setup()
	defer teardown()

	svc := client.PartnerConnect
	path := "/v2/partner_connect/attachments"
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

func TestPartnerConnect_ListRoutes(t *testing.T) {
	setup()
	defer teardown()

	svc := client.PartnerConnect
	path := "/v2/partner_connect/attachments"
	want := []*RemoteRoute{
		{
			ID:   "a0eb6eb0-fa38-41a8-a5de-1a75524667fe",
			Cidr: "169.250.0.0/29",
		},
	}
	links := &Links{
		Pages: &Pages{
			Last: "http://localhost/v2/partner_connect/attachments?page=2&per_page=1",
			Next: "http://localhost/v2/partner_connect/attachments?page=2&per_page=1",
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
      "last": "http://localhost/v2/partner_connect/attachments?page=2&per_page=1",
      "next": "http://localhost/v2/partner_connect/attachments?page=2&per_page=1"
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

func TestPartnerConnect_Set(t *testing.T) {
	tests := []struct {
		desc                        string
		id                          string
		req                         *PartnerConnectSetRoutesRequest
		mockResponse                string
		expectedRequestBody         string
		expectedUpdatedInterconnect *PartnerConnect
	}{
		{
			desc: "set remote routes",
			id:   "880b7f98-f062-404d-b33c-458d545696f6",
			req: &PartnerConnectSetRoutesRequest{
				Routes: []string{"169.250.0.1/29", "169.250.0.6/29"},
			},
			mockResponse: `
{
	"partner_connect":
` + vPartnerConnectTestJSON + `
}
			`,
			expectedRequestBody:         `{"routes":["169.250.0.1/29", "169.250.0.6/29"]}`,
			expectedUpdatedInterconnect: vPartnerConnectTestObj,
		},
	}

	for _, tt := range tests {
		setup()

		mux.HandleFunc("/v2/partner_connect/attachments/"+tt.id+"/remote_routes", func(w http.ResponseWriter, r *http.Request) {
			v := new(PartnerConnectSetRoutesRequest)
			err := json.NewDecoder(r.Body).Decode(v)
			if err != nil {
				t.Fatal(err)
			}

			testMethod(t, r, http.MethodPut)
			require.Equal(t, v, tt.req)
			w.Write([]byte(tt.mockResponse))
		})

		got, _, err := client.PartnerConnect.SetRoutes(ctx, tt.id, tt.req)

		teardown()

		require.NoError(t, err)
		require.Equal(t, tt.expectedUpdatedInterconnect, got)
	}
}

func TestPartnerConnect_GetBGPAuthKey(t *testing.T) {
	setup()
	defer teardown()

	svc := client.PartnerConnect
	path := "/v2/partner_connect/attachments"
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

func TestPartnerConnect_RegenerateServiceKey(t *testing.T) {
	setup()
	defer teardown()

	svc := client.PartnerConnect
	path := "/v2/partner_connect/attachments"
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
