package godo

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var egressGatewayGetJSONResponse = `
{
  "egress_gateway": {
    "id": "97c46619-1f53-493b-8638-00c899f30152",
    "name": "test-egress-gateway-01",
    "type": "PUBLIC",
    "state": "STATE_ACTIVE",
    "region": "nyc3",
    "vpcs": [
      {
        "vpc_uuid": "4637280e-3842-4661-a628-a6f0392959d3",
        "gateway_ip": "10.100.0.110"
      }
    ],
    "egresses": {
      "public_gateways": [
        {
          "ipv4": "10.36.48.172"
        }
      ]
    },
    "udp_timeout_seconds": 30,
    "icmp_timeout_seconds": 30,
    "tcp_timeout_seconds": 300,
    "created_at": "2025-04-28T14:20:00Z",
    "updated_at": "2025-04-28T14:20:03Z"
  }
}
`

var egressGatewayListJSONResponse = `
{
  "egress_gateways": [
    {
      "id": "97c46619-1f53-493b-8638-00c899f30152",
      "name": "test-egress-gateway-01",
      "type": "PUBLIC",
      "state": "STATE_ACTIVE",
      "region": "nyc3",
      "vpcs": [
        {
          "vpc_uuid": "4637280e-3842-4661-a628-a6f0392959d3",
          "gateway_ip": "10.100.0.110"
        }
      ],
      "egresses": {
        "public_gateways": [
          {
            "ipv4": "10.36.48.172"
          }
        ]
      },
      "udp_timeout_seconds": 30,
      "icmp_timeout_seconds": 30,
      "tcp_timeout_seconds": 300,
      "created_at": "2025-04-28T14:20:00Z",
      "updated_at": "2025-04-28T14:20:03Z"
    },
    {
      "id": "8e2fedf5-ce55-4ec3-82ca-e607be36ee08",
      "name": "test-egress-gateway-02",
      "type": "PUBLIC",
      "state": "STATE_ACTIVE",
      "region": "nyc3",
      "vpcs": [
        {
          "vpc_uuid": "4637280e-3842-4661-a628-a6f0392959d3",
          "gateway_ip": "10.100.0.106"
        }
      ],
      "egresses": {
        "public_gateways": [
          {
            "ipv4": "10.100.16.9"
          }
        ]
      },
      "udp_timeout_seconds": 30,
      "icmp_timeout_seconds": 30,
      "tcp_timeout_seconds": 300,
      "created_at": "2025-04-28T14:20:29Z",
      "updated_at": "2025-04-28T14:20:32Z"
    }
  ],
  "links": {},
  "meta": {
    "total": 2
  }
}
`

var egressGatewayUpdateJSONResponse = `
{
  "egress_gateway": {
    "id": "97c46619-1f53-493b-8638-00c899f30152",
    "name": "test-egress-gateway-renamed-01",
    "type": "PUBLIC",
    "state": "STATE_ACTIVE",
    "region": "nyc3",
    "vpcs": [
      {
        "vpc_uuid": "4637280e-3842-4661-a628-a6f0392959d3",
        "gateway_ip": "10.100.0.110"
      }
    ],
    "egresses": {
      "public_gateways": [
        {
          "ipv4": "10.36.48.172"
        }
      ]
    },
    "udp_timeout_seconds": 30,
    "icmp_timeout_seconds": 30,
    "tcp_timeout_seconds": 300,
    "created_at": "2025-04-28T14:20:00Z",
    "updated_at": "2025-04-28T14:20:03Z"
  }
}
`

func TestEgressGateways_Create(t *testing.T) {
	setup()
	defer teardown()

	createReq := &EgressGatewayRequest{
		Name:   "test-egress-gateway-01",
		Type:   "PUBLIC",
		Region: "nyc3",
		VPCs: []*IngressVPC{
			{VpcUUID: "4637280e-3842-4661-a628-a6f0392959d3"},
		},
		UDPTimeoutSeconds:  30,
		ICMPTimeoutSeconds: 30,
		TCPTimeoutSeconds:  300,
	}

	mux.HandleFunc(egressGatewaysBasePath, func(w http.ResponseWriter, r *http.Request) {
		req := new(EgressGatewayRequest)
		err := json.NewDecoder(r.Body).Decode(req)
		if err != nil {
			t.Fatal(err)
		}
		testMethod(t, r, http.MethodPost)
		assert.Equal(t, createReq, req)
		fmt.Fprintf(w, egressGatewayGetJSONResponse)
	})

	expectedGatewayResp := &EgressGateway{
		ID:     "97c46619-1f53-493b-8638-00c899f30152",
		Name:   "test-egress-gateway-01",
		Type:   "PUBLIC",
		State:  "STATE_ACTIVE",
		Region: "nyc3",
		VPCs: []*IngressVPC{
			{VpcUUID: "4637280e-3842-4661-a628-a6f0392959d3", GatewayIP: "10.100.0.110"},
		},
		Egresses: &Egresses{
			PublicGateways: []*PublicGateway{
				{IPv4: "10.36.48.172"},
			},
		},
		UDPTimeoutSeconds:  30,
		ICMPTimeoutSeconds: 30,
		TCPTimeoutSeconds:  300,
	}

	createGatewayResp, _, err := client.EgressGateways.Create(ctx, createReq)
	require.NoError(t, err)
	require.NotEmpty(t, createGatewayResp)
	expectedGatewayResp.CreatedAt = createGatewayResp.CreatedAt
	expectedGatewayResp.UpdatedAt = createGatewayResp.UpdatedAt
	assert.Equal(t, expectedGatewayResp, createGatewayResp)
}

func TestEgressGateways_Get(t *testing.T) {
	setup()
	defer teardown()

	gatewayID := "97c46619-1f53-493b-8638-00c899f30152"
	mux.HandleFunc(fmt.Sprintf("%s/%s", egressGatewaysBasePath, gatewayID), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprintf(w, egressGatewayGetJSONResponse)
	})

	expectedGatewayResp := &EgressGateway{
		ID:     "97c46619-1f53-493b-8638-00c899f30152",
		Name:   "test-egress-gateway-01",
		Type:   "PUBLIC",
		State:  "STATE_ACTIVE",
		Region: "nyc3",
		VPCs: []*IngressVPC{
			{VpcUUID: "4637280e-3842-4661-a628-a6f0392959d3", GatewayIP: "10.100.0.110"},
		},
		Egresses: &Egresses{
			PublicGateways: []*PublicGateway{
				{IPv4: "10.36.48.172"},
			},
		},
		UDPTimeoutSeconds:  30,
		ICMPTimeoutSeconds: 30,
		TCPTimeoutSeconds:  300,
	}

	getGatewayResp, _, err := client.EgressGateways.Get(ctx, gatewayID)
	require.NoError(t, err)
	require.NotNil(t, getGatewayResp)
	expectedGatewayResp.CreatedAt = getGatewayResp.CreatedAt
	expectedGatewayResp.UpdatedAt = getGatewayResp.UpdatedAt
	assert.Equal(t, expectedGatewayResp, getGatewayResp)
}

func TestEgressGateways_List(t *testing.T) {
	setup()
	defer teardown()

	stateOpts := []string{"STATE_NEW", "STATE_ACTIVE"}
	regionOpts := []string{"nyc3", "ams3"}
	typeOpts := []string{"public"}
	nameOpts := []string{"test-ngw-01", "test-ngw-02"}

	mux.HandleFunc(egressGatewaysBasePath, func(w http.ResponseWriter, r *http.Request) {
		queryParams := r.URL.Query()
		assert.True(t, queryParams.Has("state"))
		assert.ElementsMatch(t, stateOpts, queryParams["state"])
		assert.True(t, queryParams.Has("region"))
		assert.ElementsMatch(t, regionOpts, queryParams["region"])
		assert.True(t, queryParams.Has("type"))
		assert.ElementsMatch(t, typeOpts, queryParams["type"])
		assert.True(t, queryParams.Has("name"))
		assert.ElementsMatch(t, nameOpts, queryParams["name"])

		testMethod(t, r, http.MethodGet)
		fmt.Fprintf(w, egressGatewayListJSONResponse)
	})

	expectedGatewaysResp := []*EgressGateway{
		{
			ID:     "97c46619-1f53-493b-8638-00c899f30152",
			Name:   "test-egress-gateway-01",
			Type:   "PUBLIC",
			State:  "STATE_ACTIVE",
			Region: "nyc3",
			VPCs: []*IngressVPC{
				{VpcUUID: "4637280e-3842-4661-a628-a6f0392959d3", GatewayIP: "10.100.0.110"},
			},
			Egresses: &Egresses{
				PublicGateways: []*PublicGateway{
					{IPv4: "10.36.48.172"},
				},
			},
			UDPTimeoutSeconds:  30,
			ICMPTimeoutSeconds: 30,
			TCPTimeoutSeconds:  300,
		},
		{
			ID:     "8e2fedf5-ce55-4ec3-82ca-e607be36ee08",
			Name:   "test-egress-gateway-02",
			Type:   "PUBLIC",
			State:  "STATE_ACTIVE",
			Region: "nyc3",
			VPCs: []*IngressVPC{
				{VpcUUID: "4637280e-3842-4661-a628-a6f0392959d3", GatewayIP: "10.100.0.106"},
			},
			Egresses: &Egresses{
				PublicGateways: []*PublicGateway{
					{IPv4: "10.100.16.9"},
				},
			},
			UDPTimeoutSeconds:  30,
			ICMPTimeoutSeconds: 30,
			TCPTimeoutSeconds:  300,
		},
	}

	listGatewaysResp, _, err := client.EgressGateways.List(ctx, &EgressGatewaysListOptions{
		ListOptions: ListOptions{Page: 1},
		State:       stateOpts,
		Region:      regionOpts,
		Type:        typeOpts,
		Name:        nameOpts,
	})
	require.NoError(t, err)
	require.NotEmpty(t, listGatewaysResp)
	sort.SliceStable(listGatewaysResp, func(i, j int) bool {
		return listGatewaysResp[i].Name < listGatewaysResp[j].Name
	})
	for idx := range listGatewaysResp {
		expectedGatewaysResp[idx].CreatedAt = listGatewaysResp[idx].CreatedAt
		expectedGatewaysResp[idx].UpdatedAt = listGatewaysResp[idx].UpdatedAt
	}
	assert.Equal(t, expectedGatewaysResp, listGatewaysResp)
}

func TestEgressGateways_Update(t *testing.T) {
	setup()
	defer teardown()

	updateReq := &EgressGatewayRequest{
		Name:   "test-egress-gateway-renamed-01",
		Type:   "PUBLIC",
		Region: "nyc3",
		VPCs: []*IngressVPC{
			{VpcUUID: "72b0812c-7535-4388-8507-5ad29b4487b3"},
		},
		UDPTimeoutSeconds:  30,
		ICMPTimeoutSeconds: 30,
		TCPTimeoutSeconds:  300,
	}

	gatewayID := "97c46619-1f53-493b-8638-00c899f30152"
	mux.HandleFunc(fmt.Sprintf("%s/%s", egressGatewaysBasePath, gatewayID), func(w http.ResponseWriter, r *http.Request) {
		req := new(EgressGatewayRequest)
		err := json.NewDecoder(r.Body).Decode(req)
		if err != nil {
			t.Fatal(err)
		}
		testMethod(t, r, http.MethodPut)
		assert.Equal(t, updateReq, req)
		fmt.Fprintf(w, egressGatewayUpdateJSONResponse)
	})

	expectedGatewayResp := &EgressGateway{
		ID:     "97c46619-1f53-493b-8638-00c899f30152",
		Name:   "test-egress-gateway-renamed-01",
		Type:   "PUBLIC",
		State:  "STATE_ACTIVE",
		Region: "nyc3",
		VPCs: []*IngressVPC{
			{VpcUUID: "4637280e-3842-4661-a628-a6f0392959d3", GatewayIP: "10.100.0.110"},
		},
		Egresses: &Egresses{
			PublicGateways: []*PublicGateway{
				{IPv4: "10.36.48.172"},
			},
		},
		UDPTimeoutSeconds:  30,
		ICMPTimeoutSeconds: 30,
		TCPTimeoutSeconds:  300,
	}

	updateGatewayResp, _, err := client.EgressGateways.Update(ctx, gatewayID, updateReq)
	require.NoError(t, err)
	require.NotEmpty(t, updateGatewayResp)
	expectedGatewayResp.CreatedAt = updateGatewayResp.CreatedAt
	expectedGatewayResp.UpdatedAt = updateGatewayResp.UpdatedAt
	assert.Equal(t, expectedGatewayResp, updateGatewayResp)
}

func TestEgressGateways_Delete(t *testing.T) {
	setup()
	defer teardown()

	gatewayID := "97c46619-1f53-493b-8638-00c899f30152"
	mux.HandleFunc(fmt.Sprintf("%s/%s", egressGatewaysBasePath, gatewayID), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
	})

	_, err := client.EgressGateways.Delete(ctx, gatewayID)
	assert.NoError(t, err)
}
