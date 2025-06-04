package godo

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

var listAgentResponse = `
{
	"agents": [
		{
			"uuid": "00000000-0000-0000-0000-000000000000",
			"name": "AMD ROCm Documentation Agent",
			"created_at": "2025-05-08T03:37:28Z",
			"updated_at": "2025-05-09T16:02:39Z",
			"instruction": "providing answers to common questions about the platform",
			"model": {
				"uuid": "00000000-0000-0000-0000-000000000000",
				"name": "Llama 3.3 Instruct (70B)",
				"inference_name": "llama3.3-70b-instruct",
				"version": {
					"major": 1
				},
				"inference_version": "4200000096",
				"is_foundational": true,
				"upload_complete": true,
				"created_at": "2025-01-13T20:56:20Z",
				"updated_at": "2025-05-13T15:16:21Z",
				"metadata": {
					"agreements": {
						"list": [
							{
								"title": "Licensing Agreement",
								"url": "https://llama.meta.com/llama3_1/license/"
							}
						],
						"title": "Llama 3.3"
					},
					"description": "An advanced language model with greater capabilities due to its larger size.",
					"max_tokens": {
						"default": 512,
						"max": 2048,
						"min": 256
					},
					"temperature": {
						"default": 0.7,
						"max": 1,
						"min": 0
					},
					"top_p": {
						"default": 0.9,
						"max": 1,
						"min": 0.1
					}
				},
				"parent_uuid": "00000000-0000-0000-0000-000000000000",
				"agreement": {
					"uuid": "00000000-0000-0000-0000-000000000000",
					"name": "Meta Llama 3.3 Community License",
					"description": "Meta Llama 3.3 is licensed under the Meta Llama 3.3 Community License, Copyright © Meta Platforms, Inc. All Rights Reserved. By purchasing, deploying, accessing, or using this model, you agree to comply with the",
					"url": "https://www.llama.com/llama3_3/license/"
				},
				"usecases": [
					"MODEL_USECASE_AGENT",
					"MODEL_USECASE_SERVERLESS"
				]
			},
			"deployment": {
				"uuid": "00000000-0000-0000-0000-000000000000",
				"url": "https://bndu2aqk2tnxj6ldwsatdmyu.agents.do-ai.run",
				"status": "STATUS_RUNNING",
				"visibility": "VISIBILITY_PUBLIC",
				"created_at": "2025-05-08T03:37:28Z",
				"updated_at": "2025-05-12T21:07:15Z"
			},
			"k": 0,
			"temperature": 0.7,
			"top_p": 0.9,
			"max_tokens": 2048,
			"project_id": "00000000-0000-0000-0000-000000000000",
			"route_uuid": "00000000-0000-0000-0000-000000000000",
			"region": "tor1",
			"chatbot": {
				"name": "AMD ROCm Documentation Agent",
				"primary_color": "#031B4E",
				"secondary_color": "#E5E8ED",
				"starting_message": "Hello! How can assist you with some documentation today?",
				"button_background_color": "#0061EB",
				"logo": "http://159.203.26.153/AMD-Logo.svg"
			},
			"route_created_at": "0001-01-01T00:00:00Z",
			"user_id": "14567334",
			"chatbot_identifiers": [
				{
					"agent_chatbot_identifier": "dfsvgdvscfvgf"
				}
			],
			"retrieval_method": "RETRIEVAL_METHOD_NONE"
		}
	],
	"links": {
		"pages": {
			"first": "https://api.digitalocean.com/v2/gen-ai/agents?page=1&per_page=1",
			"previous": "https://api.digitalocean.com/v2/gen-ai/agents?page=1&per_page=1",
			"next": "https://api.digitalocean.com/v2/gen-ai/agents?page=3&per_page=1",
			"last": "https://api.digitalocean.com/v2/gen-ai/agents?page=34&per_page=1"
		}
	},
	"meta": {
		"total": 34,
		"page": 2,
		"pages": 34
	}
}
`

var agentResponse = `
{
	"agent": {
		"uuid": "00000000-0000-0000-0000-000000000000",
		"name": "testing-godo",
		"created_at": "2025-05-14T13:18:05Z",
		"updated_at": "2025-05-14T13:18:05Z",
		"instruction": "You are an agent who thinks deeply about the world",
		"description": "My Agent Description",
		"model": {
			"uuid": "00000000-0000-0000-0000-000000000000",
			"name": "Llama 3.3 Instruct (70B)",
			"inference_name": "llama3.3-70b-instruct",
			"version": {
				"major": 1
			},
			"inference_version": "423242296",
			"is_foundational": true,
			"upload_complete": true,
			"created_at": "2025-01-13T20:56:20Z",
			"updated_at": "2025-05-13T15:16:21Z",
			"metadata": {
				"agreements": {
					"list": [
						{
							"title": "Licensing Agreement",
							"url": "https://llama.meta.com/llama3_1/license/"
						}
					],
					"title": "Llama 3.3"
				},
				"description": "An advanced language model with greater capabilities due to its larger size.",
				"max_tokens": {
					"default": 512,
					"max": 2048,
					"min": 256
				},
				"temperature": {
					"default": 0.7,
					"max": 1,
					"min": 0
				},
				"top_p": {
					"default": 0.9,
					"max": 1,
					"min": 0.1
				}
			},
			"parent_uuid": "00000000-0000-0000-0000-000000000000",
			"agreement": {
				"uuid": "00000000-0000-0000-0000-000000000000",
				"name": "Meta Llama 3.3 Community License",
				"description": "Meta Llama 3.3 is licensed under the Meta Llama 3.3 Community License, Copyright © Meta Platforms, Inc. All Rights Reserved. By purchasing, deploying, accessing, or using this model, you agree to comply with the",
				"url": "https://www.llama.com/llama3_3/license/"
			},
			"usecases": [
				"MODEL_USECASE_AGENT",
				"MODEL_USECASE_SERVERLESS"
			]
		},
		"deployment": {
			"uuid": "00000000-0000-0000-0000-000000000000",
			"status": "STATUS_WAITING_FOR_DEPLOYMENT",
			"visibility": "VISIBILITY_PRIVATE",
			"created_at": "2025-05-14T13:18:05Z",
			"updated_at": "2025-05-14T13:18:05Z"
		},
		"knowledge_bases": [
			{
				"uuid": "00000000-0000-0000-0000-000000000000",
				"name": "genai-api-testing",
				"created_at": "2025-05-14T07:52:27Z",
				"updated_at": "2025-05-14T07:57:27Z",
				"region": "nyc3",
				"embedding_model_uuid": "00000000-0000-0000-0000-000000000000",
				"project_id": "00000000-0000-0000-0000-000000000000",
				"database_id": "00000000-0000-0000-0000-000000000000",
				"last_indexing_job": {
					"uuid": "00000000-0000-0000-0000-000000000000",
					"knowledge_base_uuid": "00000000-0000-0000-0000-000000000000",
					"created_at": "2025-05-14T07:57:27Z",
					"updated_at": "2025-05-14T08:07:14Z",
					"started_at": "2025-05-14T07:57:27Z",
					"finished_at": "2025-05-14T07:58:05Z",
					"phase": "BATCH_JOB_PHASE_SUCCEEDED",
					"total_datasources": 1,
					"completed_datasources": 1,
					"tokens": 10060
				}
			}
		],
		"api_keys": [
			{
				"api_key": "dassfbgnhgbdsfvbgdvscfvg"
			}
		],
		"k": 10,
		"temperature": 0.7,
		"top_p": 0.9,
		"max_tokens": 512,
		"tags": [
			"string"
		],
		"project_id": "00000000-0000-0000-0000-000000000000",
		"route_uuid": "00000000-0000-0000-0000-000000000000",
		"region": "TOR1",
		"route_created_at": "0001-01-01T00:00:00Z",
		"user_id": "18919793",
		"chatbot_identifiers": [
			{
				"agent_chatbot_identifier": "dsfgbfsdsadsfbgfdfvsc"
			}
		],
		"retrieval_method": "RETRIEVAL_METHOD_NONE"
	}
}
`

var agentUpdateResponse = `
{
	"agent": {
		"uuid": "00000000-0000-0000-0000-000000000000",
		"name": "My Agent",
		"created_at": "2025-05-14T13:22:05Z",
		"updated_at": "2025-05-14T13:46:48Z",
		"instruction": "You are an agent who thinks deeply about the world",
		"description": "My Agent Description",
		"deployment": {
			"uuid": "00000000-0000-0000-0000-000000000000",
			"url": "https://hv5kn7pdaawp5772uk7nmyzd.agents.do-ai.run",
			"status": "STATUS_RUNNING",
			"visibility": "VISIBILITY_PUBLIC",
			"created_at": "2025-05-14T13:22:05Z",
			"updated_at": "2025-05-14T13:59:39Z"
		},
		"k": 10,
		"temperature": 0.7,
		"top_p": 0.9,
		"max_tokens": 512,
		"tags": [
			"updated",
			"example"
		],
		"project_id": "00000000-0000-0000-0000-000000000000",
		"route_uuid": "00000000-0000-0000-0000-000000000000",
		"region": "tor1",
		"chatbot": {
			"name": "My Agent Chatbot",
			"primary_color": "#031B4E",
			"secondary_color": "#E5E8ED",
			"starting_message": "Hello! How can I help you today?",
			"button_background_color": "#0061EB"
		},
		"route_created_at": "0001-01-01T00:00:00Z",
		"user_id": "18793",
		"retrieval_method": "RETRIEVAL_METHOD_NONE"
	}
}
`

var agentModelsResponse = `
{
	"models": [
		{
			"uuid": "00000000-0000-0000-0000-000000000000",
			"name": "Llama 3.3 Instruct (70B)",
			"version": {
				"major": 1
			},
			"is_foundational": true,
			"upload_complete": true,
			"created_at": "2025-01-13T20:56:20Z",
			"updated_at": "2025-05-13T15:16:21Z",
			"parent_uuid": "00000000-0000-0000-0000-000000000000",
			"agreement": {
				"uuid": "00000000-0000-0000-0000-000000000000",
				"name": "Meta Llama 3.3 Community License",
				"description": "Meta Llama 3.3 is licensed under the Meta Llama 3.3 Community License, Copyright © Meta Platforms, Inc. All Rights Reserved. By purchasing, deploying, accessing, or using this model, you agree to comply with the",
				"url": "https://www.llama.com/llama3_3/license/"
			}
		}
	],
	"links": {
		"pages": {
			"first": "https://api.digitalocean.com/v2/gen-ai/models?page=1&per_page=1",
			"next": "https://api.digitalocean.com/v2/gen-ai/models?page=2&per_page=1",
			"last": "https://api.digitalocean.com/v2/gen-ai/models?page=15&per_page=1"
		}
	},
	"meta": {
		"total": 15,
		"page": 1,
		"pages": 15
	}
}
`
var listAPIKeysResponse = `
{
    "api_key_infos": [
        {
            "uuid": "00000000-0000-0000-0000-000000000000",
            "name": "Key One",
            "secret_key": "1000000",
            "created_at": "2025-05-14T13:18:05Z",
            "created_by": "12345678"
        },
        {
            "uuid": "00000000-0000-0000-0000-000000000000",
            "name": "Key Two",
            "secret_key": "1000000",
            "created_at": "2025-05-15T13:18:05Z",
            "created_by": "12345678"
        }
    ]
}
`

var apiKeyInfoResponse = `
{
    "api_key_info": {
        "uuid": "00000000-0000-0000-0000-000000000000",
        "name": "Key One",
        "secret_key": "1000000",
        "created_at": "2025-05-14T13:18:05Z",
        "created_by": "12345678"
    }
}
`

var listKnowledgeBaseResponse = `
{
	"knowledge_bases": [
		{
			"created_at": "2025-05-08T03:37:28Z",
			"database_id": "11111111-1111-1111-1111-111111111111",
			"embedding_model_uuid": "11111111-1111-1111-1111-111111111111",
			"last_indexing_job": {
				"created_at": "2025-05-08T03:37:28Z",
				"finished_at": "2025-05-08T03:38:28Z",
				"knowledge_base_uuid": "11111111-1111-1111-1111-111111111111",
				"phase": "BATCH_JOB_PHASE_SUCCEEDED",
				"started_at": "2025-05-08T03:37:28Z",
				"updated_at": "2025-05-08T03:38:28Z",
				"total_datasources": 1,
				"completed_datasources": 1,
				"tokens": 5000,
				"uuid": "11111111-1111-1111-1111-111111111111"
			},
			"name": "Test Knowledge Base",
			"project_id": "11111111-1111-1111-1111-111111111111",
			"region": "tor1",
			"tags": ["test"],
			"uuid": "11111111-1111-1111-1111-111111111111",
			"updated_at": "2025-05-09T16:02:39Z",
			"is_public": false,
			"user_id": "14567334"
		}
	],
	"links": {
		"pages": {
			"first": "https://api.digitalocean.com/v2/gen-ai/knowledge_bases?page=1&per_page=1",
			"previous": "https://api.digitalocean.com/v2/gen-ai/knowledge_bases?page=1&per_page=1",
			"next": "https://api.digitalocean.com/v2/gen-ai/knowledge_bases?page=3&per_page=1",
			"last": "https://api.digitalocean.com/v2/gen-ai/knowledge_bases?page=10&per_page=1"
		}
	},
	"meta": {
		"total": 10,
		"page": 1,
		"pages": 10
	}
}
`

var knowledgeBaseResponse = `
{
	"knowledge_base": {
		"uuid": "11111111-1111-1111-1111-111111111111",
		"name": "testing-kb",
		"created_at": "2025-05-14T13:18:05Z",
		"updated_at": "2025-05-14T13:18:05Z",
		"region": "tor1",
		"project_id": "11111111-1111-1111-1111-111111111111",
		"embedding_model_uuid": "11111111-1111-1111-1111-111111111111",
		"database_id": "11111111-1111-1111-1111-111111111111",
		"is_public": false,
		"tags": ["string"],
		"user_id": "18919793"
	}
}
`

var knowledgeBaseGetResponse = `
{
	"database_status" : "ONLINE",
	"knowledge_base": {
		"uuid": "11111111-1111-1111-1111-111111111111",
		"name": "testing-kb",
		"created_at": "2025-05-14T13:18:05Z",
		"updated_at": "2025-05-14T13:18:05Z",
		"region": "tor1",
		"project_id": "11111111-1111-1111-1111-111111111111",
		"embedding_model_uuid": "11111111-1111-1111-1111-111111111111",
		"database_id": "11111111-1111-1111-1111-111111111111",
		"is_public": false,
		"tags": ["string"],
		"user_id": "18919793"
	}
}
`

var knowledgeBaseUpdateResponse = `
{
	"knowledge_base": {
		"uuid": "11111111-1111-1111-1111-111111111111",
		"name": "Updated Knowledge Base",
		"created_at": "2025-05-14T13:18:05Z",
		"updated_at": "2025-05-14T13:46:48Z",
		"region": "tor1",
		"project_id": "11111111-1111-1111-1111-111111111111",
		"embedding_model_uuid": "11111111-1111-1111-1111-111111111111",
		"database_id": "11111111-1111-1111-1111-111111111111",
		"is_public": true,
		"tags": ["updated", "example"],
		"user_id": "18919793"
	}
}
`

var listDataSourcesResponse = `
{
	"knowledge_base_data_sources": [
		{
			"uuid": "22222222-2222-2222-2222-222222222222",
			"created_at": "2025-05-14T13:18:05Z",
			"updated_at": "2025-05-14T13:18:05Z",
			"spaces_data_source": {
				"bucket_name": "test-bucket",
				"item_path": "/docs/test.pdf",
				"region": "tor1"
			}
		}
	],
	"links": {
		"pages": {
			"first": "https://api.digitalocean.com/v2/gen-ai/knowledge_bases/11111111-1111-1111-1111-111111111111/data_sources?page=1&per_page=1",
			"previous": "https://api.digitalocean.com/v2/gen-ai/knowledge_bases/11111111-1111-1111-1111-111111111111/data_sources?page=1&per_page=1",
			"next": "https://api.digitalocean.com/v2/gen-ai/knowledge_bases/11111111-1111-1111-1111-111111111111/data_sources?page=2&per_page=1",
			"last": "https://api.digitalocean.com/v2/gen-ai/knowledge_bases/11111111-1111-1111-1111-111111111111/data_sources?page=3&per_page=1"
		}
	},
	"meta": { 
		"total": 3,
		"page": 1,
		"pages": 3
	}
}
`

var addDataSourceResponse = `
{
	"knowledge_base_data_source": {
		"uuid": "22222222-2222-2222-2222-222222222222",
		"created_at": "2025-05-14T13:18:05Z",
		"updated_at": "2025-05-14T13:18:05Z",
		"spaces_data_source": {
			"bucket_name": "test-bucket",
			"item_path": "/docs/test.pdf",
			"region": "tor1"
		}
	}
}
`

var deleteDataSourceResponse = `
{
	"knowledge_base_uuid": "11111111-1111-1111-1111-111111111111",
	"data_source_uuid": "22222222-2222-2222-2222-222222222222"
}
`

var deleteKnowledgeBaseResponse = `
{
	"uuid": "11111111-1111-1111-1111-111111111111"
}`

func TestListAgents(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/gen-ai/agents", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		testFormValues(t, r, values{
			"page":     "1",
			"per_page": "1",
		})

		fmt.Fprint(w, listAgentResponse)
	})

	req := &ListOptions{
		Page:    1,
		PerPage: 1,
	}

	agents, resp, err := client.GenAI.ListAgents(ctx, req)
	if err != nil {
		t.Errorf("GenAI.ListAgents returned error: %v", err)
	}

	var agent Agent
	err = json.Unmarshal([]byte(listAgentResponse), &agent)
	if err != nil {
		t.Errorf("Error unmarshalling response: %v", err)
	}

	assert.Equal(t, 34, resp.Meta.Total)
	expectedString := fmt.Sprintf("%v", agents[0])
	assert.Equal(t, expectedString, agents[0].String())
	assert.Equal(t, agent.K, agents[0].K)
}

func TestCreateAgent(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/gen-ai/agents", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, agentResponse)
	})

	req := &AgentCreateRequest{
		Description:       "My Agent Description",
		Instruction:       "You are an agent who thinks deeply about the world",
		KnowledgeBaseUuid: []string{"00000000-0000-0000-0000-000000000000"},
		ModelUuid:         "00000000-0000-0000-0000-000000000000",
		Name:              "testing-godo",
		ProjectId:         "00000000-0000-0000-0000-000000000000",
		Region:            "tor1",
		Tags:              []string{"string"},
	}

	res, _, err := client.GenAI.CreateAgent(ctx, req)
	if err != nil {
		t.Errorf("GenAI.Create returned error: %v", err)
	}

	assert.Equal(t, res.Name, req.Name)
	assert.Equal(t, res.Description, req.Description)
}

func TestListAPIKeys(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/gen-ai/agents/00000000-0000-0000-0000-000000000000/api_keys", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, listAPIKeysResponse)
	})

	keys, resp, err := client.GenAI.ListAgentAPIKeys(ctx, "00000000-0000-0000-0000-000000000000", nil)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, 2, len(keys))
	assert.Equal(t, "Key One", keys[0].Name)
	assert.Equal(t, "00000000-0000-0000-0000-000000000000", keys[0].Uuid)
	assert.Equal(t, "Key Two", keys[1].Name)
	assert.Equal(t, "00000000-0000-0000-0000-000000000000", keys[1].Uuid)
}

func TestCreateAPIKey(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/gen-ai/agents/00000000-0000-0000-0000-000000000000/api_keys", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, apiKeyInfoResponse)
	})

	req := &AgentAPIKeyCreateRequest{
		AgentUuid: "00000000-0000-0000-0000-000000000000",
		Name:      "Key One",
	}

	key, resp, err := client.GenAI.CreateAgentAPIKey(ctx, "00000000-0000-0000-0000-000000000000", req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.Response.StatusCode)
	assert.Equal(t, "Key One", key.Name)
	assert.Equal(t, "00000000-0000-0000-0000-000000000000", key.Uuid)
}

func TestUpdateAPIKey(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/gen-ai/agents/00000000-0000-0000-0000-000000000000/api_keys/00000000-0000-0000-0000-000000000000", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		fmt.Fprint(w, apiKeyInfoResponse)
	})

	req := &AgentAPIKeyUpdateRequest{
		AgentUuid:  "00000000-0000-0000-0000-000000000000",
		APIKeyUuid: "00000000-0000-0000-0000-000000000000",
		Name:       "Key One",
	}

	key, resp, err := client.GenAI.UpdateAgentAPIKey(ctx, "00000000-0000-0000-0000-000000000000", "00000000-0000-0000-0000-000000000000", req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.Response.StatusCode)
	assert.Equal(t, "Key One", key.Name)
	assert.Equal(t, "00000000-0000-0000-0000-000000000000", key.Uuid)
}

func TestDeleteAPIKey(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/gen-ai/agents/00000000-0000-0000-0000-000000000000/api_keys/00000000-0000-0000-0000-000000000000", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		fmt.Fprint(w, apiKeyInfoResponse)
	})

	key, resp, err := client.GenAI.DeleteAgentAPIKey(ctx, "00000000-0000-0000-0000-000000000000", "00000000-0000-0000-0000-000000000000")
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.Response.StatusCode)
	assert.Equal(t, "Key One", key.Name)
	assert.Equal(t, "00000000-0000-0000-0000-000000000000", key.Uuid)
}

func TestRegenerateAPIKey(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/gen-ai/agents/00000000-0000-0000-0000-000000000000/api_keys/00000000-0000-0000-0000-000000000000/regenerate", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		fmt.Fprint(w, apiKeyInfoResponse)
	})

	key, resp, err := client.GenAI.RegenerateAgentAPIKey(ctx, "00000000-0000-0000-0000-000000000000", "00000000-0000-0000-0000-000000000000")
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.Response.StatusCode)
	assert.Equal(t, "Key One", key.Name)
	assert.Equal(t, "00000000-0000-0000-0000-000000000000", key.Uuid)
}

func TestGetAgent(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/gen-ai/agents/00000000-0000-0000-0000-000000000000", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, agentResponse)
	})

	res, resp, err := client.GenAI.GetAgent(ctx, "00000000-0000-0000-0000-000000000000")
	if err != nil {
		t.Errorf("GenAI.Get returned error: %v", err)
	}

	assert.Equal(t, res.Name, "testing-godo")
	assert.Equal(t, resp.Response.StatusCode, 200)
	expectedString := fmt.Sprintf("%v", res)
	assert.Equal(t, expectedString, res.String())
}

func TestDeleteAgent(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/gen-ai/agents/def5d52c-30c5-11f0-bf8f-4e013e2ddde4", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		fmt.Fprint(w, agentResponse)
	})

	res, resp, err := client.GenAI.DeleteAgent(ctx, "def5d52c-30c5-11f0-bf8f-4e013e2ddde4")
	if err != nil {
		t.Errorf("GenAI.Delete returned error: %v", err)
	}

	assert.Equal(t, resp.Response.StatusCode, 200)
	assert.Equal(t, "testing-godo", res.Name)
}

func TestUpdateAgent(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/gen-ai/agents/00000000-0000-0000-0000-000000000000", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		fmt.Fprint(w, agentUpdateResponse)
	})

	req := &AgentUpdateRequest{
		Tags: []string{"updated", "example"},
	}

	res, resp, err := client.GenAI.UpdateAgent(ctx, "00000000-0000-0000-0000-000000000000", req)
	if err != nil {
		t.Errorf("GenAI.Update returned error: %v", err)
	}

	assert.Equal(t, res.Tags[0], req.Tags[0])
	assert.Equal(t, resp.Response.StatusCode, 200)
}

func TestUpdateAgentVisibility(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/gen-ai/agents/00000000-0000-0000-0000-000000000000/deployment_visibility", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		fmt.Fprint(w, agentUpdateResponse)
	})

	req := &AgentVisibilityUpdateRequest{
		Uuid:       "00000000-0000-0000-0000-000000000000",
		Visibility: "VISIBILITY_PRIVATE",
	}

	res, resp, err := client.GenAI.UpdateAgentVisibility(ctx, "00000000-0000-0000-0000-000000000000", req)
	if err != nil {
		t.Errorf("GenAI.UpdateVisibility returned error: %v", err)
	}

	assert.Equal(t, res.Uuid, req.Uuid)
	assert.Equal(t, res.Name, "My Agent")
	assert.Equal(t, resp.Response.StatusCode, 200)
}

func TestListModels(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/gen-ai/models", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		testFormValues(t, r, values{
			"page":     "1",
			"per_page": "1",
		})

		fmt.Fprint(w, agentModelsResponse)
	})

	req := &ListOptions{
		Page:    1,
		PerPage: 1,
	}

	models, resp, err := client.GenAI.ListModels(ctx, req)
	if err != nil {
		t.Errorf("GenAI ListModels returned error: %v", err)
	}

	assert.Equal(t, models[0].Name, "Llama 3.3 Instruct (70B)")
	expectedString := fmt.Sprintf("%v", models[0])
	assert.Equal(t, resp.Response.StatusCode, 200)
	assert.Equal(t, expectedString, models[0].String())
}

func TestListKnowledgeBases(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/gen-ai/knowledge_bases", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		testFormValues(t, r, values{
			"page":     "1",
			"per_page": "1",
		})

		fmt.Fprint(w, listKnowledgeBaseResponse)
	})

	req := &ListOptions{
		Page:    1,
		PerPage: 1,
	}

	knowledgeBases, resp, err := client.GenAI.ListKnowledgeBases(ctx, req)
	if err != nil {
		t.Errorf("GenAI.ListKnowledgeBases returned error: %v", err)
	}

	assert.Equal(t, 10, resp.Meta.Total)
	assert.Equal(t, "Test Knowledge Base", knowledgeBases[0].Name)
}

func TestCreateKnowledgeBase(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/gen-ai/knowledge_bases", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, knowledgeBaseResponse)
	})

	req := &KnowledgeBaseCreateRequest{
		Name:               "testing-kb",
		ProjectID:          "11111111-1111-1111-1111-111111111111",
		Region:             "tor1",
		EmbeddingModelUuid: "11111111-1111-1111-1111-111111111111",
		Tags:               []string{"string"},
	}

	res, _, err := client.GenAI.CreateKnowledgeBase(ctx, req)
	if err != nil {
		t.Errorf("GenAI.CreateKnowledgeBase returned error: %v", err)
	}

	assert.Equal(t, res.Name, req.Name)
	assert.Equal(t, res.ProjectId, req.ProjectID)
}

func TestListDataSources(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/gen-ai/knowledge_bases/11111111-1111-1111-1111-111111111111/data_sources", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		testFormValues(t, r, values{
			"page":     "1",
			"per_page": "1",
		})

		fmt.Fprint(w, listDataSourcesResponse)
	})

	req := &ListOptions{
		Page:    1,
		PerPage: 1,
	}

	dataSources, resp, err := client.GenAI.ListDataSources(ctx, "11111111-1111-1111-1111-111111111111", req)
	if err != nil {
		t.Errorf("GenAI.ListDataSources returned error: %v", err)
	}

	assert.Equal(t, 3, resp.Meta.Total)
	assert.Equal(t, "22222222-2222-2222-2222-222222222222", dataSources[0].Uuid)
}

func TestAddDataSource(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/gen-ai/knowledge_bases/11111111-1111-1111-1111-111111111111/data_sources", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, addDataSourceResponse)
	})

	req := &AddDataSourceRequest{
		KnowledgeBaseUuid: "11111111-1111-1111-1111-111111111111",
		SpacesDataSource: &SpacesDataSource{
			BucketName: "test-bucket",
			ItemPath:   "/docs/test.pdf",
			Region:     "tor1",
		},
	}

	res, _, err := client.GenAI.AddDataSource(ctx, "11111111-1111-1111-1111-111111111111", req)
	if err != nil {
		t.Errorf("GenAI.AddDataSource returned error: %v", err)
	}

	assert.Equal(t, "test-bucket", res.SpacesDataSource.BucketName)
	assert.Equal(t, "/docs/test.pdf", res.SpacesDataSource.ItemPath)
}

func TestDeleteDataSource(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/gen-ai/knowledge_bases/11111111-1111-1111-1111-111111111111/data_sources/22222222-2222-2222-2222-222222222222", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		fmt.Fprint(w, deleteDataSourceResponse)
	})

	kbUUID, dsUUID, resp, err := client.GenAI.DeleteDataSource(ctx, "11111111-1111-1111-1111-111111111111", "22222222-2222-2222-2222-222222222222")
	if err != nil {
		t.Errorf("GenAI.DeleteDataSource returned error: %v", err)
	}

	assert.Equal(t, "11111111-1111-1111-1111-111111111111", kbUUID)
	assert.Equal(t, "22222222-2222-2222-2222-222222222222", dsUUID)
	assert.Equal(t, 200, resp.Response.StatusCode)
}

func TestGetKnowledgeBase(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/gen-ai/knowledge_bases/11111111-1111-1111-1111-111111111111", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, knowledgeBaseGetResponse)
	})

	res, dbStatus, resp, err := client.GenAI.GetKnowledgeBase(ctx, "11111111-1111-1111-1111-111111111111")
	if err != nil {
		t.Errorf("GenAI.GetKnowledgeBase returned error: %v", err)
	}

	assert.Equal(t, "testing-kb", res.Name)
	assert.Equal(t, "ONLINE", dbStatus)
	assert.Equal(t, "11111111-1111-1111-1111-111111111111", res.Uuid)
	assert.Equal(t, 200, resp.Response.StatusCode)
}

func TestUpdateKnowledgeBase(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/gen-ai/knowledge_bases/11111111-1111-1111-1111-111111111111", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		fmt.Fprint(w, knowledgeBaseUpdateResponse)
	})

	req := &UpdateKnowledgeBaseRequest{
		Name: "Updated Knowledge Base",
		Tags: []string{"updated", "example"},
	}

	res, resp, err := client.GenAI.UpdateKnowledgeBase(ctx, "11111111-1111-1111-1111-111111111111", req)
	if err != nil {
		t.Errorf("GenAI.UpdateKnowledgeBase returned error: %v", err)
	}

	assert.Equal(t, res.Tags[0], req.Tags[0])
	assert.Equal(t, 200, resp.Response.StatusCode)
}

func TestDeleteKnowledgeBase(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/gen-ai/knowledge_bases/11111111-1111-1111-1111-111111111111", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		fmt.Fprint(w, deleteKnowledgeBaseResponse)
	})

	kbUUID, resp, err := client.GenAI.DeleteKnowledgeBase(ctx, "11111111-1111-1111-1111-111111111111")
	if err != nil {
		t.Errorf("GenAI.DeleteKnowledgeBase returned error: %v", err)
	}

	assert.Equal(t, "11111111-1111-1111-1111-111111111111", kbUUID)
	assert.Equal(t, 200, resp.Response.StatusCode)
}

func TestAttachKnowledgeBase(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/gen-ai/agents/00000000-0000-0000-0000-000000000000/knowledge_bases/11111111-1111-1111-1111-111111111111", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, agentResponse)
	})

	res, resp, err := client.GenAI.AttachKnowledgeBase(ctx, "00000000-0000-0000-0000-000000000000", "11111111-1111-1111-1111-111111111111")
	if err != nil {
		t.Errorf("GenAI.AttachKnowledgBase returned error: %v", err)
	}

	assert.Equal(t, "testing-godo", res.Name)
	assert.Equal(t, 200, resp.Response.StatusCode)
}

func TestDetachKnowledgeBase(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/gen-ai/agents/00000000-0000-0000-0000-000000000000/knowledge_bases/11111111-1111-1111-1111-111111111111", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		fmt.Fprint(w, agentResponse)
	})

	res, resp, err := client.GenAI.DetachKnowledgeBase(ctx, "00000000-0000-0000-0000-000000000000", "11111111-1111-1111-1111-111111111111")
	fmt.Print(res)
	fmt.Print(resp)

	if err != nil {
		t.Fatalf("GenAI.DetachKnowledgBase returned error: %v", err)
	}

	if res == nil {
		t.Fatalf("GenAI.DetachKnowledgBase returned nil response or result")
	}

	assert.Equal(t, "testing-godo", res.Name)
	assert.Equal(t, 200, resp.Response.StatusCode)
}
