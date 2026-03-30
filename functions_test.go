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

func TestFunctions_ListNamespaces(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/functions/namespaces", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"namespaces": [
				{
					"api_host": "https://faas.do.com",
					"namespace": "123-abc",
					"created_at": "2022-06-16T12:09:13Z",
					"updated_at": "2022-06-16T12:09:13Z",
					"label": "my-namespace-1",
					"region": "nyc1",
					"uuid": "",
					"key": ""
				},
				{
					"api_host": "https://faas.do.com",
					"namespace": "456-abc",
					"created_at": "2022-11-02T18:33:44Z",
					"updated_at": "2022-11-02T18:33:44Z",
					"label": "my-namespace-2",
					"region": "nyc3",
					"uuid": "",
					"key": ""
				}
			]
		}`)
	})

	namespaces, _, err := client.Functions.ListNamespaces(ctx)
	require.NoError(t, err)

	expectedNamespaces := []FunctionsNamespace{
		{
			ApiHost:   "https://faas.do.com",
			Namespace: "123-abc",
			CreatedAt: time.Date(2022, 6, 16, 12, 9, 13, 0, time.UTC),
			UpdatedAt: time.Date(2022, 6, 16, 12, 9, 13, 0, time.UTC),
			Label:     "my-namespace-1",
			Region:    "nyc1",
			UUID:      "",
			Key:       "",
		},
		{
			ApiHost:   "https://faas.do.com",
			Namespace: "456-abc",
			CreatedAt: time.Date(2022, 11, 2, 18, 33, 44, 0, time.UTC),
			UpdatedAt: time.Date(2022, 11, 2, 18, 33, 44, 0, time.UTC),
			Label:     "my-namespace-2",
			Region:    "nyc3",
			UUID:      "",
			Key:       "",
		},
	}
	assert.Equal(t, expectedNamespaces, namespaces)
}

func TestFunctions_GetNamespace(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/functions/namespaces/123-abc", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"namespace": {
				"api_host": "https://faas.do.com",
				"namespace": "123-abc",
				"created_at": "2022-06-16T12:09:13Z",
				"updated_at": "2022-06-16T12:09:13Z",
				"label": "my-namespace-1",
				"region": "nyc1",
				"uuid": "123-456",
				"key": "abc-123"
			}
		}`)
	})

	namespace, _, err := client.Functions.GetNamespace(ctx, "123-abc")
	require.NoError(t, err)

	expectedNamespace := &FunctionsNamespace{
		ApiHost:   "https://faas.do.com",
		Namespace: "123-abc",
		CreatedAt: time.Date(2022, 6, 16, 12, 9, 13, 0, time.UTC),
		UpdatedAt: time.Date(2022, 6, 16, 12, 9, 13, 0, time.UTC),
		Label:     "my-namespace-1",
		Region:    "nyc1",
		UUID:      "123-456",
		Key:       "abc-123",
	}
	assert.Equal(t, expectedNamespace, namespace)
}

func TestFunctions_CreateNamespace(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/functions/namespaces", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `{
			"namespace": {
				"api_host": "https://faas.do.com",
				"namespace": "123-abc",
				"created_at": "2022-06-16T12:09:13Z",
				"updated_at": "2022-06-16T12:09:13Z",
				"label": "my-namespace-1",
				"region": "nyc1",
				"uuid": "123-456",
				"key": "abc-123"
			}
		}`)
	})

	opts := FunctionsNamespaceCreateRequest{Label: "my-namespace-1", Region: "nyc1"}
	namespace, _, err := client.Functions.CreateNamespace(ctx, &opts)
	require.NoError(t, err)

	expectedNamespace := &FunctionsNamespace{
		ApiHost:   "https://faas.do.com",
		Namespace: "123-abc",
		CreatedAt: time.Date(2022, 6, 16, 12, 9, 13, 0, time.UTC),
		UpdatedAt: time.Date(2022, 6, 16, 12, 9, 13, 0, time.UTC),
		Label:     "my-namespace-1",
		Region:    "nyc1",
		UUID:      "123-456",
		Key:       "abc-123",
	}
	assert.Equal(t, expectedNamespace, namespace)
}

func TestFunctions_DeleteNamespace(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/functions/namespaces/123-abc", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
	})

	_, err := client.Functions.DeleteNamespace(ctx, "123-abc")

	assert.NoError(t, err)
}

func TestFunctions_ListTriggers(t *testing.T) {
	setup()
	defer teardown()
	mux.HandleFunc("/v2/functions/namespaces/123-456/triggers", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"triggers": [
				{
					"name": "trigger",
					"namespace": "123-456",
					"function": "my_func",
					"type": "SCHEDULED",
					"is_enabled": true,
					"created_at": "2022-10-05T13:46:59Z",
					"updated_at": "2022-10-17T18:41:30Z",
					"scheduled_details": {
						"cron": "* * * * *",
						"body": {
						"foo": "bar"
						}
					},
					"scheduled_runs": {
						"next_run_at": "2022-11-03T17:03:02Z"
					}
				},			
				{
					"name": "trigger1",
					"namespace": "123-456",
					"function": "sample/hello",
					"type": "SCHEDULED",
					"is_enabled": true,
					"created_at": "2022-10-14T20:29:43Z",
					"updated_at": "2022-10-14T20:29:43Z",
					"scheduled_details": {
						"cron": "* * * * *",
						"body": {}
					},
					"scheduled_runs": {
						"last_run_at": "2022-11-03T17:02:43Z",
						"next_run_at": "2022-11-03T17:02:47Z"
					}
				}	
			]
		}`)
	})

	triggers, _, err := client.Functions.ListTriggers(ctx, "123-456")
	require.NoError(t, err)

	expectedTriggers := []FunctionsTrigger{
		{
			Name:      "trigger",
			Namespace: "123-456",
			Function:  "my_func",
			Type:      "SCHEDULED",
			IsEnabled: true,
			CreatedAt: time.Date(2022, 10, 5, 13, 46, 59, 0, time.UTC),
			UpdatedAt: time.Date(2022, 10, 17, 18, 41, 30, 0, time.UTC),
			ScheduledDetails: &TriggerScheduledDetails{
				Cron: "* * * * *",
				Body: map[string]interface{}{
					"foo": "bar",
				},
			},
			ScheduledRuns: &TriggerScheduledRuns{
				NextRunAt: time.Date(2022, 11, 3, 17, 3, 2, 0, time.UTC),
			},
		},
		{
			Name:      "trigger1",
			Namespace: "123-456",
			Function:  "sample/hello",
			Type:      "SCHEDULED",
			IsEnabled: true,
			CreatedAt: time.Date(2022, 10, 14, 20, 29, 43, 0, time.UTC),
			UpdatedAt: time.Date(2022, 10, 14, 20, 29, 43, 0, time.UTC),
			ScheduledDetails: &TriggerScheduledDetails{
				Cron: "* * * * *",
				Body: map[string]interface{}{},
			},
			ScheduledRuns: &TriggerScheduledRuns{
				LastRunAt: time.Date(2022, 11, 03, 17, 02, 43, 0, time.UTC),
				NextRunAt: time.Date(2022, 11, 03, 17, 02, 47, 0, time.UTC),
			},
		},
	}
	assert.Equal(t, expectedTriggers, triggers)
}

func TestFunctions_GetTrigger(t *testing.T) {
	setup()
	defer teardown()
	mux.HandleFunc("/v2/functions/namespaces/123-456/triggers/my-trigger", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"trigger": {
				"name": "my-trigger",
				"namespace": "123-456",
				"function": "my_func",
				"type": "SCHEDULED",
				"is_enabled": true,
				"created_at": "2022-10-05T13:46:59Z",
				"updated_at": "2022-10-17T18:41:30Z",
				"scheduled_details": {
					"cron": "* * * * *",
					"body": {
						"foo": "bar"
					}
				},
				"scheduled_runs": {
					"next_run_at": "2022-11-03T17:03:02Z"
				}
			}	
		}`)

	})

	trigger, _, err := client.Functions.GetTrigger(ctx, "123-456", "my-trigger")
	require.NoError(t, err)

	expectedTrigger := &FunctionsTrigger{
		Name:      "my-trigger",
		Namespace: "123-456",
		Function:  "my_func",
		Type:      "SCHEDULED",
		IsEnabled: true,
		CreatedAt: time.Date(2022, 10, 5, 13, 46, 59, 0, time.UTC),
		UpdatedAt: time.Date(2022, 10, 17, 18, 41, 30, 0, time.UTC),
		ScheduledDetails: &TriggerScheduledDetails{
			Cron: "* * * * *",
			Body: map[string]interface{}{
				"foo": "bar",
			},
		},
		ScheduledRuns: &TriggerScheduledRuns{
			NextRunAt: time.Date(2022, 11, 3, 17, 3, 2, 0, time.UTC),
		},
	}
	assert.Equal(t, expectedTrigger, trigger)

}

func TestFunctions_CreateTrigger(t *testing.T) {
	setup()
	defer teardown()
	mux.HandleFunc("/v2/functions/namespaces/123-456/triggers", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `{
			"trigger": {
				"name": "my-new-trigger",
				"namespace": "123-456",
				"function": "my_func",
				"type": "SCHEDULED",
				"is_enabled": true,
				"created_at": "2022-10-05T13:46:59Z",
				"updated_at": "2022-10-17T18:41:30Z",
				"scheduled_details": {
					"cron": "* * * * *",
					"body": {
						"foo": "bar"
					}
				},
				"scheduled_runs": {
					"next_run_at": "2022-11-03T17:03:02Z"
				}
			}
		}`)
	})

	opts := FunctionsTriggerCreateRequest{
		Name:      "my-new-trigger",
		Function:  "my_func",
		Type:      "SCHEDULED",
		IsEnabled: true,
		ScheduledDetails: &TriggerScheduledDetails{
			Cron: "* * * * *",
			Body: map[string]interface{}{
				"foo": "bar",
			},
		},
	}
	trigger, _, err := client.Functions.CreateTrigger(ctx, "123-456", &opts)
	require.NoError(t, err)
	expectedTrigger := &FunctionsTrigger{
		Name:      "my-new-trigger",
		Namespace: "123-456",
		Function:  "my_func",
		Type:      "SCHEDULED",
		IsEnabled: true,
		CreatedAt: time.Date(2022, 10, 5, 13, 46, 59, 0, time.UTC),
		UpdatedAt: time.Date(2022, 10, 17, 18, 41, 30, 0, time.UTC),
		ScheduledDetails: &TriggerScheduledDetails{
			Cron: "* * * * *",
			Body: map[string]interface{}{
				"foo": "bar",
			},
		},
		ScheduledRuns: &TriggerScheduledRuns{
			NextRunAt: time.Date(2022, 11, 3, 17, 3, 2, 0, time.UTC),
		},
	}

	assert.Equal(t, expectedTrigger, trigger)
}

func TestFunctions_UpdateTrigger(t *testing.T) {
	setup()
	defer teardown()
	mux.HandleFunc("/v2/functions/namespaces/123-456/triggers/my-trigger", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		fmt.Fprint(w, `{
			"trigger": {
			"name": "my-trigger",
			"namespace": "123-456",
			"function": "my_func",
			"type": "SCHEDULED",
			"is_enabled": false,
			"created_at": "2022-10-05T13:46:59Z",
			"updated_at": "2022-10-17T18:41:30Z",
			"scheduled_details": {
				"cron": "* * * * *",
				"body": {
					"foo": "bar"
				}
			},
			"scheduled_runs": {
				"next_run_at": "2022-11-03T17:03:02Z"
			}
		}
	}`)
	})

	isEnabled := false
	opts := FunctionsTriggerUpdateRequest{
		IsEnabled: &isEnabled,
	}

	trigger, _, err := client.Functions.UpdateTrigger(ctx, "123-456", "my-trigger", &opts)
	require.NoError(t, err)

	expectedTrigger := &FunctionsTrigger{
		Name:      "my-trigger",
		Namespace: "123-456",
		Function:  "my_func",
		Type:      "SCHEDULED",
		IsEnabled: false,
		CreatedAt: time.Date(2022, 10, 5, 13, 46, 59, 0, time.UTC),
		UpdatedAt: time.Date(2022, 10, 17, 18, 41, 30, 0, time.UTC),
		ScheduledDetails: &TriggerScheduledDetails{
			Cron: "* * * * *",
			Body: map[string]interface{}{
				"foo": "bar",
			},
		},
		ScheduledRuns: &TriggerScheduledRuns{
			NextRunAt: time.Date(2022, 11, 3, 17, 3, 2, 0, time.UTC),
		},
	}
	assert.Equal(t, expectedTrigger, trigger)
}

func TestFunctions_DeleteTrigger(t *testing.T) {
	setup()
	defer teardown()
	mux.HandleFunc("/v2/functions/namespaces/123-abc/triggers/my-trigger", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
	})

	_, err := client.Functions.DeleteTrigger(ctx, "123-abc", "my-trigger")
	assert.NoError(t, err)
}

func TestFunctions_ListAccessKeys(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/functions/namespaces/123-abc/keys", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"access_keys": [
				{
					"id": "dof_v1_key1",
					"name": "my-key-1",
					"expires_at": "2025-06-16T12:09:13Z",
					"created_at": "2024-06-16T12:09:13Z",
					"updated_at": "2024-06-16T12:09:13Z"
				},
				{
					"id": "dof_v1_key2",
					"name": "my-key-2",
					"expires_at": "2025-11-02T18:33:44Z",
					"created_at": "2024-11-02T18:33:44Z",
					"updated_at": "2024-11-02T18:33:44Z"
				}
			],
			"count": 2
		}`)
	})

	accessKeys, _, err := client.Functions.ListAccessKeys(ctx, "123-abc")
	require.NoError(t, err)

	expectedKeys := []FunctionsAccessKey{
		{
			ID:        "dof_v1_key1",
			Name:      "my-key-1",
			ExpiresAt: time.Date(2025, 6, 16, 12, 9, 13, 0, time.UTC),
			CreatedAt: time.Date(2024, 6, 16, 12, 9, 13, 0, time.UTC),
			UpdatedAt: time.Date(2024, 6, 16, 12, 9, 13, 0, time.UTC),
		},
		{
			ID:        "dof_v1_key2",
			Name:      "my-key-2",
			ExpiresAt: time.Date(2025, 11, 2, 18, 33, 44, 0, time.UTC),
			CreatedAt: time.Date(2024, 11, 2, 18, 33, 44, 0, time.UTC),
			UpdatedAt: time.Date(2024, 11, 2, 18, 33, 44, 0, time.UTC),
		},
	}
	assert.Equal(t, expectedKeys, accessKeys)
}

func TestFunctions_CreateAccessKey(t *testing.T) {
	setup()
	defer teardown()

	opts := &FunctionsAccessKeyCreateRequest{
		Name:      "my-new-key",
		ExpiresIn: "1h",
	}

	mux.HandleFunc("/v2/functions/namespaces/123-abc/keys", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)

		v := new(FunctionsAccessKeyCreateRequest)
		err := json.NewDecoder(r.Body).Decode(v)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, opts, v)

		fmt.Fprint(w, `{
			"access_key": {
				"id": "dof_v1_newkey",
				"name": "my-new-key",
				"secret": "dof_v1_newkey:secret123",
				"expires_at": "2025-06-16T12:09:13Z",
				"created_at": "2024-06-16T12:09:13Z",
				"updated_at": "2024-06-16T12:09:13Z"
			}
		}`)
	})

	accessKey, _, err := client.Functions.CreateAccessKey(ctx, "123-abc", opts)
	require.NoError(t, err)

	expectedKey := &FunctionsAccessKey{
		ID:        "dof_v1_newkey",
		Name:      "my-new-key",
		Secret:    "dof_v1_newkey:secret123",
		ExpiresAt: time.Date(2025, 6, 16, 12, 9, 13, 0, time.UTC),
		CreatedAt: time.Date(2024, 6, 16, 12, 9, 13, 0, time.UTC),
		UpdatedAt: time.Date(2024, 6, 16, 12, 9, 13, 0, time.UTC),
	}
	assert.Equal(t, expectedKey, accessKey)
}

func TestFunctions_UpdateAccessKey(t *testing.T) {
	setup()
	defer teardown()

	opts := &FunctionsAccessKeyUpdateRequest{
		Name: "renamed-key",
	}

	mux.HandleFunc("/v2/functions/namespaces/123-abc/keys/dof_v1_key1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)

		v := new(FunctionsAccessKeyUpdateRequest)
		err := json.NewDecoder(r.Body).Decode(v)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, opts, v)

		fmt.Fprint(w, `{
			"access_key": {
				"id": "dof_v1_key1",
				"name": "renamed-key",
				"expires_at": "2025-06-16T12:09:13Z",
				"created_at": "2024-06-16T12:09:13Z",
				"updated_at": "2024-08-20T10:00:00Z"
			}
		}`)
	})

	accessKey, _, err := client.Functions.UpdateAccessKey(ctx, "123-abc", "dof_v1_key1", opts)
	require.NoError(t, err)

	expectedKey := &FunctionsAccessKey{
		ID:        "dof_v1_key1",
		Name:      "renamed-key",
		ExpiresAt: time.Date(2025, 6, 16, 12, 9, 13, 0, time.UTC),
		CreatedAt: time.Date(2024, 6, 16, 12, 9, 13, 0, time.UTC),
		UpdatedAt: time.Date(2024, 8, 20, 10, 0, 0, 0, time.UTC),
	}
	assert.Equal(t, expectedKey, accessKey)
}

func TestFunctions_DeleteAccessKey(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/functions/namespaces/123-abc/keys/dof_v1_key1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
	})

	_, err := client.Functions.DeleteAccessKey(ctx, "123-abc", "dof_v1_key1")
	assert.NoError(t, err)
}
