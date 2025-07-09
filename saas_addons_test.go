package godo

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSaasAddonsService_GetAppsByVendor(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v1/marketplace/add-ons/vendors/apps", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{"apps": [{"id": 1, "slug": "test-app", "name": "Test App", "description": "A test application"}]}`)
	})

	apps, _, err := client.SaasAddons.GetAppsByVendor(context.Background())
	require.NoError(t, err)
	require.Len(t, apps, 1)

	expected := &SaasAddonsApp{
		ID:          1,
		Slug:        "test-app",
		Name:        "Test App",
		Description: "A test application",
	}
	assert.Equal(t, expected, apps[0])
}

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

func TestSaasAddonsService_GetAppByVendorUUID(t *testing.T) {
	setup()
	defer teardown()

	vendorUUID := "vendor-123"
	appSlug := "test-app"

	mux.HandleFunc("/v1/marketplace/add-ons/vendor/vendor-123/apps/test-app", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{"app": {"id": 1, "slug": "test-app", "name": "Test App", "vendor_uuid": "vendor-123"}}`)
	})

	app, _, err := client.SaasAddons.GetAppByVendorUUID(context.Background(), vendorUUID, appSlug)
	require.NoError(t, err)

	expected := &SaasAddonsApp{
		ID:         1,
		Slug:       "test-app",
		Name:       "Test App",
		VendorUUID: "vendor-123",
	}
	assert.Equal(t, expected, app)
}

func TestSaasAddonsService_GetPlansByApp(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v1/marketplace/add-ons/apps/test-app/plans", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{"plans": [{"id": 1, "slug": "basic", "name": "Basic Plan", "price": "5.00", "app_slug": "test-app"}]}`)
	})

	plans, _, err := client.SaasAddons.GetPlansByApp(context.Background(), "test-app")
	require.NoError(t, err)
	require.Len(t, plans, 1)

	expected := &SaasAddonsPlan{
		ID:      1,
		Slug:    "basic",
		Name:    "Basic Plan",
		Price:   "5.00",
		AppSlug: "test-app",
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
	require.Len(t, response.InfoByApp, 1)

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

func TestSaasAddonsService_GetLiveApps(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v1/marketplace/add-ons/apps", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{"apps": [{"id": 1, "slug": "live-app", "name": "Live App", "state": "live"}]}`)
	})

	apps, _, err := client.SaasAddons.GetLiveApps(context.Background())
	require.NoError(t, err)
	require.Len(t, apps, 1)

	expected := &SaasAddonsApp{
		ID:    1,
		Slug:  "live-app",
		Name:  "Live App",
		State: "live",
	}
	assert.Equal(t, expected, apps[0])
}

func TestSaasAddonsService_GetAppBySlugPublic(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v1/marketplace/add-ons/public/apps/test-app", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{"app": {"id": 1, "slug": "test-app", "name": "Test App", "config_vars_prefix": "TEST_"}}`)
	})

	app, _, err := client.SaasAddons.GetAppBySlugPublic(context.Background(), "test-app")
	require.NoError(t, err)

	expected := &SaasAddonsPublicApp{
		ID:               1,
		Slug:             "test-app",
		Name:             "Test App",
		ConfigVarsPrefix: "TEST_",
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

func TestSaasAddonsService_GetDimensionsByFeature(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v1/marketplace/add-ons/apps/test-app/features/1/dimensions", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{"dimensions": [{"id": 1, "sku": "SKU123", "slug": "storage", "display_name": "Storage"}]}`)
	})

	dimensions, _, err := client.SaasAddons.GetDimensionsByFeature(context.Background(), "test-app", 1)
	require.NoError(t, err)
	require.Len(t, dimensions, 1)

	expected := &SaasAddonsDimension{
		ID:          1,
		SKU:         "SKU123",
		Slug:        "storage",
		DisplayName: "Storage",
	}
	assert.Equal(t, expected, dimensions[0])
}

func TestSaasAddonsService_GetDimensionVolumes(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v1/marketplace/add-ons/apps/test-app/features/1/dimensions/1/volumes", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{"dimension_volumes": [{"id": 1, "low_volume": 1, "max_volume": 100}]}`)
	})

	volumes, _, err := client.SaasAddons.GetDimensionVolumes(context.Background(), "test-app", 1, 1)
	require.NoError(t, err)
	require.Len(t, volumes, 1)

	expected := &SaasAddonsDimensionVolume{
		ID:        1,
		LowVolume: 1,
		MaxVolume: 100,
	}
	assert.Equal(t, expected, volumes[0])
}

func TestSaasAddonsService_GetPlanFeaturePrices(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v1/marketplace/add-ons/apps/test-app/plan_features/1/prices", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{"plan_feature_prices": [{"dimension_volume_id": 1, "price_per_unit": "0.10"}]}`)
	})

	prices, _, err := client.SaasAddons.GetPlanFeaturePrices(context.Background(), "test-app", 1)
	require.NoError(t, err)
	require.Len(t, prices, 1)

	expected := &SaasAddonsPlanFeaturePrice{
		DimensionVolumeID: 1,
		PricePerUnit:      "0.10",
	}
	assert.Equal(t, expected, prices[0])
}
