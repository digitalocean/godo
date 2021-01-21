package godo

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

func TestSizes_List(t *testing.T) {
	setup()
	defer teardown()

	expectedSizes := []Size{
		{
			Slug:         "s-1vcpu-1gb",
			Memory:       1024,
			Vcpus:        1,
			Disk:         25,
			PriceMonthly: 5,
			PriceHourly:  0.00744,
			Regions:      []string{"nyc1", "nyc2"},
			Available:    true,
			Transfer:     1,
			Description:  "Basic",
		},
		{
			Slug:         "512mb",
			Memory:       512,
			Vcpus:        1,
			Disk:         20,
			PriceMonthly: 5,
			PriceHourly:  0.00744,
			Regions:      []string{"nyc1", "nyc2"},
			Available:    true,
			Transfer:     1,
			Description:  "Legacy Basic",
		},
	}

	mux.HandleFunc("/v2/sizes", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"sizes": [
				{
					"slug": "s-1vcpu-1gb",
					"memory": 1024,
					"vcpus": 1,
					"disk": 25,
					"transfer": 1,
					"price_monthly": 5,
					"price_hourly": 0.00744,
					"regions": [
						"nyc1",
						"nyc2"
					],
					"available": true,
					"description": "Basic"
				},
				{
					"slug": "512mb",
					"memory": 512,
					"vcpus": 1,
					"disk": 20,
					"transfer": 1,
					"price_monthly": 5,
					"price_hourly": 0.00744,
					"regions": [
						"nyc1",
						"nyc2"
					],
					"available": true,
					"description": "Legacy Basic"
				}
			],
			"meta": {
				"total": 2
			}
		}`)
	})

	sizes, resp, err := client.Sizes.List(ctx, nil)
	if err != nil {
		t.Errorf("Sizes.List returned error: %v", err)
	}

	if !reflect.DeepEqual(sizes, expectedSizes) {
		t.Errorf("Sizes.List returned sizes %+v, expected %+v", sizes, expectedSizes)
	}

	expectedMeta := &Meta{Total: 2}
	if !reflect.DeepEqual(resp.Meta, expectedMeta) {
		t.Errorf("Sizes.List returned meta %+v, expected %+v", resp.Meta, expectedMeta)
	}
}

func TestSizes_ListSizesMultiplePages(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/sizes", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{"sizes": [{"id":1},{"id":2}], "links":{"pages":{"next":"http://example.com/v2/sizes/?page=2"}}}`)
	})

	_, resp, err := client.Sizes.List(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}

	checkCurrentPage(t, resp, 1)
}

func TestSizes_RetrievePageByNumber(t *testing.T) {
	setup()
	defer teardown()

	jBlob := `
	{
		"sizes": [{"id":1},{"id":2}],
		"links":{
			"pages":{
				"next":"http://example.com/v2/sizes/?page=3",
				"prev":"http://example.com/v2/sizes/?page=1",
				"last":"http://example.com/v2/sizes/?page=3",
				"first":"http://example.com/v2/sizes/?page=1"
			}
		}
	}`

	mux.HandleFunc("/v2/sizes", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, jBlob)
	})

	opt := &ListOptions{Page: 2}
	_, resp, err := client.Sizes.List(ctx, opt)
	if err != nil {
		t.Fatal(err)
	}

	checkCurrentPage(t, resp, 2)
}

func TestSize_String(t *testing.T) {
	size := &Size{
		Slug:         "slize",
		Memory:       123,
		Vcpus:        456,
		Disk:         789,
		PriceMonthly: 123,
		PriceHourly:  456,
		Regions:      []string{"1", "2"},
		Available:    true,
		Transfer:     789,
		Description:  "Basic",
	}

	stringified := size.String()
	expected := `godo.Size{Slug:"slize", Memory:123, Vcpus:456, Disk:789, PriceMonthly:123, PriceHourly:456, Regions:["1" "2"], Available:true, Transfer:789, Description:"Basic"}`
	if expected != stringified {
		t.Errorf("Size.String returned %+v, expected %+v", stringified, expected)
	}
}
