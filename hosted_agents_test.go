package godo

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHostedAgents_CreateSession_WithOrigin(t *testing.T) {
	setup()
	defer teardown()

	wantCreate := &HostedAgentSessionCreateRequest{
		AgentKind: HostedAgentKindClaudeCode,
		RepoHint:  "github.com/digitalocean/example",
		Origin: &HostedAgentSessionOriginRequest{
			Product:    HostedAgentSessionOriginProductSimulation,
			ResourceID: "sim-run-123",
		},
	}

	mux.HandleFunc("/v2/agents/sessions", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)

		var got HostedAgentSessionCreateRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&got))
		assert.Equal(t, wantCreate.AgentKind, got.AgentKind)
		assert.Equal(t, wantCreate.RepoHint, got.RepoHint)
		require.NotNil(t, got.Origin)
		assert.Equal(t, HostedAgentSessionOriginProductSimulation, got.Origin.Product)
		assert.Equal(t, "sim-run-123", got.Origin.ResourceID)

		// Request wire must not include verified.
		raw, err := json.Marshal(got.Origin)
		require.NoError(t, err)
		assert.NotContains(t, string(raw), "verified")

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{
			"session": {
				"session_id": "sess-1",
				"name": "session-claude-code-2026-07-27",
				"team_id": 42,
				"agent_kind": "AGENT_KIND_CLAUDE_CODE",
				"status": "SESSION_STATUS_PROVISIONING",
				"created_at": "2026-07-27T12:00:00Z",
				"last_event_at": "2026-07-27T12:00:00Z",
				"repo_hint": "github.com/digitalocean/example",
				"origin": {
					"product": "simulation",
					"resource_id": "sim-run-123",
					"verified": true
				}
			}
		}`)
	})

	session, resp, err := client.HostedAgents.CreateSession(ctx, wantCreate)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, session)
	assert.Equal(t, "sess-1", session.SessionID)
	require.NotNil(t, session.Origin)
	assert.Equal(t, HostedAgentSessionOriginProductSimulation, session.Origin.Product)
	assert.Equal(t, "sim-run-123", session.Origin.ResourceID)
	assert.True(t, session.Origin.Verified)
}

func TestHostedAgents_CreateSession_DirectOmitsOrigin(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/sessions", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)

		var raw map[string]json.RawMessage
		require.NoError(t, json.NewDecoder(r.Body).Decode(&raw))
		_, hasOrigin := raw["origin"]
		assert.False(t, hasOrigin, "omitted Origin should not appear in JSON body")

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{
			"session": {
				"session_id": "sess-direct",
				"team_id": 7,
				"agent_kind": "AGENT_KIND_OPENCODE",
				"status": "SESSION_STATUS_READY",
				"created_at": "2026-07-27T12:00:00Z",
				"last_event_at": "2026-07-27T12:00:00Z",
				"origin": {
					"product": "direct",
					"verified": true
				}
			}
		}`)
	})

	session, _, err := client.HostedAgents.CreateSession(ctx, &HostedAgentSessionCreateRequest{
		AgentKind: HostedAgentKindOpenCode,
	})
	require.NoError(t, err)
	require.NotNil(t, session.Origin)
	assert.Equal(t, HostedAgentSessionOriginProductDirect, session.Origin.Product)
	assert.Empty(t, session.Origin.ResourceID)
	assert.True(t, session.Origin.Verified)
}

func TestHostedAgents_GetSession_WithOrigin(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/sessions/sess-eval", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{
			"session": {
				"session_id": "sess-eval",
				"team_id": 9,
				"agent_kind": "AGENT_KIND_CODEX_CLI",
				"status": "SESSION_STATUS_READY",
				"created_at": "2026-07-27T12:00:00Z",
				"last_event_at": "2026-07-27T12:05:00Z",
				"origin": {
					"product": "evaluation",
					"resource_id": "eval-456",
					"verified": true
				}
			}
		}`)
	})

	session, _, err := client.HostedAgents.GetSession(ctx, "sess-eval")
	require.NoError(t, err)
	require.NotNil(t, session.Origin)
	assert.Equal(t, HostedAgentSessionOriginProductEvaluation, session.Origin.Product)
	assert.Equal(t, "eval-456", session.Origin.ResourceID)
	assert.True(t, session.Origin.Verified)
}

func TestHostedAgents_ListSessions(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/sessions", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "direct-only", r.URL.Query().Get("name"))
		assert.Equal(t, "10", r.URL.Query().Get("page_size"))

		w.Header().Set("Content-Type", "application/json")
		// Server-driven list omits simulation/evaluation; fixture mirrors that.
		fmt.Fprint(w, `{
			"sessions": [{
				"session_id": "sess-1",
				"name": "direct-only",
				"team_id": 1,
				"agent_kind": "AGENT_KIND_CLAUDE_CODE",
				"status": "SESSION_STATUS_READY",
				"created_at": "2026-07-27T12:00:00Z",
				"last_event_at": "2026-07-27T12:00:00Z",
				"origin": {"product": "direct", "verified": true}
			}],
			"next_page_token": ""
		}`)
	})

	list, _, err := client.HostedAgents.ListSessions(ctx, &HostedAgentSessionListOptions{
		Name:     "direct-only",
		PageSize: 10,
	})
	require.NoError(t, err)
	require.Len(t, list.Sessions, 1)
	assert.Equal(t, "sess-1", list.Sessions[0].SessionID)
	require.NotNil(t, list.Sessions[0].Origin)
	assert.Equal(t, HostedAgentSessionOriginProductDirect, list.Sessions[0].Origin.Product)
}

func TestHostedAgentSessionOriginRequest_JSONOmitsVerified(t *testing.T) {
	body, err := json.Marshal(&HostedAgentSessionOriginRequest{
		Product:    HostedAgentSessionOriginProductEvaluation,
		ResourceID: "eval-1",
	})
	require.NoError(t, err)
	assert.JSONEq(t, `{"product":"evaluation","resource_id":"eval-1"}`, string(body))
}

func TestHostedAgentSessionOrigin_JSONIncludesVerified(t *testing.T) {
	body, err := json.Marshal(&HostedAgentSessionOrigin{
		Product:  HostedAgentSessionOriginProductDirect,
		Verified: true,
	})
	require.NoError(t, err)
	assert.JSONEq(t, `{"product":"direct","verified":true}`, string(body))
}

func TestHostedAgents_CreateSession_Validation(t *testing.T) {
	setup()
	defer teardown()

	_, _, err := client.HostedAgents.CreateSession(ctx, nil)
	require.Error(t, err)

	_, _, err = client.HostedAgents.CreateSession(ctx, &HostedAgentSessionCreateRequest{})
	require.Error(t, err)
}

func TestHostedAgents_DestroySession(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/sessions/sess-1", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		w.WriteHeader(http.StatusNoContent)
	})

	resp, err := client.HostedAgents.DestroySession(ctx, "sess-1")
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}
