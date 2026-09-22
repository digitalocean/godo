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

var hostedAgentCheckpoint = HostedAgentCheckpoint{
	CheckpointID: "cp_9f2c1a4b",
	SessionID:    "sess-abc123",
	Status:       HostedAgentCheckpointStatusReady,
	Kind:         HostedAgentCheckpointKindExplicit,
	Label:        "before-refactor",
	EventID:      "01JEVENTULID00000000000000",
	SizeBytes:    734003200,
	CreatedAt:    Timestamp{Time: time.Date(2026, 8, 12, 21, 4, 5, 123000000, time.UTC)},
}

var hostedAgentCheckpointJSON = `
{
	"checkpoint_id": "cp_9f2c1a4b",
	"session_id": "sess-abc123",
	"status": "READY",
	"kind": "explicit",
	"label": "before-refactor",
	"event_id": "01JEVENTULID00000000000000",
	"size_bytes": 734003200,
	"created_at": "2026-08-12T21:04:05.123Z"
}
`

func TestHostedAgents_CreateCheckpoint(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/sessions/sess-abc123/checkpoints", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body HostedAgentCheckpointCreateRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		assert.Equal(t, "before-refactor", body.Label)
		fmt.Fprintf(w, `{"checkpoint":%s}`, hostedAgentCheckpointJSON)
	})

	got, resp, err := client.HostedAgents.CreateCheckpoint(ctx, "sess-abc123", &HostedAgentCheckpointCreateRequest{
		Label: "before-refactor",
	})
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, hostedAgentCheckpoint, *got)
}

func TestHostedAgents_CreateCheckpoint_NilBody(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/sessions/sess-abc123/checkpoints", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body HostedAgentCheckpointCreateRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		assert.Empty(t, body.Label)
		fmt.Fprintf(w, `{"checkpoint":%s}`, hostedAgentCheckpointJSON)
	})

	got, _, err := client.HostedAgents.CreateCheckpoint(ctx, "sess-abc123", nil)
	require.NoError(t, err)
	assert.Equal(t, "cp_9f2c1a4b", got.CheckpointID)
}

func TestHostedAgents_CreateCheckpoint_EmptySessionID(t *testing.T) {
	_, _, err := client.HostedAgents.CreateCheckpoint(ctx, "", nil)
	require.Error(t, err)
}

func TestHostedAgents_ListCheckpoints(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/sessions/sess-abc123/checkpoints", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		assert.Equal(t, "25", r.URL.Query().Get("page_size"))
		assert.Equal(t, "tok-1", r.URL.Query().Get("page_token"))
		fmt.Fprintf(w, `{"checkpoints":[%s],"next_page_token":"tok-2"}`, hostedAgentCheckpointJSON)
	})

	got, resp, err := client.HostedAgents.ListCheckpoints(ctx, "sess-abc123", &HostedAgentCheckpointListOptions{
		PageSize:  25,
		PageToken: "tok-1",
	})
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	require.Len(t, got.Checkpoints, 1)
	assert.Equal(t, hostedAgentCheckpoint, got.Checkpoints[0])
	assert.Equal(t, "tok-2", got.NextPageToken)
}

func TestHostedAgents_GetCheckpoint(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/sessions/sess-abc123/checkpoints/cp_9f2c1a4b", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprintf(w, `{"checkpoint":%s}`, hostedAgentCheckpointJSON)
	})

	got, _, err := client.HostedAgents.GetCheckpoint(ctx, "sess-abc123", "cp_9f2c1a4b")
	require.NoError(t, err)
	assert.Equal(t, hostedAgentCheckpoint, *got)
}

func TestHostedAgents_DeleteCheckpoint(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/sessions/sess-abc123/checkpoints/cp_9f2c1a4b", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		fmt.Fprint(w, `{"checkpoint_id":"cp_9f2c1a4b","deleted":true}`)
	})

	got, _, err := client.HostedAgents.DeleteCheckpoint(ctx, "sess-abc123", "cp_9f2c1a4b")
	require.NoError(t, err)
	assert.Equal(t, "cp_9f2c1a4b", got.CheckpointID)
	assert.True(t, got.Deleted)
}

func TestHostedAgents_ForkSession(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/sessions/sess-abc123/fork", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body HostedAgentForkSessionRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		assert.Equal(t, "cp_9f2c1a4b", body.FromCheckpointID)
		assert.Equal(t, 2, body.Count)
		fmt.Fprint(w, `{
			"sessions": [
				{
					"session_id": "sess_fork_a",
					"team_id": 42,
					"agent_kind": "AGENT_KIND_CLAUDE_CODE",
					"status": "SESSION_STATUS_READY",
					"created_at": "2026-08-12T21:10:00Z",
					"last_event_at": "2026-08-12T21:10:00Z",
					"parent_session_id": "sess-abc123",
					"fork_id": "fork-aaa"
				},
				{
					"session_id": "sess_fork_b",
					"team_id": 42,
					"agent_kind": "AGENT_KIND_CLAUDE_CODE",
					"status": "SESSION_STATUS_READY",
					"created_at": "2026-08-12T21:10:00Z",
					"last_event_at": "2026-08-12T21:10:00Z",
					"parent_session_id": "sess-abc123",
					"fork_id": "fork-bbb"
				}
			]
		}`)
	})

	got, _, err := client.HostedAgents.ForkSession(ctx, "sess-abc123", &HostedAgentForkSessionRequest{
		FromCheckpointID: "cp_9f2c1a4b",
		Count:            2,
	})
	require.NoError(t, err)
	require.Len(t, got.Sessions, 2)
	assert.Equal(t, "sess_fork_a", got.Sessions[0].SessionID)
	assert.Equal(t, "sess-abc123", got.Sessions[0].ParentSessionID)
	assert.Equal(t, "fork-aaa", got.Sessions[0].ForkID)
	assert.Equal(t, "sess_fork_b", got.Sessions[1].SessionID)
}

func TestHostedAgents_ForkSession_InvalidCount(t *testing.T) {
	_, _, err := client.HostedAgents.ForkSession(ctx, "sess-abc123", &HostedAgentForkSessionRequest{Count: 5})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "fork count")
}

func TestHostedAgents_RollbackToCheckpoint(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/sessions/sess-abc123/checkpoints/cp_9f2c1a4b/rollback", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprintf(w, `{"session":%s}`, hostedAgentSessionJSON)
	})

	got, _, err := client.HostedAgents.RollbackToCheckpoint(ctx, "sess-abc123", "cp_9f2c1a4b")
	require.NoError(t, err)
	assert.Equal(t, hostedAgentSession, *got)
}

func TestHostedAgents_ListSessions_ParentSessionID(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/sessions", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		assert.Equal(t, "sess-abc123", r.URL.Query().Get("parent_session_id"))
		fmt.Fprint(w, `{
			"sessions": [{
				"session_id": "sess_fork_a",
				"team_id": 42,
				"agent_kind": "AGENT_KIND_CLAUDE_CODE",
				"status": "SESSION_STATUS_READY",
				"created_at": "2026-08-12T21:10:00Z",
				"last_event_at": "2026-08-12T21:10:00Z",
				"parent_session_id": "sess-abc123",
				"fork_id": "fork-aaa"
			}],
			"next_page_token": ""
		}`)
	})

	got, _, err := client.HostedAgents.ListSessions(ctx, &HostedAgentSessionListOptions{
		ParentSessionID: "sess-abc123",
	})
	require.NoError(t, err)
	require.Len(t, got.Sessions, 1)
	assert.Equal(t, "sess-abc123", got.Sessions[0].ParentSessionID)
	assert.Equal(t, "fork-aaa", got.Sessions[0].ForkID)
}
