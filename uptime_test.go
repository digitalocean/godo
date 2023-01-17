package godo

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

func TestUptimeChecks_List(t *testing.T) {
	setup()
	defer teardown()

	expectedUptimeChecks := []UptimeCheck{
		{
			ID:   "uptimecheck-1",
			Name: "uptimecheck-1",
		},
		{
			ID:   "uptimecheck-2",
			Name: "uptimecheck-2",
		},
	}

	mux.HandleFunc("/v2/uptime/checks", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		resp, _ := json.Marshal(expectedUptimeChecks)
		fmt.Fprint(w, fmt.Sprintf(`{"checks":%s, "meta": {"total": 2}}`, string(resp)))
	})

	uptimeChecks, resp, err := client.UptimeChecks.List(ctx, nil)
	if err != nil {
		t.Errorf("UptimeChecks.List returned error: %v", err)
	}

	if !reflect.DeepEqual(uptimeChecks, expectedUptimeChecks) {
		t.Errorf("UptimeChecks.List returned uptime checks %+v, expected %+v", uptimeChecks, expectedUptimeChecks)
	}

	expectedMeta := &Meta{Total: 2}
	if !reflect.DeepEqual(resp.Meta, expectedMeta) {
		t.Errorf("UptimeChecks.List returned meta %+v, expected %+v", resp.Meta, expectedMeta)
	}
}

func TestUptimeChecks_ListWithMultiplePages(t *testing.T) {
	setup()
	defer teardown()

	mockResp := `
	{
		"checks": [
			{
				"uuid": "check-1",
				"name": "check-1"
			},
			{
				"uuid": "check-2",
				"name": "check-2"
			}
		],
		"links": {
			"pages": {
				"next": "http://example.com/v2/uptime/checks?page=2"
			}
		}
	}`

	mux.HandleFunc("/v2/uptime/checks", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, mockResp)
	})

	_, resp, err := client.UptimeChecks.List(ctx, nil)
	if err != nil {
		t.Errorf("UptimeChecks.List returned error: %v", err)
	}

	checkCurrentPage(t, resp, 1)
}

func TestUptimeChecks_ListWithPageNumber(t *testing.T) {
	setup()
	defer teardown()

	mockResp := `
	{
		"checks": [
			{
				"uuid": "check-1",
				"name": "check-1"
			},
			{
				"uuid": "check-2",
				"name": "check-2"
			}
		],
		"links": {
			"pages": {
				"next": "http://example.com/v2/uptime/checks?page=3",
				"prev": "http://example.com/v2/uptime/checks?page=1",
				"last": "http://example.com/v2/uptime/checks?page=3",
				"first": "http://example.com/v2/uptime/checks?page=1"
			}
		}
	}`

	mux.HandleFunc("/v2/uptime/checks", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, mockResp)
	})

	_, resp, err := client.UptimeChecks.List(ctx, &ListOptions{Page: 2})
	if err != nil {
		t.Errorf("UptimeChecks.List returned error: %v", err)
	}

	checkCurrentPage(t, resp, 2)
}

func TestUptimeChecks_GetState(t *testing.T) {
	setup()
	defer teardown()

	uptimeCheckState := &UptimeCheckState{
		Regions: map[string]UptimeRegion{
			"us_east": {
				Status:                    "UP",
				StatusChangedAt:           "2022-03-17T22:28:51Z",
				ThirtyDayUptimePercentage: 97.99,
			},
			"eu_west": {
				Status:                    "UP",
				StatusChangedAt:           "2022-03-17T22:28:51Z",
				ThirtyDayUptimePercentage: 97.99,
			},
		},
		PreviousOutage: UptimePreviousOutage{
			Region:          "us_east",
			StartedAt:       "2022-03-17T18:04:55Z",
			EndedAt:         "2022-03-17T18:06:55Z",
			DurationSeconds: 120,
		},
	}

	mux.HandleFunc("/v2/uptime/checks/check-1/state", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		resp, _ := json.Marshal(uptimeCheckState)
		fmt.Fprint(w, fmt.Sprintf(`{"state":%s}`, string(resp)))
	})

	resp, _, err := client.UptimeChecks.GetState(ctx, "check-1")
	if err != nil {
		t.Errorf("UptimeChecks.GetState returned error: %v", err)
	}

	if !reflect.DeepEqual(resp, uptimeCheckState) {
		t.Errorf("UptimeChecks.GetUptimeCheckState returned %+v, expected %+v", resp, uptimeCheckState)
	}
}

func TestUptimeChecks_GetWithID(t *testing.T) {
	setup()
	defer teardown()

	uptimeCheck := &UptimeCheck{
		ID:   "check-1",
		Name: "check-1",
	}

	mux.HandleFunc("/v2/uptime/checks/check-1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		resp, _ := json.Marshal(uptimeCheck)
		fmt.Fprint(w, fmt.Sprintf(`{"check":%s}`, string(resp)))
	})

	resp, _, err := client.UptimeChecks.Get(ctx, "check-1")
	if err != nil {
		t.Errorf("UptimeChecks.Get returned error: %v", err)
	}

	if !reflect.DeepEqual(resp, uptimeCheck) {
		t.Errorf("UptimeChecks.Get returned %+v, expected %+v", resp, uptimeCheck)
	}
}

func TestUptimeChecks_Create(t *testing.T) {
	setup()
	defer teardown()

	createRequest := &CreateUptimeCheckRequest{
		Name:    "my check",
		Type:    "https",
		Target:  "https://www.landingpage.com",
		Enabled: true,
	}

	createResp := &UptimeCheck{
		ID:      "check-id",
		Name:    createRequest.Name,
		Type:    createRequest.Type,
		Target:  createRequest.Target,
		Enabled: createRequest.Enabled,
	}

	mux.HandleFunc("/v2/uptime/checks", func(w http.ResponseWriter, r *http.Request) {
		v := new(CreateUptimeCheckRequest)
		err := json.NewDecoder(r.Body).Decode(v)
		if err != nil {
			t.Fatalf("decode json: %v", err)
		}

		testMethod(t, r, http.MethodPost)
		if !reflect.DeepEqual(v, createRequest) {
			t.Errorf("Request body = %+v, expected %+v", v, createRequest)
		}

		resp, _ := json.Marshal(createResp)
		fmt.Fprintf(w, fmt.Sprintf(`{"check":%s}`, string(resp)))
	})

	uptimeCheck, _, err := client.UptimeChecks.Create(ctx, createRequest)
	if err != nil {
		t.Errorf("UptimeChecks.Create returned error: %v", err)
	}

	if !reflect.DeepEqual(uptimeCheck, createResp) {
		t.Errorf("UptimeChecks.Create returned %+v, expected %+v", uptimeCheck, createResp)
	}
}

func TestUptimeChecks_Update(t *testing.T) {
	setup()
	defer teardown()

	updateRequest := &UpdateUptimeCheckRequest{
		Name:    "my check",
		Type:    "https",
		Target:  "https://www.landingpage.com",
		Enabled: true,
	}
	updateResp := &UptimeCheck{
		ID:      "check-id",
		Name:    updateRequest.Name,
		Type:    updateRequest.Type,
		Target:  updateRequest.Target,
		Enabled: updateRequest.Enabled,
	}

	mux.HandleFunc("/v2/uptime/checks/check-id", func(w http.ResponseWriter, r *http.Request) {
		reqBytes, respErr := ioutil.ReadAll(r.Body)
		if respErr != nil {
			t.Error("uptime checks mock didn't work")
		}

		req := strings.TrimSuffix(string(reqBytes), "\n")
		expectedReq := `{"name":"my check","type":"https","target":"https://www.landingpage.com","regions":null,"enabled":true}`
		if req != expectedReq {
			t.Errorf("check req didn't match up:\n expected %+v\n got %+v\n", expectedReq, req)
		}

		resp, _ := json.Marshal(updateResp)
		fmt.Fprintf(w, fmt.Sprintf(`{"check":%s}`, string(resp)))
	})

	uptimeCheck, _, err := client.UptimeChecks.Update(ctx, "check-id", updateRequest)
	if err != nil {
		t.Errorf("UptimeChecks.Update returned error: %v", err)
	}
	if !reflect.DeepEqual(uptimeCheck, updateResp) {
		t.Errorf("UptimeChecks.Update returned %+v, expected %+v", uptimeCheck, updateResp)
	}
}

func TestUptimeChecks_Delete(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/uptime/checks/check-1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
	})

	_, err := client.UptimeChecks.Delete(ctx, "check-1")
	if err != nil {
		t.Errorf("UptimeChecks.Delete returned error: %v", err)
	}
}

func TestUptimeAlert_Delete(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/uptime/checks/check-1/alerts/alert-1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
	})

	_, err := client.UptimeChecks.DeleteAlert(ctx, "check-1", "alert-1")
	if err != nil {
		t.Errorf("UptimeChecks.Delete returned error: %v", err)
	}
}

func TestUptimeAlert_Update(t *testing.T) {
	setup()
	defer teardown()

	updateRequest := &UpdateUptimeAlertRequest{
		Name:       "my alert",
		Type:       "latency",
		Threshold:  300,
		Comparison: "greater_than",
		Period:     "2m",
		Notifications: &Notifications{
			Email: []string{
				"email",
			},
			Slack: []SlackDetails{},
		},
	}
	updateResp := &UptimeAlert{
		ID:            "alert-id",
		Name:          updateRequest.Name,
		Type:          updateRequest.Type,
		Threshold:     updateRequest.Threshold,
		Comparison:    updateRequest.Comparison,
		Period:        updateRequest.Period,
		Notifications: updateRequest.Notifications,
	}

	mux.HandleFunc("/v2/uptime/checks/check-id/alerts/alert-id", func(w http.ResponseWriter, r *http.Request) {
		reqBytes, respErr := ioutil.ReadAll(r.Body)
		if respErr != nil {
			t.Error("alerts mock didn't work")
		}

		req := strings.TrimSuffix(string(reqBytes), "\n")
		expectedReq := `{"name":"my alert","type":"latency","threshold":300,"comparison":"greater_than","notifications":{"email":["email"],"slack":[]},"period":"2m"}`
		if req != expectedReq {
			t.Errorf("check req didn't match up:\n expected %+v\n got %+v\n", expectedReq, req)
		}

		resp, _ := json.Marshal(updateResp)
		fmt.Fprintf(w, fmt.Sprintf(`{"alert":%s}`, string(resp)))
	})

	alert, _, err := client.UptimeChecks.UpdateAlert(ctx, "check-id", "alert-id", updateRequest)
	if err != nil {
		t.Errorf("UptimeChecks.UpdateAlertreturned error: %v", err)
	}
	if !reflect.DeepEqual(alert, updateResp) {
		t.Errorf("UptimeChecks.UpdateAlert returned %+v, expected %+v", alert, updateResp)
	}
}

func TestUptimeAlert_Create(t *testing.T) {
	setup()
	defer teardown()

	createRequest := &CreateUptimeAlertRequest{
		Name:       "my alert",
		Type:       "latency",
		Threshold:  300,
		Comparison: "greater_than",
		Period:     "2m",
		Notifications: &Notifications{
			Email: []string{
				"email",
			},
			Slack: []SlackDetails{},
		},
	}

	createResp := &UptimeAlert{
		ID:            "alert-id",
		Name:          createRequest.Name,
		Type:          createRequest.Type,
		Threshold:     createRequest.Threshold,
		Comparison:    createRequest.Comparison,
		Period:        createRequest.Period,
		Notifications: createRequest.Notifications,
	}
	mux.HandleFunc("/v2/uptime/checks/check-id/alerts", func(w http.ResponseWriter, r *http.Request) {
		v := new(CreateUptimeAlertRequest)
		err := json.NewDecoder(r.Body).Decode(v)
		if err != nil {
			t.Fatalf("decode json: %v", err)
		}

		testMethod(t, r, http.MethodPost)
		if !reflect.DeepEqual(v, createRequest) {
			t.Errorf("Request body = %+v, expected %+v", v, createRequest)
		}

		resp, _ := json.Marshal(createResp)
		fmt.Fprintf(w, fmt.Sprintf(`{"alert":%s}`, string(resp)))
	})

	uptimeCheck, _, err := client.UptimeChecks.CreateAlert(ctx, "check-id", createRequest)
	if err != nil {
		t.Errorf("UptimeChecks.CreateAlert returned error: %v", err)
	}

	if !reflect.DeepEqual(uptimeCheck, createResp) {
		t.Errorf("UptimeChecks.CreateAlert returned %+v, expected %+v", uptimeCheck, createResp)
	}
}

func TestUptimeAlert_GetWithID(t *testing.T) {
	setup()
	defer teardown()

	alert := &UptimeAlert{
		ID:         "alert-1",
		Name:       "my alert",
		Type:       "latency",
		Threshold:  300,
		Comparison: "greater_than",
		Period:     "2m",
		Notifications: &Notifications{
			Email: []string{
				"email",
			},
			Slack: []SlackDetails{},
		},
	}

	mux.HandleFunc("/v2/uptime/checks/check-1/alerts/alert-1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		resp, _ := json.Marshal(alert)
		fmt.Fprint(w, fmt.Sprintf(`{"alert":%s}`, string(resp)))
	})

	resp, _, err := client.UptimeChecks.GetAlert(ctx, "check-1", "alert-1")
	if err != nil {
		t.Errorf("UptimeChecks.GetAlert returned error: %v", err)
	}

	if !reflect.DeepEqual(resp, alert) {
		t.Errorf("UptimeChecks.GetAlert returned %+v, expected %+v", resp, alert)
	}
}

func TestUptimeAlerts_List(t *testing.T) {
	setup()
	defer teardown()

	expectedAlerts := []UptimeAlert{
		{
			ID:         "alert-1",
			Name:       "my alert",
			Type:       "latency",
			Threshold:  300,
			Comparison: "greater_than",
			Period:     "2m",
			Notifications: &Notifications{
				Email: []string{
					"email",
				},
				Slack: []SlackDetails{},
			},
		},
		{
			ID:         "alert-2",
			Name:       "my alert",
			Type:       "latency",
			Threshold:  300,
			Comparison: "greater_than",
			Period:     "2m",
			Notifications: &Notifications{
				Email: []string{
					"email2",
				},
				Slack: []SlackDetails{},
			},
		},
	}

	mux.HandleFunc("/v2/uptime/checks/check-1/alerts", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		resp, _ := json.Marshal(expectedAlerts)
		fmt.Fprint(w, fmt.Sprintf(`{"alerts":%s, "meta": {"total": 2}}`, string(resp)))
	})

	alerts, resp, err := client.UptimeChecks.ListAlerts(ctx, "check-1", nil)
	if err != nil {
		t.Errorf("UptimeChecks.ListAlerts returned error: %v", err)
	}

	if !reflect.DeepEqual(alerts, expectedAlerts) {
		t.Errorf("UptimeChecks.ListAlerts returned uptime checks %+v, expected %+v", alerts, expectedAlerts)
	}

	expectedMeta := &Meta{Total: 2}
	if !reflect.DeepEqual(resp.Meta, expectedMeta) {
		t.Errorf("UptimeChecks.List returned meta %+v, expected %+v", resp.Meta, expectedMeta)
	}
}

func TestUptimeAlerts_ListWithMultiplePages(t *testing.T) {
	setup()
	defer teardown()

	mockResp := `
	{
		"alerts": [{
			"id": "alert-1",
			"name": "Landing page degraded performance",
			"type": "latency",
			"threshold": 300,
			"comparison": "greater_than",
			"notifications": {
				"email": [
					"bob@example.com"
				],
				"slack": [{
					"channel": "Production Alerts",
					"url": "https://hooks.slack.com/services/T1234567/AAAAAAAA/ZZZZZZ"
				}]
			},
			"period": "2m"
		}],
		"links": {
			"pages": {
				"next": "http://example.com/v2/uptime/checks/check-1/alerts?page=2"
			}
		},
		"meta": {
			"total": 1
		}
	}`

	mux.HandleFunc("/v2/uptime/checks/check-1/alerts", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, mockResp)
	})

	_, resp, err := client.UptimeChecks.ListAlerts(ctx, "check-1", nil)
	if err != nil {
		t.Errorf("UptimeChecks.ListAlerts returned error: %v", err)
	}

	checkCurrentPage(t, resp, 1)
}

func TestUptimeAlerts_ListWithPageNumber(t *testing.T) {
	setup()
	defer teardown()

	mockResp := `
	{
		"alerts": [
		  {
			"id": "alert-1",
			"name": "Landing page degraded performance",
			"type": "latency",
			"threshold": 300,
			"comparison": "greater_than",
			"notifications": {
			  "email": [
				"bob@example.com"
			  ],
			  "slack": [
				{
				  "channel": "Production Alerts",
				  "url": "https://hooks.slack.com/services/T1234567/AAAAAAAA/ZZZZZZ"
				}
			  ]
			},
			"period": "2m"
		  }
		],
		"links": {
		  "pages": {
			"next": "http://example.com/v2/uptime/checks?page=3",
			"prev": "http://example.com/v2/uptime/checks?page=1",
			"last": "http://example.com/v2/uptime/checks?page=3",
			"first": "http://example.com/v2/uptime/checks?page=1"
		  }
		}
	  }`

	mux.HandleFunc("/v2/uptime/checks/check-1/alerts", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, mockResp)
	})

	_, resp, err := client.UptimeChecks.ListAlerts(ctx, "check-1", &ListOptions{Page: 2})
	if err != nil {
		t.Errorf("UptimeChecks.ListAlerts returned error: %v", err)
	}

	checkCurrentPage(t, resp, 2)
}
