package godo

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

const saasAddonsBasePath = "/api/v1/marketplace/add-ons"

// SaasAddonsService is an interface for interacting with the SaasAddons/Marketplace Add-ons API
type SaasAddonsService interface {
	GetAllApps(context.Context) ([]*SaasAddonsApp, *Response, error)
	GetAppDetails(context.Context, string) (*SaasAddonsAppDetails, *Response, error)
	ListAddons(context.Context) ([]*SaasAddonsPublicResource, *Response, error)
	GetAddon(context.Context, string) (*SaasAddonsPublicResource, *Response, error)
	CreateAddon(context.Context, *CreateAddonRequest) (*SaasAddonsPublicResource, *Response, error)
	UpdateAddon(context.Context, string, *UpdateAddonRequest) (*SaasAddonsPublicResource, *Response, error)
	DeleteAddon(context.Context, string) (*Response, error)
	GetAddonMetadata(context.Context, string) (*SaasAddonsAddonMetadata, *Response, error)
}

// SaasAddonsServiceOp handles communication with the SaasAddons/Marketplace Add-ons related methods
type SaasAddonsServiceOp struct {
	client *Client
}

var _ SaasAddonsService = &SaasAddonsServiceOp{}

// SaasAddonsApp represents a SaasAddons application
type SaasAddonsApp struct {
	ID          string                `json:"id"`
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
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// SaasAddonsPlan represents a SaasAddons plan
type SaasAddonsPlan struct {
	ID          string               `json:"id"`
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
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	DataType    string    `json:"data_type"`
	AppSlug     string    `json:"app_slug"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// SaasAddonsAppDetails represents detailed SaasAddons application information
type SaasAddonsAppDetails struct {
	ID               string                    `json:"id"`
	Slug             string                    `json:"slug"`
	Name             string                    `json:"name"`
	Description      string                    `json:"description"`
	Categories       []*SaasAddonsCategory     `json:"categories"`
	Plans            []*SaasAddonsDetailedPlan `json:"plans"`
	Metadata         []*SaasAddonsMetadata     `json:"metadata"`
	ConfigVarsPrefix string                    `json:"config_vars_prefix"`
}

// SaasAddonsDetailedPlan represents a detailed SaasAddons plan
type SaasAddonsDetailedPlan struct {
	ID          string                           `json:"id"`
	Slug        string                           `json:"slug"`
	Name        string                           `json:"name"`
	Description string                           `json:"description"`
	Price       string                           `json:"price"`
	Features    []*SaasAddonsDetailedPlanFeature `json:"features"`
}

// SaasAddonsDetailedPlanFeature represents a detailed SaasAddons plan feature
type SaasAddonsDetailedPlanFeature struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	DataType    string `json:"data_type"`
}

// SaasAddonsMetadata represents SaasAddons metadata
type SaasAddonsMetadata struct {
	Name     string `json:"name"`
	DataType string `json:"data_type"`
}

// SaasAddonsPublicApp represents a public SaasAddons application
type SaasAddonsPublicApp struct {
	ID               string                      `json:"id"`
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
	ID          string                         `json:"id"`
	Slug        string                         `json:"slug"`
	Name        string                         `json:"name"`
	Description string                         `json:"description"`
	Price       string                         `json:"price"`
	Features    []*SaasAddonsPublicPlanFeature `json:"features"`
}

// SaasAddonsPublicPlanFeature represents a public SaasAddons plan feature
type SaasAddonsPublicPlanFeature struct {
	ID          string `json:"id"`
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

// GetAppsInfoRequest represents the request for getting apps info
type GetAppsInfoRequest struct {
	AppSlugs []string `json:"app_slugs"`
}

// GetAppsInfoResponse represents the response for getting apps info
type GetAppsInfoResponse struct {
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

type saasAddonsAppDetailsRoot struct {
	App *SaasAddonsAppDetails `json:"app"`
}

type saasAddonsPublicResourcesRoot struct {
	Resources []*SaasAddonsPublicResource `json:"resources"`
}

type saasAddonsPublicResourceRoot struct {
	Resource *SaasAddonsPublicResource `json:"resource"`
}

// CreateAddonRequest represents the request for creating an addon
type CreateAddonRequest struct {
	AppSlug         string                        `json:"app_slug"`
	PlanSlug        string                        `json:"plan_slug"`
	Name            string                        `json:"name"`
	Metadata        []*SaasAddonsResourceMetadata `json:"metadata,omitempty"`
	LinkedDropletID uint64                        `json:"linked_droplet_id,omitempty"`
	FleetUUID       string                        `json:"fleet_uuid,omitempty"`
}

// UpdateAddonRequest represents the request for updating an addon
type UpdateAddonRequest struct {
	Name string `json:"name"`
}

// GetAppDetails returns detailed app information
func (s *SaasAddonsServiceOp) GetAppDetails(ctx context.Context, appSlug string) (*SaasAddonsAppDetails, *Response, error) {
	path := fmt.Sprintf("%s/public/apps/%s", saasAddonsBasePath, appSlug)

	req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	root := new(saasAddonsAppDetailsRoot)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}

	return root.App, resp, nil
}

// ListAddons returns all addons
func (s *SaasAddonsServiceOp) ListAddons(ctx context.Context) ([]*SaasAddonsPublicResource, *Response, error) {
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

// GetAddon returns an addon by UUID
func (s *SaasAddonsServiceOp) GetAddon(ctx context.Context, resourceUUID string) (*SaasAddonsPublicResource, *Response, error) {
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

// CreateAddon creates an addon
func (s *SaasAddonsServiceOp) CreateAddon(ctx context.Context, request *CreateAddonRequest) (*SaasAddonsPublicResource, *Response, error) {
	path := fmt.Sprintf("%s/public/resources", saasAddonsBasePath)

	req, err := s.client.NewRequest(ctx, http.MethodPost, path, request)
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

// UpdateAddon updates an addon
func (s *SaasAddonsServiceOp) UpdateAddon(ctx context.Context, resourceUUID string, request *UpdateAddonRequest) (*SaasAddonsPublicResource, *Response, error) {
	path := fmt.Sprintf("%s/public/resources/%s", saasAddonsBasePath, resourceUUID)

	req, err := s.client.NewRequest(ctx, http.MethodPatch, path, request)
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

// DeleteAddon deletes an addon
func (s *SaasAddonsServiceOp) DeleteAddon(ctx context.Context, resourceUUID string) (*Response, error) {
	path := fmt.Sprintf("%s/public/resources/%s", saasAddonsBasePath, resourceUUID)

	req, err := s.client.NewRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(ctx, req, nil)
	if err != nil {
		return resp, err
	}

	return resp, nil
}

// GetAllApps returns all live apps (public, no permissions needed)
func (s *SaasAddonsServiceOp) GetAllApps(ctx context.Context) ([]*SaasAddonsApp, *Response, error) {
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
