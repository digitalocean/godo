package godo

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

func TestSaasAddonsService_GetAllApps(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v1/marketplace/add-ons/apps/live", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		json.NewEncoder(w).Encode(&saasAddonsAppsRoot{
			Apps: []*SaasAddonsApp{
				{
					ID:          "1",
					Slug:        "test-app-1",
					Name:        "Test App 1",
					Description: "First test application",
					VendorUUID:  "vendor-123",
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				},
				{
					ID:          "2",
					Slug:        "test-app-2",
					Name:        "Test App 2",
					Description: "Second test application",
					VendorUUID:  "vendor-456",
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				},
			},
		})
	})

	apps, _, err := client.SaasAddons.GetAllApps(ctx)
	if err != nil {
		t.Errorf("SaasAddons.GetAllApps returned error: %v", err)
	}

	if len(apps) != 2 {
		t.Errorf("SaasAddons.GetAllApps returned %d apps, expected 2", len(apps))
	}

	if apps[0].ID != "1" {
		t.Errorf("SaasAddons.GetAllApps returned first app ID %s, expected 1", apps[0].ID)
	}
}

func TestSaasAddonsService_GetAppDetails(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v1/marketplace/add-ons/public/apps/test-app", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		json.NewEncoder(w).Encode(&saasAddonsAppDetailsRoot{
			App: &SaasAddonsAppDetails{
				ID:               "1",
				Slug:             "test-app",
				Name:             "Test App",
				Description:      "Test application",
				ConfigVarsPrefix: "TEST_APP_",
			},
		})
	})

	app, _, err := client.SaasAddons.GetAppDetails(ctx, "test-app")
	if err != nil {
		t.Errorf("SaasAddons.GetAppDetails returned error: %v", err)
	}

	if app.ID != "1" {
		t.Errorf("SaasAddons.GetAppDetails returned ID %s, expected 1", app.ID)
	}
}

func TestSaasAddonsService_ListAddons(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v1/marketplace/add-ons/public/resources", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		json.NewEncoder(w).Encode(&saasAddonsPublicResourcesRoot{
			Resources: []*SaasAddonsPublicResource{
				{
					UUID:      "resource-1",
					AppSlug:   "test-app",
					PlanSlug:  "basic-plan",
					State:     "active",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				{
					UUID:      "resource-2",
					AppSlug:   "test-app",
					PlanSlug:  "premium-plan",
					State:     "active",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
			},
		})
	})

	resources, _, err := client.SaasAddons.ListAddons(ctx)
	if err != nil {
		t.Errorf("SaasAddons.ListAddons returned error: %v", err)
	}

	if len(resources) != 2 {
		t.Errorf("SaasAddons.ListAddons returned %d resources, expected 2", len(resources))
	}

	if resources[0].UUID != "resource-1" {
		t.Errorf("SaasAddons.ListAddons returned UUID %s, expected resource-1", resources[0].UUID)
	}
}

func TestSaasAddonsService_GetAddon(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v1/marketplace/add-ons/public/resources/resource-1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		json.NewEncoder(w).Encode(&saasAddonsPublicResourceRoot{
			Resource: &SaasAddonsPublicResource{
				UUID:      "resource-1",
				AppSlug:   "test-app",
				PlanSlug:  "basic-plan",
				State:     "active",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		})
	})

	resource, _, err := client.SaasAddons.GetAddon(ctx, "resource-1")
	if err != nil {
		t.Errorf("SaasAddons.GetAddon returned error: %v", err)
	}

	if resource.UUID != "resource-1" {
		t.Errorf("SaasAddons.GetAddon returned UUID %s, expected resource-1", resource.UUID)
	}
}

func TestSaasAddonsService_CreateAddon(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v1/marketplace/add-ons/public/resources", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)

		var req CreateAddonRequest
		json.NewDecoder(r.Body).Decode(&req)

		if req.AppSlug != "test-app" {
			t.Errorf("CreateAddon request AppSlug = %v, expected test-app", req.AppSlug)
		}
		if req.PlanSlug != "basic-plan" {
			t.Errorf("CreateAddon request PlanSlug = %v, expected basic-plan", req.PlanSlug)
		}

		json.NewEncoder(w).Encode(&saasAddonsPublicResourceRoot{
			Resource: &SaasAddonsPublicResource{
				UUID:      "resource-1",
				AppSlug:   req.AppSlug,
				PlanSlug:  req.PlanSlug,
				State:     "provisioning",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		})
	})

	req := &CreateAddonRequest{
		AppSlug:  "test-app",
		PlanSlug: "basic-plan",
		Name:     "Test Resource 1",
	}

	resource, _, err := client.SaasAddons.CreateAddon(ctx, req)
	if err != nil {
		t.Errorf("SaasAddons.CreateAddon returned error: %v", err)
	}

	if resource.UUID != "resource-1" {
		t.Errorf("SaasAddons.CreateAddon returned UUID %s, expected resource-1", resource.UUID)
	}
}

func TestSaasAddonsService_UpdateAddon(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v1/marketplace/add-ons/public/resources/resource-1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPatch)

		var req UpdateAddonRequest
		json.NewDecoder(r.Body).Decode(&req)

		if req.Name != "Updated Resource Name" {
			t.Errorf("UpdateAddon request Name = %v, expected Updated Resource Name", req.Name)
		}

		json.NewEncoder(w).Encode(&saasAddonsPublicResourceRoot{
			Resource: &SaasAddonsPublicResource{
				UUID:      "resource-1",
				AppSlug:   "test-app",
				PlanSlug:  "basic-plan",
				State:     "active",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		})
	})

	req := &UpdateAddonRequest{
		Name: "Updated Resource Name",
	}

	resource, _, err := client.SaasAddons.UpdateAddon(ctx, "resource-1", req)
	if err != nil {
		t.Errorf("SaasAddons.UpdateAddon returned error: %v", err)
	}

	if resource.UUID != "resource-1" {
		t.Errorf("SaasAddons.UpdateAddon returned UUID %s, expected resource-1", resource.UUID)
	}
}

func TestSaasAddonsService_DeleteAddon(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v1/marketplace/add-ons/public/resources/resource-1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		w.WriteHeader(http.StatusNoContent)
	})

	_, err := client.SaasAddons.DeleteAddon(ctx, "resource-1")
	if err != nil {
		t.Errorf("SaasAddons.DeleteAddon returned error: %v", err)
	}
}

func TestSaasAddonsService_GetAddonMetadata(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v1/marketplace/add-ons/apps/test-app/metadata", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		json.NewEncoder(w).Encode(&SaasAddonsAddonMetadata{
			AppSlug: "test-app",
			Metadata: []*SaasAddonsAddonMetadataItem{
				{
					Name:     "API_KEY",
					DataType: "string",
				},
			},
		})
	})

	metadata, _, err := client.SaasAddons.GetAddonMetadata(ctx, "test-app")
	if err != nil {
		t.Errorf("SaasAddons.GetAddonMetadata returned error: %v", err)
	}

	if metadata.AppSlug != "test-app" {
		t.Errorf("SaasAddons.GetAddonMetadata returned AppSlug %s, expected test-app", metadata.AppSlug)
	}
}

func TestSaasAddonsService_CreateAddonRequestValidation(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v1/marketplace/add-ons/public/resources", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)

		var req CreateAddonRequest
		json.NewDecoder(r.Body).Decode(&req)

		if req.AppSlug == "" {
			http.Error(w, "app_slug is required", http.StatusBadRequest)
			return
		}
		if req.PlanSlug == "" {
			http.Error(w, "plan_slug is required", http.StatusBadRequest)
			return
		}

		json.NewEncoder(w).Encode(&saasAddonsPublicResourceRoot{
			Resource: &SaasAddonsPublicResource{
				UUID:      "resource-1",
				AppSlug:   req.AppSlug,
				PlanSlug:  req.PlanSlug,
				State:     "provisioning",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		})
	})

	// Test with missing AppSlug
	req := &CreateAddonRequest{
		PlanSlug: "basic-plan",
		Name:     "Test Resource 1",
	}

	_, resp, err := client.SaasAddons.CreateAddon(ctx, req)
	if err == nil {
		t.Errorf("SaasAddons.CreateAddon should have returned an error for missing AppSlug")
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("SaasAddons.CreateAddon returned status %d, expected %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestSaasAddonsService_ErrorHandling(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v1/marketplace/add-ons/public/resources/nonexistent", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		http.Error(w, "Resource not found", http.StatusNotFound)
	})

	_, resp, err := client.SaasAddons.GetAddon(ctx, "nonexistent")
	if err == nil {
		t.Errorf("SaasAddons.GetAddon should have returned an error for nonexistent resource")
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("SaasAddons.GetAddon returned status %d, expected %d", resp.StatusCode, http.StatusNotFound)
	}
}
