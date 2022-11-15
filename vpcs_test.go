package godo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var vTestObj = &VPC{
	ID:          "880b7f98-f062-404d-b33c-458d545696f6",
	URN:         "do:vpc:880b7f98-f062-404d-b33c-458d545696f6",
	Name:        "my-new-vpc",
	Description: "vpc description",
	IPRange:     "10.122.0.0/20",
	RegionSlug:  "s2r7",
	CreatedAt:   time.Date(2019, 2, 4, 21, 48, 40, 995304079, time.UTC),
	Default:     false,
}

var vTestJSON = `
    {
      "id":"880b7f98-f062-404d-b33c-458d545696f6",
      "urn":"do:vpc:880b7f98-f062-404d-b33c-458d545696f6",
      "name":"my-new-vpc",
      "description":"vpc description",
      "ip_range":"10.122.0.0/20",
      "region":"s2r7",
      "created_at":"2019-02-04T21:48:40.995304079Z",
      "default":false
    }
`

func TestVPCs_Get(t *testing.T) {
	setup()
	defer teardown()

	svc := client.VPCs
	path := "/v2/vpcs"
	want := vTestObj
	id := "880b7f98-f062-404d-b33c-458d545696f6"
	jsonBlob := `
{
  "vpc":
` + vTestJSON + `
}
`

	mux.HandleFunc(path+"/"+id, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, jsonBlob)
	})

	got, _, err := svc.Get(ctx, id)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestVPCs_List(t *testing.T) {
	setup()
	defer teardown()

	svc := client.VPCs
	path := "/v2/vpcs"
	want := []*VPC{
		vTestObj,
	}
	links := &Links{
		Pages: &Pages{
			Last: "http://localhost/v2/vpcs?page=3&per_page=1",
			Next: "http://localhost/v2/vpcs?page=2&per_page=1",
		},
	}
	meta := &Meta{
		Total: 3,
	}
	jsonBlob := `
{
  "vpcs": [
` + vTestJSON + `
  ],
  "links": {
    "pages": {
      "last": "http://localhost/v2/vpcs?page=3&per_page=1",
      "next": "http://localhost/v2/vpcs?page=2&per_page=1"
    }
  },
  "meta": {"total": 3}
}
`
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, jsonBlob)
	})

	got, resp, err := svc.List(ctx, nil)
	require.NoError(t, err)
	assert.Equal(t, want, got)
	assert.Equal(t, resp.Links, links)
	assert.Equal(t, resp.Meta, meta)
}

func TestVPCs_Create(t *testing.T) {
	setup()
	defer teardown()

	svc := client.VPCs
	path := "/v2/vpcs"
	want := vTestObj
	req := &VPCCreateRequest{
		Name:       "my-new-vpc",
		RegionSlug: "s2r7",
	}
	jsonBlob := `
{
  "vpc":
` + vTestJSON + `
}
`

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		c := new(VPCCreateRequest)
		err := json.NewDecoder(r.Body).Decode(c)
		if err != nil {
			t.Fatal(err)
		}

		testMethod(t, r, http.MethodPost)
		require.Equal(t, c, req)
		fmt.Fprint(w, jsonBlob)
	})

	got, _, err := svc.Create(ctx, req)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestVPCs_Update(t *testing.T) {

	tests := []struct {
		desc                string
		id                  string
		req                 *VPCUpdateRequest
		mockResponse        string
		expectedRequestBody string
		expectedUpdatedVPC  *VPC
	}{
		{
			desc: "setting name and description without default argument",
			id:   "880b7f98-f062-404d-b33c-458d545696f6",
			req: &VPCUpdateRequest{
				Name:        "my-new-vpc",
				Description: "vpc description",
			},
			mockResponse: `
			{
			  "vpc":
			` + vTestJSON + `
			}
			`,
			expectedRequestBody: `{"name":"my-new-vpc","description":"vpc description"}`,
			expectedUpdatedVPC:  vTestObj,
		},

		{
			desc: "setting the default vpc option",
			id:   "880b7f98-f062-404d-b33c-458d545696f6",
			req: &VPCUpdateRequest{
				Name:        "my-new-vpc",
				Description: "vpc description",
				Default:     PtrTo(false),
			},
			mockResponse: `
			{
			  "vpc":
			` + vTestJSON + `
			}
			`,
			expectedRequestBody: `{"name":"my-new-vpc","description":"vpc description","default":false}`,
			expectedUpdatedVPC:  vTestObj,
		},

		{
			desc: "setting the default vpc option",
			id:   "880b7f98-f062-404d-b33c-458d545696f6",
			req: &VPCUpdateRequest{
				Name:        "my-new-vpc",
				Description: "vpc description",
				Default:     PtrTo(true),
			},
			mockResponse: `
			{
			  "vpc":
			` + vTestJSON + `
			}
			`,
			expectedRequestBody: `{"name":"my-new-vpc","description":"vpc description","default":true}`,
			expectedUpdatedVPC:  vTestObj,
		},
	}

	for _, tt := range tests {
		setup()

		mux.HandleFunc("/v2/vpcs/"+tt.id, func(w http.ResponseWriter, r *http.Request) {
			buf := new(bytes.Buffer)
			buf.ReadFrom(r.Body)
			require.Equal(t, tt.expectedRequestBody, strings.TrimSpace(buf.String()))

			v := new(VPCUpdateRequest)
			err := json.NewDecoder(buf).Decode(v)
			if err != nil {
				t.Fatal(err)
			}

			testMethod(t, r, http.MethodPut)
			fmt.Fprint(w, tt.mockResponse)
		})

		got, _, err := client.VPCs.Update(ctx, tt.id, tt.req)

		teardown()

		require.NoError(t, err)
		require.Equal(t, tt.expectedUpdatedVPC, got)
	}
}

func TestVPCs_Set(t *testing.T) {

	tests := []struct {
		desc                string
		id                  string
		updateFields        []VPCSetField
		mockResponse        string
		expectedRequestBody string
		expectedUpdatedVPC  *VPC
	}{
		{
			desc: "setting name and description",
			id:   "880b7f98-f062-404d-b33c-458d545696f6",
			updateFields: []VPCSetField{
				VPCSetName("my-new-vpc"),
				VPCSetDescription("vpc description"),
			},
			mockResponse: `
			{
			  "vpc":
			` + vTestJSON + `
			}
			`,
			expectedRequestBody: `{"description":"vpc description","name":"my-new-vpc"}`,
			expectedUpdatedVPC:  vTestObj,
		},

		{
			desc: "setting the default vpc option",
			id:   "880b7f98-f062-404d-b33c-458d545696f6",
			updateFields: []VPCSetField{
				VPCSetName("my-new-vpc"),
				VPCSetDescription("vpc description"),
				VPCSetDefault(),
			},
			mockResponse: `
			{
			  "vpc":
			` + vTestJSON + `
			}
			`,
			expectedRequestBody: `{"default":true,"description":"vpc description","name":"my-new-vpc"}`,
			expectedUpdatedVPC:  vTestObj,
		},
	}

	for _, tt := range tests {
		setup()

		mux.HandleFunc("/v2/vpcs/"+tt.id, func(w http.ResponseWriter, r *http.Request) {
			buf := new(bytes.Buffer)
			buf.ReadFrom(r.Body)
			require.Equal(t, tt.expectedRequestBody, strings.TrimSpace(buf.String()))

			v := new(VPCUpdateRequest)
			err := json.NewDecoder(buf).Decode(v)
			if err != nil {
				t.Fatal(err)
			}

			testMethod(t, r, http.MethodPatch)
			fmt.Fprint(w, tt.mockResponse)
		})

		got, _, err := client.VPCs.Set(ctx, tt.id, tt.updateFields...)

		teardown()

		require.NoError(t, err)
		require.Equal(t, tt.expectedUpdatedVPC, got)
	}
}

func TestVPCs_Delete(t *testing.T) {
	setup()
	defer teardown()

	svc := client.VPCs
	path := "/v2/vpcs"
	id := "880b7f98-f062-404d-b33c-458d545696f6"

	mux.HandleFunc(path+"/"+id, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
	})

	_, err := svc.Delete(ctx, id)
	require.NoError(t, err)
}

func TestVPCs_ListMembers(t *testing.T) {
	tests := []struct {
		desc          string
		expectedQuery string
		query         *VPCListMembersRequest
		resp          string
		want          []*VPCMember
	}{
		{
			desc:          "list all members",
			expectedQuery: "",
			query:         nil,
			resp: `{
		"members": [
			{
				"urn": "do:loadbalancer:fb294d78-d193-4cb2-8737-ea620993591b",
				"name": "nyc1-load-balancer-01",
				"created_at": "2020-03-16T19:30:48Z"
			},
			{
				"urn": "do:dbaas:13f7a2f6-43df-4c4a-8129-8733267ddeea",
				"name": "db-postgresql-nyc1-55986",
				"created_at": "2020-03-15T19:30:48Z"
			},
			{
				"urn": "do:kubernetes:da39d893-96e1-4e4d-971d-1fdda33a46b1",
				"name": "k8s-nyc1-1584127772221",
				"created_at": "2020-03-14T19:30:48Z"
			},
			{
				"urn": "do:droplet:86e29982-03a7-4946-8a07-a0114dff8754",
				"name": "ubuntu-s-1vcpu-1gb-nyc1-01",
				"created_at": "2020-03-13T19:30:48Z"
			}
		],
		"links": {
		},
		"meta": {
			"total": 4
		}
		}`,
			want: []*VPCMember{
				{
					URN:       "do:loadbalancer:fb294d78-d193-4cb2-8737-ea620993591b",
					Name:      "nyc1-load-balancer-01",
					CreatedAt: time.Date(2020, 3, 16, 19, 30, 48, 0, time.UTC),
				},
				{
					URN:       "do:dbaas:13f7a2f6-43df-4c4a-8129-8733267ddeea",
					Name:      "db-postgresql-nyc1-55986",
					CreatedAt: time.Date(2020, 3, 15, 19, 30, 48, 0, time.UTC),
				},
				{
					URN:       "do:kubernetes:da39d893-96e1-4e4d-971d-1fdda33a46b1",
					Name:      "k8s-nyc1-1584127772221",
					CreatedAt: time.Date(2020, 3, 14, 19, 30, 48, 0, time.UTC),
				},
				{
					URN:       "do:droplet:86e29982-03a7-4946-8a07-a0114dff8754",
					Name:      "ubuntu-s-1vcpu-1gb-nyc1-01",
					CreatedAt: time.Date(2020, 3, 13, 19, 30, 48, 0, time.UTC),
				},
			},
		},
		{
			desc:          "list droplet members",
			expectedQuery: "droplet",
			query:         &VPCListMembersRequest{ResourceType: "droplet"},
			resp: `{
		"members": [
			{
				"urn": "do:droplet:86e29982-03a7-4946-8a07-a0114dff8754",
				"name": "ubuntu-s-1vcpu-1gb-nyc1-01",
				"created_at": "2020-03-13T19:30:48Z"
			}
		],
		"links": {
		},
		"meta": {
			"total": 1
		}
		}`,
			want: []*VPCMember{
				{
					URN:       "do:droplet:86e29982-03a7-4946-8a07-a0114dff8754",
					Name:      "ubuntu-s-1vcpu-1gb-nyc1-01",
					CreatedAt: time.Date(2020, 3, 13, 19, 30, 48, 0, time.UTC),
				},
			},
		},
	}

	id := "880b7f98-f062-404d-b33c-458d545696f6"
	path := "/v2/vpcs/" + id + "/members"

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			setup()
			defer teardown()

			mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
				testMethod(t, r, http.MethodGet)
				require.Equal(t, tt.expectedQuery, r.URL.Query().Get("resource_type"))
				fmt.Fprint(w, tt.resp)
			})

			got, _, err := client.VPCs.ListMembers(ctx, id, tt.query, nil)

			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}
