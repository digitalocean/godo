package godo

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"testing"
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
