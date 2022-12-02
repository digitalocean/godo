package godo

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var lbListJSONResponse = `
{
	"load_balancers":[
        {
            "id":"37e6be88-01ec-4ec7-9bc6-a514d4719057",
            "name":"example-lb-01",
            "ip":"46.214.185.203",
            "algorithm":"round_robin",
            "status":"active",
            "created_at":"2016-12-15T14:16:36Z",
            "forwarding_rules":[
                {
                    "entry_protocol":"https",
                    "entry_port":443,
                    "target_protocol":"http",
                    "target_port":80,
                    "certificate_id":"a-b-c"
                }
            ],
            "health_check":{
                "protocol":"http",
                "port":80,
                "path":"/index.html",
                "check_interval_seconds":10,
                "response_timeout_seconds":5,
                "healthy_threshold":5,
                "unhealthy_threshold":3
            },
            "sticky_sessions":{
                "type":"cookies",
                "cookie_name":"DO-LB",
                "cookie_ttl_seconds":5
            },
            "region":{
            	"name":"New York 1",
                "slug":"nyc1",
                "sizes":[
                    "512mb",
                    "1gb",
                    "2gb",
                    "4gb",
                    "8gb",
                    "16gb"
                ],
                "features":[
                    "private_networking",
                    "backups",
                    "ipv6",
                    "metadata",
                    "storage"
                ],
                "available":true
            },
            "droplet_ids":[
                2,
                21
            ],
            "disable_lets_encrypt_dns_records": true,
            "project_id": "6929eef6-4e45-11ed-bdc3-0242ac120002",
            "http_idle_timeout_seconds": 60,
            "firewall": {
                "deny": ["cidr:1.2.0.0/16"],
                "allow": ["ip:1.2.3.4"]
            }
        }
    ],
    "links":{
        "pages":{
            "last":"http://localhost:3001/v2/load_balancers?page=3&per_page=1",
            "next":"http://localhost:3001/v2/load_balancers?page=2&per_page=1"
        }
    },
    "meta":{
        "total":3
    }
}
`

var lbCreateJSONResponse = `
{
    "load_balancer":{
        "id":"8268a81c-fcf5-423e-a337-bbfe95817f23",
        "name":"example-lb-01",
        "ip":"",
        "algorithm":"round_robin",
        "status":"new",
        "created_at":"2016-12-15T14:19:09Z",
        "forwarding_rules":[
            {
                "entry_protocol":"https",
                "entry_port":443,
                "target_protocol":"http",
                "target_port":80,
                "certificate_id":"a-b-c"
            },
            {
                "entry_protocol":"https",
                "entry_port":444,
                "target_protocol":"https",
                "target_port":443,
                "tls_passthrough":true
            }
        ],
        "health_check":{
            "protocol":"http",
            "port":80,
            "path":"/index.html",
            "check_interval_seconds":10,
            "response_timeout_seconds":5,
            "healthy_threshold":5,
            "unhealthy_threshold":3
        },
        "sticky_sessions":{
            "type":"cookies",
            "cookie_name":"DO-LB",
            "cookie_ttl_seconds":5
        },
        "region":{
            "name":"New York 1",
            "slug":"nyc1",
            "sizes":[
                "512mb",
                "1gb",
                "2gb",
                "4gb",
                "8gb",
                "16gb"
            ],
            "features":[
                "private_networking",
                "backups",
                "ipv6",
                "metadata",
                "storage"
            ],
            "available":true
		},
		"tags": ["my-tag"],
        "droplet_ids":[
            2,
            21
        ],
        "redirect_http_to_https":true,
        "vpc_uuid":"880b7f98-f062-404d-b33c-458d545696f6",
        "disable_lets_encrypt_dns_records": true,
        "project_id": "6929eef6-4e45-11ed-bdc3-0242ac120002",
        "http_idle_timeout_seconds": 60,
        "firewall": {
            "deny": ["cidr:1.2.0.0/16"],
            "allow": ["ip:1.2.3.4"]
        }
    }
}
`

var lbGetJSONResponse = `
{
    "load_balancer":{
        "id":"37e6be88-01ec-4ec7-9bc6-a514d4719057",
        "name":"example-lb-01",
        "ip":"46.214.185.203",
        "algorithm":"round_robin",
        "status":"active",
        "created_at":"2016-12-15T14:16:36Z",
        "forwarding_rules":[
            {
                "entry_protocol":"https",
                "entry_port":443,
                "target_protocol":"http",
                "target_port":80,
                "certificate_id":"a-b-c"
            }
        ],
        "health_check":{
            "protocol":"http",
            "port":80,
            "path":"/index.html",
            "check_interval_seconds":10,
            "response_timeout_seconds":5,
            "healthy_threshold":5,
            "unhealthy_threshold":3
        },
        "sticky_sessions":{
            "type":"cookies",
            "cookie_name":"DO-LB",
            "cookie_ttl_seconds":5
        },
        "region":{
            "name":"New York 1",
            "slug":"nyc1",
            "sizes":[
                "512mb",
                "1gb",
                "2gb",
                "4gb",
                "8gb",
                "16gb"
            ],
            "features":[
                "private_networking",
                "backups",
                "ipv6",
                "metadata",
                "storage"
            ],
            "available":true
        },
        "droplet_ids":[
            2,
            21
        ],
        "disable_lets_encrypt_dns_records": false,
        "project_id": "6929eef6-4e45-11ed-bdc3-0242ac120002",
        "http_idle_timeout_seconds": 60,
        "firewall": {
            "deny": ["cidr:1.2.0.0/16"],
            "allow": ["ip:1.2.3.4"]
        }
    }
}
`

var lbUpdateJSONResponse = `
{
    "load_balancer":{
        "id":"8268a81c-fcf5-423e-a337-bbfe95817f23",
        "name":"example-lb-01",
        "ip":"12.34.56.78",
        "algorithm":"least_connections",
        "status":"active",
        "size_unit":2,
        "created_at":"2016-12-15T14:19:09Z",
        "forwarding_rules":[
            {
                "entry_protocol":"http",
                "entry_port":80,
                "target_protocol":"http",
                "target_port":80
            },
            {
                "entry_protocol":"https",
                "entry_port":443,
                "target_protocol":"http",
                "target_port":80,
                "certificate_id":"a-b-c"
            }
        ],
        "health_check":{
            "protocol":"tcp",
            "port":80,
            "path":"",
            "check_interval_seconds":10,
            "response_timeout_seconds":5,
            "healthy_threshold":5,
            "unhealthy_threshold":3
        },
        "sticky_sessions":{
            "type":"none"
        },
        "region":{
            "name":"New York 1",
            "slug":"nyc1",
            "sizes":[
                "512mb",
                "1gb",
                "2gb",
                "4gb",
                "8gb",
                "16gb"
            ],
            "features":[
                "private_networking",
                "backups",
                "ipv6",
                "metadata",
                "storage"
            ],
            "available":true
        },
        "droplet_ids":[
            2,
            21
        ],
        "project_id": "6929eef6-4e45-11ed-bdc3-0242ac120002",
        "http_idle_timeout_seconds": 60,
        "firewall": {
            "deny": ["cidr:1.3.0.0/16"],
            "allow": ["ip:1.2.3.5"]
        }
    }
}
`

func TestLoadBalancers_Get(t *testing.T) {
	setup()
	defer teardown()

	path := "/v2/load_balancers"
	loadBalancerID := "37e6be88-01ec-4ec7-9bc6-a514d4719057"
	path = fmt.Sprintf("%s/%s", path, loadBalancerID)
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, lbGetJSONResponse)
	})

	loadBalancer, _, err := client.LoadBalancers.Get(ctx, loadBalancerID)
	require.NoError(t, err)
	expectedTimeout := uint64(60)

	expected := &LoadBalancer{
		ID:        "37e6be88-01ec-4ec7-9bc6-a514d4719057",
		Name:      "example-lb-01",
		IP:        "46.214.185.203",
		Algorithm: "round_robin",
		Status:    "active",
		Created:   "2016-12-15T14:16:36Z",
		ForwardingRules: []ForwardingRule{
			{
				EntryProtocol:  "https",
				EntryPort:      443,
				TargetProtocol: "http",
				TargetPort:     80,
				CertificateID:  "a-b-c",
				TlsPassthrough: false,
			},
		},
		HealthCheck: &HealthCheck{
			Protocol:               "http",
			Port:                   80,
			Path:                   "/index.html",
			CheckIntervalSeconds:   10,
			ResponseTimeoutSeconds: 5,
			HealthyThreshold:       5,
			UnhealthyThreshold:     3,
		},
		StickySessions: &StickySessions{
			Type:             "cookies",
			CookieName:       "DO-LB",
			CookieTtlSeconds: 5,
		},
		Region: &Region{
			Slug:      "nyc1",
			Name:      "New York 1",
			Sizes:     []string{"512mb", "1gb", "2gb", "4gb", "8gb", "16gb"},
			Available: true,
			Features:  []string{"private_networking", "backups", "ipv6", "metadata", "storage"},
		},
		DropletIDs:             []int{2, 21},
		ProjectID:              "6929eef6-4e45-11ed-bdc3-0242ac120002",
		HTTPIdleTimeoutSeconds: &expectedTimeout,
		Firewall: &LBFirewall{
			Allow: []string{"ip:1.2.3.4"},
			Deny:  []string{"cidr:1.2.0.0/16"},
		},
	}

	disableLetsEncryptDNSRecords := false
	expected.DisableLetsEncryptDNSRecords = &disableLetsEncryptDNSRecords

	assert.Equal(t, expected, loadBalancer)
}

func TestLoadBalancers_Create(t *testing.T) {
	setup()
	defer teardown()

	createRequest := &LoadBalancerRequest{
		Name:      "example-lb-01",
		Algorithm: "round_robin",
		Region:    "nyc1",
		ForwardingRules: []ForwardingRule{
			{
				EntryProtocol:  "https",
				EntryPort:      443,
				TargetProtocol: "http",
				TargetPort:     80,
				CertificateID:  "a-b-c",
			},
		},
		HealthCheck: &HealthCheck{
			Protocol:               "http",
			Port:                   80,
			Path:                   "/index.html",
			CheckIntervalSeconds:   10,
			ResponseTimeoutSeconds: 5,
			UnhealthyThreshold:     3,
			HealthyThreshold:       5,
		},
		StickySessions: &StickySessions{
			Type:             "cookies",
			CookieName:       "DO-LB",
			CookieTtlSeconds: 5,
		},
		Tag:                 "my-tag",
		Tags:                []string{"my-tag"},
		DropletIDs:          []int{2, 21},
		RedirectHttpToHttps: true,
		VPCUUID:             "880b7f98-f062-404d-b33c-458d545696f6",
		ProjectID:           "6929eef6-4e45-11ed-bdc3-0242ac120002",
		Firewall: &LBFirewall{
			Allow: []string{"ip:1.2.3.4"},
			Deny:  []string{"cidr:1.2.0.0/16"},
		},
	}

	path := "/v2/load_balancers"
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		v := new(LoadBalancerRequest)
		err := json.NewDecoder(r.Body).Decode(v)
		if err != nil {
			t.Fatal(err)
		}

		testMethod(t, r, http.MethodPost)
		assert.Equal(t, createRequest, v)

		fmt.Fprint(w, lbCreateJSONResponse)
	})

	loadBalancer, _, err := client.LoadBalancers.Create(ctx, createRequest)
	require.NoError(t, err)

	expectedTimeout := uint64(60)
	expected := &LoadBalancer{
		ID:        "8268a81c-fcf5-423e-a337-bbfe95817f23",
		Name:      "example-lb-01",
		Algorithm: "round_robin",
		Status:    "new",
		Created:   "2016-12-15T14:19:09Z",
		ForwardingRules: []ForwardingRule{
			{
				EntryProtocol:  "https",
				EntryPort:      443,
				TargetProtocol: "http",
				TargetPort:     80,
				CertificateID:  "a-b-c",
				TlsPassthrough: false,
			},
			{
				EntryProtocol:  "https",
				EntryPort:      444,
				TargetProtocol: "https",
				TargetPort:     443,
				CertificateID:  "",
				TlsPassthrough: true,
			},
		},
		HealthCheck: &HealthCheck{
			Protocol:               "http",
			Port:                   80,
			Path:                   "/index.html",
			CheckIntervalSeconds:   10,
			ResponseTimeoutSeconds: 5,
			HealthyThreshold:       5,
			UnhealthyThreshold:     3,
		},
		StickySessions: &StickySessions{
			Type:             "cookies",
			CookieName:       "DO-LB",
			CookieTtlSeconds: 5,
		},
		Region: &Region{
			Slug:      "nyc1",
			Name:      "New York 1",
			Sizes:     []string{"512mb", "1gb", "2gb", "4gb", "8gb", "16gb"},
			Available: true,
			Features:  []string{"private_networking", "backups", "ipv6", "metadata", "storage"},
		},
		Tags:                   []string{"my-tag"},
		DropletIDs:             []int{2, 21},
		RedirectHttpToHttps:    true,
		VPCUUID:                "880b7f98-f062-404d-b33c-458d545696f6",
		ProjectID:              "6929eef6-4e45-11ed-bdc3-0242ac120002",
		HTTPIdleTimeoutSeconds: &expectedTimeout,
		Firewall: &LBFirewall{
			Allow: []string{"ip:1.2.3.4"},
			Deny:  []string{"cidr:1.2.0.0/16"},
		},
	}

	disableLetsEncryptDNSRecords := true
	expected.DisableLetsEncryptDNSRecords = &disableLetsEncryptDNSRecords

	assert.Equal(t, expected, loadBalancer)
}

func TestLoadBalancers_CreateValidateSucceeds(t *testing.T) {
	setup()
	defer teardown()

	createRequest := &LoadBalancerRequest{
		Name:      "example-lb-01",
		Algorithm: "round_robin",
		Region:    "nyc1",
		ForwardingRules: []ForwardingRule{
			{
				EntryProtocol:  "https",
				EntryPort:      443,
				TargetProtocol: "http",
				TargetPort:     80,
				CertificateID:  "a-b-c",
			},
		},
		HealthCheck: &HealthCheck{
			Protocol:               "http",
			Port:                   80,
			Path:                   "/index.html",
			CheckIntervalSeconds:   10,
			ResponseTimeoutSeconds: 5,
			UnhealthyThreshold:     3,
			HealthyThreshold:       5,
		},
		StickySessions: &StickySessions{
			Type:             "cookies",
			CookieName:       "DO-LB",
			CookieTtlSeconds: 5,
		},
		Tag:                 "my-tag",
		Tags:                []string{"my-tag"},
		DropletIDs:          []int{2, 21},
		RedirectHttpToHttps: true,
		VPCUUID:             "880b7f98-f062-404d-b33c-458d545696f6",
		ValidateOnly:        true,
	}

	path := "/v2/load_balancers"
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		v := new(LoadBalancerRequest)
		err := json.NewDecoder(r.Body).Decode(v)
		if err != nil {
			t.Fatal(err)
		}

		w.WriteHeader(http.StatusAccepted)

		testMethod(t, r, http.MethodPost)
		assert.Equal(t, createRequest, v)

		fmt.Fprint(w, lbCreateJSONResponse)
	})

	loadBalancer, resp, err := client.LoadBalancers.Create(ctx, createRequest)
	require.NoError(t, err)
	assert.Equal(t, http.StatusAccepted, resp.StatusCode)
	assert.Nil(t, nil, loadBalancer)
}

func TestLoadBalancers_CreateValidateFails(t *testing.T) {
	setup()
	defer teardown()

	createRequest := &LoadBalancerRequest{
		Name:      "example-lb-01",
		Algorithm: "round_robin",
		Region:    "nyc1",
		ForwardingRules: []ForwardingRule{
			{
				EntryProtocol:  "https",
				EntryPort:      443,
				TargetProtocol: "http",
				TargetPort:     80,
				CertificateID:  "a-b-c",
			},
		},
		HealthCheck: &HealthCheck{
			Protocol:               "http",
			Port:                   80,
			Path:                   "/index.html",
			CheckIntervalSeconds:   10,
			ResponseTimeoutSeconds: 5,
			UnhealthyThreshold:     3,
			HealthyThreshold:       5,
		},
		StickySessions: &StickySessions{
			Type:             "cookies",
			CookieName:       "DO-LB",
			CookieTtlSeconds: 5,
		},
		Tag:                 "my-tag",
		Tags:                []string{"my-tag"},
		DropletIDs:          []int{2, 21},
		RedirectHttpToHttps: true,
		VPCUUID:             "880b7f98-f062-404d-b33c-458d545696f6",
		ValidateOnly:        true,
	}

	path := "/v2/load_balancers"
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		v := new(LoadBalancerRequest)
		err := json.NewDecoder(r.Body).Decode(v)
		if err != nil {
			t.Fatal(err)
		}

		w.WriteHeader(http.StatusUnprocessableEntity)
		testMethod(t, r, http.MethodPost)
		assert.Equal(t, createRequest, v)
	})

	loadBalancer, resp, err := client.LoadBalancers.Create(ctx, createRequest)
	require.Error(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
	assert.Nil(t, nil, loadBalancer)
}

func TestLoadBalancers_Update(t *testing.T) {
	setup()
	defer teardown()

	updateRequest := &LoadBalancerRequest{
		Name:      "example-lb-01",
		Algorithm: "least_connections",
		Region:    "nyc1",
		SizeUnit:  2,
		ForwardingRules: []ForwardingRule{
			{
				EntryProtocol:  "http",
				EntryPort:      80,
				TargetProtocol: "http",
				TargetPort:     80,
			},
			{
				EntryProtocol:  "https",
				EntryPort:      443,
				TargetProtocol: "http",
				TargetPort:     80,
				CertificateID:  "a-b-c",
			},
		},
		HealthCheck: &HealthCheck{
			Protocol:               "tcp",
			Port:                   80,
			Path:                   "",
			CheckIntervalSeconds:   10,
			ResponseTimeoutSeconds: 5,
			UnhealthyThreshold:     3,
			HealthyThreshold:       5,
		},
		StickySessions: &StickySessions{
			Type: "none",
		},
		DropletIDs: []int{2, 21},
		ProjectID:  "6929eef6-4e45-11ed-bdc3-0242ac120002",
		Firewall: &LBFirewall{
			Allow: []string{"ip:1.2.3.5"},
			Deny:  []string{"cidr:1.3.0.0/16"},
		},
	}

	path := "/v2/load_balancers"
	loadBalancerID := "8268a81c-fcf5-423e-a337-bbfe95817f23"
	path = fmt.Sprintf("%s/%s", path, loadBalancerID)

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		v := new(LoadBalancerRequest)
		err := json.NewDecoder(r.Body).Decode(v)
		if err != nil {
			t.Fatal(err)
		}

		testMethod(t, r, "PUT")
		assert.Equal(t, updateRequest, v)

		fmt.Fprint(w, lbUpdateJSONResponse)
	})

	loadBalancer, _, err := client.LoadBalancers.Update(ctx, loadBalancerID, updateRequest)
	require.NoError(t, err)

	expectedTimeout := uint64(60)
	expected := &LoadBalancer{
		ID:        "8268a81c-fcf5-423e-a337-bbfe95817f23",
		Name:      "example-lb-01",
		IP:        "12.34.56.78",
		Algorithm: "least_connections",
		Status:    "active",
		SizeUnit:  2,
		Created:   "2016-12-15T14:19:09Z",
		ForwardingRules: []ForwardingRule{
			{
				EntryProtocol:  "http",
				EntryPort:      80,
				TargetProtocol: "http",
				TargetPort:     80,
			},
			{
				EntryProtocol:  "https",
				EntryPort:      443,
				TargetProtocol: "http",
				TargetPort:     80,
				CertificateID:  "a-b-c",
			},
		},
		HealthCheck: &HealthCheck{
			Protocol:               "tcp",
			Port:                   80,
			Path:                   "",
			CheckIntervalSeconds:   10,
			ResponseTimeoutSeconds: 5,
			UnhealthyThreshold:     3,
			HealthyThreshold:       5,
		},
		StickySessions: &StickySessions{
			Type: "none",
		},
		Region: &Region{
			Slug:      "nyc1",
			Name:      "New York 1",
			Sizes:     []string{"512mb", "1gb", "2gb", "4gb", "8gb", "16gb"},
			Available: true,
			Features:  []string{"private_networking", "backups", "ipv6", "metadata", "storage"},
		},
		DropletIDs:                   []int{2, 21},
		DisableLetsEncryptDNSRecords: nil,
		ProjectID:                    "6929eef6-4e45-11ed-bdc3-0242ac120002",
		HTTPIdleTimeoutSeconds:       &expectedTimeout,
		Firewall: &LBFirewall{
			Allow: []string{"ip:1.2.3.5"},
			Deny:  []string{"cidr:1.3.0.0/16"},
		},
	}

	assert.Equal(t, expected, loadBalancer)
}

func TestLoadBalancers_List(t *testing.T) {
	setup()
	defer teardown()

	path := "/v2/load_balancers"
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, lbListJSONResponse)
	})

	loadBalancers, resp, err := client.LoadBalancers.List(ctx, nil)

	require.NoError(t, err)

	expectedTimeout := uint64(60)
	expectedLBs := []LoadBalancer{
		{
			ID:        "37e6be88-01ec-4ec7-9bc6-a514d4719057",
			Name:      "example-lb-01",
			IP:        "46.214.185.203",
			Algorithm: "round_robin",
			Status:    "active",
			Created:   "2016-12-15T14:16:36Z",
			ForwardingRules: []ForwardingRule{
				{
					EntryProtocol:  "https",
					EntryPort:      443,
					TargetProtocol: "http",
					TargetPort:     80,
					CertificateID:  "a-b-c",
				},
			},
			HealthCheck: &HealthCheck{
				Protocol:               "http",
				Port:                   80,
				Path:                   "/index.html",
				CheckIntervalSeconds:   10,
				ResponseTimeoutSeconds: 5,
				HealthyThreshold:       5,
				UnhealthyThreshold:     3,
			},
			StickySessions: &StickySessions{
				Type:             "cookies",
				CookieName:       "DO-LB",
				CookieTtlSeconds: 5,
			},
			Region: &Region{
				Slug:      "nyc1",
				Name:      "New York 1",
				Sizes:     []string{"512mb", "1gb", "2gb", "4gb", "8gb", "16gb"},
				Available: true,
				Features:  []string{"private_networking", "backups", "ipv6", "metadata", "storage"},
			},
			DropletIDs:             []int{2, 21},
			ProjectID:              "6929eef6-4e45-11ed-bdc3-0242ac120002",
			HTTPIdleTimeoutSeconds: &expectedTimeout,
			Firewall: &LBFirewall{
				Allow: []string{"ip:1.2.3.4"},
				Deny:  []string{"cidr:1.2.0.0/16"},
			},
		},
	}
	disableLetsEncryptDNSRecords := true
	expectedLBs[0].DisableLetsEncryptDNSRecords = &disableLetsEncryptDNSRecords

	assert.Equal(t, expectedLBs, loadBalancers)

	expectedMeta := &Meta{Total: 3}
	assert.Equal(t, expectedMeta, resp.Meta)
}

func TestLoadBalancers_List_Pagination(t *testing.T) {
	setup()
	defer teardown()

	path := "/v2/load_balancers"
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		testFormValues(t, r, map[string]string{"page": "2"})
		fmt.Fprint(w, lbListJSONResponse)
	})

	opts := &ListOptions{Page: 2}
	_, resp, err := client.LoadBalancers.List(ctx, opts)

	require.NoError(t, err)

	assert.Equal(t, "http://localhost:3001/v2/load_balancers?page=2&per_page=1", resp.Links.Pages.Next)
	assert.Equal(t, "http://localhost:3001/v2/load_balancers?page=3&per_page=1", resp.Links.Pages.Last)
}

func TestLoadBalancers_Delete(t *testing.T) {
	setup()
	defer teardown()

	lbID := "37e6be88-01ec-4ec7-9bc6-a514d4719057"
	path := "/v2/load_balancers"
	path = fmt.Sprintf("%s/%s", path, lbID)
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
	})

	_, err := client.LoadBalancers.Delete(ctx, lbID)

	assert.NoError(t, err)
}

func TestLoadBalancers_AddDroplets(t *testing.T) {
	setup()
	defer teardown()

	dropletIdsRequest := &dropletIDsRequest{
		IDs: []int{42, 44},
	}

	lbID := "37e6be88-01ec-4ec7-9bc6-a514d4719057"
	path := fmt.Sprintf("/v2/load_balancers/%s/droplets", lbID)
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		v := new(dropletIDsRequest)
		err := json.NewDecoder(r.Body).Decode(v)
		if err != nil {
			t.Fatal(err)
		}

		testMethod(t, r, http.MethodPost)
		assert.Equal(t, dropletIdsRequest, v)

		fmt.Fprint(w, nil)
	})

	_, err := client.LoadBalancers.AddDroplets(ctx, lbID, dropletIdsRequest.IDs...)

	assert.NoError(t, err)
}

func TestLoadBalancers_RemoveDroplets(t *testing.T) {
	setup()
	defer teardown()

	dropletIdsRequest := &dropletIDsRequest{
		IDs: []int{2, 21},
	}

	lbID := "37e6be88-01ec-4ec7-9bc6-a514d4719057"
	path := fmt.Sprintf("/v2/load_balancers/%s/droplets", lbID)
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		v := new(dropletIDsRequest)
		err := json.NewDecoder(r.Body).Decode(v)
		if err != nil {
			t.Fatal(err)
		}

		testMethod(t, r, http.MethodDelete)
		assert.Equal(t, dropletIdsRequest, v)

		fmt.Fprint(w, nil)
	})

	_, err := client.LoadBalancers.RemoveDroplets(ctx, lbID, dropletIdsRequest.IDs...)

	assert.NoError(t, err)
}

func TestLoadBalancers_AddForwardingRules(t *testing.T) {
	setup()
	defer teardown()

	frr := &forwardingRulesRequest{
		Rules: []ForwardingRule{
			{
				EntryProtocol:  "https",
				EntryPort:      444,
				TargetProtocol: "http",
				TargetPort:     81,
				CertificateID:  "b2abc00f-d3c4-426c-9f0b-b2f7a3ff7527",
			},
			{
				EntryProtocol:  "tcp",
				EntryPort:      8080,
				TargetProtocol: "tcp",
				TargetPort:     8081,
			},
		},
	}

	lbID := "37e6be88-01ec-4ec7-9bc6-a514d4719057"
	path := fmt.Sprintf("/v2/load_balancers/%s/forwarding_rules", lbID)
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		v := new(forwardingRulesRequest)
		err := json.NewDecoder(r.Body).Decode(v)
		if err != nil {
			t.Fatal(err)
		}

		testMethod(t, r, http.MethodPost)
		assert.Equal(t, frr, v)

		fmt.Fprint(w, nil)
	})

	_, err := client.LoadBalancers.AddForwardingRules(ctx, lbID, frr.Rules...)

	assert.NoError(t, err)
}

func TestLoadBalancers_RemoveForwardingRules(t *testing.T) {
	setup()
	defer teardown()

	frr := &forwardingRulesRequest{
		Rules: []ForwardingRule{
			{
				EntryProtocol:  "https",
				EntryPort:      444,
				TargetProtocol: "http",
				TargetPort:     81,
			},
			{
				EntryProtocol:  "tcp",
				EntryPort:      8080,
				TargetProtocol: "tcp",
				TargetPort:     8081,
			},
		},
	}

	lbID := "37e6be88-01ec-4ec7-9bc6-a514d4719057"
	path := fmt.Sprintf("/v2/load_balancers/%s/forwarding_rules", lbID)
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		v := new(forwardingRulesRequest)
		err := json.NewDecoder(r.Body).Decode(v)
		if err != nil {
			t.Fatal(err)
		}

		testMethod(t, r, http.MethodDelete)
		assert.Equal(t, frr, v)

		fmt.Fprint(w, nil)
	})

	_, err := client.LoadBalancers.RemoveForwardingRules(ctx, lbID, frr.Rules...)

	assert.NoError(t, err)
}

func TestLoadBalancers_AsRequest(t *testing.T) {
	lbIdleTimeout := uint64(60)
	lb := &LoadBalancer{
		ID:        "37e6be88-01ec-4ec7-9bc6-a514d4719057",
		Name:      "test-loadbalancer",
		IP:        "10.0.0.1",
		SizeSlug:  "lb-small",
		Algorithm: "least_connections",
		Status:    "active",
		Created:   "2011-06-24T12:00:00Z",
		HealthCheck: &HealthCheck{
			Protocol:               "http",
			Port:                   80,
			Path:                   "/ping",
			CheckIntervalSeconds:   30,
			ResponseTimeoutSeconds: 10,
			HealthyThreshold:       3,
			UnhealthyThreshold:     3,
		},
		StickySessions: &StickySessions{
			Type:             "cookies",
			CookieName:       "nomnom",
			CookieTtlSeconds: 32,
		},
		Region: &Region{
			Slug: "lon1",
		},
		RedirectHttpToHttps:    true,
		EnableProxyProtocol:    true,
		EnableBackendKeepalive: true,
		VPCUUID:                "880b7f98-f062-404d-b33c-458d545696f6",
		ProjectID:              "6929eef6-4e45-11ed-bdc3-0242ac120002",
		ValidateOnly:           true,
		HTTPIdleTimeoutSeconds: &lbIdleTimeout,
		Firewall: &LBFirewall{
			Allow: []string{"ip:1.2.3.5"},
			Deny:  []string{"cidr:1.3.0.0/16"},
		},
	}

	lb.DropletIDs = make([]int, 1, 2)
	lb.DropletIDs[0] = 12345
	lb.ForwardingRules = make([]ForwardingRule, 1, 2)
	lb.ForwardingRules[0] = ForwardingRule{
		EntryProtocol:  "http",
		EntryPort:      80,
		TargetProtocol: "http",
		TargetPort:     80,
	}

	want := &LoadBalancerRequest{
		Name:      "test-loadbalancer",
		Algorithm: "least_connections",
		Region:    "lon1",
		SizeSlug:  "lb-small",
		ForwardingRules: []ForwardingRule{{
			EntryProtocol:  "http",
			EntryPort:      80,
			TargetProtocol: "http",
			TargetPort:     80,
		}},
		HealthCheck: &HealthCheck{
			Protocol:               "http",
			Port:                   80,
			Path:                   "/ping",
			CheckIntervalSeconds:   30,
			ResponseTimeoutSeconds: 10,
			HealthyThreshold:       3,
			UnhealthyThreshold:     3,
		},
		StickySessions: &StickySessions{
			Type:             "cookies",
			CookieName:       "nomnom",
			CookieTtlSeconds: 32,
		},
		DropletIDs:             []int{12345},
		RedirectHttpToHttps:    true,
		EnableProxyProtocol:    true,
		EnableBackendKeepalive: true,
		VPCUUID:                "880b7f98-f062-404d-b33c-458d545696f6",
		ProjectID:              "6929eef6-4e45-11ed-bdc3-0242ac120002",
		HTTPIdleTimeoutSeconds: &lbIdleTimeout,
		ValidateOnly:           true,
		Firewall: &LBFirewall{
			Allow: []string{"ip:1.2.3.5"},
			Deny:  []string{"cidr:1.3.0.0/16"},
		},
	}

	r := lb.AsRequest()
	assert.Equal(t, want, r)
	assert.False(t, r.HealthCheck == lb.HealthCheck, "HealthCheck points to same struct")
	assert.False(t, r.StickySessions == lb.StickySessions, "StickySessions points to same struct")

	r.DropletIDs = append(r.DropletIDs, 54321)
	r.ForwardingRules = append(r.ForwardingRules, ForwardingRule{
		EntryProtocol:  "https",
		EntryPort:      443,
		TargetProtocol: "https",
		TargetPort:     443,
		TlsPassthrough: true,
	})

	// Check that original LoadBalancer hasn't changed
	lb.DropletIDs = append(lb.DropletIDs, 13579)
	lb.ForwardingRules = append(lb.ForwardingRules, ForwardingRule{
		EntryProtocol:  "tcp",
		EntryPort:      587,
		TargetProtocol: "tcp",
		TargetPort:     587,
	})
	assert.Equal(t, []int{12345, 54321}, r.DropletIDs)
	assert.Equal(t, []ForwardingRule{
		{
			EntryProtocol:  "http",
			EntryPort:      80,
			TargetProtocol: "http",
			TargetPort:     80,
		},
		{
			EntryProtocol:  "https",
			EntryPort:      443,
			TargetProtocol: "https",
			TargetPort:     443,
			TlsPassthrough: true,
		},
	}, r.ForwardingRules)
}
