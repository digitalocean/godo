package godo

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

var listKnowledgeBaseResponse = `
{
	"knowledge_bases": [
		{
			"created_at": "2025-05-08T03:37:28Z",
			"database_id": "11111111-1111-1111-1111-111111111111",
			"embedding_model_uuid": "11111111-1111-1111-1111-111111111111",
			"last_indexing_job": {
				"created_at": "2025-05-08T03:37:28Z",
				"finished_at": "2025-05-08T03:38:28Z",
				"knowledge_base_uuid": "11111111-1111-1111-1111-111111111111",
				"phase": "BATCH_JOB_PHASE_SUCCEEDED",
				"started_at": "2025-05-08T03:37:28Z",
				"updated_at": "2025-05-08T03:38:28Z",
				"total_datasources": 1,
				"completed_datasources": 1,
				"tokens": 5000,
				"uuid": "11111111-1111-1111-1111-111111111111"
			},
			"name": "Test Knowledge Base",
			"project_id": "11111111-1111-1111-1111-111111111111",
			"region": "tor1",
			"tags": ["test"],
			"uuid": "11111111-1111-1111-1111-111111111111",
			"updated_at": "2025-05-09T16:02:39Z",
			"is_public": false,
			"user_id": "14567334"
		}
	],
	"links": {
		"pages": {
			"first": "https://api.digitalocean.com/v2/gen-ai/knowledge_bases?page=1&per_page=1",
			"previous": "https://api.digitalocean.com/v2/gen-ai/knowledge_bases?page=1&per_page=1",
			"next": "https://api.digitalocean.com/v2/gen-ai/knowledge_bases?page=3&per_page=1",
			"last": "https://api.digitalocean.com/v2/gen-ai/knowledge_bases?page=10&per_page=1"
		}
	},
	"meta": {
		"total": 10,
		"page": 1,
		"pages": 10
	}
}
`

var knowledgeBaseResponse = `
{
	"knowledge_base": {
		"uuid": "11111111-1111-1111-1111-111111111111",
		"name": "testing-kb",
		"created_at": "2025-05-14T13:18:05Z",
		"updated_at": "2025-05-14T13:18:05Z",
		"region": "tor1",
		"project_id": "11111111-1111-1111-1111-111111111111",
		"embedding_model_uuid": "11111111-1111-1111-1111-111111111111",
		"database_id": "11111111-1111-1111-1111-111111111111",
		"is_public": false,
		"tags": ["string"],
		"user_id": "18919793"
	}
}
`

var knowledgeBaseUpdateResponse = `
{
	"knowledge_base": {
		"uuid": "11111111-1111-1111-1111-111111111111",
		"name": "Updated Knowledge Base",
		"created_at": "2025-05-14T13:18:05Z",
		"updated_at": "2025-05-14T13:46:48Z",
		"region": "tor1",
		"project_id": "11111111-1111-1111-1111-111111111111",
		"embedding_model_uuid": "11111111-1111-1111-1111-111111111111",
		"database_id": "11111111-1111-1111-1111-111111111111",
		"is_public": true,
		"tags": ["updated", "example"],
		"user_id": "18919793"
	}
}
`

var listDataSourcesResponse = `
{
	"knowledge_base_data_sources": [
		{
			"uuid": "22222222-2222-2222-2222-222222222222",
			"bucket_name": "test-bucket",
			"item_path": "/docs/test.pdf",
			"region": "tor1",
			"created_at": "2025-05-14T13:18:05Z",
			"updated_at": "2025-05-14T13:18:05Z",
			"spaces_data_source": {
				"bucket_name": "test-bucket",
				"item_path": "/docs/test.pdf",
				"region": "tor1"
			}
		}
	],
	"links": {
		"pages": {
			"first": "https://api.digitalocean.com/v2/gen-ai/knowledge_bases/11111111-1111-1111-1111-111111111111/data_sources?page=1&per_page=1",
			"previous": "https://api.digitalocean.com/v2/gen-ai/knowledge_bases/11111111-1111-1111-1111-111111111111/data_sources?page=1&per_page=1",
			"next": "https://api.digitalocean.com/v2/gen-ai/knowledge_bases/11111111-1111-1111-1111-111111111111/data_sources?page=2&per_page=1",
			"last": "https://api.digitalocean.com/v2/gen-ai/knowledge_bases/11111111-1111-1111-1111-111111111111/data_sources?page=3&per_page=1"
		}
	},
	"meta": {
		"total": 3,
		"page": 1,
		"pages": 3
	}
}
`

var addDataSourceResponse = `
{
	"knowledge_base_data_source": {
		"uuid": "22222222-2222-2222-2222-222222222222",
		"bucket_name": "test-bucket",
		"item_path": "/docs/test.pdf",
		"region": "tor1",
		"created_at": "2025-05-14T13:18:05Z",
		"updated_at": "2025-05-14T13:18:05Z",
		"spaces_data_source": {
			"bucket_name": "test-bucket",
			"item_path": "/docs/test.pdf",
			"region": "tor1"
		}
	}
}
`

var deleteDataSourceResponse = `
{
	"knowledge_base_uuid": "11111111-1111-1111-1111-111111111111",
	"data_source_uuid": "22222222-2222-2222-2222-222222222222"
}
`

var deleteKnowledgeBaseResponse = `
{
	"uuid": "11111111-1111-1111-1111-111111111111"
}
`

var agentResponse = `
{
	"agent": {
		"uuid": "00000000-0000-0000-0000-000000000000",
		"name": "testing-godo",
		"created_at": "2025-05-14T13:18:05Z",
		"updated_at": "2025-05-14T13:18:05Z",
		"instruction": "You are an agent who thinks deeply about the world",
		"description": "My Agent Description",
		"knowledge_bases": [
			{
				"uuid": "11111111-1111-1111-1111-111111111111",
				"name": "testing-kb",
				"created_at": "2025-05-14T07:52:27Z",
				"updated_at": "2025-05-14T07:57:27Z",
				"region": "tor1",
				"embedding_model_uuid": "11111111-1111-1111-1111-111111111111",
				"project_id": "11111111-1111-1111-1111-111111111111",
				"database_id": "11111111-1111-1111-1111-111111111111"
			}
		],
		"k": 10,
		"temperature": 0.7,
		"top_p": 0.9,
		"max_tokens": 512,
		"project_id": "00000000-0000-0000-0000-000000000000",
		"region": "tor1",
		"user_id": "18919793",
		"retrieval_method": "RETRIEVAL_METHOD_NONE"
	}
}
`

func TestListKnowledgeBases(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/gen-ai/knowledge_bases", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		testFormValues(t, r, values{
			"page":     "1",
			"per_page": "1",
		})

		fmt.Fprint(w, listKnowledgeBaseResponse)
	})

	req := &ListOptions{
		Page:    1,
		PerPage: 1,
	}

	knowledgeBases, resp, err := client.GenAI.ListKnowledgeBases(ctx, req)
	if err != nil {
		t.Errorf("GenAI.ListKnowledgeBases returned error: %v", err)
	}

	assert.Equal(t, 10, resp.Meta.Total)
	assert.Equal(t, "Test Knowledge Base", knowledgeBases[0].Name)
}

func TestCreateKnowledgeBase(t *testing.T) { //works
	setup()
	defer teardown()

	mux.HandleFunc("/v2/gen-ai/knowledge_bases", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, knowledgeBaseResponse)
	})

	req := &KnowledgeBaseCreateRequest{
		Name:               "testing-kb",
		ProjectID:          "11111111-1111-1111-1111-111111111111",
		Region:             "tor1",
		EmbeddingModelUUID: "11111111-1111-1111-1111-111111111111",
		Tags:               []string{"string"},
	}

	res, _, err := client.GenAI.CreateKnowledgeBase(ctx, req)
	if err != nil {
		t.Errorf("GenAI.CreateKnowledgeBase returned error: %v", err)
	}

	assert.Equal(t, res.Name, req.Name)
	assert.Equal(t, res.ProjectId, req.ProjectID)
}

func TestListDataSources(t *testing.T) { //works
	setup()
	defer teardown()

	mux.HandleFunc("/v2/gen-ai/knowledge_bases/11111111-1111-1111-1111-111111111111/data_sources", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		testFormValues(t, r, values{
			"page":     "1",
			"per_page": "1",
		})

		fmt.Fprint(w, listDataSourcesResponse)
	})

	req := &ListOptions{
		Page:    1,
		PerPage: 1,
	}

	dataSources, resp, err := client.GenAI.ListDataSources(ctx, "11111111-1111-1111-1111-111111111111", req)
	if err != nil {
		t.Errorf("GenAI.ListDataSources returned error: %v", err)
	}

	assert.Equal(t, 3, resp.Meta.Total)
	assert.Equal(t, "test-bucket", dataSources[0].BucketName)
}

func TestAddDataSource(t *testing.T) { //works
	setup()
	defer teardown()

	mux.HandleFunc("/v2/gen-ai/knowledge_bases/11111111-1111-1111-1111-111111111111/data_sources", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, addDataSourceResponse)
	})

	req := &AddDataSourceRequest{
		KnowledgeBaseUUID: "11111111-1111-1111-1111-111111111111",
		SpacesDataSource: &SpacesDataSource{
			BucketName: "test-bucket",
			ItemPath:   "/docs/test.pdf",
			Region:     "tor1",
		},
	}

	res, _, err := client.GenAI.AddDataSource(ctx, "11111111-1111-1111-1111-111111111111", req)
	if err != nil {
		t.Errorf("GenAI.AddDataSource returned error: %v", err)
	}

	assert.Equal(t, "test-bucket", res.BucketName)
	assert.Equal(t, "/docs/test.pdf", res.ItemPath)
}

func TestDeleteDataSource(t *testing.T) { //works
	setup()
	defer teardown()

	mux.HandleFunc("/v2/gen-ai/knowledge_bases/11111111-1111-1111-1111-111111111111/data_sources/22222222-2222-2222-2222-222222222222", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		fmt.Fprint(w, deleteDataSourceResponse)
	})

	kbUUID, dsUUID, resp, err := client.GenAI.DeleteDataSource(ctx, "11111111-1111-1111-1111-111111111111", "22222222-2222-2222-2222-222222222222")
	if err != nil {
		t.Errorf("GenAI.DeleteDataSource returned error: %v", err)
	}

	assert.Equal(t, "11111111-1111-1111-1111-111111111111", kbUUID)
	assert.Equal(t, "22222222-2222-2222-2222-222222222222", dsUUID)
	assert.Equal(t, 200, resp.Response.StatusCode)
}

func TestGetKnowledgeBase(t *testing.T) { //works
	setup()
	defer teardown()

	mux.HandleFunc("/v2/gen-ai/knowledge_bases/11111111-1111-1111-1111-111111111111", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, knowledgeBaseResponse)
	})

	res, resp, err := client.GenAI.GetKnowledgeBase(ctx, "11111111-1111-1111-1111-111111111111")
	if err != nil {
		t.Errorf("GenAI.GetKnowledgeBase returned error: %v", err)
	}

	assert.Equal(t, "testing-kb", res.Name)
	assert.Equal(t, 200, resp.Response.StatusCode)
}

func TestUpdateKnowledgeBase(t *testing.T) { //works
	setup()
	defer teardown()

	mux.HandleFunc("/v2/gen-ai/knowledge_bases/11111111-1111-1111-1111-111111111111", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		fmt.Fprint(w, knowledgeBaseUpdateResponse)
	})

	req := &UpdateKnowledgeBaseRequest{
		Name: "Updated Knowledge Base",
		Tags: []string{"updated", "example"},
	}

	res, resp, err := client.GenAI.UpdateKnowledgeBase(ctx, "11111111-1111-1111-1111-111111111111", req)
	if err != nil {
		t.Errorf("GenAI.UpdateKnowledgeBase returned error: %v", err)
	}

	assert.Equal(t, res.Tags[0], req.Tags[0])
	assert.Equal(t, 200, resp.Response.StatusCode)
}

func TestDeleteKnowledgeBase(t *testing.T) { //works
	setup()
	defer teardown()

	mux.HandleFunc("/v2/gen-ai/knowledge_bases/11111111-1111-1111-1111-111111111111", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		fmt.Fprint(w, deleteKnowledgeBaseResponse)
	})

	kbUUID, resp, err := client.GenAI.DeleteKnowledgeBase(ctx, "11111111-1111-1111-1111-111111111111")
	if err != nil {
		t.Errorf("GenAI.DeleteKnowledgeBase returned error: %v", err)
	}

	assert.Equal(t, "11111111-1111-1111-1111-111111111111", kbUUID)
	assert.Equal(t, 200, resp.Response.StatusCode)
}

func TestAttachKnowledgeBase(t *testing.T) { //fail
	setup()
	defer teardown()

	mux.HandleFunc("/v2/gen-ai/agents/00000000-0000-0000-0000-000000000000/knowledge_bases/11111111-1111-1111-1111-111111111111", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, agentResponse)
	})

	res, resp, err := client.GenAI.AttachKnowledgBase(ctx, "00000000-0000-0000-0000-000000000000", "11111111-1111-1111-1111-111111111111")
	if err != nil {
		t.Errorf("GenAI.AttachKnowledgBase returned error: %v", err)
	}

	assert.Equal(t, "testing-godo", res.Name)
	assert.Equal(t, 200, resp.Response.StatusCode)
}

func TestDetachKnowledgeBase(t *testing.T) {
	setup()
	defer teardown()

	// Mock the expected API endpoint
	mux.HandleFunc("/v2/gen-ai/agents/00000000-0000-0000-0000-000000000000/knowledge_bases/11111111-1111-1111-1111-111111111111", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		fmt.Fprint(w, agentResponse) // Mock response
	})

	// Call the method
	res, resp, err := client.GenAI.DetachKnowledgBase(ctx, "00000000-0000-0000-0000-000000000000", "11111111-1111-1111-1111-111111111111")
	fmt.Print(res)
	fmt.Print(resp)

	if err != nil {
		t.Fatalf("GenAI.DetachKnowledgBase returned error: %v", err)
	}

	// Validate the response
	if res == nil {
		t.Fatalf("GenAI.DetachKnowledgBase returned nil response or result")
	}

	assert.Equal(t, "testing-godo", res.Name)      // Validate the agent name
	assert.Equal(t, 200, resp.Response.StatusCode) // Validate the HTTP status code
}
