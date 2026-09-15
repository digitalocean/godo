package godo

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"
	"time"
)

func TestPrepaymentGetConfig(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/customers/my/prepayment_config", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)

		response := `
		{
			"config": {
				"spend_limit": "100.00",
				"is_auto_prepay_enabled": true,
				"prepay_amount": "50.00",
				"prepay_threshold": "10.00",
				"created_at": "2026-06-22T19:31:51Z",
				"updated_at": "2026-06-25T18:40:07Z"
			},
			"status": {
				"balance": "25.00",
				"is_auto_prepay_enabled": true,
				"blocked": true,
				"eligible": true,
				"month_to_date_balance": "75.00"
			}
		}
		`

		fmt.Fprint(w, response)
	})

	got, _, err := client.Prepayment.GetConfig(ctx)
	if err != nil {
		t.Errorf("Prepayment.GetConfig returned error: %v", err)
	}

	expected := &PrepaymentConfigResponse{
		Config: &PrepaymentConfig{
			SpendLimit:          "100.00",
			IsAutoPrepayEnabled: true,
			PrepayAmount:        "50.00",
			PrepayThreshold:     "10.00",
			CreatedAt:           time.Date(2026, 6, 22, 19, 31, 51, 0, time.UTC),
			UpdatedAt:           time.Date(2026, 6, 25, 18, 40, 7, 0, time.UTC),
		},
		Status: &PrepaymentStatus{
			Balance:             "25.00",
			IsAutoPrepayEnabled: true,
			Blocked:             true,
			Eligible:            true,
			MonthToDateBalance:  "75.00",
		},
	}
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("Prepayment.GetConfig returned %+v, expected %+v", got, expected)
	}
}

func TestPrepaymentGetStatus(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/customers/my/prepayment_status", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)

		response := `
		{
			"status": {
				"balance": "25.00",
				"is_auto_prepay_enabled": true,
				"blocked": true,
				"eligible": true,
				"month_to_date_balance": "75.00"
			}
		}
		`

		fmt.Fprint(w, response)
	})

	got, _, err := client.Prepayment.GetStatus(ctx)
	if err != nil {
		t.Errorf("Prepayment.GetStatus returned error: %v", err)
	}

	expected := &PrepaymentStatusResponse{
		Status: &PrepaymentStatus{
			Balance:             "25.00",
			IsAutoPrepayEnabled: true,
			Blocked:             true,
			Eligible:            true,
			MonthToDateBalance:  "75.00",
		},
	}
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("Prepayment.GetStatus returned %+v, expected %+v", got, expected)
	}
}
