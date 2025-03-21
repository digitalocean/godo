package godo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const partnerNetworkConnectBasePath = "/v2/partner_network_connect/attachments"

// PartnerNetworkConnectService is an interface for managing Partner Attachment with the
// DigitalOcean API.
// See: https://docs.digitalocean.com/reference/api/api-reference/#tag/PartnerNetworkConnect
type PartnerNetworkConnectService interface {
	List(context.Context, *ListOptions) ([]*Attachment, *Response, error)
	Create(context.Context, *PartnerNetworkConnectCreateRequest) (*Attachment, *Response, error)
	Get(context.Context, string) (*Attachment, *Response, error)
	Update(context.Context, string, *PartnerNetworkConnectUpdateRequest) (*Attachment, *Response, error)
	Delete(context.Context, string) (*Response, error)
	GetServiceKey(context.Context, string) (*ServiceKey, *Response, error)
	SetRoutes(context.Context, string, *PartnerNetworkConnectSetRoutesRequest) (*Attachment, *Response, error)
	ListRoutes(context.Context, string, *ListOptions) ([]*RemoteRoute, *Response, error)
	GetBGPAuthKey(ctx context.Context, iaID string) (*BgpAuthKey, *Response, error)
	RegenerateServiceKey(ctx context.Context, iaID string) (*RegenerateServiceKey, *Response, error)
}

var _ PartnerNetworkConnectService = &PartnerNetworkConnectsServiceOp{}

// PartnerNetworkConnectsServiceOp interfaces with the Partner Attachment endpoints in the DigitalOcean API.
type PartnerNetworkConnectsServiceOp struct {
	client *Client
}

// PartnerNetworkConnectCreateRequest represents a request to create a Partner Attachment.
type PartnerNetworkConnectCreateRequest struct {
	// Name is the name of the Partner Attachment
	Name string `json:"name,omitempty"`
	// ConnectionBandwidthInMbps is the bandwidth of the connection in Mbps
	ConnectionBandwidthInMbps int `json:"connection_bandwidth_in_mbps,omitempty"`
	// Region is the region where the Partner Attachment is created
	Region string `json:"region,omitempty"`
	// NaaSProvider is the name of the Network as a Service provider
	NaaSProvider string `json:"naas_provider,omitempty"`
	// VPCIDs is the IDs of the VPCs to which the Partner Attachment is connected
	VPCIDs []string `json:"vpc_ids,omitempty"`
	// BGP is the BGP configuration of the Partner Attachment
	BGP BGP `json:"bgp,omitempty"`
}

type partnerNetworkConnectRequestBody struct {
	// Name is the name of the Partner Attachment
	Name string `json:"name,omitempty"`
	// ConnectionBandwidthInMbps is the bandwidth of the connection in Mbps
	ConnectionBandwidthInMbps int `json:"connection_bandwidth_in_mbps,omitempty"`
	// Region is the region where the Partner Attachment is created
	Region string `json:"region,omitempty"`
	// NaaSProvider is the name of the Network as a Service provider
	NaaSProvider string `json:"naas_provider,omitempty"`
	// VPCIDs is the IDs of the VPCs to which the Partner Attachment is connected
	VPCIDs []string `json:"vpc_ids,omitempty"`
	// BGP is the BGP configuration of the Partner Attachment
	BGP *BGPInput `json:"bgp,omitempty"`
}

func (req *PartnerNetworkConnectCreateRequest) buildReq() *partnerNetworkConnectRequestBody {
	request := &partnerNetworkConnectRequestBody{
		Name:                      req.Name,
		ConnectionBandwidthInMbps: req.ConnectionBandwidthInMbps,
		Region:                    req.Region,
		NaaSProvider:              req.NaaSProvider,
		VPCIDs:                    req.VPCIDs,
	}

	if req.BGP != (BGP{}) {
		request.BGP = &BGPInput{
			LocalASN:      req.BGP.LocalASN,
			LocalRouterIP: req.BGP.LocalRouterIP,
			PeerASN:       req.BGP.PeerASN,
			PeerRouterIP:  req.BGP.PeerRouterIP,
			AuthKey:       req.BGP.AuthKey,
		}
	}

	return request
}

// PartnerNetworkConnectUpdateRequest represents a request to update a Partner Attachment.
type PartnerNetworkConnectUpdateRequest struct {
	// Name is the name of the Partner Attachment
	Name string `json:"name,omitempty"`
	//VPCIDs is the IDs of the VPCs to which the Partner Attachment is connected
	VPCIDs []string `json:"vpc_ids,omitempty"`
}

type PartnerNetworkConnectSetRoutesRequest struct {
	// Routes is the list of routes to be used for the Partner Attachment
	Routes []string `json:"routes,omitempty"`
}

// BGP represents the BGP configuration of a Partner Attachment.
type BGP struct {
	// LocalASN is the local ASN
	LocalASN int `json:"local_asn,omitempty"`
	// LocalRouterIP is the local router IP
	LocalRouterIP string `json:"local_router_ip,omitempty"`
	// PeerASN is the peer ASN
	PeerASN int `json:"peer_asn,omitempty"`
	// PeerRouterIP is the peer router IP
	PeerRouterIP string `json:"peer_router_ip,omitempty"`
	// AuthKey is the authentication key
	AuthKey string `json:"auth_key,omitempty"`
}

func (b *BGP) UnmarshalJSON(data []byte) error {
	type Alias BGP
	aux := &struct {
		LocalASN       *int `json:"local_asn,omitempty"`
		LocalRouterASN *int `json:"local_router_asn,omitempty"`
		PeerASN        *int `json:"peer_asn,omitempty"`
		PeerRouterASN  *int `json:"peer_router_asn,omitempty"`
		*Alias
	}{
		Alias: (*Alias)(b),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if aux.LocalASN != nil {
		b.LocalASN = *aux.LocalASN
	} else if aux.LocalRouterASN != nil {
		b.LocalASN = *aux.LocalRouterASN
	}

	if aux.PeerASN != nil {
		b.PeerASN = *aux.PeerASN
	} else if aux.PeerRouterASN != nil {
		b.PeerASN = *aux.PeerRouterASN
	}
	return nil
}

// BGPInput represents the BGP configuration of a Partner Attachment.
type BGPInput struct {
	// LocalASN is the local ASN
	LocalASN int `json:"local_router_asn,omitempty"`
	// LocalRouterIP is the local router IP
	LocalRouterIP string `json:"local_router_ip,omitempty"`
	// PeerASN is the peer ASN
	PeerASN int `json:"peer_router_asn,omitempty"`
	// PeerRouterIP is the peer router IP
	PeerRouterIP string `json:"peer_router_ip,omitempty"`
	// AuthKey is the authentication key
	AuthKey string `json:"auth_key,omitempty"`
}

// ServiceKey represents the service key of a Partner Attachment.
type ServiceKey struct {
	Value     string    `json:"value,omitempty"`
	State     string    `json:"state,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
}

// RemoteRoute represents a route for a Partner Attachment.
type RemoteRoute struct {
	// ID is the generated ID of the Route
	ID string `json:"id,omitempty"`
	// Cidr is the CIDR of the route
	Cidr string `json:"cidr,omitempty"`
}

// Attachment represents a DigitalOcean Partner Attachment.
type Attachment struct {
	// ID is the generated ID of the Partner Attachment
	ID string `json:"id,omitempty"`
	// Name is the name of the Partner Attachment
	Name string `json:"name,omitempty"`
	// State is the state of the Partner Attachment
	State string `json:"state,omitempty"`
	// ConnectionBandwidthInMbps is the bandwidth of the connection in Mbps
	ConnectionBandwidthInMbps int `json:"connection_bandwidth_in_mbps,omitempty"`
	// Region is the region where the Partner Attachment is created
	Region string `json:"region,omitempty"`
	// NaaSProvider is the name of the Network as a Service provider
	NaaSProvider string `json:"naas_provider,omitempty"`
	// VPCIDs is the IDs of the VPCs to which the Partner Attachment is connected
	VPCIDs []string `json:"vpc_ids,omitempty"`
	// BGP is the BGP configuration of the Partner Attachment
	BGP BGP `json:"bgp,omitempty"`
	// CreatedAt is time when this Partner Attachment was first created
	CreatedAt time.Time `json:"created_at,omitempty"`
}

type partnerNetworkConnectAttachmentRoot struct {
	Attachment *Attachment `json:"-"`
}

func (r *partnerNetworkConnectAttachmentRoot) UnmarshalJSON(data []byte) error {
	// auxiliary structure to capture both potential keys
	var aux struct {
		PartnerNetworkConnect *Attachment `json:"partner_network_connect"`
		Attachment            *Attachment `json:"attachment"`
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if aux.PartnerNetworkConnect != nil {
		r.Attachment = aux.PartnerNetworkConnect
	} else {
		r.Attachment = aux.Attachment
	}
	return nil
}

type partnerNetworkConnectAttachmentsRoot struct {
	Attachments []*Attachment `json:"-"`
	Links       *Links        `json:"links"`
	Meta        *Meta         `json:"meta"`
}

func (r *partnerNetworkConnectAttachmentsRoot) UnmarshalJSON(data []byte) error {
	var aux struct {
		Attachments            []*Attachment `json:"attachments"`
		PartnerNetworkConnects []*Attachment `json:"partner_network_connects"`
		Links                  *Links        `json:"links"`
		Meta                   *Meta         `json:"meta"`
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if aux.PartnerNetworkConnects != nil {
		r.Attachments = aux.PartnerNetworkConnects
	} else {
		r.Attachments = aux.Attachments
	}

	r.Links = aux.Links
	r.Meta = aux.Meta

	return nil
}

type serviceKeyRoot struct {
	ServiceKey *ServiceKey `json:"service_key"`
}

type remoteRoutesRoot struct {
	RemoteRoutes []*RemoteRoute `json:"remote_routes"`
	Links        *Links         `json:"links"`
	Meta         *Meta          `json:"meta"`
}

type BgpAuthKey struct {
	Value string `json:"value"`
}

type bgpAuthKeyRoot struct {
	BgpAuthKey *BgpAuthKey `json:"bgp_auth_key"`
}

type RegenerateServiceKey struct {
}

type regenerateServiceKeyRoot struct {
	RegenerateServiceKey *RegenerateServiceKey `json:"-"`
}

// List returns a list of all Partner Attachment, with optional pagination.
func (s *PartnerNetworkConnectsServiceOp) List(ctx context.Context, opt *ListOptions) ([]*Attachment, *Response, error) {
	path, err := addOptions(partnerNetworkConnectBasePath, opt)
	if err != nil {
		return nil, nil, err
	}
	req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	root := new(partnerNetworkConnectAttachmentsRoot)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}
	if l := root.Links; l != nil {
		resp.Links = l
	}
	if m := root.Meta; m != nil {
		resp.Meta = m
	}
	return root.Attachments, resp, nil
}

// Create creates a new Partner Attachment.
func (s *PartnerNetworkConnectsServiceOp) Create(ctx context.Context, create *PartnerNetworkConnectCreateRequest) (*Attachment, *Response, error) {
	path := partnerNetworkConnectBasePath

	req, err := s.client.NewRequest(ctx, http.MethodPost, path, create.buildReq())
	if err != nil {
		return nil, nil, err
	}

	root := new(partnerNetworkConnectAttachmentRoot)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}

	return root.Attachment, resp, nil
}

// Get returns the details of a Partner Attachment.
func (s *PartnerNetworkConnectsServiceOp) Get(ctx context.Context, id string) (*Attachment, *Response, error) {
	path := fmt.Sprintf("%s/%s", partnerNetworkConnectBasePath, id)
	req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	root := new(partnerNetworkConnectAttachmentRoot)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}

	return root.Attachment, resp, nil
}

// Update updates a Partner Attachment properties.
func (s *PartnerNetworkConnectsServiceOp) Update(ctx context.Context, id string, update *PartnerNetworkConnectUpdateRequest) (*Attachment, *Response, error) {
	path := fmt.Sprintf("%s/%s", partnerNetworkConnectBasePath, id)
	req, err := s.client.NewRequest(ctx, http.MethodPatch, path, update)
	if err != nil {
		return nil, nil, err
	}

	root := new(partnerNetworkConnectAttachmentRoot)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}

	return root.Attachment, resp, nil
}

// Delete deletes a Partner Attachment.
func (s *PartnerNetworkConnectsServiceOp) Delete(ctx context.Context, id string) (*Response, error) {
	path := fmt.Sprintf("%s/%s", partnerNetworkConnectBasePath, id)
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

func (s *PartnerNetworkConnectsServiceOp) GetServiceKey(ctx context.Context, id string) (*ServiceKey, *Response, error) {
	path := fmt.Sprintf("%s/%s/service_key", partnerNetworkConnectBasePath, id)
	req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	root := new(serviceKeyRoot)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}

	return root.ServiceKey, resp, nil
}

// ListRoutes lists all remote routes for a Partner Attachment.
func (s *PartnerNetworkConnectsServiceOp) ListRoutes(ctx context.Context, id string, opt *ListOptions) ([]*RemoteRoute, *Response, error) {
	path, err := addOptions(fmt.Sprintf("%s/%s/remote_routes", partnerNetworkConnectBasePath, id), opt)
	if err != nil {
		return nil, nil, err
	}
	req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	root := new(remoteRoutesRoot)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}
	if l := root.Links; l != nil {
		resp.Links = l
	}
	if m := root.Meta; m != nil {
		resp.Meta = m
	}

	return root.RemoteRoutes, resp, nil
}

// SetRoutes updates specific properties of a Partner Attachment.
func (s *PartnerNetworkConnectsServiceOp) SetRoutes(ctx context.Context, id string, set *PartnerNetworkConnectSetRoutesRequest) (*Attachment, *Response, error) {
	path := fmt.Sprintf("%s/%s/remote_routes", partnerNetworkConnectBasePath, id)
	req, err := s.client.NewRequest(ctx, http.MethodPut, path, set)
	if err != nil {
		return nil, nil, err
	}

	root := new(partnerNetworkConnectAttachmentRoot)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}

	return root.Attachment, resp, nil
}

// GetBGPAuthKey returns Partner Attachment bgp auth key
func (s *PartnerNetworkConnectsServiceOp) GetBGPAuthKey(ctx context.Context, iaID string) (*BgpAuthKey, *Response, error) {
	path := fmt.Sprintf("%s/%s/bgp_auth_key", partnerNetworkConnectBasePath, iaID)
	req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	root := new(bgpAuthKeyRoot)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}

	return root.BgpAuthKey, resp, nil
}

// RegenerateServiceKey regenerates the service key of a Partner Attachment.
func (s *PartnerNetworkConnectsServiceOp) RegenerateServiceKey(ctx context.Context, iaID string) (*RegenerateServiceKey, *Response, error) {
	path := fmt.Sprintf("%s/%s/service_key", partnerNetworkConnectBasePath, iaID)
	req, err := s.client.NewRequest(ctx, http.MethodPost, path, nil)
	if err != nil {
		return nil, nil, err
	}

	root := new(regenerateServiceKeyRoot)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}

	return root.RegenerateServiceKey, resp, nil
}
