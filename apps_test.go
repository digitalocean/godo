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
		Region: testAppRegion.Slug,
		Services: []*AppServiceSpec{{
			Name: "service-name",
			Routes: []*AppRouteSpec{{
				Path: "/",
			}},
			RunCommand:     "run-command",
			BuildCommand:   "build-command",
			DockerfilePath: "Dockerfile",
			GitHub: &GitHubSourceSpec{
				Repo:   "owner/service",
				Branch: "branch",
			},
			InstanceSizeSlug: "professional-xs",
			InstanceCount:    1,
		}},
		Workers: []*AppWorkerSpec{{
			Name:           "worker-name",
			RunCommand:     "run-command",
			BuildCommand:   "build-command",
			DockerfilePath: "Dockerfile",
			GitHub: &GitHubSourceSpec{
				Repo:   "owner/worker",
				Branch: "branch",
			},
			InstanceSizeSlug: "professional-xs",
			InstanceCount:    1,
		}},
		StaticSites: []*AppStaticSiteSpec{{
			Name:         "static-name",
			BuildCommand: "build-command",
			Git: &GitSourceSpec{
				RepoCloneURL: "git@githost.com/owner/static.git",
				Branch:       "branch",
			},
			OutputDir: "out",
		}},
		Jobs: []*AppJobSpec{{
			Name:           "job-name",
			RunCommand:     "run-command",
			BuildCommand:   "build-command",
			DockerfilePath: "Dockerfile",
			GitHub: &GitHubSourceSpec{
				Repo:   "owner/job",
				Branch: "branch",
			},
			InstanceSizeSlug: "professional-xs",
			InstanceCount:    1,
		}},
		Databases: []*AppDatabaseSpec{{
			Name:        "db",
			Engine:      AppDatabaseSpecEngine_MySQL,
			Version:     "8",
			Size:        "size",
			NumNodes:    1,
			Production:  true,
			ClusterName: "cluster-name",
			DBName:      "app",
			DBUser:      "appuser",
		}},
		Functions: []*AppFunctionsSpec{{
			Name: "functions-name",
			GitHub: &GitHubSourceSpec{
				Repo:   "git@githost.com/owner/functions.git",
				Branch: "branch",
			},
		}},
		Domains: []*AppDomainSpec{
			{
				Domain: "example.com",
				Type:   AppDomainSpecType_Primary,
			},
		},
	}

	testAppRegion = AppRegion{
		Slug:        "ams",
		Label:       "Amsterdam",
		Flag:        "netherlands",
		Continent:   "Europe",
		DataCenters: []string{"ams3"},
		Default:     true,
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
		Jobs: []*DeploymentJob{{
			Name:             "job-name",
			SourceCommitHash: "job-hash",
		}},
		Functions: []*DeploymentFunctions{{
			Name:             "functions-name",
			SourceCommitHash: "functions-hash",
		}},
		CreatedAt:          time.Unix(1595959200, 0).UTC(),
		UpdatedAt:          time.Unix(1595959200, 0).UTC(),
		PhaseLastUpdatedAt: time.Unix(1595959200, 0).UTC(),
		Phase:              DeploymentPhase_Active,
		Progress: &DeploymentProgress{
			SuccessSteps: 1,
			TotalSteps:   1,
			Steps: []*DeploymentProgressStep{{
				Name:      "step",
				Status:    DeploymentProgressStepStatus_Success,
				StartedAt: time.Unix(1595959200, 0).UTC(),
				EndedAt:   time.Unix(1595959200, 0).UTC(),
				Steps: []*DeploymentProgressStep{{
					Name:      "sub",
					Status:    DeploymentProgressStepStatus_Success,
					StartedAt: time.Unix(1595959200, 0).UTC(),
					EndedAt:   time.Unix(1595959200, 0).UTC(),
				}},
			}},
		},
	}

	testApp = App{
		ID:                      "1c70f8f3-106e-428b-ae6d-bfc693c77536",
		Spec:                    testAppSpec,
		DefaultIngress:          "example.com",
		LiveURL:                 "https://example.com",
		LiveURLBase:             "https://example.com",
		LiveDomain:              "example.com",
		ActiveDeployment:        &testDeployment,
		InProgressDeployment:    &testDeployment,
		LastDeploymentCreatedAt: time.Unix(1595959200, 0).UTC(),
		LastDeploymentActiveAt:  time.Unix(1595959200, 0).UTC(),
		CreatedAt:               time.Unix(1595959200, 0).UTC(),
		UpdatedAt:               time.Unix(1595959200, 0).UTC(),
		Region:                  &testAppRegion,
		TierSlug:                testAppTier.Slug,
	}

	testAppTier = AppTier{
		Name:                 "Test",
		Slug:                 "test",
		EgressBandwidthBytes: "10240",
		BuildSeconds:         "3000",
	}

	testInstanceSize = AppInstanceSize{
		Name:            "Basic XXS",
		Slug:            "basic-xxs",
		CPUType:         AppInstanceSizeCPUType_Dedicated,
		CPUs:            "1",
		MemoryBytes:     "536870912",
		USDPerMonth:     "5",
		USDPerSecond:    "0.0000018896447",
		TierSlug:        "basic",
		TierUpgradeTo:   "professional-xs",
		TierDowngradeTo: "basic-xxxs",
	}

	testAlerts = []*AppAlert{
		{
			ID: "c586fc0d-e8e2-4c50-9bf6-6c0a6b2ed2a7",
			Spec: &AppAlertSpec{
				Rule: AppAlertSpecRule_DeploymentFailed,
			},
			Emails: []string{"test@example.com", "test2@example.com"},
			SlackWebhooks: []*AppAlertSlackWebhook{
				{
					URL:     "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX",
					Channel: "channel name",
				},
			},
		},
	}

	testAlert = AppAlert{
		ID: "c586fc0d-e8e2-4c50-9bf6-6c0a6b2ed2a7",
		Spec: &AppAlertSpec{
			Rule: AppAlertSpecRule_DeploymentFailed,
		},
		Emails: []string{"test@example.com", "test2@example.com"},
		SlackWebhooks: []*AppAlertSlackWebhook{
			{
				URL:     "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX",
				Channel: "channel name",
			},
		},
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

		json.NewEncoder(w).Encode(&appsRoot{Apps: []*App{&testApp}, Meta: &Meta{Total: 1}, Links: &Links{}})
	})

	apps, resp, err := client.Apps.List(ctx, nil)
	require.NoError(t, err)
	assert.Equal(t, []*App{&testApp}, apps)
	assert.Equal(t, 1, resp.Meta.Total)
	currentPage, err := resp.Links.CurrentPage()
	require.NoError(t, err)
	assert.Equal(t, 1, currentPage)
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

func TestApps_ProposeApp(t *testing.T) {
	setup()
	defer teardown()

	ctx := context.Background()

	spec := &AppSpec{
		Name: "sample-golang",
		Services: []*AppServiceSpec{{
			Name:            "web",
			EnvironmentSlug: "go",
			RunCommand:      "bin/sample-golang",
			GitHub: &GitHubSourceSpec{
				Repo:   "digitalocean/sample-golang",
				Branch: "branch",
			},
		}},
	}

	mux.HandleFunc("/v2/apps/propose", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var req AppProposeRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		assert.Equal(t, spec, req.Spec)
		assert.Equal(t, testApp.ID, req.AppID)

		json.NewEncoder(w).Encode(&AppProposeResponse{
			Spec: &AppSpec{
				Name: "sample-golang",
				Services: []*AppServiceSpec{{
					Name:            "web",
					EnvironmentSlug: "go",
					RunCommand:      "bin/sample-golang",
					GitHub: &GitHubSourceSpec{
						Repo:   "digitalocean/sample-golang",
						Branch: "branch",
					},
					InstanceCount: 1,
					Routes: []*AppRouteSpec{{
						Path: "/",
					}},
				}},
			},
			AppNameAvailable: true,
		})
	})

	res, _, err := client.Apps.Propose(ctx, &AppProposeRequest{
		Spec:  spec,
		AppID: testApp.ID,
	})
	require.NoError(t, err)
	assert.Equal(t, int64(1), res.Spec.Services[0].InstanceCount)
	assert.Equal(t, "/", res.Spec.Services[0].Routes[0].Path)
	assert.True(t, res.AppNameAvailable)
}

func TestApps_CreateDeployment(t *testing.T) {
	for _, forceBuild := range []bool{true, false} {
		t.Run(fmt.Sprintf("ForceBuild_%t", forceBuild), func(t *testing.T) {
			setup()
			defer teardown()

			ctx := context.Background()

			mux.HandleFunc(fmt.Sprintf("/v2/apps/%s/deployments", testApp.ID), func(w http.ResponseWriter, r *http.Request) {
				testMethod(t, r, http.MethodPost)

				var req DeploymentCreateRequest
				err := json.NewDecoder(r.Body).Decode(&req)
				require.NoError(t, err)
				assert.Equal(t, forceBuild, req.ForceBuild)

				json.NewEncoder(w).Encode(&deploymentRoot{Deployment: &testDeployment})
			})

			deployment, _, err := client.Apps.CreateDeployment(ctx, testApp.ID, &DeploymentCreateRequest{
				ForceBuild: forceBuild,
			})
			require.NoError(t, err)
			assert.Equal(t, &testDeployment, deployment)
		})
	}
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

func TestApps_ListDeployments(t *testing.T) {
	setup()
	defer teardown()

	ctx := context.Background()

	mux.HandleFunc(fmt.Sprintf("/v2/apps/%s/deployments", testApp.ID), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)

		json.NewEncoder(w).Encode(&deploymentsRoot{Deployments: []*Deployment{&testDeployment}, Meta: &Meta{Total: 1}, Links: &Links{}})
	})

	deployments, resp, err := client.Apps.ListDeployments(ctx, testApp.ID, nil)
	require.NoError(t, err)
	assert.Equal(t, []*Deployment{&testDeployment}, deployments)
	assert.Equal(t, 1, resp.Meta.Total)
	currentPage, err := resp.Links.CurrentPage()
	require.NoError(t, err)
	assert.Equal(t, 1, currentPage)
}

func TestApps_GetLogs(t *testing.T) {
	setup()
	defer teardown()

	ctx := context.Background()

	mux.HandleFunc(fmt.Sprintf("/v2/apps/%s/deployments/%s/logs", testApp.ID, testDeployment.ID), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)

		assert.Equal(t, "RUN", r.URL.Query().Get("type"))
		assert.Equal(t, "true", r.URL.Query().Get("follow"))
		assert.Equal(t, "1", r.URL.Query().Get("tail_lines"))
		_, hasComponent := r.URL.Query()["component_name"]
		assert.False(t, hasComponent)

		json.NewEncoder(w).Encode(&AppLogs{LiveURL: "https://live.logs.url"})
	})

	logs, _, err := client.Apps.GetLogs(ctx, testApp.ID, testDeployment.ID, "", AppLogTypeRun, true, 1)
	require.NoError(t, err)
	assert.NotEmpty(t, logs.LiveURL)
}

func TestApps_GetLogs_component(t *testing.T) {
	setup()
	defer teardown()

	ctx := context.Background()

	mux.HandleFunc(fmt.Sprintf("/v2/apps/%s/deployments/%s/logs", testApp.ID, testDeployment.ID), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)

		assert.Equal(t, "RUN", r.URL.Query().Get("type"))
		assert.Equal(t, "true", r.URL.Query().Get("follow"))
		assert.Equal(t, "1", r.URL.Query().Get("tail_lines"))
		assert.Equal(t, "service-name", r.URL.Query().Get("component_name"))

		json.NewEncoder(w).Encode(&AppLogs{LiveURL: "https://live.logs.url"})
	})

	logs, _, err := client.Apps.GetLogs(ctx, testApp.ID, testDeployment.ID, "service-name", AppLogTypeRun, true, 1)
	require.NoError(t, err)
	assert.NotEmpty(t, logs.LiveURL)
}

func TestApps_ListRegions(t *testing.T) {
	setup()
	defer teardown()

	ctx := context.Background()

	mux.HandleFunc("/v2/apps/regions", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)

		json.NewEncoder(w).Encode(&appRegionsRoot{Regions: []*AppRegion{&testAppRegion}})
	})

	regions, _, err := client.Apps.ListRegions(ctx)
	require.NoError(t, err)
	assert.Equal(t, []*AppRegion{&testAppRegion}, regions)
}

func TestApps_ListTiers(t *testing.T) {
	setup()
	defer teardown()

	ctx := context.Background()

	mux.HandleFunc("/v2/apps/tiers", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)

		json.NewEncoder(w).Encode(&appTiersRoot{Tiers: []*AppTier{&testAppTier}})
	})

	tiers, _, err := client.Apps.ListTiers(ctx)
	require.NoError(t, err)
	assert.Equal(t, []*AppTier{&testAppTier}, tiers)
}

func TestApps_GetTier(t *testing.T) {
	setup()
	defer teardown()

	ctx := context.Background()

	mux.HandleFunc(fmt.Sprintf("/v2/apps/tiers/%s", testAppTier.Slug), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)

		json.NewEncoder(w).Encode(&appTierRoot{Tier: &testAppTier})
	})

	tier, _, err := client.Apps.GetTier(ctx, testAppTier.Slug)
	require.NoError(t, err)
	assert.Equal(t, &testAppTier, tier)
}

func TestApps_ListInstanceSizes(t *testing.T) {
	setup()
	defer teardown()

	ctx := context.Background()

	mux.HandleFunc("/v2/apps/tiers/instance_sizes", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)

		json.NewEncoder(w).Encode(&instanceSizesRoot{InstanceSizes: []*AppInstanceSize{&testInstanceSize}})
	})

	instanceSizes, _, err := client.Apps.ListInstanceSizes(ctx)
	require.NoError(t, err)
	assert.Equal(t, []*AppInstanceSize{&testInstanceSize}, instanceSizes)
}

func TestApps_GetInstanceSize(t *testing.T) {
	setup()
	defer teardown()

	ctx := context.Background()

	mux.HandleFunc(fmt.Sprintf("/v2/apps/tiers/instance_sizes/%s", testInstanceSize.Slug), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)

		json.NewEncoder(w).Encode(&instanceSizeRoot{InstanceSize: &testInstanceSize})
	})

	instancesize, _, err := client.Apps.GetInstanceSize(ctx, testInstanceSize.Slug)
	require.NoError(t, err)
	assert.Equal(t, &testInstanceSize, instancesize)
}

func TestApps_ListAppAlerts(t *testing.T) {
	setup()
	defer teardown()

	ctx := context.Background()

	mux.HandleFunc(fmt.Sprintf("/v2/apps/%s/alerts", testApp.ID), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)

		json.NewEncoder(w).Encode(&appAlertsRoot{Alerts: testAlerts})
	})

	appAlerts, _, err := client.Apps.ListAlerts(ctx, testApp.ID)
	require.NoError(t, err)
	assert.Equal(t, testAlerts, appAlerts)
}

func TestApps_UpdateAppAlertDestinations(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc(fmt.Sprintf("/v2/apps/%s/alerts/%s/destinations", testApp.ID, testAlert.ID), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)

		json.NewEncoder(w).Encode(&appAlertRoot{Alert: &testAlert})
	})

	appAlert, _, err := client.Apps.UpdateAlertDestinations(ctx, testApp.ID, testAlert.ID, &AlertDestinationUpdateRequest{Emails: testAlert.Emails, SlackWebhooks: testAlert.SlackWebhooks})
	require.NoError(t, err)
	assert.Equal(t, &testAlert, appAlert)
}
