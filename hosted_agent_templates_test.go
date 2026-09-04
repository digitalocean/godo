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

var hostedAgentTemplate = HostedAgentTemplate{
	TemplateID: "019fb39c-14d9-7080-933e-b9b90e25acda",
	Name:       "hermes-prod",
	Spec: &HostedAgentTemplateSpec{
		BaseTemplate: "coding-opencode",
		Image: &HostedAgentTemplateImageSource{
			Registry:   "registry.digitalocean.com",
			Repository: "acme-mars/hermes",
			Tag:        "2026-09-01",
			Digest:     "sha256:abc123",
		},
	},
	Status:    HostedAgentTemplateStatusBuilding,
	CreatedAt: Timestamp{Time: time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)},
	UpdatedAt: Timestamp{Time: time.Date(2026, 9, 1, 12, 1, 0, 0, time.UTC)},
}

const hostedAgentTemplateJSON = `{
	"template_id": "019fb39c-14d9-7080-933e-b9b90e25acda",
	"name": "hermes-prod",
	"spec": {
		"base_template": "coding-opencode",
		"image": {
			"registry": "registry.digitalocean.com",
			"repository": "acme-mars/hermes",
			"tag": "2026-09-01",
			"digest": "sha256:abc123"
		}
	},
	"status": "TEMPLATE_STATUS_BUILDING",
	"created_at": "2026-09-01T12:00:00Z",
	"updated_at": "2026-09-01T12:01:00Z"
}`

var hostedAgentTemplateBuild = HostedAgentTemplateBuild{
	BuildID:    "019fb39c-aaaa-7080-933e-b9b90e25acda",
	TemplateID: "019fb39c-14d9-7080-933e-b9b90e25acda",
	Name:       "hermes-prod",
	Spec: &HostedAgentTemplateSpec{
		BaseTemplate: "coding-opencode",
		Image: &HostedAgentTemplateImageSource{
			Registry:   "registry.digitalocean.com",
			Repository: "acme-mars/hermes",
			Tag:        "2026-09-01",
		},
	},
	Status:    HostedAgentTemplateBuildStatusSucceeded,
	CreatedAt: Timestamp{Time: time.Date(2026, 9, 1, 12, 0, 5, 0, time.UTC)},
	UpdatedAt: Timestamp{Time: time.Date(2026, 9, 1, 12, 0, 30, 0, time.UTC)},
}

const hostedAgentTemplateBuildJSON = `{
	"build_id": "019fb39c-aaaa-7080-933e-b9b90e25acda",
	"template_id": "019fb39c-14d9-7080-933e-b9b90e25acda",
	"name": "hermes-prod",
	"spec": {
		"base_template": "coding-opencode",
		"image": {
			"registry": "registry.digitalocean.com",
			"repository": "acme-mars/hermes",
			"tag": "2026-09-01"
		}
	},
	"status": "SUCCEEDED",
	"created_at": "2026-09-01T12:00:05Z",
	"updated_at": "2026-09-01T12:00:30Z"
}`

func TestHostedAgents_ListTemplates(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/templates", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		assert.Equal(t, "25", r.URL.Query().Get("page_size"))
		assert.Equal(t, "tok-1", r.URL.Query().Get("page_token"))
		fmt.Fprintf(w, `{"templates":[%s],"next_page_token":"tok-2"}`, hostedAgentTemplateJSON)
	})

	got, resp, err := client.HostedAgents.ListTemplates(ctx, &HostedAgentTemplateListOptions{
		PageSize:  25,
		PageToken: "tok-1",
	})
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	require.Len(t, got.Templates, 1)
	assert.Equal(t, hostedAgentTemplate, got.Templates[0])
	assert.Equal(t, "tok-2", got.NextPageToken)
}

func TestHostedAgents_CreateTemplate(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/templates", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body HostedAgentTemplateCreateRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		assert.Equal(t, "hermes-prod", body.Name)
		assert.Equal(t, "coding-opencode", body.BaseTemplate)
		assert.Equal(t, "registry.digitalocean.com/acme-mars/hermes:2026-09-01", body.SourceOCIRef)
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, `{"template":%s}`, hostedAgentTemplateJSON)
	})

	got, resp, err := client.HostedAgents.CreateTemplate(ctx, &HostedAgentTemplateCreateRequest{
		Name:         "hermes-prod",
		BaseTemplate: "coding-opencode",
		SourceOCIRef: "registry.digitalocean.com/acme-mars/hermes:2026-09-01",
	})
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	assert.Equal(t, hostedAgentTemplate, *got)
}

func TestHostedAgents_GetTemplate(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/templates/019fb39c-14d9-7080-933e-b9b90e25acda", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprintf(w, `{"template":%s}`, hostedAgentTemplateJSON)
	})

	got, resp, err := client.HostedAgents.GetTemplate(ctx, "019fb39c-14d9-7080-933e-b9b90e25acda")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, hostedAgentTemplate, *got)
}

func TestHostedAgents_UpdateTemplate(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/templates/019fb39c-14d9-7080-933e-b9b90e25acda", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		var body HostedAgentTemplateUpdateRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		assert.Equal(t, "registry.digitalocean.com/acme-mars/hermes:2026-09-02", body.SourceOCIRef)
		assert.Equal(t, "coding-codex", body.BaseTemplate)
		fmt.Fprintf(w, `{"template":%s}`, hostedAgentTemplateJSON)
	})

	got, resp, err := client.HostedAgents.UpdateTemplate(ctx, "019fb39c-14d9-7080-933e-b9b90e25acda", &HostedAgentTemplateUpdateRequest{
		SourceOCIRef: "registry.digitalocean.com/acme-mars/hermes:2026-09-02",
		BaseTemplate: "coding-codex",
	})
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, hostedAgentTemplate, *got)
}

func TestHostedAgents_DeleteTemplate(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/templates/019fb39c-14d9-7080-933e-b9b90e25acda", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		fmt.Fprint(w, `{"template_id":"019fb39c-14d9-7080-933e-b9b90e25acda","deleted":true}`)
	})

	got, resp, err := client.HostedAgents.DeleteTemplate(ctx, "019fb39c-14d9-7080-933e-b9b90e25acda")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "019fb39c-14d9-7080-933e-b9b90e25acda", got.TemplateID)
	assert.True(t, got.Deleted)
}

func TestHostedAgents_ListTemplateBuilds(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/templates/019fb39c-14d9-7080-933e-b9b90e25acda/builds", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		assert.Equal(t, "10", r.URL.Query().Get("page_size"))
		fmt.Fprintf(w, `{"builds":[%s],"next_page_token":"tok-b"}`, hostedAgentTemplateBuildJSON)
	})

	got, resp, err := client.HostedAgents.ListTemplateBuilds(ctx, "019fb39c-14d9-7080-933e-b9b90e25acda", &HostedAgentTemplateBuildListOptions{
		PageSize: 10,
	})
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	require.Len(t, got.Builds, 1)
	assert.Equal(t, hostedAgentTemplateBuild, got.Builds[0])
	assert.Equal(t, "tok-b", got.NextPageToken)
}

func TestHostedAgents_GetTemplateBuild(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/templates/019fb39c-14d9-7080-933e-b9b90e25acda/builds/019fb39c-aaaa-7080-933e-b9b90e25acda", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprintf(w, `{"build":%s}`, hostedAgentTemplateBuildJSON)
	})

	got, resp, err := client.HostedAgents.GetTemplateBuild(ctx, "019fb39c-14d9-7080-933e-b9b90e25acda", "019fb39c-aaaa-7080-933e-b9b90e25acda")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, hostedAgentTemplateBuild, *got)
}

func TestHostedAgents_GetTemplateBuildLogs(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/templates/019fb39c-14d9-7080-933e-b9b90e25acda/builds/019fb39c-aaaa-7080-933e-b9b90e25acda/logs", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{"signed_url":"https://spaces.example/logs?sig=abc"}`)
	})

	got, resp, err := client.HostedAgents.GetTemplateBuildLogs(ctx, "019fb39c-14d9-7080-933e-b9b90e25acda", "019fb39c-aaaa-7080-933e-b9b90e25acda")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "https://spaces.example/logs?sig=abc", got.SignedURL)
}

func TestHostedAgents_TemplateValidationErrors(t *testing.T) {
	setup()
	defer teardown()

	_, _, err := client.HostedAgents.CreateTemplate(ctx, nil)
	require.EqualError(t, err, "hosted agents: create template request is required")

	_, _, err = client.HostedAgents.CreateTemplate(ctx, &HostedAgentTemplateCreateRequest{})
	require.EqualError(t, err, "hosted agents: name is required")

	_, _, err = client.HostedAgents.CreateTemplate(ctx, &HostedAgentTemplateCreateRequest{Name: "x"})
	require.EqualError(t, err, "hosted agents: base_template is required")

	_, _, err = client.HostedAgents.CreateTemplate(ctx, &HostedAgentTemplateCreateRequest{
		Name:         "x",
		BaseTemplate: "coding-base",
	})
	require.EqualError(t, err, "hosted agents: source_oci_ref is required")

	_, _, err = client.HostedAgents.GetTemplate(ctx, "")
	require.EqualError(t, err, "hosted agents: template id is required")

	_, _, err = client.HostedAgents.UpdateTemplate(ctx, "id", nil)
	require.EqualError(t, err, "hosted agents: update template request is required")

	_, _, err = client.HostedAgents.DeleteTemplate(ctx, "")
	require.EqualError(t, err, "hosted agents: template id is required")

	_, _, err = client.HostedAgents.ListTemplateBuilds(ctx, "", nil)
	require.EqualError(t, err, "hosted agents: template id is required")

	_, _, err = client.HostedAgents.GetTemplateBuild(ctx, "id", "")
	require.EqualError(t, err, "hosted agents: build id is required")

	_, _, err = client.HostedAgents.GetTemplateBuildLogs(ctx, "", "build")
	require.EqualError(t, err, "hosted agents: template id is required")
}

func TestHostedAgents_CreateTemplate_EmptyBody(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/agents/templates", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{}`)
	})

	_, resp, err := client.HostedAgents.CreateTemplate(ctx, &HostedAgentTemplateCreateRequest{
		Name:         "hermes-prod",
		BaseTemplate: "coding-opencode",
		SourceOCIRef: "registry.digitalocean.com/acme-mars/hermes:2026-09-01",
	})
	require.EqualError(t, err, "hosted agents: create template returned no template")
	require.NotNil(t, resp)
}
