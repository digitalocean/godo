package godo

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"testing"
	"time"
)

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
