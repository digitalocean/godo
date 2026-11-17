// Smoke the Insights client against a live API.
//
//	DIGITALOCEAN_TOKEN=... go run ./examples/insights/smoke
//
// Optional:
//
//	DIGITALOCEAN_API_URL   API base override, including trailing slash (staging)
//	INSIGHTS_REGION        run PromQL, logs, spans, and traces calls in this region, for example nyc3
//	INSIGHTS_QUERY         PromQL expression (default do.droplets.cpu_time)
//	INSIGHTS_SMOKE_WRITE=1 create, update, then delete a dashboard; and create, then
//	                       delete, an email channel and a paused alert rule
//	INSIGHTS_EMAIL         verified team email, required for the channel and alert
//	                       rule writes
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"time"

	"github.com/digitalocean/godo"
)

// dashboardSpec builds a minimal single-panel Perses dashboard document.
func dashboardSpec(name, description string) json.RawMessage {
	doc := map[string]any{
		"kind":       "Dashboard",
		"apiVersion": "perses.dev/v1alpha1",
		"spec": map[string]any{
			"display":   map[string]any{"name": name, "description": description},
			"duration":  "1h",
			"variables": []any{},
			"layouts": []any{map[string]any{"kind": "Grid", "spec": map[string]any{"items": []any{
				map[string]any{"x": 0, "y": 0, "width": 12, "height": 6, "content": map[string]any{"$ref": "#/spec/panels/cpu"}},
			}}}},
			"panels": map[string]any{"cpu": map[string]any{"kind": "Panel", "spec": map[string]any{
				"display": map[string]any{"name": "CPU utilization"},
				"plugin":  map[string]any{"kind": "TimeSeriesChart", "spec": map[string]any{}},
				"queries": []any{map[string]any{"kind": "TimeSeriesQuery", "spec": map[string]any{
					"plugin": map[string]any{"kind": "MetricsTimeSeriesQuery", "spec": map[string]any{
						"metricName": "do.droplets.cpu_utilization", "kind": "gauge", "filters": []any{}, "groupBy": []any{},
					}},
				}}},
			}}},
		},
	}
	b, err := json.Marshal(doc)
	if err != nil {
		panic(err)
	}
	return b
}

// sameJSON reports whether two JSON documents are equal ignoring key order and
// whitespace. The API stores specs in a JSON column that normalizes both.
func sameJSON(a, b json.RawMessage) bool {
	var x, y any
	if json.Unmarshal(a, &x) != nil || json.Unmarshal(b, &y) != nil {
		return false
	}
	return reflect.DeepEqual(x, y)
}

func main() {
	token := os.Getenv("DIGITALOCEAN_TOKEN")
	if token == "" {
		fmt.Fprintln(os.Stderr, "DIGITALOCEAN_TOKEN is required")
		os.Exit(2)
	}
	client := godo.NewFromToken(token)
	if base := os.Getenv("DIGITALOCEAN_API_URL"); base != "" {
		u, err := url.Parse(base)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
		client.BaseURL = u
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	failed := 0

	check := func(name string, fn func() error) {
		if err := fn(); err != nil {
			failed++
			fmt.Printf("FAIL %s: %v\n", name, err)
			return
		}
		fmt.Printf("OK   %s\n", name)
	}

	check("ListNotificationChannels", func() error {
		channels, resp, err := client.Insights.ListNotificationChannels(ctx, &godo.ListOptions{Page: 1, PerPage: 20})
		if err != nil {
			return err
		}
		fmt.Printf("       → %d channel(s)", len(channels))
		if resp.Meta != nil {
			fmt.Printf(", total %d", resp.Meta.Total)
		}
		fmt.Println()
		return nil
	})

	check("ListAlertRules", func() error {
		rules, _, err := client.Insights.ListAlertRules(ctx, &godo.AlertRuleListOptions{PerPage: 20})
		if err != nil {
			return err
		}
		fmt.Printf("       → %d rule(s)\n", len(rules))
		return nil
	})

	check("ListAlertInstances", func() error {
		instances, _, err := client.Insights.ListAlertInstances(ctx, &godo.AlertInstanceListOptions{
			Status:  godo.InsightsAlertInstanceStatusActive,
			PerPage: 20,
		})
		if err != nil {
			return err
		}
		fmt.Printf("       → %d active instance(s)\n", len(instances))
		return nil
	})

	check("ListDashboards", func() error {
		out, _, err := client.Insights.ListDashboards(ctx, &godo.DashboardListOptions{PerPage: 20})
		if err != nil {
			return err
		}
		fmt.Printf("       → %d dashboard(s)", len(out.Dashboards))
		if out.Pagination != nil {
			fmt.Printf(", total %d", out.Pagination.Total)
		}
		fmt.Println()
		return nil
	})

	check("GetDashboardSchema", func() error {
		schema, _, err := client.Insights.GetDashboardSchema(ctx)
		if err != nil {
			return err
		}
		var doc map[string]any
		if err := json.Unmarshal(schema, &doc); err != nil {
			return err
		}
		if _, ok := doc["$schema"]; !ok {
			return fmt.Errorf("body is not an unwrapped JSON Schema document")
		}
		fmt.Printf("       → %d bytes, title=%v\n", len(schema), doc["title"])
		return nil
	})

	region := os.Getenv("INSIGHTS_REGION")
	if region == "" {
		fmt.Println("SKIP PromQL (set INSIGHTS_REGION)")
	} else {
		query := os.Getenv("INSIGHTS_QUERY")
		if query == "" {
			query = "do.droplets.cpu_time"
		}
		check("Query", func() error {
			out, _, err := client.Insights.Query(ctx, region, &godo.PromQueryOptions{Query: query})
			if err != nil {
				return err
			}
			fmt.Printf("       → resultType=%s samples=%d\n", out.ResultType, len(out.Vector)+len(out.Matrix))
			return nil
		})
		check("SearchLogs", func() error {
			out, _, err := client.Insights.SearchLogs(ctx, region, &godo.LogsSearchRequest{
				TimeRange: godo.LogsTimeRange{
					From: godo.LogsTimeInstant{Relative: "5m"},
					To:   godo.LogsTimeInstant{Relative: "now"},
				},
				OrderBy: []godo.LogsOrderBy{{
					Field:     godo.LogsFieldRef{Name: "timestamp"},
					Direction: godo.InsightsLogsSortDirectionDescending,
				}},
				Pagination: &godo.LogsPaginationRequest{Limit: 10},
			})
			if err != nil {
				return err
			}
			fmt.Printf("       → %d log record(s)\n", len(out.Data))
			return nil
		})
		check("SearchSpans", func() error {
			out, _, err := client.Insights.SearchSpans(ctx, region, &godo.SpansSearchRequest{
				TimeRange: godo.SpansTimeRange{
					From: godo.SpansTimeBound{Relative: "5m"},
					To:   godo.SpansTimeBound{Relative: "now"},
				},
				OrderBy: []godo.SpansOrderBy{{
					Field:     godo.SpansFilterField{Name: "start_time"},
					Direction: godo.InsightsSpansSortDirectionDescending,
				}},
				Pagination: &godo.SpansPaginationRequest{Limit: 10},
			})
			if err != nil {
				return err
			}
			fmt.Printf("       → %d span(s)\n", len(out.Data))
			return nil
		})
		var traceID string
		check("SearchTraces", func() error {
			out, _, err := client.Insights.SearchTraces(ctx, region, &godo.TracesSearchRequest{
				TimeRange: godo.TracesTimeRange{
					From: godo.TracesTimeBound{Relative: "1h"},
					To:   godo.TracesTimeBound{Relative: "now"},
				},
				OrderBy: []godo.TracesOrderBy{{
					Field:     godo.TracesFilterField{Name: "started"},
					Direction: godo.InsightsTracesSortDirectionDescending,
				}},
				Pagination: &godo.TracesPaginationRequest{Limit: 10},
			})
			if err != nil {
				return err
			}
			if len(out.Data) > 0 {
				traceID = out.Data[0].TraceID
			}
			fmt.Printf("       → %d trace(s)\n", len(out.Data))
			return nil
		})
		if traceID == "" {
			fmt.Println("SKIP GetTrace (no traces in the last hour)")
		} else {
			check("GetTrace", func() error {
				out, _, err := client.Insights.GetTrace(ctx, region, traceID, &godo.TraceGetOptions{From: "1h", To: "now"})
				if err != nil {
					return err
				}
				fmt.Printf("       → %d span(s), truncated=%t\n", len(out.Spans), out.Truncated)
				return nil
			})
		}
		check("PostQuery", func() error {
			_, _, err := client.Insights.PostQuery(ctx, region, &godo.PromQueryOptions{Query: query, Timeout: "30s"})
			return err
		})
		check("QueryRange", func() error {
			now := time.Now().UTC()
			out, _, err := client.Insights.QueryRange(ctx, region, &godo.PromQueryRangeOptions{
				Query: query,
				Start: now.Add(-5 * time.Minute).Format(time.RFC3339),
				End:   now.Format(time.RFC3339),
				Step:  "60s",
			})
			if err != nil {
				return err
			}
			fmt.Printf("       → series=%d\n", len(out.Data.Result))
			return nil
		})
		check("Labels", func() error {
			out, _, err := client.Insights.Labels(ctx, region, nil)
			if err != nil {
				return err
			}
			fmt.Printf("       → %d label(s)\n", len(out.Data))
			return nil
		})
		check("LabelValues", func() error {
			out, _, err := client.Insights.LabelValues(ctx, region, "__name__", nil)
			if err != nil {
				return err
			}
			fmt.Printf("       → %d value(s)\n", len(out.Data))
			return nil
		})
		check("Series", func() error {
			out, _, err := client.Insights.Series(ctx, region, &godo.PromSelectorOptions{
				Match: []string{`{__name__="do.droplets.cpu_time"}`},
				Start: time.Now().UTC().Add(-5 * time.Minute).Format(time.RFC3339),
				End:   time.Now().UTC().Format(time.RFC3339),
			})
			if err != nil {
				return err
			}
			fmt.Printf("       → %d series\n", len(out.Data))
			return nil
		})
	}

	if os.Getenv("INSIGHTS_SMOKE_WRITE") != "1" {
		fmt.Println("SKIP write (set INSIGHTS_SMOKE_WRITE=1, and INSIGHTS_EMAIL for channels and alert rules)")
	} else {
		var dashboardUUID string
		dashboardName := fmt.Sprintf("godo insights smoke %d", time.Now().Unix())
		spec := dashboardSpec(dashboardName, "Created by the godo Insights smoke example. Safe to delete.")
		check("CreateDashboard", func() error {
			d, resp, err := client.Insights.CreateDashboard(ctx, &godo.DashboardRequest{Spec: spec})
			if err != nil {
				return err
			}
			dashboardUUID = d.UUID // record first so cleanup runs even if a check below fails
			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("status %d, want 200", resp.StatusCode)
			}
			if d.Name != dashboardName {
				return fmt.Errorf("name %q, want %q", d.Name, dashboardName)
			}
			if !sameJSON(d.Spec, spec) {
				return fmt.Errorf("spec differs from the request: %s", d.Spec)
			}
			fmt.Printf("       → uuid=%s name=%q\n", d.UUID, d.Name)
			return nil
		})
		if dashboardUUID != "" {
			check("GetDashboard", func() error {
				d, _, err := client.Insights.GetDashboard(ctx, dashboardUUID)
				if err != nil {
					return err
				}
				if d.UUID != dashboardUUID {
					return fmt.Errorf("got uuid %q", d.UUID)
				}
				if !sameJSON(d.Spec, spec) {
					return fmt.Errorf("spec differs from the request: %s", d.Spec)
				}
				return nil
			})
			check("ListDashboards sorted by name", func() error {
				out, _, err := client.Insights.ListDashboards(ctx, &godo.DashboardListOptions{
					PerPage:       200,
					Sort:          godo.InsightsDashboardSortName,
					SortDirection: godo.InsightsDashboardSortAscending,
				})
				if err != nil {
					return err
				}
				for _, d := range out.Dashboards {
					if d.UUID == dashboardUUID {
						return nil
					}
				}
				return fmt.Errorf("dashboard %s not in the first %d", dashboardUUID, len(out.Dashboards))
			})
			check("CreateDashboard duplicate name is 409", func() error {
				d, resp, err := client.Insights.CreateDashboard(ctx, &godo.DashboardRequest{Spec: spec})
				if err == nil {
					_, _ = client.Insights.DeleteDashboard(ctx, d.UUID)
					return fmt.Errorf("expected an error")
				}
				if resp == nil || resp.StatusCode != http.StatusConflict {
					return fmt.Errorf("got %v", err)
				}
				return nil
			})
			check("UpdateDashboard", func() error {
				updated := dashboardSpec(dashboardName+" updated", "Updated by the godo Insights smoke example.")
				d, _, err := client.Insights.UpdateDashboard(ctx, dashboardUUID, &godo.DashboardRequest{Spec: updated})
				if err != nil {
					return err
				}
				if d.Name != dashboardName+" updated" {
					return fmt.Errorf("name %q not replaced", d.Name)
				}
				if !sameJSON(d.Spec, updated) {
					return fmt.Errorf("spec not replaced: %s", d.Spec)
				}
				fmt.Printf("       → name=%q\n", d.Name)
				return nil
			})
			check("DeleteDashboard", func() error {
				resp, err := client.Insights.DeleteDashboard(ctx, dashboardUUID)
				if err != nil {
					return err
				}
				if resp.StatusCode != http.StatusOK {
					return fmt.Errorf("status %d, want 200", resp.StatusCode)
				}
				return nil
			})
			check("GetDashboard after delete is 404", func() error {
				_, resp, err := client.Insights.GetDashboard(ctx, dashboardUUID)
				if err == nil {
					return fmt.Errorf("expected an error")
				}
				if resp == nil || resp.StatusCode != http.StatusNotFound {
					return fmt.Errorf("got %v", err)
				}
				return nil
			})
		}
	}

	email := os.Getenv("INSIGHTS_EMAIL")
	switch {
	case os.Getenv("INSIGHTS_SMOKE_WRITE") != "1":
		// reported above
	case email == "":
		fmt.Println("SKIP channel and alert rule write (INSIGHTS_EMAIL is required)")
	default:
		var channelID, ruleID string
		check("CreateNotificationChannel", func() error {
			ch, _, err := client.Insights.CreateNotificationChannel(ctx, &godo.NotificationChannelRequest{
				Name:  "godo insights smoke",
				Email: &godo.EmailNotificationConfig{To: email},
			})
			if err != nil {
				return err
			}
			channelID = ch.ID
			fmt.Printf("       → id=%s type=%s\n", ch.ID, ch.ChannelType)
			return nil
		})
		if channelID != "" {
			check("GetNotificationChannel", func() error {
				ch, _, err := client.Insights.GetNotificationChannel(ctx, channelID)
				if err != nil {
					return err
				}
				if ch.ID != channelID {
					return fmt.Errorf("got id %q", ch.ID)
				}
				return nil
			})
			critical := 1e9
			bindings := []godo.NotificationChannelBinding{{
				NotificationChannelID: channelID,
				NotifyOn:              []string{godo.InsightsSeverityCritical},
			}}
			check("CreateAlertRule", func() error {
				rule, _, err := client.Insights.CreateAlertRule(ctx, &godo.AlertRuleRequest{
					Status: godo.InsightsAlertRuleStatusPaused,
					Spec: godo.AlertRuleSpec{
						Name: "godo insights smoke",
						Query: godo.AlertMetricsQuery{
							Metric: "do.droplets.cpu_utilization",
						},
						Condition:            &godo.AlertCondition{Window: godo.InsightsEvaluationWindow5m},
						Thresholds:           godo.AlertThresholds{Critical: &critical, Operator: godo.InsightsThresholdOperatorGreaterThan},
						NotificationChannels: &bindings,
						ReAlertDuration:      godo.InsightsReAlertDurationNever,
					},
				})
				if err != nil {
					return err
				}
				ruleID = rule.ID
				fmt.Printf("       → id=%s status=%s\n", rule.ID, rule.Status)
				return nil
			})
			if ruleID != "" {
				check("DeleteAlertRule", func() error {
					_, err := client.Insights.DeleteAlertRule(ctx, ruleID)
					return err
				})
			}
			check("DeleteNotificationChannel", func() error {
				_, err := client.Insights.DeleteNotificationChannel(ctx, channelID)
				return err
			})
		}
	}

	if failed > 0 {
		fmt.Printf("%d failed\n", failed)
		os.Exit(1)
	}
}
