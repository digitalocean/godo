package godo

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNfsCreate(t *testing.T) {
	setup()
	defer teardown()

	createRequest := &NfsCreateRequest{
		Name:    "test-nfs-share",
		SizeGib: 50,
		Region:  "atl1",
	}

	mux.HandleFunc("/v2/nfs", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, `{"share": {"id": "test-nfs-id", "name": "test-nfs-share", "size_gib": 50, "region": "atl1", "status": "CREATING", "created_at":"2023-10-01T00:00:00Z", "vpc_ids": []}}`)
	})

	share, resp, err := client.Nfs.Create(context.Background(), createRequest)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "test-nfs-share", share.Name)
	assert.Equal(t, "atl1", share.Region)
	assert.Equal(t, 50, share.SizeGib)
	assert.Equal(t, NfsShareCreating, share.Status)

	invalidCreateRequest := &NfsCreateRequest{
		Name:    "test-nfs-share-invalid-size",
		SizeGib: 20,
		Region:  "atl1",
	}

	share, resp, err = client.Nfs.Create(context.Background(), invalidCreateRequest)
	assert.Error(t, err)
	assert.Equal(t, "size_gib is invalid because it cannot be less than 50Gib", err.Error())
	assert.Nil(t, share)
	assert.Nil(t, resp)
}

func TestNfsDelete(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/nfs/test-nfs-id", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		w.WriteHeader(http.StatusNoContent)
	})

	resp, err := client.Nfs.Delete(context.Background(), "test-nfs-id", "atl1")
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestNfsGet(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/nfs/test-nfs-id", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"share": {"id": "test-nfs-id", "name": "test-nfs-share", "size_gib": 50, "region": "atl1", "status": "ACTIVE", "created_at":"2023-10-01T00:00:00Z", "vpc_ids": []}}`)
	})

	share, resp, err := client.Nfs.Get(context.Background(), "test-nfs-id", "atl1")
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "test-nfs-share", share.Name)
	assert.Equal(t, "atl1", share.Region)
	assert.Equal(t, 50, share.SizeGib)
	assert.Equal(t, NfsShareActive, share.Status)
}

func TestNfsList(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/nfs", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		page := r.URL.Query().Get("page")
		if page == "2" {
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `{"shares": [{"id": "test-nfs-id-2", "name": "test-nfs-share-2", "size_gib": 50, "region": "atl1", "status": "ACTIVE", "created_at":"2023-10-01T00:00:00Z", "vpc_ids": []}]}`)
		} else {
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `{"shares": [{"id": "test-nfs-id-1", "name": "test-nfs-share-1", "size_gib": 50, "region": "atl1", "status": "CREATING", "created_at":"2023-10-01T00:00:00Z", "vpc_ids": []}]}`)
		}
	})

	// Test first page
	shares, resp, err := client.Nfs.List(context.Background(), &ListOptions{Page: 1}, "atl1")
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, shares, 1)
	assert.Equal(t, "test-nfs-share-1", shares[0].Name)
	assert.Equal(t, "atl1", shares[0].Region)
	assert.Equal(t, 50, shares[0].SizeGib)
	assert.Equal(t, NfsShareCreating, shares[0].Status)

	// Test second page
	shares, resp, err = client.Nfs.List(context.Background(), &ListOptions{Page: 2}, "atl1")
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, shares, 1)
	assert.Equal(t, "test-nfs-share-2", shares[0].Name)
	assert.Equal(t, "atl1", shares[0].Region)
	assert.Equal(t, 50, shares[0].SizeGib)
	assert.Equal(t, NfsShareActive, shares[0].Status)
}

func TestNfsSnapshotGet(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/nfs/snapshots/test-nfs-snapshot-id", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{ "snapshot": {"id": "test-nfs-snapshot-id", "name": "daily-backup", "size_gib": 1024, "region": "atl1", "status": "ACTIVE", "created_at": "2023-11-14T16:29:21Z", "share_id": "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d"}}`)
	})

	snapshot, resp, err := client.Nfs.GetSnapshot(context.Background(), "test-nfs-snapshot-id", "atl1")
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "daily-backup", snapshot.Name)
	assert.Equal(t, "atl1", snapshot.Region)
	assert.Equal(t, 1024, snapshot.SizeGib)
	assert.Equal(t, NfsSnapshotActive, snapshot.Status)
}

func TestNfsListSnapshots(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/nfs/snapshots", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		// Check that region query parameter is present
		page := r.URL.Query().Get("page")
		if page == "2" {
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `{"snapshots": [{"id": "test-nfs-snapshot-id-2", "name": "daily-backup-2", "size_gib": 2048, "region": "atl1", "status": "ACTIVE", "created_at": "2023-11-14T16:29:21Z", "share_id": "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d"}]}`)
		} else {
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `{"snapshots": [{"id": "test-nfs-snapshot-id-1", "name": "daily-backup-1", "size_gib": 1024, "region": "atl1", "status": "CREATING", "created_at": "2023-11-14T16:29:21Z", "share_id": "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d"}]}`)
		}
	})

	// Test first page
	snapshots, resp, err := client.Nfs.ListSnapshots(context.Background(), &ListOptions{Page: 1}, "", "atl1")
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, snapshots, 1)
	assert.Equal(t, "daily-backup-1", snapshots[0].Name)
	assert.Equal(t, "atl1", snapshots[0].Region)
	assert.Equal(t, 1024, snapshots[0].SizeGib)
	assert.Equal(t, NfsSnapshotCreating, snapshots[0].Status)

	// Test second page
	snapshots, resp, err = client.Nfs.ListSnapshots(context.Background(), &ListOptions{Page: 2}, "", "atl1")
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, snapshots, 1)
	assert.Equal(t, "daily-backup-2", snapshots[0].Name)
	assert.Equal(t, "atl1", snapshots[0].Region)
	assert.Equal(t, 2048, snapshots[0].SizeGib)
	assert.Equal(t, NfsSnapshotActive, snapshots[0].Status)
}

func TestNfsSnapshotsDelete(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/nfs/snapshots/my-snapshot-id", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		w.WriteHeader(http.StatusNoContent)
	})

	resp, err := client.Nfs.DeleteSnapshot(context.Background(), "my-snapshot-id", "atl1")
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}
