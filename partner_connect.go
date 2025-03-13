package godo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const partnerConnectBasePath = "/v2/partner_connect/attachments"

// PartnerConnectService is an interface for managing Partner Connect with the
// DigitalOcean API.
// See: https://docs.digitalocean.com/reference/api/api-reference/#tag/PartnerConnect
type PartnerConnectService interface {
	List(context.Context, *ListOptions) ([]*PartnerConnect, *Response, error)
	Create(context.Context, *PartnerConnectCreateRequest) (*PartnerConnect, *Response, error)
	Get(context.Context, string) (*PartnerConnect, *Response, error)
	Update(context.Context, string, *PartnerConnectUpdateRequest) (*PartnerConnect, *Response, error)
	Delete(context.Context, string) (*Response, error)
	GetServiceKey(context.Context, string) (*ServiceKey, *Response, error)
	SetRoutes(context.Context, string, *PartnerConnectSetRoutesRequest) (*PartnerConnect, *Response, error)
	ListRoutes(context.Context, string, *ListOptions) ([]*RemoteRoute, *Response, error)
	GetBGPAuthKey(ctx context.Context, iaID string) (*BgpAuthKey, *Response, error)
	RegenerateServiceKey(ctx context.Context, iaID string) (*RegenerateServiceKey, *Response, error)
}

var _ PartnerConnectService = &PartnerConnectsServiceOp{}

// PartnerConnectsServiceOp interfaces with the Partner Connect endpoints in the DigitalOcean API.
type PartnerConnectsServiceOp struct {
	client *Client
}

// PartnerConnectCreateRequest represents a request to create a Partner Connect.
type PartnerConnectCreateRequest struct {
	// Name is the name of the Partner Connect
	Name string `json:"name,omitempty"`
	// ConnectionBandwidthInMbps is the bandwidth of the connection in Mbps
	ConnectionBandwidthInMbps int `json:"connection_bandwidth_in_mbps,omitempty"`
	// Region is the region where the Partner Connect is created
	Region string `json:"region,omitempty"`
	// NaaSProvider is the name of the Network as a Service provider
	NaaSProvider string `json:"naas_provider,omitempty"`
	// VPCIDs is the IDs of the VPCs to which the Partner Connect is connected
	VPCIDs []string `json:"vpc_ids,omitempty"`
	// BGP is the BGP configuration of the Partner Connect
	BGP BGP `json:"bgp,omitempty"`
}

type partnerConnectRequestBody struct {
	// Name is the name of the Partner Connect
	Name string `json:"name,omitempty"`
	// ConnectionBandwidthInMbps is the bandwidth of the connection in Mbps
	ConnectionBandwidthInMbps int `json:"connection_bandwidth_in_mbps,omitempty"`
	// Region is the region where the Partner Connect is created
	Region string `json:"region,omitempty"`
	// NaaSProvider is the name of the Network as a Service provider
	NaaSProvider string `json:"naas_provider,omitempty"`
	// VPCIDs is the IDs of the VPCs to which the Partner Connect is connected
	VPCIDs []string `json:"vpc_ids,omitempty"`
	// BGP is the BGP configuration of the Partner Connect
	BGP *BGPInput `json:"bgp,omitempty"`
}

func (req *PartnerConnectCreateRequest) buildReq() *partnerConnectRequestBody {
	request := &partnerConnectRequestBody{
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

// PartnerConnectUpdateRequest represents a request to update a Partner Connect.
type PartnerConnectUpdateRequest struct {
	// Name is the name of the Partner Connect
	Name string `json:"name,omitempty"`
	//VPCIDs is the IDs of the VPCs to which the Partner Connect is connected
	VPCIDs []string `json:"vpc_ids,omitempty"`
}

type PartnerConnectSetRoutesRequest struct {
	// Routes is the list of routes to be used for the Partner Connect
	Routes []string `json:"routes,omitempty"`
}

// BGP represents the BGP configuration of a Partner Connect.
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

// BGPInput represents the BGP configuration of a Partner Connect.
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

// ServiceKey represents the service key of a Partner Connect.
type ServiceKey struct {
	Value     string    `json:"value,omitempty"`
	State     string    `json:"state,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
}

// RemoteRoute represents a route for a Partner Connect.
type RemoteRoute struct {
	// ID is the generated ID of the Route
	ID string `json:"id,omitempty"`
	// Cidr is the CIDR of the route
	Cidr string `json:"cidr,omitempty"`
}

// PartnerConnect represents a DigitalOcean Partner Connect.
type PartnerConnect struct {
	// ID is the generated ID of the Partner Connect
	ID string `json:"id,omitempty"`
	// Name is the name of the Partner Connect
	Name string `json:"name,omitempty"`
	// State is the state of the Partner Connect
	State string `json:"state,omitempty"`
	// ConnectionBandwidthInMbps is the bandwidth of the connection in Mbps
	ConnectionBandwidthInMbps int `json:"connection_bandwidth_in_mbps,omitempty"`
	// Region is the region where the Partner Connect is created
	Region string `json:"region,omitempty"`
	// NaaSProvider is the name of the Network as a Service provider
	NaaSProvider string `json:"naas_provider,omitempty"`
	// VPCIDs is the IDs of the VPCs to which the Partner Connect is connected
	VPCIDs []string `json:"vpc_ids,omitempty"`
	// BGP is the BGP configuration of the Partner Connect
	BGP BGP `json:"bgp,omitempty"`
	// CreatedAt is time when this Partner Connect was first created
	CreatedAt time.Time `json:"created_at,omitempty"`
}

type partnerConnectAttachmentRoot struct {
	PartnerConnect *PartnerConnect `json:"-"`
}

func (r *partnerConnectAttachmentRoot) UnmarshalJSON(data []byte) error {
	// auxiliary structure to capture both potential keys
	var aux struct {
		PartnerConnect                *PartnerConnect `json:"partner_connect"`
		PartnerInterconnectAttachment *PartnerConnect `json:"partner_interconnect_attachment"`
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if aux.PartnerConnect != nil {
		r.PartnerConnect = aux.PartnerConnect
	} else {
		r.PartnerConnect = aux.PartnerInterconnectAttachment
	}
	return nil
}

type partnerConnectsRoot struct {
	PartnerConnects []*PartnerConnect `json:"-"`
	Links           *Links            `json:"links"`
	Meta            *Meta             `json:"meta"`
}

func (r *partnerConnectsRoot) UnmarshalJSON(data []byte) error {
	var aux struct {
		PartnerInterconnectAttachments []*PartnerConnect `json:"partner_interconnect_attachments"`
		PartnerConnects                []*PartnerConnect `json:"partner_connects"`
		Links                          *Links            `json:"links"`
		Meta                           *Meta             `json:"meta"`
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if aux.PartnerConnects != nil {
		r.PartnerConnects = aux.PartnerConnects
	} else {
		r.PartnerConnects = aux.PartnerInterconnectAttachments
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

// List returns a list of all Partner Connect, with optional pagination.
func (s *PartnerConnectsServiceOp) List(ctx context.Context, opt *ListOptions) ([]*PartnerConnect, *Response, error) {
	path, err := addOptions(partnerConnectBasePath, opt)
	if err != nil {
		return nil, nil, err
	}
	req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	root := new(partnerConnectsRoot)
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
	return root.PartnerConnects, resp, nil
}

// Create creates a new Partner Connect.
func (s *PartnerConnectsServiceOp) Create(ctx context.Context, create *PartnerConnectCreateRequest) (*PartnerConnect, *Response, error) {
	path := partnerConnectBasePath

	req, err := s.client.NewRequest(ctx, http.MethodPost, path, create.buildReq())
	if err != nil {
		return nil, nil, err
	}

	root := new(partnerConnectAttachmentRoot)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}

	return root.PartnerConnect, resp, nil
}

// Get returns the details of a Partner Connect.
func (s *PartnerConnectsServiceOp) Get(ctx context.Context, id string) (*PartnerConnect, *Response, error) {
	path := fmt.Sprintf("%s/%s", partnerConnectBasePath, id)
	req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	root := new(partnerConnectAttachmentRoot)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}

	return root.PartnerConnect, resp, nil
}

// Update updates a Partner Connect properties.
func (s *PartnerConnectsServiceOp) Update(ctx context.Context, id string, update *PartnerConnectUpdateRequest) (*PartnerConnect, *Response, error) {
	path := fmt.Sprintf("%s/%s", partnerConnectBasePath, id)
	req, err := s.client.NewRequest(ctx, http.MethodPatch, path, update)
	if err != nil {
		return nil, nil, err
	}

	root := new(partnerConnectAttachmentRoot)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}

	return root.PartnerConnect, resp, nil
}

// Delete deletes a Partner Connect.
func (s *PartnerConnectsServiceOp) Delete(ctx context.Context, id string) (*Response, error) {
	path := fmt.Sprintf("%s/%s", partnerConnectBasePath, id)
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

func (s *PartnerConnectsServiceOp) GetServiceKey(ctx context.Context, id string) (*ServiceKey, *Response, error) {
	path := fmt.Sprintf("%s/%s/service_key", partnerConnectBasePath, id)
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

// ListRoutes lists all remote routes for a Partner Connect.
func (s *PartnerConnectsServiceOp) ListRoutes(ctx context.Context, id string, opt *ListOptions) ([]*RemoteRoute, *Response, error) {
	path, err := addOptions(fmt.Sprintf("%s/%s/remote_routes", partnerConnectBasePath, id), opt)
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

// SetRoutes updates specific properties of a Partner Connect.
func (s *PartnerConnectsServiceOp) SetRoutes(ctx context.Context, id string, set *PartnerConnectSetRoutesRequest) (*PartnerConnect, *Response, error) {
	path := fmt.Sprintf("%s/%s/remote_routes", partnerConnectBasePath, id)
	req, err := s.client.NewRequest(ctx, http.MethodPut, path, set)
	if err != nil {
		return nil, nil, err
	}

	root := new(partnerConnectAttachmentRoot)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}

	return root.PartnerConnect, resp, nil
}

// GetBGPAuthKey returns Partner Connect bgp auth key
func (s *PartnerConnectsServiceOp) GetBGPAuthKey(ctx context.Context, iaID string) (*BgpAuthKey, *Response, error) {
	path := fmt.Sprintf("%s/%s/bgp_auth_key", partnerConnectBasePath, iaID)
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

// RegenerateServiceKey regenerates the service key of a Partner Connect.
func (s *PartnerConnectsServiceOp) RegenerateServiceKey(ctx context.Context, iaID string) (*RegenerateServiceKey, *Response, error) {
	path := fmt.Sprintf("%s/%s/service_key", partnerConnectBasePath, iaID)
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
