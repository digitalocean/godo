package godo

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const hostedAgentPolicyValidateManifest = `apiVersion: agents.digitalocean.com/v1alpha1
kind: Agent
metadata:
  name: policy-check
spec:
  runtime:
    adapter: codex
  sandbox:
    template: coding-codex
  permissions:
    defaultAction: ask
    rules:
      - tool: bash
        action: ask
`

const hostedAgentPolicyValidationResultJSON = `{
	"agent": "codex",
	"version": "0.38",
	"ok": true,
	"defaultAction": "ask",
	"defaultActionVerdict": "exact",
	"defaultActionRendered": "ask",
	"verdicts": [
		{
			"rule": {
				"id": "rule-1",
				"tool": "bash",
				"action": "ask",
				"enforcement": "best-effort"
			},
			"verdict": "degraded",
			"rendered": "deny",
			"reason": "codex cannot enforce ask for bash; degrades to deny",
			"suggestion": "set enforcement: best-effort, or target a capable agent"
		}
	]
}`

func TestHostedAgents_ValidatePolicy(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/sessions/policy/validate", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		assert.Equal(t, "application/x-yaml", r.Header.Get("Content-Type"))
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		assert.Contains(t, string(body), "adapter: codex")
		assert.Contains(t, string(body), "defaultAction: ask")
		fmt.Fprint(w, hostedAgentPolicyValidationResultJSON)
	})

	got, resp, err := client.HostedAgents.ValidatePolicy(ctx, []byte(hostedAgentPolicyValidateManifest))
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.True(t, got.OK)
	assert.Equal(t, "codex", got.Agent)
	assert.Equal(t, "0.38", got.Version)
	assert.Equal(t, HostedAgentPolicyActionAsk, got.DefaultAction)
	assert.Equal(t, HostedAgentPolicyVerdictExact, got.DefaultActionVerdict)
	assert.Equal(t, HostedAgentPolicyActionAsk, got.DefaultActionRendered)
	require.Len(t, got.Verdicts, 1)
	assert.Equal(t, "rule-1", got.Verdicts[0].Rule.ID)
	assert.Equal(t, "bash", got.Verdicts[0].Rule.Tool)
	assert.Equal(t, HostedAgentPolicyVerdictDegraded, got.Verdicts[0].Verdict)
	assert.Equal(t, HostedAgentPolicyActionDeny, got.Verdicts[0].Rendered)
}

func TestHostedAgents_ValidatePolicy_EmptyManifest(t *testing.T) {
	setup()
	defer teardown()

	_, _, err := client.HostedAgents.ValidatePolicy(ctx, []byte("  \n"))
	require.EqualError(t, err, "hosted agents: manifest is required")
}

func TestHostedAgents_ValidatePolicy_NotOK(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/sessions/policy/validate", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{
			"agent": "codex",
			"version": "0.38",
			"ok": false,
			"defaultAction": "allow",
			"defaultActionVerdict": "unverified",
			"defaultActionRendered": "ask",
			"defaultActionReason": "no capability descriptor; allow downgraded to ask",
			"verdicts": []
		}`)
	})

	got, resp, err := client.HostedAgents.ValidatePolicy(ctx, []byte(hostedAgentPolicyValidateManifest))
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.False(t, got.OK)
	assert.Equal(t, HostedAgentPolicyActionAllow, got.DefaultAction)
	assert.Equal(t, HostedAgentPolicyVerdictUnverified, got.DefaultActionVerdict)
	assert.Equal(t, HostedAgentPolicyActionAsk, got.DefaultActionRendered)
	assert.Equal(t, "no capability descriptor; allow downgraded to ask", got.DefaultActionReason)
}

func TestHostedAgentPolicyValidationResult_JSONOmitsEmptyVerdicts(t *testing.T) {
	body, err := json.Marshal(&HostedAgentPolicyValidationResult{
		Agent:   "codex",
		Version: "0.38",
		OK:      true,
	})
	require.NoError(t, err)
	assert.JSONEq(t, `{"agent":"codex","version":"0.38","ok":true}`, string(body))
}
