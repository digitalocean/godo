package godo

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

const saasAddonsBasePath = "v1/marketplace/add-ons"

// SaasAddonsService is an interface for interacting with the SaasAddons/Marketplace Add-ons API
type SaasAddonsService interface {
	GetAppsByVendor(context.Context) ([]*SaasAddonsApp, *Response, error)
	GetAppBySlug(context.Context, string) (*SaasAddonsApp, *Response, error)
	GetAppByVendorUUID(context.Context, string, string) (*SaasAddonsApp, *Response, error)
	GetPlansByApp(context.Context, string) ([]*SaasAddonsPlan, *Response, error)
	GetPublicInfoByApps(context.Context, *GetPublicInfoByAppsRequest) (*GetPublicInfoByAppsResponse, *Response, error)
	GetAppFeatures(context.Context, string) ([]*SaasAddonsFeature, *Response, error)
	GetLiveApps(context.Context) ([]*SaasAddonsApp, *Response, error)
	GetAppBySlugPublic(context.Context, string) (*SaasAddonsPublicApp, *Response, error)
	GetAllResourcesPublic(context.Context) ([]*SaasAddonsPublicResource, *Response, error)
	GetPublicResource(context.Context, string) (*SaasAddonsPublicResource, *Response, error)
	GetAddonMetadata(context.Context, string) (*SaasAddonsAddonMetadata, *Response, error)
	GetDimensionsByFeature(context.Context, string, uint64) ([]*SaasAddonsDimension, *Response, error)
	GetDimensionVolumes(context.Context, string, uint64, uint64) ([]*SaasAddonsDimensionVolume, *Response, error)
	GetPlanFeaturePrices(context.Context, string, uint64) ([]*SaasAddonsPlanFeaturePrice, *Response, error)
}

// SaasAddonsServiceOp handles communication with the SaasAddons/Marketplace Add-ons related methods
type SaasAddonsServiceOp struct {
	client *Client
}

var _ SaasAddonsService = &SaasAddonsServiceOp{}


// SaasAddonsApp represents a SaasAddons application
type SaasAddonsApp struct {
	ID          uint64                `json:"id"`
	Slug        string                `json:"slug"`
	Name        string                `json:"name"`
	Description string                `json:"description"`
	Categories  []*SaasAddonsCategory `json:"categories"`
	State       string                `json:"state"`
	VendorUUID  string                `json:"vendor_uuid"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
}

// SaasAddonsCategory represents a SaasAddons application category
type SaasAddonsCategory struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// SaasAddonsPlan represents a SaasAddons plan
type SaasAddonsPlan struct {
	ID          uint64               `json:"id"`
	Slug        string               `json:"slug"`
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Price       string               `json:"price"`
	Features    []*SaasAddonsFeature `json:"features"`
	AppSlug     string               `json:"app_slug"`
	CreatedAt   time.Time            `json:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at"`
}

// SaasAddonsFeature represents a SaasAddons feature
type SaasAddonsFeature struct {
	ID          uint64    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	DataType    string    `json:"data_type"`
	AppSlug     string    `json:"app_slug"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// SaasAddonsPublicApp represents a public SaasAddons application
type SaasAddonsPublicApp struct {
	ID               uint64                      `json:"id"`
	Slug             string                      `json:"slug"`
	Name             string                      `json:"name"`
	Description      string                      `json:"description"`
	Categories       []*SaasAddonsCategory       `json:"categories"`
	Plans            []*SaasAddonsPublicPlan     `json:"plans"`
	Metadata         []*SaasAddonsPublicMetadata `json:"metadata"`
	ConfigVarsPrefix string                      `json:"config_vars_prefix"`
}

// SaasAddonsPublicPlan represents a public SaasAddons plan
type SaasAddonsPublicPlan struct {
	ID          uint64                         `json:"id"`
	Slug        string                         `json:"slug"`
	Name        string                         `json:"name"`
	Description string                         `json:"description"`
	Price       string                         `json:"price"`
	Features    []*SaasAddonsPublicPlanFeature `json:"features"`
}

// SaasAddonsPublicPlanFeature represents a public SaasAddons plan feature
type SaasAddonsPublicPlanFeature struct {
	ID          uint64 `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	DataType    string `json:"data_type"`
}

// SaasAddonsPublicMetadata represents public SaasAddons metadata
type SaasAddonsPublicMetadata struct {
	Name     string `json:"name"`
	DataType string `json:"data_type"`
}

// SaasAddonsPublicResource represents a public SaasAddons resource
type SaasAddonsPublicResource struct {
	UUID      string                        `json:"uuid"`
	AppSlug   string                        `json:"app_slug"`
	PlanSlug  string                        `json:"plan_slug"`
	State     string                        `json:"state"`
	Metadata  []*SaasAddonsResourceMetadata `json:"metadata"`
	CreatedAt time.Time                     `json:"created_at"`
	UpdatedAt time.Time                     `json:"updated_at"`
}

// SaasAddonsResourceMetadata represents SaasAddons resource metadata
type SaasAddonsResourceMetadata struct {
	Name     string      `json:"name"`
	Value    interface{} `json:"value"`
	DataType string      `json:"data_type"`
}

// SaasAddonsAddonMetadata represents SaasAddons addon metadata
type SaasAddonsAddonMetadata struct {
	AppSlug  string                         `json:"app_slug"`
	Metadata []*SaasAddonsAddonMetadataItem `json:"metadata"`
}

// SaasAddonsAddonMetadataItem represents a SaasAddons addon metadata item
type SaasAddonsAddonMetadataItem struct {
	Name     string `json:"name"`
	DataType string `json:"type"`
}

// SaasAddonsDimension represents a SaasAddons dimension
type SaasAddonsDimension struct {
	ID          uint64                       `json:"id"`
	SKU         string                       `json:"sku"`
	Slug        string                       `json:"slug"`
	DisplayName string                       `json:"display_name"`
	Volumes     []*SaasAddonsDimensionVolume `json:"volumes"`
}

// SaasAddonsDimensionVolume represents a SaasAddons dimension volume
type SaasAddonsDimensionVolume struct {
	ID        uint64 `json:"id"`
	LowVolume uint64 `json:"low_volume"`
	MaxVolume int64  `json:"max_volume"`
}

// SaasAddonsPlanFeaturePrice represents a SaasAddons plan feature price
type SaasAddonsPlanFeaturePrice struct {
	DimensionVolumeID uint64 `json:"dimension_volume_id"`
	PricePerUnit      string `json:"price_per_unit"`
}

// GetPublicInfoByAppsRequest represents the request for getting public info by apps
type GetPublicInfoByAppsRequest struct {
	AppSlugs []string `json:"app_slugs"`
}

// GetPublicInfoByAppsResponse represents the response for getting public info by apps
type GetPublicInfoByAppsResponse struct {
	InfoByApp []*SaasAddonsInfoByApp `json:"info_by_app"`
}

// SaasAddonsInfoByApp represents info by app
type SaasAddonsInfoByApp struct {
	AppSlug string            `json:"app_slug"`
	TOS     string            `json:"tos"`
	EULA    string            `json:"eula"`
	Plans   []*SaasAddonsPlan `json:"plans"`
}

// Root types for API responses
type saasAddonsAppsRoot struct {
	Apps []*SaasAddonsApp `json:"apps"`
}

type saasAddonsAppRoot struct {
	App *SaasAddonsApp `json:"app"`
}

type saasAddonsPlansRoot struct {
	Plans []*SaasAddonsPlan `json:"plans"`
}

type saasAddonsFeaturesRoot struct {
	Features []*SaasAddonsFeature `json:"features"`
}

type saasAddonsPublicAppRoot struct {
	App *SaasAddonsPublicApp `json:"app"`
}

type saasAddonsPublicResourcesRoot struct {
	Resources []*SaasAddonsPublicResource `json:"resources"`
}

type saasAddonsPublicResourceRoot struct {
	Resource *SaasAddonsPublicResource `json:"resource"`
}

type saasAddonsDimensionsRoot struct {
	Dimensions []*SaasAddonsDimension `json:"dimensions"`
}

type saasAddonsDimensionVolumesRoot struct {
	DimensionVolumes []*SaasAddonsDimensionVolume `json:"dimension_volumes"`
}

type saasAddonsPlanFeaturePricesRoot struct {
	PlanFeaturePrices []*SaasAddonsPlanFeaturePrice `json:"plan_feature_prices"`
}

// GetAppsByVendor returns apps by vendor (public, no permissions needed)
func (s *SaasAddonsServiceOp) GetAppsByVendor(ctx context.Context) ([]*SaasAddonsApp, *Response, error) {
	path := fmt.Sprintf("%s/vendors/apps", saasAddonsBasePath)

	req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	root := new(saasAddonsAppsRoot)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}

	return root.Apps, resp, nil
}

// GetAppBySlug returns an app by slug (public, no permissions needed)
func (s *SaasAddonsServiceOp) GetAppBySlug(ctx context.Context, appSlug string) (*SaasAddonsApp, *Response, error) {
	path := fmt.Sprintf("%s/apps/%s", saasAddonsBasePath, appSlug)

	req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	root := new(saasAddonsAppRoot)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}

	return root.App, resp, nil
}

// GetAppByVendorUUID returns an app by vendor UUID and app slug (public, no permissions needed)
func (s *SaasAddonsServiceOp) GetAppByVendorUUID(ctx context.Context, vendorUUID, appSlug string) (*SaasAddonsApp, *Response, error) {
	path := fmt.Sprintf("%s/vendor/%s/apps/%s", saasAddonsBasePath, vendorUUID, appSlug)

	req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	root := new(saasAddonsAppRoot)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}

	return root.App, resp, nil
}

// GetPlansByApp returns plans for an app (public, no permissions needed)
func (s *SaasAddonsServiceOp) GetPlansByApp(ctx context.Context, appSlug string) ([]*SaasAddonsPlan, *Response, error) {
	path := fmt.Sprintf("%s/apps/%s/plans", saasAddonsBasePath, appSlug)

	req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	root := new(saasAddonsPlansRoot)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}

	return root.Plans, resp, nil
}

// GetPublicInfoByApps returns public info for multiple apps (public, no permissions needed)
func (s *SaasAddonsServiceOp) GetPublicInfoByApps(ctx context.Context, request *GetPublicInfoByAppsRequest) (*GetPublicInfoByAppsResponse, *Response, error) {
	path := fmt.Sprintf("%s/apps/public_info", saasAddonsBasePath)

	req, err := s.client.NewRequest(ctx, http.MethodPost, path, request)
	if err != nil {
		return nil, nil, err
	}

	root := new(GetPublicInfoByAppsResponse)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}

	return root, resp, nil
}

// GetAppFeatures returns features for an app (public, no permissions needed)
func (s *SaasAddonsServiceOp) GetAppFeatures(ctx context.Context, appSlug string) ([]*SaasAddonsFeature, *Response, error) {
	path := fmt.Sprintf("%s/apps/%s/features", saasAddonsBasePath, appSlug)

	req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	root := new(saasAddonsFeaturesRoot)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}

	return root.Features, resp, nil
}

// GetLiveApps returns live apps (public, no permissions needed)
func (s *SaasAddonsServiceOp) GetLiveApps(ctx context.Context) ([]*SaasAddonsApp, *Response, error) {
	path := fmt.Sprintf("%s/apps", saasAddonsBasePath)

	req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	root := new(saasAddonsAppsRoot)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}

	return root.Apps, resp, nil
}

// GetAppBySlugPublic returns an app by slug using the public endpoint
func (s *SaasAddonsServiceOp) GetAppBySlugPublic(ctx context.Context, appSlug string) (*SaasAddonsPublicApp, *Response, error) {
	path := fmt.Sprintf("%s/public/apps/%s", saasAddonsBasePath, appSlug)

	req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	root := new(saasAddonsPublicAppRoot)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}

	return root.App, resp, nil
}

// GetAllResourcesPublic returns all public resources
func (s *SaasAddonsServiceOp) GetAllResourcesPublic(ctx context.Context) ([]*SaasAddonsPublicResource, *Response, error) {
	path := fmt.Sprintf("%s/public/resources", saasAddonsBasePath)

	req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	root := new(saasAddonsPublicResourcesRoot)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}

	return root.Resources, resp, nil
}

// GetPublicResource returns a public resource by UUID
func (s *SaasAddonsServiceOp) GetPublicResource(ctx context.Context, resourceUUID string) (*SaasAddonsPublicResource, *Response, error) {
	path := fmt.Sprintf("%s/public/resources/%s", saasAddonsBasePath, resourceUUID)

	req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	root := new(saasAddonsPublicResourceRoot)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}

	return root.Resource, resp, nil
}

// GetAddonMetadata returns addon metadata for an app
func (s *SaasAddonsServiceOp) GetAddonMetadata(ctx context.Context, appSlug string) (*SaasAddonsAddonMetadata, *Response, error) {
	path := fmt.Sprintf("%s/apps/%s/metadata", saasAddonsBasePath, appSlug)

	req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	root := new(SaasAddonsAddonMetadata)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}

	return root, resp, nil
}

// GetDimensionsByFeature returns dimensions for a feature
func (s *SaasAddonsServiceOp) GetDimensionsByFeature(ctx context.Context, appSlug string, featureID uint64) ([]*SaasAddonsDimension, *Response, error) {
	path := fmt.Sprintf("%s/apps/%s/features/%d/dimensions", saasAddonsBasePath, appSlug, featureID)

	req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	root := new(saasAddonsDimensionsRoot)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}

	return root.Dimensions, resp, nil
}

// GetDimensionVolumes returns dimension volumes for a dimension
func (s *SaasAddonsServiceOp) GetDimensionVolumes(ctx context.Context, appSlug string, featureID, dimensionID uint64) ([]*SaasAddonsDimensionVolume, *Response, error) {
	path := fmt.Sprintf("%s/apps/%s/features/%d/dimensions/%d/volumes", saasAddonsBasePath, appSlug, featureID, dimensionID)

	req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	root := new(saasAddonsDimensionVolumesRoot)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}

	return root.DimensionVolumes, resp, nil
}

// GetPlanFeaturePrices returns plan feature prices
func (s *SaasAddonsServiceOp) GetPlanFeaturePrices(ctx context.Context, appSlug string, planFeatureID uint64) ([]*SaasAddonsPlanFeaturePrice, *Response, error) {
	path := fmt.Sprintf("%s/apps/%s/plan_features/%d/prices", saasAddonsBasePath, appSlug, planFeatureID)

	req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	root := new(saasAddonsPlanFeaturePricesRoot)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}

	return root.PlanFeaturePrices, resp, nil
}
