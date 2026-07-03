package godo

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var hostedAgentSession = HostedAgentSession{
	SessionID:   "sess-abc123",
	Name:        "godo-fixture",
	TeamID:      42,
	AgentKind:   HostedAgentKindClaudeCode,
	Status:      HostedAgentSessionStatusReady,
	CreatedAt:   Timestamp{Time: time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)},
	LastEventAt: Timestamp{Time: time.Date(2026, 3, 1, 12, 5, 0, 0, time.UTC)},
	RepoHint:    "digitalocean/godo",
	ProviderAuth: map[string]HostedAgentProviderAuthState{
		"github": HostedAgentProviderAuthStateAuthorized,
	},
}

var hostedAgentSessionJSON = `
{
	"session_id": "sess-abc123",
	"name": "godo-fixture",
	"team_id": 42,
	"agent_kind": "AGENT_KIND_CLAUDE_CODE",
	"status": "SESSION_STATUS_READY",
	"created_at": "2026-03-01T12:00:00Z",
	"last_event_at": "2026-03-01T12:05:00Z",
	"repo_hint": "digitalocean/godo",
	"provider_auth": {
		"github": "PROVIDER_AUTH_STATE_AUTHORIZED"
	}
}
`

func TestHostedAgents_CreateSession(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/sessions", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body HostedAgentSessionCreateRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		assert.Equal(t, HostedAgentKindClaudeCode, body.AgentKind)
		assert.Equal(t, "digitalocean/godo", body.RepoHint)
		fmt.Fprintf(w, `{"session":%s}`, hostedAgentSessionJSON)
	})

	got, resp, err := client.HostedAgents.CreateSession(ctx, &HostedAgentSessionCreateRequest{
		AgentKind: HostedAgentKindClaudeCode,
		RepoHint:  "digitalocean/godo",
	})
	require.NoError(t, err)
	assert.Equal(t, hostedAgentSession, *got)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHostedAgents_CreateSessionFromManifest(t *testing.T) {
	setup()
	defer teardown()

	const manifest = `apiVersion: agents.digitalocean.com/v1alpha1
kind: Agent
metadata:
  name: opencode-coding
spec:
  runtime:
    adapter: opencode
  sandbox:
    template: coding
`

	mux.HandleFunc("/v2/agents/sessions", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		assert.Equal(t, "application/x-yaml", r.Header.Get("Content-Type"))
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		assert.Contains(t, string(body), "agents.digitalocean.com/v1alpha1")
		assert.Contains(t, string(body), "adapter: opencode")
		fmt.Fprintf(w, `{"session":%s}`, hostedAgentSessionJSON)
	})

	got, resp, err := client.HostedAgents.CreateSessionFromManifest(ctx, []byte(manifest))
	require.NoError(t, err)
	assert.Equal(t, hostedAgentSession, *got)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHostedAgents_CreateSessionFromManifest_Empty(t *testing.T) {
	setup()
	defer teardown()

	_, _, err := client.HostedAgents.CreateSessionFromManifest(ctx, []byte("  \n"))
	require.EqualError(t, err, "hosted agents: manifest is required")
}

func TestHostedAgents_ListSessions(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/sessions", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		assert.Equal(t, "SESSION_STATUS_READY", r.URL.Query().Get("status"))
		assert.Equal(t, "25", r.URL.Query().Get("page_size"))
		fmt.Fprintf(w, `{"sessions":[%s],"next_page_token":""}`, hostedAgentSessionJSON)
	})

	got, resp, err := client.HostedAgents.ListSessions(ctx, &HostedAgentSessionListOptions{
		Status:   HostedAgentSessionStatusReady,
		PageSize: 25,
	})
	require.NoError(t, err)
	require.Len(t, got.Sessions, 1)
	assert.Equal(t, hostedAgentSession, got.Sessions[0])
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHostedAgents_GetSession(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/sessions/sess-abc123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprintf(w, `{"session":%s}`, hostedAgentSessionJSON)
	})

	got, resp, err := client.HostedAgents.GetSession(ctx, "sess-abc123")
	require.NoError(t, err)
	assert.Equal(t, hostedAgentSession, *got)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHostedAgents_DestroySession(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/sessions/sess-abc123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		w.WriteHeader(http.StatusNoContent)
	})

	resp, err := client.HostedAgents.DestroySession(ctx, "sess-abc123")
	require.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestHostedAgents_PauseSession(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/sessions/sess-abc123/pause", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		assert.JSONEq(t, "{}", string(body))
		w.WriteHeader(http.StatusNoContent)
	})

	resp, err := client.HostedAgents.PauseSession(ctx, "sess-abc123")
	require.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestHostedAgents_ResumeSession(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/sessions/sess-abc123/resume", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		assert.JSONEq(t, "{}", string(body))
		w.WriteHeader(http.StatusNoContent)
	})

	resp, err := client.HostedAgents.ResumeSession(ctx, "sess-abc123")
	require.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestHostedAgents_SendInput(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/sessions/sess-abc123/input", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body HostedAgentSendInputRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		assert.Equal(t, "fix the failing test", body.Text)
		fmt.Fprint(w, `{"run_id":"run-001"}`)
	})

	got, resp, err := client.HostedAgents.SendInput(ctx, "sess-abc123", &HostedAgentSendInputRequest{
		Text: "fix the failing test",
	})
	require.NoError(t, err)
	assert.Equal(t, "run-001", got.RunID)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHostedAgents_ResolveHITL(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/sessions/sess-abc123/hitl/req-001", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body HostedAgentResolveHITLRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		assert.Equal(t, HostedAgentHITLOutcomeApprove, body.Outcome)
		w.WriteHeader(http.StatusNoContent)
	})

	resp, err := client.HostedAgents.ResolveHITL(ctx, "sess-abc123", "req-001", &HostedAgentResolveHITLRequest{
		Outcome: HostedAgentHITLOutcomeApprove,
	})
	require.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestHostedAgents_StreamSession(t *testing.T) {
	setup()
	defer teardown()

	// The server serializes the SPI canonical event envelope: the discriminator
	// is `type` (dot-separated), the body is `data`, the timestamp is
	// `timestamp`, and the team id rides as a decimal string in `tenant_id`.
	const eventJSON = `{"event_id":"ev-1","run_id":"run-1","tenant_id":"42","session_id":"sess-abc123","timestamp":"2026-03-01T12:01:00Z","seq":1,"type":"run.token_delta","data":{"text":"hello"}}`

	mux.HandleFunc("/v2/agents/sessions/sess-abc123/stream", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		assert.Equal(t, "text/event-stream", r.Header.Get("Accept"))
		assert.Equal(t, "ev-0", r.URL.Query().Get("replay_from"))
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprintf(w, ": connected\n\n")
		fmt.Fprintf(w, "id: ev-1\nevent: run.token_delta\ndata: %s\n\n", eventJSON)
	})

	stream, resp, err := client.HostedAgents.StreamSession(ctx, "sess-abc123", &HostedAgentSessionStreamOptions{
		ReplayFrom: "ev-0",
	})
	require.NoError(t, err)
	require.NotNil(t, stream)
	defer stream.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	require.True(t, stream.Next())
	ev := stream.Current()
	assert.Equal(t, HostedAgentEventKindTokenChunk, ev.Kind)
	assert.Equal(t, "ev-1", ev.EventID)
	assert.Equal(t, "run-1", ev.RunID)
	assert.Equal(t, uint64(42), ev.TeamID)
	assert.JSONEq(t, `{"text":"hello"}`, string(ev.Payload))
	assert.NoError(t, stream.Err())
	assert.False(t, stream.Next())
}

// TestHostedAgentEvent_UnmarshalSPIWire pins the SPI canonical envelope decode:
// type->Kind (dot-separated), data->Payload, timestamp->At, tenant_id(string)->TeamID.
func TestHostedAgentEvent_UnmarshalSPIWire(t *testing.T) {
	const frame = `{"event_id":"ev-9","run_id":"run-7","tenant_id":"120","session_id":"sess-1","timestamp":"2026-06-05T12:56:24.774753219Z","seq":3,"type":"run.token_delta","data":{"text":"Paris"}}`

	var ev HostedAgentEvent
	require.NoError(t, json.Unmarshal([]byte(frame), &ev))

	assert.Equal(t, "ev-9", ev.EventID)
	assert.Equal(t, "run-7", ev.RunID)
	assert.Equal(t, "sess-1", ev.SessionID)
	assert.Equal(t, uint64(120), ev.TeamID)
	assert.Equal(t, uint64(3), ev.Seq)
	assert.Equal(t, HostedAgentEventKindTokenChunk, ev.Kind)
	assert.False(t, ev.At.IsZero())
	assert.JSONEq(t, `{"text":"Paris"}`, string(ev.Payload))
}

func TestHostedAgents_ListSessions_PageToken(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/sessions", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		assert.Equal(t, "cursor-abc", r.URL.Query().Get("page_token"))
		fmt.Fprintf(w, `{"sessions":[],"next_page_token":"cursor-def"}`)
	})

	got, resp, err := client.HostedAgents.ListSessions(ctx, &HostedAgentSessionListOptions{
		PageToken: "cursor-abc",
	})
	require.NoError(t, err)
	assert.Empty(t, got.Sessions)
	assert.Equal(t, "cursor-def", got.NextPageToken)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHostedAgents_ListSessions_ByName(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/sessions", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		assert.Equal(t, "godo-fixture", r.URL.Query().Get("name"))
		fmt.Fprintf(w, `{"sessions":[%s]}`, hostedAgentSessionJSON)
	})

	got, resp, err := client.HostedAgents.ListSessions(ctx, &HostedAgentSessionListOptions{
		Name: "godo-fixture",
	})
	require.NoError(t, err)
	require.Len(t, got.Sessions, 1)
	assert.Equal(t, "godo-fixture", got.Sessions[0].Name)
	assert.Equal(t, hostedAgentSession, got.Sessions[0])
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHostedAgents_ExecInSandbox(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/sessions/sess-abc123/sandbox/exec", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body HostedAgentSandboxExecRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		assert.Equal(t, []string{"echo", "hello"}, body.Argv)
		fmt.Fprint(w, `{"exit_code":0,"stdout":"hello\n"}`)
	})

	got, resp, err := client.HostedAgents.ExecInSandbox(ctx, "sess-abc123", &HostedAgentSandboxExecRequest{
		Argv: []string{"echo", "hello"},
	})
	require.NoError(t, err)
	assert.Equal(t, 0, got.ExitCode)
	assert.Equal(t, "hello\n", got.Stdout)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHostedAgents_StreamSession_ReplayOnly(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/sessions/sess-abc123/stream", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		assert.Equal(t, "true", r.URL.Query().Get("replay_only"))
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, ": replay only\n\n")
	})

	stream, resp, err := client.HostedAgents.StreamSession(ctx, "sess-abc123", &HostedAgentSessionStreamOptions{
		ReplayOnly: true,
	})
	require.NoError(t, err)
	defer stream.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.False(t, stream.Next())
}

func TestHostedAgents_ValidationErrors(t *testing.T) {
	setup()
	defer teardown()

	_, _, err := client.HostedAgents.CreateSession(ctx, nil)
	require.EqualError(t, err, "hosted agents: create request is required")

	_, _, err = client.HostedAgents.CreateSession(ctx, &HostedAgentSessionCreateRequest{})
	require.EqualError(t, err, "hosted agents: agent_kind is required")

	_, _, err = client.HostedAgents.GetSession(ctx, "")
	require.EqualError(t, err, "hosted agents: session id is required")

	_, err = client.HostedAgents.DestroySession(ctx, "")
	require.EqualError(t, err, "hosted agents: session id is required")

	_, err = client.HostedAgents.PauseSession(ctx, "")
	require.EqualError(t, err, "hosted agents: session id is required")

	_, err = client.HostedAgents.ResumeSession(ctx, "")
	require.EqualError(t, err, "hosted agents: session id is required")

	_, _, err = client.HostedAgents.SendInput(ctx, "sess-abc123", nil)
	require.EqualError(t, err, "hosted agents: input is required")

	_, _, err = client.HostedAgents.SendInput(ctx, "sess-abc123", &HostedAgentSendInputRequest{})
	require.EqualError(t, err, "hosted agents: text is required")

	_, err = client.HostedAgents.ResolveHITL(ctx, "", "req-001", &HostedAgentResolveHITLRequest{
		Outcome: HostedAgentHITLOutcomeApprove,
	})
	require.EqualError(t, err, "hosted agents: session id is required")

	_, err = client.HostedAgents.ResolveHITL(ctx, "sess-abc123", "req-001", &HostedAgentResolveHITLRequest{})
	require.EqualError(t, err, "hosted agents: outcome is required")

	_, _, err = client.HostedAgents.ExecInSandbox(ctx, "sess-abc123", &HostedAgentSandboxExecRequest{})
	require.EqualError(t, err, "hosted agents: argv is required")
}

func TestHostedAgents_UploadWorkspace(t *testing.T) {
	setup()
	defer teardown()

	const payload = "hello world"

	mux.HandleFunc("/v2/agents/sessions/sess-abc123/workspace/upload", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		assert.Equal(t, "application/octet-stream", r.Header.Get("Content-Type"))
		assert.Equal(t, "src/main.go", r.URL.Query().Get("path"))
		assert.Equal(t, "true", r.URL.Query().Get("is_archive"))
		assert.Equal(t, "deadbeef", r.Header.Get("X-Content-Sha256"))
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		assert.Equal(t, payload, string(body))
		fmt.Fprintf(w, `{"path":"/workspace/src/main.go","bytes_written":%d}`, len(payload))
	})

	got, resp, err := client.HostedAgents.UploadWorkspace(ctx, "sess-abc123", &HostedAgentWorkspaceUploadRequest{
		Path:          "src/main.go",
		IsArchive:     true,
		ContentSHA256: "deadbeef",
		Body:          strings.NewReader(payload),
	})
	require.NoError(t, err)
	assert.Equal(t, "/workspace/src/main.go", got.Path)
	assert.Equal(t, int64(len(payload)), got.BytesWritten)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHostedAgents_DownloadWorkspace(t *testing.T) {
	setup()
	defer teardown()

	const payload = "the quick brown fox"
	sum := sha256.Sum256([]byte(payload))
	digest := hex.EncodeToString(sum[:])

	mux.HandleFunc("/v2/agents/sessions/sess-abc123/workspace/download", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		assert.Equal(t, "notes.txt", r.URL.Query().Get("path"))
		assert.Equal(t, "true", r.URL.Query().Get("as_archive"))

		w.Header().Set("Trailer", "X-Content-Sha256")
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("X-Workspace-Is-Archive", "true")
		w.Header().Set("X-Workspace-Size-Bytes", strconv.Itoa(len(payload)))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(payload))
		w.Header().Set("X-Content-Sha256", digest)
	})

	dl, resp, err := client.HostedAgents.DownloadWorkspace(ctx, "sess-abc123", &HostedAgentWorkspaceDownloadRequest{
		Path:      "notes.txt",
		AsArchive: true,
	})
	require.NoError(t, err)
	require.NotNil(t, dl)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.True(t, dl.IsArchive)
	assert.Equal(t, int64(len(payload)), dl.SizeBytes)

	body, err := io.ReadAll(dl.Body)
	require.NoError(t, err)
	assert.Equal(t, payload, string(body))
	require.NoError(t, dl.Body.Close())
}

func TestHostedAgents_DownloadWorkspace_ChecksumMismatch(t *testing.T) {
	setup()
	defer teardown()

	const payload = "the quick brown fox"

	mux.HandleFunc("/v2/agents/sessions/sess-abc123/workspace/download", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Trailer", "X-Content-Sha256")
		w.Header().Set("Content-Type", "application/octet-stream")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(payload))
		w.Header().Set("X-Content-Sha256", "0000000000000000000000000000000000000000000000000000000000000000")
	})

	dl, _, err := client.HostedAgents.DownloadWorkspace(ctx, "sess-abc123", &HostedAgentWorkspaceDownloadRequest{
		Path: "notes.txt",
	})
	require.NoError(t, err)
	defer dl.Body.Close()

	_, err = io.ReadAll(dl.Body)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "checksum mismatch")
}

func TestHostedAgents_DownloadWorkspace_MissingTrailer(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/sessions/sess-abc123/workspace/download", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("partial data"))
	})

	dl, _, err := client.HostedAgents.DownloadWorkspace(ctx, "sess-abc123", &HostedAgentWorkspaceDownloadRequest{
		Path: "notes.txt",
	})
	require.NoError(t, err)
	defer dl.Body.Close()

	_, err = io.ReadAll(dl.Body)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing X-Content-Sha256")
}

func TestHostedAgents_WorkspaceValidationErrors(t *testing.T) {
	setup()
	defer teardown()

	_, _, err := client.HostedAgents.UploadWorkspace(ctx, "", &HostedAgentWorkspaceUploadRequest{})
	require.EqualError(t, err, "hosted agents: session id is required")

	_, _, err = client.HostedAgents.UploadWorkspace(ctx, "sess-abc123", nil)
	require.EqualError(t, err, "hosted agents: upload request is required")

	_, _, err = client.HostedAgents.UploadWorkspace(ctx, "sess-abc123", &HostedAgentWorkspaceUploadRequest{})
	require.EqualError(t, err, "hosted agents: path is required")

	_, _, err = client.HostedAgents.UploadWorkspace(ctx, "sess-abc123", &HostedAgentWorkspaceUploadRequest{Path: "x"})
	require.EqualError(t, err, "hosted agents: body is required")

	_, _, err = client.HostedAgents.DownloadWorkspace(ctx, "", &HostedAgentWorkspaceDownloadRequest{})
	require.EqualError(t, err, "hosted agents: session id is required")

	_, _, err = client.HostedAgents.DownloadWorkspace(ctx, "sess-abc123", nil)
	require.EqualError(t, err, "hosted agents: download request is required")

	_, _, err = client.HostedAgents.DownloadWorkspace(ctx, "sess-abc123", &HostedAgentWorkspaceDownloadRequest{})
	require.EqualError(t, err, "hosted agents: path is required")
}

func TestHostedAgents_GetSession_EmptyBody(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/sessions/sess-abc123", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{}`)
	})

	_, resp, err := client.HostedAgents.GetSession(ctx, "sess-abc123")
	require.EqualError(t, err, "hosted agents: get session returned no session")
	require.NotNil(t, resp)
}
