package godo

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHostedAgents_ListRegions(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/harness-runtime/regions", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		assert.Empty(t, r.URL.RawQuery)
		fmt.Fprint(w, `{
			"regions": [
				{"slug": "mem1", "is_default": false},
				{"slug": "ric1", "is_default": true}
			]
		}`)
	})

	got, resp, err := client.HostedAgents.ListRegions(ctx)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, []HostedAgentRegion{
		{Slug: "mem1", IsDefault: false},
		{Slug: "ric1", IsDefault: true},
	}, got.Regions)
}

// A team with no grants still sees the default region, so a picker built from
// the catalog is never empty.
func TestHostedAgents_ListRegions_DefaultOnly(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/harness-runtime/regions", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{"regions": [{"slug": "ric1", "is_default": true}]}`)
	})

	got, _, err := client.HostedAgents.ListRegions(ctx)
	require.NoError(t, err)
	require.Len(t, got.Regions, 1)
	assert.Equal(t, "ric1", got.Regions[0].Slug)
	assert.True(t, got.Regions[0].IsDefault)
}

func TestHostedAgents_ListRegions_Error(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/harness-runtime/regions", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, `{"id":"unauthorized","message":"Unable to authenticate you."}`)
	})

	got, resp, err := client.HostedAgents.ListRegions(ctx)
	require.Error(t, err)
	assert.Nil(t, got)
	require.NotNil(t, resp)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestHostedAgentRegionsResponse_DefaultRegion(t *testing.T) {
	tests := []struct {
		name     string
		in       *HostedAgentRegionsResponse
		wantSlug string
		wantOK   bool
	}{
		{
			name: "the marked region wins regardless of position",
			in: &HostedAgentRegionsResponse{Regions: []HostedAgentRegion{
				{Slug: "mem1"},
				{Slug: "ric1", IsDefault: true},
			}},
			wantSlug: "ric1",
			wantOK:   true,
		},
		{
			name:     "a catalog that marks none reports none, rather than guessing the first",
			in:       &HostedAgentRegionsResponse{Regions: []HostedAgentRegion{{Slug: "mem1"}}},
			wantSlug: "",
			wantOK:   false,
		},
		{
			name:     "empty",
			in:       &HostedAgentRegionsResponse{},
			wantSlug: "",
			wantOK:   false,
		},
		{
			name:     "nil receiver",
			in:       nil,
			wantSlug: "",
			wantOK:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			slug, ok := tt.in.DefaultRegion()
			assert.Equal(t, tt.wantSlug, slug)
			assert.Equal(t, tt.wantOK, ok)
		})
	}
}
