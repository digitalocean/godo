package godo

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFunctions_ListNamespaces(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/functions/namespaces", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"namespaces": [
				{
					"api_host": "https://faas.do.com",
					"namespace": "123-abc",
					"created_at": "2022-06-16T12:09:13Z",
					"updated_at": "2022-06-16T12:09:13Z",
					"label": "my-namespace-1",
					"region": "nyc1",
					"uuid": "",
					"key": ""
				},
				{
					"api_host": "https://faas.do.com",
					"namespace": "456-abc",
					"created_at": "2022-11-02T18:33:44Z",
					"updated_at": "2022-11-02T18:33:44Z",
					"label": "my-namespace-2",
					"region": "nyc3",
					"uuid": "",
					"key": ""
				}
			]
		}`)
	})

	namespaces, _, err := client.Functions.ListNamespaces(ctx)
	require.NoError(t, err)

	expectedNamespaces := []FunctionsNamespace{
		{
			ApiHost:   "https://faas.do.com",
			Namespace: "123-abc",
			CreatedAt: time.Date(2022, 6, 16, 12, 9, 13, 0, time.UTC),
			UpdatedAt: time.Date(2022, 6, 16, 12, 9, 13, 0, time.UTC),
			Label:     "my-namespace-1",
			Region:    "nyc1",
			UUID:      "",
			Key:       "",
		},
		{
			ApiHost:   "https://faas.do.com",
			Namespace: "456-abc",
			CreatedAt: time.Date(2022, 11, 2, 18, 33, 44, 0, time.UTC),
			UpdatedAt: time.Date(2022, 11, 2, 18, 33, 44, 0, time.UTC),
			Label:     "my-namespace-2",
			Region:    "nyc3",
			UUID:      "",
			Key:       "",
		},
	}
	assert.Equal(t, expectedNamespaces, namespaces)
}

func TestFunctions_GetNamespace(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/functions/namespaces/123-abc", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"namespace": {
				"api_host": "https://faas.do.com",
				"namespace": "123-abc",
				"created_at": "2022-06-16T12:09:13Z",
				"updated_at": "2022-06-16T12:09:13Z",
				"label": "my-namespace-1",
				"region": "nyc1",
				"uuid": "123-456",
				"key": "abc-123"
			}
		}`)
	})

	namespace, _, err := client.Functions.GetNamespace(ctx, "123-abc")
	require.NoError(t, err)

	expectedNamespace := &FunctionsNamespace{
		ApiHost:   "https://faas.do.com",
		Namespace: "123-abc",
		CreatedAt: time.Date(2022, 6, 16, 12, 9, 13, 0, time.UTC),
		UpdatedAt: time.Date(2022, 6, 16, 12, 9, 13, 0, time.UTC),
		Label:     "my-namespace-1",
		Region:    "nyc1",
		UUID:      "123-456",
		Key:       "abc-123",
	}
	assert.Equal(t, expectedNamespace, namespace)
}

func TestFunctions_CreateNamespace(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/functions/namespaces", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `{
			"namespace": {
				"api_host": "https://faas.do.com",
				"namespace": "123-abc",
				"created_at": "2022-06-16T12:09:13Z",
				"updated_at": "2022-06-16T12:09:13Z",
				"label": "my-namespace-1",
				"region": "nyc1",
				"uuid": "123-456",
				"key": "abc-123"
			}
		}`)
	})

	opts := FunctionsNamespaceCreateRequest{Label: "my-namespace-1", Region: "nyc1"}
	namespace, _, err := client.Functions.CreateNamespace(ctx, &opts)
	require.NoError(t, err)

	expectedNamespace := &FunctionsNamespace{
		ApiHost:   "https://faas.do.com",
		Namespace: "123-abc",
		CreatedAt: time.Date(2022, 6, 16, 12, 9, 13, 0, time.UTC),
		UpdatedAt: time.Date(2022, 6, 16, 12, 9, 13, 0, time.UTC),
		Label:     "my-namespace-1",
		Region:    "nyc1",
		UUID:      "123-456",
		Key:       "abc-123",
	}
	assert.Equal(t, expectedNamespace, namespace)
}

func TestFunctions_DeleteNamespace(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/functions/namespaces/123-abc", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
	})

	_, err := client.Functions.DeleteNamespace(ctx, "123-abc")

	assert.NoError(t, err)
}
