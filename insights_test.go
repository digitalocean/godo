package godo

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"testing"
	"time"
)

func TestInsights_NotificationChannels(t *testing.T) {
	createdAt := time.Date(2026, 9, 3, 10, 15, 0, 0, time.UTC)
	channel := &NotificationChannel{
		ID:          "550e8400-e29b-41d4-a716-446655440000",
		Name:        "Platform alerts",
		ChannelType: InsightsChannelTypeSlack,
		Slack: &SlackNotificationConfig{
			WebhookURL: "********",
			Channel:    "#platform-alerts",
		},
		Usage:     &NotificationChannelUsage{RuleCount: 2},
		CreatedAt: createdAt,
		UpdatedAt: createdAt,
	}

	t.Run("list", func(t *testing.T) {
		setup()
		defer teardown()

		mux.HandleFunc("/v2/insights/notification-channels", func(w http.ResponseWriter, r *http.Request) {
			testMethod(t, r, http.MethodGet)
			if r.URL.Query().Get("page") != "1" || r.URL.Query().Get("per_page") != "20" {
				t.Errorf("query = %s", r.URL.RawQuery)
			}
			body, _ := json.Marshal(channel)
			fmt.Fprintf(w, `{"notification_channels":[%s],"links":{"pages":{"next":"https://api.digitalocean.com/v2/insights/notification-channels?page=2"}},"meta":{"total":1}}`, body)
		})

		got, resp, err := client.Insights.ListNotificationChannels(ctx, &ListOptions{Page: 1, PerPage: 20})
		if err != nil {
			t.Fatalf("ListNotificationChannels: %v", err)
		}
		if !reflect.DeepEqual(got, []NotificationChannel{*channel}) {
			t.Errorf("got %+v", got)
		}
		if resp.Meta == nil || resp.Meta.Total != 1 {
			t.Errorf("meta = %+v", resp.Meta)
		}
		checkCurrentPage(t, resp, 1)
	})

	t.Run("get", func(t *testing.T) {
		setup()
		defer teardown()
		mux.HandleFunc("/v2/insights/notification-channels/"+channel.ID, func(w http.ResponseWriter, r *http.Request) {
			testMethod(t, r, http.MethodGet)
			body, _ := json.Marshal(channel)
			fmt.Fprintf(w, `{"notification_channel":%s}`, body)
		})
		got, _, err := client.Insights.GetNotificationChannel(ctx, channel.ID)
		if err != nil {
			t.Fatalf("GetNotificationChannel: %v", err)
		}
		if !reflect.DeepEqual(got, channel) {
			t.Errorf("got %+v", got)
		}
	})

	t.Run("create", func(t *testing.T) {
		setup()
		defer teardown()
		create := &NotificationChannelRequest{
			Name: "Platform alerts",
			Slack: &SlackNotificationConfig{
				WebhookURL: "https://hooks.slack.com/services/T000/B000/XXXXXXXX",
				Channel:    "#platform-alerts",
			},
		}
		mux.HandleFunc("/v2/insights/notification-channels", func(w http.ResponseWriter, r *http.Request) {
			testMethod(t, r, http.MethodPost)
			var got NotificationChannelRequest
			if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
				t.Fatalf("decode: %v", err)
			}
			if !reflect.DeepEqual(&got, create) {
				t.Errorf("body = %+v", got)
			}
			body, _ := json.Marshal(channel)
			w.WriteHeader(http.StatusCreated)
			fmt.Fprintf(w, `{"notification_channel":%s}`, body)
		})
		got, _, err := client.Insights.CreateNotificationChannel(ctx, create)
		if err != nil {
			t.Fatalf("CreateNotificationChannel: %v", err)
		}
		if !reflect.DeepEqual(got, channel) {
			t.Errorf("got %+v", got)
		}
	})

	t.Run("update omits empty secret", func(t *testing.T) {
		setup()
		defer teardown()
		update := &NotificationChannelRequest{
			Name:  "Platform alerts",
			Slack: &SlackNotificationConfig{Channel: "#platform-alerts-prod"},
		}
		mux.HandleFunc("/v2/insights/notification-channels/"+channel.ID, func(w http.ResponseWriter, r *http.Request) {
			testMethod(t, r, http.MethodPut)
			raw, _ := io.ReadAll(r.Body)
			var body map[string]any
			if err := json.Unmarshal(raw, &body); err != nil {
				t.Fatalf("decode: %v", err)
			}
			slack, _ := body["slack"].(map[string]any)
			if _, ok := slack["webhook_url"]; ok {
				t.Errorf("webhook_url should be omitted, body = %s", raw)
			}
			fmt.Fprintf(w, `{"notification_channel":%s}`, mustJSON(channel))
		})
		if _, _, err := client.Insights.UpdateNotificationChannel(ctx, channel.ID, update); err != nil {
			t.Fatalf("UpdateNotificationChannel: %v", err)
		}
	})

	t.Run("delete", func(t *testing.T) {
		setup()
		defer teardown()
		mux.HandleFunc("/v2/insights/notification-channels/"+channel.ID, func(w http.ResponseWriter, r *http.Request) {
			testMethod(t, r, http.MethodDelete)
			w.WriteHeader(http.StatusNoContent)
		})
		if _, err := client.Insights.DeleteNotificationChannel(ctx, channel.ID); err != nil {
			t.Fatalf("DeleteNotificationChannel: %v", err)
		}
	})
}

func TestInsights_AlertRules(t *testing.T) {
	createdAt := time.Date(2026, 9, 3, 10, 0, 0, 0, time.UTC)
	warning := 80.0
	critical := 95.0
	channels := []NotificationChannelBinding{{
		NotificationChannelID: "550e8400-e29b-41d4-a716-446655440000",
		NotifyOn:              []string{InsightsSeverityCritical},
	}}
	rule := &AlertRule{
		ID: "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
		Spec: AlertRuleSpec{
			Name: "High CPU",
			Query: AlertMetricsQuery{
				Metric:       "do.droplets.cpu_utilization",
				ResourceURNs: []string{"do:droplet:12345"},
			},
			Condition:            &AlertCondition{Window: InsightsEvaluationWindow5m},
			Thresholds:           AlertThresholds{Warning: &warning, Critical: &critical, Operator: InsightsThresholdOperatorGreaterThan},
			NotificationChannels: &channels,
			ReAlertDuration:      InsightsReAlertDuration4h,
		},
		Status:    InsightsAlertRuleStatusActive,
		CreatedAt: createdAt,
		UpdatedAt: createdAt,
	}

	t.Run("list", func(t *testing.T) {
		setup()
		defer teardown()
		mux.HandleFunc("/v2/insights/alert-rules", func(w http.ResponseWriter, r *http.Request) {
			testMethod(t, r, http.MethodGet)
			if r.URL.Query().Get("resource_urn") != "do:droplet:12345" {
				t.Errorf("query = %s", r.URL.RawQuery)
			}
			fmt.Fprintf(w, `{"alert_rules":[%s],"meta":{"total":1}}`, mustJSON(rule))
		})
		got, resp, err := client.Insights.ListAlertRules(ctx, &AlertRuleListOptions{ResourceURN: "do:droplet:12345"})
		if err != nil {
			t.Fatalf("ListAlertRules: %v", err)
		}
		if !reflect.DeepEqual(got, []AlertRule{*rule}) {
			t.Errorf("got %+v, want %+v", got, []AlertRule{*rule})
		}
		if resp.Meta == nil || resp.Meta.Total != 1 {
			t.Errorf("meta = %+v", resp.Meta)
		}
	})

	t.Run("create", func(t *testing.T) {
		setup()
		defer teardown()
		create := &AlertRuleRequest{Spec: rule.Spec}
		mux.HandleFunc("/v2/insights/alert-rules", func(w http.ResponseWriter, r *http.Request) {
			testMethod(t, r, http.MethodPost)
			var got AlertRuleRequest
			if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
				t.Fatalf("decode: %v", err)
			}
			if !reflect.DeepEqual(got.Spec, create.Spec) {
				t.Errorf("spec = %+v", got.Spec)
			}
			w.WriteHeader(http.StatusCreated)
			fmt.Fprintf(w, `{"alert_rule":%s}`, mustJSON(rule))
		})
		got, _, err := client.Insights.CreateAlertRule(ctx, create)
		if err != nil {
			t.Fatalf("CreateAlertRule: %v", err)
		}
		if !reflect.DeepEqual(got, rule) {
			t.Errorf("got %+v", got)
		}
	})

	t.Run("update omits bindings and status", func(t *testing.T) {
		setup()
		defer teardown()
		update := &AlertRuleRequest{Spec: AlertRuleSpec{
			Name:       "High CPU",
			Query:      AlertMetricsQuery{Metric: "do.droplets.cpu_utilization"},
			Thresholds: AlertThresholds{Critical: &critical, Operator: InsightsThresholdOperatorGreaterThan},
		}}
		mux.HandleFunc("/v2/insights/alert-rules/"+rule.ID, func(w http.ResponseWriter, r *http.Request) {
			testMethod(t, r, http.MethodPut)
			raw, _ := io.ReadAll(r.Body)
			var body map[string]any
			if err := json.Unmarshal(raw, &body); err != nil {
				t.Fatalf("decode: %v", err)
			}
			if _, ok := body["status"]; ok {
				t.Errorf("status should be omitted: %s", raw)
			}
			spec, _ := body["spec"].(map[string]any)
			if _, ok := spec["notification_channels"]; ok {
				t.Errorf("notification_channels should be omitted: %s", raw)
			}
			fmt.Fprintf(w, `{"alert_rule":%s}`, mustJSON(rule))
		})
		if _, _, err := client.Insights.UpdateAlertRule(ctx, rule.ID, update); err != nil {
			t.Fatalf("UpdateAlertRule: %v", err)
		}
	})

	t.Run("get and delete", func(t *testing.T) {
		setup()
		defer teardown()
		mux.HandleFunc("/v2/insights/alert-rules/"+rule.ID, func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet:
				fmt.Fprintf(w, `{"alert_rule":%s}`, mustJSON(rule))
			case http.MethodDelete:
				w.WriteHeader(http.StatusNoContent)
			default:
				t.Errorf("method %s", r.Method)
			}
		})
		got, _, err := client.Insights.GetAlertRule(ctx, rule.ID)
		if err != nil {
			t.Fatalf("GetAlertRule: %v", err)
		}
		if !reflect.DeepEqual(got, rule) {
			t.Errorf("got %+v", got)
		}
		if _, err := client.Insights.DeleteAlertRule(ctx, rule.ID); err != nil {
			t.Fatalf("DeleteAlertRule: %v", err)
		}
	})
}

func TestInsights_AlertInstances(t *testing.T) {
	triggered := time.Date(2026, 9, 3, 10, 15, 0, 0, time.UTC)
	last := time.Date(2026, 9, 3, 10, 45, 0, 0, time.UTC)
	resolved := time.Date(2026, 9, 3, 11, 0, 0, 0, time.UTC)
	instance := &AlertInstance{
		ID:              "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
		RuleID:          "8f3a2b1c-4d5e-6f7a-8b9c-0d1e2f3a4b5c",
		Severity:        InsightsAlertInstanceSeverityWarning,
		Status:          InsightsAlertInstanceStatusResolved,
		ResourceURN:     "do:droplet:12345",
		Value:           87.5,
		TriggeredAt:     triggered,
		LastTriggeredAt: last,
		ResolvedAt:      &resolved,
	}

	setup()
	defer teardown()

	mux.HandleFunc("/v2/insights/alert-instances", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		q := r.URL.Query()
		if q.Get("status") != InsightsAlertInstanceStatusActive || q.Get("rule_id") != instance.RuleID || q.Get("resource_urn") != instance.ResourceURN {
			t.Errorf("query = %s", r.URL.RawQuery)
		}
		fmt.Fprintf(w, `{"alert_instances":[%s],"meta":{"total":1}}`, mustJSON(instance))
	})
	mux.HandleFunc("/v2/insights/alert-instances/"+instance.ID, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprintf(w, `{"alert_instance":%s}`, mustJSON(instance))
	})

	got, resp, err := client.Insights.ListAlertInstances(ctx, &AlertInstanceListOptions{
		Status:      InsightsAlertInstanceStatusActive,
		RuleID:      instance.RuleID,
		ResourceURN: instance.ResourceURN,
	})
	if err != nil {
		t.Fatalf("ListAlertInstances: %v", err)
	}
	if !reflect.DeepEqual(got, []AlertInstance{*instance}) {
		t.Errorf("list got %+v", got)
	}
	if resp.Meta == nil || resp.Meta.Total != 1 {
		t.Errorf("meta = %+v", resp.Meta)
	}

	one, _, err := client.Insights.GetAlertInstance(ctx, instance.ID)
	if err != nil {
		t.Fatalf("GetAlertInstance: %v", err)
	}
	if !reflect.DeepEqual(one, instance) {
		t.Errorf("get got %+v", one)
	}
}

func TestInsights_PromQL(t *testing.T) {
	vectorBody := `{
		"status":"success",
		"data":{
			"resultType":"vector",
			"result":[{"metric":{"__name__":"do_droplets_cpu_time"},"value":[1620683817.0,"1"]}]
		}
	}`
	rangeBody := `{
		"status":"success",
		"data":{
			"resultType":"matrix",
			"result":[{"metric":{"__name__":"do_droplets_cpu_time"},"values":[[1620683817.0,"1"],[1620683832.0,"2"]]}]
		}
	}`

	t.Run("query get", func(t *testing.T) {
		setup()
		defer teardown()
		mux.HandleFunc("/v2/insights/query/nyc3/prom/api/v1/query", func(w http.ResponseWriter, r *http.Request) {
			testMethod(t, r, http.MethodGet)
			if r.URL.Query().Get("query") != "do.droplets.cpu_time" || r.URL.Query().Get("time") != "1620683817" {
				t.Errorf("query = %s", r.URL.RawQuery)
			}
			fmt.Fprint(w, vectorBody)
		})
		got, _, err := client.Insights.Query(ctx, "nyc3", &PromQueryOptions{Query: "do.droplets.cpu_time", Time: "1620683817"})
		if err != nil {
			t.Fatalf("Query: %v", err)
		}
		if got.ResultType != "vector" || len(got.Vector) != 1 || got.Vector[0].Value.Value != "1" || got.Vector[0].Metric["__name__"] != "do_droplets_cpu_time" {
			t.Errorf("got %+v", got)
		}
	})

	t.Run("query post and string result", func(t *testing.T) {
		setup()
		defer teardown()
		mux.HandleFunc("/v2/insights/query/nyc3/prom/api/v1/query", func(w http.ResponseWriter, r *http.Request) {
			testMethod(t, r, http.MethodPost)
			if r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
				t.Errorf("content-type = %s", r.Header.Get("Content-Type"))
			}
			if err := r.ParseForm(); err != nil {
				t.Fatal(err)
			}
			if r.Form.Get("query") != `rate(do.droplets.cpu_time[5m])` || r.Form.Get("timeout") != "30s" {
				t.Errorf("form = %v", r.Form)
			}
			fmt.Fprint(w, `{"status":"success","data":{"resultType":"string","result":[1620683817.0,"up"]}}`)
		})
		got, _, err := client.Insights.PostQuery(ctx, "nyc3", &PromQueryOptions{Query: `rate(do.droplets.cpu_time[5m])`, Timeout: "30s"})
		if err != nil {
			t.Fatalf("PostQuery: %v", err)
		}
		if got.ResultType != "string" || got.Sample == nil || got.Sample.Value != "up" || got.Sample.Timestamp != 1620683817 {
			t.Errorf("got %+v", got)
		}
	})

	t.Run("query range", func(t *testing.T) {
		setup()
		defer teardown()
		mux.HandleFunc("/v2/insights/query/nyc3/prom/api/v1/query_range", func(w http.ResponseWriter, r *http.Request) {
			testMethod(t, r, http.MethodGet)
			q := r.URL.Query()
			if q.Get("start") != "1620683817" || q.Get("end") != "1620705417" || q.Get("step") != "15s" {
				t.Errorf("query = %s", r.URL.RawQuery)
			}
			fmt.Fprint(w, rangeBody)
		})
		got, _, err := client.Insights.QueryRange(ctx, "nyc3", &PromQueryRangeOptions{
			Query: "do.droplets.cpu_time", Start: "1620683817", End: "1620705417", Step: "15s",
		})
		if err != nil {
			t.Fatalf("QueryRange: %v", err)
		}
		if got.Data.ResultType != "matrix" || len(got.Data.Result) != 1 || len(got.Data.Result[0].Values) != 2 {
			t.Errorf("got %+v", got)
		}
	})

	t.Run("series post matchers", func(t *testing.T) {
		setup()
		defer teardown()
		mux.HandleFunc("/v2/insights/query/nyc3/prom/api/v1/series", func(w http.ResponseWriter, r *http.Request) {
			testMethod(t, r, http.MethodPost)
			if err := r.ParseForm(); err != nil {
				t.Fatal(err)
			}
			got := r.Form["match[]"]
			want := []string{`{__name__="do.droplets.cpu_time"}`, `{job="node"}`}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("match[] = %#v", got)
			}
			fmt.Fprint(w, `{"status":"success","data":[{"__name__":"do_droplets_cpu_time"}]}`)
		})
		got, _, err := client.Insights.PostSeries(ctx, "nyc3", &PromSelectorOptions{Match: []string{`{__name__="do.droplets.cpu_time"}`, `{job="node"}`}})
		if err != nil {
			t.Fatalf("PostSeries: %v", err)
		}
		if len(got.Data) != 1 || got.Data[0]["__name__"] != "do_droplets_cpu_time" {
			t.Errorf("got %+v", got)
		}
	})

	t.Run("labels and label values", func(t *testing.T) {
		setup()
		defer teardown()
		mux.HandleFunc("/v2/insights/query/nyc3/prom/api/v1/labels", func(w http.ResponseWriter, r *http.Request) {
			testMethod(t, r, http.MethodGet)
			fmt.Fprint(w, `{"status":"success","data":["__name__","job"]}`)
		})
		mux.HandleFunc("/v2/insights/query/nyc3/prom/api/v1/label/__name__/values", func(w http.ResponseWriter, r *http.Request) {
			testMethod(t, r, http.MethodGet)
			if r.URL.Query().Get("match[]") != `{__name__=~".+"}` {
				t.Errorf("query = %s", r.URL.RawQuery)
			}
			fmt.Fprint(w, `{"status":"success","data":["do_droplets_cpu_time"]}`)
		})
		labels, _, err := client.Insights.Labels(ctx, "nyc3", nil)
		if err != nil {
			t.Fatalf("Labels: %v", err)
		}
		if !reflect.DeepEqual(labels.Data, []string{"__name__", "job"}) {
			t.Errorf("labels = %+v", labels)
		}
		values, _, err := client.Insights.LabelValues(ctx, "nyc3", "__name__", &PromSelectorOptions{Match: []string{`{__name__=~".+"}`}})
		if err != nil {
			t.Fatalf("LabelValues: %v", err)
		}
		if !reflect.DeepEqual(values.Data, []string{"do_droplets_cpu_time"}) {
			t.Errorf("values = %+v", values)
		}
	})

	t.Run("post range labels", func(t *testing.T) {
		setup()
		defer teardown()
		mux.HandleFunc("/v2/insights/query/nyc3/prom/api/v1/query_range", func(w http.ResponseWriter, r *http.Request) {
			testMethod(t, r, http.MethodPost)
			fmt.Fprint(w, rangeBody)
		})
		mux.HandleFunc("/v2/insights/query/nyc3/prom/api/v1/labels", func(w http.ResponseWriter, r *http.Request) {
			testMethod(t, r, http.MethodPost)
			fmt.Fprint(w, `{"status":"success","data":["job"]}`)
		})
		mux.HandleFunc("/v2/insights/query/nyc3/prom/api/v1/series", func(w http.ResponseWriter, r *http.Request) {
			testMethod(t, r, http.MethodGet)
			fmt.Fprint(w, `{"status":"success","data":[]}`)
		})
		if _, _, err := client.Insights.PostQueryRange(ctx, "nyc3", &PromQueryRangeOptions{Query: "up", Start: "1", End: "2", Step: "15s"}); err != nil {
			t.Fatalf("PostQueryRange: %v", err)
		}
		if _, _, err := client.Insights.PostLabels(ctx, "nyc3", nil); err != nil {
			t.Fatalf("PostLabels: %v", err)
		}
		if _, _, err := client.Insights.Series(ctx, "nyc3", &PromSelectorOptions{Match: []string{"up"}}); err != nil {
			t.Fatalf("Series: %v", err)
		}
	})
}

func TestPromQueryResponse_UnknownType(t *testing.T) {
	var resp PromQueryResponse
	err := json.Unmarshal([]byte(`{"status":"success","data":{"resultType":"histogram","result":[]}}`), &resp)
	if err == nil {
		t.Fatal("expected error")
	}
}

func mustJSON(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(b)
}

func TestInsights_SearchLogs(t *testing.T) {
	stringValue := "Error"
	boolValue := false
	numberValue := 0.0
	search := &LogsSearchRequest{
		TimeRange: LogsTimeRange{
			From: LogsTimeInstant{Absolute: "2026-09-30T00:00:00Z"},
			To:   LogsTimeInstant{Relative: "now"},
		},
		Filter: &LogsFilterExpression{
			And: &LogsFilterGroup{
				Expressions: []LogsFilterExpression{
					{
						Condition: &LogsFilterCondition{
							Field:    LogsFieldRef{Name: "severity_text"},
							Operator: InsightsLogsFilterOperatorEqual,
							Value:    &LogsFilterValue{StringValue: &stringValue},
						},
					},
					{
						Or: &LogsFilterGroup{
							Expressions: []LogsFilterExpression{
								{
									Condition: &LogsFilterCondition{
										Field:    LogsFieldRef{Name: "debug", Scope: InsightsLogsFieldScopeAttributes},
										Operator: InsightsLogsFilterOperatorEqual,
										Value:    &LogsFilterValue{BoolValue: &boolValue},
									},
								},
								{
									Condition: &LogsFilterCondition{
										Field:    LogsFieldRef{Name: "retry_count", Scope: InsightsLogsFieldScopeAttributes},
										Operator: InsightsLogsFilterOperatorGreaterThanOrEqual,
										Value:    &LogsFilterValue{NumberValue: &numberValue},
									},
								},
							},
						},
					},
					{
						Not: &LogsFilterExpression{
							TextSearch: &LogsTextSearch{Query: "healthcheck"},
						},
					},
				},
			},
		},
		OrderBy: []LogsOrderBy{
			{
				Field:     LogsFieldRef{Name: "timestamp"},
				Direction: InsightsLogsSortDirectionDescending,
			},
		},
		Pagination: &LogsPaginationRequest{Limit: 100},
	}

	setup()
	defer teardown()

	mux.HandleFunc("/v2/insights/query/nyc3/logs/search", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Errorf("Content-Type = %q", got)
		}

		var got LogsSearchRequest
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if !reflect.DeepEqual(got, *search) {
			t.Errorf("request = %#v, want %#v", got, *search)
		}

		fmt.Fprint(w, `{
			"data": [{
				"timestamp": "2026-09-30T00:30:00Z",
				"severity_number": 17,
				"severity_text": "Error",
				"body": "Connection timeout",
				"trace_id": "abc123",
				"span_id": "def456",
				"service_name": "api-gateway",
				"resource": {"do.component": "droplet", "do.droplet.id": "12345678"},
				"attributes": {"http.method": "GET", "http.status_code": "504"}
			}],
			"pagination": {"has_more": true, "next_cursor": "next-page"}
		}`)
	})

	got, resp, err := client.Insights.SearchLogs(ctx, "nyc3", search)
	if err != nil {
		t.Fatalf("SearchLogs: %v", err)
	}
	if resp == nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("response = %#v", resp)
	}

	want := &LogsSearchResponse{
		Data: []InsightsLogRecord{
			{
				Timestamp:      time.Date(2026, 9, 30, 0, 30, 0, 0, time.UTC),
				SeverityNumber: 17,
				SeverityText:   "Error",
				Body:           "Connection timeout",
				TraceID:        "abc123",
				SpanID:         "def456",
				ServiceName:    "api-gateway",
				Resource:       map[string]string{"do.component": "droplet", "do.droplet.id": "12345678"},
				Attributes:     map[string]string{"http.method": "GET", "http.status_code": "504"},
			},
		},
		Pagination: &LogsPaginationResponse{HasMore: true, NextCursor: "next-page"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("SearchLogs = %#v, want %#v", got, want)
	}
}

func TestLogsFilterValueArrays(t *testing.T) {
	tests := []struct {
		name  string
		value LogsFilterValue
		want  string
	}{
		{
			name: "strings",
			value: LogsFilterValue{
				StringArrayValue: &LogsStringArrayValue{Values: []string{"droplet-123", "droplet-456"}},
			},
			want: `{"string_array_value":{"values":["droplet-123","droplet-456"]}}`,
		},
		{
			name: "numbers",
			value: LogsFilterValue{
				NumberArrayValue: &LogsNumberArrayValue{Values: []float64{42, 84}},
			},
			want: `{"number_array_value":{"values":[42,84]}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.value)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			if string(got) != tt.want {
				t.Errorf("JSON = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestInsights_SearchLogsCursorAndEmptyData(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/insights/query/nyc3/logs/search", func(w http.ResponseWriter, r *http.Request) {
		var got LogsSearchRequest
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		want := &LogsPaginationRequest{Limit: 25, Cursor: "next-page"}
		if !reflect.DeepEqual(got.Pagination, want) {
			t.Errorf("pagination = %#v, want %#v", got.Pagination, want)
		}
		fmt.Fprint(w, `{"pagination":{"has_more":false}}`)
	})

	got, _, err := client.Insights.SearchLogs(ctx, "nyc3", &LogsSearchRequest{
		TimeRange: LogsTimeRange{
			From: LogsTimeInstant{UnixNano: "1790726400000000000"},
			To:   LogsTimeInstant{Absolute: "2026-09-30T01:00:00Z"},
		},
		Pagination: &LogsPaginationRequest{Limit: 25, Cursor: "next-page"},
	})
	if err != nil {
		t.Fatalf("SearchLogs: %v", err)
	}
	if got.Data != nil {
		t.Errorf("data = %#v, want nil", got.Data)
	}
	if got.Pagination == nil || got.Pagination.HasMore || got.Pagination.NextCursor != "" {
		t.Errorf("pagination = %#v", got.Pagination)
	}
}

func TestInsights_SearchLogsEscapesRegion(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.URL.EscapedPath(), "/v2/insights/query/nyc3%2Fother/logs/search"; got != want {
			t.Errorf("escaped path = %q, want %q", got, want)
		}
		fmt.Fprint(w, `{}`)
	})

	_, _, err := client.Insights.SearchLogs(ctx, "nyc3/other", &LogsSearchRequest{
		TimeRange: LogsTimeRange{
			From: LogsTimeInstant{Relative: "1h"},
			To:   LogsTimeInstant{Relative: "now"},
		},
	})
	if err != nil {
		t.Fatalf("SearchLogs: %v", err)
	}
}

func TestInsights_SearchLogsError(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/insights/query/nyc3/logs/search", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"message":"time range must not exceed 7 days"}`)
	})

	got, resp, err := client.Insights.SearchLogs(ctx, "nyc3", &LogsSearchRequest{})
	if err == nil {
		t.Fatal("expected error")
	}
	if got != nil {
		t.Errorf("result = %#v, want nil", got)
	}
	if resp == nil || resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("response = %#v", resp)
	}
	var apiErr *ErrorResponse
	if !errors.As(err, &apiErr) || apiErr.Message != "time range must not exceed 7 days" {
		t.Errorf("error = %#v", err)
	}
}

func TestInsights_SearchSpans(t *testing.T) {
	search := &SpansSearchRequest{
		TimeRange: SpansTimeRange{
			From: SpansTimeBound{Absolute: "2026-09-30T00:00:00Z"},
			To:   SpansTimeBound{Relative: "now"},
		},
		Filter: &SpansFilterExpression{
			Type: InsightsSpansFilterTypeAnd,
			Expressions: []SpansFilterExpression{
				{
					Type:     InsightsSpansFilterTypeCondition,
					Field:    &SpansFilterField{Name: "serviceName"},
					Operator: InsightsSpansFilterOperatorEqual,
					Value:    &SpansFilterValue{String: String("api-gateway")},
				},
				{
					Type: InsightsSpansFilterTypeOr,
					Expressions: []SpansFilterExpression{
						{
							Type:     InsightsSpansFilterTypeCondition,
							Field:    &SpansFilterField{Name: "statusCode"},
							Operator: InsightsSpansFilterOperatorEqual,
							Value:    &SpansFilterValue{String: String("Error")},
						},
						{
							Type:     InsightsSpansFilterTypeCondition,
							Field:    &SpansFilterField{Name: "durationNs"},
							Operator: InsightsSpansFilterOperatorGreaterThanOrEqual,
							Value:    &SpansFilterValue{Number: PtrTo(1000000.0)},
						},
						{
							Type:     InsightsSpansFilterTypeCondition,
							Field:    &SpansFilterField{Name: "http.method", Scope: InsightsSpansFieldScopeAttributes},
							Operator: InsightsSpansFilterOperatorIn,
							Value:    &SpansFilterValue{StringArray: []string{"GET", "POST"}},
						},
					},
				},
				{
					Type: InsightsSpansFilterTypeNot,
					Expressions: []SpansFilterExpression{
						{Type: InsightsSpansFilterTypeTextSearch, Query: "healthcheck"},
					},
				},
			},
		},
		OrderBy: []SpansOrderBy{
			{
				Field:     SpansFilterField{Name: "startTime"},
				Direction: InsightsSpansSortDirectionDescending,
			},
		},
		Pagination: &SpansPaginationRequest{Limit: 100},
	}

	setup()
	defer teardown()

	mux.HandleFunc("/v2/insights/query/nyc3/spans/search", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Errorf("Content-Type = %q", got)
		}

		var got SpansSearchRequest
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if !reflect.DeepEqual(got, *search) {
			t.Errorf("request = %#v, want %#v", got, *search)
		}

		fmt.Fprint(w, `{
			"data": [{
				"traceId": "4bf92f3577b34da6a3ce929d0e0e4736",
				"spanId": "00f067aa0ba902b7",
				"parentSpanId": "00f067aa0ba902b6",
				"startTime": "2026-09-30T00:00:00Z",
				"endTime": "2026-09-30T00:00:01Z",
				"durationNs": 1000000,
				"name": "GET /v2/droplets",
				"kind": "Server",
				"statusCode": "Ok",
				"serviceName": "api-gateway",
				"resource": {"service.name": "api-gateway", "service.version": "1.4.2"},
				"attributes": {"http.method": "GET", "http.status_code": "200"},
				"region": "nyc3"
			}],
			"pagination": {"hasMore": true}
		}`)
	})

	got, resp, err := client.Insights.SearchSpans(ctx, "nyc3", search)
	if err != nil {
		t.Fatalf("SearchSpans: %v", err)
	}
	if resp == nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("response = %#v", resp)
	}

	want := &SpansSearchResponse{
		Data: []SpanRecord{
			{
				TraceID:      "4bf92f3577b34da6a3ce929d0e0e4736",
				SpanID:       "00f067aa0ba902b7",
				ParentSpanID: "00f067aa0ba902b6",
				StartTime:    time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC),
				EndTime:      time.Date(2026, 9, 30, 0, 0, 1, 0, time.UTC),
				DurationNs:   1000000,
				Name:         "GET /v2/droplets",
				Kind:         "Server",
				StatusCode:   "Ok",
				ServiceName:  "api-gateway",
				Resource:     map[string]string{"service.name": "api-gateway", "service.version": "1.4.2"},
				Attributes:   map[string]string{"http.method": "GET", "http.status_code": "200"},
				Region:       "nyc3",
			},
		},
		Pagination: &SpansPaginationResponse{HasMore: true},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("SearchSpans = %#v, want %#v", got, want)
	}
}

func TestSpansFilterValueMarshal(t *testing.T) {
	tests := []struct {
		name  string
		value SpansFilterValue
		want  string
	}{
		{"string", SpansFilterValue{String: String("api-gateway")}, `{"type":"string","value":"api-gateway"}`},
		{"number", SpansFilterValue{Number: PtrTo(503.0)}, `{"type":"number","value":503}`},
		{"bool", SpansFilterValue{Bool: PtrTo(false)}, `{"type":"bool","value":false}`},
		{"string_array", SpansFilterValue{StringArray: []string{"GET", "POST"}}, `{"type":"string_array","value":["GET","POST"]}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.value)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			if string(got) != tt.want {
				t.Errorf("JSON = %s, want %s", got, tt.want)
			}
		})
	}

	if _, err := json.Marshal(SpansFilterValue{}); err == nil {
		t.Error("expected error marshaling empty SpansFilterValue")
	}
	if _, err := json.Marshal(SpansFilterValue{String: String("x"), Bool: PtrTo(true)}); err == nil {
		t.Error("expected error marshaling multi-field SpansFilterValue")
	}
}

func TestSpansFilterValueUnmarshal(t *testing.T) {
	tests := []struct {
		name string
		data string
		want SpansFilterValue
	}{
		{"typed string", `{"type":"string","value":"api-gateway"}`, SpansFilterValue{String: String("api-gateway")}},
		{"bare string", `"api-gateway"`, SpansFilterValue{String: String("api-gateway")}},
		{"typed number", `{"type":"number","value":503}`, SpansFilterValue{Number: PtrTo(503.0)}},
		{"bare number", `503`, SpansFilterValue{Number: PtrTo(503.0)}},
		{"typed bool", `{"type":"bool","value":true}`, SpansFilterValue{Bool: PtrTo(true)}},
		{"bare bool", `false`, SpansFilterValue{Bool: PtrTo(false)}},
		{"typed string_array", `{"type":"string_array","value":["GET","POST"]}`, SpansFilterValue{StringArray: []string{"GET", "POST"}}},
		{"bare string_array", `["GET","POST"]`, SpansFilterValue{StringArray: []string{"GET", "POST"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got SpansFilterValue
			if err := json.Unmarshal([]byte(tt.data), &got); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestInsights_SearchSpansEscapesRegion(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.URL.EscapedPath(), "/v2/insights/query/nyc3%2Fother/spans/search"; got != want {
			t.Errorf("escaped path = %q, want %q", got, want)
		}
		fmt.Fprint(w, `{}`)
	})

	_, _, err := client.Insights.SearchSpans(ctx, "nyc3/other", &SpansSearchRequest{
		TimeRange: SpansTimeRange{
			From: SpansTimeBound{Relative: "1h"},
			To:   SpansTimeBound{Relative: "now"},
		},
	})
	if err != nil {
		t.Fatalf("SearchSpans: %v", err)
	}
}

func TestInsights_SearchSpansError(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/insights/query/nyc3/spans/search", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"message":"time range must not exceed 7 days"}`)
	})

	got, resp, err := client.Insights.SearchSpans(ctx, "nyc3", &SpansSearchRequest{})
	if err == nil {
		t.Fatal("expected error")
	}
	if got != nil {
		t.Errorf("result = %#v, want nil", got)
	}
	if resp == nil || resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("response = %#v", resp)
	}
	var apiErr *ErrorResponse
	if !errors.As(err, &apiErr) || apiErr.Message != "time range must not exceed 7 days" {
		t.Errorf("error = %#v", err)
	}
}

func TestInsights_SearchTraces(t *testing.T) {
	search := &TracesSearchRequest{
		TimeRange: TracesTimeRange{
			From: TracesTimeBound{Absolute: "2026-10-09T00:00:00Z"},
			To:   TracesTimeBound{Relative: "now"},
		},
		Filter: &TracesFilterExpression{
			Type: InsightsTracesFilterTypeAnd,
			Expressions: []TracesFilterExpression{
				{
					Type:     InsightsTracesFilterTypeCondition,
					Field:    &TracesFilterField{Name: "service.name"},
					Operator: InsightsTracesFilterOperatorEqual,
					Value:    &TracesFilterValue{String: String("checkout-api")},
				},
				{
					Type: InsightsTracesFilterTypeOr,
					Expressions: []TracesFilterExpression{
						{
							Type:     InsightsTracesFilterTypeCondition,
							Field:    &TracesFilterField{Name: "statusCode"},
							Operator: InsightsTracesFilterOperatorEqual,
							Value:    &TracesFilterValue{String: String("Error")},
						},
						{
							Type:     InsightsTracesFilterTypeCondition,
							Field:    &TracesFilterField{Name: "durationNs"},
							Operator: InsightsTracesFilterOperatorGreaterThanOrEqual,
							Value:    &TracesFilterValue{Number: PtrTo(500000000.0)},
						},
						{
							Type:     InsightsTracesFilterTypeCondition,
							Field:    &TracesFilterField{Name: "host.name", Scope: InsightsTracesFieldScopeResource},
							Operator: InsightsTracesFilterOperatorIn,
							Value:    &TracesFilterValue{StringArray: []string{"web-01", "web-02"}},
						},
					},
				},
				{
					Type: InsightsTracesFilterTypeNot,
					Expressions: []TracesFilterExpression{
						{Type: InsightsTracesFilterTypeTextSearch, Query: "healthcheck"},
					},
				},
			},
		},
		OrderBy: []TracesOrderBy{
			{Field: TracesFilterField{Name: "durationNs"}, Direction: InsightsTracesSortDirectionDescending},
			{Field: TracesFilterField{Name: "started"}, Direction: InsightsTracesSortDirectionAscending},
		},
		Pagination: &TracesPaginationRequest{Limit: 50},
	}

	setup()
	defer teardown()

	mux.HandleFunc("/v2/insights/query/nyc3/traces/search", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Errorf("Content-Type = %q", got)
		}

		var got TracesSearchRequest
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if !reflect.DeepEqual(got, *search) {
			t.Errorf("request = %#v, want %#v", got, *search)
		}

		fmt.Fprint(w, `{
			"data": [{
				"traceId": "112419ad0cdd6a78f20c44a173c594d5",
				"started": "2026-10-09T09:36:32.95Z",
				"serviceName": "checkout-api",
				"durationNs": 50000000,
				"spanCount": 3,
				"errorCount": 1,
				"status": "Error",
				"spanNames": ["GET /checkout", "SELECT orders"]
			}],
			"pagination": {"hasMore": true}
		}`)
	})

	got, resp, err := client.Insights.SearchTraces(ctx, "nyc3", search)
	if err != nil {
		t.Fatalf("SearchTraces: %v", err)
	}
	if resp == nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("response = %#v", resp)
	}

	want := &TracesSearchResponse{
		Data: []TraceSummary{
			{
				TraceID:     "112419ad0cdd6a78f20c44a173c594d5",
				Started:     time.Date(2026, 10, 9, 9, 36, 32, 950000000, time.UTC),
				ServiceName: "checkout-api",
				DurationNs:  50000000,
				SpanCount:   3,
				ErrorCount:  1,
				Status:      InsightsTraceStatusError,
				SpanNames:   []string{"GET /checkout", "SELECT orders"},
			},
		},
		Pagination: &TracesPaginationResponse{HasMore: true},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("SearchTraces = %#v, want %#v", got, want)
	}
}

func TestInsights_SearchTracesMinimalRequest(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/insights/query/nyc3/traces/search", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		want := `{"timeRange":{"from":{"relative":"15m"},"to":{"relative":"now"}}}` + "\n"
		if string(body) != want {
			t.Errorf("body = %s, want %s", body, want)
		}
		fmt.Fprint(w, `{"data":[],"pagination":{"hasMore":false}}`)
	})

	got, _, err := client.Insights.SearchTraces(ctx, "nyc3", &TracesSearchRequest{
		TimeRange: TracesTimeRange{
			From: TracesTimeBound{Relative: "15m"},
			To:   TracesTimeBound{Relative: "now"},
		},
	})
	if err != nil {
		t.Fatalf("SearchTraces: %v", err)
	}
	if len(got.Data) != 0 || got.Pagination == nil || got.Pagination.HasMore {
		t.Errorf("SearchTraces = %#v", got)
	}
}

func TestTracesFilterValueMarshal(t *testing.T) {
	tests := []struct {
		name  string
		value TracesFilterValue
		want  string
	}{
		{"string", TracesFilterValue{String: String("checkout-api")}, `{"type":"string","value":"checkout-api"}`},
		{"empty string", TracesFilterValue{String: String("")}, `{"type":"string","value":""}`},
		{"number", TracesFilterValue{Number: PtrTo(0.0)}, `{"type":"number","value":0}`},
		{"bool", TracesFilterValue{Bool: PtrTo(false)}, `{"type":"bool","value":false}`},
		{"string_array", TracesFilterValue{StringArray: []string{"web-01", "web-02"}}, `{"type":"string_array","value":["web-01","web-02"]}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.value)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			if string(got) != tt.want {
				t.Errorf("JSON = %s, want %s", got, tt.want)
			}
		})
	}

	if _, err := json.Marshal(TracesFilterValue{}); err == nil {
		t.Error("expected error marshaling empty TracesFilterValue")
	}
	if _, err := json.Marshal(TracesFilterValue{String: String("x"), Bool: PtrTo(true)}); err == nil {
		t.Error("expected error marshaling multi-field TracesFilterValue")
	}
}

func TestTracesFilterValueUnmarshal(t *testing.T) {
	tests := []struct {
		name string
		data string
		want TracesFilterValue
	}{
		{"typed string", `{"type":"string","value":"checkout-api"}`, TracesFilterValue{String: String("checkout-api")}},
		{"bare string", `"checkout-api"`, TracesFilterValue{String: String("checkout-api")}},
		{"typed number", `{"type":"number","value":500000000}`, TracesFilterValue{Number: PtrTo(500000000.0)}},
		{"bare number", `500000000`, TracesFilterValue{Number: PtrTo(500000000.0)}},
		{"typed bool", `{"type":"bool","value":true}`, TracesFilterValue{Bool: PtrTo(true)}},
		{"bare bool", `false`, TracesFilterValue{Bool: PtrTo(false)}},
		{"typed string_array", `{"type":"string_array","value":["web-01","web-02"]}`, TracesFilterValue{StringArray: []string{"web-01", "web-02"}}},
		{"bare string_array", `["web-01","web-02"]`, TracesFilterValue{StringArray: []string{"web-01", "web-02"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got TracesFilterValue
			if err := json.Unmarshal([]byte(tt.data), &got); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %#v, want %#v", got, tt.want)
			}
		})
	}

	var v TracesFilterValue
	if err := json.Unmarshal([]byte(`{"type":"date","value":"x"}`), &v); err == nil {
		t.Error("expected error for unknown value type")
	}
	if err := json.Unmarshal([]byte(`["a",1]`), &v); err == nil {
		t.Error("expected error for non-string array element")
	}
}

func TestInsights_SearchTracesEscapesRegion(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.URL.EscapedPath(), "/v2/insights/query/nyc3%2Fother/traces/search"; got != want {
			t.Errorf("escaped path = %q, want %q", got, want)
		}
		fmt.Fprint(w, `{}`)
	})

	_, _, err := client.Insights.SearchTraces(ctx, "nyc3/other", &TracesSearchRequest{
		TimeRange: TracesTimeRange{
			From: TracesTimeBound{Relative: "1h"},
			To:   TracesTimeBound{Relative: "now"},
		},
	})
	if err != nil {
		t.Fatalf("SearchTraces: %v", err)
	}
}

func TestInsights_SearchTracesError(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/insights/query/nyc3/traces/search", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "pagination.limit exceeds maximum of 1000")
	})

	got, resp, err := client.Insights.SearchTraces(ctx, "nyc3", &TracesSearchRequest{
		Pagination: &TracesPaginationRequest{Limit: 1001},
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if got != nil {
		t.Errorf("result = %#v, want nil", got)
	}
	if resp == nil || resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("response = %#v", resp)
	}
	var apiErr *ErrorResponse
	if !errors.As(err, &apiErr) || apiErr.Message != "pagination.limit exceeds maximum of 1000" {
		t.Errorf("error = %#v", err)
	}
}

func TestInsights_GetTrace(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/insights/query/nyc3/traces/05928fa10bbc6bb65b33405ea4d6c403", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		testFormValues(t, r, values{"from": "24h", "to": "now"})

		fmt.Fprint(w, `{
			"traceId": "05928fa10bbc6bb65b33405ea4d6c403",
			"spans": [
				{
					"traceId": "05928fa10bbc6bb65b33405ea4d6c403",
					"spanId": "1367b0627cbd1f04",
					"startTime": "2026-10-09T09:36:32Z",
					"endTime": "2026-10-09T09:36:33Z",
					"durationNs": 1000000000,
					"name": "GET /checkout",
					"kind": "Server",
					"statusCode": "Error",
					"statusMessage": "upstream timeout",
					"serviceName": "checkout-api",
					"resource": {"service.name": "checkout-api"},
					"attributes": {"http.method": "GET"}
				},
				{
					"traceId": "05928fa10bbc6bb65b33405ea4d6c403",
					"spanId": "a2fb4a1d1a96d312",
					"parentSpanId": "1367b0627cbd1f04",
					"startTime": "2026-10-09T09:36:32.5Z",
					"endTime": "2026-10-09T09:36:32.7Z",
					"durationNs": 200000000,
					"name": "SELECT orders",
					"kind": "Client",
					"statusCode": "Ok",
					"serviceName": "checkout-api",
					"resource": {"service.name": "checkout-api"}
				}
			],
			"truncated": true
		}`)
	})

	got, resp, err := client.Insights.GetTrace(ctx, "nyc3", "05928fa10bbc6bb65b33405ea4d6c403", &TraceGetOptions{From: "24h", To: "now"})
	if err != nil {
		t.Fatalf("GetTrace: %v", err)
	}
	if resp == nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("response = %#v", resp)
	}

	want := &Trace{
		TraceID: "05928fa10bbc6bb65b33405ea4d6c403",
		Spans: []TraceSpan{
			{
				TraceID:       "05928fa10bbc6bb65b33405ea4d6c403",
				SpanID:        "1367b0627cbd1f04",
				StartTime:     time.Date(2026, 10, 9, 9, 36, 32, 0, time.UTC),
				EndTime:       time.Date(2026, 10, 9, 9, 36, 33, 0, time.UTC),
				DurationNs:    1000000000,
				Name:          "GET /checkout",
				Kind:          "Server",
				StatusCode:    "Error",
				StatusMessage: "upstream timeout",
				ServiceName:   "checkout-api",
				Resource:      map[string]string{"service.name": "checkout-api"},
				Attributes:    map[string]string{"http.method": "GET"},
			},
			{
				TraceID:      "05928fa10bbc6bb65b33405ea4d6c403",
				SpanID:       "a2fb4a1d1a96d312",
				ParentSpanID: "1367b0627cbd1f04",
				StartTime:    time.Date(2026, 10, 9, 9, 36, 32, 500000000, time.UTC),
				EndTime:      time.Date(2026, 10, 9, 9, 36, 32, 700000000, time.UTC),
				DurationNs:   200000000,
				Name:         "SELECT orders",
				Kind:         "Client",
				StatusCode:   "Ok",
				ServiceName:  "checkout-api",
				Resource:     map[string]string{"service.name": "checkout-api"},
			},
		},
		Truncated: true,
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("GetTrace = %#v, want %#v", got, want)
	}
}

func TestInsights_GetTraceNilOptions(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/insights/query/nyc3/traces/05928fa10bbc6bb65b33405ea4d6c403", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		if r.URL.RawQuery != "" {
			t.Errorf("query = %q, want empty", r.URL.RawQuery)
		}
		fmt.Fprint(w, `{"traceId":"05928fa10bbc6bb65b33405ea4d6c403","spans":[]}`)
	})

	got, _, err := client.Insights.GetTrace(ctx, "nyc3", "05928fa10bbc6bb65b33405ea4d6c403", nil)
	if err != nil {
		t.Fatalf("GetTrace: %v", err)
	}
	if got.TraceID != "05928fa10bbc6bb65b33405ea4d6c403" || got.Truncated {
		t.Errorf("GetTrace = %#v", got)
	}
}

func TestInsights_GetTraceEscapesPath(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.URL.EscapedPath(), "/v2/insights/query/nyc3%2Fother/traces/abc%2F..%2Fdef"; got != want {
			t.Errorf("escaped path = %q, want %q", got, want)
		}
		if got, want := r.URL.Query().Get("from"), "now-2h"; got != want {
			t.Errorf("from = %q, want %q", got, want)
		}
		fmt.Fprint(w, `{}`)
	})

	_, _, err := client.Insights.GetTrace(ctx, "nyc3/other", "abc/../def", &TraceGetOptions{From: "now-2h"})
	if err != nil {
		t.Fatalf("GetTrace: %v", err)
	}
}

func TestInsights_GetTraceNotFound(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/insights/query/nyc3/traces/0000000000000000000000000000dead", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, "trace not found")
	})

	got, resp, err := client.Insights.GetTrace(ctx, "nyc3", "0000000000000000000000000000dead", nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if got != nil {
		t.Errorf("result = %#v, want nil", got)
	}
	if resp == nil || resp.StatusCode != http.StatusNotFound {
		t.Fatalf("response = %#v", resp)
	}
	var apiErr *ErrorResponse
	if !errors.As(err, &apiErr) || apiErr.Message != "trace not found" {
		t.Errorf("error = %#v", err)
	}
}
