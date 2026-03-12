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
			HuggingFaceToken: "hf_test-token",
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

var diGetSizesJSONResponse = `
{
  "enabled_regions": ["atl1", "nyc2"],
  "sizes": [
    {
      "gpu_slug": "gpu-mi300x1-192gb",
      "price_per_hour": "1.99",
      "regions": ["atl1", "nyc2"],
      "currency": "USD",
      "cpu": 20,
      "memory": 128000,
      "gpu": {
        "count": 1,
        "vram_gb": 192,
        "slug": "gpu-mi300x1-192gb"
      },
      "size_category": {
        "name": "AMD MI300X",
        "fleet_name": "do:compute-fleet:gpu-amd-mi300x"
      },
      "disks": [
        {
          "type": "Local",
          "size_gb": 720
        },
        {
          "type": "Scratch",
          "size_gb": 5120
        }
      ]
    },
    {
      "gpu_slug": "gpu-h100x1-80gb",
      "price_per_hour": "3.39",
      "regions": ["nyc2", "tor1"],
      "currency": "USD",
      "cpu": 26,
      "memory": 200000,
      "gpu": {
        "count": 1,
        "vram_gb": 80,
        "slug": "gpu-h100x1-80gb"
      },
      "size_category": {
        "name": "NVIDIA H100",
        "fleet_name": "do:compute-fleet:gpu-nvidia-h100"
      },
      "disks": [
        {
          "type": "Local",
          "size_gb": 480
        }
      ]
    }
  ]
}
`

func TestDedicatedInference_GetSizes(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/dedicated-inferences/sizes", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, diGetSizesJSONResponse)
	})

	sizes, resp, err := client.DedicatedInference.GetSizes(ctx)
	if err != nil {
		t.Fatalf("DedicatedInference.GetSizes returned error: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	if len(sizes.EnabledRegions) != 2 {
		t.Fatalf("expected 2 enabled regions, got %d", len(sizes.EnabledRegions))
	}

	if sizes.EnabledRegions[0] != "atl1" {
		t.Errorf("expected first enabled region %q, got %q", "atl1", sizes.EnabledRegions[0])
	}

	if len(sizes.Sizes) != 2 {
		t.Fatalf("expected 2 sizes, got %d", len(sizes.Sizes))
	}

	size := sizes.Sizes[0]
	if size.GPUSlug != "gpu-mi300x1-192gb" {
		t.Errorf("expected GPUSlug %q, got %q", "gpu-mi300x1-192gb", size.GPUSlug)
	}

	if size.PricePerHour != "1.99" {
		t.Errorf("expected PricePerHour %q, got %q", "1.99", size.PricePerHour)
	}

	if size.Currency != "USD" {
		t.Errorf("expected Currency %q, got %q", "USD", size.Currency)
	}

	if size.CPU != 20 {
		t.Errorf("expected CPU %d, got %d", 20, size.CPU)
	}

	if size.Memory != 128000 {
		t.Errorf("expected Memory %d, got %d", 128000, size.Memory)
	}

	if size.GPU == nil {
		t.Fatal("expected GPU to be non-nil")
	}

	if size.GPU.Count != 1 {
		t.Errorf("expected GPU Count %d, got %d", 1, size.GPU.Count)
	}

	if size.GPU.VramGb != 192 {
		t.Errorf("expected GPU VramGb %d, got %d", 192, size.GPU.VramGb)
	}

	if size.SizeCategory == nil {
		t.Fatal("expected SizeCategory to be non-nil")
	}

	if size.SizeCategory.Name != "AMD MI300X" {
		t.Errorf("expected SizeCategory Name %q, got %q", "AMD MI300X", size.SizeCategory.Name)
	}

	if len(size.Disks) != 2 {
		t.Fatalf("expected 2 disks, got %d", len(size.Disks))
	}

	if size.Disks[0].Type != "Local" {
		t.Errorf("expected first disk type %q, got %q", "Local", size.Disks[0].Type)
	}

	if size.Disks[0].SizeGb != 720 {
		t.Errorf("expected first disk size %d, got %d", 720, size.Disks[0].SizeGb)
	}
}
