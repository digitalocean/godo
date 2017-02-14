package godo

import "fmt"

const loadBalancerBasePath = "v2/load_balancers"

//var errNoNetworks = errors.New("no networks have been defined")

// LoadBalancerService is an interface for interfacing with the Load Balancer
// endpoints of the DigitalOcean API
// See: https://developers.digitalocean.com/documentation/v2#load_balancers
type LoadBalancersService interface {
	List(*ListOptions) ([]LoadBalancer, *Response, error)
	Get(string) (*LoadBalancer, *Response, error)
}

// LoadBalancersServiceOp handles communication with the Load Balancer related methods of the
// DigitalOcean API.
type LoadBalancersServiceOp struct {
	client *Client
}

var _ LoadBalancersService = &LoadBalancersServiceOp{}

type LoadBalancer struct {
	ID                  string            `json:"id,omitempty"`
	Name                string            `json:"name,omitempty"`
	IP                  string            `json:"ip,omitempty"`
	Algorithm           string            `json:"algorithm,omitempty"`
	Status              string            `json:"status,omitempty"`
	Created             string            `json:"created_at,omitempty"`
	ForwardingRules     []*ForwardingRule `json:"forwarding_rules,omitempty"`
	HealthCheck         *HealthCheck      `json:"health_check,omitempty"`
	StickySessions      *StickySession    `json:"sticky_sessions,omitempty"`
	Region              *Region           `json:"region,omitempty"`
	Tag                 string            `json:"tag,omitempty"`
	DropletIDs          []int             `json:"droplet_ids,omitempty"`
	RedirectHttpToHttps bool              `json:"redirect_http_to_https,omitempty"`
}

type ForwardingRule struct {
	EntryProtocol  string `json:"entry_protocol,omitempty"`
	EntryPort      int    `json:"entry_port,omitempty"`
	TargetProtocol string `json:"target_protocol,omitempty"`
	TargetPort     int    `json:"target_port,omitempty"`
	CertificateID  string `json:"certificate_id,omitempty"`
	TLSPassthrough bool   `json:"tls_passthrough,omitempty"`
}

type HealthCheck struct {
	Protocol               string `json:"protocol,omitempty"`
	Port                   int    `json:"port,omitempty"`
	Path                   string `json:"path,omitempty"`
	CheckIntervalSeconds   int    `json:"check_interval_seconds,omitempty"`
	RepsonseTimeoutSeconds int    `json:"repsonse_timeout_seconds,omitempty"`
	UnhealthyThreshold     int    `json:"unhealthy_threshold,omitempty"`
	HealthThreshold        int    `json:"health_threshold,omitempty"`
}

type StickySession struct {
	Type             string `json:"type,omitempty"`
	CookieName       string `json:"cookie_name,omitempty"`
	CookieTTLSeconds string `json:"cookie_ttl_seconds,omitempty"`
}

type loadBalancersRoot struct {
	LoadBalancers []LoadBalancer `json:"load_balancers"`
	Links         *Links         `json:"links"`
}

type loadBalancerRoot struct {
	LoadBalancer *LoadBalancer `json:"load_balancer,omitempty"`
	Links        *Links        `json:"links,omitempty"`
}

func (r LoadBalancer) String() string {
	return Stringify(r)
}

// Performs a list request given a path.
func (lb *LoadBalancersServiceOp) list(path string) ([]LoadBalancer, *Response, error) {
	req, err := lb.client.NewRequest("GET", path, nil)
	if err != nil {
		return nil, nil, err
	}

	root := new(loadBalancersRoot)
	resp, err := lb.client.Do(req, root)
	if err != nil {
		return nil, resp, err
	}
	if l := root.Links; l != nil {
		resp.Links = l
	}

	return root.LoadBalancers, resp, err
}

// List all Load Balancers.
func (lb *LoadBalancersServiceOp) List(opt *ListOptions) ([]LoadBalancer, *Response, error) {
	path := loadBalancerBasePath
	path, err := addOptions(path, opt)
	if err != nil {
		return nil, nil, err
	}

	return lb.list(path)
}

// Get an individual Load Balancer.
func (lb *LoadBalancersServiceOp) Get(id string) (*LoadBalancer, *Response, error) {
	path := fmt.Sprintf("%s/%s", loadBalancerBasePath, id)

	req, err := lb.client.NewRequest("GET", path, nil)
	if err != nil {
		return nil, nil, err
	}

	root := new(loadBalancerRoot)
	resp, err := lb.client.Do(req, root)
	if err != nil {
		return nil, resp, err
	}

	return root.LoadBalancer, resp, err
}
