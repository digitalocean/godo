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
