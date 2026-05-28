package godo

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	scanUUID    = "497dcba3-ecbf-4587-a2dd-5eb0665e6880"
	scan2UUID   = "7edb3b2e-869c-485b-af70-76a934e0fcfd"
	findingUUID = "50e14f43-dd4e-412f-864d-78943ea28d91"
)

func TestSecurityCreateScan(t *testing.T) {
	setup()
	defer teardown()

	createRequest := &CreateScanRequest{
		Resources: []string{"do:droplet"},
	}

	mux.HandleFunc("/v2/security/scans", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, `{"id": "`+scanUUID+`", "status": "COMPLETED", "created_at": "2025-12-04T00:00:00Z", "findings": []}`)
	})

	scan, resp, err := client.Security.CreateScan(context.Background(), createRequest)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, scanUUID, scan.ID)
	assert.Equal(t, "COMPLETED", scan.Status)
}

func TestSecurityListScans(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/security/scans", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		page := r.URL.Query().Get("page")
		if page == "2" {
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `{"scans": [{"id": "`+scan2UUID+`", "status": "COMPLETED", "created_at": "2025-12-05T00:00:00Z"}]}`)
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"scans": [{"id": "`+scanUUID+`", "status": "RUNNING", "created_at": "2025-12-04T00:00:00Z"}]}`)
	})

	scans, resp, err := client.Security.ListScans(context.Background(), &ListOptions{Page: 1})
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, scans, 1)
	assert.Equal(t, scanUUID, scans[0].ID)
	assert.Equal(t, "RUNNING", scans[0].Status)

	scans, resp, err = client.Security.ListScans(context.Background(), &ListOptions{Page: 2})
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, scans, 1)
	assert.Equal(t, scan2UUID, scans[0].ID)
	assert.Equal(t, "COMPLETED", scans[0].Status)
}

func TestSecurityGetScan(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("GET /v2/security/scans/{scanUUID}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("scanUUID")
		assert.Equal(t, "critical", r.URL.Query().Get("severity"))
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"scan": {"id": "`+id+`", "status": "COMPLETED", "created_at": "2025-12-04T00:00:00Z", "findings": [{"rule_uuid": "rule-1", "name": "test", "severity": "critical", "affected_resources_count": 2}]}}`)
	})

	opts := &ScanFindingsOptions{Severity: "critical", Type: "CSPM"}
	scan, resp, err := client.Security.GetScan(context.Background(), scanUUID, opts)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, scanUUID, scan.ID)
	assert.Len(t, scan.Findings, 1)
	assert.Equal(t, "critical", scan.Findings[0].Severity)
}

func TestSecurityGetLatestScan(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/security/scans/latest", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "high", r.URL.Query().Get("severity"))
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"scan": {"id": "`+scanUUID+`", "status": "COMPLETED", "created_at": "2025-12-06T00:00:00Z"}}`)
	})

	opts := &ScanFindingsOptions{Severity: "high"}
	scan, resp, err := client.Security.GetLatestScan(context.Background(), opts)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, scanUUID, scan.ID)
}

func TestSecurityListFindingAffectedResources(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("GET /v2/security/scans/{scanUUID}/findings/{findingUUID}/affected_resources", func(w http.ResponseWriter, r *http.Request) {
		page := r.URL.Query().Get("page")
		if page == "2" {
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `{"affected_resources": [{"urn": "do:droplet:2", "name": "droplet-2", "type": "Droplet"}]}`)
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"affected_resources": [{"urn": "do:droplet:1", "name": "droplet-1", "type": "Droplet"}]}`)
	})

	resources, resp, err := client.Security.ListFindingAffectedResources(
		context.Background(),
		&ListFindingAffectedResourcesRequest{
			ScanUUID:    scanUUID,
			FindingUUID: findingUUID,
		},
		&ListOptions{Page: 1},
	)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resources, 1)
	assert.Equal(t, "do:droplet:1", resources[0].URN)

	resources, resp, err = client.Security.ListFindingAffectedResources(
		context.Background(),
		&ListFindingAffectedResourcesRequest{
			ScanUUID:    scanUUID,
			FindingUUID: findingUUID,
		},
		&ListOptions{Page: 2},
	)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resources, 1)
	assert.Equal(t, "do:droplet:2", resources[0].URN)
}
