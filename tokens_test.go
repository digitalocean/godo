package godo

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTokens_Get(t *testing.T) {
	tests := []struct {
		id       int
		body     string
		expected *Token
	}{
		{
			id: 123,
			body: `
            {
              "token": {
                "id":123,
                "name":"droplets-read-token",
                "scopes":["droplet:read"],
                "created_at":"2022-12-05T20:32:15Z",
                "last_used_at":"2022-12-05",
                "expiry_seconds":604800
              }
            }`,
			expected: &Token{
				ID:            123,
				Name:          "droplets-read-token",
				Scopes:        []string{"droplet:read"},
				CreatedAt:     time.Date(2022, time.December, 5, 20, 32, 15, 0, time.UTC),
				LastUsedAt:    time.Date(2022, time.December, 5, 20, 32, 15, 0, time.UTC).Format("2006-01-02"),
				ExpirySeconds: PtrTo(604800),
			},
		},
		{
			id: 456,
			body: `
            {
              "token": {
                "id":456,
                "name":"droplets-read-token-without-expiry",
                "scopes":["droplet:read"],
                "created_at":"2022-12-05T20:32:15Z",
                "last_used_at":"2022-12-05"
              }
            }`,
			expected: &Token{
				ID:         456,
				Name:       "droplets-read-token-without-expiry",
				Scopes:     []string{"droplet:read"},
				CreatedAt:  time.Date(2022, time.December, 5, 20, 32, 15, 0, time.UTC),
				LastUsedAt: time.Date(2022, time.December, 5, 20, 32, 15, 0, time.UTC).Format("2006-01-02"),
			},
		},
	}

	for _, tt := range tests {
		setup()
		defer teardown()

		mux.HandleFunc(fmt.Sprintf("/v2/tokens/%d", tt.id), func(w http.ResponseWriter, r *http.Request) {
			testMethod(t, r, http.MethodGet)
			fmt.Fprint(w, tt.body)
		})

		got, _, err := client.Tokens.Get(ctx, tt.id)
		require.NoError(t, err)
		require.Equal(t, tt.expected, got)
	}
}

func TestTokens_List(t *testing.T) {
	setup()
	defer teardown()

	expected := []Token{
		{
			ID:            123,
			Name:          "droplets-read-token",
			Scopes:        []string{"droplet:read"},
			CreatedAt:     time.Date(2022, time.December, 5, 20, 32, 15, 0, time.UTC),
			LastUsedAt:    time.Date(2022, time.December, 5, 20, 32, 15, 0, time.UTC).Format("2006-01-02"),
			ExpirySeconds: PtrTo(604800),
		},
		{
			ID:         456,
			Name:       "droplets-read-token-without-expiry",
			Scopes:     []string{"droplet:read"},
			CreatedAt:  time.Date(2022, time.December, 5, 20, 32, 15, 0, time.UTC),
			LastUsedAt: time.Date(2022, time.December, 5, 20, 32, 15, 0, time.UTC).Format("2006-01-02"),
		},
	}

	body := `
            {
              "tokens": [
              {
                "id":123,
                "name":"droplets-read-token",
                "scopes":["droplet:read"],
                "created_at":"2022-12-05T20:32:15Z",
                "last_used_at":"2022-12-05",
                "expiry_seconds":604800
              },
              {
                "id":456,
                "name":"droplets-read-token-without-expiry",
                "scopes":["droplet:read"],
                "created_at":"2022-12-05T20:32:15Z",
                "last_used_at":"2022-12-05"
              }
            ],
            "links": {
              "pages": {}
            },
            "meta": {
              "total": 2
            }
          }`

	mux.HandleFunc("/v2/tokens", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)

		params := r.URL.Query()
		assert.Equal(t, "200", params.Get("per_page"))
		assert.Equal(t, "1", params.Get("page"))

		fmt.Fprint(w, body)
	})

	got, resp, err := client.Tokens.List(ctx, &ListOptions{PerPage: 200, Page: 1})
	require.NoError(t, err)
	require.Equal(t, expected, got)
	require.Equal(t, &Meta{Total: 2}, resp.Meta)
}

func TestTokens_Create(t *testing.T) {
	tests := []struct {
		request     *TokenCreateRequest
		expectedReq string
		response    string
		expected    *Token
	}{
		{
			request: &TokenCreateRequest{
				Name:          "droplets-read-token",
				Scopes:        []string{"droplet:read"},
				ExpirySeconds: PtrTo(604800),
			},
			expectedReq: `{
                "name":"droplets-read-token",
                "scopes":["droplet:read"],
                "expiry_seconds":604800
            }`,
			response: `
            {
              "token": {
                "id":123,
                "name":"droplets-read-token",
                "scopes":["droplet:read"],
                "created_at":"2022-12-05T20:32:15Z",
                "last_used_at":"2022-12-05",
                "expiry_seconds":604800,
                "access_token": "dop_v1_foo"
              }
            }`,
			expected: &Token{
				ID:            123,
				Name:          "droplets-read-token",
				Scopes:        []string{"droplet:read"},
				CreatedAt:     time.Date(2022, time.December, 5, 20, 32, 15, 0, time.UTC),
				LastUsedAt:    time.Date(2022, time.December, 5, 20, 32, 15, 0, time.UTC).Format("2006-01-02"),
				ExpirySeconds: PtrTo(604800),
				AccessToken:   "dop_v1_foo",
			},
		},
		{
			request: &TokenCreateRequest{
				Name:   "droplets-read-token",
				Scopes: []string{"droplet:read"},
			},
			expectedReq: `{
                "name":"droplets-read-token",
                "scopes":["droplet:read"]
            }`,
			response: `
            {
              "token": {
                "id":456,
                "name":"droplets-read-token-without-expiry",
                "scopes":["droplet:read"],
                "created_at":"2022-12-05T20:32:15Z",
                "last_used_at":"2022-12-05",
                "access_token": "dop_v1_foo"
              }
            }`,
			expected: &Token{
				ID:          456,
				Name:        "droplets-read-token-without-expiry",
				Scopes:      []string{"droplet:read"},
				CreatedAt:   time.Date(2022, time.December, 5, 20, 32, 15, 0, time.UTC),
				LastUsedAt:  time.Date(2022, time.December, 5, 20, 32, 15, 0, time.UTC).Format("2006-01-02"),
				AccessToken: "dop_v1_foo",
			},
		},
	}

	for _, tt := range tests {
		setup()
		defer teardown()

		mux.HandleFunc("/v2/tokens", func(w http.ResponseWriter, r *http.Request) {
			testMethod(t, r, http.MethodPost)

			defer r.Body.Close()
			bodyBytes, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			cleanedExpect := strings.Replace(tt.expectedReq, " ", "", -1)
			cleanedExpect = strings.Replace(cleanedExpect, "\n", "", -1)
			cleanedReq := strings.Replace(string(bodyBytes), "\n", "", -1)
			assert.Equal(t, cleanedExpect, cleanedReq)

			fmt.Fprint(w, tt.response)
		})

		got, _, err := client.Tokens.Create(ctx, tt.request)
		require.NoError(t, err)
		require.Equal(t, tt.expected, got)
	}
}

func TestTokens_Update(t *testing.T) {
	tests := []struct {
		id          int
		request     *TokenUpdateRequest
		expectedReq string
		response    string
		expected    *Token
	}{
		{
			id: 123,
			request: &TokenUpdateRequest{
				Name: "updated-name",
			},
			expectedReq: `{"name":"updated-name"}`,
			response: `
            {
              "token": {
                "id":123,
                "name":"updated-name",
                "scopes":["droplet:read"],
                "created_at":"2022-12-05T20:32:15Z",
                "last_used_at":"2022-12-05",
                "expiry_seconds":604800
              }
            }`,
			expected: &Token{
				ID:            123,
				Name:          "updated-name",
				Scopes:        []string{"droplet:read"},
				CreatedAt:     time.Date(2022, time.December, 5, 20, 32, 15, 0, time.UTC),
				LastUsedAt:    time.Date(2022, time.December, 5, 20, 32, 15, 0, time.UTC).Format("2006-01-02"),
				ExpirySeconds: PtrTo(604800),
			},
		},
		{
			id: 456,
			request: &TokenUpdateRequest{
				Scopes: []string{"droplet:create"},
			},
			expectedReq: `{
                "scopes":["droplet:create"]
            }`,
			response: `
            {
              "token": {
                "id":456,
                "name":"droplets-read-token-without-expiry",
                "scopes":["droplet:create"],
                "created_at":"2022-12-05T20:32:15Z",
                "last_used_at":"2022-12-05"
              }
            }`,
			expected: &Token{
				ID:         456,
				Name:       "droplets-read-token-without-expiry",
				Scopes:     []string{"droplet:create"},
				CreatedAt:  time.Date(2022, time.December, 5, 20, 32, 15, 0, time.UTC),
				LastUsedAt: time.Date(2022, time.December, 5, 20, 32, 15, 0, time.UTC).Format("2006-01-02"),
			},
		},
		{
			id: 789,
			request: &TokenUpdateRequest{
				Name:   "both-updated",
				Scopes: []string{"droplet:create"},
			},
			expectedReq: `{
                "name": "both-updated",
                "scopes":["droplet:create"]
            }`,
			response: `
            {
              "token": {
                "id":456,
                "name":"both-updated",
                "scopes":["droplet:create"],
                "created_at":"2022-12-05T20:32:15Z",
                "last_used_at":"2022-12-05"
              }
            }`,
			expected: &Token{
				ID:         456,
				Name:       "both-updated",
				Scopes:     []string{"droplet:create"},
				CreatedAt:  time.Date(2022, time.December, 5, 20, 32, 15, 0, time.UTC),
				LastUsedAt: time.Date(2022, time.December, 5, 20, 32, 15, 0, time.UTC).Format("2006-01-02"),
			},
		},
	}

	for _, tt := range tests {
		setup()
		defer teardown()

		mux.HandleFunc(fmt.Sprintf("/v2/tokens/%d", tt.id), func(w http.ResponseWriter, r *http.Request) {
			testMethod(t, r, http.MethodPatch)

			defer r.Body.Close()
			bodyBytes, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			cleanedExpect := strings.Replace(tt.expectedReq, " ", "", -1)
			cleanedExpect = strings.Replace(cleanedExpect, "\n", "", -1)
			cleanedReq := strings.Replace(string(bodyBytes), "\n", "", -1)
			assert.Equal(t, cleanedExpect, cleanedReq)

			fmt.Fprint(w, tt.response)
		})

		got, _, err := client.Tokens.Update(ctx, tt.id, tt.request)
		require.NoError(t, err)
		require.Equal(t, tt.expected, got)
	}
}

func TestTokens_Delete(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/tokens/123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		fmt.Fprint(w, http.NoBody)
	})

	_, err := client.Tokens.Revoke(ctx, 123)
	require.NoError(t, err)
}

func TestTokens_ListScopes(t *testing.T) {
	setup()
	defer teardown()

	expected := []TokenScope{
		{
			Name: "account:read",
		},
		{
			Name: "droplet:create",
		},
		{
			Name: "droplet:delete",
		},
	}

	body := `
            {
              "scopes": [
                {
                  "name":"account:read"
                },
                {
                  "name":"droplet:create"
                },
                {
                  "name":"droplet:delete"
                }
              ],
              "links": {
                "pages": {}
              },
              "meta": {
                "total": 3
               }
            }`

	mux.HandleFunc("/v2/tokens/scopes", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)

		params := r.URL.Query()
		assert.Equal(t, "200", params.Get("per_page"))
		assert.Equal(t, "1", params.Get("page"))

		fmt.Fprint(w, body)
	})

	got, resp, err := client.Tokens.ListScopes(ctx, &ListOptions{PerPage: 200, Page: 1})
	require.NoError(t, err)
	require.Equal(t, expected, got)
	require.Equal(t, &Meta{Total: 3}, resp.Meta)
}

func TestTokens_ListScopesByNamespace(t *testing.T) {
	setup()
	defer teardown()

	expected := []TokenScope{
		{
			Name: "droplet:create",
		},
		{
			Name: "droplet:delete",
		},
	}

	body := `
            {
              "scopes": [
                {
                  "name":"droplet:create"
                },
                {
                  "name":"droplet:delete"
                }
              ],
              "links": {
                "pages": {}
              },
              "meta": {
                "total": 2
               }
            }`

	mux.HandleFunc("/v2/tokens/scopes", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)

		params := r.URL.Query()
		assert.Equal(t, "200", params.Get("per_page"))
		assert.Equal(t, "1", params.Get("page"))
		assert.Equal(t, "droplet", params.Get("namespace"))

		fmt.Fprint(w, body)
	})

	got, resp, err := client.Tokens.ListScopesByNamespace(ctx, "droplet", &ListOptions{PerPage: 200, Page: 1})
	require.NoError(t, err)
	require.Equal(t, expected, got)
	require.Equal(t, &Meta{Total: 2}, resp.Meta)
}
