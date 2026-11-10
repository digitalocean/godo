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

const dataCloudTestDataFlowID = "df-019fb39c-14d9-7080-933e-b9b90e25acda"

func TestDataCloudPlatform_GetWorkspaceSecret(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/data-cloud-platform/tor1/workspaces/"+dataCloudTestUUID+"/secrets/db-password", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		assert.Equal(t, "/v2/data-cloud-platform/tor1/workspaces/"+dataCloudTestUUID+"/secrets/db-password", r.URL.EscapedPath())
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

	got, resp, err := client.DataCloudPlatform.GetWorkspaceSecret(ctx, dataCloudTestRegion, dataCloudTestUUID, "db-password")
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, &DataCloudGetWorkspaceSecretResponse{
		Data: &DataCloudGetSecretData{
			Data: map[string]any{},
			Metadata: &DataCloudSecretMetadata{
				CreatedTime:    "2026-11-07T00:00:00Z",
				CustomMetadata: map[string]any{"owner": "analytics"},
				Destroyed:      false,
				Version:        3,
			},
			Keys: []string{"password", "username"},
		},
	}, got)
}

func TestDataCloudPlatform_GetWorkspaceSecret_EscapesPath(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/v2/data-cloud-platform/tor1%2Fother/workspaces/ws%2Fabc/secrets/db%2Fpassword", r.URL.EscapedPath())
		assert.Empty(t, r.URL.RawQuery)
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"data":{"keys":[]}}`)
	})

	got, resp, err := client.DataCloudPlatform.GetWorkspaceSecret(ctx, "tor1/other", "ws/abc", "db/password")
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, &DataCloudGetWorkspaceSecretResponse{
		Data: &DataCloudGetSecretData{Keys: []string{}},
	}, got)
}

func TestDataCloudPlatform_GetWorkspaceSecret_Error(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/data-cloud-platform/tor1/workspaces/"+dataCloudTestUUID+"/secrets/db-password", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"message":"secret not found","request_id":"req-get-secret"}`)
	})

	got, resp, err := client.DataCloudPlatform.GetWorkspaceSecret(ctx, dataCloudTestRegion, dataCloudTestUUID, "db-password")
	require.Error(t, err)
	assert.Nil(t, got)
	require.NotNil(t, resp)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	var apiErr *ErrorResponse
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, "secret not found", apiErr.Message)
}

func TestDataCloudPlatform_ListWorkspaceSecrets(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/data-cloud-platform/tor1/workspaces/"+dataCloudTestUUID+"/secrets", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		assert.Equal(t, "/v2/data-cloud-platform/tor1/workspaces/"+dataCloudTestUUID+"/secrets", r.URL.EscapedPath())
		assert.Empty(t, r.URL.RawQuery)
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"secret_keys":["db-password","api-token"]}`)
	})

	got, resp, err := client.DataCloudPlatform.ListWorkspaceSecrets(ctx, dataCloudTestRegion, dataCloudTestUUID)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, &DataCloudListWorkspaceSecretsResponse{
		SecretKeys: []string{"db-password", "api-token"},
	}, got)
}

func TestDataCloudPlatform_ListWorkspaceSecrets_EscapesPath(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/v2/data-cloud-platform/tor1%2Fother/workspaces/ws%2Fabc/secrets", r.URL.EscapedPath())
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"secret_keys":[]}`)
	})

	got, resp, err := client.DataCloudPlatform.ListWorkspaceSecrets(ctx, "tor1/other", "ws/abc")
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, &DataCloudListWorkspaceSecretsResponse{SecretKeys: []string{}}, got)
}

func TestDataCloudPlatform_ListWorkspaceSecrets_Error(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/data-cloud-platform/tor1/workspaces/"+dataCloudTestUUID+"/secrets", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"message":"workspace not found","request_id":"req-list-secrets"}`)
	})

	got, resp, err := client.DataCloudPlatform.ListWorkspaceSecrets(ctx, dataCloudTestRegion, dataCloudTestUUID)
	require.Error(t, err)
	assert.Nil(t, got)
	require.NotNil(t, resp)
	var apiErr *ErrorResponse
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, "workspace not found", apiErr.Message)
}

func TestDataCloudPlatform_DeleteWorkspaceSecret(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/data-cloud-platform/tor1/workspaces/"+dataCloudTestUUID+"/secrets/db-password", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		assert.Equal(t, "/v2/data-cloud-platform/tor1/workspaces/"+dataCloudTestUUID+"/secrets/db-password", r.URL.EscapedPath())
		assert.Empty(t, r.URL.RawQuery)
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{}`)
	})

	resp, err := client.DataCloudPlatform.DeleteWorkspaceSecret(ctx, dataCloudTestRegion, dataCloudTestUUID, "db-password")
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestDataCloudPlatform_DeleteWorkspaceSecret_EscapesPath(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Equal(t, "/v2/data-cloud-platform/tor1%2Fother/workspaces/ws%2Fabc/secrets/db%2Fpassword", r.URL.EscapedPath())
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{}`)
	})

	resp, err := client.DataCloudPlatform.DeleteWorkspaceSecret(ctx, "tor1/other", "ws/abc", "db/password")
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestDataCloudPlatform_DeleteWorkspaceSecret_Error(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/data-cloud-platform/tor1/workspaces/"+dataCloudTestUUID+"/secrets/db-password", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"message":"secret not found","request_id":"req-delete-secret"}`)
	})

	resp, err := client.DataCloudPlatform.DeleteWorkspaceSecret(ctx, dataCloudTestRegion, dataCloudTestUUID, "db-password")
	require.Error(t, err)
	require.NotNil(t, resp)
	var apiErr *ErrorResponse
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, "secret not found", apiErr.Message)
}

func TestDataCloudPlatform_GetWorkspaceQueryCredentials(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/data-cloud-platform/tor1/workspaces/"+dataCloudTestUUID+"/query/credentials", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		assert.Equal(t, "/v2/data-cloud-platform/tor1/workspaces/"+dataCloudTestUUID+"/query/credentials", r.URL.EscapedPath())
		assert.Empty(t, r.URL.RawQuery)
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"user":"query_user","password":"secret_value"}`)
	})

	got, resp, err := client.DataCloudPlatform.GetWorkspaceQueryCredentials(ctx, dataCloudTestRegion, dataCloudTestUUID)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, &DataCloudQueryCredentials{
		User:     "query_user",
		Password: "secret_value",
	}, got)
}

func TestDataCloudPlatform_GetWorkspaceQueryCredentials_EscapesPath(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/v2/data-cloud-platform/tor1%2Fother/workspaces/ws%2Fabc/query/credentials", r.URL.EscapedPath())
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"user":"query_user","password":"secret_value"}`)
	})

	got, resp, err := client.DataCloudPlatform.GetWorkspaceQueryCredentials(ctx, "tor1/other", "ws/abc")
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, &DataCloudQueryCredentials{User: "query_user", Password: "secret_value"}, got)
}

func TestDataCloudPlatform_GetWorkspaceQueryCredentials_Error(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/data-cloud-platform/tor1/workspaces/"+dataCloudTestUUID+"/query/credentials", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"message":"workspace not found","request_id":"req-query-creds"}`)
	})

	got, resp, err := client.DataCloudPlatform.GetWorkspaceQueryCredentials(ctx, dataCloudTestRegion, dataCloudTestUUID)
	require.Error(t, err)
	assert.Nil(t, got)
	require.NotNil(t, resp)
	var apiErr *ErrorResponse
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, "workspace not found", apiErr.Message)
}

func TestDataCloudPlatform_CreateDataFlow(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/data-cloud-platform/workspaces/"+dataCloudTestUUID+"/data-flows", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		assert.Equal(t, "/v2/data-cloud-platform/workspaces/"+dataCloudTestUUID+"/data-flows", r.URL.EscapedPath())
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

	got, resp, err := client.DataCloudPlatform.CreateDataFlow(ctx, &DataCloudDataFlowCreateRequest{
		Name:            "ingest",
		DesiredReplicas: 1,
		MaxReplicas:     3,
		Workspace:       dataCloudTestUUID,
		ImageRef:        "registry.example/flow:1",
		Manifest:        json.RawMessage(`{"schema":"v1"}`),
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, &DataCloudDataFlowCreateResponse{
		UUID:  dataCloudTestDataFlowID,
		State: "pending",
		Upload: &DataCloudTarballUpload{
			URL:       "https://upload.example/secret_value",
			ObjectKey: "fchain/abc",
			ExpiresAt: "2026-11-07T01:00:00Z",
		},
	}, got)
}

func TestDataCloudPlatform_CreateDataFlow_EscapesPath(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/v2/data-cloud-platform/workspaces/ws%2Fabc/data-flows", r.URL.EscapedPath())
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"uuid":"df-1","state":"pending"}`)
	})

	got, resp, err := client.DataCloudPlatform.CreateDataFlow(ctx, &DataCloudDataFlowCreateRequest{
		Workspace: "ws/abc",
		Name:      "ingest",
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, &DataCloudDataFlowCreateResponse{UUID: "df-1", State: "pending"}, got)
}

func TestDataCloudPlatform_CreateDataFlow_Error(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/data-cloud-platform/workspaces/"+dataCloudTestUUID+"/data-flows", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"message":"invalid manifest","request_id":"req-create-df"}`)
	})

	got, resp, err := client.DataCloudPlatform.CreateDataFlow(ctx, &DataCloudDataFlowCreateRequest{Workspace: dataCloudTestUUID})
	require.Error(t, err)
	assert.Nil(t, got)
	require.NotNil(t, resp)
	var apiErr *ErrorResponse
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, "invalid manifest", apiErr.Message)
}

func TestDataCloudPlatform_ListDataFlows(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/data-cloud-platform/workspaces/"+dataCloudTestUUID+"/data-flows", func(w http.ResponseWriter, r *http.Request) {
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

	got, resp, err := client.DataCloudPlatform.ListDataFlows(ctx, &DataCloudDataFlowListRequest{
		Workspace: dataCloudTestUUID,
		PageSize:  25,
		PageToken: "tok-1",
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, &DataCloudListDataFlowsResponse{
		DataFlows: []DataCloudDataFlow{{
			UUID:            dataCloudTestDataFlowID,
			Name:            "ingest",
			State:           "active",
			Workspace:       dataCloudTestUUID,
			DesiredReplicas: 1,
			MaxReplicas:     3,
			Tarball:         &DataCloudTarballStatus{Requested: "sha256:abc", Attached: "sha256:abc"},
			DesiredState:    "running",
		}},
		NextPageToken: "tok-2",
	}, got)
}

func TestDataCloudPlatform_ListDataFlows_EscapesPathAndQuery(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/v2/data-cloud-platform/workspaces/ws%2Fabc/data-flows", r.URL.EscapedPath())
		assert.Equal(t, "a/b", r.URL.Query().Get("page_token"))
		assert.Contains(t, r.URL.RawQuery, "page_token=a%2Fb")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"data_flows":[]}`)
	})

	got, resp, err := client.DataCloudPlatform.ListDataFlows(ctx, &DataCloudDataFlowListRequest{
		Workspace: "ws/abc",
		PageToken: "a/b",
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, &DataCloudListDataFlowsResponse{DataFlows: []DataCloudDataFlow{}}, got)
}

func TestDataCloudPlatform_ListDataFlows_Error(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/data-cloud-platform/workspaces/"+dataCloudTestUUID+"/data-flows", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"message":"workspace not found","request_id":"req-list-df"}`)
	})

	got, resp, err := client.DataCloudPlatform.ListDataFlows(ctx, &DataCloudDataFlowListRequest{Workspace: dataCloudTestUUID})
	require.Error(t, err)
	assert.Nil(t, got)
	require.NotNil(t, resp)
	var apiErr *ErrorResponse
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, "workspace not found", apiErr.Message)
}

func TestDataCloudPlatform_GetDataFlow(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/data-cloud-platform/workspaces/"+dataCloudTestUUID+"/data-flows/"+dataCloudTestDataFlowID, func(w http.ResponseWriter, r *http.Request) {
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

	got, resp, err := client.DataCloudPlatform.GetDataFlow(ctx, dataCloudTestUUID, dataCloudTestDataFlowID)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, dataCloudTestDataFlowID, got.UUID)
	assert.Equal(t, "ingest", got.Name)
	assert.Equal(t, "active", got.State)
	assert.JSONEq(t, `{"schema":"v1"}`, string(got.Spec))
}

func TestDataCloudPlatform_GetDataFlow_EscapesPath(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/v2/data-cloud-platform/workspaces/ws%2Fabc/data-flows/df%2F1", r.URL.EscapedPath())
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"uuid":"df/1","state":"active"}`)
	})

	got, resp, err := client.DataCloudPlatform.GetDataFlow(ctx, "ws/abc", "df/1")
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "df/1", got.UUID)
}

func TestDataCloudPlatform_GetDataFlow_Error(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/data-cloud-platform/workspaces/"+dataCloudTestUUID+"/data-flows/"+dataCloudTestDataFlowID, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"message":"data flow not found","request_id":"req-get-df"}`)
	})

	got, resp, err := client.DataCloudPlatform.GetDataFlow(ctx, dataCloudTestUUID, dataCloudTestDataFlowID)
	require.Error(t, err)
	assert.Nil(t, got)
	require.NotNil(t, resp)
	var apiErr *ErrorResponse
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, "data flow not found", apiErr.Message)
}

func TestDataCloudPlatform_UpdateDataFlow(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/data-cloud-platform/workspaces/"+dataCloudTestUUID+"/data-flows/"+dataCloudTestDataFlowID, func(w http.ResponseWriter, r *http.Request) {
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

	got, resp, err := client.DataCloudPlatform.UpdateDataFlow(ctx, &DataCloudDataFlowUpdateRequest{
		ID:              dataCloudTestDataFlowID,
		Workspace:       dataCloudTestUUID,
		DesiredReplicas: 2,
		SecretRefs:      []string{"db-password"},
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, &DataCloudDataFlowUpdateResponse{UUID: dataCloudTestDataFlowID, State: "active"}, got)
}

func TestDataCloudPlatform_UpdateDataFlow_EscapesPath(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPatch, r.Method)
		assert.Equal(t, "/v2/data-cloud-platform/workspaces/ws%2Fabc/data-flows/df%2F1", r.URL.EscapedPath())
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"uuid":"df/1","state":"active"}`)
	})

	got, resp, err := client.DataCloudPlatform.UpdateDataFlow(ctx, &DataCloudDataFlowUpdateRequest{
		ID:        "df/1",
		Workspace: "ws/abc",
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "df/1", got.UUID)
}

func TestDataCloudPlatform_UpdateDataFlow_Error(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/data-cloud-platform/workspaces/"+dataCloudTestUUID+"/data-flows/"+dataCloudTestDataFlowID, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"message":"invalid patch","request_id":"req-patch-df"}`)
	})

	got, resp, err := client.DataCloudPlatform.UpdateDataFlow(ctx, &DataCloudDataFlowUpdateRequest{
		ID:        dataCloudTestDataFlowID,
		Workspace: dataCloudTestUUID,
	})
	require.Error(t, err)
	assert.Nil(t, got)
	require.NotNil(t, resp)
	var apiErr *ErrorResponse
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, "invalid patch", apiErr.Message)
}

func TestDataCloudPlatform_DeleteDataFlow(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/data-cloud-platform/workspaces/"+dataCloudTestUUID+"/data-flows/"+dataCloudTestDataFlowID, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{}`)
	})

	resp, err := client.DataCloudPlatform.DeleteDataFlow(ctx, dataCloudTestUUID, dataCloudTestDataFlowID)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestDataCloudPlatform_DeleteDataFlow_EscapesPath(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Equal(t, "/v2/data-cloud-platform/workspaces/ws%2Fabc/data-flows/df%2F1", r.URL.EscapedPath())
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{}`)
	})

	resp, err := client.DataCloudPlatform.DeleteDataFlow(ctx, "ws/abc", "df/1")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestDataCloudPlatform_DeleteDataFlow_Error(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/data-cloud-platform/workspaces/"+dataCloudTestUUID+"/data-flows/"+dataCloudTestDataFlowID, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"message":"data flow not found","request_id":"req-delete-df"}`)
	})

	resp, err := client.DataCloudPlatform.DeleteDataFlow(ctx, dataCloudTestUUID, dataCloudTestDataFlowID)
	require.Error(t, err)
	require.NotNil(t, resp)
	var apiErr *ErrorResponse
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, "data flow not found", apiErr.Message)
}

func TestDataCloudPlatform_PauseDataFlow(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/data-cloud-platform/workspaces/"+dataCloudTestUUID+"/data-flows/"+dataCloudTestDataFlowID+"/pause", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{}`)
	})

	resp, err := client.DataCloudPlatform.PauseDataFlow(ctx, dataCloudTestUUID, dataCloudTestDataFlowID)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestDataCloudPlatform_PauseDataFlow_EscapesPath(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/v2/data-cloud-platform/workspaces/ws%2Fabc/data-flows/df%2F1/pause", r.URL.EscapedPath())
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{}`)
	})

	resp, err := client.DataCloudPlatform.PauseDataFlow(ctx, "ws/abc", "df/1")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestDataCloudPlatform_PauseDataFlow_Error(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/data-cloud-platform/workspaces/"+dataCloudTestUUID+"/data-flows/"+dataCloudTestDataFlowID+"/pause", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		fmt.Fprint(w, `{"message":"not active","request_id":"req-pause"}`)
	})

	resp, err := client.DataCloudPlatform.PauseDataFlow(ctx, dataCloudTestUUID, dataCloudTestDataFlowID)
	require.Error(t, err)
	require.NotNil(t, resp)
	var apiErr *ErrorResponse
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, "not active", apiErr.Message)
}

func TestDataCloudPlatform_ResumeDataFlow(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/data-cloud-platform/workspaces/"+dataCloudTestUUID+"/data-flows/"+dataCloudTestDataFlowID+"/resume", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{}`)
	})

	resp, err := client.DataCloudPlatform.ResumeDataFlow(ctx, dataCloudTestUUID, dataCloudTestDataFlowID)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestDataCloudPlatform_ResumeDataFlow_EscapesPath(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v2/data-cloud-platform/workspaces/ws%2Fabc/data-flows/df%2F1/resume", r.URL.EscapedPath())
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{}`)
	})

	resp, err := client.DataCloudPlatform.ResumeDataFlow(ctx, "ws/abc", "df/1")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestDataCloudPlatform_ResumeDataFlow_Error(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/data-cloud-platform/workspaces/"+dataCloudTestUUID+"/data-flows/"+dataCloudTestDataFlowID+"/resume", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		fmt.Fprint(w, `{"message":"not paused","request_id":"req-resume"}`)
	})

	resp, err := client.DataCloudPlatform.ResumeDataFlow(ctx, dataCloudTestUUID, dataCloudTestDataFlowID)
	require.Error(t, err)
	require.NotNil(t, resp)
	var apiErr *ErrorResponse
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, "not paused", apiErr.Message)
}

func TestDataCloudPlatform_RestartDataFlow(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/data-cloud-platform/workspaces/"+dataCloudTestUUID+"/data-flows/"+dataCloudTestDataFlowID+"/restart", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{}`)
	})

	resp, err := client.DataCloudPlatform.RestartDataFlow(ctx, dataCloudTestUUID, dataCloudTestDataFlowID)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestDataCloudPlatform_RestartDataFlow_EscapesPath(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v2/data-cloud-platform/workspaces/ws%2Fabc/data-flows/df%2F1/restart", r.URL.EscapedPath())
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{}`)
	})

	resp, err := client.DataCloudPlatform.RestartDataFlow(ctx, "ws/abc", "df/1")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestDataCloudPlatform_RestartDataFlow_Error(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/data-cloud-platform/workspaces/"+dataCloudTestUUID+"/data-flows/"+dataCloudTestDataFlowID+"/restart", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"message":"data flow not found","request_id":"req-restart"}`)
	})

	resp, err := client.DataCloudPlatform.RestartDataFlow(ctx, dataCloudTestUUID, dataCloudTestDataFlowID)
	require.Error(t, err)
	require.NotNil(t, resp)
	var apiErr *ErrorResponse
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, "data flow not found", apiErr.Message)
}

func TestDataCloudPlatform_RewindDataFlow(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/data-cloud-platform/workspaces/"+dataCloudTestUUID+"/data-flows/"+dataCloudTestDataFlowID+"/rewind", func(w http.ResponseWriter, r *http.Request) {
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

	resp, err := client.DataCloudPlatform.RewindDataFlow(ctx, &DataCloudDataFlowRewindRequest{
		ID:        dataCloudTestDataFlowID,
		Workspace: dataCloudTestUUID,
		RewindTo:  "2026-11-01T00:00:00Z",
	})
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestDataCloudPlatform_RewindDataFlow_EscapesPath(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v2/data-cloud-platform/workspaces/ws%2Fabc/data-flows/df%2F1/rewind", r.URL.EscapedPath())
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{}`)
	})

	resp, err := client.DataCloudPlatform.RewindDataFlow(ctx, &DataCloudDataFlowRewindRequest{
		ID:        "df/1",
		Workspace: "ws/abc",
	})
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestDataCloudPlatform_RewindDataFlow_Error(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/data-cloud-platform/workspaces/"+dataCloudTestUUID+"/data-flows/"+dataCloudTestDataFlowID+"/rewind", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"message":"rewind is not supported","request_id":"req-rewind"}`)
	})

	resp, err := client.DataCloudPlatform.RewindDataFlow(ctx, &DataCloudDataFlowRewindRequest{
		ID:        dataCloudTestDataFlowID,
		Workspace: dataCloudTestUUID,
		RewindTo:  "2026-11-01T00:00:00Z",
	})
	require.Error(t, err)
	require.NotNil(t, resp)
	var apiErr *ErrorResponse
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, "rewind is not supported", apiErr.Message)
}

func TestDataCloudPlatform_CreateFchainUploadURL(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/data-cloud-platform/workspaces/"+dataCloudTestUUID+"/fchain/upload-url", func(w http.ResponseWriter, r *http.Request) {
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

	got, resp, err := client.DataCloudPlatform.CreateFchainUploadURL(ctx, &DataCloudFchainUploadURLRequest{
		SHA256:    "abc123",
		Workspace: dataCloudTestUUID,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, &DataCloudFchainUploadURLResponse{
		UploadURL: "https://upload.example/secret_value",
		ObjectKey: "fchain/abc",
		ExpiresAt: "2026-11-07T01:00:00Z",
	}, got)
}

func TestDataCloudPlatform_CreateFchainUploadURL_EscapesPath(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v2/data-cloud-platform/workspaces/ws%2Fabc/fchain/upload-url", r.URL.EscapedPath())
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"upload_url":"https://upload.example/secret_value","object_key":"k"}`)
	})

	got, resp, err := client.DataCloudPlatform.CreateFchainUploadURL(ctx, &DataCloudFchainUploadURLRequest{Workspace: "ws/abc"})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "https://upload.example/secret_value", got.UploadURL)
}

func TestDataCloudPlatform_CreateFchainUploadURL_Error(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/data-cloud-platform/workspaces/"+dataCloudTestUUID+"/fchain/upload-url", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"message":"sha256 required","request_id":"req-upload-url"}`)
	})

	got, resp, err := client.DataCloudPlatform.CreateFchainUploadURL(ctx, &DataCloudFchainUploadURLRequest{Workspace: dataCloudTestUUID})
	require.Error(t, err)
	assert.Nil(t, got)
	require.NotNil(t, resp)
	var apiErr *ErrorResponse
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, "sha256 required", apiErr.Message)
}

func TestDataCloudPlatform_AttachDataFlowFchain(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/data-cloud-platform/workspaces/"+dataCloudTestUUID+"/data-flows/"+dataCloudTestDataFlowID+"/fchain", func(w http.ResponseWriter, r *http.Request) {
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

	resp, err := client.DataCloudPlatform.AttachDataFlowFchain(ctx, &DataCloudAttachDataFlowFchainRequest{
		ID:        dataCloudTestDataFlowID,
		Workspace: dataCloudTestUUID,
		ObjectKey: "fchain/abc",
	})
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestDataCloudPlatform_AttachDataFlowFchain_EscapesPath(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v2/data-cloud-platform/workspaces/ws%2Fabc/data-flows/df%2F1/fchain", r.URL.EscapedPath())
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{}`)
	})

	resp, err := client.DataCloudPlatform.AttachDataFlowFchain(ctx, &DataCloudAttachDataFlowFchainRequest{
		ID:        "df/1",
		Workspace: "ws/abc",
		ObjectKey: "fchain/abc",
	})
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestDataCloudPlatform_AttachDataFlowFchain_Error(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/data-cloud-platform/workspaces/"+dataCloudTestUUID+"/data-flows/"+dataCloudTestDataFlowID+"/fchain", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"message":"object missing","request_id":"req-attach"}`)
	})

	resp, err := client.DataCloudPlatform.AttachDataFlowFchain(ctx, &DataCloudAttachDataFlowFchainRequest{
		ID:        dataCloudTestDataFlowID,
		Workspace: dataCloudTestUUID,
		ObjectKey: "fchain/abc",
	})
	require.Error(t, err)
	require.NotNil(t, resp)
	var apiErr *ErrorResponse
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, "object missing", apiErr.Message)
}

func TestDataCloudPlatform_GetWorkspaceMetrics(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/data-cloud-platform/workspaces/"+dataCloudTestUUID+"/metrics", func(w http.ResponseWriter, r *http.Request) {
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

	got, resp, err := client.DataCloudPlatform.GetWorkspaceMetrics(ctx, &DataCloudWorkspaceMetricsRequest{
		Workspace:   dataCloudTestUUID,
		MetricNames: []string{"events"},
		DataFlowID:  dataCloudTestDataFlowID,
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
	assert.Equal(t, dataCloudTestDataFlowID, got.Rows[0]["data_flow_id"])
	assert.EqualValues(t, 42, got.Rows[0]["value"])
}

func TestDataCloudPlatform_GetWorkspaceMetrics_EscapesPath(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v2/data-cloud-platform/workspaces/ws%2Fabc/metrics", r.URL.EscapedPath())
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"count":0,"columns":[],"rows":[]}`)
	})

	got, resp, err := client.DataCloudPlatform.GetWorkspaceMetrics(ctx, &DataCloudWorkspaceMetricsRequest{Workspace: "ws/abc"})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, int32(0), got.Count)
}

func TestDataCloudPlatform_GetWorkspaceMetrics_Error(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/data-cloud-platform/workspaces/"+dataCloudTestUUID+"/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"message":"invalid range","request_id":"req-metrics"}`)
	})

	got, resp, err := client.DataCloudPlatform.GetWorkspaceMetrics(ctx, &DataCloudWorkspaceMetricsRequest{Workspace: dataCloudTestUUID})
	require.Error(t, err)
	assert.Nil(t, got)
	require.NotNil(t, resp)
	var apiErr *ErrorResponse
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, "invalid range", apiErr.Message)
}
