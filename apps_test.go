package godo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	testAppSpec = &AppSpec{
		Name:   "app-name",
		Region: "ams3",
		Services: []AppServiceSpec{{
			Name: "service-name",
			Routes: []AppRouteSpec{{
				Path: "/",
			}},
			RunCommand:     "run-command",
			BuildCommand:   "build-command",
			DockerfilePath: "Dockerfile",
			GitHub: GitHubSourceSpec{
				Repo:   "owner/service",
				Branch: "branch",
			},
			InstanceSizeSlug: "professional-xs",
			InstanceCount:    1,
		}},
		Workers: []AppWorkerSpec{{
			Name:           "worker-name",
			RunCommand:     "run-command",
			BuildCommand:   "build-command",
			DockerfilePath: "Dockerfile",
			GitHub: GitHubSourceSpec{
				Repo:   "owner/worker",
				Branch: "branch",
			},
			InstanceSizeSlug: "professional-xs",
			InstanceCount:    1,
		}},
		StaticSites: []AppStaticSiteSpec{{
			Name:         "static-name",
			BuildCommand: "build-command",
			Git: GitSourceSpec{
				RepoCloneURL: "git@githost.com/owner/static.git",
				Branch:       "branch",
			},
		}},
	}

	testDeployment = Deployment{
		ID:   "08f10d33-94c3-4492-b9a3-1603e9ab7fe4",
		Spec: testAppSpec,
		Services: []*DeploymentService{{
			Name:             "service-name",
			SourceCommitHash: "service-hash",
		}},
		Workers: []*DeploymentWorker{{
			Name:             "worker-name",
			SourceCommitHash: "worker-hash",
		}},
		StaticSites: []*DeploymentStaticSite{{
			Name:             "static-name",
			SourceCommitHash: "static-hash",
		}},
	}

	testApp = App{
		ID:                   "1c70f8f3-106e-428b-ae6d-bfc693c77536",
		Spec:                 testAppSpec,
		DefaultIngress:       "test.ingress.com",
		ActiveDeployment:     &testDeployment,
		InProgressDeployment: &testDeployment,
		CreatedAt:            time.Unix(1, 0).UTC(),
		UpdatedAt:            time.Unix(1, 0).UTC(),
	}
)

func TestApps_CreateApp(t *testing.T) {
	setup()
	defer teardown()

	ctx := context.Background()

	mux.HandleFunc("/v2/apps", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var req AppCreateRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		assert.Equal(t, testAppSpec, req.Spec)

		json.NewEncoder(w).Encode(&appRoot{App: &testApp})
	})

	app, _, err := client.Apps.Create(ctx, &AppCreateRequest{Spec: testAppSpec})
	require.NoError(t, err)
	assert.Equal(t, &testApp, app)
}

func TestApps_GetApp(t *testing.T) {
	setup()
	defer teardown()

	ctx := context.Background()

	mux.HandleFunc(fmt.Sprintf("/v2/apps/%s", testApp.ID), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)

		json.NewEncoder(w).Encode(&appRoot{App: &testApp})
	})

	app, _, err := client.Apps.Get(ctx, testApp.ID)
	require.NoError(t, err)
	assert.Equal(t, &testApp, app)
}

func TestApps_ListApp(t *testing.T) {
	setup()
	defer teardown()

	ctx := context.Background()

	mux.HandleFunc("/v2/apps", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)

		json.NewEncoder(w).Encode(&appsRoot{Apps: []*App{&testApp}})
	})

	apps, _, err := client.Apps.List(ctx, nil)
	require.NoError(t, err)
	assert.Equal(t, []*App{&testApp}, apps)
}

func TestApps_UpdateApp(t *testing.T) {
	setup()
	defer teardown()

	ctx := context.Background()

	updatedSpec := *testAppSpec
	updatedSpec.Name = "new-name"

	mux.HandleFunc(fmt.Sprintf("/v2/apps/%s", testApp.ID), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		var req AppUpdateRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		assert.Equal(t, &updatedSpec, req.Spec)

		json.NewEncoder(w).Encode(&appRoot{App: &testApp})
	})

	app, _, err := client.Apps.Update(ctx, testApp.ID, &AppUpdateRequest{Spec: &updatedSpec})
	require.NoError(t, err)
	assert.Equal(t, &testApp, app)
}

func TestApps_DeleteApp(t *testing.T) {
	setup()
	defer teardown()

	ctx := context.Background()

	mux.HandleFunc(fmt.Sprintf("/v2/apps/%s", testApp.ID), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
	})

	_, err := client.Apps.Delete(ctx, testApp.ID)
	require.NoError(t, err)
}

func TestApps_CreateDeployment(t *testing.T) {
	setup()
	defer teardown()

	ctx := context.Background()

	mux.HandleFunc(fmt.Sprintf("/v2/apps/%s/deployments", testApp.ID), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)

		json.NewEncoder(w).Encode(&deploymentRoot{Deployment: &testDeployment})
	})

	deployment, _, err := client.Apps.CreateDeployment(ctx, testApp.ID)
	require.NoError(t, err)
	assert.Equal(t, &testDeployment, deployment)
}

func TestApps_GetDeployment(t *testing.T) {
	setup()
	defer teardown()

	ctx := context.Background()

	mux.HandleFunc(fmt.Sprintf("/v2/apps/%s/deployments/%s", testApp.ID, testDeployment.ID), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)

		json.NewEncoder(w).Encode(&deploymentRoot{Deployment: &testDeployment})
	})

	deployment, _, err := client.Apps.GetDeployment(ctx, testApp.ID, testDeployment.ID)
	require.NoError(t, err)
	assert.Equal(t, &testDeployment, deployment)
}

func TestApps_ListDeployment(t *testing.T) {
	setup()
	defer teardown()

	ctx := context.Background()

	mux.HandleFunc(fmt.Sprintf("/v2/apps/%s/deployments", testApp.ID), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)

		json.NewEncoder(w).Encode(&deploymentsRoot{Deployments: []*Deployment{&testDeployment}})
	})

	deployments, _, err := client.Apps.ListDeployments(ctx, testApp.ID, nil)
	require.NoError(t, err)
	assert.Equal(t, []*Deployment{&testDeployment}, deployments)
}

func TestApps_GetLogs(t *testing.T) {
	setup()
	defer teardown()

	ctx := context.Background()

	mux.HandleFunc(fmt.Sprintf("/v2/apps/%s/deployments/%s/components/%s/logs", testApp.ID, testDeployment.ID, "service-name"), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)

		json.NewEncoder(w).Encode(&AppLogs{LiveURL: "https://live.logs.url"})
	})

	logs, _, err := client.Apps.GetLogs(ctx, testApp.ID, testDeployment.ID, "service-name", AppLogTypeRun)
	require.NoError(t, err)
	assert.NotEmpty(t, logs.LiveURL)
}
