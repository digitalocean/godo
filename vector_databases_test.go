package godo

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var vectorDB = VectorDB{
	ID:        "da4e0206-d019-41d7-b51f-deadbeefbb8f",
	Name:      "vectortest",
	Region:    "nyc3",
	OwnerUUID: "880b7f98-f062-404d-b33c-458d545696f6",
	Status:    "active",
	Config: &VectorDBConfig{
		DefaultQuantization: "rq",
		EnableAutoSchema:    true,
		WeaviateVersion:     "1.27.0",
	},
	CreatedAt: time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
	UpdatedAt: time.Date(2024, 2, 1, 8, 30, 0, 0, time.UTC),
	Endpoints: &VectorDBEndpoints{
		HTTP: "https://vectortest-abc123.vectordb.digitalocean.com",
		GRPC: "vectortest-abc123.vectordb.digitalocean.com:443",
	},
	Size: "small",
	Tags: []string{"production", "ml"},
}

var vectorDBJSON = `
{
	"id": "da4e0206-d019-41d7-b51f-deadbeefbb8f",
	"name": "vectortest",
	"region": "nyc3",
	"owner_uuid": "880b7f98-f062-404d-b33c-458d545696f6",
	"status": "active",
	"config": {
		"default_quantization": "rq",
		"enable_auto_schema": true,
		"weaviate_version": "1.27.0"
	},
	"created_at": "2024-01-15T12:00:00Z",
	"updated_at": "2024-02-01T08:30:00Z",
	"endpoints": {
		"http": "https://vectortest-abc123.vectordb.digitalocean.com",
		"grpc": "vectortest-abc123.vectordb.digitalocean.com:443"
	},
	"size": "small",
	"tags": ["production", "ml"]
}
`

var vectorDBsJSON = fmt.Sprintf(`
{
	"vector_dbs": [%s],
	"total": 1
}
`, vectorDBJSON)

func TestVectorDBs_List(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/vector-databases", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, vectorDBsJSON)
	})

	got, resp, err := client.VectorDBs.List(ctx, nil)
	require.NoError(t, err)
	require.Equal(t, []VectorDB{vectorDB}, got)
	require.Equal(t, &Meta{Total: 1}, resp.Meta)
}

func TestVectorDBs_List_Pagination(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/vector-databases", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		assert.Equal(t, "2", r.URL.Query().Get("page"))
		assert.Equal(t, "10", r.URL.Query().Get("per_page"))
		fmt.Fprint(w, vectorDBsJSON)
	})

	_, _, err := client.VectorDBs.List(ctx, &ListOptions{Page: 2, PerPage: 10})
	require.NoError(t, err)
}

func TestVectorDBs_Get(t *testing.T) {
	setup()
	defer teardown()

	id := vectorDB.ID
	path := fmt.Sprintf("/v2/vector-databases/%s", id)

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprintf(w, `{"vector_db": %s}`, vectorDBJSON)
	})

	got, _, err := client.VectorDBs.Get(ctx, id)
	require.NoError(t, err)
	require.Equal(t, &vectorDB, got)
}

func TestVectorDBs_Create(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/vector-databases", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprintf(w, `{"vector_db": %s}`, vectorDBJSON)
	})

	createRequest := &VectorDBCreateRequest{
		Name:      "vectortest",
		Region:    "nyc3",
		Size:      "small",
		Tags:      []string{"production", "ml"},
		ProjectID: "49c369ee-c6a7-4c13-b8b4-ba0ac8d4180b",
	}

	got, _, err := client.VectorDBs.Create(ctx, createRequest)
	require.NoError(t, err)
	require.Equal(t, &vectorDB, got)
}

func TestVectorDBs_Update(t *testing.T) {
	setup()
	defer teardown()

	id := vectorDB.ID
	path := fmt.Sprintf("/v2/vector-databases/%s", id)

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		fmt.Fprintf(w, `{"vector_db": %s}`, vectorDBJSON)
	})

	updateRequest := &VectorDBUpdateRequest{
		ID: id,
		Config: &VectorDBConfig{
			DefaultQuantization: "pq",
			EnableAutoSchema:    false,
		},
	}

	got, _, err := client.VectorDBs.Update(ctx, id, updateRequest)
	require.NoError(t, err)
	require.Equal(t, &vectorDB, got)
}

func TestVectorDBs_Delete(t *testing.T) {
	setup()
	defer teardown()

	id := vectorDB.ID
	path := fmt.Sprintf("/v2/vector-databases/%s", id)

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
	})

	_, err := client.VectorDBs.Delete(ctx, id)
	require.NoError(t, err)
}

func TestVectorDBs_Resize(t *testing.T) {
	setup()
	defer teardown()

	id := vectorDB.ID
	path := fmt.Sprintf("/v2/vector-databases/%s/resize", id)

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprintf(w, `{"vector_db": %s}`, vectorDBJSON)
	})

	resizeRequest := &VectorDBResizeRequest{
		ID:   id,
		Size: "medium",
	}

	got, _, err := client.VectorDBs.Resize(ctx, id, resizeRequest)
	require.NoError(t, err)
	require.Equal(t, &vectorDB, got)
}

func TestVectorDBs_UpdateTags(t *testing.T) {
	setup()
	defer teardown()

	id := vectorDB.ID
	path := fmt.Sprintf("/v2/vector-databases/%s/tags", id)

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		fmt.Fprintf(w, `{"vector_db": %s}`, vectorDBJSON)
	})

	updateTagsRequest := &VectorDBUpdateTagsRequest{
		ID:   id,
		Tags: []string{"staging"},
	}

	got, _, err := client.VectorDBs.UpdateTags(ctx, id, updateTagsRequest)
	require.NoError(t, err)
	require.Equal(t, &vectorDB, got)
}

func TestVectorDBs_GetCredentials(t *testing.T) {
	setup()
	defer teardown()

	id := vectorDB.ID
	path := fmt.Sprintf("/v2/vector-databases/%s/credentials", id)

	want := &VectorDBAdminCredentials{
		UserID:   "weaviate-admin-user",
		APIToken: "secret-api-token",
	}

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"user_id": "weaviate-admin-user",
			"api_token": "secret-api-token"
		}`)
	})

	got, _, err := client.VectorDBs.GetCredentials(ctx, id)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestVectorDBs_ListBackups(t *testing.T) {
	setup()
	defer teardown()

	id := vectorDB.ID
	path := fmt.Sprintf("/v2/vector-databases/%s/backups", id)

	want := []VectorDBBackup{
		{
			BackupID:    "vectordb-da4e0206-20240101-120000",
			Status:      "SUCCESS",
			StartedAt:   time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			CompletedAt: time.Date(2024, 1, 1, 12, 5, 0, 0, time.UTC),
		},
	}

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"backups": [
				{
					"backup_id": "vectordb-da4e0206-20240101-120000",
					"status": "SUCCESS",
					"started_at": "2024-01-01T12:00:00Z",
					"completed_at": "2024-01-01T12:05:00Z"
				}
			]
		}`)
	})

	got, _, err := client.VectorDBs.ListBackups(ctx, id)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestVectorDBs_RestoreBackup(t *testing.T) {
	setup()
	defer teardown()

	id := vectorDB.ID
	backupID := "vectordb-da4e0206-20240101-120000"
	path := fmt.Sprintf("/v2/vector-databases/%s/backups/%s/restore", id, backupID)

	want := &VectorDBRestoreBackupResponse{
		BackupID: backupID,
		Status:   "STARTED",
	}

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprintf(w, `{
			"backup_id": %q,
			"status": "STARTED"
		}`, backupID)
	})

	restoreRequest := &VectorDBRestoreBackupRequest{
		ID:       id,
		BackupID: backupID,
	}

	got, _, err := client.VectorDBs.RestoreBackup(ctx, id, backupID, restoreRequest)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestVectorDBs_GetRestoreStatus(t *testing.T) {
	setup()
	defer teardown()

	id := vectorDB.ID
	backupID := "vectordb-da4e0206-20240101-120000"
	path := fmt.Sprintf("/v2/vector-databases/%s/backups/%s/restore", id, backupID)

	want := &VectorDBRestoreStatus{
		BackupID: backupID,
		Status:   "TRANSFERRING",
		Error:    "",
	}

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprintf(w, `{
			"backup_id": %q,
			"status": "TRANSFERRING"
		}`, backupID)
	})

	got, _, err := client.VectorDBs.GetRestoreStatus(ctx, id, backupID)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestVectorDBs_GetRestoreStatus_Failed(t *testing.T) {
	setup()
	defer teardown()

	id := vectorDB.ID
	backupID := "vectordb-da4e0206-20240101-120000"
	path := fmt.Sprintf("/v2/vector-databases/%s/backups/%s/restore", id, backupID)

	want := &VectorDBRestoreStatus{
		BackupID: backupID,
		Status:   "FAILED",
		Error:    "restore failed: disk full",
	}

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprintf(w, `{
			"backup_id": %q,
			"status": "FAILED",
			"error": "restore failed: disk full"
		}`, backupID)
	})

	got, _, err := client.VectorDBs.GetRestoreStatus(ctx, id, backupID)
	require.NoError(t, err)
	require.Equal(t, want, got)
}
