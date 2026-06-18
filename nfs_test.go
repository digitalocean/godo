package godo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNfsCreate(t *testing.T) {
	setup()
	defer teardown()

	createRequest := &NfsCreateRequest{
		Name:            "test-nfs-share",
		SizeGib:         50,
		Region:          "atl1",
		PerformanceTier: "standard",
	}

	mux.HandleFunc("/v2/nfs", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, `{"share": {"id": "test-nfs-id", "name": "test-nfs-share", "size_gib": 50, "region": "atl1", "status": "CREATING", "created_at":"2023-10-01T00:00:00Z", "vpc_ids": [], "host": "10.128.32.2", "mount_path": "/123456/test-nfs-id", "performance_tier": "standard"}}`)
	})

	share, resp, err := client.Nfs.Create(context.Background(), createRequest)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "test-nfs-share", share.Name)
	assert.Equal(t, "atl1", share.Region)
	assert.Equal(t, 50, share.SizeGib)
	assert.Equal(t, NfsShareCreating, share.Status)
	assert.Equal(t, "10.128.32.2", share.Host)
	assert.Equal(t, "/123456/test-nfs-id", share.MountPath)
	assert.Equal(t, "standard", share.PerformanceTier)

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
		fmt.Fprint(w, `{"share": {"id": "test-nfs-id", "name": "test-nfs-share", "size_gib": 50, "region": "atl1", "status": "ACTIVE", "created_at":"2023-10-01T00:00:00Z", "vpc_ids": [], "host": "10.128.32.2", "mount_path": "/123456/test-nfs-id", "performance_tier": "standard"}}`)
	})

	share, resp, err := client.Nfs.Get(context.Background(), "test-nfs-id", "atl1")
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "test-nfs-share", share.Name)
	assert.Equal(t, "atl1", share.Region)
	assert.Equal(t, 50, share.SizeGib)
	assert.Equal(t, NfsShareActive, share.Status)
	assert.Equal(t, "10.128.32.2", share.Host)
	assert.Equal(t, "/123456/test-nfs-id", share.MountPath)
	assert.Equal(t, "standard", share.PerformanceTier)
}

func TestNfsList(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/nfs", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		page := r.URL.Query().Get("page")
		if page == "2" {
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `{"shares": [{"id": "test-nfs-id-2", "name": "test-nfs-share-2", "size_gib": 50, "region": "atl1", "status": "ACTIVE", "created_at":"2023-10-01T00:00:00Z", "vpc_ids": [], "host": "10.128.32.3", "mount_path": "/123456/test-nfs-id-2", "performance_tier": "standard"}]}`)
		} else {
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `{"shares": [{"id": "test-nfs-id-1", "name": "test-nfs-share-1", "size_gib": 50, "region": "atl1", "status": "CREATING", "created_at":"2023-10-01T00:00:00Z", "vpc_ids": [], "host": "10.128.32.2", "mount_path": "/123456/test-nfs-id-1", "performance_tier": "standard"}]}`)
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
	assert.Equal(t, "10.128.32.2", shares[0].Host)
	assert.Equal(t, "/123456/test-nfs-id-1", shares[0].MountPath)
	assert.Equal(t, "standard", shares[0].PerformanceTier)
	// Test second page
	shares, resp, err = client.Nfs.List(context.Background(), &ListOptions{Page: 2}, "atl1")
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, shares, 1)
	assert.Equal(t, "test-nfs-share-2", shares[0].Name)
	assert.Equal(t, "atl1", shares[0].Region)
	assert.Equal(t, 50, shares[0].SizeGib)
	assert.Equal(t, NfsShareActive, shares[0].Status)
	assert.Equal(t, "10.128.32.3", shares[0].Host)
	assert.Equal(t, "/123456/test-nfs-id-2", shares[0].MountPath)
	assert.Equal(t, "standard", shares[0].PerformanceTier)
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

func TestNfsCreateAccessPoint(t *testing.T) {
	setup()
	defer teardown()

	vpcID := "my-vpc-id"
	createRequest := &NfsCreateAccessPointRequest{
		Name: "my-access-point",
		Path: "/exports/data",
		AccessPolicy: NfsAccessPolicy{
			Anonuid:                    65534,
			Anongid:                    65534,
			Protocols:                  []NfsAccessPolicyProtocol{NfsAccessPolicyProtocolNFS4},
			SquashConfig:               NfsSquashConfigRootSquash,
			IdentityEnforcementEnabled: true,
		},
		VpcID: &vpcID,
	}

	mux.HandleFunc("/v2/nfs/shares/test-share-id/access_points", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)

		var got NfsCreateAccessPointRequest
		err := json.NewDecoder(r.Body).Decode(&got)
		assert.NoError(t, err)
		assert.Equal(t, createRequest.Name, got.Name)
		assert.Equal(t, createRequest.Path, got.Path)
		assert.Equal(t, createRequest.AccessPolicy.SquashConfig, got.AccessPolicy.SquashConfig)
		assert.Equal(t, createRequest.AccessPolicy.Protocols, got.AccessPolicy.Protocols)
		assert.NotNil(t, got.VpcID)
		assert.Equal(t, *createRequest.VpcID, *got.VpcID)

		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, `{"access_point":{"id":"test-access-point-id","name":"my-access-point","share_id":"test-share-id","path":"/exports/data","status":"ACCESS_POINT_CREATING","access_policy":{"anonuid":65534,"anongid":65534,"protocols":["NFS4"],"squash_config":"ROOT_SQUASH","identity_enforcement_enabled":true},"created_at":"2026-06-15T00:00:00Z","updated_at":"2026-06-15T00:00:00Z","is_default":false,"vpc_id":"my-vpc-id"},"action":{"id":"1","status":"in-progress","type":"CREATE_ACCESS_POINT","resource_type":"network_file_share","resource_id":"test-share-id","region_slug":"nyc3","started_at":"2026-06-15T00:00:00Z"}}`)
	})

	apResp, resp, err := client.Nfs.CreateAccessPoint(context.Background(), "test-share-id", createRequest)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "test-access-point-id", apResp.AccessPoint.ID)
	assert.Equal(t, NfsAccessPointCreating, apResp.AccessPoint.Status)
	assert.Equal(t, "CREATE_ACCESS_POINT", apResp.Action.Type)
}

func TestNfsGetAccessPoint(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/nfs/access_points/test-access-point-id", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"access_point":{"id":"test-access-point-id","name":"my-access-point","share_id":"test-share-id","path":"/exports/data","status":"ACCESS_POINT_ACTIVE","access_policy":{"anonuid":65534,"anongid":65534,"protocols":["NFS4"],"squash_config":"ROOT_SQUASH","identity_enforcement_enabled":false},"created_at":"2026-06-15T00:00:00Z","updated_at":"2026-06-15T01:00:00Z","is_default":true}}`)
	})

	ap, resp, err := client.Nfs.GetAccessPoint(context.Background(), "test-access-point-id")
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "my-access-point", ap.Name)
	assert.Equal(t, NfsAccessPointActive, ap.Status)
	assert.Equal(t, NfsSquashConfigRootSquash, ap.AccessPolicy.SquashConfig)
}

func TestNfsListAccessPoints(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/nfs/shares/test-share-id/access_points", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, string(NfsAccessPointActive), r.URL.Query().Get("status"))
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"access_points":[{"id":"test-access-point-id","name":"my-access-point","share_id":"test-share-id","path":"/exports/data","status":"ACCESS_POINT_ACTIVE","access_policy":{"anonuid":65534,"anongid":65534,"protocols":["NFS","NFS4"],"squash_config":"NO_SQUASH","identity_enforcement_enabled":false},"created_at":"2026-06-15T00:00:00Z","updated_at":"2026-06-15T01:00:00Z","is_default":false}]}`)
	})

	accessPoints, resp, err := client.Nfs.ListAccessPoints(context.Background(), "test-share-id", &NfsListAccessPointsOptions{Status: NfsAccessPointActive})
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, accessPoints, 1)
	assert.Equal(t, "test-access-point-id", accessPoints[0].ID)
	assert.Equal(t, NfsSquashConfigNoSquash, accessPoints[0].AccessPolicy.SquashConfig)
}

func TestNfsDeleteAccessPoint(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/nfs/access_points/test-access-point-id", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"access_point":{"id":"test-access-point-id","name":"my-access-point","share_id":"test-share-id","path":"/exports/data","status":"ACCESS_POINT_DELETED","access_policy":{"anonuid":65534,"anongid":65534,"protocols":["NFS4"],"squash_config":"ROOT_SQUASH","identity_enforcement_enabled":false},"created_at":"2026-06-15T00:00:00Z","updated_at":"2026-06-15T01:00:00Z","is_default":false},"action":{"id":"2","status":"completed","type":"DELETE_ACCESS_POINT","resource_type":"network_file_share","resource_id":"test-share-id","region_slug":"nyc3","started_at":"2026-06-15T02:00:00Z"}}`)
	})

	apResp, resp, err := client.Nfs.DeleteAccessPoint(context.Background(), "test-access-point-id")
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, NfsAccessPointDeleted, apResp.AccessPoint.Status)
	assert.Equal(t, "DELETE_ACCESS_POINT", apResp.Action.Type)
}

