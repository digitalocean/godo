package godo

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"testing"
	"time"
)

var (
	diListJSONResponse = `
{
  "dedicated_inferences": [
    {
      "id": "di-uuid-1",
      "name": "test-di-1",
      "region": "s2r1",
      "status": "active",
      "vpc_uuid": "246de291-05af-461f-956a-a7be58f65367",
      "endpoints": {
        "public_endpoint_fqdn": "test-di-1.di.s2r1.digitalocean.com",
        "private_endpoint_fqdn": "test-di-1.internal.di.s2r1.digitalocean.com"
      },
      "created_at": "2024-01-09T20:44:32Z",
      "updated_at": "2024-01-09T20:44:32Z"
    },
    {
      "id": "di-uuid-2",
      "name": "test-di-2",
      "region": "s2r1",
      "status": "provisioning",
      "vpc_uuid": "246de291-05af-461f-956a-a7be58f65367",
      "created_at": "2024-01-09T21:00:00Z",
      "updated_at": "2024-01-09T21:00:00Z"
    }
  ],
  "links": {
    "pages": {
      "last": "https://api.digitalocean.com/v2/dedicated-inferences?page=1",
      "next": ""
    }
  },
  "meta": {
    "total": 2
  }
}
`
	diUpdateJSONResponse = `
{
  "dedicated_inference": {
    "id": "di-uuid",
    "name": "test-di-updated",
    "region": "s2r1",
    "status": "updating",
    "vpc_uuid": "246de291-05af-461f-956a-a7be58f65367",
    "endpoints": {
      "public_endpoint_fqdn": "test-di.di.s2r1.digitalocean.com",
      "private_endpoint_fqdn": "test-di.internal.di.s2r1.digitalocean.com"
    },
    "spec": {
      "version": 1,
      "id": "spec-uuid",
      "dedicated_inference_id": "di-uuid",
      "state": "active",
      "enable_public_endpoint": true,
      "vpc_config": {
        "vpc_uuid": "246de291-05af-461f-956a-a7be58f65367"
      },
      "model_deployments": [
        {
          "model_id": "model-uuid",
          "model_slug": "meta-llama/Llama-3.1-8B-Instruct",
          "model_provider": "hugging_face",
          "accelerators": [
            {
              "accelerator_id": "acc-uuid",
              "accelerator_slug": "gpu-mi300x1-192gb",
              "state": "active",
              "type": "prefill_decode",
              "scale": 1
            }
          ]
        }
      ],
      "created_at": "2024-01-09T20:44:32Z",
      "updated_at": "2024-01-09T20:44:32Z"
    },
    "pending_deployment_spec": {
      "version": 2,
      "id": "spec-uuid-2",
      "dedicated_inference_id": "di-uuid",
      "state": "pending",
      "enable_public_endpoint": true,
      "vpc_config": {
        "vpc_uuid": "246de291-05af-461f-956a-a7be58f65367"
      },
      "model_deployments": [
        {
          "model_id": "model-uuid",
          "model_slug": "meta-llama/Llama-3.1-8B-Instruct",
          "model_provider": "hugging_face",
          "accelerators": [
            {
              "accelerator_id": "acc-uuid",
              "accelerator_slug": "gpu-mi300x1-192gb",
              "state": "pending",
              "type": "prefill_decode",
              "scale": 2
            }
          ]
        }
      ],
      "created_at": "2024-01-09T20:50:00Z",
      "updated_at": "2024-01-09T20:50:00Z"
    },
    "created_at": "2024-01-09T20:44:32Z",
    "updated_at": "2024-01-09T20:50:00Z"
  }
}
`
	diCreateJSONResponse = `
{
  "dedicated_inference": {
    "id": "di-uuid",
    "name": "test-di",
    "region": "s2r1",
    "status": "provisioning",
    "vpc_uuid": "246de291-05af-461f-956a-a7be58f65367",
    "pending_deployment_spec": {
      "version": 1,
      "id": "spec-uuid",
      "dedicated_inference_id": "di-uuid",
      "state": "pending",
      "enable_public_endpoint": true,
      "vpc_config": {
        "vpc_uuid": "246de291-05af-461f-956a-a7be58f65367"
      },
      "model_deployments": [
        {
          "model_id": "",
          "model_slug": "meta-llama/Llama-3.1-8B-Instruct",
          "model_provider": "hugging_face",
          "accelerators": [
            {
              "accelerator_id": "",
              "accelerator_slug": "gpu-mi300x1-192gb",
              "state": "invalid",
              "type": "prefill_decode",
              "scale": 1
            }
          ]
        }
      ],
      "created_at": "1970-01-01T00:00:00Z",
      "updated_at": "1970-01-01T00:00:00Z"
    },
    "created_at": "1970-01-01T00:00:00Z",
    "updated_at": "1970-01-01T00:00:00Z"
  },
  "token": {
    "id": "auth-token-uuid",
    "name": "auto-generated-during-provisioning",
    "value": "auth-token-value",
    "created_at": "1970-01-01T00:00:00Z"
  }
}
`
	diGetJSONResponse = `
{
  "dedicated_inference": {
    "id": "di-uuid",
    "name": "test-di",
    "region": "s2r1",
    "status": "active",
    "vpc_uuid": "246de291-05af-461f-956a-a7be58f65367",
    "endpoints": {
      "public_endpoint_fqdn": "test-di.di.s2r1.digitalocean.com",
      "private_endpoint_fqdn": "test-di.internal.di.s2r1.digitalocean.com"
    },
    "spec": {
      "version": 1,
      "id": "spec-uuid",
      "dedicated_inference_id": "di-uuid",
      "state": "active",
      "enable_public_endpoint": true,
      "vpc_config": {
        "vpc_uuid": "246de291-05af-461f-956a-a7be58f65367"
      },
      "model_deployments": [
        {
          "model_id": "model-uuid",
          "model_slug": "meta-llama/Llama-3.1-8B-Instruct",
          "model_provider": "hugging_face",
          "accelerators": [
            {
              "accelerator_id": "acc-uuid",
              "accelerator_slug": "gpu-mi300x1-192gb",
              "state": "active",
              "type": "prefill_decode",
              "scale": 1
            }
          ]
        }
      ],
      "created_at": "2024-01-09T20:44:32Z",
      "updated_at": "2024-01-09T20:44:32Z"
    },
    "created_at": "2024-01-09T20:44:32Z",
    "updated_at": "2024-01-09T20:44:32Z"
  }
}
`
	diListAcceleratorsJSONResponse = `
{
  "accelerators": [
    {
      "id": "acc-uuid-1",
      "name": "gpu-acc-1",
      "slug": "gpu-mi300x1-192gb",
      "status": "running",
      "created_at": "2024-01-09T20:44:32Z"
    },
    {
      "id": "acc-uuid-2",
      "name": "gpu-acc-2",
      "slug": "gpu-mi300x1-192gb",
      "status": "provisioning",
      "created_at": "2024-01-09T21:00:00Z"
    }
  ],
  "links": {
    "pages": {
      "last": "https://api.digitalocean.com/v2/dedicated-inferences/di-uuid/accelerators?page=1",
      "next": ""
    }
  },
  "meta": {
    "total": 2
  }
}
`
	diCreateTokenJSONResponse = `
{
  "token": {
    "id": "token-uuid-123",
    "name": "new-inference-token",
    "value": "test-token-value-placeholder",
    "created_at": "2024-01-09T20:44:32Z"
  }
}
`
	diListTokensJSONResponse = `
{
  "tokens": [
    {
      "id": "token-uuid-1",
      "name": "first-token",
      "created_at": "2024-01-09T20:44:32Z"
    },
    {
      "id": "token-uuid-2",
      "name": "second-token",
      "created_at": "2024-01-09T21:00:00Z"
    }
  ],
  "links": {
    "pages": {
      "last": "https://api.digitalocean.com/v2/dedicated-inferences/di-uuid/tokens?page=1",
      "next": ""
    }
  },
  "meta": {
    "total": 2
  }
}
`
)

func TestDedicatedInference_Create(t *testing.T) {
	setup()
	defer teardown()

	createRequest := &DedicatedInferenceCreateRequest{
		Spec: &DedicatedInferenceSpecRequest{
			Version:              1,
			Name:                 "test-di",
			Region:               "s2r1",
			EnablePublicEndpoint: true,
			VPC: &DedicatedInferenceVPCRequest{
				UUID: "246de291-05af-461f-956a-a7be58f65367",
			},
			ModelDeployments: []*DedicatedInferenceModelRequest{
				{
					ModelSlug:      "meta-llama/Llama-3.1-8B-Instruct",
					ModelProvider:  "hugging_face",
					WorkloadConfig: &DedicatedInferenceWorkloadConfig{},
					Accelerators: []*DedicatedInferenceAcceleratorRequest{
						{
							AcceleratorSlug: "gpu-mi300x1-192gb",
							Scale:           1,
							Type:            "prefill_decode",
						},
					},
				},
			},
		},
		Secrets: &DedicatedInferenceSecrets{
			HuggingFaceToken: "test-hf-token-placeholder",
		},
	}

	mux.HandleFunc("/v2/dedicated-inferences", func(w http.ResponseWriter, r *http.Request) {
		v := new(DedicatedInferenceCreateRequest)
		err := json.NewDecoder(r.Body).Decode(v)
		if err != nil {
			t.Fatal(err)
		}

		testMethod(t, r, http.MethodPost)

		if v.Spec.Name != "test-di" {
			t.Errorf("Request body name = %q, expected %q", v.Spec.Name, "test-di")
		}
		if v.Spec.Region != "s2r1" {
			t.Errorf("Request body region = %q, expected %q", v.Spec.Region, "s2r1")
		}
		if v.Spec.VPC.UUID != "246de291-05af-461f-956a-a7be58f65367" {
			t.Errorf("Request body VPC UUID = %q, expected %q", v.Spec.VPC.UUID, "246de291-05af-461f-956a-a7be58f65367")
		}
		if len(v.Spec.ModelDeployments) != 1 {
			t.Fatalf("Request body model deployments count = %d, expected 1", len(v.Spec.ModelDeployments))
		}
		if v.Spec.ModelDeployments[0].ModelSlug != "meta-llama/Llama-3.1-8B-Instruct" {
			t.Errorf("Request body model slug = %q, expected %q", v.Spec.ModelDeployments[0].ModelSlug, "meta-llama/Llama-3.1-8B-Instruct")
		}

		w.WriteHeader(http.StatusAccepted)
		fmt.Fprint(w, diCreateJSONResponse)
	})

	di, token, resp, err := client.DedicatedInference.Create(ctx, createRequest)
	if err != nil {
		t.Errorf("DedicatedInference.Create returned error: %v", err)
	}

	if resp.StatusCode != http.StatusAccepted {
		t.Errorf("expected status %d, got %d", http.StatusAccepted, resp.StatusCode)
	}

	zeroTime := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)

	expectedDI := &DedicatedInference{
		ID:      "di-uuid",
		Name:    "test-di",
		Region:  "s2r1",
		Status:  "provisioning",
		VPCUUID: "246de291-05af-461f-956a-a7be58f65367",
		PendingDeploymentSpec: &DedicatedInferenceDeployment{
			Version:              1,
			ID:                   "spec-uuid",
			DedicatedInferenceID: "di-uuid",
			State:                "pending",
			EnablePublicEndpoint: true,
			VPCConfig: &DedicatedInferenceVPCConfig{
				VPCUUID: "246de291-05af-461f-956a-a7be58f65367",
			},
			ModelDeployments: []*DedicatedInferenceModelDeployment{
				{
					ModelID:       "",
					ModelSlug:     "meta-llama/Llama-3.1-8B-Instruct",
					ModelProvider: "hugging_face",
					Accelerators: []*DedicatedInferenceAccelerator{
						{
							AcceleratorID:   "",
							AcceleratorSlug: "gpu-mi300x1-192gb",
							State:           "invalid",
							Type:            "prefill_decode",
							Scale:           1,
						},
					},
				},
			},
			CreatedAt: zeroTime,
			UpdatedAt: zeroTime,
		},
		CreatedAt: zeroTime,
		UpdatedAt: zeroTime,
	}

	if !reflect.DeepEqual(di, expectedDI) {
		t.Errorf("DedicatedInference.Create returned %+v, expected %+v", di, expectedDI)
	}

	expectedToken := &DedicatedInferenceToken{
		ID:        "auth-token-uuid",
		Name:      "auto-generated-during-provisioning",
		Value:     "auth-token-value",
		CreatedAt: zeroTime,
	}

	if !reflect.DeepEqual(token, expectedToken) {
		t.Errorf("DedicatedInference.Create token = %+v, expected %+v", token, expectedToken)
	}
}

func TestDedicatedInference_Get(t *testing.T) {
	setup()
	defer teardown()

	diID := "di-uuid"

	mux.HandleFunc(fmt.Sprintf("/v2/dedicated-inferences/%s", diID), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, diGetJSONResponse)
	})

	di, _, err := client.DedicatedInference.Get(ctx, diID)
	if err != nil {
		t.Errorf("DedicatedInference.Get returned error: %v", err)
	}

	createdAt, _ := time.Parse(time.RFC3339, "2024-01-09T20:44:32Z")

	expectedDI := &DedicatedInference{
		ID:      "di-uuid",
		Name:    "test-di",
		Region:  "s2r1",
		Status:  "active",
		VPCUUID: "246de291-05af-461f-956a-a7be58f65367",
		Endpoints: &DedicatedInferenceEndpoints{
			PublicEndpointFQDN:  "test-di.di.s2r1.digitalocean.com",
			PrivateEndpointFQDN: "test-di.internal.di.s2r1.digitalocean.com",
		},
		DeploymentSpec: &DedicatedInferenceDeployment{
			Version:              1,
			ID:                   "spec-uuid",
			DedicatedInferenceID: "di-uuid",
			State:                "active",
			EnablePublicEndpoint: true,
			VPCConfig: &DedicatedInferenceVPCConfig{
				VPCUUID: "246de291-05af-461f-956a-a7be58f65367",
			},
			ModelDeployments: []*DedicatedInferenceModelDeployment{
				{
					ModelID:       "model-uuid",
					ModelSlug:     "meta-llama/Llama-3.1-8B-Instruct",
					ModelProvider: "hugging_face",
					Accelerators: []*DedicatedInferenceAccelerator{
						{
							AcceleratorID:   "acc-uuid",
							AcceleratorSlug: "gpu-mi300x1-192gb",
							State:           "active",
							Type:            "prefill_decode",
							Scale:           1,
						},
					},
				},
			},
			CreatedAt: createdAt,
			UpdatedAt: createdAt,
		},
		CreatedAt: createdAt,
		UpdatedAt: createdAt,
	}

	if !reflect.DeepEqual(di, expectedDI) {
		t.Errorf("DedicatedInference.Get returned %+v, expected %+v", di, expectedDI)
	}
}

func TestDedicatedInference_Get_NoToken(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/dedicated-inferences/di-uuid", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, diGetJSONResponse)
	})

	di, _, err := client.DedicatedInference.Get(ctx, "di-uuid")
	if err != nil {
		t.Errorf("DedicatedInference.Get returned error: %v", err)
	}

	if di.ID != "di-uuid" {
		t.Errorf("expected ID %q, got %q", "di-uuid", di.ID)
	}

	if di.Status != "active" {
		t.Errorf("expected Status %q, got %q", "active", di.Status)
	}

	if di.Endpoints == nil {
		t.Fatal("expected endpoints, got nil")
	}

	if di.Endpoints.PublicEndpointFQDN != "test-di.di.s2r1.digitalocean.com" {
		t.Errorf("expected public fqdn %q, got %q", "test-di.di.s2r1.digitalocean.com", di.Endpoints.PublicEndpointFQDN)
	}
}

func TestDedicatedInference_String(t *testing.T) {
	di := DedicatedInference{
		ID:     "di-123",
		Name:   "test",
		Region: "nyc2",
		Status: "active",
	}

	if di.String() == "" {
		t.Error("DedicatedInference.String() returned empty string")
	}
}

func TestDedicatedInferenceToken_String(t *testing.T) {
	token := DedicatedInferenceToken{
		ID:   "token-123",
		Name: "test-token",
	}

	if token.String() == "" {
		t.Error("DedicatedInferenceToken.String() returned empty string")
	}
}

func TestDedicatedInference_Update(t *testing.T) {
	setup()
	defer teardown()

	diID := "di-uuid"

	mux.HandleFunc(fmt.Sprintf("/v2/dedicated-inferences/%s", diID), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPatch)

		var req DedicatedInferenceUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request body: %v", err)
		}

		if req.Spec == nil {
			t.Error("expected spec in request")
		}

		if req.Spec.Name != "test-di-updated" {
			t.Errorf("expected name %q, got %q", "test-di-updated", req.Spec.Name)
		}

		w.WriteHeader(http.StatusAccepted)
		fmt.Fprint(w, diUpdateJSONResponse)
	})

	updateReq := &DedicatedInferenceUpdateRequest{
		Spec: &DedicatedInferenceSpecRequest{
			Version:              2,
			Name:                 "test-di-updated",
			Region:               "s2r1",
			EnablePublicEndpoint: true,
			VPC: &DedicatedInferenceVPCRequest{
				UUID: "246de291-05af-461f-956a-a7be58f65367",
			},
			ModelDeployments: []*DedicatedInferenceModelRequest{
				{
					ModelID:       "model-uuid",
					ModelSlug:     "meta-llama/Llama-3.1-8B-Instruct",
					ModelProvider: "hugging_face",
					Accelerators: []*DedicatedInferenceAcceleratorRequest{
						{
							AcceleratorSlug: "gpu-mi300x1-192gb",
							Scale:           2,
							Type:            "prefill_decode",
						},
					},
				},
			},
		},
		Secrets: &DedicatedInferenceSecrets{
			HuggingFaceToken: "test-hf-token-placeholder",
		},
	}

	di, _, err := client.DedicatedInference.Update(ctx, diID, updateReq)
	if err != nil {
		t.Errorf("DedicatedInference.Update returned error: %v", err)
	}

	if di.ID != diID {
		t.Errorf("expected ID %q, got %q", diID, di.ID)
	}

	if di.Name != "test-di-updated" {
		t.Errorf("expected name %q, got %q", "test-di-updated", di.Name)
	}

	if di.Status != "updating" {
		t.Errorf("expected status %q, got %q", "updating", di.Status)
	}

	if di.PendingDeploymentSpec == nil {
		t.Fatal("expected pending_deployment_spec, got nil")
	}

	if di.PendingDeploymentSpec.ModelDeployments[0].Accelerators[0].Scale != 2 {
		t.Errorf("expected scale 2, got %d", di.PendingDeploymentSpec.ModelDeployments[0].Accelerators[0].Scale)
	}
}

func TestDedicatedInference_Delete(t *testing.T) {
	setup()
	defer teardown()

	diID := "di-uuid"

	mux.HandleFunc(fmt.Sprintf("/v2/dedicated-inferences/%s", diID), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		w.WriteHeader(http.StatusAccepted)
	})

	resp, err := client.DedicatedInference.Delete(ctx, diID)
	if err != nil {
		t.Errorf("DedicatedInference.Delete returned error: %v", err)
	}

	if resp.StatusCode != http.StatusAccepted {
		t.Errorf("expected status %d, got %d", http.StatusAccepted, resp.StatusCode)
	}
}

func TestDedicatedInference_List(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/dedicated-inferences", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, diListJSONResponse)
	})

	diList, resp, err := client.DedicatedInference.List(ctx, nil)
	if err != nil {
		t.Fatalf("DedicatedInference.List returned error: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	if len(diList) != 2 {
		t.Fatalf("expected 2 dedicated inferences, got %d", len(diList))
	}

	if diList[0].ID != "di-uuid-1" {
		t.Errorf("expected ID %q, got %q", "di-uuid-1", diList[0].ID)
	}

	if diList[0].Name != "test-di-1" {
		t.Errorf("expected Name %q, got %q", "test-di-1", diList[0].Name)
	}

	if diList[0].Status != "active" {
		t.Errorf("expected Status %q, got %q", "active", diList[0].Status)
	}

	if diList[1].Status != "provisioning" {
		t.Errorf("expected Status %q, got %q", "provisioning", diList[1].Status)
	}

	if resp.Meta == nil || resp.Meta.Total != 2 {
		t.Errorf("expected Meta.Total to be 2")
	}
}

func TestDedicatedInference_ListWithOptions(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/dedicated-inferences", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)

		if got := r.URL.Query().Get("region"); got != "s2r1" {
			t.Errorf("expected region query param %q, got %q", "s2r1", got)
		}
		if got := r.URL.Query().Get("page"); got != "1" {
			t.Errorf("expected page query param %q, got %q", "1", got)
		}
		if got := r.URL.Query().Get("per_page"); got != "10" {
			t.Errorf("expected per_page query param %q, got %q", "10", got)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, diListJSONResponse)
	})

	opts := &DedicatedInferenceListOptions{
		Region: "s2r1",
		ListOptions: ListOptions{
			Page:    1,
			PerPage: 10,
		},
	}

	diList, _, err := client.DedicatedInference.List(ctx, opts)
	if err != nil {
		t.Fatalf("DedicatedInference.List returned error: %v", err)
	}

	if len(diList) != 2 {
		t.Fatalf("expected 2 dedicated inferences, got %d", len(diList))
	}
}

func TestDedicatedInference_ListAccelerators(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/dedicated-inferences/di-uuid/accelerators", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, diListAcceleratorsJSONResponse)
	})

	accelerators, resp, err := client.DedicatedInference.ListAccelerators(ctx, "di-uuid", nil)
	if err != nil {
		t.Fatalf("DedicatedInference.ListAccelerators returned error: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	if len(accelerators) != 2 {
		t.Fatalf("expected 2 accelerators, got %d", len(accelerators))
	}

	if accelerators[0].ID != "acc-uuid-1" {
		t.Errorf("expected ID %q, got %q", "acc-uuid-1", accelerators[0].ID)
	}

	if accelerators[0].Slug != "gpu-mi300x1-192gb" {
		t.Errorf("expected Slug %q, got %q", "gpu-mi300x1-192gb", accelerators[0].Slug)
	}

	if accelerators[0].Status != "running" {
		t.Errorf("expected Status %q, got %q", "running", accelerators[0].Status)
	}

	if accelerators[1].Status != "provisioning" {
		t.Errorf("expected Status %q, got %q", "provisioning", accelerators[1].Status)
	}

	if resp.Meta == nil || resp.Meta.Total != 2 {
		t.Errorf("expected Meta.Total to be 2")
	}
}

func TestDedicatedInference_ListAcceleratorsWithOptions(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/dedicated-inferences/di-uuid/accelerators", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)

		if got := r.URL.Query().Get("slug"); got != "gpu-mi300x1-192gb" {
			t.Errorf("expected slug query param %q, got %q", "gpu-mi300x1-192gb", got)
		}
		if got := r.URL.Query().Get("page"); got != "1" {
			t.Errorf("expected page query param %q, got %q", "1", got)
		}
		if got := r.URL.Query().Get("per_page"); got != "20" {
			t.Errorf("expected per_page query param %q, got %q", "20", got)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, diListAcceleratorsJSONResponse)
	})

	opts := &DedicatedInferenceListAcceleratorsOptions{
		Slug: "gpu-mi300x1-192gb",
		ListOptions: ListOptions{
			Page:    1,
			PerPage: 20,
		},
	}

	accelerators, _, err := client.DedicatedInference.ListAccelerators(ctx, "di-uuid", opts)
	if err != nil {
		t.Fatalf("DedicatedInference.ListAccelerators returned error: %v", err)
	}

	if len(accelerators) != 2 {
		t.Fatalf("expected 2 accelerators, got %d", len(accelerators))
	}
}

func TestDedicatedInference_CreateToken(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/dedicated-inferences/di-uuid/tokens", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)

		var req DedicatedInferenceTokenCreateRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}

		if req.Name != "new-inference-token" {
			t.Errorf("expected name %q, got %q", "new-inference-token", req.Name)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, diCreateTokenJSONResponse)
	})

	createReq := &DedicatedInferenceTokenCreateRequest{
		Name: "new-inference-token",
	}

	token, resp, err := client.DedicatedInference.CreateToken(ctx, "di-uuid", createReq)
	if err != nil {
		t.Fatalf("DedicatedInference.CreateToken returned error: %v", err)
	}

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected status code %d, got %d", http.StatusCreated, resp.StatusCode)
	}

	if token.ID != "token-uuid-123" {
		t.Errorf("expected ID %q, got %q", "token-uuid-123", token.ID)
	}

	if token.Name != "new-inference-token" {
		t.Errorf("expected Name %q, got %q", "new-inference-token", token.Name)
	}

	if token.Value != "test-token-value-placeholder" {
		t.Errorf("expected Value %q, got %q", "test-token-value-placeholder", token.Value)
	}
}

func TestDedicatedInference_ListTokens(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/dedicated-inferences/di-uuid/tokens", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, diListTokensJSONResponse)
	})

	tokens, resp, err := client.DedicatedInference.ListTokens(ctx, "di-uuid", nil)
	if err != nil {
		t.Fatalf("DedicatedInference.ListTokens returned error: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	if len(tokens) != 2 {
		t.Fatalf("expected 2 tokens, got %d", len(tokens))
	}

	if tokens[0].ID != "token-uuid-1" {
		t.Errorf("expected ID %q, got %q", "token-uuid-1", tokens[0].ID)
	}

	if tokens[0].Name != "first-token" {
		t.Errorf("expected Name %q, got %q", "first-token", tokens[0].Name)
	}

	if tokens[1].Name != "second-token" {
		t.Errorf("expected Name %q, got %q", "second-token", tokens[1].Name)
	}

	if resp.Meta == nil || resp.Meta.Total != 2 {
		t.Errorf("expected Meta.Total to be 2")
	}
}

func TestDedicatedInference_ListTokensWithPagination(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/dedicated-inferences/di-uuid/tokens", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)

		if got := r.URL.Query().Get("page"); got != "2" {
			t.Errorf("expected page query param %q, got %q", "2", got)
		}
		if got := r.URL.Query().Get("per_page"); got != "10" {
			t.Errorf("expected per_page query param %q, got %q", "10", got)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, diListTokensJSONResponse)
	})

	opts := &ListOptions{
		Page:    2,
		PerPage: 10,
	}

	tokens, _, err := client.DedicatedInference.ListTokens(ctx, "di-uuid", opts)
	if err != nil {
		t.Fatalf("DedicatedInference.ListTokens returned error: %v", err)
	}

	if len(tokens) != 2 {
		t.Fatalf("expected 2 tokens, got %d", len(tokens))
	}
}

func TestDedicatedInference_RevokeToken(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/dedicated-inferences/di-uuid/tokens/token-uuid-123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		w.WriteHeader(http.StatusNoContent)
	})

	resp, err := client.DedicatedInference.RevokeToken(ctx, "di-uuid", "token-uuid-123")
	if err != nil {
		t.Fatalf("DedicatedInference.RevokeToken returned error: %v", err)
	}

	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("expected status code %d, got %d", http.StatusNoContent, resp.StatusCode)
	}
}
