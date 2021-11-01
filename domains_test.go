package godo

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDomains_ListDomains(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/domains", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"domains": [
				{
					"name":"foo.com"
				},
				{
					"name":"bar.com"
				}
			],
			"meta": {
				"total": 2
			}
		}`)
	})

	domains, resp, err := client.Domains.List(ctx, nil)
	require.NoError(t, err)

	expectedDomains := []Domain{{Name: "foo.com"}, {Name: "bar.com"}}
	assert.Equal(t, expectedDomains, domains)

	expectedMeta := &Meta{Total: 2}
	assert.Equal(t, expectedMeta, resp.Meta)
}

func TestDomains_ListDomainsMultiplePages(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/domains", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{"domains": [{"id":1},{"id":2}], "links":{"pages":{"next":"http://example.com/v2/domains/?page=2"}}}`)
	})

	_, resp, err := client.Domains.List(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}

	checkCurrentPage(t, resp, 1)
}

func TestDomains_RetrievePageByNumber(t *testing.T) {
	setup()
	defer teardown()

	jBlob := `
	{
		"domains": [{"id":1},{"id":2}],
		"links":{
			"pages":{
				"next":"http://example.com/v2/domains/?page=3",
				"prev":"http://example.com/v2/domains/?page=1",
				"last":"http://example.com/v2/domains/?page=3",
				"first":"http://example.com/v2/domains/?page=1"
			}
		},
		"meta":{
			"total":2
		}
	}`

	mux.HandleFunc("/v2/domains", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, jBlob)
	})

	opt := &ListOptions{Page: 2}
	_, resp, err := client.Domains.List(ctx, opt)
	if err != nil {
		t.Fatal(err)
	}

	checkCurrentPage(t, resp, 2)
}

func TestDomains_GetDomain(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/domains/example.com", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{"domain":{"name":"example.com"}}`)
	})

	domains, _, err := client.Domains.Get(ctx, "example.com")
	require.NoError(t, err)

	expected := &Domain{Name: "example.com"}
	assert.Equal(t, expected, domains)
}

func TestDomains_Create(t *testing.T) {
	setup()
	defer teardown()

	createRequest := &DomainCreateRequest{
		Name:      "example.com",
		IPAddress: "127.0.0.1",
	}

	mux.HandleFunc("/v2/domains", func(w http.ResponseWriter, r *http.Request) {
		v := new(DomainCreateRequest)
		err := json.NewDecoder(r.Body).Decode(v)
		if err != nil {
			t.Fatal(err)
		}

		testMethod(t, r, http.MethodPost)
		assert.Equal(t, createRequest, v)

		fmt.Fprint(w, `{"domain":{"name":"example.com"}}`)
	})

	domain, _, err := client.Domains.Create(ctx, createRequest)
	require.NoError(t, err)

	expected := &Domain{Name: "example.com"}
	assert.Equal(t, expected, domain)
}

func TestDomains_Destroy(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/domains/example.com", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
	})

	_, err := client.Domains.Delete(ctx, "example.com")

	assert.NoError(t, err)
}

func TestDomains_AllRecordsForDomainName(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/domains/example.com/records", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{"domain_records":[{"id":1},{"id":2}]}`)
	})

	records, _, err := client.Domains.Records(ctx, "example.com", nil)
	require.NoError(t, err)

	expected := []DomainRecord{{ID: 1}, {ID: 2}}
	assert.Equal(t, expected, records)
}

func TestDomains_AllRecordsForDomainName_PerPage(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/domains/example.com/records", func(w http.ResponseWriter, r *http.Request) {
		perPage := r.URL.Query().Get("per_page")
		if perPage != "2" {
			t.Fatalf("expected '2', got '%s'", perPage)
		}

		fmt.Fprint(w, `{"domain_records":[{"id":1},{"id":2}]}`)
	})

	dro := &ListOptions{PerPage: 2}
	records, _, err := client.Domains.Records(ctx, "example.com", dro)
	require.NoError(t, err)

	expected := []DomainRecord{{ID: 1}, {ID: 2}}
	assert.Equal(t, expected, records)
}

func TestDomains_RecordsByType(t *testing.T) {
	tests := []struct {
		name        string
		recordType  string
		pagination  *ListOptions
		expectedErr *ArgError
	}{
		{
			name:       "success",
			recordType: "CNAME",
		},
		{
			name:        "when record type is empty it returns argument error",
			expectedErr: &ArgError{arg: "type", reason: "cannot be an empty string"},
		},
		{
			name:       "with pagination",
			recordType: "CNAME",
			pagination: &ListOptions{Page: 1, PerPage: 10},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setup()
			defer teardown()

			mux.HandleFunc("/v2/domains/example.com/records", func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, tt.recordType, r.URL.Query().Get("type"))
				if tt.pagination != nil {
					require.Equal(t, strconv.Itoa(tt.pagination.Page), r.URL.Query().Get("page"))
					require.Equal(t, strconv.Itoa(tt.pagination.PerPage), r.URL.Query().Get("per_page"))
				}
				testMethod(t, r, http.MethodGet)
				fmt.Fprint(w, `{"domain_records":[{"id":1},{"id":2}]}`)
			})

			records, _, err := client.Domains.RecordsByType(ctx, "example.com", tt.recordType, tt.pagination)
			if tt.expectedErr != nil {
				assert.Equal(t, tt.expectedErr, err)
			} else {
				expected := []DomainRecord{{ID: 1}, {ID: 2}}
				assert.Equal(t, expected, records)
			}
		})
	}
}

func TestDomains_RecordsByName(t *testing.T) {
	tests := []struct {
		name        string
		recordName  string
		pagination  *ListOptions
		expectedErr *ArgError
	}{
		{
			name:       "success",
			recordName: "foo.com",
		},
		{
			name:        "when record name is empty it returns argument error",
			expectedErr: &ArgError{arg: "name", reason: "cannot be an empty string"},
		},
		{
			name:       "with pagination",
			recordName: "foo.com",
			pagination: &ListOptions{Page: 2, PerPage: 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setup()
			defer teardown()

			mux.HandleFunc("/v2/domains/example.com/records", func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, tt.recordName, r.URL.Query().Get("name"))
				if tt.pagination != nil {
					require.Equal(t, strconv.Itoa(tt.pagination.Page), r.URL.Query().Get("page"))
					require.Equal(t, strconv.Itoa(tt.pagination.PerPage), r.URL.Query().Get("per_page"))
				}
				testMethod(t, r, http.MethodGet)
				fmt.Fprint(w, `{"domain_records":[{"id":1},{"id":2}]}`)
			})

			records, _, err := client.Domains.RecordsByName(ctx, "example.com", tt.recordName, tt.pagination)
			if tt.expectedErr != nil {
				assert.Equal(t, tt.expectedErr, err)
			} else {
				expected := []DomainRecord{{ID: 1}, {ID: 2}}
				assert.Equal(t, expected, records)
			}
		})
	}
}

func TestDomains_RecordsByTypeAndName(t *testing.T) {
	tests := []struct {
		name        string
		recordType  string
		recordName  string
		pagination  *ListOptions
		expectedErr *ArgError
	}{
		{
			name:       "success",
			recordType: "NS",
			recordName: "foo.com",
		},
		{
			name:        "when record type is empty it returns argument error",
			recordName:  "foo.com",
			expectedErr: &ArgError{arg: "type", reason: "cannot be an empty string"},
		},
		{
			name:        "when record name is empty it returns argument error",
			recordType:  "NS",
			expectedErr: &ArgError{arg: "name", reason: "cannot be an empty string"},
		},
		{
			name:       "with pagination",
			recordType: "CNAME",
			recordName: "foo.com",
			pagination: &ListOptions{Page: 1, PerPage: 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setup()
			defer teardown()

			mux.HandleFunc("/v2/domains/example.com/records", func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, tt.recordType, r.URL.Query().Get("type"))
				require.Equal(t, tt.recordName, r.URL.Query().Get("name"))
				if tt.pagination != nil {
					require.Equal(t, strconv.Itoa(tt.pagination.Page), r.URL.Query().Get("page"))
					require.Equal(t, strconv.Itoa(tt.pagination.PerPage), r.URL.Query().Get("per_page"))
				}
				testMethod(t, r, http.MethodGet)
				fmt.Fprint(w, `{"domain_records":[{"id":1},{"id":2}]}`)
			})

			records, _, err := client.Domains.RecordsByTypeAndName(ctx, "example.com", tt.recordType, tt.recordName, tt.pagination)
			if tt.expectedErr != nil {
				assert.Equal(t, tt.expectedErr, err)
			} else {
				expected := []DomainRecord{{ID: 1}, {ID: 2}}
				assert.Equal(t, expected, records)
			}
		})
	}
}

func TestDomains_GetRecordforDomainName(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/domains/example.com/records/1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{"domain_record":{"id":1}}`)
	})

	record, _, err := client.Domains.Record(ctx, "example.com", 1)
	require.NoError(t, err)

	expected := &DomainRecord{ID: 1}
	assert.Equal(t, expected, record)
}

func TestDomains_DeleteRecordForDomainName(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/domains/example.com/records/1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
	})

	_, err := client.Domains.DeleteRecord(ctx, "example.com", 1)

	assert.NoError(t, err)
}

func TestDomains_CreateRecordForDomainName(t *testing.T) {
	setup()
	defer teardown()

	createRequest := &DomainRecordEditRequest{
		Type:     "CNAME",
		Name:     "example",
		Data:     "@",
		Priority: 10,
		Port:     10,
		TTL:      1800,
		Weight:   10,
		Flags:    1,
		Tag:      "test",
	}

	mux.HandleFunc("/v2/domains/example.com/records",
		func(w http.ResponseWriter, r *http.Request) {
			v := new(DomainRecordEditRequest)
			err := json.NewDecoder(r.Body).Decode(v)

			require.NoError(t, err)

			testMethod(t, r, http.MethodPost)
			assert.Equal(t, createRequest, v)

			fmt.Fprintf(w, `{"domain_record": {"id":1}}`)
		})

	record, _, err := client.Domains.CreateRecord(ctx, "example.com", createRequest)
	require.NoError(t, err)

	expected := &DomainRecord{ID: 1}
	assert.Equal(t, expected, record)
}

func TestDomains_EditRecordForDomainName(t *testing.T) {
	setup()
	defer teardown()

	editRequest := &DomainRecordEditRequest{
		Type:     "CNAME",
		Name:     "example",
		Data:     "@",
		Priority: 10,
		Port:     10,
		TTL:      1800,
		Weight:   10,
		Flags:    1,
		Tag:      "test",
	}

	mux.HandleFunc("/v2/domains/example.com/records/1", func(w http.ResponseWriter, r *http.Request) {
		v := new(DomainRecordEditRequest)
		err := json.NewDecoder(r.Body).Decode(v)
		if err != nil {
			t.Fatalf("decode json: %v", err)
		}

		testMethod(t, r, http.MethodPut)
		assert.Equal(t, editRequest, v)

		fmt.Fprintf(w, `{"domain_record": {"id":1, "type": "CNAME", "name": "example"}}`)
	})

	record, _, err := client.Domains.EditRecord(ctx, "example.com", 1, editRequest)
	require.NoError(t, err)

	expected := &DomainRecord{ID: 1, Type: "CNAME", Name: "example"}
	assert.Equal(t, expected, record)
}

func TestDomainRecord_String(t *testing.T) {
	record := &DomainRecord{
		ID:       1,
		Type:     "CNAME",
		Name:     "example",
		Data:     "@",
		Priority: 10,
		Port:     10,
		TTL:      1800,
		Weight:   10,
		Flags:    1,
		Tag:      "test",
	}

	stringified := record.String()
	expected := `godo.DomainRecord{ID:1, Type:"CNAME", Name:"example", Data:"@", Priority:10, Port:10, TTL:1800, Weight:10, Flags:1, Tag:"test"}`
	assert.Equal(t, expected, stringified)
}

func TestDomainRecordEditRequest_String(t *testing.T) {
	record := &DomainRecordEditRequest{
		Type:     "CNAME",
		Name:     "example",
		Data:     "@",
		Priority: 10,
		Port:     10,
		TTL:      1800,
		Weight:   10,
		Flags:    1,
		Tag:      "test",
	}

	stringified := record.String()
	expected := `godo.DomainRecordEditRequest{Type:"CNAME", Name:"example", Data:"@", Priority:10, Port:10, TTL:1800, Weight:10, Flags:1, Tag:"test"}`
	assert.Equal(t, expected, stringified)
}
