package godo

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSpacesKeyCreate(t *testing.T) {
	setup()
	defer teardown()

	createRequest := &SpacesKeyCreateRequest{
		Name: "test-key",
		Grants: []*Grant{
			{
				Bucket:     "test-bucket",
				Permission: PermissionRead,
			},
		},
	}

	mux.HandleFunc("/v2/spaces/keys", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, `{"key":{"name":"test-key","access_key":"test-access-key","secret_key":"test-secret-key","created_at":"2023-10-01T00:00:00Z"}}`)
	})

	key, resp, err := client.SpacesKeys.CreateSpacesKey(context.Background(), createRequest)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "test-key", key.Name)
	assert.Equal(t, "test-access-key", key.AccessKey)
	assert.Equal(t, "test-secret-key", key.SecretKey)
}

func TestSpacesKeyDelete(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/spaces/keys/test-access-key", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		w.WriteHeader(http.StatusNoContent)
	})

	resp, err := client.SpacesKeys.DeleteSpacesKey(context.Background(), "test-access-key")
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestSpacesKeyUpdate(t *testing.T) {
	setup()
	defer teardown()

	updateRequest := &SpacesKeyUpdateRequest{
		Name: "updated-key",
		Grants: []*Grant{
			{
				Bucket:     "updated-bucket",
				Permission: PermissionReadWrite,
			},
		},
	}

	mux.HandleFunc("/v2/spaces/keys/test-access-key", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"key":{"name":"updated-key","access_key":"test-access-key","created_at":"2023-10-01T00:00:00Z"}}`)
	})

	key, resp, err := client.SpacesKeys.UpdateSpacesKey(context.Background(), "test-access-key", updateRequest)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "updated-key", key.Name)
	assert.Equal(t, "test-access-key", key.AccessKey)
}

func TestSpacesKeyList(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/spaces/keys", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"keys":[{"name":"test-key","access_key":"test-access-key","created_at":"2023-10-01T00:00:00Z"}]}`)
	})

	keys, resp, err := client.SpacesKeys.ListSpacesKeys(context.Background(), nil)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, keys, 1)
	assert.Equal(t, "test-key", keys[0].Name)
	assert.Equal(t, "test-access-key", keys[0].AccessKey)
}

func TestSpacesKeyList_Pagination(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/spaces/keys", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		page := r.URL.Query().Get("page")
		if page == "2" {
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `{"keys":[{"name":"test-key-2","access_key":"test-access-key-2","created_at":"2023-10-02T00:00:00Z"}]}`)
		} else {
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `{"keys":[{"name":"test-key-1","access_key":"test-access-key-1","created_at":"2023-10-01T00:00:00Z"}]}`)
		}
	})

	// Test first page
	keys, resp, err := client.SpacesKeys.ListSpacesKeys(context.Background(), &ListOptions{Page: 1})
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, keys, 1)
	assert.Equal(t, "test-key-1", keys[0].Name)
	assert.Equal(t, "test-access-key-1", keys[0].AccessKey)

	// Test second page
	keys, resp, err = client.SpacesKeys.ListSpacesKeys(context.Background(), &ListOptions{Page: 2})
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, keys, 1)
	assert.Equal(t, "test-key-2", keys[0].Name)
	assert.Equal(t, "test-access-key-2", keys[0].AccessKey)
}
