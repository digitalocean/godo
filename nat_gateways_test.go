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

var natGatewayGetJSONResponse = `
{
  "nat_gateway": {
    "id": "97c46619-1f53-493b-8638-00c899f30152",
    "name": "test-nat-gateway-01",
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

var natGatewayListJSONResponse = `
{
  "nat_gateways": [
    {
      "id": "97c46619-1f53-493b-8638-00c899f30152",
      "name": "test-nat-gateway-01",
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
      "name": "test-nat-gateway-02",
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

var natGatewayUpdateJSONResponse = `
{
  "nat_gateway": {
    "id": "97c46619-1f53-493b-8638-00c899f30152",
    "name": "test-nat-gateway-renamed-01",
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

func TestNatGateway_Create(t *testing.T) {
	setup()
	defer teardown()

	createReq := &NatGatewayRequest{
		Name:   "test-nat-gateway-01",
		Type:   "PUBLIC",
		Region: "nyc3",
		VPCs: []*IngressVPC{
			{VpcUUID: "4637280e-3842-4661-a628-a6f0392959d3"},
		},
		UDPTimeoutSeconds:  30,
		ICMPTimeoutSeconds: 30,
		TCPTimeoutSeconds:  300,
	}

	mux.HandleFunc(natGatewaysBasePath, func(w http.ResponseWriter, r *http.Request) {
		req := new(NatGatewayRequest)
		err := json.NewDecoder(r.Body).Decode(req)
		if err != nil {
			t.Fatal(err)
		}
		testMethod(t, r, http.MethodPost)
		assert.Equal(t, createReq, req)
		fmt.Fprintf(w, natGatewayGetJSONResponse)
	})

	expectedGatewayResp := &NatGateway{
		ID:     "97c46619-1f53-493b-8638-00c899f30152",
		Name:   "test-nat-gateway-01",
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

	createGatewayResp, _, err := client.NatGateways.Create(ctx, createReq)
	require.NoError(t, err)
	require.NotEmpty(t, createGatewayResp)
	expectedGatewayResp.CreatedAt = createGatewayResp.CreatedAt
	expectedGatewayResp.UpdatedAt = createGatewayResp.UpdatedAt
	assert.Equal(t, expectedGatewayResp, createGatewayResp)
}

func TestNatGateway_Get(t *testing.T) {
	setup()
	defer teardown()

	gatewayID := "97c46619-1f53-493b-8638-00c899f30152"
	mux.HandleFunc(fmt.Sprintf("%s/%s", natGatewaysBasePath, gatewayID), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprintf(w, natGatewayGetJSONResponse)
	})

	expectedGatewayResp := &NatGateway{
		ID:     "97c46619-1f53-493b-8638-00c899f30152",
		Name:   "test-nat-gateway-01",
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

	getGatewayResp, _, err := client.NatGateways.Get(ctx, gatewayID)
	require.NoError(t, err)
	require.NotNil(t, getGatewayResp)
	expectedGatewayResp.CreatedAt = getGatewayResp.CreatedAt
	expectedGatewayResp.UpdatedAt = getGatewayResp.UpdatedAt
	assert.Equal(t, expectedGatewayResp, getGatewayResp)
}

func TestNatGateway_List(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc(natGatewaysBasePath, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprintf(w, natGatewayListJSONResponse)
	})

	expectedGatewaysResp := []*NatGateway{
		{
			ID:     "97c46619-1f53-493b-8638-00c899f30152",
			Name:   "test-nat-gateway-01",
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
			Name:   "test-nat-gateway-02",
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

	listGatewaysResp, _, err := client.NatGateways.List(ctx, nil)
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

func TestNatGateway_Update(t *testing.T) {
	setup()
	defer teardown()

	updateReq := &NatGatewayRequest{
		Name:   "test-nat-gateway-renamed-01",
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
	mux.HandleFunc(fmt.Sprintf("%s/%s", natGatewaysBasePath, gatewayID), func(w http.ResponseWriter, r *http.Request) {
		req := new(NatGatewayRequest)
		err := json.NewDecoder(r.Body).Decode(req)
		if err != nil {
			t.Fatal(err)
		}
		testMethod(t, r, http.MethodPut)
		assert.Equal(t, updateReq, req)
		fmt.Fprintf(w, natGatewayUpdateJSONResponse)
	})

	expectedGatewayResp := &NatGateway{
		ID:     "97c46619-1f53-493b-8638-00c899f30152",
		Name:   "test-nat-gateway-renamed-01",
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

	updateGatewayResp, _, err := client.NatGateways.Update(ctx, gatewayID, updateReq)
	require.NoError(t, err)
	require.NotEmpty(t, updateGatewayResp)
	expectedGatewayResp.CreatedAt = updateGatewayResp.CreatedAt
	expectedGatewayResp.UpdatedAt = updateGatewayResp.UpdatedAt
	assert.Equal(t, expectedGatewayResp, updateGatewayResp)
}

func TestNatGateway_Delete(t *testing.T) {
	setup()
	defer teardown()

	gatewayID := "97c46619-1f53-493b-8638-00c899f30152"
	mux.HandleFunc(fmt.Sprintf("%s/%s", natGatewaysBasePath, gatewayID), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
	})

	_, err := client.NatGateways.Delete(ctx, gatewayID)
	assert.NoError(t, err)
}
