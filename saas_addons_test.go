package godo

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSaasAddonsService_GetAppBySlug(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v1/marketplace/add-ons/apps/test-app", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{"app": {"id": 1, "slug": "test-app", "name": "Test App", "description": "A test application"}}`)
	})

	app, _, err := client.SaasAddons.GetAppBySlug(context.Background(), "test-app")
	require.NoError(t, err)

	expected := &SaasAddonsApp{
		ID:          1,
		Slug:        "test-app",
		Name:        "Test App",
		Description: "A test application",
	}
	assert.Equal(t, expected, app)
}

func TestSaasAddonsService_GetPlansByApp(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v1/marketplace/add-ons/apps/test-app/plans", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{"plans": [{"id": 1, "slug": "basic", "name": "Basic Plan", "description": "Basic plan for testing"}]}`)
	})

	plans, _, err := client.SaasAddons.GetPlansByApp(context.Background(), "test-app")
	require.NoError(t, err)
	require.Len(t, plans, 1)

	expected := &SaasAddonsPlan{
		ID:          1,
		Slug:        "basic",
		Name:        "Basic Plan",
		Description: "Basic plan for testing",
	}
	assert.Equal(t, expected, plans[0])
}

func TestSaasAddonsService_GetPublicInfoByApps(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v1/marketplace/add-ons/apps/public_info", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `{"info_by_app": [{"app_slug": "test-app", "tos": "Terms of Service", "eula": "End User License Agreement"}]}`)
	})

	request := &GetPublicInfoByAppsRequest{
		AppSlugs: []string{"test-app"},
	}

	response, _, err := client.SaasAddons.GetPublicInfoByApps(context.Background(), request)
	require.NoError(t, err)

	expected := &SaasAddonsInfoByApp{
		AppSlug: "test-app",
		TOS:     "Terms of Service",
		EULA:    "End User License Agreement",
	}
	assert.Equal(t, expected, response.InfoByApp[0])
}

func TestSaasAddonsService_GetAppFeatures(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v1/marketplace/add-ons/apps/test-app/features", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{"features": [{"id": 1, "name": "Storage", "description": "Storage space", "data_type": "string", "app_slug": "test-app"}]}`)
	})

	features, _, err := client.SaasAddons.GetAppFeatures(context.Background(), "test-app")
	require.NoError(t, err)
	require.Len(t, features, 1)

	expected := &SaasAddonsFeature{
		ID:          1,
		Name:        "Storage",
		Description: "Storage space",
		DataType:    "string",
		AppSlug:     "test-app",
	}
	assert.Equal(t, expected, features[0])
}

func TestSaasAddonsService_GetAllApps(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v1/marketplace/add-ons/apps/live", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{"apps": [{"id": 1, "slug": "test-app", "name": "Test App", "description": "A live application"}]}`)
	})

	apps, _, err := client.SaasAddons.GetAllApps(context.Background())
	require.NoError(t, err)
	require.Len(t, apps, 1)

	expected := &SaasAddonsApp{
		ID:          1,
		Slug:        "test-app",
		Name:        "Test App",
		Description: "A live application",
	}
	assert.Equal(t, expected, apps[0])
}

func TestSaasAddonsService_GetAppBySlugPublic(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v1/marketplace/add-ons/public/apps/test-app", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{"app": {"id": 1, "slug": "test-app", "name": "Test App", "description": "A public application"}}`)
	})

	app, _, err := client.SaasAddons.GetAppBySlugPublic(context.Background(), "test-app")
	require.NoError(t, err)

	expected := &SaasAddonsPublicApp{
		ID:          1,
		Slug:        "test-app",
		Name:        "Test App",
		Description: "A public application",
	}
	assert.Equal(t, expected, app)
}

func TestSaasAddonsService_GetAllResourcesPublic(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v1/marketplace/add-ons/public/resources", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{"resources": [{"uuid": "resource-123", "app_slug": "test-app", "plan_slug": "basic", "state": "provisioned"}]}`)
	})

	resources, _, err := client.SaasAddons.GetAllResourcesPublic(context.Background())
	require.NoError(t, err)
	require.Len(t, resources, 1)

	expected := &SaasAddonsPublicResource{
		UUID:     "resource-123",
		AppSlug:  "test-app",
		PlanSlug: "basic",
		State:    "provisioned",
	}
	assert.Equal(t, expected, resources[0])
}

func TestSaasAddonsService_GetPublicResource(t *testing.T) {
	setup()
	defer teardown()

	resourceUUID := "resource-123"
	mux.HandleFunc("/v1/marketplace/add-ons/public/resources/resource-123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{"resource": {"uuid": "resource-123", "app_slug": "test-app", "plan_slug": "basic", "state": "provisioned"}}`)
	})

	resource, _, err := client.SaasAddons.GetPublicResource(context.Background(), resourceUUID)
	require.NoError(t, err)

	expected := &SaasAddonsPublicResource{
		UUID:     "resource-123",
		AppSlug:  "test-app",
		PlanSlug: "basic",
		State:    "provisioned",
	}
	assert.Equal(t, expected, resource)
}

func TestSaasAddonsService_CreateResourcePublic(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v1/marketplace/add-ons/public/resources", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `{"resource": {"uuid": "resource-123", "app_slug": "test-app", "plan_slug": "basic", "state": "provisioned"}}`)
	})

	request := &CreateResourceRequest{
		AppSlug:  "test-app",
		PlanSlug: "basic",
		Name:     "test-resource",
	}

	resource, _, err := client.SaasAddons.CreateResourcePublic(context.Background(), request)
	require.NoError(t, err)

	expected := &SaasAddonsPublicResource{
		UUID:     "resource-123",
		AppSlug:  "test-app",
		PlanSlug: "basic",
		State:    "provisioned",
	}
	assert.Equal(t, expected, resource)
}

func TestSaasAddonsService_UpdateResourcePublic(t *testing.T) {
	setup()
	defer teardown()

	resourceUUID := "resource-123"
	mux.HandleFunc("/v1/marketplace/add-ons/public/resources/resource-123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPatch)
		fmt.Fprint(w, `{"resource": {"uuid": "resource-123", "app_slug": "test-app", "plan_slug": "basic", "state": "provisioned"}}`)
	})

	request := &UpdateResourceRequest{
		Name: "updated-resource",
	}

	resource, _, err := client.SaasAddons.UpdateResourcePublic(context.Background(), resourceUUID, request)
	require.NoError(t, err)

	expected := &SaasAddonsPublicResource{
		UUID:     "resource-123",
		AppSlug:  "test-app",
		PlanSlug: "basic",
		State:    "provisioned",
	}
	assert.Equal(t, expected, resource)
}

func TestSaasAddonsService_DeprovisionResourcePublic(t *testing.T) {
	setup()
	defer teardown()

	resourceUUID := "resource-123"
	mux.HandleFunc("/v1/marketplace/add-ons/public/resources/resource-123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		w.WriteHeader(http.StatusNoContent)
	})

	_, err := client.SaasAddons.DeprovisionResourcePublic(context.Background(), resourceUUID)
	require.NoError(t, err)
}

func TestSaasAddonsService_GetAddonMetadata(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v1/marketplace/add-ons/apps/test-app/metadata", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{"app_slug": "test-app", "metadata": [{"name": "database_url", "type": "string"}]}`)
	})

	metadata, _, err := client.SaasAddons.GetAddonMetadata(context.Background(), "test-app")
	require.NoError(t, err)

	expected := &SaasAddonsAddonMetadata{
		AppSlug: "test-app",
		Metadata: []*SaasAddonsAddonMetadataItem{
			{
				Name:     "database_url",
				DataType: "string",
			},
		},
	}
	assert.Equal(t, expected, metadata)
}
