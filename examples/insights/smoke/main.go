// Smoke the Insights client against a live API.
//
//	DIGITALOCEAN_TOKEN=... go run ./examples/insights/smoke
//
// Optional:
//
//	DIGITALOCEAN_API_URL   API base override, including trailing slash (staging)
//	INSIGHTS_REGION        run PromQL calls in this region, for example nyc3
//	INSIGHTS_QUERY         PromQL expression (default do.droplets.cpu_time)
//	INSIGHTS_SMOKE_WRITE=1 create, then delete, an email channel and a paused alert rule
//	INSIGHTS_EMAIL         verified team email, required with INSIGHTS_SMOKE_WRITE=1
package main

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/digitalocean/godo"
)

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
		fmt.Println("SKIP write (set INSIGHTS_SMOKE_WRITE=1 and INSIGHTS_EMAIL)")
	} else if email := os.Getenv("INSIGHTS_EMAIL"); email == "" {
		fmt.Println("SKIP write (INSIGHTS_EMAIL is required)")
	} else {
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
