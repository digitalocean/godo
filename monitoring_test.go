package godo

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/digitalocean/godo/metrics"
	"github.com/stretchr/testify/assert"
)

var (
	listEmptyPoliciesJSON = `
	{
		"policies": [
		],
		"meta": {
			"total": 0
		}
	}
	`

	listPoliciesJSON = `
	{
		"policies": [
		{
		  "uuid": "669befc9-3cbc-45fc-85f0-2c966f133730",
		  "type": "v1/insights/droplet/cpu",
		  "description": "description of policy",
		  "compare": "LessThan",
		  "value": 75,
		  "window": "5m",
		  "entities": [],
		  "tags": [
			"test-tag"
		  ],
		  "alerts": {
			"slack": [
			  {
				"url": "https://hooks.slack.com/services/T1234567/AAAAAAAAA/ZZZZZZ",
				"channel": "#alerts-test"
			  }
			],
			"email": ["bob@example.com"]
		  },
		  "enabled": true
		},
		{
		  "uuid": "777befc9-3cbc-45fc-85f0-2c966f133737",
		  "type": "v1/insights/droplet/cpu",
		  "description": "description of policy #2",
		  "compare": "LessThan",
		  "value": 90,
		  "window": "5m",
		  "entities": [],
		  "tags": [
			"test-tag-2"
		  ],
		  "alerts": {
			"slack": [
			  {
				"url": "https://hooks.slack.com/services/T1234567/AAAAAAAAA/ZZZZZZ",
				"channel": "#alerts-test"
			  }
			],
			"email": ["bob@example.com", "alice@example.com"]
		  },
		  "enabled": false
		}
		],
		"links": {
			"pages":{
				"next":"http://example.com/v2/monitoring/alerts/?page=3",
				"prev":"http://example.com/v2/monitoring/alerts/?page=1",
				"last":"http://example.com/v2/monitoring/alerts/?page=3",
				"first":"http://example.com/v2/monitoring/alerts/?page=1"
			}
		},
		"meta": {
			"total": 2
		}
	}
	`

	createAlertPolicyJSON = `
	{
		"policy": {
          "uuid": "669befc9-3cbc-45fc-85f0-2c966f133730",
		  "alerts": {
			"email": [
			  "bob@example.com"
			],
			"slack": [
			  {
				"channel": "#alerts-test",
				"url": "https://hooks.slack.com/services/T1234567/AAAAAAAA/ZZZZZZ"
			  }
			]
		  },
		  "compare": "LessThan",
		  "description": "description of policy",
		  "enabled": true,
		  "entities": [
		  ],
		  "tags": [
			"test-tag"
		  ],
		  "type": "v1/insights/droplet/cpu",
		  "value": 75,
		  "window": "5m"
		}
	}
	`

	updateAlertPolicyJSON = `
	{
		"policy": {
          "uuid": "769befc9-3cbc-45fc-85f0-2c966f133730",
		  "alerts": {
			"email": [
			  "bob@example.com"
			],
			"slack": [
			  {
				"channel": "#alerts-test",
				"url": "https://hooks.slack.com/services/T1234567/AAAAAAAA/ZZZZZZ"
			  }
			]
		  },
		  "compare": "GreaterThan",
		  "description": "description of updated policy",
		  "enabled": true,
		  "entities": [
		  ],
		  "tags": [
			"test-tag"
		  ],
		  "type": "v1/insights/droplet/cpu",
		  "value": 75,
		  "window": "5m"
		}
	}
	`

	getPolicyJSON = `
	{
		"policy": {
          "uuid": "669befc9-3cbc-45fc-85f0-2c966f133730",
		  "alerts": {
			"email": [
			  "bob@example.com"
			],
			"slack": [
			  {
				"channel": "#alerts-test",
				"url": "https://hooks.slack.com/services/T1234567/AAAAAAAA/ZZZZZZ"
			  }
			]
		  },
		  "compare": "LessThan",
		  "description": "description of policy",
		  "enabled": true,
		  "entities": [
		  ],
		  "tags": [
			"test-tag"
		  ],
		  "type": "v1/insights/droplet/cpu",
		  "value": 75,
		  "window": "5m"
		}
	}
	`

	bandwidthRespJSON = `
	{
		"status": "success",
		"data": {
			"resultType": "matrix",
			"result": [
				{
					"metric": {
						"direction": "inbound",
						"host_id": "222651441",
						"interface": "private"
					},
					"values": [
						[
							1634052360,
							"0.016600450090265357"
						],
						[
							1634052480,
							"0.015085955677299055"
						],
						[
							1634052600,
							"0.014941163855322308"
						],
						[
							1634052720,
							"0.016214285714285712"
						]
					]
				}
			]
		}
	}`

	memoryRespJSON = `
	{
		"status": "success",
		"data": {
			"resultType": "matrix",
			"result": [
			{
				"metric": {
				"host_id": "123"
				},
				"values": [
				[
					1635386880,
					"1028956160"
				],
				[
					1635387000,
					"1028956160"
				],
				[
					1635387120,
					"1028956160"
				]
				]
			}
			]
		}
	}`

	filesystemRespJSON = `
			{
		"status": "success",
		"data": {
			"resultType": "matrix",
			"result": [
			{
				"metric": {
					"device": "/dev/vda1",
					"fstype": "ext4",
					"host_id": "123",
					"mountpoint": "/"
				},
				"values": [
					[
						1635386880,
						"25832407040"
					],
					[
						1635387000,
						"25832407040"
					],
					[
						1635387120,
						"25832407040"
					]
				]
			}
			]
		}
	}`

	loadRespJSON = `
	{
		"status": "success",
		"data": {
			"resultType": "matrix",
			"result": [
			{
				"metric": {
				"host_id": "123"
				},
				"values": [
				[
					1635386880,
					"0.04"
				],
				[
					1635387000,
					"0.03"
				],
				[
					1635387120,
					"0.01"
				]
				]
			}
			]
		}
	}`

	cpuRespJSON = `
	{
		"status": "success",
		"data": {
			"resultType": "matrix",
			"result": [
			{
				"metric": {
				"host_id": "123",
				"mode": "idle"
				},
				"values": [
				[
					1635386880,
					"122901.18"
				],
				[
					1635387000,
					"123020.92"
				],
				[
					1635387120,
					"123140.8"
				]
				]
			},
			{
				"metric": {
				"host_id": "123",
				"mode": "iowait"
				},
				"values": [
				[
					1635386880,
					"14.99"
				],
				[
					1635387000,
					"15.01"
				],
				[
					1635387120,
					"15.01"
				]
				]
			},
			{
				"metric": {
				"host_id": "123",
				"mode": "irq"
				},
				"values": [
				[
					1635386880,
					"0"
				],
				[
					1635387000,
					"0"
				],
				[
					1635387120,
					"0"
				]
				]
			},
			{
				"metric": {
				"host_id": "123",
				"mode": "nice"
				},
				"values": [
				[
					1635386880,
					"66.35"
				],
				[
					1635387000,
					"66.35"
				],
				[
					1635387120,
					"66.35"
				]
				]
			},
			{
				"metric": {
				"host_id": "123",
				"mode": "softirq"
				},
				"values": [
				[
					1635386880,
					"2.13"
				],
				[
					1635387000,
					"2.13"
				],
				[
					1635387120,
					"2.13"
				]
				]
			},
			{
				"metric": {
				"host_id": "123",
				"mode": "steal"
				},
				"values": [
				[
					1635386880,
					"7.89"
				],
				[
					1635387000,
					"7.9"
				],
				[
					1635387120,
					"7.91"
				]
				]
			},
			{
				"metric": {
				"host_id": "123",
				"mode": "system"
				},
				"values": [
				[
					1635386880,
					"140.09"
				],
				[
					1635387000,
					"140.2"
				],
				[
					1635387120,
					"140.23"
				]
				]
			},
			{
				"metric": {
				"host_id": "123",
				"mode": "user"
				},
				"values": [
				[
					1635386880,
					"278.57"
				],
				[
					1635387000,
					"278.65"
				],
				[
					1635387120,
					"278.69"
				]
				]
			}
			]
		}
	}`

	testCPUResponse = &MetricsResponse{
		Status: "success",
		Data: MetricsData{
			ResultType: "matrix",
			Result: []metrics.SampleStream{
				{
					Metric: metrics.Metric{
						"host_id": "123",
						"mode":    "idle",
					},
					Values: []metrics.SamplePair{
						{
							Timestamp: 1635386880000,
							Value:     122901.18,
						},
						{
							Timestamp: 1635387000000,
							Value:     123020.92,
						},
						{
							Timestamp: 1635387120000,
							Value:     123140.8,
						},
					},
				},
				{
					Metric: metrics.Metric{
						"host_id": "123",
						"mode":    "iowait",
					},
					Values: []metrics.SamplePair{
						{
							Timestamp: 1635386880000,
							Value:     14.99,
						},
						{
							Timestamp: 1635387000000,
							Value:     15.01,
						},
						{
							Timestamp: 1635387120000,
							Value:     15.01,
						},
					},
				},
				{
					Metric: metrics.Metric{
						"host_id": "123",
						"mode":    "irq",
					},
					Values: []metrics.SamplePair{
						{
							Timestamp: 1635386880000,
							Value:     0,
						},
						{
							Timestamp: 1635387000000,
							Value:     0,
						},
						{
							Timestamp: 1635387120000,
							Value:     0,
						},
					},
				},
				{
					Metric: metrics.Metric{
						"host_id": "123",
						"mode":    "nice",
					},
					Values: []metrics.SamplePair{
						{
							Timestamp: 1635386880000,
							Value:     66.35,
						},
						{
							Timestamp: 1635387000000,
							Value:     66.35,
						},
						{
							Timestamp: 1635387120000,
							Value:     66.35,
						},
					},
				},
				{
					Metric: metrics.Metric{
						"host_id": "123",
						"mode":    "softirq",
					},
					Values: []metrics.SamplePair{
						{
							Timestamp: 1635386880000,
							Value:     2.13,
						},
						{
							Timestamp: 1635387000000,
							Value:     2.13,
						},
						{
							Timestamp: 1635387120000,
							Value:     2.13,
						},
					},
				},
				{
					Metric: metrics.Metric{
						"host_id": "123",
						"mode":    "steal",
					},
					Values: []metrics.SamplePair{
						{
							Timestamp: 1635386880000,
							Value:     7.89,
						},
						{
							Timestamp: 1635387000000,
							Value:     7.9,
						},
						{
							Timestamp: 1635387120000,
							Value:     7.91,
						},
					},
				},
				{
					Metric: metrics.Metric{
						"host_id": "123",
						"mode":    "system",
					},
					Values: []metrics.SamplePair{
						{
							Timestamp: 1635386880000,
							Value:     140.09,
						},
						{
							Timestamp: 1635387000000,
							Value:     140.2,
						},
						{
							Timestamp: 1635387120000,
							Value:     140.23,
						},
					},
				},
				{
					Metric: metrics.Metric{
						"host_id": "123",
						"mode":    "user",
					},
					Values: []metrics.SamplePair{
						{
							Timestamp: 1635386880000,
							Value:     278.57,
						},
						{
							Timestamp: 1635387000000,
							Value:     278.65,
						},
						{
							Timestamp: 1635387120000,
							Value:     278.69,
						},
					},
				},
			},
		},
	}

	testLoadResponse = &MetricsResponse{
		Status: "success",
		Data: MetricsData{
			ResultType: "matrix",
			Result: []metrics.SampleStream{
				{
					Metric: metrics.Metric{
						"host_id": "123",
					},
					Values: []metrics.SamplePair{
						{
							Timestamp: 1635386880000,
							Value:     0.04,
						},
						{
							Timestamp: 1635387000000,
							Value:     0.03,
						},
						{
							Timestamp: 1635387120000,
							Value:     0.01,
						},
					},
				},
			},
		},
	}

	testFilesystemResponse = &MetricsResponse{
		Status: "success",
		Data: MetricsData{
			ResultType: "matrix",
			Result: []metrics.SampleStream{
				{
					Metric: metrics.Metric{
						"device":     "/dev/vda1",
						"fstype":     "ext4",
						"host_id":    "123",
						"mountpoint": "/",
					},
					Values: []metrics.SamplePair{
						{
							Timestamp: 1635386880000,
							Value:     25832407040,
						},
						{
							Timestamp: 1635387000000,
							Value:     25832407040,
						},
						{
							Timestamp: 1635387120000,
							Value:     25832407040,
						},
					},
				},
			},
		},
	}

	testMemoryResponse = &MetricsResponse{
		Status: "success",
		Data: MetricsData{
			ResultType: "matrix",
			Result: []metrics.SampleStream{
				{
					Metric: metrics.Metric{
						"host_id": "123",
					},
					Values: []metrics.SamplePair{
						{
							Timestamp: 1635386880000,
							Value:     1.02895616e+09,
						},
						{
							Timestamp: 1635387000000,
							Value:     1.02895616e+09,
						},
						{
							Timestamp: 1635387120000,
							Value:     1.02895616e+09,
						},
					},
				},
			},
		},
	}
)

func TestAlertPolicies_List(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/monitoring/alerts", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, listPoliciesJSON)
	})

	policies, resp, err := client.Monitoring.ListAlertPolicies(ctx, nil)
	if err != nil {
		t.Errorf("Monitoring.ListAlertPolicies returned error: %v", err)
	}

	expectedPolicies := []AlertPolicy{
		{UUID: "669befc9-3cbc-45fc-85f0-2c966f133730", Type: DropletCPUUtilizationPercent, Description: "description of policy", Compare: "LessThan", Value: 75, Window: "5m", Entities: []string{}, Tags: []string{"test-tag"}, Alerts: Alerts{Slack: []SlackDetails{{URL: "https://hooks.slack.com/services/T1234567/AAAAAAAAA/ZZZZZZ", Channel: "#alerts-test"}}, Email: []string{"bob@example.com"}}, Enabled: true},
		{UUID: "777befc9-3cbc-45fc-85f0-2c966f133737", Type: DropletCPUUtilizationPercent, Description: "description of policy #2", Compare: "LessThan", Value: 90, Window: "5m", Entities: []string{}, Tags: []string{"test-tag-2"}, Alerts: Alerts{Slack: []SlackDetails{{URL: "https://hooks.slack.com/services/T1234567/AAAAAAAAA/ZZZZZZ", Channel: "#alerts-test"}}, Email: []string{"bob@example.com", "alice@example.com"}}, Enabled: false},
	}
	if !reflect.DeepEqual(policies, expectedPolicies) {
		t.Errorf("Monitoring.ListAlertPolicies returned policies %+v, expected %+v", policies, expectedPolicies)
	}

	expectedMeta := &Meta{Total: 2}
	if !reflect.DeepEqual(resp.Meta, expectedMeta) {
		t.Errorf("Monitoring.ListAlertPolicies returned meta %+v, expected %+v", resp.Meta, expectedMeta)
	}
}

func TestAlertPolicies_ListEmpty(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/monitoring/alerts", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, listEmptyPoliciesJSON)
	})

	policies, _, err := client.Monitoring.ListAlertPolicies(ctx, nil)
	if err != nil {
		t.Errorf("Monitoring.ListAlertPolicies returned error: %v", err)
	}

	expected := []AlertPolicy{}
	if !reflect.DeepEqual(policies, expected) {
		t.Errorf("Monitoring.ListAlertPolicies returned %+v, expected %+v", policies, expected)
	}
}

func TestAlertPolicies_ListPaging(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/monitoring/alerts", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, listPoliciesJSON)
	})

	_, resp, err := client.Monitoring.ListAlertPolicies(ctx, nil)
	if err != nil {
		t.Errorf("Monitoring.ListAlertPolicies returned error: %v", err)
	}
	checkCurrentPage(t, resp, 2)
}

func TestAlertPolicy_Get(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/monitoring/alerts/669befc9-3cbc-45fc-85f0-2c966f133730", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, getPolicyJSON)
	})

	policy, _, err := client.Monitoring.GetAlertPolicy(ctx, "669befc9-3cbc-45fc-85f0-2c966f133730")
	if err != nil {
		t.Errorf("Monitoring.GetAlertPolicy returned error: %v", err)
	}
	expected := &AlertPolicy{UUID: "669befc9-3cbc-45fc-85f0-2c966f133730", Type: DropletCPUUtilizationPercent, Description: "description of policy", Compare: "LessThan", Value: 75, Window: "5m", Entities: []string{}, Tags: []string{"test-tag"}, Alerts: Alerts{Slack: []SlackDetails{{URL: "https://hooks.slack.com/services/T1234567/AAAAAAAA/ZZZZZZ", Channel: "#alerts-test"}}, Email: []string{"bob@example.com"}}, Enabled: true}
	if !reflect.DeepEqual(policy, expected) {
		t.Errorf("Monitoring.CreateAlertPolicy returned %+v, expected %+v", policy, expected)
	}
}

func TestAlertPolicy_Create(t *testing.T) {
	setup()
	defer teardown()

	createRequest := &AlertPolicyCreateRequest{
		Type:        DropletCPUUtilizationPercent,
		Description: "description of policy",
		Compare:     "LessThan",
		Value:       75,
		Window:      "5m",
		Entities:    []string{},
		Tags:        []string{"test-tag"},
		Alerts: Alerts{
			Email: []string{"bob@example.com"},
			Slack: []SlackDetails{
				{
					Channel: "#alerts-test",
					URL:     "https://hooks.slack.com/services/T1234567/AAAAAAAAA/ZZZZZZ",
				},
			},
		},
	}

	mux.HandleFunc("/v2/monitoring/alerts", func(w http.ResponseWriter, r *http.Request) {
		v := new(AlertPolicyCreateRequest)
		err := json.NewDecoder(r.Body).Decode(v)
		if err != nil {
			t.Fatalf("decode json: %v", err)
		}

		testMethod(t, r, http.MethodPost)
		if !reflect.DeepEqual(v, createRequest) {
			t.Errorf("Request body = %+v, expected %+v", v, createRequest)
		}

		fmt.Fprintf(w, createAlertPolicyJSON)
	})

	policy, _, err := client.Monitoring.CreateAlertPolicy(ctx, createRequest)
	if err != nil {
		t.Errorf("Monitoring.CreateAlertPolicy returned error: %v", err)
	}

	expected := &AlertPolicy{UUID: "669befc9-3cbc-45fc-85f0-2c966f133730", Type: DropletCPUUtilizationPercent, Description: "description of policy", Compare: "LessThan", Value: 75, Window: "5m", Entities: []string{}, Tags: []string{"test-tag"}, Alerts: Alerts{Slack: []SlackDetails{{URL: "https://hooks.slack.com/services/T1234567/AAAAAAAA/ZZZZZZ", Channel: "#alerts-test"}}, Email: []string{"bob@example.com"}}, Enabled: true}

	if !reflect.DeepEqual(policy, expected) {
		t.Errorf("Monitoring.CreateAlertPolicy returned %+v, expected %+v", policy, expected)
	}
}

func TestAlertPolicy_Delete(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/monitoring/alerts/669befc9-3cbc-45fc-85f0-2c966f133730", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
	})

	_, err := client.Monitoring.DeleteAlertPolicy(ctx, "669befc9-3cbc-45fc-85f0-2c966f133730")
	if err != nil {
		t.Errorf("Monitoring.DeleteAlertPolicy returned error: %v", err)
	}
}

func TestAlertPolicy_Update(t *testing.T) {
	setup()
	defer teardown()

	updateRequest := &AlertPolicyUpdateRequest{
		Type:        DropletCPUUtilizationPercent,
		Description: "description of updated policy",
		Compare:     "GreaterThan",
		Value:       75,
		Window:      "5m",
		Entities:    []string{},
		Tags:        []string{"test-tag"},
		Alerts: Alerts{
			Email: []string{"bob@example.com"},
			Slack: []SlackDetails{
				{
					Channel: "#alerts-test",
					URL:     "https://hooks.slack.com/services/T1234567/AAAAAAAAA/ZZZZZZ",
				},
			},
		},
	}

	mux.HandleFunc("/v2/monitoring/alerts/769befc9-3cbc-45fc-85f0-2c966f133730", func(w http.ResponseWriter, r *http.Request) {
		v := new(AlertPolicyUpdateRequest)
		err := json.NewDecoder(r.Body).Decode(v)
		if err != nil {
			t.Fatalf("decode json: %v", err)
		}

		testMethod(t, r, http.MethodPut)
		if !reflect.DeepEqual(v, updateRequest) {
			t.Errorf("Request body = %+v, expected %+v", v, updateRequest)
		}

		fmt.Fprintf(w, updateAlertPolicyJSON)
	})

	policy, _, err := client.Monitoring.UpdateAlertPolicy(ctx, "769befc9-3cbc-45fc-85f0-2c966f133730", updateRequest)
	if err != nil {
		t.Errorf("Monitoring.UpdateAlertPolicy returned error: %v", err)
	}

	expected := &AlertPolicy{UUID: "769befc9-3cbc-45fc-85f0-2c966f133730", Type: DropletCPUUtilizationPercent, Description: "description of updated policy", Compare: "GreaterThan", Value: 75, Window: "5m", Entities: []string{}, Tags: []string{"test-tag"}, Alerts: Alerts{Slack: []SlackDetails{{URL: "https://hooks.slack.com/services/T1234567/AAAAAAAA/ZZZZZZ", Channel: "#alerts-test"}}, Email: []string{"bob@example.com"}}, Enabled: true}

	if !reflect.DeepEqual(policy, expected) {
		t.Errorf("Monitoring.UpdateAlertPolicy returned %+v, expected %+v", policy, expected)
	}
}

func TestGetDropletBandwidth(t *testing.T) {
	setup()
	defer teardown()
	now := time.Now()
	metricReq := &DropletBandwidthMetricsRequest{
		DropletMetricsRequest: DropletMetricsRequest{HostID: "123",
			Start: now.Add(-300 * time.Second),
			End:   now,
		},
		Interface: "private",
		Direction: "inbound",
	}

	mux.HandleFunc("/v2/monitoring/metrics/droplet/bandwidth", func(w http.ResponseWriter, r *http.Request) {
		hostID := r.URL.Query().Get("host_id")
		inter := r.URL.Query().Get("interface")
		direction := r.URL.Query().Get("direction")
		start := r.URL.Query().Get("start")
		end := r.URL.Query().Get("end")

		assert.Equal(t, metricReq.HostID, hostID)
		assert.Equal(t, metricReq.Interface, inter)
		assert.Equal(t, metricReq.Direction, direction)
		assert.Equal(t, fmt.Sprintf("%d", metricReq.Start.Unix()), start)
		assert.Equal(t, fmt.Sprintf("%d", metricReq.End.Unix()), end)
		testMethod(t, r, http.MethodGet)

		fmt.Fprintf(w, bandwidthRespJSON)
	})

	metricsResp, _, err := client.Monitoring.GetDropletBandwidth(ctx, metricReq)
	if err != nil {
		t.Errorf("Monitoring.GetDropletBandwidthMetrics returned error: %v", err)
	}

	expected := &MetricsResponse{
		Status: "success",
		Data: MetricsData{
			ResultType: "matrix",
			Result: []metrics.SampleStream{
				{
					Metric: metrics.Metric{
						"host_id":   "222651441",
						"direction": "inbound",
						"interface": "private",
					},
					Values: []metrics.SamplePair{
						{
							Timestamp: 1634052360000,
							Value:     0.016600450090265357,
						},
						{
							Timestamp: 1634052480000,
							Value:     0.015085955677299055,
						},
						{
							Timestamp: 1634052600000,
							Value:     0.014941163855322308,
						},
						{
							Timestamp: 1634052720000,
							Value:     0.016214285714285712,
						},
					},
				},
			},
		},
	}

	assert.Equal(t, expected, metricsResp)
}

func TestGetDropletTotalMemory(t *testing.T) {
	setup()
	defer teardown()
	now := time.Now()
	metricReq := &DropletMetricsRequest{
		HostID: "123",
		Start:  now.Add(-300 * time.Second),
		End:    now,
	}

	mux.HandleFunc("/v2/monitoring/metrics/droplet/memory_total", func(w http.ResponseWriter, r *http.Request) {
		hostID := r.URL.Query().Get("host_id")
		start := r.URL.Query().Get("start")
		end := r.URL.Query().Get("end")

		assert.Equal(t, metricReq.HostID, hostID)
		assert.Equal(t, fmt.Sprintf("%d", metricReq.Start.Unix()), start)
		assert.Equal(t, fmt.Sprintf("%d", metricReq.End.Unix()), end)
		testMethod(t, r, http.MethodGet)

		fmt.Fprintf(w, memoryRespJSON)
	})

	metricsResp, _, err := client.Monitoring.GetDropletTotalMemory(ctx, metricReq)
	if err != nil {
		t.Errorf("Monitoring.GetDropletTotalMemory returned error: %v", err)
	}

	assert.Equal(t, testMemoryResponse, metricsResp)
}

func TestGetDropletFreeMemory(t *testing.T) {
	setup()
	defer teardown()
	now := time.Now()
	metricReq := &DropletMetricsRequest{
		HostID: "123",
		Start:  now.Add(-300 * time.Second),
		End:    now,
	}

	mux.HandleFunc("/v2/monitoring/metrics/droplet/memory_free", func(w http.ResponseWriter, r *http.Request) {
		hostID := r.URL.Query().Get("host_id")
		start := r.URL.Query().Get("start")
		end := r.URL.Query().Get("end")

		assert.Equal(t, metricReq.HostID, hostID)
		assert.Equal(t, fmt.Sprintf("%d", metricReq.Start.Unix()), start)
		assert.Equal(t, fmt.Sprintf("%d", metricReq.End.Unix()), end)
		testMethod(t, r, http.MethodGet)

		fmt.Fprintf(w, memoryRespJSON)
	})

	metricsResp, _, err := client.Monitoring.GetDropletFreeMemory(ctx, metricReq)
	if err != nil {
		t.Errorf("Monitoring.GetDropletFreeMemory returned error: %v", err)
	}

	assert.Equal(t, testMemoryResponse, metricsResp)
}

func TestGetDropletAvailableMemory(t *testing.T) {
	setup()
	defer teardown()
	now := time.Now()
	metricReq := &DropletMetricsRequest{
		HostID: "123",
		Start:  now.Add(-300 * time.Second),
		End:    now,
	}

	mux.HandleFunc("/v2/monitoring/metrics/droplet/memory_available", func(w http.ResponseWriter, r *http.Request) {
		hostID := r.URL.Query().Get("host_id")
		start := r.URL.Query().Get("start")
		end := r.URL.Query().Get("end")

		assert.Equal(t, metricReq.HostID, hostID)
		assert.Equal(t, fmt.Sprintf("%d", metricReq.Start.Unix()), start)
		assert.Equal(t, fmt.Sprintf("%d", metricReq.End.Unix()), end)
		testMethod(t, r, http.MethodGet)

		fmt.Fprintf(w, memoryRespJSON)
	})

	metricsResp, _, err := client.Monitoring.GetDropletAvailableMemory(ctx, metricReq)
	if err != nil {
		t.Errorf("Monitoring.GetDropletAvailableMemory returned error: %v", err)
	}

	assert.Equal(t, testMemoryResponse, metricsResp)
}

func TestGetDropletCachedMemory(t *testing.T) {
	setup()
	defer teardown()
	now := time.Now()
	metricReq := &DropletMetricsRequest{
		HostID: "123",
		Start:  now.Add(-300 * time.Second),
		End:    now,
	}

	mux.HandleFunc("/v2/monitoring/metrics/droplet/memory_cached", func(w http.ResponseWriter, r *http.Request) {
		hostID := r.URL.Query().Get("host_id")
		start := r.URL.Query().Get("start")
		end := r.URL.Query().Get("end")

		assert.Equal(t, metricReq.HostID, hostID)
		assert.Equal(t, fmt.Sprintf("%d", metricReq.Start.Unix()), start)
		assert.Equal(t, fmt.Sprintf("%d", metricReq.End.Unix()), end)
		testMethod(t, r, http.MethodGet)

		fmt.Fprintf(w, memoryRespJSON)
	})

	metricsResp, _, err := client.Monitoring.GetDropletCachedMemory(ctx, metricReq)
	if err != nil {
		t.Errorf("Monitoring.GetDropletCachedMemory returned error: %v", err)
	}

	assert.Equal(t, testMemoryResponse, metricsResp)
}

func TestGetDropletFilesystemFree(t *testing.T) {
	setup()
	defer teardown()
	now := time.Now()
	metricReq := &DropletMetricsRequest{
		HostID: "123",
		Start:  now.Add(-300 * time.Second),
		End:    now,
	}

	mux.HandleFunc("/v2/monitoring/metrics/droplet/filesystem_free", func(w http.ResponseWriter, r *http.Request) {
		hostID := r.URL.Query().Get("host_id")
		start := r.URL.Query().Get("start")
		end := r.URL.Query().Get("end")

		assert.Equal(t, metricReq.HostID, hostID)
		assert.Equal(t, fmt.Sprintf("%d", metricReq.Start.Unix()), start)
		assert.Equal(t, fmt.Sprintf("%d", metricReq.End.Unix()), end)
		testMethod(t, r, http.MethodGet)

		fmt.Fprintf(w, filesystemRespJSON)
	})

	metricsResp, _, err := client.Monitoring.GetDropletFilesystemFree(ctx, metricReq)
	if err != nil {
		t.Errorf("Monitoring.GetDropletFilesystemFree returned error: %v", err)
	}

	assert.Equal(t, testFilesystemResponse, metricsResp)
}

func TestGetDropletFilesystemSize(t *testing.T) {
	setup()
	defer teardown()
	now := time.Now()
	metricReq := &DropletMetricsRequest{
		HostID: "123",
		Start:  now.Add(-300 * time.Second),
		End:    now,
	}

	mux.HandleFunc("/v2/monitoring/metrics/droplet/filesystem_size", func(w http.ResponseWriter, r *http.Request) {
		hostID := r.URL.Query().Get("host_id")
		start := r.URL.Query().Get("start")
		end := r.URL.Query().Get("end")

		assert.Equal(t, metricReq.HostID, hostID)
		assert.Equal(t, fmt.Sprintf("%d", metricReq.Start.Unix()), start)
		assert.Equal(t, fmt.Sprintf("%d", metricReq.End.Unix()), end)
		testMethod(t, r, http.MethodGet)

		fmt.Fprintf(w, filesystemRespJSON)
	})

	metricsResp, _, err := client.Monitoring.GetDropletFilesystemSize(ctx, metricReq)
	if err != nil {
		t.Errorf("Monitoring.GetDropletFilesystemSize returned error: %v", err)
	}

	assert.Equal(t, testFilesystemResponse, metricsResp)
}

func TestGetDropletLoad1(t *testing.T) {
	setup()
	defer teardown()
	now := time.Now()
	metricReq := &DropletMetricsRequest{
		HostID: "123",
		Start:  now.Add(-300 * time.Second),
		End:    now,
	}

	mux.HandleFunc("/v2/monitoring/metrics/droplet/load_1", func(w http.ResponseWriter, r *http.Request) {
		hostID := r.URL.Query().Get("host_id")
		start := r.URL.Query().Get("start")
		end := r.URL.Query().Get("end")

		assert.Equal(t, metricReq.HostID, hostID)
		assert.Equal(t, fmt.Sprintf("%d", metricReq.Start.Unix()), start)
		assert.Equal(t, fmt.Sprintf("%d", metricReq.End.Unix()), end)
		testMethod(t, r, http.MethodGet)

		fmt.Fprintf(w, loadRespJSON)
	})

	metricsResp, _, err := client.Monitoring.GetDropletLoad1(ctx, metricReq)
	if err != nil {
		t.Errorf("Monitoring.GetDropletLoad1 returned error: %v", err)
	}

	assert.Equal(t, testLoadResponse, metricsResp)
}

func TestGetDropletLoad5(t *testing.T) {
	setup()
	defer teardown()
	now := time.Now()
	metricReq := &DropletMetricsRequest{
		HostID: "123",
		Start:  now.Add(-300 * time.Second),
		End:    now,
	}

	mux.HandleFunc("/v2/monitoring/metrics/droplet/load_5", func(w http.ResponseWriter, r *http.Request) {
		hostID := r.URL.Query().Get("host_id")
		start := r.URL.Query().Get("start")
		end := r.URL.Query().Get("end")

		assert.Equal(t, metricReq.HostID, hostID)
		assert.Equal(t, fmt.Sprintf("%d", metricReq.Start.Unix()), start)
		assert.Equal(t, fmt.Sprintf("%d", metricReq.End.Unix()), end)
		testMethod(t, r, http.MethodGet)

		fmt.Fprintf(w, loadRespJSON)
	})

	metricsResp, _, err := client.Monitoring.GetDropletLoad5(ctx, metricReq)
	if err != nil {
		t.Errorf("Monitoring.GetDropletLoad5 returned error: %v", err)
	}

	assert.Equal(t, testLoadResponse, metricsResp)
}

func TestGetDropletLoad15(t *testing.T) {
	setup()
	defer teardown()
	now := time.Now()
	metricReq := &DropletMetricsRequest{
		HostID: "123",
		Start:  now.Add(-300 * time.Second),
		End:    now,
	}

	mux.HandleFunc("/v2/monitoring/metrics/droplet/load_15", func(w http.ResponseWriter, r *http.Request) {
		hostID := r.URL.Query().Get("host_id")
		start := r.URL.Query().Get("start")
		end := r.URL.Query().Get("end")

		assert.Equal(t, metricReq.HostID, hostID)
		assert.Equal(t, fmt.Sprintf("%d", metricReq.Start.Unix()), start)
		assert.Equal(t, fmt.Sprintf("%d", metricReq.End.Unix()), end)
		testMethod(t, r, http.MethodGet)

		fmt.Fprintf(w, loadRespJSON)
	})

	metricsResp, _, err := client.Monitoring.GetDropletLoad15(ctx, metricReq)
	if err != nil {
		t.Errorf("Monitoring.GetDropletLoad15 returned error: %v", err)
	}

	assert.Equal(t, testLoadResponse, metricsResp)
}

func TestGetDropletCPU(t *testing.T) {
	setup()
	defer teardown()
	now := time.Now()
	metricReq := &DropletMetricsRequest{
		HostID: "123",
		Start:  now.Add(-300 * time.Second),
		End:    now,
	}

	mux.HandleFunc("/v2/monitoring/metrics/droplet/cpu", func(w http.ResponseWriter, r *http.Request) {
		hostID := r.URL.Query().Get("host_id")
		start := r.URL.Query().Get("start")
		end := r.URL.Query().Get("end")

		assert.Equal(t, metricReq.HostID, hostID)
		assert.Equal(t, fmt.Sprintf("%d", metricReq.Start.Unix()), start)
		assert.Equal(t, fmt.Sprintf("%d", metricReq.End.Unix()), end)
		testMethod(t, r, http.MethodGet)

		fmt.Fprintf(w, cpuRespJSON)
	})

	metricsResp, _, err := client.Monitoring.GetDropletCPU(ctx, metricReq)
	if err != nil {
		t.Errorf("Monitoring.GetDropletCPU returned error: %v", err)
	}

	assert.Equal(t, testCPUResponse, metricsResp)
}
