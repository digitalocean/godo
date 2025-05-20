package godo

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	listAPIKeysResponse = `{
        "api_key_infos": [
            {
                "uuid": "123e4567-e89b-12d3-a456-426614174000",
                "name": "key1",
                "secret_key": "secret1",
                "created_at": "2024-01-01T10:00:00Z",
                "created_by": "12345"
            },
            {
                "uuid": "123e4567-e89b-12d3-a456-426614174001",
                "name": "key2",
                "secret_key": "secret2",
                "created_at": "2024-01-02T10:00:00Z",
                "created_by": "12345"
            }
        ]
    }`

	apiKeyInfoResponse = `{
        "api_key_info": {
            "uuid": "123e4567-e89b-12d3-a456-426614174000",
            "name": "key1",
            "secret_key": "secret1",
            "created_at": "2024-01-01T10:00:00Z",
            "created_by": "12345"
        }
    }`
)

func expectedAPIKeyList() []*AgentAPIKeyInfo {
	return []*AgentAPIKeyInfo{
		{
			UUID:      "123e4567-e89b-12d3-a456-426614174000",
			Name:      "key1",
			SecretKey: "secret1",
			CreatedAt: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
			CreatedBy: "12345",
		},
		{
			UUID:      "123e4567-e89b-12d3-a456-426614174001",
			Name:      "key2",
			SecretKey: "secret2",
			CreatedAt: time.Date(2024, 1, 2, 10, 0, 0, 0, time.UTC),
			CreatedBy: "12345",
		},
	}
}

func expectedAPIKeyInfo() *AgentAPIKeyInfo {
	return &AgentAPIKeyInfo{
		UUID:      "123e4567-e89b-12d3-a456-426614174000",
		Name:      "key1",
		SecretKey: "secret1",
		CreatedAt: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
		CreatedBy: "12345",
	}
}

func TestListAPIKeys(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/gen-ai/agents/agent-id/api_keys", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, listAPIKeysResponse)
	})

	keys, resp, err := client.GenAI.ListAPIKeys(ctx, "agent-id", nil)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.Response.StatusCode)
	assert.Equal(t, expectedAPIKeyList(), keys)
}

func TestCreateAPIKey(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/gen-ai/agents/agent-id/api_keys", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, apiKeyInfoResponse)
	})

	key, resp, err := client.GenAI.CreateAPIKey(ctx, "agent-id", &AgentCreateAPIRequest{})
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.Response.StatusCode)
	assert.Equal(t, expectedAPIKeyInfo(), key)
}

func TestUpdateAPIKey(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/gen-ai/agents/agent-id/api_keys/key-id", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		fmt.Fprint(w, apiKeyInfoResponse)
	})

	key, resp, err := client.GenAI.UpdateAPIKey(ctx, "agent-id", "key-id", &AgentUpdateAPIRequest{})
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.Response.StatusCode)
	assert.Equal(t, expectedAPIKeyInfo(), key)
}

func TestDeleteAPIKey(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/gen-ai/agents/agent-id/api_keys/key-id", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		fmt.Fprint(w, apiKeyInfoResponse)
	})

	key, resp, err := client.GenAI.DeleteAPIKey(ctx, "agent-id", "key-id")
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.Response.StatusCode)
	assert.Equal(t, expectedAPIKeyInfo(), key)
}

func TestRegenerateAPIKey(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/gen-ai/agents/agent-id/api_keys/key-id/regenerate", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		fmt.Fprint(w, apiKeyInfoResponse)
	})

	key, resp, err := client.GenAI.RegenerateAPIKey(ctx, "agent-id", "key-id")
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.Response.StatusCode)
	assert.Equal(t, expectedAPIKeyInfo(), key)
}
