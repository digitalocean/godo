package godo

import (
	"context"
	"fmt"
	"net/http"
)

const billingInsightsBasePath = "v2/billing"

// BillingInsightsService is an interface for interfacing with the BillingInsights
// endpoint of the DigitalOcean API
// See: https://docs.digitalocean.com/reference/api/reference/billing/#billingInsights_list
type BillingInsightsService interface {
	List(ctx context.Context, accountURN, startDate, endDate string, opt *ListOptions) (*BillingInsights, *Response, error)
}

// BillingInsightsServiceOp handles communication with the BillingInsights related methods of
// the DigitalOcean API.
type BillingInsightsServiceOp struct {
	client *Client
}

var _ BillingInsightsService = &BillingInsightsServiceOp{}

// BillingInsights represents a DigitalOcean Billing Insight
type BillingInsights struct {
	DataPoints  []BillingInsightsEntry `json:"data_points"`
	CurrentPage int                    `json:"current_page,string"`
	TotalItems  int                    `json:"total_items,string"`
	TotalPages  int                    `json:"total_pages,string"`
}

// BillingInsightsEntry represents an entry in a customer's Billing Insights
type BillingInsightsEntry struct {
	Description      string `json:"description"`
	GroupDescription string `json:"group_description"`
	Region           string `json:"region"`
	SKU              string `json:"sku"`
	StartDate        string `json:"start_date"` // Explicitly string-based YYYY-MM-DD, not full timestamp
	TotalAmount      string `json:"total_amount"`
	UsageTeamURN     string `json:"usage_team_urn"`
}

func (b BillingInsights) String() string {
	return Stringify(b)
}

// List the Billing Insights for a customer
func (s *BillingInsightsServiceOp) List(ctx context.Context, accountURN, startDate, endDate string, opt *ListOptions) (*BillingInsights, *Response, error) {
	path := fmt.Sprintf("%s/%s/insights/%s/%s", billingInsightsBasePath, accountURN, startDate, endDate)
	path, err := addOptions(path, opt)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	root := new(BillingInsights)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}

	resp.Meta = &Meta{
		Page:  root.CurrentPage,
		Pages: root.TotalPages,
		Total: root.TotalItems,
	}

	return root, resp, err
}
