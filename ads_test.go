package godo

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	adsTestRegion = "tor1"
	adsTestUUID   = "ws-019fb39c-14d9-7080-933e-b9b90e25acda"
)

var adsWorkspace = ADSWorkspace{
	UUID:              adsTestUUID,
	OwnerUUID:         "7f3c1e52-9a4b-4d6e-8c21-5b0a9d3e6f14",
	Name:              "analytics",
	Region:            adsTestRegion,
	Status:            "active",
	QueryEndpoint:     "https://query.example.internal",
	QueryInstanceSize: "small",
}

const adsWorkspaceJSON = `{
	"uuid": "ws-019fb39c-14d9-7080-933e-b9b90e25acda",
	"owner_uuid": "7f3c1e52-9a4b-4d6e-8c21-5b0a9d3e6f14",
	"name": "analytics",
	"region": "tor1",
	"status": "active",
	"query_endpoint": "https://query.example.internal",
	"query_instance_size": "small"
}`

func TestADSTokenCreateRequest_JSON(t *testing.T) {
	raw, err := json.Marshal(ADSTokenCreateRequest{
		Region:           "tor1",
		Purpose:          ADSTokenPurposeQuery,
		WorkspaceUUIDs:   []string{adsTestUUID},
		Name:             "analytics",
		ExpiresInSeconds: 3600,
		Permissions:      []string{"read"},
	})
	require.NoError(t, err)
	assert.JSONEq(t, `{
		"region": "tor1",
		"purpose": "TOKEN_PURPOSE_QUERY",
		"workspace_uuids": ["ws-019fb39c-14d9-7080-933e-b9b90e25acda"],
		"name": "analytics",
		"expires_in_seconds": "3600",
		"permissions": ["read"]
	}`, string(raw))
	assert.NotContains(t, string(raw), "sandbox_uuids")
}

func TestADSTokenCreateRequest_OmitsOptionalFields(t *testing.T) {
	raw, err := json.Marshal(ADSTokenCreateRequest{Region: "tor1"})
	require.NoError(t, err)
	assert.JSONEq(t, `{"region":"tor1"}`, string(raw))
}

func TestADS_ListPlans(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/data-cloud-platform/plans", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		assert.Empty(t, r.URL.RawQuery)
		assert.Equal(t, "/v2/data-cloud-platform/plans", r.URL.EscapedPath())
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{
			"plans": [
				{
					"region_slug": "tor1",
					"available_for_create": true,
					"query_engine_sizes": [
						{"size": "small", "droplet_slug": "s-1vcpu-2gb", "cpu": 1, "memory_gi": 2}
					]
				},
				{
					"region_slug": "nyc3",
					"available_for_create": false
				}
			]
		}`)
	})

	got, resp, err := client.ADS.ListPlans(ctx)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, &ADSPlansResponse{
		Plans: []ADSPlan{
			{
				RegionSlug:         "tor1",
				AvailableForCreate: true,
				QueryEngineSizes: []ADSQueryEngineSize{{
					Size:        "small",
					DropletSlug: "s-1vcpu-2gb",
					CPU:         1,
					MemoryGi:    2,
				}},
			},
			{
				RegionSlug:         "nyc3",
				AvailableForCreate: false,
			},
		},
	}, got)
}

func TestADS_ListPlans_Error(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/data-cloud-platform/plans", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"message":"plans unavailable","request_id":"req-plans"}`)
	})

	got, resp, err := client.ADS.ListPlans(ctx)
	require.Error(t, err)
	assert.Nil(t, got)
	require.NotNil(t, resp)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	var apiErr *ErrorResponse
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, "plans unavailable", apiErr.Message)
	assert.Equal(t, "req-plans", apiErr.RequestID)
}

func TestADS_CreateWorkspace(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/data-cloud-platform/tor1/workspaces", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		assert.Equal(t, "/v2/data-cloud-platform/tor1/workspaces", r.URL.EscapedPath())
		assert.Empty(t, r.URL.RawQuery)
		raw, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		assert.JSONEq(t, `{"name":"analytics","query_instance_size":"medium"}`, string(raw))
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"workspace":%s}`, adsWorkspaceJSON)
	})

	got, resp, err := client.ADS.CreateWorkspace(ctx, adsTestRegion, &ADSWorkspaceCreateRequest{
		Name:              "analytics",
		QueryInstanceSize: "medium",
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, &adsWorkspace, got)
}

func TestADS_CreateWorkspace_OmitsEmptyFields(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/data-cloud-platform/tor1/workspaces", func(w http.ResponseWriter, r *http.Request) {
		raw, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		assert.JSONEq(t, `{}`, string(raw))
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"workspace":%s}`, adsWorkspaceJSON)
	})

	got, resp, err := client.ADS.CreateWorkspace(ctx, adsTestRegion, &ADSWorkspaceCreateRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, &adsWorkspace, got)
}

func TestADS_CreateWorkspace_EscapesRegion(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/v2/data-cloud-platform/tor1%2Fother/workspaces", r.URL.EscapedPath())
		assert.Empty(t, r.URL.RawQuery)
		raw, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		assert.JSONEq(t, `{"name":"analytics"}`, string(raw))
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"workspace":%s}`, adsWorkspaceJSON)
	})

	got, resp, err := client.ADS.CreateWorkspace(ctx, "tor1/other", &ADSWorkspaceCreateRequest{
		Name: "analytics",
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, &adsWorkspace, got)
}

func TestADS_CreateWorkspace_Error(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/data-cloud-platform/tor1/workspaces", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"message":"region not found","request_id":"req-create"}`)
	})

	got, resp, err := client.ADS.CreateWorkspace(ctx, adsTestRegion, &ADSWorkspaceCreateRequest{Name: "analytics"})
	require.Error(t, err)
	assert.Nil(t, got)
	require.NotNil(t, resp)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	var apiErr *ErrorResponse
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, "region not found", apiErr.Message)
}

func TestADS_CreateWorkspace_RequiresRequest(t *testing.T) {
	setup()
	defer teardown()

	got, resp, err := client.ADS.CreateWorkspace(ctx, adsTestRegion, nil)
	require.EqualError(t, err, "ads: create request is required")
	assert.Nil(t, got)
	assert.Nil(t, resp)
}

func TestADS_CreateWorkspace_RequiresRegion(t *testing.T) {
	setup()
	defer teardown()

	got, resp, err := client.ADS.CreateWorkspace(ctx, "", &ADSWorkspaceCreateRequest{Name: "analytics"})
	require.EqualError(t, err, "ads: region is required")
	assert.Nil(t, got)
	assert.Nil(t, resp)
}

func TestADS_CreateWorkspace_EmptyWorkspace(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/data-cloud-platform/tor1/workspaces", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{}`)
	})

	got, resp, err := client.ADS.CreateWorkspace(ctx, adsTestRegion, &ADSWorkspaceCreateRequest{Name: "analytics"})
	require.EqualError(t, err, "ads: create workspace returned no workspace")
	assert.Nil(t, got)
	require.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestADS_ListWorkspaces(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/data-cloud-platform/tor1/workspaces", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		assert.Equal(t, "/v2/data-cloud-platform/tor1/workspaces", r.URL.EscapedPath())
		assert.Empty(t, r.URL.RawQuery)
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"workspaces":[%s]}`, adsWorkspaceJSON)
	})

	got, resp, err := client.ADS.ListWorkspaces(ctx, adsTestRegion)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, &ADSWorkspacesResponse{Workspaces: []ADSWorkspace{adsWorkspace}}, got)
}

func TestADS_ListWorkspaces_EscapesRegion(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/v2/data-cloud-platform/tor1%2Fother/workspaces", r.URL.EscapedPath())
		assert.Empty(t, r.URL.RawQuery)
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"workspaces":[]}`)
	})

	got, resp, err := client.ADS.ListWorkspaces(ctx, "tor1/other")
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, &ADSWorkspacesResponse{Workspaces: []ADSWorkspace{}}, got)
}

func TestADS_ListWorkspaces_Error(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/data-cloud-platform/tor1/workspaces", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"message":"region not found","request_id":"req-list"}`)
	})

	got, resp, err := client.ADS.ListWorkspaces(ctx, adsTestRegion)
	require.Error(t, err)
	assert.Nil(t, got)
	require.NotNil(t, resp)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	var apiErr *ErrorResponse
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, "region not found", apiErr.Message)
}

func TestADS_ListWorkspaces_RequiresRegion(t *testing.T) {
	setup()
	defer teardown()

	got, resp, err := client.ADS.ListWorkspaces(ctx, "")
	require.EqualError(t, err, "ads: region is required")
	assert.Nil(t, got)
	assert.Nil(t, resp)
}

func TestADS_GetWorkspace(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/data-cloud-platform/tor1/workspaces/"+adsTestUUID, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		assert.Equal(t, "/v2/data-cloud-platform/tor1/workspaces/"+adsTestUUID, r.URL.EscapedPath())
		assert.Empty(t, r.URL.RawQuery)
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"workspace":%s}`, adsWorkspaceJSON)
	})

	got, resp, err := client.ADS.GetWorkspace(ctx, adsTestRegion, adsTestUUID)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, &adsWorkspace, got)
}

func TestADS_GetWorkspace_EscapesPath(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/v2/data-cloud-platform/tor1%2Fother/workspaces/abc%2Fdef", r.URL.EscapedPath())
		assert.Empty(t, r.URL.RawQuery)
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"workspace":%s}`, adsWorkspaceJSON)
	})

	got, resp, err := client.ADS.GetWorkspace(ctx, "tor1/other", "abc/def")
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, &adsWorkspace, got)
}

func TestADS_GetWorkspace_Error(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/data-cloud-platform/tor1/workspaces/"+adsTestUUID, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"message":"workspace not found","request_id":"req-get"}`)
	})

	got, resp, err := client.ADS.GetWorkspace(ctx, adsTestRegion, adsTestUUID)
	require.Error(t, err)
	assert.Nil(t, got)
	require.NotNil(t, resp)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	var apiErr *ErrorResponse
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, "workspace not found", apiErr.Message)
}

func TestADS_GetWorkspace_Forbidden(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/data-cloud-platform/tor1/workspaces/"+adsTestUUID, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, `{"message":"forbidden","request_id":"req-forbidden"}`)
	})

	got, resp, err := client.ADS.GetWorkspace(ctx, adsTestRegion, adsTestUUID)
	require.Error(t, err)
	assert.Nil(t, got)
	require.NotNil(t, resp)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	var apiErr *ErrorResponse
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, "forbidden", apiErr.Message)
	assert.Equal(t, "req-forbidden", apiErr.RequestID)
}

func TestADS_GetWorkspace_RequiresIDs(t *testing.T) {
	setup()
	defer teardown()

	got, resp, err := client.ADS.GetWorkspace(ctx, "", adsTestUUID)
	require.EqualError(t, err, "ads: region is required")
	assert.Nil(t, got)
	assert.Nil(t, resp)

	got, resp, err = client.ADS.GetWorkspace(ctx, adsTestRegion, "")
	require.EqualError(t, err, "ads: workspace uuid is required")
	assert.Nil(t, got)
	assert.Nil(t, resp)
}

func TestADS_GetWorkspace_EmptyWorkspace(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/data-cloud-platform/tor1/workspaces/"+adsTestUUID, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{}`)
	})

	got, resp, err := client.ADS.GetWorkspace(ctx, adsTestRegion, adsTestUUID)
	require.EqualError(t, err, "ads: get workspace returned no workspace")
	assert.Nil(t, got)
	require.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestADS_DeleteWorkspace(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/data-cloud-platform/tor1/workspaces/"+adsTestUUID, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		assert.Equal(t, "/v2/data-cloud-platform/tor1/workspaces/"+adsTestUUID, r.URL.EscapedPath())
		assert.Empty(t, r.URL.RawQuery)
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{}`)
	})

	resp, err := client.ADS.DeleteWorkspace(ctx, adsTestRegion, adsTestUUID)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestADS_DeleteWorkspace_EscapesPath(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Equal(t, "/v2/data-cloud-platform/tor1%2Fother/workspaces/abc%2Fdef", r.URL.EscapedPath())
		assert.Empty(t, r.URL.RawQuery)
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{}`)
	})

	resp, err := client.ADS.DeleteWorkspace(ctx, "tor1/other", "abc/def")
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestADS_DeleteWorkspace_Error(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/data-cloud-platform/tor1/workspaces/"+adsTestUUID, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"message":"workspace not found","request_id":"req-delete"}`)
	})

	resp, err := client.ADS.DeleteWorkspace(ctx, adsTestRegion, adsTestUUID)
	require.Error(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	var apiErr *ErrorResponse
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, "workspace not found", apiErr.Message)
}

func TestADS_DeleteWorkspace_RequiresIDs(t *testing.T) {
	setup()
	defer teardown()

	resp, err := client.ADS.DeleteWorkspace(ctx, "", adsTestUUID)
	require.EqualError(t, err, "ads: region is required")
	assert.Nil(t, resp)

	resp, err = client.ADS.DeleteWorkspace(ctx, adsTestRegion, "")
	require.EqualError(t, err, "ads: workspace uuid is required")
	assert.Nil(t, resp)
}

func TestADS_CreateToken(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/data-cloud-platform/tor1/tokens", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		assert.Equal(t, "/v2/data-cloud-platform/tor1/tokens", r.URL.EscapedPath())
		assert.Empty(t, r.URL.RawQuery)
		raw, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		assert.JSONEq(t, `{
			"region": "tor1",
			"purpose": "TOKEN_PURPOSE_QUERY",
			"workspace_uuids": ["ws-019fb39c-14d9-7080-933e-b9b90e25acda"],
			"name": "analytics",
			"expires_in_seconds": "3600",
			"permissions": ["read"]
		}`, string(raw))
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{
			"id": "tok-019fb39c-14d9-7080-933e-b9b90e25acda",
			"name": "analytics",
			"access_token": "secret_value",
			"expires_at": "2026-11-07T00:00:00Z"
		}`)
	})

	got, resp, err := client.ADS.CreateToken(ctx, &ADSTokenCreateRequest{
		Region:           adsTestRegion,
		Purpose:          ADSTokenPurposeQuery,
		WorkspaceUUIDs:   []string{adsTestUUID},
		Name:             "analytics",
		ExpiresInSeconds: 3600,
		Permissions:      []string{"read"},
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, &ADSTokenCreateResponse{
		ID:          "tok-019fb39c-14d9-7080-933e-b9b90e25acda",
		Name:        "analytics",
		AccessToken: "secret_value",
		ExpiresAt:   "2026-11-07T00:00:00Z",
	}, got)
}

func TestADS_CreateToken_EscapesRegion(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/v2/data-cloud-platform/tor1%2Fother/tokens", r.URL.EscapedPath())
		assert.Empty(t, r.URL.RawQuery)
		raw, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		assert.JSONEq(t, `{"region":"tor1/other"}`, string(raw))
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{
			"id": "tok-019fb39c-14d9-7080-933e-b9b90e25acda",
			"access_token": "secret_value"
		}`)
	})

	got, resp, err := client.ADS.CreateToken(ctx, &ADSTokenCreateRequest{
		Region: "tor1/other",
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, &ADSTokenCreateResponse{
		ID:          "tok-019fb39c-14d9-7080-933e-b9b90e25acda",
		AccessToken: "secret_value",
	}, got)
}

func TestADS_CreateToken_Error(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/data-cloud-platform/tor1/tokens", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"message":"region not found","request_id":"req-create-token"}`)
	})

	got, resp, err := client.ADS.CreateToken(ctx, &ADSTokenCreateRequest{
		Region: adsTestRegion,
	})
	require.Error(t, err)
	assert.Nil(t, got)
	require.NotNil(t, resp)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	var apiErr *ErrorResponse
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, "region not found", apiErr.Message)
	assert.Equal(t, "req-create-token", apiErr.RequestID)
}

func TestADS_CreateToken_Forbidden(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/data-cloud-platform/tor1/tokens", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, `{"message":"forbidden","request_id":"req-forbidden"}`)
	})

	got, resp, err := client.ADS.CreateToken(ctx, &ADSTokenCreateRequest{
		Region: adsTestRegion,
	})
	require.Error(t, err)
	assert.Nil(t, got)
	require.NotNil(t, resp)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	var apiErr *ErrorResponse
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, "forbidden", apiErr.Message)
	assert.Equal(t, "req-forbidden", apiErr.RequestID)
}

func TestADS_ListTokens(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/data-cloud-platform/tor1/tokens", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		assert.Equal(t, "/v2/data-cloud-platform/tor1/tokens", r.URL.EscapedPath())
		q := r.URL.Query()
		assert.Equal(t, "TOKEN_PURPOSE_QUERY", q.Get("purpose"))
		assert.Equal(t, adsTestUUID, q.Get("workspace_uuid"))
		assert.Empty(t, q.Get("sandbox_uuid"))
		assert.Equal(t, "2", q.Get("page"))
		assert.Equal(t, "10", q.Get("per_page"))
		assert.Empty(t, q.Get("region"))
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{
			"tokens": [
				{
					"id": "tok-019fb39c-14d9-7080-933e-b9b90e25acda",
					"name": "analytics",
					"created_at": "2026-11-06T00:00:00Z",
					"expires_at": "2026-11-07T00:00:00Z",
					"purpose": "TOKEN_PURPOSE_QUERY",
					"query_hosts": ["query.example.internal"],
					"workspace_uuids": ["ws-019fb39c-14d9-7080-933e-b9b90e25acda"]
				}
			],
			"total": "1"
		}`)
	})

	got, resp, err := client.ADS.ListTokens(ctx, &ADSTokenListRequest{
		Region:        adsTestRegion,
		Purpose:       ADSTokenPurposeQuery,
		Page:          2,
		PerPage:       10,
		WorkspaceUUID: adsTestUUID,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, &ADSListTokensResponse{
		Tokens: []ADSTokenSummary{{
			ID:             "tok-019fb39c-14d9-7080-933e-b9b90e25acda",
			Name:           "analytics",
			CreatedAt:      "2026-11-06T00:00:00Z",
			ExpiresAt:      "2026-11-07T00:00:00Z",
			Purpose:        ADSTokenPurposeQuery,
			QueryHosts:     []string{"query.example.internal"},
			WorkspaceUUIDs: []string{adsTestUUID},
		}},
		Total: "1",
	}, got)
}

func TestADS_ListTokens_EscapesPathAndQuery(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/v2/data-cloud-platform/tor1%2Fother/tokens", r.URL.EscapedPath())
		assert.Equal(t, "ws/abc", r.URL.Query().Get("workspace_uuid"))
		assert.Contains(t, r.URL.RawQuery, "workspace_uuid=ws%2Fabc")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"tokens":[],"total":"0"}`)
	})

	got, resp, err := client.ADS.ListTokens(ctx, &ADSTokenListRequest{
		Region:        "tor1/other",
		WorkspaceUUID: "ws/abc",
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, &ADSListTokensResponse{
		Tokens: []ADSTokenSummary{},
		Total:  "0",
	}, got)
}

func TestADS_ListTokens_OmitsEmptyFilters(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/data-cloud-platform/tor1/tokens", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		assert.Equal(t, "/v2/data-cloud-platform/tor1/tokens", r.URL.EscapedPath())
		assert.Empty(t, r.URL.RawQuery)
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"tokens":[],"total":"0"}`)
	})

	got, resp, err := client.ADS.ListTokens(ctx, &ADSTokenListRequest{
		Region: adsTestRegion,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, &ADSListTokensResponse{
		Tokens: []ADSTokenSummary{},
		Total:  "0",
	}, got)
}

func TestADS_ListTokens_Error(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/data-cloud-platform/tor1/tokens", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"message":"region not found","request_id":"req-list-tokens"}`)
	})

	got, resp, err := client.ADS.ListTokens(ctx, &ADSTokenListRequest{
		Region: adsTestRegion,
	})
	require.Error(t, err)
	assert.Nil(t, got)
	require.NotNil(t, resp)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	var apiErr *ErrorResponse
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, "region not found", apiErr.Message)
	assert.Equal(t, "req-list-tokens", apiErr.RequestID)
}

func TestADS_ListTokens_RequiresRequest(t *testing.T) {
	setup()
	defer teardown()

	got, resp, err := client.ADS.ListTokens(ctx, nil)
	require.EqualError(t, err, "ads: list request is required")
	assert.Nil(t, got)
	assert.Nil(t, resp)
}

func TestADS_ListTokens_RequiresRegion(t *testing.T) {
	setup()
	defer teardown()

	got, resp, err := client.ADS.ListTokens(ctx, &ADSTokenListRequest{})
	require.EqualError(t, err, "ads: region is required")
	assert.Nil(t, got)
	assert.Nil(t, resp)
}

func TestADS_PutWorkspaceSecret(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/data-cloud-platform/tor1/workspaces/"+adsTestUUID+"/secrets/db-password", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		assert.Equal(t, "/v2/data-cloud-platform/tor1/workspaces/"+adsTestUUID+"/secrets/db-password", r.URL.EscapedPath())
		assert.Empty(t, r.URL.RawQuery)
		raw, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		assert.JSONEq(t, `{
			"region": "tor1",
			"uuid": "ws-019fb39c-14d9-7080-933e-b9b90e25acda",
			"secret_key": "db-password",
			"data": {"password": "secret_value"},
			"options": {"cas": 1}
		}`, string(raw))
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{
			"data": {
				"created_time": "2026-11-07T00:00:00Z",
				"custom_metadata": {"owner": "analytics"},
				"deletion_time": "",
				"destroyed": false,
				"version": 3
			}
		}`)
	})

	got, resp, err := client.ADS.PutWorkspaceSecret(ctx, &ADSPutWorkspaceSecretRequest{
		Region:    adsTestRegion,
		UUID:      adsTestUUID,
		SecretKey: "db-password",
		Data:      map[string]string{"password": "secret_value"},
		Options:   &ADSSecretOptions{CAS: 1},
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, &ADSPutWorkspaceSecretResponse{
		Data: &ADSSecretMetadata{
			CreatedTime:    "2026-11-07T00:00:00Z",
			CustomMetadata: map[string]any{"owner": "analytics"},
			DeletionTime:   "",
			Destroyed:      false,
			Version:        3,
		},
	}, got)
}

func TestADS_PutWorkspaceSecret_EscapesPath(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		assert.Equal(t, "/v2/data-cloud-platform/tor1%2Fother/workspaces/ws%2Fabc/secrets/db%2Fpassword", r.URL.EscapedPath())
		assert.Empty(t, r.URL.RawQuery)
		raw, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		assert.JSONEq(t, `{
			"region": "tor1/other",
			"uuid": "ws/abc",
			"secret_key": "db/password",
			"data": {"password": "secret_value"}
		}`, string(raw))
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{
			"data": {
				"created_time": "2026-11-07T00:00:00Z",
				"destroyed": false,
				"version": 1
			}
		}`)
	})

	got, resp, err := client.ADS.PutWorkspaceSecret(ctx, &ADSPutWorkspaceSecretRequest{
		Region:    "tor1/other",
		UUID:      "ws/abc",
		SecretKey: "db/password",
		Data:      map[string]string{"password": "secret_value"},
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, &ADSPutWorkspaceSecretResponse{
		Data: &ADSSecretMetadata{
			CreatedTime: "2026-11-07T00:00:00Z",
			Destroyed:   false,
			Version:     1,
		},
	}, got)
}

func TestADS_PutWorkspaceSecret_Error(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/data-cloud-platform/tor1/workspaces/"+adsTestUUID+"/secrets/db-password", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"message":"workspace not found","request_id":"req-put-secret"}`)
	})

	got, resp, err := client.ADS.PutWorkspaceSecret(ctx, &ADSPutWorkspaceSecretRequest{
		Region:    adsTestRegion,
		UUID:      adsTestUUID,
		SecretKey: "db-password",
		Data:      map[string]string{"password": "secret_value"},
	})
	require.Error(t, err)
	assert.Nil(t, got)
	require.NotNil(t, resp)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	var apiErr *ErrorResponse
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, "workspace not found", apiErr.Message)
	assert.Equal(t, "req-put-secret", apiErr.RequestID)
}

func TestADS_PutWorkspaceSecret_RequiresRequest(t *testing.T) {
	setup()
	defer teardown()

	got, resp, err := client.ADS.PutWorkspaceSecret(ctx, nil)
	require.EqualError(t, err, "ads: put request is required")
	assert.Nil(t, got)
	assert.Nil(t, resp)
}

func TestADS_PutWorkspaceSecret_RequiresData(t *testing.T) {
	setup()
	defer teardown()

	got, resp, err := client.ADS.PutWorkspaceSecret(ctx, &ADSPutWorkspaceSecretRequest{
		Region:    adsTestRegion,
		UUID:      adsTestUUID,
		SecretKey: "db-password",
	})
	require.EqualError(t, err, "ads: secret data is required")
	assert.Nil(t, got)
	assert.Nil(t, resp)

	got, resp, err = client.ADS.PutWorkspaceSecret(ctx, &ADSPutWorkspaceSecretRequest{
		Region:    adsTestRegion,
		UUID:      adsTestUUID,
		SecretKey: "db-password",
		Data:      map[string]string{},
	})
	require.EqualError(t, err, "ads: secret data is required")
	assert.Nil(t, got)
	assert.Nil(t, resp)
}

func TestADS_PutWorkspaceSecret_RequiresPathArgs(t *testing.T) {
	setup()
	defer teardown()

	got, resp, err := client.ADS.PutWorkspaceSecret(ctx, &ADSPutWorkspaceSecretRequest{
		UUID:      adsTestUUID,
		SecretKey: "db-password",
		Data:      map[string]string{"password": "secret_value"},
	})
	require.EqualError(t, err, "ads: region is required")
	assert.Nil(t, got)
	assert.Nil(t, resp)

	got, resp, err = client.ADS.PutWorkspaceSecret(ctx, &ADSPutWorkspaceSecretRequest{
		Region:    adsTestRegion,
		SecretKey: "db-password",
		Data:      map[string]string{"password": "secret_value"},
	})
	require.EqualError(t, err, "ads: workspace uuid is required")
	assert.Nil(t, got)
	assert.Nil(t, resp)

	got, resp, err = client.ADS.PutWorkspaceSecret(ctx, &ADSPutWorkspaceSecretRequest{
		Region: adsTestRegion,
		UUID:   adsTestUUID,
		Data:   map[string]string{"password": "secret_value"},
	})
	require.EqualError(t, err, "ads: secret key is required")
	assert.Nil(t, got)
	assert.Nil(t, resp)
}
