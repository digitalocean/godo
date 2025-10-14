package godo

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNfsResize(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/nfs/my-nfs-id/actions", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, `{"action": {"id": 1, "status": "in-progress", "type": "resize", "resource_type": "network_file_share", "region_slug": "atl1", "started_at": "2025-10-14T11:55:31.615157397Z"}}`)
	})

	action, resp, err := client.NfsActions.Resize(context.Background(), "my-nfs-id", 1024, "atl1")
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "atl1", action.RegionSlug)
	assert.Equal(t, "network_file_share", action.ResourceType)
	assert.Equal(t, "in-progress", action.Status)
	assert.Equal(t, "resize", action.Type)
}

func TestNfsSnapshot(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/nfs/my-nfs-id/actions", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, `{"action": {"id": 1, "status": "in-progress", "type": "snapshot", "resource_type": "network_file_share", "region_slug": "atl1", "started_at": "2025-10-14T11:55:31.615157397Z"}}`)
	})

	action, resp, err := client.NfsActions.Snapshot(context.Background(), "my-nfs-id", "my-snapshot", "atl1")
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "atl1", action.RegionSlug)
	assert.Equal(t, "network_file_share", action.ResourceType)
	assert.Equal(t, "in-progress", action.Status)
	assert.Equal(t, "snapshot", action.Type)
}
