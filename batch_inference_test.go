package godo

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

// batchInferenceSetup creates a test server and configures the client so that
// BatchInferenceServiceOp routes requests to the local mock server.
func batchInferenceSetup() func() {
	setup()

	// Point the batch-inference service's base URL at the test server.
	u, _ := url.Parse(server.URL + "/")
	client.BatchInference.(*BatchInferenceServiceOp).baseURL = u

	return teardown
}

var (
	batchFileUploadResponse = `{
  "file_id": "b7d562e0-49ae-43f7-820d-6a37a2f05435",
  "upload_url": "https://spaces.example.com/batch-inputs/1/b7d562e0.jsonl?token=xyz",
  "expires_at": "2026-04-21T17:49:27.536939587Z"
}`

	batchCreateResponse = `{
  "batch_id": "11f13da9-64a4-fe5b-8567-a23ae3abd3e2",
  "cancel_requested_at": null,
  "completion_window": "24h",
  "created_at": "2026-04-21T17:41:57Z",
  "expires_at": "2026-04-21T17:49:28Z",
  "file_id": "b7d562e0-49ae-43f7-820d-6a37a2f05435",
  "provider": "anthropic",
  "request_counts": {
    "total": 0,
    "completed": 0,
    "failed": 0
  },
  "request_id": "postman-1776793316",
  "result_available": false,
  "status": "queued",
  "updated_at": "2026-04-21T17:41:57Z"
}`

	batchGetResponse = `{
  "batch_id": "11f13da9-64a4-fe5b-8567-a23ae3abd3e2",
  "cancel_requested_at": null,
  "completion_window": "24h",
  "created_at": "2026-04-21T17:41:57Z",
  "expires_at": "2026-04-21T17:49:28Z",
  "file_id": "b7d562e0-49ae-43f7-820d-6a37a2f05435",
  "provider": "anthropic",
  "request_counts": {
    "total": 2,
    "completed": 2,
    "failed": 0
  },
  "request_id": "postman-1776793316",
  "result_available": true,
  "status": "completed",
  "updated_at": "2026-04-21T17:44:26Z"
}`

	batchCancelResponse = `{
  "batch_id": "11f13dad-0b91-104b-8567-a23ae3abd3e2",
  "cancel_requested_at": "2026-04-21T18:10:00Z",
  "completion_window": "24h",
  "created_at": "2026-04-21T18:05:00Z",
  "expires_at": null,
  "file_id": "b7d562e0-49ae-43f7-820d-6a37a2f05435",
  "provider": "anthropic",
  "request_counts": {
    "total": 2,
    "completed": 0,
    "failed": 0
  },
  "request_id": "postman-cancel-test",
  "result_available": false,
  "status": "cancelled",
  "updated_at": "2026-04-21T18:10:00Z"
}`

	batchListResponse = `{
  "edges": [
    {
      "cursor": "eyJjIjoiMjAyNi0wNC0yMVQxNzo0MTo1N1oiLCJpIjoyMH0=",
      "node": {
        "batch_id": "11f13da9-64a4-fe5b-8567-a23ae3abd3e2",
        "cancel_requested_at": null,
        "completion_window": "24h",
        "created_at": "2026-04-21T17:41:57Z",
        "expires_at": null,
        "provider": "anthropic",
        "request_counts": {
          "total": 2,
          "completed": 2,
          "failed": 0
        },
        "request_id": "postman-1776793316",
        "result_available": true,
        "status": "completed",
        "updated_at": "2026-04-21T17:44:26Z"
      }
    },
    {
      "cursor": "eyJjIjoiMjAyNi0wNC0yMFQwNToyMjo1MVoiLCJpIjoxM30=",
      "node": {
        "batch_id": "11f13c78-fa24-cbfc-8567-a23ae3abd3e2",
        "cancel_requested_at": null,
        "completion_window": "24h",
        "created_at": "2026-04-20T05:22:51Z",
        "expires_at": null,
        "provider": "openai",
        "request_counts": {
          "total": 2,
          "completed": 2,
          "failed": 0
        },
        "request_id": "postman-1776662571",
        "result_available": true,
        "status": "completed",
        "updated_at": "2026-04-20T05:25:12Z"
      }
    }
  ],
  "page_info": {
    "endCursor": "eyJjIjoiMjAyNi0wNC0yMFQwNToyMjo1MVoiLCJpIjoxM30=",
    "hasNextPage": true
  }
}`

	batchResultsResponse = `{
  "download": {
    "presigned_url": "https://spaces.example.com/batch-results/1/output.jsonl?token=abc",
    "expires_at": "2026-04-21T18:07:47.228188268Z"
  },
  "output_file_id": "msgbatch_015UytMT8cCLD8332Hb3BhNe"
}`
)

func TestBatchInference_CreatePresignedUploadURL(t *testing.T) {
	cleanup := batchInferenceSetup()
	defer cleanup()

	mux.HandleFunc("/v1/batches/files", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)

		var req CreateBatchFileRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if req.FileName != "input.jsonl" {
			t.Errorf("expected FileName 'input.jsonl', got '%s'", req.FileName)
		}
		fmt.Fprint(w, batchFileUploadResponse)
	})

	upload, _, err := client.BatchInference.CreatePresignedUploadURL(ctx, &CreateBatchFileRequest{
		FileName: "input.jsonl",
	})
	if err != nil {
		t.Fatalf("CreatePresignedUploadURL: %v", err)
	}
	if upload.FileID != "b7d562e0-49ae-43f7-820d-6a37a2f05435" {
		t.Errorf("expected FileID 'b7d562e0-49ae-43f7-820d-6a37a2f05435', got '%s'", upload.FileID)
	}
	if upload.UploadURL != "https://spaces.example.com/batch-inputs/1/b7d562e0.jsonl?token=xyz" {
		t.Errorf("unexpected UploadURL: %s", upload.UploadURL)
	}
	if upload.ExpiresAt != "2026-04-21T17:49:27.536939587Z" {
		t.Errorf("expected ExpiresAt '2026-04-21T17:49:27.536939587Z', got '%s'", upload.ExpiresAt)
	}
}

func TestBatchInference_UploadInputFile(t *testing.T) {
	cleanup := batchInferenceSetup()
	defer cleanup()

	body := `{"custom_id":"req-1","method":"POST","url":"/v1/messages","body":{"model":"claude-sonnet-4-20250514","max_tokens":1024,"messages":[{"role":"user","content":"Hello"}]}}` + "\n"

	mux.HandleFunc("/upload/test.jsonl", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)

		if ct := r.Header.Get("Content-Type"); ct != "application/x-ndjson" {
			t.Errorf("expected Content-Type 'application/x-ndjson', got '%s'", ct)
		}
		if auth := r.Header.Get("Authorization"); auth != "" {
			t.Errorf("expected no Authorization header, got '%s'", auth)
		}

		got, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		if string(got) != body {
			t.Errorf("unexpected body: %s", string(got))
		}

		w.WriteHeader(http.StatusOK)
	})

	uploadURL := server.URL + "/upload/test.jsonl"
	resp, err := client.BatchInference.UploadInputFile(ctx, uploadURL, strings.NewReader(body))
	if err != nil {
		t.Fatalf("UploadInputFile: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestBatchInference_CreateJob(t *testing.T) {
	cleanup := batchInferenceSetup()
	defer cleanup()

	mux.HandleFunc("/v1/batches", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			return
		}
		testMethod(t, r, http.MethodPost)

		var req CreateBatchRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if req.Provider != "anthropic" {
			t.Errorf("expected Provider 'anthropic', got '%s'", req.Provider)
		}
		if req.FileID != "b7d562e0-49ae-43f7-820d-6a37a2f05435" {
			t.Errorf("expected FileID 'b7d562e0-49ae-43f7-820d-6a37a2f05435', got '%s'", req.FileID)
		}
		if req.CompletionWindow != "24h" {
			t.Errorf("expected CompletionWindow '24h', got '%s'", req.CompletionWindow)
		}
		fmt.Fprint(w, batchCreateResponse)
	})

	batch, _, err := client.BatchInference.CreateJob(ctx, &CreateBatchRequest{
		Provider:         "anthropic",
		FileID:           "b7d562e0-49ae-43f7-820d-6a37a2f05435",
		CompletionWindow: "24h",
		RequestID:        "postman-1776793316",
	})
	if err != nil {
		t.Fatalf("CreateJob: %v", err)
	}
	if batch.BatchID != "11f13da9-64a4-fe5b-8567-a23ae3abd3e2" {
		t.Errorf("expected BatchID '11f13da9-64a4-fe5b-8567-a23ae3abd3e2', got '%s'", batch.BatchID)
	}
	if batch.Provider != "anthropic" {
		t.Errorf("expected Provider 'anthropic', got '%s'", batch.Provider)
	}
	if batch.Status != "queued" {
		t.Errorf("expected Status 'queued', got '%s'", batch.Status)
	}
	if batch.CreatedAt != "2026-04-21T17:41:57Z" {
		t.Errorf("expected CreatedAt '2026-04-21T17:41:57Z', got '%s'", batch.CreatedAt)
	}
	if batch.RequestCounts == nil {
		t.Fatal("expected RequestCounts to be non-nil")
	}
	if batch.RequestCounts.Total != 0 {
		t.Errorf("expected RequestCounts.Total 0, got %d", batch.RequestCounts.Total)
	}
	if batch.ResultAvailable {
		t.Error("expected ResultAvailable to be false")
	}
}

func TestBatchInference_CreateJobOpenAI(t *testing.T) {
	cleanup := batchInferenceSetup()
	defer cleanup()

	mux.HandleFunc("/v1/batches", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			return
		}
		testMethod(t, r, http.MethodPost)

		var req CreateBatchRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if req.Provider != "openai" {
			t.Errorf("expected Provider 'openai', got '%s'", req.Provider)
		}
		if req.Endpoint != "/v1/chat/completions" {
			t.Errorf("expected Endpoint '/v1/chat/completions', got '%s'", req.Endpoint)
		}

		resp := `{
  "batch_id": "11f13c78-fa24-cbfc-8567-a23ae3abd3e2",
  "cancel_requested_at": null,
  "completion_window": "24h",
  "created_at": "2026-04-20T05:22:51Z",
  "expires_at": null,
  "file_id": "b7d562e0-49ae-43f7-820d-6a37a2f05435",
  "provider": "openai",
  "request_counts": {"total": 0, "completed": 0, "failed": 0},
  "request_id": "openai-test",
  "result_available": false,
  "status": "queued",
  "updated_at": "2026-04-20T05:22:51Z"
}`
		fmt.Fprint(w, resp)
	})

	batch, _, err := client.BatchInference.CreateJob(ctx, &CreateBatchRequest{
		Provider:         "openai",
		FileID:           "b7d562e0-49ae-43f7-820d-6a37a2f05435",
		CompletionWindow: "24h",
		Endpoint:         "/v1/chat/completions",
		RequestID:        "openai-test",
	})
	if err != nil {
		t.Fatalf("CreateJob: %v", err)
	}
	if batch.Provider != "openai" {
		t.Errorf("expected Provider 'openai', got '%s'", batch.Provider)
	}
}

func TestBatchInference_GetJob(t *testing.T) {
	cleanup := batchInferenceSetup()
	defer cleanup()

	mux.HandleFunc("/v1/batches/11f13da9-64a4-fe5b-8567-a23ae3abd3e2", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, batchGetResponse)
	})

	batch, _, err := client.BatchInference.GetJob(ctx, "11f13da9-64a4-fe5b-8567-a23ae3abd3e2")
	if err != nil {
		t.Fatalf("GetJob: %v", err)
	}
	if batch.BatchID != "11f13da9-64a4-fe5b-8567-a23ae3abd3e2" {
		t.Errorf("expected BatchID '11f13da9-64a4-fe5b-8567-a23ae3abd3e2', got '%s'", batch.BatchID)
	}
	if batch.Status != "completed" {
		t.Errorf("expected Status 'completed', got '%s'", batch.Status)
	}
	if batch.Provider != "anthropic" {
		t.Errorf("expected Provider 'anthropic', got '%s'", batch.Provider)
	}
	if batch.RequestCounts == nil {
		t.Fatal("expected RequestCounts to be non-nil")
	}
	if batch.RequestCounts.Completed != 2 {
		t.Errorf("expected RequestCounts.Completed 2, got %d", batch.RequestCounts.Completed)
	}
	if batch.RequestCounts.Total != 2 {
		t.Errorf("expected RequestCounts.Total 2, got %d", batch.RequestCounts.Total)
	}
	if !batch.ResultAvailable {
		t.Error("expected ResultAvailable to be true")
	}
}

func TestBatchInference_CancelJob(t *testing.T) {
	cleanup := batchInferenceSetup()
	defer cleanup()

	mux.HandleFunc("/v1/batches/11f13dad-0b91-104b-8567-a23ae3abd3e2/cancel", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, batchCancelResponse)
	})

	batch, _, err := client.BatchInference.CancelJob(ctx, "11f13dad-0b91-104b-8567-a23ae3abd3e2")
	if err != nil {
		t.Fatalf("CancelJob: %v", err)
	}
	if batch.BatchID != "11f13dad-0b91-104b-8567-a23ae3abd3e2" {
		t.Errorf("expected BatchID '11f13dad-0b91-104b-8567-a23ae3abd3e2', got '%s'", batch.BatchID)
	}
	if batch.Status != "cancelled" {
		t.Errorf("expected Status 'cancelled', got '%s'", batch.Status)
	}
	if batch.CancelRequestedAt == nil || *batch.CancelRequestedAt != "2026-04-21T18:10:00Z" {
		t.Errorf("expected CancelRequestedAt '2026-04-21T18:10:00Z', got %v", batch.CancelRequestedAt)
	}
}

func TestBatchInference_ListJobs(t *testing.T) {
	cleanup := batchInferenceSetup()
	defer cleanup()

	mux.HandleFunc("/v1/batches", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)

		status := r.URL.Query().Get("status")
		if status != "completed" {
			t.Errorf("expected query status=completed, got '%s'", status)
		}
		limit := r.URL.Query().Get("limit")
		if limit != "10" {
			t.Errorf("expected query limit=10, got '%s'", limit)
		}
		fmt.Fprint(w, batchListResponse)
	})

	list, _, err := client.BatchInference.ListJobs(ctx, &ListBatchesOptions{
		Status: "completed",
		Limit:  10,
	})
	if err != nil {
		t.Fatalf("ListJobs: %v", err)
	}
	if len(list.Edges) != 2 {
		t.Fatalf("expected 2 edges, got %d", len(list.Edges))
	}
	if list.Edges[0].Node.BatchID != "11f13da9-64a4-fe5b-8567-a23ae3abd3e2" {
		t.Errorf("expected first batch BatchID '11f13da9-64a4-fe5b-8567-a23ae3abd3e2', got '%s'", list.Edges[0].Node.BatchID)
	}
	if list.Edges[0].Node.Provider != "anthropic" {
		t.Errorf("expected first batch Provider 'anthropic', got '%s'", list.Edges[0].Node.Provider)
	}
	if list.Edges[0].Node.Status != "completed" {
		t.Errorf("expected first batch Status 'completed', got '%s'", list.Edges[0].Node.Status)
	}
	if list.Edges[0].Cursor != "eyJjIjoiMjAyNi0wNC0yMVQxNzo0MTo1N1oiLCJpIjoyMH0=" {
		t.Errorf("unexpected first edge cursor: %s", list.Edges[0].Cursor)
	}
	if list.Edges[1].Node.BatchID != "11f13c78-fa24-cbfc-8567-a23ae3abd3e2" {
		t.Errorf("expected second batch BatchID '11f13c78-fa24-cbfc-8567-a23ae3abd3e2', got '%s'", list.Edges[1].Node.BatchID)
	}
	if list.Edges[1].Node.Provider != "openai" {
		t.Errorf("expected second batch Provider 'openai', got '%s'", list.Edges[1].Node.Provider)
	}
	if !list.PageInfo.HasNextPage {
		t.Error("expected HasNextPage to be true")
	}
	if list.PageInfo.EndCursor != "eyJjIjoiMjAyNi0wNC0yMFQwNToyMjo1MVoiLCJpIjoxM30=" {
		t.Errorf("unexpected EndCursor: %s", list.PageInfo.EndCursor)
	}
}

func TestBatchInference_GetJobResult(t *testing.T) {
	cleanup := batchInferenceSetup()
	defer cleanup()

	mux.HandleFunc("/v1/batches/11f13da9-64a4-fe5b-8567-a23ae3abd3e2/results", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, batchResultsResponse)
	})

	results, _, err := client.BatchInference.GetJobResult(ctx, "11f13da9-64a4-fe5b-8567-a23ae3abd3e2")
	if err != nil {
		t.Fatalf("GetJobResult: %v", err)
	}
	if results.Download.PresignedURL != "https://spaces.example.com/batch-results/1/output.jsonl?token=abc" {
		t.Errorf("unexpected PresignedURL: %s", results.Download.PresignedURL)
	}
	if results.Download.ExpiresAt != "2026-04-21T18:07:47.228188268Z" {
		t.Errorf("unexpected Download.ExpiresAt: %s", results.Download.ExpiresAt)
	}
	if results.OutputFileID != "msgbatch_015UytMT8cCLD8332Hb3BhNe" {
		t.Errorf("unexpected OutputFileID: %s", results.OutputFileID)
	}
}
