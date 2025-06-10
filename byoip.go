package godo

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

const byoipsBasePath = "/v2/byoip_prefixes"

// BYOIPsService is an interface for interacting with the BYOIPs
// endpoints of the Digital Ocean API.

type BYOIPsService interface {
	Create(context.Context, *BYOIPCreateReq) (*BYOIPPrefixCreateResp, *Response, error)
	List(context.Context, *ListOptions) ([]BYOIP, *Response, error)
	Get(context.Context, string) (*BYOIP, *Response, error)
	GetResources(context.Context, string) ([]BYOIPResource, *Response, error)
}

// BYOIPServiceOp handles communication with the BYOIP related methods of the
// DigitalOcean API.
type BYOIPServiceOp struct {
	client *Client
}

var _ BYOIPsService = (*BYOIPServiceOp)(nil)

// BYOIP represents a Digital Ocean BYOIP resource.
type BYOIP struct {
	UUID          string                   `json:"uuid"`
	Cidr          string                   `json:"cidr"`
	RegionSlug    string                   `json:"region"`
	Status        string                   `json:"status"`
	FailureReason string                   `json:"failure_reason"`
	Validations   []map[string]interface{} `json:"validations"`
}

// BYOIPCreateReq represents a request to create a BYOIP prefix.
type BYOIPCreateReq struct {
	Prefix    string `json:"prefix"`
	Signature string `json:"signature"`
	Region    string `json:"region"`
}

// BYOIPPrefixCreateResp represents the response from creating a BYOIP prefix.
type BYOIPPrefixCreateResp struct {
	ID string `json:"id"`
}

// BYOIPResource represents a BYOIP resource allocations
type BYOIPResource struct {
	ID         uint64    `json:"id"`
	BYOIP      string    `json:"byoip"`
	Resource   string    `json:"resource"`
	RegionSlug string    `json:"region"`
	AssignedAt time.Time `json:"assigned_at"`
}

type byoipRoot struct {
	BYOIP *BYOIP `json:"byoip"`
}

type byoipsRoot struct {
	BYOIPs []BYOIP `json:"byoips"`
	Links  *Links  `json:"links"`
	Meta   *Meta   `json:"meta"`
}

type byoipResourcesRoot struct {
	Resources []BYOIPResource `json:"ips"`
}

func (r BYOIP) String() string {
	return Stringify(r)
}

// List all BYOIP prefixes.
func (r *BYOIPServiceOp) List(ctx context.Context, opt *ListOptions) ([]BYOIP, *Response, error) {
	path := byoipsBasePath
	path, err := addOptions(path, opt)
	if err != nil {
		return nil, nil, err
	}

	req, err := r.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	root := new(byoipsRoot)
	resp, err := r.client.Do(ctx, req, root)
	if err != nil {
		return nil, nil, err
	}
	if root.Meta != nil {
		resp.Meta = root.Meta
	}
	if root.Links != nil {
		resp.Links = root.Links
	}

	return root.BYOIPs, resp, err
}

// Get an individual BYOIP prefix details.
func (r *BYOIPServiceOp) Get(ctx context.Context, uuid string) (*BYOIP, *Response, error) {
	path := fmt.Sprintf("%s/%s", byoipsBasePath, uuid)

	req, err := r.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	root := new(byoipRoot)
	resp, err := r.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}

	return root.BYOIP, resp, err
}

// GetResources return all existing BYOIP allocations for given BYOIP prefix id.
func (r *BYOIPServiceOp) GetResources(ctx context.Context, uuid string) ([]BYOIPResource, *Response, error) {
	path := fmt.Sprintf("%s/%s/ips", byoipsBasePath, uuid)

	req, err := r.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	root := new(byoipResourcesRoot)

	resp, err := r.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}

	return root.Resources, resp, err
}

// Create a BYOIP prefix
func (r *BYOIPServiceOp) Create(ctx context.Context, byoip *BYOIPCreateReq) (*BYOIPPrefixCreateResp, *Response, error) {
	path := byoipsBasePath

	req, err := r.client.NewRequest(ctx, http.MethodPost, path, byoip)
	if err != nil {
		return nil, nil, err
	}

	root := new(BYOIPPrefixCreateResp)
	resp, err := r.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}

	return root, resp, err
}
