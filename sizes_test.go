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
			DiskInfo: []DiskInfo{
				{
					Type: "local",
					Size: &DiskSize{
						Amount: 25,
						Unit:   "gib",
					},
				},
			},
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
			DiskInfo: []DiskInfo{
				{
					Type: "local",
					Size: &DiskSize{
						Amount: 20,
						Unit:   "gib",
					},
				},
			},
		},
		{
			Slug:         "gpu-h100x8-640gb-200",
			Memory:       1966080,
			Vcpus:        160,
			Disk:         200,
			PriceMonthly: 35414.4,
			PriceHourly:  52.7,
			Regions:      []string{"tor1"},
			Available:    true,
			Transfer:     60,
			Description:  "H100 GPU - 8X (small disk)",
			GPUInfo: &GPUInfo{
				Count: 8,
				VRAM: &VRAM{
					Amount: 640,
					Unit:   "gib",
				},
				Model: "nvidia_h100",
			},
			DiskInfo: []DiskInfo{
				{
					Type: "local",
					Size: &DiskSize{
						Amount: 200,
						Unit:   "gib",
					},
				},
				{
					Type: "scratch",
					Size: &DiskSize{
						Amount: 40960,
						Unit:   "gib",
					},
				},
			},
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
					"description": "Basic",
					"disk_info": [
						{
							"type": "local",
							"size": {
								"amount": 25,
								"unit": "gib"
							}
						}
					]
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
					"description": "Legacy Basic",
					"disk_info": [
						{
							"type": "local",
							"size": {
								"amount": 20,
								"unit": "gib"
							}
						}
					]
				},
				{
					"slug": "gpu-h100x8-640gb-200",
					"memory": 1966080,
					"vcpus": 160,
					"disk": 200,
					"transfer": 60,
					"price_monthly": 35414.4,
					"price_hourly": 52.7,
					"regions": [
						"tor1"
					],
					"available": true,
					"description": "H100 GPU - 8X (small disk)",
					"networking_throughput": 10000,
					"gpu_info": {
						"count": 8,
						"vram": {
							"amount": 640,
							"unit": "gib"
						},
						"model": "nvidia_h100"
					},
					"disk_info": [
						{
							"type": "local",
							"size": {
								"amount": 200,
								"unit": "gib"
							}
						},
						{
							"type": "scratch",
							"size": {
								"amount": 40960,
								"unit": "gib"
							}
						}
					]
				}
			],
			"meta": {
				"total": 3
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

	expectedMeta := &Meta{Total: 3}
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
