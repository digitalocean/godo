package godo

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

const (
	natGatewaysBasePath = "/v2/nat_gateways"
)

// NatGatewaysService defines an interface for managing NAT Gateways through the DigitalOcean API
type NatGatewaysService interface {
	Create(context.Context, *NatGatewayRequest) (*NatGateway, *Response, error)
	Get(context.Context, string) (*NatGateway, *Response, error)
	List(context.Context, *NatGatewaysListOptions) ([]*NatGateway, *Response, error)
	Update(context.Context, string, *NatGatewayRequest) (*NatGateway, *Response, error)
	Delete(context.Context, string) (*Response, error)
}

// NatGatewayRequest represents a DigitalOcean NAT Gateway create/update request
type NatGatewayRequest struct {
	Name               string        `json:"name"`
	Type               string        `json:"type"`
	Region             string        `json:"region"`
	VPCs               []*IngressVPC `json:"vpcs"`
	UDPTimeoutSeconds  uint32        `json:"udp_timeout_seconds,omitempty"`
	ICMPTimeoutSeconds uint32        `json:"icmp_timeout_seconds,omitempty"`
	TCPTimeoutSeconds  uint32        `json:"tcp_timeout_seconds,omitempty"`
}

// NatGateway represents a DigitalOcean NAT Gateway resource
type NatGateway struct {
	ID                 string        `json:"id"`
	Name               string        `json:"name"`
	Type               string        `json:"type"`
	State              string        `json:"state"`
	Region             string        `json:"region"`
	VPCs               []*IngressVPC `json:"vpcs"`
	Egresses           *Egresses     `json:"egresses,omitempty"`
	UDPTimeoutSeconds  uint32        `json:"udp_timeout_seconds,omitempty"`
	ICMPTimeoutSeconds uint32        `json:"icmp_timeout_seconds,omitempty"`
	TCPTimeoutSeconds  uint32        `json:"tcp_timeout_seconds,omitempty"`
	CreatedAt          time.Time     `json:"created_at"`
	UpdatedAt          time.Time     `json:"updated_at"`
}

// IngressVPC defines the ingress configs supported by a NAT Gateway
type IngressVPC struct {
	VpcUUID           string `json:"vpc_uuid"`
	GatewayIP         string `json:"gateway_ip,omitempty"`
	DefaultNATGateway bool   `json:"default_nat_gateway,omitempty"`
}

// Egresses define egress routes supported by a NAT Gateway
type Egresses struct {
	PublicGateways []*PublicGateway `json:"public_gateways,omitempty"`
}

// PublicGateway defines the public egress supported by a NAT Gateway
type PublicGateway struct {
	IPv4 string `json:"ipv4"`
}

// NatGatewaysListOptions define custom options for listing NAT Gateways
type NatGatewaysListOptions struct {
	ListOptions
	State  []string `json:"state,omitempty"`
	Region []string `json:"region,omitempty"`
	Type   []string `json:"type,omitempty"`
	Name   []string `json:"name,omitempty"`
}

type natGatewayRoot struct {
	NatGateway *NatGateway `json:"nat_gateway"`
}

type natGatewaysRoot struct {
	NatGateways []*NatGateway `json:"nat_gateways"`
	Links       *Links        `json:"links"`
	Meta        *Meta         `json:"meta"`
}

// NatGatewaysServiceOp handles communication with NAT Gateway methods of the DigitalOcean API
type NatGatewaysServiceOp struct {
	client *Client
}

var _ NatGatewaysService = &NatGatewaysServiceOp{}

// Create a new NAT Gateway
func (n *NatGatewaysServiceOp) Create(ctx context.Context, createReq *NatGatewayRequest) (*NatGateway, *Response, error) {
	req, err := n.client.NewRequest(ctx, http.MethodPost, natGatewaysBasePath, createReq)
	if err != nil {
		return nil, nil, err
	}
	root := new(natGatewayRoot)
	resp, err := n.client.Do(ctx, req, root)
	if err != nil {
		return nil, nil, err
	}
	return root.NatGateway, resp, nil
}

// Get an existing NAT Gateway
func (n *NatGatewaysServiceOp) Get(ctx context.Context, id string) (*NatGateway, *Response, error) {
	req, err := n.client.NewRequest(ctx, http.MethodGet, fmt.Sprintf("%s/%s", natGatewaysBasePath, id), nil)
	if err != nil {
		return nil, nil, err
	}
	root := new(natGatewayRoot)
	resp, err := n.client.Do(ctx, req, root)
	if err != nil {
		return nil, nil, err
	}
	return root.NatGateway, resp, nil
}

// List all active NAT Gateways
func (n *NatGatewaysServiceOp) List(ctx context.Context, opts *NatGatewaysListOptions) ([]*NatGateway, *Response, error) {
	path, err := addOptions(natGatewaysBasePath, opts)
	if err != nil {
		return nil, nil, err
	}
	req, err := n.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}
	root := new(natGatewaysRoot)
	resp, err := n.client.Do(ctx, req, root)
	if err != nil {
		return nil, nil, err
	}
	if root.Links != nil {
		resp.Links = root.Links
	}
	if root.Meta != nil {
		resp.Meta = root.Meta
	}
	return root.NatGateways, resp, nil
}

// Update an existing NAT Gateway
func (n *NatGatewaysServiceOp) Update(ctx context.Context, id string, updateReq *NatGatewayRequest) (*NatGateway, *Response, error) {
	req, err := n.client.NewRequest(ctx, http.MethodPut, fmt.Sprintf("%s/%s", natGatewaysBasePath, id), updateReq)
	if err != nil {
		return nil, nil, err
	}
	root := new(natGatewayRoot)
	resp, err := n.client.Do(ctx, req, root)
	if err != nil {
		return nil, nil, err
	}
	return root.NatGateway, resp, nil
}

// Delete an existing NAT Gateway
func (n *NatGatewaysServiceOp) Delete(ctx context.Context, id string) (*Response, error) {
	req, err := n.client.NewRequest(ctx, http.MethodDelete, fmt.Sprintf("%s/%s", natGatewaysBasePath, id), nil)
	if err != nil {
		return nil, err
	}
	return n.client.Do(ctx, req, nil)
}
