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
	List(context.Context, *ListOptions) ([]*BYOIP, *Response, error)
	Get(context.Context, string) (*BYOIP, *Response, error)
	GetResources(context.Context, string, *ListOptions) ([]BYOIPResource, *Response, error)
	Delete(context.Context, string) (*Response, error)
}

// BYOIPServiceOp handles communication with the BYOIP related methods of the
// DigitalOcean API.
type BYOIPServiceOp struct {
	client *Client
}

var _ BYOIPsService = (*BYOIPServiceOp)(nil)

type BYOIP struct {
	Prefix        string `json:"prefix"`
	Status        string `json:"status"`
	UUID          string `json:"uuid"`
	Region        string `json:"region"`
	Validations   []any  `json:"validations"`
	FailureReason string `json:"failure_reason"`
}

// BYOIPCreateReq represents a request to create a BYOIP prefix.
type BYOIPCreateReq struct {
	Prefix    string `json:"prefix"`
	Signature string `json:"signature"`
	Region    string `json:"region"`
}

// BYOIPPrefixCreateResp represents the response from creating a BYOIP prefix.
type BYOIPPrefixCreateResp struct {
	UUID   string `json:"uuid"`
	Region string `json:"region"`
	Status string `json:"status"`
}

// BYOIPResource represents a BYOIP resource allocations
type BYOIPResource struct {
	ID         uint64    `json:"id"`
	BYOIP      string    `json:"byoip"`
	Resource   string    `json:"resource"`
	Region     string    `json:"region"`
	AssignedAt time.Time `json:"assigned_at"`
}

type byoipRoot struct {
	BYOIP *BYOIP `json:"byoip_prefix"`
}

type byoipsRoot struct {
	BYOIPs []*BYOIP `json:"byoip_prefixes"`
	Links  *Links   `json:"links"`
	Meta   *Meta    `json:"meta"`
}

type byoipResourcesRoot struct {
	Resources []BYOIPResource `json:"ips"`
	Links     *Links          `json:"links"`
	Meta      *Meta           `json:"meta"`
}

func (r BYOIP) String() string {
	return Stringify(r)
}

// List all BYOIP prefixes.
func (r *BYOIPServiceOp) List(ctx context.Context, opt *ListOptions) ([]*BYOIP, *Response, error) {
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
func (r *BYOIPServiceOp) GetResources(ctx context.Context, uuid string, opt *ListOptions) ([]BYOIPResource, *Response, error) {
	path := fmt.Sprintf("%s/%s/ips", byoipsBasePath, uuid)

	addOptions(path, opt)
	req, err := r.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	root := new(byoipResourcesRoot)

	resp, err := r.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}
	if root.Meta != nil {
		resp.Meta = root.Meta
	}
	if root.Links != nil {
		resp.Links = root.Links
	}
	return root.Resources, resp, err
}

// Create a BYOIP prefix
func (r *BYOIPServiceOp) Create(ctx context.Context, byoip *BYOIPCreateReq) (*BYOIPPrefixCreateResp, *Response, error) {

	if byoip.Prefix == "" {
		return nil, nil, fmt.Errorf("prefix is required")
	}
	if byoip.Signature == "" {
		return nil, nil, fmt.Errorf("signature is required")
	}
	if byoip.Region == "" {
		return nil, nil, fmt.Errorf("region is required")
	}

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

func (r *BYOIPServiceOp) Delete(ctx context.Context, uuid string) (*Response, error) {
	path := fmt.Sprintf("%s/%s", byoipsBasePath, uuid)

	req, err := r.client.NewRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return nil, err
	}
	resp, err := r.client.Do(ctx, req, nil)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
