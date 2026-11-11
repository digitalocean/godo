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

const adsTestDataFlowID = "df-019fb39c-14d9-7080-933e-b9b90e25acda"

func TestADS_GetWorkspaceSecret(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/data-cloud-platform/tor1/workspaces/"+adsTestUUID+"/secrets/db-password", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		assert.Equal(t, "/v2/data-cloud-platform/tor1/workspaces/"+adsTestUUID+"/secrets/db-password", r.URL.EscapedPath())
		assert.Empty(t, r.URL.RawQuery)
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{
			"data": {
				"data": {},
				"metadata": {
					"created_time": "2026-11-07T00:00:00Z",
					"custom_metadata": {"owner": "analytics"},
					"destroyed": false,
					"version": 3
				},
				"keys": ["password", "username"]
			}
		}`)
	})

	got, resp, err := client.ADS.GetWorkspaceSecret(ctx, adsTestRegion, adsTestUUID, "db-password")
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, &ADSGetWorkspaceSecretResponse{
		Data: &ADSGetSecretData{
			Data: map[string]any{},
			Metadata: &ADSSecretMetadata{
				CreatedTime:    "2026-11-07T00:00:00Z",
				CustomMetadata: map[string]any{"owner": "analytics"},
				Destroyed:      false,
				Version:        3,
			},
			Keys: []string{"password", "username"},
		},
	}, got)
}

func TestADS_GetWorkspaceSecret_EscapesPath(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/v2/data-cloud-platform/tor1%2Fother/workspaces/ws%2Fabc/secrets/db%2Fpassword", r.URL.EscapedPath())
		assert.Empty(t, r.URL.RawQuery)
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"data":{"keys":[]}}`)
	})

	got, resp, err := client.ADS.GetWorkspaceSecret(ctx, "tor1/other", "ws/abc", "db/password")
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, &ADSGetWorkspaceSecretResponse{
		Data: &ADSGetSecretData{Keys: []string{}},
	}, got)
}

func TestADS_GetWorkspaceSecret_Error(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/data-cloud-platform/tor1/workspaces/"+adsTestUUID+"/secrets/db-password", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"message":"secret not found","request_id":"req-get-secret"}`)
	})

	got, resp, err := client.ADS.GetWorkspaceSecret(ctx, adsTestRegion, adsTestUUID, "db-password")
	require.Error(t, err)
	assert.Nil(t, got)
	require.NotNil(t, resp)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	var apiErr *ErrorResponse
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, "secret not found", apiErr.Message)
}

func TestADS_GetWorkspaceSecret_Forbidden(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/data-cloud-platform/tor1/workspaces/"+adsTestUUID+"/secrets/db-password", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, `{"message":"forbidden","request_id":"req-forbidden"}`)
	})

	got, resp, err := client.ADS.GetWorkspaceSecret(ctx, adsTestRegion, adsTestUUID, "db-password")
	require.Error(t, err)
	assert.Nil(t, got)
	require.NotNil(t, resp)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	var apiErr *ErrorResponse
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, "forbidden", apiErr.Message)
	assert.Equal(t, "req-forbidden", apiErr.RequestID)
}

func TestADS_ListWorkspaceSecrets(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/data-cloud-platform/tor1/workspaces/"+adsTestUUID+"/secrets", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		assert.Equal(t, "/v2/data-cloud-platform/tor1/workspaces/"+adsTestUUID+"/secrets", r.URL.EscapedPath())
		assert.Empty(t, r.URL.RawQuery)
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"secret_keys":["db-password","api-token"]}`)
	})

	got, resp, err := client.ADS.ListWorkspaceSecrets(ctx, adsTestRegion, adsTestUUID)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, &ADSListWorkspaceSecretsResponse{
		SecretKeys: []string{"db-password", "api-token"},
	}, got)
}

func TestADS_ListWorkspaceSecrets_EscapesPath(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/v2/data-cloud-platform/tor1%2Fother/workspaces/ws%2Fabc/secrets", r.URL.EscapedPath())
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"secret_keys":[]}`)
	})

	got, resp, err := client.ADS.ListWorkspaceSecrets(ctx, "tor1/other", "ws/abc")
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, &ADSListWorkspaceSecretsResponse{SecretKeys: []string{}}, got)
}

func TestADS_ListWorkspaceSecrets_Error(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/data-cloud-platform/tor1/workspaces/"+adsTestUUID+"/secrets", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"message":"workspace not found","request_id":"req-list-secrets"}`)
	})

	got, resp, err := client.ADS.ListWorkspaceSecrets(ctx, adsTestRegion, adsTestUUID)
	require.Error(t, err)
	assert.Nil(t, got)
	require.NotNil(t, resp)
	var apiErr *ErrorResponse
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, "workspace not found", apiErr.Message)
}

func TestADS_DeleteWorkspaceSecret(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/data-cloud-platform/tor1/workspaces/"+adsTestUUID+"/secrets/db-password", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		assert.Equal(t, "/v2/data-cloud-platform/tor1/workspaces/"+adsTestUUID+"/secrets/db-password", r.URL.EscapedPath())
		assert.Empty(t, r.URL.RawQuery)
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{}`)
	})

	resp, err := client.ADS.DeleteWorkspaceSecret(ctx, adsTestRegion, adsTestUUID, "db-password")
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestADS_DeleteWorkspaceSecret_EscapesPath(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Equal(t, "/v2/data-cloud-platform/tor1%2Fother/workspaces/ws%2Fabc/secrets/db%2Fpassword", r.URL.EscapedPath())
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{}`)
	})

	resp, err := client.ADS.DeleteWorkspaceSecret(ctx, "tor1/other", "ws/abc", "db/password")
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestADS_DeleteWorkspaceSecret_Error(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/data-cloud-platform/tor1/workspaces/"+adsTestUUID+"/secrets/db-password", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"message":"secret not found","request_id":"req-delete-secret"}`)
	})

	resp, err := client.ADS.DeleteWorkspaceSecret(ctx, adsTestRegion, adsTestUUID, "db-password")
	require.Error(t, err)
	require.NotNil(t, resp)
	var apiErr *ErrorResponse
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, "secret not found", apiErr.Message)
}

func TestADS_GetWorkspaceQueryCredentials(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/data-cloud-platform/tor1/workspaces/"+adsTestUUID+"/query/credentials", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		assert.Equal(t, "/v2/data-cloud-platform/tor1/workspaces/"+adsTestUUID+"/query/credentials", r.URL.EscapedPath())
		assert.Empty(t, r.URL.RawQuery)
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"user":"query_user","password":"secret_value"}`)
	})

	got, resp, err := client.ADS.GetWorkspaceQueryCredentials(ctx, adsTestRegion, adsTestUUID)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, &ADSQueryCredentials{
		User:     "query_user",
		Password: "secret_value",
	}, got)
}

func TestADS_GetWorkspaceQueryCredentials_EscapesPath(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/v2/data-cloud-platform/tor1%2Fother/workspaces/ws%2Fabc/query/credentials", r.URL.EscapedPath())
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"user":"query_user","password":"secret_value"}`)
	})

	got, resp, err := client.ADS.GetWorkspaceQueryCredentials(ctx, "tor1/other", "ws/abc")
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, &ADSQueryCredentials{User: "query_user", Password: "secret_value"}, got)
}

func TestADS_GetWorkspaceQueryCredentials_Error(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/data-cloud-platform/tor1/workspaces/"+adsTestUUID+"/query/credentials", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"message":"workspace not found","request_id":"req-query-creds"}`)
	})

	got, resp, err := client.ADS.GetWorkspaceQueryCredentials(ctx, adsTestRegion, adsTestUUID)
	require.Error(t, err)
	assert.Nil(t, got)
	require.NotNil(t, resp)
	var apiErr *ErrorResponse
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, "workspace not found", apiErr.Message)
}

func TestADS_GetWorkspaceQueryCredentials_Forbidden(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/data-cloud-platform/tor1/workspaces/"+adsTestUUID+"/query/credentials", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, `{"message":"forbidden","request_id":"req-forbidden"}`)
	})

	got, resp, err := client.ADS.GetWorkspaceQueryCredentials(ctx, adsTestRegion, adsTestUUID)
	require.Error(t, err)
	assert.Nil(t, got)
	require.NotNil(t, resp)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	var apiErr *ErrorResponse
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, "forbidden", apiErr.Message)
	assert.Equal(t, "req-forbidden", apiErr.RequestID)
}

func TestADS_CreateDataFlow(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/data-cloud-platform/tor1/workspaces/"+adsTestUUID+"/data-flows", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		assert.Equal(t, "/v2/data-cloud-platform/tor1/workspaces/"+adsTestUUID+"/data-flows", r.URL.EscapedPath())
		assert.Empty(t, r.URL.RawQuery)
		raw, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		assert.JSONEq(t, `{
			"name": "ingest",
			"desired_replicas": 1,
			"max_replicas": 3,
			"workspace": "ws-019fb39c-14d9-7080-933e-b9b90e25acda",
			"image_ref": "registry.example/flow:1",
			"manifest": {"schema":"v1"}
		}`, string(raw))
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{
			"uuid": "df-019fb39c-14d9-7080-933e-b9b90e25acda",
			"state": "pending",
			"upload": {
				"url": "https://upload.example/secret_value",
				"object_key": "fchain/abc",
				"expires_at": "2026-11-07T01:00:00Z"
			}
		}`)
	})

	got, resp, err := client.ADS.CreateDataFlow(ctx, &ADSDataFlowCreateRequest{
		Region:          adsTestRegion,
		Name:            "ingest",
		DesiredReplicas: 1,
		MaxReplicas:     3,
		Workspace:       adsTestUUID,
		ImageRef:        "registry.example/flow:1",
		Manifest:        json.RawMessage(`{"schema":"v1"}`),
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, &ADSDataFlowCreateResponse{
		UUID:  adsTestDataFlowID,
		State: "pending",
		Upload: &ADSTarballUpload{
			URL:       "https://upload.example/secret_value",
			ObjectKey: "fchain/abc",
			ExpiresAt: "2026-11-07T01:00:00Z",
		},
	}, got)
}

func TestADS_CreateDataFlow_EscapesPath(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/v2/data-cloud-platform/tor1%2Fother/workspaces/ws%2Fabc/data-flows", r.URL.EscapedPath())
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"uuid":"df-1","state":"pending"}`)
	})

	got, resp, err := client.ADS.CreateDataFlow(ctx, &ADSDataFlowCreateRequest{
		Region:    "tor1/other",
		Workspace: "ws/abc",
		Name:      "ingest",
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, &ADSDataFlowCreateResponse{UUID: "df-1", State: "pending"}, got)
}

func TestADS_CreateDataFlow_Error(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/data-cloud-platform/tor1/workspaces/"+adsTestUUID+"/data-flows", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"message":"invalid manifest","request_id":"req-create-df"}`)
	})

	got, resp, err := client.ADS.CreateDataFlow(ctx, &ADSDataFlowCreateRequest{Region: adsTestRegion, Workspace: adsTestUUID})
	require.Error(t, err)
	assert.Nil(t, got)
	require.NotNil(t, resp)
	var apiErr *ErrorResponse
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, "invalid manifest", apiErr.Message)
}

func TestADS_ListDataFlows(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/data-cloud-platform/tor1/workspaces/"+adsTestUUID+"/data-flows", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		q := r.URL.Query()
		assert.Equal(t, "25", q.Get("page_size"))
		assert.Equal(t, "tok-1", q.Get("page_token"))
		assert.Empty(t, q.Get("workspace"))
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{
			"data_flows": [{
				"uuid": "df-019fb39c-14d9-7080-933e-b9b90e25acda",
				"name": "ingest",
				"state": "active",
				"workspace": "ws-019fb39c-14d9-7080-933e-b9b90e25acda",
				"desired_replicas": 1,
				"max_replicas": 3,
				"tarball": {"requested": "sha256:abc", "attached": "sha256:abc"},
				"desired_state": "running"
			}],
			"next_page_token": "tok-2"
		}`)
	})

	got, resp, err := client.ADS.ListDataFlows(ctx, &ADSDataFlowListRequest{
		Region:    adsTestRegion,
		Workspace: adsTestUUID,
		PageSize:  25,
		PageToken: "tok-1",
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, &ADSListDataFlowsResponse{
		DataFlows: []ADSDataFlow{{
			UUID:            adsTestDataFlowID,
			Name:            "ingest",
			State:           "active",
			Workspace:       adsTestUUID,
			DesiredReplicas: 1,
			MaxReplicas:     3,
			Tarball:         &ADSTarballStatus{Requested: "sha256:abc", Attached: "sha256:abc"},
			DesiredState:    "running",
		}},
		NextPageToken: "tok-2",
	}, got)
}

func TestADS_ListDataFlows_EscapesPathAndQuery(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/v2/data-cloud-platform/tor1%2Fother/workspaces/ws%2Fabc/data-flows", r.URL.EscapedPath())
		assert.Equal(t, "a/b", r.URL.Query().Get("page_token"))
		assert.Contains(t, r.URL.RawQuery, "page_token=a%2Fb")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"data_flows":[]}`)
	})

	got, resp, err := client.ADS.ListDataFlows(ctx, &ADSDataFlowListRequest{
		Region:    "tor1/other",
		Workspace: "ws/abc",
		PageToken: "a/b",
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, &ADSListDataFlowsResponse{DataFlows: []ADSDataFlow{}}, got)
}

func TestADS_ListDataFlows_Error(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/data-cloud-platform/tor1/workspaces/"+adsTestUUID+"/data-flows", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"message":"workspace not found","request_id":"req-list-df"}`)
	})

	got, resp, err := client.ADS.ListDataFlows(ctx, &ADSDataFlowListRequest{Region: adsTestRegion, Workspace: adsTestUUID})
	require.Error(t, err)
	assert.Nil(t, got)
	require.NotNil(t, resp)
	var apiErr *ErrorResponse
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, "workspace not found", apiErr.Message)
}

func TestADS_GetDataFlow(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/data-cloud-platform/tor1/workspaces/"+adsTestUUID+"/data-flows/"+adsTestDataFlowID, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		assert.Empty(t, r.URL.RawQuery)
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{
			"uuid": "df-019fb39c-14d9-7080-933e-b9b90e25acda",
			"name": "ingest",
			"state": "active",
			"workspace": "ws-019fb39c-14d9-7080-933e-b9b90e25acda",
			"spec": {"schema":"v1"},
			"last_error": ""
		}`)
	})

	got, resp, err := client.ADS.GetDataFlow(ctx, adsTestRegion, adsTestUUID, adsTestDataFlowID)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, adsTestDataFlowID, got.UUID)
	assert.Equal(t, "ingest", got.Name)
	assert.Equal(t, "active", got.State)
	assert.JSONEq(t, `{"schema":"v1"}`, string(got.Spec))
}

func TestADS_GetDataFlow_EscapesPath(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/v2/data-cloud-platform/tor1%2Fother/workspaces/ws%2Fabc/data-flows/df%2F1", r.URL.EscapedPath())
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"uuid":"df/1","state":"active"}`)
	})

	got, resp, err := client.ADS.GetDataFlow(ctx, "tor1/other", "ws/abc", "df/1")
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "df/1", got.UUID)
}

func TestADS_GetDataFlow_Error(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/data-cloud-platform/tor1/workspaces/"+adsTestUUID+"/data-flows/"+adsTestDataFlowID, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"message":"data flow not found","request_id":"req-get-df"}`)
	})

	got, resp, err := client.ADS.GetDataFlow(ctx, adsTestRegion, adsTestUUID, adsTestDataFlowID)
	require.Error(t, err)
	assert.Nil(t, got)
	require.NotNil(t, resp)
	var apiErr *ErrorResponse
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, "data flow not found", apiErr.Message)
}

func TestADS_GetDataFlow_Forbidden(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/data-cloud-platform/tor1/workspaces/"+adsTestUUID+"/data-flows/"+adsTestDataFlowID, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, `{"message":"forbidden","request_id":"req-forbidden"}`)
	})

	got, resp, err := client.ADS.GetDataFlow(ctx, adsTestRegion, adsTestUUID, adsTestDataFlowID)
	require.Error(t, err)
	assert.Nil(t, got)
	require.NotNil(t, resp)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	var apiErr *ErrorResponse
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, "forbidden", apiErr.Message)
	assert.Equal(t, "req-forbidden", apiErr.RequestID)
}

func TestADS_UpdateDataFlow(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/data-cloud-platform/tor1/workspaces/"+adsTestUUID+"/data-flows/"+adsTestDataFlowID, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPatch)
		raw, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		assert.JSONEq(t, `{
			"id": "df-019fb39c-14d9-7080-933e-b9b90e25acda",
			"workspace": "ws-019fb39c-14d9-7080-933e-b9b90e25acda",
			"desired_replicas": 2,
			"secret_refs": ["db-password"]
		}`, string(raw))
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"uuid":"df-019fb39c-14d9-7080-933e-b9b90e25acda","state":"active"}`)
	})

	got, resp, err := client.ADS.UpdateDataFlow(ctx, &ADSDataFlowUpdateRequest{
		Region:          adsTestRegion,
		ID:              adsTestDataFlowID,
		Workspace:       adsTestUUID,
		DesiredReplicas: 2,
		SecretRefs:      []string{"db-password"},
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, &ADSDataFlowUpdateResponse{UUID: adsTestDataFlowID, State: "active"}, got)
}

func TestADS_UpdateDataFlow_EscapesPath(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPatch, r.Method)
		assert.Equal(t, "/v2/data-cloud-platform/tor1%2Fother/workspaces/ws%2Fabc/data-flows/df%2F1", r.URL.EscapedPath())
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"uuid":"df/1","state":"active"}`)
	})

	got, resp, err := client.ADS.UpdateDataFlow(ctx, &ADSDataFlowUpdateRequest{
		Region:    "tor1/other",
		ID:        "df/1",
		Workspace: "ws/abc",
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "df/1", got.UUID)
}

func TestADS_UpdateDataFlow_Error(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/data-cloud-platform/tor1/workspaces/"+adsTestUUID+"/data-flows/"+adsTestDataFlowID, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"message":"invalid patch","request_id":"req-patch-df"}`)
	})

	got, resp, err := client.ADS.UpdateDataFlow(ctx, &ADSDataFlowUpdateRequest{
		Region:    adsTestRegion,
		ID:        adsTestDataFlowID,
		Workspace: adsTestUUID,
	})
	require.Error(t, err)
	assert.Nil(t, got)
	require.NotNil(t, resp)
	var apiErr *ErrorResponse
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, "invalid patch", apiErr.Message)
}

func TestADS_DeleteDataFlow(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/data-cloud-platform/tor1/workspaces/"+adsTestUUID+"/data-flows/"+adsTestDataFlowID, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{}`)
	})

	resp, err := client.ADS.DeleteDataFlow(ctx, adsTestRegion, adsTestUUID, adsTestDataFlowID)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestADS_DeleteDataFlow_EscapesPath(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Equal(t, "/v2/data-cloud-platform/tor1%2Fother/workspaces/ws%2Fabc/data-flows/df%2F1", r.URL.EscapedPath())
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{}`)
	})

	resp, err := client.ADS.DeleteDataFlow(ctx, "tor1/other", "ws/abc", "df/1")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestADS_DeleteDataFlow_Error(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/data-cloud-platform/tor1/workspaces/"+adsTestUUID+"/data-flows/"+adsTestDataFlowID, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"message":"data flow not found","request_id":"req-delete-df"}`)
	})

	resp, err := client.ADS.DeleteDataFlow(ctx, adsTestRegion, adsTestUUID, adsTestDataFlowID)
	require.Error(t, err)
	require.NotNil(t, resp)
	var apiErr *ErrorResponse
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, "data flow not found", apiErr.Message)
}

func TestADS_PauseDataFlow(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/data-cloud-platform/tor1/workspaces/"+adsTestUUID+"/data-flows/"+adsTestDataFlowID+"/pause", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{}`)
	})

	resp, err := client.ADS.PauseDataFlow(ctx, adsTestRegion, adsTestUUID, adsTestDataFlowID)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestADS_PauseDataFlow_EscapesPath(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/v2/data-cloud-platform/tor1%2Fother/workspaces/ws%2Fabc/data-flows/df%2F1/pause", r.URL.EscapedPath())
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{}`)
	})

	resp, err := client.ADS.PauseDataFlow(ctx, "tor1/other", "ws/abc", "df/1")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestADS_PauseDataFlow_Error(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/data-cloud-platform/tor1/workspaces/"+adsTestUUID+"/data-flows/"+adsTestDataFlowID+"/pause", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		fmt.Fprint(w, `{"message":"not active","request_id":"req-pause"}`)
	})

	resp, err := client.ADS.PauseDataFlow(ctx, adsTestRegion, adsTestUUID, adsTestDataFlowID)
	require.Error(t, err)
	require.NotNil(t, resp)
	var apiErr *ErrorResponse
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, "not active", apiErr.Message)
}

func TestADS_ResumeDataFlow(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/data-cloud-platform/tor1/workspaces/"+adsTestUUID+"/data-flows/"+adsTestDataFlowID+"/resume", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{}`)
	})

	resp, err := client.ADS.ResumeDataFlow(ctx, adsTestRegion, adsTestUUID, adsTestDataFlowID)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestADS_ResumeDataFlow_EscapesPath(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v2/data-cloud-platform/tor1%2Fother/workspaces/ws%2Fabc/data-flows/df%2F1/resume", r.URL.EscapedPath())
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{}`)
	})

	resp, err := client.ADS.ResumeDataFlow(ctx, "tor1/other", "ws/abc", "df/1")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestADS_ResumeDataFlow_Error(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/data-cloud-platform/tor1/workspaces/"+adsTestUUID+"/data-flows/"+adsTestDataFlowID+"/resume", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		fmt.Fprint(w, `{"message":"not paused","request_id":"req-resume"}`)
	})

	resp, err := client.ADS.ResumeDataFlow(ctx, adsTestRegion, adsTestUUID, adsTestDataFlowID)
	require.Error(t, err)
	require.NotNil(t, resp)
	var apiErr *ErrorResponse
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, "not paused", apiErr.Message)
}

func TestADS_RestartDataFlow(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/data-cloud-platform/tor1/workspaces/"+adsTestUUID+"/data-flows/"+adsTestDataFlowID+"/restart", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{}`)
	})

	resp, err := client.ADS.RestartDataFlow(ctx, adsTestRegion, adsTestUUID, adsTestDataFlowID)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestADS_RestartDataFlow_EscapesPath(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v2/data-cloud-platform/tor1%2Fother/workspaces/ws%2Fabc/data-flows/df%2F1/restart", r.URL.EscapedPath())
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{}`)
	})

	resp, err := client.ADS.RestartDataFlow(ctx, "tor1/other", "ws/abc", "df/1")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestADS_RestartDataFlow_Error(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/data-cloud-platform/tor1/workspaces/"+adsTestUUID+"/data-flows/"+adsTestDataFlowID+"/restart", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"message":"data flow not found","request_id":"req-restart"}`)
	})

	resp, err := client.ADS.RestartDataFlow(ctx, adsTestRegion, adsTestUUID, adsTestDataFlowID)
	require.Error(t, err)
	require.NotNil(t, resp)
	var apiErr *ErrorResponse
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, "data flow not found", apiErr.Message)
}

func TestADS_RewindDataFlow(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/data-cloud-platform/tor1/workspaces/"+adsTestUUID+"/data-flows/"+adsTestDataFlowID+"/rewind", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		raw, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		assert.JSONEq(t, `{
			"id": "df-019fb39c-14d9-7080-933e-b9b90e25acda",
			"workspace": "ws-019fb39c-14d9-7080-933e-b9b90e25acda",
			"rewind_to": "2026-11-01T00:00:00Z"
		}`, string(raw))
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{}`)
	})

	resp, err := client.ADS.RewindDataFlow(ctx, &ADSDataFlowRewindRequest{
		Region:    adsTestRegion,
		ID:        adsTestDataFlowID,
		Workspace: adsTestUUID,
		RewindTo:  "2026-11-01T00:00:00Z",
	})
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestADS_RewindDataFlow_EscapesPath(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v2/data-cloud-platform/tor1%2Fother/workspaces/ws%2Fabc/data-flows/df%2F1/rewind", r.URL.EscapedPath())
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{}`)
	})

	resp, err := client.ADS.RewindDataFlow(ctx, &ADSDataFlowRewindRequest{
		Region:    "tor1/other",
		ID:        "df/1",
		Workspace: "ws/abc",
	})
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestADS_RewindDataFlow_Error(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/data-cloud-platform/tor1/workspaces/"+adsTestUUID+"/data-flows/"+adsTestDataFlowID+"/rewind", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"message":"rewind is not supported","request_id":"req-rewind"}`)
	})

	resp, err := client.ADS.RewindDataFlow(ctx, &ADSDataFlowRewindRequest{
		Region:    adsTestRegion,
		ID:        adsTestDataFlowID,
		Workspace: adsTestUUID,
		RewindTo:  "2026-11-01T00:00:00Z",
	})
	require.Error(t, err)
	require.NotNil(t, resp)
	var apiErr *ErrorResponse
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, "rewind is not supported", apiErr.Message)
}

func TestADS_CreateFchainUploadURL(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/data-cloud-platform/tor1/workspaces/"+adsTestUUID+"/fchain/upload-url", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		raw, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		assert.JSONEq(t, `{
			"sha256": "abc123",
			"workspace": "ws-019fb39c-14d9-7080-933e-b9b90e25acda"
		}`, string(raw))
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{
			"upload_url": "https://upload.example/secret_value",
			"object_key": "fchain/abc",
			"expires_at": "2026-11-07T01:00:00Z"
		}`)
	})

	got, resp, err := client.ADS.CreateFchainUploadURL(ctx, &ADSFchainUploadURLRequest{
		Region:    adsTestRegion,
		SHA256:    "abc123",
		Workspace: adsTestUUID,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, &ADSFchainUploadURLResponse{
		UploadURL: "https://upload.example/secret_value",
		ObjectKey: "fchain/abc",
		ExpiresAt: "2026-11-07T01:00:00Z",
	}, got)
}

func TestADS_CreateFchainUploadURL_EscapesPath(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v2/data-cloud-platform/tor1%2Fother/workspaces/ws%2Fabc/fchain/upload-url", r.URL.EscapedPath())
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"upload_url":"https://upload.example/secret_value","object_key":"k"}`)
	})

	got, resp, err := client.ADS.CreateFchainUploadURL(ctx, &ADSFchainUploadURLRequest{Region: "tor1/other", Workspace: "ws/abc"})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "https://upload.example/secret_value", got.UploadURL)
}

func TestADS_CreateFchainUploadURL_Error(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/data-cloud-platform/tor1/workspaces/"+adsTestUUID+"/fchain/upload-url", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"message":"sha256 required","request_id":"req-upload-url"}`)
	})

	got, resp, err := client.ADS.CreateFchainUploadURL(ctx, &ADSFchainUploadURLRequest{Region: adsTestRegion, Workspace: adsTestUUID})
	require.Error(t, err)
	assert.Nil(t, got)
	require.NotNil(t, resp)
	var apiErr *ErrorResponse
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, "sha256 required", apiErr.Message)
}

func TestADS_AttachDataFlowFchain(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/data-cloud-platform/tor1/workspaces/"+adsTestUUID+"/data-flows/"+adsTestDataFlowID+"/fchain", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		raw, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		assert.JSONEq(t, `{
			"id": "df-019fb39c-14d9-7080-933e-b9b90e25acda",
			"workspace": "ws-019fb39c-14d9-7080-933e-b9b90e25acda",
			"object_key": "fchain/abc"
		}`, string(raw))
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{}`)
	})

	resp, err := client.ADS.AttachDataFlowFchain(ctx, &ADSAttachDataFlowFchainRequest{
		Region:    adsTestRegion,
		ID:        adsTestDataFlowID,
		Workspace: adsTestUUID,
		ObjectKey: "fchain/abc",
	})
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestADS_AttachDataFlowFchain_EscapesPath(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v2/data-cloud-platform/tor1%2Fother/workspaces/ws%2Fabc/data-flows/df%2F1/fchain", r.URL.EscapedPath())
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{}`)
	})

	resp, err := client.ADS.AttachDataFlowFchain(ctx, &ADSAttachDataFlowFchainRequest{
		Region:    "tor1/other",
		ID:        "df/1",
		Workspace: "ws/abc",
		ObjectKey: "fchain/abc",
	})
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestADS_AttachDataFlowFchain_Error(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/data-cloud-platform/tor1/workspaces/"+adsTestUUID+"/data-flows/"+adsTestDataFlowID+"/fchain", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"message":"object missing","request_id":"req-attach"}`)
	})

	resp, err := client.ADS.AttachDataFlowFchain(ctx, &ADSAttachDataFlowFchainRequest{
		Region:    adsTestRegion,
		ID:        adsTestDataFlowID,
		Workspace: adsTestUUID,
		ObjectKey: "fchain/abc",
	})
	require.Error(t, err)
	require.NotNil(t, resp)
	var apiErr *ErrorResponse
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, "object missing", apiErr.Message)
}

func TestADS_GetWorkspaceMetrics(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/data-cloud-platform/tor1/workspaces/"+adsTestUUID+"/metrics", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		raw, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		assert.JSONEq(t, `{
			"workspace": "ws-019fb39c-14d9-7080-933e-b9b90e25acda",
			"metric_names": ["events"],
			"data_flow_id": "df-019fb39c-14d9-7080-933e-b9b90e25acda",
			"from": "2026-11-01T00:00:00Z",
			"to": "2026-11-07T00:00:00Z",
			"group_by": ["data_flow_id"],
			"limit": 100,
			"metric_type": "counter"
		}`, string(raw))
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{
			"count": 1,
			"columns": ["data_flow_id", "value"],
			"rows": [{"data_flow_id": "df-019fb39c-14d9-7080-933e-b9b90e25acda", "value": 42}]
		}`)
	})

	got, resp, err := client.ADS.GetWorkspaceMetrics(ctx, &ADSWorkspaceMetricsRequest{
		Region:      adsTestRegion,
		Workspace:   adsTestUUID,
		MetricNames: []string{"events"},
		DataFlowID:  adsTestDataFlowID,
		From:        "2026-11-01T00:00:00Z",
		To:          "2026-11-07T00:00:00Z",
		GroupBy:     []string{"data_flow_id"},
		Limit:       100,
		MetricType:  "counter",
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, int32(1), got.Count)
	assert.Equal(t, []string{"data_flow_id", "value"}, got.Columns)
	require.Len(t, got.Rows, 1)
	assert.Equal(t, adsTestDataFlowID, got.Rows[0]["data_flow_id"])
	assert.EqualValues(t, 42, got.Rows[0]["value"])
}

func TestADS_GetWorkspaceMetrics_EscapesPath(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v2/data-cloud-platform/tor1%2Fother/workspaces/ws%2Fabc/metrics", r.URL.EscapedPath())
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"count":0,"columns":[],"rows":[]}`)
	})

	got, resp, err := client.ADS.GetWorkspaceMetrics(ctx, &ADSWorkspaceMetricsRequest{Region: "tor1/other", Workspace: "ws/abc"})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, int32(0), got.Count)
}

func TestADS_GetWorkspaceMetrics_Error(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/data-cloud-platform/tor1/workspaces/"+adsTestUUID+"/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"message":"invalid range","request_id":"req-metrics"}`)
	})

	got, resp, err := client.ADS.GetWorkspaceMetrics(ctx, &ADSWorkspaceMetricsRequest{Region: adsTestRegion, Workspace: adsTestUUID})
	require.Error(t, err)
	assert.Nil(t, got)
	require.NotNil(t, resp)
	var apiErr *ErrorResponse
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, "invalid range", apiErr.Message)
}
