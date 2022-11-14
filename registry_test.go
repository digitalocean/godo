package godo

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testRegistry          = "test-registry"
	testRegion            = "r1"
	testRepository        = "test/repository"
	testEncodedRepository = "test%2Frepository"
	testTag               = "test-tag"
	testDigest            = "sha256:e692418e4cbaf90ca69d05a66403747baa33ee08806650b51fab815ad7fc331f"
	testCompressedSize    = 2789669
	testSize              = 5843968
	testGCBlobsDeleted    = 42
	testGCFreedBytes      = 666
	testGCStatus          = "requested"
	testGCUUID            = "mew-mew-id"
	testGCType            = GCTypeUnreferencedBlobsOnly
)

var (
	testTime              = time.Date(2020, 4, 1, 0, 0, 0, 0, time.UTC)
	testTimeString        = testTime.Format(time.RFC3339)
	testGarbageCollection = &GarbageCollection{
		UUID:         testGCUUID,
		RegistryName: testRegistry,
		Status:       testGCStatus,
		CreatedAt:    testTime,
		UpdatedAt:    testTime,
		BlobsDeleted: testGCBlobsDeleted,
		FreedBytes:   testGCFreedBytes,
		Type:         testGCType,
	}
)

func TestRegistry_Create(t *testing.T) {
	setup()
	defer teardown()

	want := &Registry{
		Name:                       testRegistry,
		StorageUsageBytes:          0,
		StorageUsageBytesUpdatedAt: testTime,
		CreatedAt:                  testTime,
		Region:                     testRegion,
	}

	createRequest := &RegistryCreateRequest{
		Name:                 want.Name,
		SubscriptionTierSlug: "basic",
		Region:               testRegion,
	}

	createResponseJSON := `
{
	"registry": {
		"name": "` + testRegistry + `",
		"storage_usage_bytes": 0,
        "storage_usage_bytes_updated_at": "` + testTimeString + `",
        "created_at": "` + testTimeString + `",
		"region": "` + testRegion + `"
	},
    "subscription": {
      "tier": {
        "name": "Basic",
        "slug": "basic",
        "included_repositories": 5,
        "included_storage_bytes": 5368709120,
        "allow_storage_overage": true,
        "included_bandwidth_bytes": 5368709120,
        "monthly_price_in_cents": 500
      },
      "created_at": "` + testTimeString + `",
      "updated_at": "` + testTimeString + `"
    }
}`

	mux.HandleFunc("/v2/registry", func(w http.ResponseWriter, r *http.Request) {
		v := new(RegistryCreateRequest)
		err := json.NewDecoder(r.Body).Decode(v)
		if err != nil {
			t.Fatal(err)
		}

		testMethod(t, r, http.MethodPost)
		require.Equal(t, v, createRequest)
		fmt.Fprint(w, createResponseJSON)
	})

	got, _, err := client.Registry.Create(ctx, createRequest)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestRegistry_Get(t *testing.T) {
	setup()
	defer teardown()

	want := &Registry{
		Name:                       testRegistry,
		StorageUsageBytes:          0,
		StorageUsageBytesUpdatedAt: testTime,
		CreatedAt:                  testTime,
		Region:                     testRegion,
	}

	getResponseJSON := `
{
	"registry": {
		"name": "` + testRegistry + `",
		"storage_usage_bytes": 0,
        "storage_usage_bytes_updated_at": "` + testTimeString + `",
        "created_at": "` + testTimeString + `",
		"region": "` + testRegion + `"
	}
}`

	mux.HandleFunc("/v2/registry", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, getResponseJSON)
	})
	got, _, err := client.Registry.Get(ctx)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestRegistry_Delete(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/registry", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
	})

	_, err := client.Registry.Delete(ctx)
	require.NoError(t, err)
}

func TestRegistry_DockerCredentials(t *testing.T) {
	returnedConfig := "this could be a docker config"
	tests := []struct {
		name                  string
		params                *RegistryDockerCredentialsRequest
		expectedReadWrite     string
		expectedExpirySeconds string
	}{
		{
			name:              "read-only (default)",
			params:            &RegistryDockerCredentialsRequest{},
			expectedReadWrite: "false",
		},
		{
			name:              "read/write",
			params:            &RegistryDockerCredentialsRequest{ReadWrite: true},
			expectedReadWrite: "true",
		},
		{
			name:                  "read-only + custom expiry",
			params:                &RegistryDockerCredentialsRequest{ExpirySeconds: PtrTo(60 * 60)},
			expectedReadWrite:     "false",
			expectedExpirySeconds: "3600",
		},
		{
			name:                  "read/write + custom expiry",
			params:                &RegistryDockerCredentialsRequest{ReadWrite: true, ExpirySeconds: PtrTo(60 * 60)},
			expectedReadWrite:     "true",
			expectedExpirySeconds: "3600",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			setup()
			defer teardown()

			mux.HandleFunc("/v2/registry/docker-credentials", func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, test.expectedReadWrite, r.URL.Query().Get("read_write"))
				require.Equal(t, test.expectedExpirySeconds, r.URL.Query().Get("expiry_seconds"))
				testMethod(t, r, http.MethodGet)
				fmt.Fprint(w, returnedConfig)
			})

			got, _, err := client.Registry.DockerCredentials(ctx, test.params)
			require.NoError(t, err)
			require.Equal(t, []byte(returnedConfig), got.DockerConfigJSON)
		})
	}
}

func TestRepository_List(t *testing.T) {
	setup()
	defer teardown()

	wantRepositories := []*Repository{
		{
			RegistryName: testRegistry,
			Name:         testRepository,
			TagCount:     1,
			LatestTag: &RepositoryTag{
				RegistryName:        testRegistry,
				Repository:          testRepository,
				Tag:                 testTag,
				ManifestDigest:      testDigest,
				CompressedSizeBytes: testCompressedSize,
				SizeBytes:           testSize,
				UpdatedAt:           testTime,
			},
		},
	}
	getResponseJSON := `{
	"repositories": [
		{
			"registry_name": "` + testRegistry + `",
			"name": "` + testRepository + `",
			"tag_count": 1,
			"latest_tag": {
				"registry_name": "` + testRegistry + `",
				"repository": "` + testRepository + `",
				"tag": "` + testTag + `",
				"manifest_digest": "` + testDigest + `",
				"compressed_size_bytes": ` + fmt.Sprintf("%d", testCompressedSize) + `,
				"size_bytes": ` + fmt.Sprintf("%d", testSize) + `,
				"updated_at": "` + testTimeString + `"
			}
		}
	],
	"links": {
	    "pages": {
			"next": "https://api.digitalocean.com/v2/registry/` + testRegistry + `/repositories?page=2",
			"last": "https://api.digitalocean.com/v2/registry/` + testRegistry + `/repositories?page=2"
		}
	},
	"meta": {
	    "total": 2
	}
}`

	mux.HandleFunc(fmt.Sprintf("/v2/registry/%s/repositories", testRegistry), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		testFormValues(t, r, map[string]string{"page": "1", "per_page": "1"})
		fmt.Fprint(w, getResponseJSON)
	})
	got, response, err := client.Registry.ListRepositories(ctx, testRegistry, &ListOptions{Page: 1, PerPage: 1})
	require.NoError(t, err)
	require.Equal(t, wantRepositories, got)

	gotRespLinks := response.Links
	wantRespLinks := &Links{
		Pages: &Pages{
			Next: fmt.Sprintf("https://api.digitalocean.com/v2/registry/%s/repositories?page=2", testRegistry),
			Last: fmt.Sprintf("https://api.digitalocean.com/v2/registry/%s/repositories?page=2", testRegistry),
		},
	}
	assert.Equal(t, wantRespLinks, gotRespLinks)

	gotRespMeta := response.Meta
	wantRespMeta := &Meta{
		Total: 2,
	}
	assert.Equal(t, wantRespMeta, gotRespMeta)
}

func TestRepository_ListV2(t *testing.T) {
	setup()
	defer teardown()

	wantRepositories := []*RepositoryV2{
		{
			RegistryName:  testRegistry,
			Name:          testRepository,
			TagCount:      2,
			ManifestCount: 1,
			LatestManifest: &RepositoryManifest{
				Digest:              "sha256:abc",
				RegistryName:        testRegistry,
				Repository:          testRepository,
				CompressedSizeBytes: testCompressedSize,
				SizeBytes:           testSize,
				UpdatedAt:           testTime,
				Tags:                []string{"v1", "v2"},
				Blobs: []*Blob{
					{
						Digest:              "sha256:blob1",
						CompressedSizeBytes: 100,
					},
					{
						Digest:              "sha256:blob2",
						CompressedSizeBytes: 200,
					},
				},
			},
		},
	}
	baseLinkPage := fmt.Sprintf("https://api.digitalocean.com/v2/registry/%s/repositoriesV2", testRegistry)
	getResponseJSON := `{
	"repositories": [
		{
			"registry_name": "` + testRegistry + `",
			"name": "` + testRepository + `",
			"tag_count": 2,
			"manifest_count": 1,
			"latest_manifest": {
				"digest": "sha256:abc",
				"registry_name": "` + testRegistry + `",
				"repository": "` + testRepository + `",
				"compressed_size_bytes": ` + fmt.Sprintf("%d", testCompressedSize) + `,
				"size_bytes": ` + fmt.Sprintf("%d", testSize) + `,
				"updated_at": "` + testTimeString + `",
				"tags": [
					"v1",
					"v2"
				],
				"blobs": [
					{
						"digest": "sha256:blob1",
						"compressed_size_bytes": 100
					},
					{
						"digest": "sha256:blob2",
						"compressed_size_bytes": 200
					}
				]
			}
		}
	],
	"links": {
	    "pages": {
			"first":    "` + baseLinkPage + `?page=1&page_size=1",
			"prev":     "` + baseLinkPage + `?page=2&page_size=1&page_token=aaa",
			"next":     "` + baseLinkPage + `?page=4&page_size=1&page_token=ccc",
			"last":     "` + baseLinkPage + `?page=5&page_size=1"
		}
	},
	"meta": {
	    "total": 5
	}
}`

	mux.HandleFunc(fmt.Sprintf("/v2/registry/%s/repositoriesV2", testRegistry), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		testFormValues(t, r, map[string]string{"page": "3", "per_page": "1", "page_token": "bbb"})
		fmt.Fprint(w, getResponseJSON)
	})
	got, response, err := client.Registry.ListRepositoriesV2(ctx, testRegistry, &TokenListOptions{Page: 3, PerPage: 1, Token: "bbb"})
	require.NoError(t, err)
	require.Equal(t, wantRepositories, got)

	gotRespLinks := response.Links
	wantRespLinks := &Links{
		Pages: &Pages{
			First: fmt.Sprintf("https://api.digitalocean.com/v2/registry/%s/repositoriesV2?page=1&page_size=1", testRegistry),
			Prev:  fmt.Sprintf("https://api.digitalocean.com/v2/registry/%s/repositoriesV2?page=2&page_size=1&page_token=aaa", testRegistry),
			Next:  fmt.Sprintf("https://api.digitalocean.com/v2/registry/%s/repositoriesV2?page=4&page_size=1&page_token=ccc", testRegistry),
			Last:  fmt.Sprintf("https://api.digitalocean.com/v2/registry/%s/repositoriesV2?page=5&page_size=1", testRegistry),
		},
	}
	assert.Equal(t, wantRespLinks, gotRespLinks)

	gotRespMeta := response.Meta
	wantRespMeta := &Meta{
		Total: 5,
	}
	assert.Equal(t, wantRespMeta, gotRespMeta)
}

func TestRepository_ListTags(t *testing.T) {
	setup()
	defer teardown()

	wantTags := []*RepositoryTag{
		{
			RegistryName:        testRegistry,
			Repository:          testRepository,
			Tag:                 testTag,
			ManifestDigest:      testDigest,
			CompressedSizeBytes: testCompressedSize,
			SizeBytes:           testSize,
			UpdatedAt:           testTime,
		},
	}
	getResponseJSON := `{
	"tags": [
		{
			"registry_name": "` + testRegistry + `",
			"repository": "` + testRepository + `",
			"tag": "` + testTag + `",
			"manifest_digest": "` + testDigest + `",
			"compressed_size_bytes": ` + fmt.Sprintf("%d", testCompressedSize) + `,
			"size_bytes": ` + fmt.Sprintf("%d", testSize) + `,
			"updated_at": "` + testTimeString + `"
		}
	],
	"links": {
	    "pages": {
			"next": "https://api.digitalocean.com/v2/registry/` + testRegistry + `/repositories/` + testEncodedRepository + `/tags?page=2",
			"last": "https://api.digitalocean.com/v2/registry/` + testRegistry + `/repositories/` + testEncodedRepository + `/tags?page=2"
		}
	},
	"meta": {
	    "total": 2
	}
}`

	mux.HandleFunc(fmt.Sprintf("/v2/registry/%s/repositories/%s/tags", testRegistry, testRepository), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		testFormValues(t, r, map[string]string{"page": "1", "per_page": "1"})
		fmt.Fprint(w, getResponseJSON)
	})
	got, response, err := client.Registry.ListRepositoryTags(ctx, testRegistry, testRepository, &ListOptions{Page: 1, PerPage: 1})
	require.NoError(t, err)
	require.Equal(t, wantTags, got)

	gotRespLinks := response.Links
	wantRespLinks := &Links{
		Pages: &Pages{
			Next: fmt.Sprintf("https://api.digitalocean.com/v2/registry/%s/repositories/%s/tags?page=2", testRegistry, testEncodedRepository),
			Last: fmt.Sprintf("https://api.digitalocean.com/v2/registry/%s/repositories/%s/tags?page=2", testRegistry, testEncodedRepository),
		},
	}
	assert.Equal(t, wantRespLinks, gotRespLinks)

	gotRespMeta := response.Meta
	wantRespMeta := &Meta{
		Total: 2,
	}
	assert.Equal(t, wantRespMeta, gotRespMeta)
}

func TestRegistry_DeleteTag(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc(fmt.Sprintf("/v2/registry/%s/repositories/%s/tags/%s", testRegistry, testRepository, testTag), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
	})

	_, err := client.Registry.DeleteTag(ctx, testRegistry, testRepository, testTag)
	require.NoError(t, err)
}

func TestRegistry_ListManifests(t *testing.T) {
	setup()
	defer teardown()

	wantTags := []*RepositoryManifest{
		{
			RegistryName:        testRegistry,
			Repository:          testRepository,
			Digest:              testDigest,
			CompressedSizeBytes: testCompressedSize,
			SizeBytes:           testSize,
			UpdatedAt:           testTime,
			Tags:                []string{"latest", "v1", "v2"},
			Blobs: []*Blob{
				{
					Digest:              "sha256:blob1",
					CompressedSizeBytes: 998,
				},
				{
					Digest:              "sha256:blob2",
					CompressedSizeBytes: 1,
				},
			},
		},
	}
	getResponseJSON := `{
	"manifests": [
		{
			"registry_name": "` + testRegistry + `",
			"repository": "` + testRepository + `",
			"digest": "` + testDigest + `",
			"compressed_size_bytes": ` + fmt.Sprintf("%d", testCompressedSize) + `,
			"size_bytes": ` + fmt.Sprintf("%d", testSize) + `,
			"updated_at": "` + testTimeString + `",
			"tags": [ "latest", "v1", "v2" ],
			"blobs": [
				{
					"digest": "sha256:blob1",
					"compressed_size_bytes": 998
				},
				{

					"digest": "sha256:blob2",
					"compressed_size_bytes": 1
				}
			]
		}
	],
	"links": {
	    "pages": {
			"first": "https://api.digitalocean.com/v2/registry/` + testRegistry + `/repositories/` + testEncodedRepository + `/digests?page=1&page_size=1",
			"prev": "https://api.digitalocean.com/v2/registry/` + testRegistry + `/repositories/` + testEncodedRepository + `/digests?page=2&page_size=1",
			"next": "https://api.digitalocean.com/v2/registry/` + testRegistry + `/repositories/` + testEncodedRepository + `/digests?page=4&page_size=1",
			"last": "https://api.digitalocean.com/v2/registry/` + testRegistry + `/repositories/` + testEncodedRepository + `/digests?page=5&page_size=1"
		}
	},
	"meta": {
	    "total": 5
	}
}`

	mux.HandleFunc(fmt.Sprintf("/v2/registry/%s/repositories/%s/digests", testRegistry, testRepository), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		testFormValues(t, r, map[string]string{"page": "3", "per_page": "1"})
		fmt.Fprint(w, getResponseJSON)
	})
	got, response, err := client.Registry.ListRepositoryManifests(ctx, testRegistry, testRepository, &ListOptions{Page: 3, PerPage: 1})
	require.NoError(t, err)
	require.Equal(t, wantTags, got)

	gotRespLinks := response.Links
	wantRespLinks := &Links{
		Pages: &Pages{
			First: fmt.Sprintf("https://api.digitalocean.com/v2/registry/%s/repositories/%s/digests?page=1&page_size=1", testRegistry, testEncodedRepository),
			Prev:  fmt.Sprintf("https://api.digitalocean.com/v2/registry/%s/repositories/%s/digests?page=2&page_size=1", testRegistry, testEncodedRepository),
			Next:  fmt.Sprintf("https://api.digitalocean.com/v2/registry/%s/repositories/%s/digests?page=4&page_size=1", testRegistry, testEncodedRepository),
			Last:  fmt.Sprintf("https://api.digitalocean.com/v2/registry/%s/repositories/%s/digests?page=5&page_size=1", testRegistry, testEncodedRepository),
		},
	}
	assert.Equal(t, wantRespLinks, gotRespLinks)

	gotRespMeta := response.Meta
	wantRespMeta := &Meta{
		Total: 5,
	}
	assert.Equal(t, wantRespMeta, gotRespMeta)
}

func TestRegistry_DeleteManifest(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc(fmt.Sprintf("/v2/registry/%s/repositories/%s/digests/%s", testRegistry, testRepository, testDigest), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
	})

	_, err := client.Registry.DeleteManifest(ctx, testRegistry, testRepository, testDigest)
	require.NoError(t, err)
}

func reifyTemplateStr(t *testing.T, tmplStr string, v interface{}) string {
	tmpl, err := template.New("meow").Parse(tmplStr)
	require.NoError(t, err)

	s := &strings.Builder{}
	err = tmpl.Execute(s, v)
	require.NoError(t, err)

	return s.String()
}

func TestGarbageCollection_Start(t *testing.T) {
	setup()
	defer teardown()

	want := testGarbageCollection
	requestResponseJSONTmpl := `
{
  "garbage_collection": {
    "uuid": "{{.UUID}}",
    "registry_name": "{{.RegistryName}}",
    "status": "{{.Status}}",
    "type": "{{.Type}}",
    "created_at": "{{.CreatedAt.Format "2006-01-02T15:04:05Z07:00"}}",
    "updated_at": "{{.UpdatedAt.Format "2006-01-02T15:04:05Z07:00"}}",
    "blobs_deleted": {{.BlobsDeleted}},
    "freed_bytes": {{.FreedBytes}}
  }
}`
	requestResponseJSON := reifyTemplateStr(t, requestResponseJSONTmpl, want)

	createRequest := &StartGarbageCollectionRequest{
		Type: GCTypeUnreferencedBlobsOnly,
	}
	mux.HandleFunc("/v2/registry/"+testRegistry+"/garbage-collection",
		func(w http.ResponseWriter, r *http.Request) {
			v := new(StartGarbageCollectionRequest)
			err := json.NewDecoder(r.Body).Decode(v)
			if err != nil {
				t.Fatal(err)
			}

			testMethod(t, r, http.MethodPost)
			require.Equal(t, v, createRequest)
			fmt.Fprint(w, requestResponseJSON)
		})

	got, _, err := client.Registry.StartGarbageCollection(ctx, testRegistry)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestGarbageCollection_Get(t *testing.T) {
	setup()
	defer teardown()

	want := testGarbageCollection
	requestResponseJSONTmpl := `
{
  "garbage_collection": {
    "uuid": "{{.UUID}}",
    "registry_name": "{{.RegistryName}}",
    "status": "{{.Status}}",
    "type": "{{.Type}}",
    "created_at": "{{.CreatedAt.Format "2006-01-02T15:04:05Z07:00"}}",
    "updated_at": "{{.UpdatedAt.Format "2006-01-02T15:04:05Z07:00"}}",
    "blobs_deleted": {{.BlobsDeleted}},
    "freed_bytes": {{.FreedBytes}}
  }
}`
	requestResponseJSON := reifyTemplateStr(t, requestResponseJSONTmpl, want)

	mux.HandleFunc("/v2/registry/"+testRegistry+"/garbage-collection",
		func(w http.ResponseWriter, r *http.Request) {
			testMethod(t, r, http.MethodGet)
			fmt.Fprint(w, requestResponseJSON)
		})

	got, _, err := client.Registry.GetGarbageCollection(ctx, testRegistry)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestGarbageCollection_List(t *testing.T) {
	setup()
	defer teardown()

	want := []*GarbageCollection{testGarbageCollection}
	requestResponseJSONTmpl := `
{
  "garbage_collections": [
    {
      "uuid": "{{.UUID}}",
      "registry_name": "{{.RegistryName}}",
      "status": "{{.Status}}",
      "type": "{{.Type}}",
      "created_at": "{{.CreatedAt.Format "2006-01-02T15:04:05Z07:00"}}",
      "updated_at": "{{.UpdatedAt.Format "2006-01-02T15:04:05Z07:00"}}",
      "blobs_deleted": {{.BlobsDeleted}},
      "freed_bytes": {{.FreedBytes}}
    }
  ],
	"links": {
	    "pages": {
			"next": "https://api.digitalocean.com/v2/registry/` + testRegistry + `/garbage-collections?page=2",
			"last": "https://api.digitalocean.com/v2/registry/` + testRegistry + `/garbage-collections?page=2"
		}
	},
	"meta": {
	    "total": 2
	}
}`
	requestResponseJSON := reifyTemplateStr(t, requestResponseJSONTmpl, testGarbageCollection)

	mux.HandleFunc("/v2/registry/"+testRegistry+"/garbage-collections",
		func(w http.ResponseWriter, r *http.Request) {
			testMethod(t, r, http.MethodGet)
			testFormValues(t, r, map[string]string{"page": "1", "per_page": "1"})
			fmt.Fprint(w, requestResponseJSON)
		})

	got, resp, err := client.Registry.ListGarbageCollections(ctx, testRegistry, &ListOptions{Page: 1, PerPage: 1})
	require.NoError(t, err)
	require.Equal(t, want, got)

	gotRespLinks := resp.Links
	wantRespLinks := &Links{
		Pages: &Pages{
			Next: fmt.Sprintf("https://api.digitalocean.com/v2/registry/%s/garbage-collections?page=2", testRegistry),
			Last: fmt.Sprintf("https://api.digitalocean.com/v2/registry/%s/garbage-collections?page=2", testRegistry),
		},
	}
	assert.Equal(t, wantRespLinks, gotRespLinks)

	gotRespMeta := resp.Meta
	wantRespMeta := &Meta{
		Total: 2,
	}
	assert.Equal(t, wantRespMeta, gotRespMeta)
}

func TestGarbageCollection_Update(t *testing.T) {
	setup()
	defer teardown()

	updateRequest := &UpdateGarbageCollectionRequest{
		Cancel: true,
	}

	want := testGarbageCollection
	requestResponseJSONTmpl := `
{
  "garbage_collection": {
    "uuid": "{{.UUID}}",
    "registry_name": "{{.RegistryName}}",
    "status": "{{.Status}}",
    "type": "{{.Type}}",
    "created_at": "{{.CreatedAt.Format "2006-01-02T15:04:05Z07:00"}}",
    "updated_at": "{{.UpdatedAt.Format "2006-01-02T15:04:05Z07:00"}}",
    "blobs_deleted": {{.BlobsDeleted}},
    "freed_bytes": {{.FreedBytes}}
  }
}`
	requestResponseJSON := reifyTemplateStr(t, requestResponseJSONTmpl, want)

	mux.HandleFunc("/v2/registry/"+testRegistry+"/garbage-collection/"+testGCUUID,
		func(w http.ResponseWriter, r *http.Request) {
			v := new(UpdateGarbageCollectionRequest)
			err := json.NewDecoder(r.Body).Decode(v)
			if err != nil {
				t.Fatal(err)
			}

			testMethod(t, r, http.MethodPut)
			require.Equal(t, v, updateRequest)
			fmt.Fprint(w, requestResponseJSON)
		})

	got, _, err := client.Registry.UpdateGarbageCollection(ctx, testRegistry, testGCUUID, updateRequest)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestRegistry_GetOptions(t *testing.T) {
	responseJSON := `
{
  "options": {
	"available_regions": [
		"r1",
		"r2"
	],
    "subscription_tiers": [
      {
        "name": "Starter",
        "slug": "starter",
        "included_repositories": 1,
        "included_storage_bytes": 524288000,
        "allow_storage_overage": false,
        "included_bandwidth_bytes": 524288000,
        "monthly_price_in_cents": 0,
        "eligible": false,
        "eligibility_reasons": [
          "OverStorageLimit",
          "OverRepositoryLimit"
        ]
      },
      {
        "name": "Basic",
        "slug": "basic",
        "included_repositories": 5,
        "included_storage_bytes": 5368709120,
        "allow_storage_overage": true,
        "included_bandwidth_bytes": 5368709120,
        "monthly_price_in_cents": 500,
        "eligible": false,
        "eligibility_reasons": [
          "OverRepositoryLimit"
        ]
      },
      {
        "name": "Professional",
        "slug": "professional",
        "included_repositories": 0,
        "included_storage_bytes": 107374182400,
        "allow_storage_overage": true,
        "included_bandwidth_bytes": 107374182400,
        "monthly_price_in_cents": 2000,
        "eligible": true
      }
    ]
  }
}`
	want := &RegistryOptions{
		AvailableRegions: []string{
			"r1",
			"r2",
		},
		SubscriptionTiers: []*RegistrySubscriptionTier{
			{
				Name:                   "Starter",
				Slug:                   "starter",
				IncludedRepositories:   1,
				IncludedStorageBytes:   524288000,
				AllowStorageOverage:    false,
				IncludedBandwidthBytes: 524288000,
				MonthlyPriceInCents:    0,
				Eligible:               false,
				EligibilityReasons: []string{
					"OverStorageLimit",
					"OverRepositoryLimit",
				},
			},
			{
				Name:                   "Basic",
				Slug:                   "basic",
				IncludedRepositories:   5,
				IncludedStorageBytes:   5368709120,
				AllowStorageOverage:    true,
				IncludedBandwidthBytes: 5368709120,
				MonthlyPriceInCents:    500,
				Eligible:               false,
				EligibilityReasons: []string{
					"OverRepositoryLimit",
				},
			},
			{
				Name:                   "Professional",
				Slug:                   "professional",
				IncludedRepositories:   0,
				IncludedStorageBytes:   107374182400,
				AllowStorageOverage:    true,
				IncludedBandwidthBytes: 107374182400,
				MonthlyPriceInCents:    2000,
				Eligible:               true,
			},
		},
	}

	setup()
	defer teardown()

	mux.HandleFunc("/v2/registry/options", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, responseJSON)
	})

	got, _, err := client.Registry.GetOptions(ctx)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestRegistry_GetSubscription(t *testing.T) {
	setup()
	defer teardown()

	want := &RegistrySubscription{
		Tier: &RegistrySubscriptionTier{
			Name:                   "Basic",
			Slug:                   "basic",
			IncludedRepositories:   5,
			IncludedStorageBytes:   5368709120,
			AllowStorageOverage:    true,
			IncludedBandwidthBytes: 5368709120,
			MonthlyPriceInCents:    500,
		},
		CreatedAt: testTime,
		UpdatedAt: testTime,
	}

	getResponseJSON := `
{
  "subscription": {
    "tier": {
      "name": "Basic",
      "slug": "basic",
      "included_repositories": 5,
      "included_storage_bytes": 5368709120,
      "allow_storage_overage": true,
      "included_bandwidth_bytes": 5368709120,
      "monthly_price_in_cents": 500
    },
    "created_at": "` + testTimeString + `",
    "updated_at": "` + testTimeString + `"
  }
}
`

	mux.HandleFunc("/v2/registry/subscription", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, getResponseJSON)
	})
	got, _, err := client.Registry.GetSubscription(ctx)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestRegistry_UpdateSubscription(t *testing.T) {
	setup()
	defer teardown()

	updateRequest := &RegistrySubscriptionUpdateRequest{
		TierSlug: "professional",
	}

	want := &RegistrySubscription{
		Tier: &RegistrySubscriptionTier{
			Name:                   "Professional",
			Slug:                   "professional",
			IncludedRepositories:   0,
			IncludedStorageBytes:   107374182400,
			AllowStorageOverage:    true,
			IncludedBandwidthBytes: 107374182400,
			MonthlyPriceInCents:    2000,
			Eligible:               true,
		},
		CreatedAt: testTime,
		UpdatedAt: testTime,
	}

	updateResponseJSON := `{
  "subscription": {
    "tier": {
        "name": "Professional",
        "slug": "professional",
        "included_repositories": 0,
        "included_storage_bytes": 107374182400,
        "allow_storage_overage": true,
        "included_bandwidth_bytes": 107374182400,
        "monthly_price_in_cents": 2000,
        "eligible": true
      },
    "created_at": "` + testTimeString + `",
    "updated_at": "` + testTimeString + `"
  }
}`

	mux.HandleFunc("/v2/registry/subscription",
		func(w http.ResponseWriter, r *http.Request) {
			v := new(RegistrySubscriptionUpdateRequest)
			err := json.NewDecoder(r.Body).Decode(v)
			if err != nil {
				t.Fatal(err)
			}

			testMethod(t, r, http.MethodPost)
			require.Equal(t, v, updateRequest)
			fmt.Fprint(w, updateResponseJSON)
		})

	got, _, err := client.Registry.UpdateSubscription(ctx, updateRequest)
	require.NoError(t, err)
	require.Equal(t, want, got)
}
