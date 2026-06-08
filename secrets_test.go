package godo

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (fn roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

const (
	secretName   = "my-secret"
	secretRegion = "nyc3"
)

func TestSecretsList(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/security/secrets", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "1", r.URL.Query().Get("page"))
		assert.Equal(t, "20", r.URL.Query().Get("per_page"))
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{
			"secrets": [{
				"secret": "`+secretName+`",
				"region": "`+secretRegion+`",
				"version": 1,
				"created_at": "2026-06-08T12:00:00Z",
				"updated_at": "2026-06-08T12:00:00Z"
			}],
			"meta": {"page": 1, "pages": 1, "total": 1},
			"unavailable_regions": ["sfo3"]
		}`)
	})

	secrets, resp, err := client.Secrets.List(context.Background(), &ListOptions{Page: 1, PerPage: 20})
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, secrets, 1)
	assert.Equal(t, secretName, secrets[0].Name)
	assert.Equal(t, secretRegion, secrets[0].Region)
	assert.Equal(t, 1, secrets[0].Version)
}

func TestSecretsGet(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/security/secrets/"+secretName, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, secretRegion, r.URL.Query().Get("region"))
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{
			"secret": "`+secretName+`",
			"version": 1,
			"values": {"key": "val"},
			"created_at": "2026-06-08T12:00:00Z",
			"updated_at": "2026-06-08T12:00:00Z"
		}`)
	})

	secret, resp, err := client.Secrets.Get(context.Background(), secretName, secretRegion)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, secretName, secret.Name)
	assert.Equal(t, 1, secret.Version)
	assert.Equal(t, "val", secret.Values["key"])
}

func TestSecretsListVersions(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/security/secrets/"+secretName+"/versions", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, secretRegion, r.URL.Query().Get("region"))
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{
			"versions": [{
				"version": 1,
				"created_at": "2026-06-08T11:00:00Z",
				"updated_at": "2026-06-08T11:00:00Z"
			}]
		}`)
	})

	versions, resp, err := client.Secrets.ListVersions(context.Background(), secretName, secretRegion)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, versions, 1)
	assert.Equal(t, 1, versions[0].Version)
}

func TestSecretsCreate(t *testing.T) {
	setup()
	defer teardown()

	createRequest := &SecretCreateRequest{
		Name:   secretName,
		Region: secretRegion,
		Values: map[string]string{"key": "val"},
	}

	mux.HandleFunc("/v2/security/secrets", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		var body SecretCreateRequest
		assert.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		assert.Equal(t, createRequest.Name, body.Name)
		assert.Equal(t, createRequest.Region, body.Region)
		assert.Equal(t, createRequest.Values, body.Values)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"name": "`+secretName+`", "region": "`+secretRegion+`", "version": 1}`)
	})

	result, resp, err := client.Secrets.Create(context.Background(), createRequest)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, secretName, result.Name)
	assert.Equal(t, secretRegion, result.Region)
	assert.Equal(t, 1, result.Version)
}

func TestSecretsCreateNoContentWithBody(t *testing.T) {
	setup()
	defer teardown()

	client.HTTPClient = &http.Client{
		Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			assert.Equal(t, http.MethodPost, req.Method)
			return &http.Response{
				StatusCode: http.StatusNoContent,
				Body:       io.NopCloser(strings.NewReader(`{"name": "` + secretName + `", "region": "` + secretRegion + `", "version": 1}`)),
				Header:     make(http.Header),
			}, nil
		}),
	}

	createRequest := &SecretCreateRequest{
		Name:   secretName,
		Region: secretRegion,
		Values: map[string]string{"key": "val"},
	}

	result, resp, err := client.Secrets.Create(context.Background(), createRequest)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	assert.Equal(t, secretName, result.Name)
	assert.Equal(t, secretRegion, result.Region)
	assert.Equal(t, 1, result.Version)
}

func TestSecretsUpdate(t *testing.T) {
	setup()
	defer teardown()

	updateRequest := &SecretUpdateRequest{
		Region:  secretRegion,
		Version: 1,
		Values:  map[string]string{"key": "new-val"},
	}

	mux.HandleFunc("/v2/security/secrets/"+secretName, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		var body SecretUpdateRequest
		assert.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		assert.Equal(t, updateRequest.Region, body.Region)
		assert.Equal(t, updateRequest.Version, body.Version)
		assert.Equal(t, updateRequest.Values, body.Values)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"name": "`+secretName+`", "region": "`+secretRegion+`", "version": 2}`)
	})

	result, resp, err := client.Secrets.Update(context.Background(), secretName, updateRequest)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, secretName, result.Name)
	assert.Equal(t, secretRegion, result.Region)
	assert.Equal(t, 2, result.Version)
}

func TestSecretsDelete(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/security/secrets/"+secretName, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Equal(t, secretRegion, r.URL.Query().Get("region"))
		w.WriteHeader(http.StatusNoContent)
	})

	resp, err := client.Secrets.Delete(context.Background(), secretName, secretRegion)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestSecretsRestore(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/security/secrets/"+secretName+"/restore", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, secretRegion, r.URL.Query().Get("region"))
		w.WriteHeader(http.StatusNoContent)
	})

	resp, err := client.Secrets.Restore(context.Background(), secretName, secretRegion)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}
