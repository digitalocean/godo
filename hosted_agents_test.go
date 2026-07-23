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

func workspaceDownloadFooter(payload string) string {
	sum := sha256.Sum256([]byte(payload))
	return workspaceDownloadFooterPrefix + hex.EncodeToString(sum[:]) + "\n"
}

func TestHostedAgents_DownloadWorkspace(t *testing.T) {
	setup()
	defer teardown()

	const payload = "the quick brown fox"
	footer := workspaceDownloadFooter(payload)

	mux.HandleFunc("/v2/agents/sessions/sess-abc123/workspace/download", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		assert.Equal(t, "notes.txt", r.URL.Query().Get("path"))
		assert.Equal(t, "true", r.URL.Query().Get("as_archive"))

		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("X-Workspace-Is-Archive", "true")
		w.Header().Set("X-Workspace-Size-Bytes", strconv.Itoa(len(payload)))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(payload + footer))
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
	badFooter := workspaceDownloadFooterPrefix + strings.Repeat("0", 64) + "\n"

	mux.HandleFunc("/v2/agents/sessions/sess-abc123/workspace/download", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(payload + badFooter))
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

func TestHostedAgents_DownloadWorkspace_MissingFooter(t *testing.T) {
	setup()
	defer teardown()

	const payload = "partial data"

	mux.HandleFunc("/v2/agents/sessions/sess-abc123/workspace/download", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(payload))
	})

	dl, _, err := client.HostedAgents.DownloadWorkspace(ctx, "sess-abc123", &HostedAgentWorkspaceDownloadRequest{
		Path: "notes.txt",
	})
	require.NoError(t, err)
	defer dl.Body.Close()

	_, err = io.ReadAll(dl.Body)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "integrity footer")
}

func TestHostedAgents_DownloadWorkspace_InvalidFooter(t *testing.T) {
	setup()
	defer teardown()

	const payload = "the quick brown fox"
	invalidFooter := "NOTASHA1" + strings.Repeat("a", 64) + "\n"

	mux.HandleFunc("/v2/agents/sessions/sess-abc123/workspace/download", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(payload + invalidFooter))
	})

	dl, _, err := client.HostedAgents.DownloadWorkspace(ctx, "sess-abc123", &HostedAgentWorkspaceDownloadRequest{
		Path: "notes.txt",
	})
	require.NoError(t, err)
	defer dl.Body.Close()

	_, err = io.ReadAll(dl.Body)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid workspace download integrity footer")
}

func TestHostedAgents_DownloadWorkspace_EmptyPayload(t *testing.T) {
	setup()
	defer teardown()

	const payload = ""
	footer := workspaceDownloadFooter(payload)

	mux.HandleFunc("/v2/agents/sessions/sess-abc123/workspace/download", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("X-Workspace-Size-Bytes", "0")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(footer))
	})

	dl, _, err := client.HostedAgents.DownloadWorkspace(ctx, "sess-abc123", &HostedAgentWorkspaceDownloadRequest{
		Path: "empty.txt",
	})
	require.NoError(t, err)
	defer dl.Body.Close()

	body, err := io.ReadAll(dl.Body)
	require.NoError(t, err)
	assert.Equal(t, "", string(body))
	assert.Equal(t, int64(0), dl.SizeBytes)
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

func TestHostedAgents_CreateWorkspaceTransfer_Upload(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/sessions/sess-abc123/workspace/transfers", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var got HostedAgentWorkspaceTransferCreateRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&got))
		assert.Equal(t, HostedAgentWorkspaceTransferDirectionUpload, got.Direction)
		assert.Equal(t, "/workspace/data/big.bin", got.Path)
		assert.Equal(t, int64(524288000), got.SizeBytes)
		assert.Equal(t, "abc123", got.SHA256)
		assert.False(t, got.IsArchive)

		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, `{
			"transfer_id": "xfer-upload-1",
			"direction": "upload",
			"status": "pending",
			"upload_id": "up-1",
			"part_size": 16777216,
			"expires_at": "2026-07-21T12:00:00Z"
		}`)
	})

	got, resp, err := client.HostedAgents.CreateWorkspaceTransfer(ctx, "sess-abc123", &HostedAgentWorkspaceTransferCreateRequest{
		Direction: HostedAgentWorkspaceTransferDirectionUpload,
		Path:      "/workspace/data/big.bin",
		SizeBytes: 524288000,
		SHA256:    "abc123",
	})
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	assert.Equal(t, "xfer-upload-1", got.TransferID)
	assert.Equal(t, HostedAgentWorkspaceTransferDirectionUpload, got.Direction)
	assert.Equal(t, HostedAgentWorkspaceTransferStatusPending, got.Status)
	assert.Equal(t, "up-1", got.UploadID)
	assert.Equal(t, int64(16777216), got.PartSize)
	require.NotNil(t, got.ExpiresAt)
	assert.Equal(t, time.Date(2026, 7, 21, 12, 0, 0, 0, time.UTC), got.ExpiresAt.Time)
}

func TestHostedAgents_CreateWorkspaceTransfer_Download(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/sessions/sess-abc123/workspace/transfers", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var got HostedAgentWorkspaceTransferCreateRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&got))
		assert.Equal(t, HostedAgentWorkspaceTransferDirectionDownload, got.Direction)
		assert.Equal(t, "/workspace/data/big.bin", got.Path)
		assert.True(t, got.AsArchive)

		w.WriteHeader(http.StatusAccepted)
		fmt.Fprint(w, `{
			"transfer_id": "xfer-download-1",
			"direction": "download",
			"status": "pending"
		}`)
	})

	got, resp, err := client.HostedAgents.CreateWorkspaceTransfer(ctx, "sess-abc123", &HostedAgentWorkspaceTransferCreateRequest{
		Direction: HostedAgentWorkspaceTransferDirectionDownload,
		Path:      "/workspace/data/big.bin",
		AsArchive: true,
	})
	require.NoError(t, err)
	assert.Equal(t, http.StatusAccepted, resp.StatusCode)
	assert.Equal(t, "xfer-download-1", got.TransferID)
	assert.Equal(t, HostedAgentWorkspaceTransferStatusPending, got.Status)
}

func TestHostedAgents_CreateWorkspaceTransferPartUploadURL(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/sessions/sess-abc123/workspace/transfers/xfer-1/part-upload-urls", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var got HostedAgentWorkspaceTransferPartUploadURLRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&got))
		assert.Equal(t, 1, got.PartNumber)

		fmt.Fprint(w, `{
			"transfer_id": "xfer-1",
			"part_number": 1,
			"upload_url": "https://spaces.example/part-1",
			"expires_at": "2026-07-21T12:30:00Z"
		}`)
	})

	got, resp, err := client.HostedAgents.CreateWorkspaceTransferPartUploadURL(ctx, "sess-abc123", "xfer-1", &HostedAgentWorkspaceTransferPartUploadURLRequest{
		PartNumber: 1,
	})
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "xfer-1", got.TransferID)
	assert.Equal(t, 1, got.PartNumber)
	assert.Equal(t, "https://spaces.example/part-1", got.UploadURL)
	require.NotNil(t, got.ExpiresAt)
	assert.Equal(t, time.Date(2026, 7, 21, 12, 30, 0, 0, time.UTC), got.ExpiresAt.Time)
}

func TestHostedAgents_CommitWorkspaceTransfer(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/sessions/sess-abc123/workspace/transfers/xfer-1/commit", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var got HostedAgentWorkspaceTransferCommitRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&got))
		assert.Equal(t, "deadbeef", got.SHA256)

		w.WriteHeader(http.StatusAccepted)
		fmt.Fprint(w, `{
			"transfer_id": "xfer-1",
			"status": "in_progress",
			"size_bytes": 524288000
		}`)
	})

	got, resp, err := client.HostedAgents.CommitWorkspaceTransfer(ctx, "sess-abc123", "xfer-1", &HostedAgentWorkspaceTransferCommitRequest{
		SHA256: "deadbeef",
	})
	require.NoError(t, err)
	assert.Equal(t, http.StatusAccepted, resp.StatusCode)
	assert.Equal(t, HostedAgentWorkspaceTransferStatusInProgress, got.Status)
	assert.Equal(t, int64(524288000), got.SizeBytes)
}

func TestHostedAgents_GetWorkspaceTransfer(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/sessions/sess-abc123/workspace/transfers/xfer-1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"transfer_id": "xfer-1",
			"direction": "download",
			"status": "completed",
			"bytes_written": 524288000,
			"sha256": "e6b84c4839cbbfc3cde3b1bc84e8a82b9661c00eae1726fcff0dca8d643423ae",
			"download_url": "https://spaces.example/download",
			"expires_at": "2026-07-21T13:00:00Z",
			"error_message": null
		}`)
	})

	got, resp, err := client.HostedAgents.GetWorkspaceTransfer(ctx, "sess-abc123", "xfer-1")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, HostedAgentWorkspaceTransferDirectionDownload, got.Direction)
	assert.Equal(t, HostedAgentWorkspaceTransferStatusCompleted, got.Status)
	assert.Equal(t, int64(524288000), got.BytesWritten)
	assert.Equal(t, "e6b84c4839cbbfc3cde3b1bc84e8a82b9661c00eae1726fcff0dca8d643423ae", got.SHA256)
	assert.Equal(t, "https://spaces.example/download", got.DownloadURL)
	require.NotNil(t, got.ExpiresAt)
	assert.Equal(t, time.Date(2026, 7, 21, 13, 0, 0, 0, time.UTC), got.ExpiresAt.Time)
	assert.Empty(t, got.ErrorMessage)
}

func TestHostedAgents_CancelWorkspaceTransfer(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/sessions/sess-abc123/workspace/transfers/xfer-1/cancel", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var got HostedAgentWorkspaceTransferCancelRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&got))
		assert.Equal(t, "user cancelled", got.Reason)

		fmt.Fprint(w, `{
			"transfer_id": "xfer-1",
			"aborted": true,
			"status": "failed"
		}`)
	})

	got, resp, err := client.HostedAgents.CancelWorkspaceTransfer(ctx, "sess-abc123", "xfer-1", &HostedAgentWorkspaceTransferCancelRequest{
		Reason: "user cancelled",
	})
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "xfer-1", got.TransferID)
	assert.True(t, got.Aborted)
	assert.Equal(t, HostedAgentWorkspaceTransferStatusFailed, got.Status)
}

func TestHostedAgents_WorkspaceTransferValidationErrors(t *testing.T) {
	setup()
	defer teardown()

	_, _, err := client.HostedAgents.CreateWorkspaceTransfer(ctx, "", &HostedAgentWorkspaceTransferCreateRequest{})
	require.EqualError(t, err, "hosted agents: session id is required")

	_, _, err = client.HostedAgents.CreateWorkspaceTransfer(ctx, "sess-abc123", nil)
	require.EqualError(t, err, "hosted agents: transfer create request is required")

	_, _, err = client.HostedAgents.CreateWorkspaceTransfer(ctx, "sess-abc123", &HostedAgentWorkspaceTransferCreateRequest{})
	require.EqualError(t, err, "hosted agents: direction is required")

	_, _, err = client.HostedAgents.CreateWorkspaceTransfer(ctx, "sess-abc123", &HostedAgentWorkspaceTransferCreateRequest{
		Direction: HostedAgentWorkspaceTransferDirectionUpload,
	})
	require.EqualError(t, err, "hosted agents: path is required")

	_, _, err = client.HostedAgents.CreateWorkspaceTransferPartUploadURL(ctx, "sess-abc123", "xfer-1", &HostedAgentWorkspaceTransferPartUploadURLRequest{})
	require.EqualError(t, err, "hosted agents: part_number must be >= 1")

	_, _, err = client.HostedAgents.GetWorkspaceTransfer(ctx, "sess-abc123", "")
	require.EqualError(t, err, "hosted agents: transfer id is required")

	_, _, err = client.HostedAgents.CancelWorkspaceTransfer(ctx, "", "xfer-1", nil)
	require.EqualError(t, err, "hosted agents: session id is required")
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
