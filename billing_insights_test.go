package godo

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

func TestBillingInsights_List(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/billing/do:team:12345678-1234-1234-1234-123456789012/insights/2025-01-01/2025-01-31", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"current_page": 1,
			"data_points": [
				{
					"description": "droplet name (c-2-4GiB)",
					"group_description": "",
					"region": "nyc3",
					"sku": "1-DO-DROP-0109",
					"start_date": "2025-01-01",
					"total_amount": "0.86",
					"usage_team_urn": "do:team:12345678-1234-1234-1234-123456789012"
				},
				{
					"description": "3 nodes - 4 GB / 2 vCPU / 80 GB SSD",
					"group_description": "kubernetes cluster name",
					"region": "nyc3",
					"sku": "1-KS-K8SWN-00109",
					"start_date": "2025-01-01",
					"total_amount": "2.57",
					"usage_team_urn": "do:team:12345678-1234-1234-1234-123456789012"
				}
			],
			"total_items": 2,
			"total_pages": 1
		}`)
	})

	history, resp, err := client.BillingInsights.List(ctx, "do:team:12345678-1234-1234-1234-123456789012", "2025-01-01", "2025-01-31", nil)
	if err != nil {
		t.Errorf("BillingInsights.List returned error: %v", err)
	}

	expectedBillingInsights := []BillingInsightsEntry{
		{
			Description:      "droplet name (c-2-4GiB)",
			GroupDescription: "",
			Region:           "nyc3",
			SKU:              "1-DO-DROP-0109",
			StartDate:        "2025-01-01",
			TotalAmount:      "0.86",
			UsageTeamURN:     "do:team:12345678-1234-1234-1234-123456789012",
		},
		{
			Description:      "3 nodes - 4 GB / 2 vCPU / 80 GB SSD",
			GroupDescription: "kubernetes cluster name",
			Region:           "nyc3",
			SKU:              "1-KS-K8SWN-00109",
			StartDate:        "2025-01-01",
			TotalAmount:      "2.57",
			UsageTeamURN:     "do:team:12345678-1234-1234-1234-123456789012",
		},
	}
	entries := history.DataPoints
	if !reflect.DeepEqual(entries, expectedBillingInsights) {
		t.Errorf("BillingInsights.List\nBillingInsights: got=%#v\nwant=%#v", entries, expectedBillingInsights)
	}
	expectedMeta := &Meta{
		Page:  1,
		Total: 2,
		Pages: 1,
	}
	if !reflect.DeepEqual(resp.Meta, expectedMeta) {
		t.Errorf("BillingInsights.List\nMeta: got=%#v\nwant=%#v", resp.Meta, expectedMeta)
	}
}
