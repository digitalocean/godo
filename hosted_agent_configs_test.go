package godo

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var hostedAgentConfig = HostedAgentConfig{
	ID:                     "019fb39c-14d9-7080-933e-b9b90e25acda",
	Name:                   "support-agent",
	AgentSpecSchemaVersion: "agents.digitalocean.com/v1alpha1",
	Manifest:               json.RawMessage(`{"apiVersion":"agents.digitalocean.com/v1alpha1","kind":"Agent"}`),
	ContentHash:            "75803fef24dc731824ecd4a1853c76153c7d8503534092b94da3ca3f31a882f4",
	CreatedBy:              "user-1",
	CreatedAt:              Timestamp{Time: time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)},
	UpdatedAt:              Timestamp{Time: time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)},
	Credentials: []HostedAgentConfigCredentialSlot{
		{Name: "OPENAI_API_KEY", Source: "tenantSecret", Provider: "openai"},
	},
}

const hostedAgentConfigJSON = `
{
	"id": "019fb39c-14d9-7080-933e-b9b90e25acda",
	"name": "support-agent",
	"agentspec_schema_version": "agents.digitalocean.com/v1alpha1",
	"manifest": {"apiVersion":"agents.digitalocean.com/v1alpha1","kind":"Agent"},
	"content_hash": "75803fef24dc731824ecd4a1853c76153c7d8503534092b94da3ca3f31a882f4",
	"created_by": "user-1",
	"created_at": "2026-08-01T12:00:00Z",
	"updated_at": "2026-08-01T12:00:00Z",
	"credentials": [
		{"name":"OPENAI_API_KEY","source":"tenantSecret","provider":"openai"}
	]
}
`

func TestHostedAgents_ListAgentConfigs(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/configs", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		assert.Equal(t, "10", r.URL.Query().Get("page_size"))
		fmt.Fprint(w, `{"configs":[{"id":"019fb39c-14d9-7080-933e-b9b90e25acda","name":"support-agent","agentspec_schema_version":"agents.digitalocean.com/v1alpha1","content_hash":"75803fef24dc731824ecd4a1853c76153c7d8503534092b94da3ca3f31a882f4","created_by":"user-1","created_at":"2026-08-01T12:00:00Z","updated_at":"2026-08-01T12:00:00Z"}],"next_page_token":"abc"}`)
	})

	got, resp, err := client.HostedAgents.ListAgentConfigs(ctx, &HostedAgentConfigListOptions{PageSize: 10})
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	require.Len(t, got.Configs, 1)
	assert.Equal(t, "support-agent", got.Configs[0].Name)
	assert.Equal(t, "abc", got.NextPageToken)
}

func TestHostedAgents_GetAgentConfig(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/configs/019fb39c-14d9-7080-933e-b9b90e25acda", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprintf(w, `{"config":%s}`, hostedAgentConfigJSON)
	})

	got, resp, err := client.HostedAgents.GetAgentConfig(ctx, hostedAgentConfig.ID)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, hostedAgentConfig.ID, got.ID)
	assert.Equal(t, hostedAgentConfig.Name, got.Name)
	require.Len(t, got.Credentials, 1)
	assert.Equal(t, "OPENAI_API_KEY", got.Credentials[0].Name)
	assert.Equal(t, "tenantSecret", got.Credentials[0].Source)
	assert.NotContains(t, string(mustMarshal(t, got)), "plaintext")
	assert.NotContains(t, string(mustMarshal(t, got)), `"configured"`)
}

func TestHostedAgents_CreateAgentConfig(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/configs", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body HostedAgentConfigCreateRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		assert.Equal(t, "support-agent", body.Name)
		assert.Contains(t, body.ManifestYAML, "apiVersion:")
		// Credentials are declared inline in the manifest under spec.secrets;
		// there is no separate secrets field on the request body anymore.
		assert.Contains(t, body.ManifestYAML, "spec:")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, `{"config":%s}`, hostedAgentConfigJSON)
	})

	got, resp, err := client.HostedAgents.CreateAgentConfig(ctx, &HostedAgentConfigCreateRequest{
		Name: "support-agent",
		ManifestYAML: "apiVersion: agents.digitalocean.com/v1alpha1\n" +
			"kind: Agent\n" +
			"spec:\n" +
			"  secrets:\n" +
			"    - name: OPENAI_API_KEY\n" +
			"      source: tenantSecret\n" +
			"      value: sk-test\n",
	})
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	assert.Equal(t, hostedAgentConfig.ID, got.ID)
}

func TestHostedAgents_DeleteAgentConfig(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/configs/019fb39c-14d9-7080-933e-b9b90e25acda", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		w.WriteHeader(http.StatusNoContent)
	})

	resp, err := client.HostedAgents.DeleteAgentConfig(ctx, hostedAgentConfig.ID)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestHostedAgents_ListAgentConfigSessions(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/configs/019fb39c-14d9-7080-933e-b9b90e25acda/sessions", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		assert.Equal(t, "SESSION_STATUS_READY", r.URL.Query().Get("status"))
		fmt.Fprintf(w, `{"sessions":[%s],"next_page_token":""}`, hostedAgentSessionJSON)
	})

	got, resp, err := client.HostedAgents.ListAgentConfigSessions(ctx, hostedAgentConfig.ID, &HostedAgentSessionListOptions{
		Status: HostedAgentSessionStatusReady,
	})
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	require.Len(t, got.Sessions, 1)
	assert.Equal(t, "sess-abc123", got.Sessions[0].SessionID)
}

func mustMarshal(t *testing.T, v any) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return b
}
