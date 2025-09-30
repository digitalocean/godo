package godo

import (
	"context"
	"fmt"
	"net/http"
)

const nfsBasePath = "v2/nfs"

type NfsService interface {
	List(context.Context, *ListOptions) ([]*Nfs, *Response, error)
	Create(context.Context, *NfsCreateRequest) (*Nfs, *Response, error)
	Delete(context.Context, string) (*Response, error)
	Get(context.Context, string) (*Nfs, *Response, error)
}

// NfsServiceOp handles communication with the NFS related methods of the
// DigitalOcean API.
type NfsServiceOp struct {
	client *Client
}

var _ NfsService = &NfsServiceOp{}

// Nfs represents a DigitalOcean NFS share
type Nfs struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	SizeGib   int      `json:"size_gib"`
	Region    string   `json:"region"`
	Status    string   `json:"status"`
	CreatedAt string   `json:"created_at"`
	VpcIDs    []string `json:"vpc_ids"`
	UserID    int      `json:"user_id"`
}

// NfsCreateRequest represents a request to create an NFS share.
type NfsCreateRequest struct {
	Name    string   `json:"name"`
	SizeGib int      `json:"size_gib"`
	Region  string   `json:"region"`
	VpcIDs  []string `json:"vpc_ids,omitempty"`
	UserID  int      `json:"user_id"`
}

// nfsRoot represents a response from the DigitalOcean API
type nfsRoot struct {
	Share *Nfs `json:"share"`
}

// nfsListRoot represents a response from the DigitalOcean API
type nfsListRoot struct {
	Shares []*Nfs `json:"shares,omitempty"`
	Links  *Links `json:"links,omitempty"`
	Meta   *Meta  `json:"meta"`
}

// Create creates a new NFS share.
func (s *NfsServiceOp) Create(ctx context.Context, createRequest *NfsCreateRequest) (*Nfs, *Response, error) {
	if createRequest == nil {
		return nil, nil, NewArgError("createRequest", "cannot be nil")
	}

	req, err := s.client.NewRequest(ctx, http.MethodPost, nfsBasePath, createRequest)
	if err != nil {
		return nil, nil, err
	}

	root := new(nfsRoot)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}

	return root.Share, resp, nil
}

// Get retrieves an NFS share by ID.
func (s *NfsServiceOp) Get(ctx context.Context, id string) (*Nfs, *Response, error) {
	if id == "" {
		return nil, nil, NewArgError("id", "cannot be empty")
	}

	path := fmt.Sprintf("%s/%s", nfsBasePath, id)
	req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	root := new(nfsRoot)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}

	return root.Share, resp, nil
}

// List returns a list of NFS shares.
func (s *NfsServiceOp) List(ctx context.Context, opts *ListOptions) ([]*Nfs, *Response, error) {
	path, err := addOptions(nfsBasePath, opts)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	root := new(nfsListRoot)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}

	if root.Links != nil {
		resp.Links = root.Links
	}
	if root.Meta != nil {
		resp.Meta = root.Meta
	}

	return root.Shares, resp, nil
}

// Delete deletes an NFS share.
func (s *NfsServiceOp) Delete(ctx context.Context, id string) (*Response, error) {
	if id == "" {
		return nil, NewArgError("id", "cannot be empty")
	}

	path := fmt.Sprintf("%s/%s", nfsBasePath, id)
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
