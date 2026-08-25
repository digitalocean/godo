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

var hostedAgentTrigger = HostedAgentTrigger{
	TriggerID:      "trig-abc123",
	TeamID:         42,
	Kind:           HostedAgentTriggerKindWebhook,
	Name:           "github-prs",
	Status:         HostedAgentTriggerStatusActive,
	SessionMode:    HostedAgentTriggerSessionModeFresh,
	AgentKind:      HostedAgentKindOpenCode,
	PromptTemplate: "Review {{payload.pull_request.title}}",
	Output: &HostedAgentTriggerOutputRead{
		Mode:            HostedAgentTriggerOutputModeEmail,
		EmailConfigured: true,
		SlackConfigured: false,
	},
	SessionTemplate: "apiVersion: agents.digitalocean.com/v1alpha1\nkind: Agent\n",
	Webhook: &HostedAgentWebhookConfig{
		Provider:   HostedAgentWebhookProviderGitHub,
		WebhookURL: "https://api.digitalocean.com/v2/agents/triggers/trig-abc123/webhook",
	},
	CreatedAt: Timestamp{Time: time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)},
	UpdatedAt: Timestamp{Time: time.Date(2026, 7, 1, 12, 5, 0, 0, time.UTC)},
}

var hostedAgentTriggerJSON = `
{
	"trigger_id": "trig-abc123",
	"team_id": 42,
	"kind": "webhook",
	"name": "github-prs",
	"status": "active",
	"session_mode": "fresh",
	"agent_kind": "AGENT_KIND_OPENCODE",
	"prompt_template": "Review {{payload.pull_request.title}}",
	"output": {
		"mode": "email",
		"email_configured": true,
		"slack_configured": false
	},
	"session_template": "apiVersion: agents.digitalocean.com/v1alpha1\nkind: Agent\n",
	"webhook": {
		"provider": "github",
		"webhook_url": "https://api.digitalocean.com/v2/agents/triggers/trig-abc123/webhook"
	},
	"created_at": "2026-07-01T12:00:00Z",
	"updated_at": "2026-07-01T12:05:00Z"
}
`

var hostedAgentCronTrigger = HostedAgentTrigger{
	TriggerID:      "trig-cron1",
	TeamID:         42,
	Kind:           HostedAgentTriggerKindCron,
	Name:           "nightly-summary",
	Status:         HostedAgentTriggerStatusActive,
	SessionMode:    HostedAgentTriggerSessionModeReuse,
	AgentKind:      HostedAgentKindClaudeCode,
	PromptTemplate: "Summarize overnight activity",
	Output: &HostedAgentTriggerOutputRead{
		Mode: HostedAgentTriggerOutputModeNone,
	},
	BoundSessionID: "sess-paused-1",
	Cron: &HostedAgentCronConfig{
		CronExpr:  "0 9 * * *",
		Timezone:  "America/New_York",
		NextRunAt: Timestamp{Time: time.Date(2026, 7, 2, 13, 0, 0, 0, time.UTC)},
	},
	CreatedAt: Timestamp{Time: time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)},
	UpdatedAt: Timestamp{Time: time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)},
}

var hostedAgentCronTriggerJSON = `
{
	"trigger_id": "trig-cron1",
	"team_id": 42,
	"kind": "cron",
	"name": "nightly-summary",
	"status": "active",
	"session_mode": "reuse",
	"agent_kind": "AGENT_KIND_CLAUDE_CODE",
	"prompt_template": "Summarize overnight activity",
	"output": {"mode": "none"},
	"bound_session_id": "sess-paused-1",
	"cron": {
		"cron_expr": "0 9 * * *",
		"timezone": "America/New_York",
		"next_run_at": "2026-07-02T13:00:00Z"
	},
	"created_at": "2026-07-01T12:00:00Z",
	"updated_at": "2026-07-01T12:00:00Z"
}
`

func TestHostedAgentTriggers_List(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/triggers", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		assert.Equal(t, "webhook", r.URL.Query().Get("kind"))
		assert.Equal(t, "active", r.URL.Query().Get("status"))
		assert.Equal(t, "25", r.URL.Query().Get("page_size"))
		fmt.Fprintf(w, `{"triggers":[%s],"next_page_token":"tok-2"}`, hostedAgentTriggerJSON)
	})

	got, resp, err := client.HostedAgentTriggers.List(ctx, &HostedAgentTriggerListOptions{
		PageSize: 25,
		Kind:     HostedAgentTriggerKindWebhook,
		Status:   HostedAgentTriggerStatusActive,
	})
	require.NoError(t, err)
	require.Len(t, got.Triggers, 1)
	assert.Equal(t, hostedAgentTrigger, got.Triggers[0])
	assert.Equal(t, "tok-2", got.NextPageToken)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHostedAgentTriggers_CreateWebhook(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/triggers", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body HostedAgentTriggerCreateRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		assert.Equal(t, HostedAgentTriggerKindWebhook, body.Kind)
		assert.Equal(t, "github-prs", body.Name)
		assert.Equal(t, HostedAgentTriggerSessionModeFresh, body.SessionMode)
		assert.Equal(t, HostedAgentWebhookProviderGitHub, body.Webhook.Provider)
		assert.Equal(t, HostedAgentTriggerOutputModeEmail, body.Output.Mode)
		assert.Equal(t, "ops@example.com", body.Output.Email)
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, `{"trigger":%s,"webhook_secret":"whsec_once"}`, hostedAgentTriggerJSON)
	})

	got, resp, err := client.HostedAgentTriggers.Create(ctx, &HostedAgentTriggerCreateRequest{
		Kind:           HostedAgentTriggerKindWebhook,
		Name:           "github-prs",
		SessionMode:    HostedAgentTriggerSessionModeFresh,
		PromptTemplate: "Review {{payload.pull_request.title}}",
		Output: HostedAgentTriggerOutputWrite{
			Mode:  HostedAgentTriggerOutputModeEmail,
			Email: "ops@example.com",
		},
		SessionTemplate: "apiVersion: agents.digitalocean.com/v1alpha1\nkind: Agent\n",
		Webhook: &HostedAgentCreateWebhookConfig{
			Provider: HostedAgentWebhookProviderGitHub,
		},
	})
	require.NoError(t, err)
	require.NotNil(t, got.Trigger)
	assert.Equal(t, hostedAgentTrigger, *got.Trigger)
	assert.Equal(t, "whsec_once", got.WebhookSecret)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
}

func TestHostedAgentTriggers_CreateWithSlackOutput(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/triggers", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body HostedAgentTriggerCreateRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		assert.Equal(t, HostedAgentTriggerOutputModeSlack, body.Output.Mode)
		require.NotNil(t, body.Output.Slack)
		assert.Equal(t, "https://hooks.slack.com/services/T/B/xxx", body.Output.Slack.WebhookURL)
		assert.Empty(t, body.Output.Email)

		raw, err := json.Marshal(body.Output)
		require.NoError(t, err)
		assert.Contains(t, string(raw), `"slack"`)
		assert.Contains(t, string(raw), `"webhook_url"`)
		assert.NotContains(t, string(raw), "slack_webhook_url")

		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, `{
			"trigger": {
				"trigger_id": "trig-slack",
				"kind": "cron",
				"name": "slack-notify",
				"status": "active",
				"session_mode": "fresh",
				"agent_kind": "AGENT_KIND_OPENCODE",
				"prompt_template": "Ping Slack",
				"output": {
					"mode": "slack",
					"email_configured": false,
					"slack_configured": true
				},
				"session_template": "apiVersion: agents.digitalocean.com/v1alpha1\nkind: Agent\n",
				"cron": {"cron_expr": "0 * * * *", "timezone": "UTC"},
				"created_at": "2026-07-01T12:00:00Z",
				"updated_at": "2026-07-01T12:00:00Z"
			}
		}`)
	})

	got, resp, err := client.HostedAgentTriggers.Create(ctx, &HostedAgentTriggerCreateRequest{
		Kind:           HostedAgentTriggerKindCron,
		Name:           "slack-notify",
		SessionMode:    HostedAgentTriggerSessionModeFresh,
		PromptTemplate: "Ping Slack",
		Output: HostedAgentTriggerOutputWrite{
			Mode: HostedAgentTriggerOutputModeSlack,
			Slack: &HostedAgentTriggerSlackOutputWrite{
				WebhookURL: "https://hooks.slack.com/services/T/B/xxx",
			},
		},
		SessionTemplate: "apiVersion: agents.digitalocean.com/v1alpha1\nkind: Agent\n",
		Cron: &HostedAgentCreateCronConfig{
			CronExpr: "0 * * * *",
			Timezone: "UTC",
		},
	})
	require.NoError(t, err)
	require.NotNil(t, got.Trigger)
	require.NotNil(t, got.Trigger.Output)
	assert.Equal(t, HostedAgentTriggerOutputModeSlack, got.Trigger.Output.Mode)
	assert.True(t, got.Trigger.Output.SlackConfigured)
	assert.False(t, got.Trigger.Output.EmailConfigured)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
}

func TestHostedAgentTriggers_CreateCron(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/triggers", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body HostedAgentTriggerCreateRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		assert.Equal(t, HostedAgentTriggerKindCron, body.Kind)
		assert.Equal(t, "0 9 * * *", body.Cron.CronExpr)
		assert.Equal(t, "America/New_York", body.Cron.Timezone)
		assert.Nil(t, body.Webhook)
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, `{"trigger":%s}`, hostedAgentCronTriggerJSON)
	})

	got, resp, err := client.HostedAgentTriggers.Create(ctx, &HostedAgentTriggerCreateRequest{
		Kind:           HostedAgentTriggerKindCron,
		Name:           "nightly-summary",
		SessionMode:    HostedAgentTriggerSessionModeReuse,
		PromptTemplate: "Summarize overnight activity",
		Output:         HostedAgentTriggerOutputWrite{Mode: HostedAgentTriggerOutputModeNone},
		BoundSessionID: "sess-paused-1",
		Cron: &HostedAgentCreateCronConfig{
			CronExpr: "0 9 * * *",
			Timezone: "America/New_York",
		},
	})
	require.NoError(t, err)
	require.NotNil(t, got.Trigger)
	assert.Equal(t, hostedAgentCronTrigger, *got.Trigger)
	assert.Empty(t, got.WebhookSecret)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
}

func TestHostedAgentTriggers_Create_Nil(t *testing.T) {
	setup()
	defer teardown()

	_, _, err := client.HostedAgentTriggers.Create(ctx, nil)
	require.EqualError(t, err, "create is invalid because cannot be nil")
}

func TestHostedAgentTriggers_Get(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/triggers/trig-abc123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprintf(w, `{"trigger":%s}`, hostedAgentTriggerJSON)
	})

	got, resp, err := client.HostedAgentTriggers.Get(ctx, "trig-abc123")
	require.NoError(t, err)
	assert.Equal(t, hostedAgentTrigger, *got)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHostedAgentTriggers_Get_EmptyID(t *testing.T) {
	setup()
	defer teardown()

	_, _, err := client.HostedAgentTriggers.Get(ctx, "")
	require.EqualError(t, err, "hosted agent triggers: trigger id is required")
}

func TestHostedAgentTriggers_Update(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/triggers/trig-abc123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPatch)
		var body HostedAgentTriggerUpdateRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		assert.Equal(t, HostedAgentTriggerStatusPaused, body.Status)
		paused := hostedAgentTrigger
		paused.Status = HostedAgentTriggerStatusPaused
		b, err := json.Marshal(map[string]any{"trigger": paused})
		require.NoError(t, err)
		w.Write(b)
	})

	got, resp, err := client.HostedAgentTriggers.Update(ctx, "trig-abc123", &HostedAgentTriggerUpdateRequest{
		Status: HostedAgentTriggerStatusPaused,
	})
	require.NoError(t, err)
	assert.Equal(t, HostedAgentTriggerStatusPaused, got.Status)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHostedAgentTriggers_Delete(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/triggers/trig-abc123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		w.WriteHeader(http.StatusNoContent)
	})

	resp, err := client.HostedAgentTriggers.Delete(ctx, "trig-abc123")
	require.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestHostedAgentTriggers_RotateSecret(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/triggers/trig-abc123/rotate-secret", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		assert.Empty(t, r.URL.Query().Get("revoke_previous"), "the default rotate must not ask for revocation")
		fmt.Fprint(w, `{"webhook_secret":"whsec_rotated","previous_secret_expires_at":"2026-07-01T12:05:00Z"}`)
	})

	got, resp, err := client.HostedAgentTriggers.RotateSecret(ctx, "trig-abc123", nil)
	require.NoError(t, err)
	assert.Equal(t, "whsec_rotated", got.WebhookSecret)
	require.NotNil(t, got.PreviousSecretExpiresAt)
	assert.Equal(t, time.Date(2026, 7, 1, 12, 5, 0, 0, time.UTC), got.PreviousSecretExpiresAt.UTC())
	assert.False(t, got.PreviousSecretRevoked)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// A non-nil options struct with RevokePrevious unset must be indistinguishable
// on the wire from passing nil. This is what the `omitempty` tag buys, and
// without it a caller building options programmatically would send
// revoke_previous=false and depend on the server parsing it.
func TestHostedAgentTriggers_RotateSecret_ExplicitFalseOmitsParam(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/triggers/trig-abc123/rotate-secret", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		assert.Empty(t, r.URL.RawQuery, "an unset RevokePrevious must put nothing on the wire")
		fmt.Fprint(w, `{"webhook_secret":"whsec_rotated","previous_secret_expires_at":"2026-07-01T12:05:00Z"}`)
	})

	got, _, err := client.HostedAgentTriggers.RotateSecret(ctx, "trig-abc123", &HostedAgentTriggerRotateSecretOptions{})
	require.NoError(t, err)
	assert.Equal(t, "whsec_rotated", got.WebhookSecret)
	assert.False(t, got.PreviousSecretRevoked)
}

func TestHostedAgentTriggers_RotateSecret_RequiresTriggerID(t *testing.T) {
	setup()
	defer teardown()

	_, _, err := client.HostedAgentTriggers.RotateSecret(ctx, "", nil)
	require.Error(t, err, "an empty id would POST to the collection path instead of a trigger")
}

func TestHostedAgentTriggers_RotateSecret_RevokePrevious(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/triggers/trig-abc123/rotate-secret", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		assert.Equal(t, "true", r.URL.Query().Get("revoke_previous"))
		fmt.Fprint(w, `{"webhook_secret":"whsec_rotated","previous_secret_revoked":true}`)
	})

	got, resp, err := client.HostedAgentTriggers.RotateSecret(ctx, "trig-abc123", &HostedAgentTriggerRotateSecretOptions{RevokePrevious: true})
	require.NoError(t, err)
	assert.Equal(t, "whsec_rotated", got.WebhookSecret)
	assert.True(t, got.PreviousSecretRevoked)
	assert.Nil(t, got.PreviousSecretExpiresAt, "there is no expiry to report when the old secret is already dead")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHostedAgentTriggers_ListExecutions(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/triggers/trig-abc123/executions", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		assert.Equal(t, "failed", r.URL.Query().Get("status"))
		fmt.Fprint(w, `{
			"executions": [{
				"execution_id": "exec-1",
				"trigger_id": "trig-abc123",
				"status": "failed",
				"session_id": "sess-1",
				"failure_reason": "session gone — re-bind",
				"created_at": "2026-07-01T12:10:00Z",
				"updated_at": "2026-07-01T12:11:00Z"
			}],
			"next_page_token": ""
		}`)
	})

	got, resp, err := client.HostedAgentTriggers.ListExecutions(ctx, "trig-abc123", &HostedAgentTriggerExecutionListOptions{
		Status: HostedAgentTriggerExecutionStatusFailed,
	})
	require.NoError(t, err)
	require.Len(t, got.Executions, 1)
	assert.Equal(t, "exec-1", got.Executions[0].ExecutionID)
	assert.Equal(t, HostedAgentTriggerExecutionStatusFailed, got.Executions[0].Status)
	assert.Empty(t, got.Executions[0].Payload)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHostedAgentTriggers_GetExecution(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/triggers/trig-abc123/executions/exec-1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"execution": {
				"execution_id": "exec-1",
				"trigger_id": "trig-abc123",
				"status": "succeeded",
				"session_id": "sess-1",
				"run_id": "run-9",
				"created_at": "2026-07-01T12:10:00Z",
				"updated_at": "2026-07-01T12:12:00Z",
				"payload": "{\"action\":\"opened\"}",
				"output_text": "hello from run",
				"output_truncated": true
			}
		}`)
	})

	got, resp, err := client.HostedAgentTriggers.GetExecution(ctx, "trig-abc123", "exec-1")
	require.NoError(t, err)
	assert.Equal(t, "exec-1", got.ExecutionID)
	assert.Equal(t, HostedAgentTriggerExecutionStatusSucceeded, got.Status)
	assert.Equal(t, `{"action":"opened"}`, got.Payload)
	assert.Equal(t, "hello from run", got.OutputText)
	assert.True(t, got.OutputTruncated)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHostedAgentTriggers_GetBySession(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/triggers/by-session/sess-paused-1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprintf(w, `{"trigger":%s}`, hostedAgentCronTriggerJSON)
	})

	got, resp, err := client.HostedAgentTriggers.GetBySession(ctx, "sess-paused-1")
	require.NoError(t, err)
	assert.Equal(t, hostedAgentCronTrigger, *got)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHostedAgentTriggers_ListReusableSessions(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/triggers/reusable-sessions", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		assert.Equal(t, "10", r.URL.Query().Get("page_size"))
		fmt.Fprint(w, `{
			"sessions": [{
				"session_id": "sess-paused-1",
				"name": "reuse-me",
				"agent_kind": "AGENT_KIND_CLAUDE_CODE",
				"status": "SESSION_STATUS_PAUSED",
				"created_at": "2026-07-01T11:00:00Z",
				"last_event_at": "2026-07-01T11:30:00Z"
			}],
			"next_page_token": ""
		}`)
	})

	got, resp, err := client.HostedAgentTriggers.ListReusableSessions(ctx, &HostedAgentReusableSessionListOptions{
		PageSize: 10,
	})
	require.NoError(t, err)
	require.Len(t, got.Sessions, 1)
	assert.Equal(t, "sess-paused-1", got.Sessions[0].SessionID)
	assert.Equal(t, HostedAgentSessionStatusPaused, got.Sessions[0].Status)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHostedAgentTriggers_ListWebhookProviders(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/webhook-providers", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"providers": [{
				"key": "github",
				"display_name": "GitHub",
				"description": "GitHub webhooks",
				"docs_url": "https://docs.github.com/webhooks",
				"paste_hint": "Paste into Secret",
				"signature": {
					"header": "X-Hub-Signature-256",
					"scheme": "hmac-sha256"
				}
			}]
		}`)
	})

	got, resp, err := client.HostedAgentTriggers.ListWebhookProviders(ctx)
	require.NoError(t, err)
	require.Len(t, got.Providers, 1)
	assert.Equal(t, HostedAgentWebhookProviderGitHub, got.Providers[0].Key)
	assert.Equal(t, "X-Hub-Signature-256", got.Providers[0].Signature.Header)
	assert.Equal(t, HostedAgentWebhookSignatureHMACSHA256, got.Providers[0].Signature.Scheme)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
