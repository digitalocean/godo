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
		"retrieval_method": "RETRIEVAL_METHOD_NONE",
		"functions": [
        {
            "name": "godo-test-function",
            "description": "Creating Function Route",
            "faas_name": "godo-test-faasname",
            "faas_namespace": "fn-00000000-0000-0000-0000-000000000000",
            "input_schema": {
                "parameters": [
                    {
                        "name": "zipCode",
                        "in": "query",
                        "schema": {
                            "type": "string"
                        },
                        "required": false,
                        "description": "The ZIP code for which to fetch the weather"
                    },
                    {
                        "name": "measurement",
                        "in": "query",
                        "schema": {
                            "type": "string",
                            "enum": [
                                "F",
                                "C"
                            ]
                        },
                        "required": false,
                        "description": "The measurement unit for temperature (F or C)"
                    }
                ]
            },
            "output_schema": {
                "properties": [
                    {
                        "name": "temperature",
                        "type": "number",
                        "description": "The temperature for the specified location"
                    },
                    {
                        "name": "measurement",
                        "type": "string",
                        "description": "The measurement unit used for the temperature (F or C)"
                    },
                    {
                        "name": "conditions",
                        "type": "string",
                        "description": "A description of the current weather conditions (Sunny, Cloudy, etc)"
                    }
                ]
            }
        }
    ]
		
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
		"retrieval_method": "RETRIEVAL_METHOD_NONE",
		"functions": [
        {
            "name": "godo-test-function",
            "description": "Updating Function Route",
            "faas_name": "godo-test-faasname",
            "faas_namespace": "fn-00000000-0000-0000-0000-000000000000",
            "input_schema": {
                "parameters": [
                    {
                        "name": "zipCode",
                        "in": "query",
                        "schema": {
                            "type": "string"
                        },
                        "required": false,
                        "description": "The ZIP code for which to fetch the weather"
                    },
                    {
                        "name": "measurement",
                        "in": "query",
                        "schema": {
                            "type": "string",
                            "enum": [
                                "F",
                                "C"
                            ]
                        },
                        "required": false,
                        "description": "The measurement unit for temperature (F or C)"
                    }
                ]
            },
            "output_schema": {
                "properties": [
                    {
                        "name": "temperature",
                        "type": "number",
                        "description": "The temperature for the specified location"
                    },
                    {
                        "name": "measurement",
                        "type": "string",
                        "description": "The measurement unit used for the temperature (F or C)"
                    },
                    {
                        "name": "conditions",
                        "type": "string",
                        "description": "A description of the current weather conditions (Sunny, Cloudy, etc)"
                    }
                ]
            }
        }
    ]
	}
}
`

var listAvailableModelsResponse = `
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

var listIndexingJobsResponse = `
{
	"jobs": [
		{
			"uuid": "22222222-2222-2222-2222-222222222222",
			"knowledge_base_uuid": "11111111-1111-1111-1111-111111111111",
			"created_at": "2025-05-08T03:37:28Z",
			"started_at": "2025-05-08T03:37:30Z",
			"finished_at": "2025-05-08T03:38:28Z",
			"updated_at": "2025-05-08T03:38:28Z",
			"phase": "BATCH_JOB_PHASE_SUCCEEDED",
			"total_datasources": 2,
			"completed_datasources": 2,
			"tokens": 10500,
			"data_source_uuids": ["33333333-3333-3333-3333-333333333333", "44444444-4444-4444-4444-444444444444"]
		},
		{
			"uuid": "55555555-5555-5555-5555-555555555555",
			"knowledge_base_uuid": "66666666-6666-6666-6666-666666666666",
			"created_at": "2025-05-07T02:30:15Z",
			"started_at": "2025-05-07T02:30:20Z",
			"finished_at": "2025-05-07T02:32:45Z",
			"updated_at": "2025-05-07T02:32:45Z",
			"phase": "BATCH_JOB_PHASE_SUCCEEDED",
			"total_datasources": 1,
			"completed_datasources": 1,
			"tokens": 7500,
			"data_source_uuids": ["77777777-7777-7777-7777-777777777777"]
		}
	],
	"links": {
		"pages": {
			"first": "https://api.digitalocean.com/v2/gen-ai/indexing_jobs?page=1&per_page=2",
			"next": "https://api.digitalocean.com/v2/gen-ai/indexing_jobs?page=2&per_page=2",
			"last": "https://api.digitalocean.com/v2/gen-ai/indexing_jobs?page=5&per_page=2"
		}
	},
	"meta": {
		"total": 9,
		"page": 1,
		"pages": 5
	}
}
`

var indexingJobDataSourcesResponse = `
{
	"indexed_data_sources": [
		{
			"completed_at": "2025-05-08T03:38:28Z",
			"data_source_uuid": "33333333-3333-3333-3333-333333333333",
			"error_details": "",
			"error_msg": "",
			"failed_item_count": "0",
			"indexed_file_count": "5",
			"indexed_item_count": "150",
			"removed_item_count": "0",
			"skipped_item_count": "2",
			"started_at": "2025-05-08T03:37:30Z",
			"status": "DATA_SOURCE_STATUS_COMPLETED",
			"total_bytes": "2048000",
			"total_bytes_indexed": "1900000",
			"total_file_count": "7"
		},
		{
			"completed_at": "2025-05-08T03:38:25Z",
			"data_source_uuid": "44444444-4444-4444-4444-444444444444",
			"error_details": "File format not supported",
			"error_msg": "Some files could not be processed",
			"failed_item_count": "3",
			"indexed_file_count": "8",
			"indexed_item_count": "200",
			"removed_item_count": "1",
			"skipped_item_count": "5",
			"started_at": "2025-05-08T03:37:32Z",
			"status": "DATA_SOURCE_STATUS_COMPLETED_WITH_ERRORS",
			"total_bytes": "3072000",
			"total_bytes_indexed": "2800000",
		"total_file_count": "12"
	}
]
}
`

var indexingJobResponse = `
{
	"job": {
		"uuid": "22222222-2222-2222-2222-222222222222",
		"knowledge_base_uuid": "11111111-1111-1111-1111-111111111111",
		"status": "INDEX_JOB_STATUS_COMPLETED",
		"phase": "BATCH_JOB_PHASE_COMPLETED",
		"created_at": "2025-05-08T03:30:00Z",
		"started_at": "2025-05-08T03:30:05Z",
		"finished_at": "2025-05-08T03:37:45Z",
		"updated_at": "2025-05-08T03:37:45Z",
		"data_source_uuids": [
			"33333333-3333-3333-3333-333333333333",
			"44444444-4444-4444-4444-444444444444"
		],
		"total_datasources": 2,
		"completed_datasources": 2,
		"tokens": 1250,
		"total_items_indexed": "350",
		"total_items_failed": "3",
		"total_items_skipped": "7"
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
		"name": "Updated-Knowledge-Base",
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

var agentRouteResponse = `
{
    "child_agent_uuid": "00000000-0000-0000-0000-000000000001",
    "parent_agent_uuid": "00000000-0000-0000-0000-000000000000",
    "rollback": false,
    "uuid": "00000000-0000-0000-0000-000000000003"
}
`

var listAgentVersionsResponse = `
{
	"agent_versions": [
		{
			"agent_uuid": "00000000-0000-0000-0000-000000000000",
			"currently_applied": true,
			"version_hash": "00000000000000000000000000000000000000000000000000000000000000",
			"created_at": "2025-06-25T09:32:26Z",
			"created_by_email": "example@gmail.com",
			"trigger_action": "Rolled back from version ABCDEDF",
			"model_name": "Llama 3.3 Instruct (70B)",
			"name": "normal-agent",
			"instruction": "You are an agent who thinks deeply about the world",
			"description": "Think about the world deeply",
			"tags": [
				"example string"
			],
			"k": 10,
			"temperature": 0.7,
			"top_p": 0.9,
			"max_tokens": 512,
			"retrieval_method": "RETRIEVAL_METHOD_NONE",
			"attached_functions": [
				{
					"name": "terraform-tf",
					"faas_name": "default/testing",
					"faas_namespace": "fn-b90f0000-0000-0000-0000-75edfbb6f397",
					"is_deleted": true
				}
			],
			"can_rollback": true
		}
	],
	"links": {
		"pages": {
			"first": "https://api.digitalocean.com/v2/gen-ai/agents/01efde4/versions?page=1&per_page=5",
			"next": "https://api.digitalocean.com/v2/gen-ai/agents/01efde4/versions?page=2&per_page=5",
			"last": "https://api.digitalocean.com/v2/gen-ai/agents/01efde4/versions?page=13&per_page=5"
		}
	},
	"meta": {
		"total": 65,
		"page": 1,
		"pages": 13
	}
}
`
var listAnthropicAPIKeysResponse = `
{
    "api_key_infos": [
        {
            "uuid": "11111111-1111-1111-1111-111111111111",
            "name": "Anthropic Key One",
             "api_key": "sk-ant-1",
            "created_at": "2025-05-14T13:18:05Z",
            "created_by": "user-1"
        },
        {
            "uuid": "22222222-2222-2222-2222-222222222222",
            "name": "Anthropic Key Two",
             "api_key": "sk-ant-2",
            "created_at": "2025-05-15T13:18:05Z",
            "created_by": "user-2"
        }
    ],
    "links": {
        "pages": {
            "first": "https://api.digitalocean.com/v2/gen-ai/anthropic/keys?page=1&per_page=1",
            "next": "https://api.digitalocean.com/v2/gen-ai/anthropic/keys?page=2&per_page=1",
            "last": "https://api.digitalocean.com/v2/gen-ai/anthropic/keys?page=10&per_page=1"
        }
    },
    "meta": {
        "total": 2,
        "page": 1,
        "pages": 10
    }
}
`

var anthropicAPIKeyInfoResponse = `
{
    "api_key_info": {
        "uuid": "11111111-1111-1111-1111-111111111111",
        "name": "Anthropic Key One",
        "api_key": "sk-ant-1",
        "created_at": "2025-05-14T13:18:05Z",
        "created_by": "user-1"
    }
}
`
var listAgentsByAnthropicAPIKeyResponse = `
{
  "agents": [
    {
      "uuid": "00000000-0000-0000-0000-000000000000",
      "name": "Anthropic Agent 1"
    },
    {
      "uuid": "00000000-0000-0000-0000-000000000001",
      "name": "Anthropic Agent 2"
    }
  ],
  "links": {
    "pages": {
      "first": "https://api.digitalocean.com/v2/gen-ai/anthropic/keys?page=1&per_page=1"
    }
  },
  "meta": {
    "total": 2,
    "page": 1,
    "pages": 1
  }
}
`

var rollbackResponse = `
{
	"version_hash": "00000000000000000000000000000000000000000000000000000000000001"
}
`

var listOpenAIAPIKeysResponse = `
{
    "api_key_infos": [
        {
            "uuid": "11111111-1111-1111-1111-111111111111",
            "created_at": "2025-05-14T13:18:05Z",
            "created_by": "user-1",
			"name": "openai-key1",
			"models": [{
				"agreement": {
					"description": "openai example",
					"name": "example name",
					"url": "example string",
					"uuid": "123e4567-e89b-12d3-a456-426614174000"
				},
				"created_at": "2023-01-01T00:00:00Z",
				"inference_name": "example name",
				"inference_version": "example string",
				"is_foundational": true,
				"metadata": { },
				"name": "example name",
				"parent_uuid": "123e4567-e89b-12d3-a456-426614174000",
				"provider": "MODEL_PROVIDER_DIGITALOCEAN",
				"updated_at": "2023-01-01T00:00:00Z",
				"upload_complete": true,
				"url": "example string",
				"usecases": ["MODEL_USECASE_AGENT","MODEL_USECASE_GUARDRAIL"],
				"uuid": "123e4567-e89b-12d3-a456-426614174000",
				"version": {
					"major": 123,
					"minor": 123,
					"patch": 123
				}
			}]
        },
		 {
            "uuid": "22222222-2222-2222-2222-222222222222",
            "created_at": "2025-05-14T13:18:05Z",
            "created_by": "user-2",
			"name": "openai-key2",
			"models": [{
				"agreement": {
					"description": "openai example",
					"name": "example name",
					"url": "example string",
					"uuid": "123e4567-e89b-12d3-a456-426614174000"
				},
				"created_at": "2023-01-01T00:00:00Z",
				"inference_name": "example name",
				"inference_version": "example string",
				"is_foundational": true,
				"metadata": { },
				"name": "example name",
				"parent_uuid": "123e4567-e89b-12d3-a456-426614174000",
				"provider": "MODEL_PROVIDER_DIGITALOCEAN",
				"updated_at": "2023-01-01T00:00:00Z",
				"upload_complete": true,
				"url": "example string",
				"usecases": ["MODEL_USECASE_AGENT","MODEL_USECASE_GUARDRAIL"],
				"uuid": "123e4567-e89b-12d3-a456-426614174000",
				"version": {
					"major": 123,
					"minor": 123,
					"patch": 123
				}
			}]
        }
    ],
    "links": {
        "pages": {
            "first": "https://api.digitalocean.com/v2/gen-ai/openai/keys?page=1&per_page=1",
            "next": "https://api.digitalocean.com/v2/gen-ai/openai/keys?page=2&per_page=1",
            "last": "https://api.digitalocean.com/v2/gen-ai/openai/keys?page=10&per_page=1"
        }
    },
    "meta": {
        "total": 2,
        "page": 1,
        "pages": 10
    }
}
`

var openaiAPIKeyInfoResponse = `
{
    "api_key_info": {
            "uuid": "11111111-1111-1111-1111-111111111111",
            "created_at": "2025-05-14T13:18:05Z",
            "created_by": "user-1",
			"name": "OpenAI One",
			"models": [{
				"agreement": {
					"description": "openai example",
					"name": "example name",
					"url": "example string",
					"uuid": "123e4567-e89b-12d3-a456-426614174000"
				},
				"created_at": "2023-01-01T00:00:00Z",
				"inference_name": "example name",
				"inference_version": "example string",
				"is_foundational": true,
				"metadata": { },
				"name": "example name",
				"parent_uuid": "123e4567-e89b-12d3-a456-426614174000",
				"provider": "MODEL_PROVIDER_DIGITALOCEAN",
				"updated_at": "2023-01-01T00:00:00Z",
				"upload_complete": true,
				"url": "example string",
				"usecases": ["MODEL_USECASE_AGENT","MODEL_USECASE_GUARDRAIL"],
				"uuid": "123e4567-e89b-12d3-a456-426614174000",
				"version": {
					"major": 123,
					"minor": 123,
					"patch": 123
				}
			}]
        }
}
`

var listAgentsByOpenAIAPIKeyResponse = `
{
  "agents": [
    {
      "uuid": "00000000-0000-0000-0000-000000000000",
      "name": "OpenAI Agent 1"
    },
    {
      "uuid": "00000000-0000-0000-0000-000000000001",
      "name": "OpenAI Agent 2"
    }
  ],
  "links": {
    "pages": {
      "first": "https://api.digitalocean.com/v2/gen-ai/openai/keys?page=1&per_page=1"
    }
  },
  "meta": {
    "total": 2,
    "page": 1,
    "pages": 1
  }
}
`

var listDatacenterRegionsResponse = `
{
	"regions": [
		{
			"region": "tor1",
			"inference_url": "https://tor1.gen-ai.digitalocean.com",
			"serves_batch": true,
			"serves_inference": true,
			"stream_inference_url": "https://tor1.gen-ai.digitalocean.com/stream"
		},
		{
			"region": "nyc3",
			"inference_url": "https://nyc3.gen-ai.digitalocean.com",
			"serves_batch": false,
			"serves_inference": true,
			"stream_inference_url": "https://nyc3.gen-ai.digitalocean.com/stream"
		}
	]
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

func TestListIndexingJobs(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/gen-ai/indexing_jobs", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		testFormValues(t, r, values{
			"page":     "1",
			"per_page": "2",
		})

		fmt.Fprint(w, listIndexingJobsResponse)
	})

	req := &ListOptions{
		Page:    1,
		PerPage: 2,
	}

	result, resp, err := client.GenAI.ListIndexingJobs(ctx, req)
	if err != nil {
		t.Errorf("GenAI.ListIndexingJobs returned error: %v", err)
	}

	assert.Equal(t, 9, resp.Meta.Total)
	assert.Equal(t, 2, len(result.Jobs))
	assert.Equal(t, "22222222-2222-2222-2222-222222222222", result.Jobs[0].Uuid)
	assert.Equal(t, "11111111-1111-1111-1111-111111111111", result.Jobs[0].KnowledgeBaseUuid)
	assert.Equal(t, "BATCH_JOB_PHASE_SUCCEEDED", result.Jobs[0].Phase)
	assert.Equal(t, 2, result.Jobs[0].TotalDatasources)
	assert.Equal(t, 2, result.Jobs[0].CompletedDatasources)
	assert.Equal(t, 10500, result.Jobs[0].Tokens)
	assert.Equal(t, 2, len(result.Jobs[0].DataSourceUuids))
	assert.Equal(t, "55555555-5555-5555-5555-555555555555", result.Jobs[1].Uuid)
}

func TestListIndexingJobDataSources(t *testing.T) {
	setup()
	defer teardown()

	indexingJobUUID := "22222222-2222-2222-2222-222222222222"

	mux.HandleFunc(fmt.Sprintf("/v2/gen-ai/indexing_jobs/%s/data_sources", indexingJobUUID), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, indexingJobDataSourcesResponse)
	})

	result, resp, err := client.GenAI.ListIndexingJobDataSources(ctx, indexingJobUUID)
	if err != nil {
		t.Errorf("GenAI.ListIndexingJobDataSources returned error: %v", err)
	}

	assert.Equal(t, 2, len(result.IndexedDataSources))

	// Test first data source
	ds1 := result.IndexedDataSources[0]
	assert.Equal(t, "33333333-3333-3333-3333-333333333333", ds1.DataSourceUuid)
	assert.Equal(t, "DATA_SOURCE_STATUS_COMPLETED", ds1.Status)
	assert.Equal(t, "5", ds1.IndexedFileCount)
	assert.Equal(t, "150", ds1.IndexedItemCount)
	assert.Equal(t, "0", ds1.FailedItemCount)
	assert.Equal(t, "2", ds1.SkippedItemCount)
	assert.Equal(t, "2048000", ds1.TotalBytes)
	assert.Equal(t, "1900000", ds1.TotalBytesIndexed)
	assert.Equal(t, "7", ds1.TotalFileCount)

	// Test second data source with errors
	ds2 := result.IndexedDataSources[1]
	assert.Equal(t, "44444444-4444-4444-4444-444444444444", ds2.DataSourceUuid)
	assert.Equal(t, "DATA_SOURCE_STATUS_COMPLETED_WITH_ERRORS", ds2.Status)
	assert.Equal(t, "8", ds2.IndexedFileCount)
	assert.Equal(t, "200", ds2.IndexedItemCount)
	assert.Equal(t, "3", ds2.FailedItemCount)
	assert.Equal(t, "5", ds2.SkippedItemCount)
	assert.Equal(t, "File format not supported", ds2.ErrorDetails)
	assert.Equal(t, "Some files could not be processed", ds2.ErrorMsg)

	_ = resp // Mark as used to avoid unused variable warning
}

func TestGetIndexingJob(t *testing.T) {
	setup()
	defer teardown()

	indexingJobUUID := "22222222-2222-2222-2222-222222222222"

	mux.HandleFunc(fmt.Sprintf("/v2/gen-ai/indexing_jobs/%s", indexingJobUUID), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, indexingJobResponse)
	})

	result, resp, err := client.GenAI.GetIndexingJob(ctx, indexingJobUUID)
	if err != nil {
		t.Errorf("GenAI.GetIndexingJob returned error: %v", err)
	}

	// Test the job details
	job := result.Job
	assert.Equal(t, "22222222-2222-2222-2222-222222222222", job.Uuid)
	assert.Equal(t, "11111111-1111-1111-1111-111111111111", job.KnowledgeBaseUuid)
	assert.Equal(t, "INDEX_JOB_STATUS_COMPLETED", job.Status)
	assert.Equal(t, "BATCH_JOB_PHASE_COMPLETED", job.Phase)
	assert.Equal(t, 2, job.TotalDatasources)
	assert.Equal(t, 2, job.CompletedDatasources)
	assert.Equal(t, 1250, job.Tokens)
	assert.Equal(t, "350", job.TotalItemsIndexed)
	assert.Equal(t, "3", job.TotalItemsFailed)
	assert.Equal(t, "7", job.TotalItemsSkipped)

	// Test data source UUIDs array
	assert.Equal(t, 2, len(job.DataSourceUuids))
	assert.Equal(t, "33333333-3333-3333-3333-333333333333", job.DataSourceUuids[0])
	assert.Equal(t, "44444444-4444-4444-4444-444444444444", job.DataSourceUuids[1])

	// Test timestamps
	assert.NotNil(t, job.CreatedAt)
	assert.NotNil(t, job.StartedAt)
	assert.NotNil(t, job.FinishedAt)
	assert.NotNil(t, job.UpdatedAt)

	_ = resp // Mark as used to avoid unused variable warning
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
		DataSources: []KnowledgeBaseDataSource{
			{
				WebCrawlerDataSource: &WebCrawlerDataSource{
					BaseUrl: "https://www.example.com",
				},
			},
		},
	}

	res, _, err := client.GenAI.CreateKnowledgeBase(ctx, req)
	if err != nil {
		t.Errorf("GenAI.CreateKnowledgeBase returned error: %v", err)
	}

	assert.Equal(t, res.Name, req.Name)
	assert.Equal(t, res.ProjectId, req.ProjectID)
	assert.Equal(t, req.EmbeddingModelUuid, res.EmbeddingModelUuid)
	assert.Equal(t, req.Region, res.Region)
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

	dataSources, resp, err := client.GenAI.ListKnowledgeBaseDataSources(ctx, "11111111-1111-1111-1111-111111111111", req)
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

	req := &AddKnowledgeBaseDataSourceRequest{
		KnowledgeBaseUuid: "11111111-1111-1111-1111-111111111111",
		SpacesDataSource: &SpacesDataSource{
			BucketName: "test-bucket",
			ItemPath:   "/docs/test.pdf",
			Region:     "tor1",
		},
	}

	res, _, err := client.GenAI.AddKnowledgeBaseDataSource(ctx, "11111111-1111-1111-1111-111111111111", req)
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

	kbUUID, dsUUID, resp, err := client.GenAI.DeleteKnowledgeBaseDataSource(ctx, "11111111-1111-1111-1111-111111111111", "22222222-2222-2222-2222-222222222222")
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
		Name: "Updated-Knowledge-Base",
		Tags: []string{"updated", "example"},
	}

	res, resp, err := client.GenAI.UpdateKnowledgeBase(ctx, "11111111-1111-1111-1111-111111111111", req)
	if err != nil {
		t.Errorf("GenAI.UpdateKnowledgeBase returned error: %v", err)
	}

	assert.Equal(t, req.Name, res.Name)
	assert.Equal(t, req.Tags, res.Tags)
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

	res, resp, err := client.GenAI.AttachKnowledgeBaseToAgent(ctx, "00000000-0000-0000-0000-000000000000", "11111111-1111-1111-1111-111111111111")
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

	res, resp, err := client.GenAI.DetachKnowledgeBaseToAgent(ctx, "00000000-0000-0000-0000-000000000000", "11111111-1111-1111-1111-111111111111")
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

func TestAddAgentRoute(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/gen-ai/agents/00000000-0000-0000-0000-000000000000/child_agents/00000000-0000-0000-0000-000000000001", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, agentRouteResponse)
	})

	req := &AgentRouteCreateRequest{
		ChildAgentUuid:  "00000000-0000-0000-0000-000000000001",
		IfCase:          "use this to get weather information",
		ParentAgentUuid: "00000000-0000-0000-0000-000000000000",
		RouteName:       "weather route app",
	}

	res, resp, err := client.GenAI.AddAgentRoute(ctx, "00000000-0000-0000-0000-000000000000", "00000000-0000-0000-0000-000000000001", req)
	if err != nil {
		t.Errorf("GenAI.AddAgentRoute returned error: %v", err)
	}
	fmt.Println(res)
	assert.Equal(t, res.ChildAgentUuid, "00000000-0000-0000-0000-000000000001")
	assert.Equal(t, resp.Response.StatusCode, 200)
}

func TestDeleteAgentRoute(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/gen-ai/agents/00000000-0000-0000-0000-000000000000/child_agents/00000000-0000-0000-0000-000000000001", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		fmt.Fprint(w, agentRouteResponse)
	})

	res, resp, err := client.GenAI.DeleteAgentRoute(ctx, "00000000-0000-0000-0000-000000000000", "00000000-0000-0000-0000-000000000001")
	if err != nil {
		t.Errorf("GenAI.DeleteAgentRoute returned error: %v", err)
	}

	assert.Equal(t, res.ChildAgentUuid, "00000000-0000-0000-0000-000000000001")
	assert.Equal(t, resp.Response.StatusCode, 200)
}

func TestUpdateAgentRoute(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/gen-ai/agents/00000000-0000-0000-0000-000000000000/child_agents/00000000-0000-0000-0000-000000000001", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		fmt.Fprint(w, agentRouteResponse)
	})

	req := &AgentRouteUpdateRequest{
		ChildAgentUuid:  "00000000-0000-0000-0000-000000000001",
		IfCase:          "use this to get weather information",
		ParentAgentUuid: "00000000-0000-0000-0000-000000000000",
		RouteName:       "weather route app",
	}

	res, resp, err := client.GenAI.UpdateAgentRoute(ctx, "00000000-0000-0000-0000-000000000000", "00000000-0000-0000-0000-000000000001", req)
	if err != nil {
		t.Errorf("GenAI.UpdateAgentRoute returned error: %v", err)
	}

	assert.Equal(t, res.ChildAgentUuid, "00000000-0000-0000-0000-000000000001")
	assert.Equal(t, resp.Response.StatusCode, 200)
}

func TestListVersions(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/gen-ai/agents/00000000-0000-0000-0000-000000000000/versions", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, listAgentVersionsResponse)
	})

	versions, resp, err := client.GenAI.ListAgentVersions(ctx, "00000000-0000-0000-0000-000000000000", nil)
	if err != nil {
		t.Errorf("GenAI.ListAgentVersions returned error: %v", err)
	}
	fmt.Println(versions)
	assert.Equal(t, 1, len(versions))
	assert.Equal(t, "00000000000000000000000000000000000000000000000000000000000000", versions[0].VersionHash)
	assert.Equal(t, 200, resp.Response.StatusCode)
}

func TestRollbackVersion(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/gen-ai/agents/00000000-0000-0000-0000-000000000000/versions", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		fmt.Fprint(w, rollbackResponse)
	})

	versions, resp, err := client.GenAI.RollbackAgentVersion(ctx, "00000000-0000-0000-0000-000000000000", "00000000000000000000000000000000000000000000000000000000000000")
	if err != nil {
		t.Errorf("GenAI.RollbackVersion returned error: %v", err)
	}

	assert.Equal(t, "00000000000000000000000000000000000000000000000000000000000001", versions)
	assert.Equal(t, 200, resp.Response.StatusCode)
}

func TestListAnthropicAPIKeys(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/gen-ai/anthropic/keys", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, listAnthropicAPIKeysResponse)
	})

	keys, resp, err := client.GenAI.ListAnthropicAPIKeys(ctx, nil)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, 2, len(keys))
	assert.Equal(t, "Anthropic Key One", keys[0].Name)
	assert.Equal(t, "11111111-1111-1111-1111-111111111111", keys[0].Uuid)
	assert.Equal(t, "Anthropic Key Two", keys[1].Name)
	assert.Equal(t, "22222222-2222-2222-2222-222222222222", keys[1].Uuid)
}

func TestCreateAnthropicAPIKey(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/gen-ai/anthropic/keys", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, anthropicAPIKeyInfoResponse)
	})

	req := &AnthropicAPIKeyCreateRequest{
		Name:   "Anthropic Key One",
		ApiKey: "11111111-1111-1111-1111-111111111111",
	}

	key, resp, err := client.GenAI.CreateAnthropicAPIKey(ctx, req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "Anthropic Key One", key.Name)
	assert.Equal(t, "11111111-1111-1111-1111-111111111111", key.Uuid)
}

func TestGetAnthropicAPIKey(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/gen-ai/anthropic/keys/11111111-1111-1111-1111-111111111111", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, anthropicAPIKeyInfoResponse)
	})

	key, resp, err := client.GenAI.GetAnthropicAPIKey(ctx, "11111111-1111-1111-1111-111111111111")
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "Anthropic Key One", key.Name)
	assert.Equal(t, "11111111-1111-1111-1111-111111111111", key.Uuid)
}

func TestUpdateAnthropicAPIKey(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/gen-ai/anthropic/keys/11111111-1111-1111-1111-111111111111", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		fmt.Fprint(w, anthropicAPIKeyInfoResponse)
	})

	req := &AnthropicAPIKeyUpdateRequest{
		Name:       "Anthropic Key One",
		ApiKey:     "11111111-1111-1111-1111-111111111111",
		ApiKeyUuid: "11111111-1111-1111-1111-111111111111",
	}

	key, resp, err := client.GenAI.UpdateAnthropicAPIKey(ctx, "11111111-1111-1111-1111-111111111111", req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "Anthropic Key One", key.Name)
	assert.Equal(t, "11111111-1111-1111-1111-111111111111", key.Uuid)
}

func TestDeleteAnthropicAPIKey(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/gen-ai/anthropic/keys/11111111-1111-1111-1111-111111111111", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		fmt.Fprint(w, anthropicAPIKeyInfoResponse)
	})

	key, resp, err := client.GenAI.DeleteAnthropicAPIKey(ctx, "11111111-1111-1111-1111-111111111111")
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "Anthropic Key One", key.Name)
	assert.Equal(t, "11111111-1111-1111-1111-111111111111", key.Uuid)
}

func TestListAgentsByAnthropicAPIKey(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/gen-ai/anthropic/keys/11111111-1111-1111-1111-111111111111/agents", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, listAgentsByAnthropicAPIKeyResponse)
	})

	agents, resp, err := client.GenAI.ListAgentsByAnthropicAPIKey(ctx, "11111111-1111-1111-1111-111111111111", nil)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, 2, len(agents))
	assert.Equal(t, "Anthropic Agent 1", agents[0].Name)
	assert.Equal(t, "00000000-0000-0000-0000-000000000000", agents[0].Uuid)
	assert.Equal(t, "Anthropic Agent 2", agents[1].Name)
	assert.Equal(t, "00000000-0000-0000-0000-000000000001", agents[1].Uuid)
	assert.NotNil(t, resp.Meta)
	assert.Equal(t, 2, resp.Meta.Total)
	assert.NotNil(t, resp.Links)
	assert.Equal(t, "https://api.digitalocean.com/v2/gen-ai/anthropic/keys?page=1&per_page=1", resp.Links.Pages.First)
}

func TestListOpenAIAPIKeys(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/gen-ai/openai/keys", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, listOpenAIAPIKeysResponse)
	})

	keys, resp, err := client.GenAI.ListOpenAIAPIKeys(ctx, nil)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, 2, len(keys))
	assert.Equal(t, "openai-key1", keys[0].Name)
	assert.Equal(t, "11111111-1111-1111-1111-111111111111", keys[0].Uuid)
	assert.Equal(t, "openai-key2", keys[1].Name)
	assert.Equal(t, "22222222-2222-2222-2222-222222222222", keys[1].Uuid)
}

func TestCreateOpenAIAPIKey(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/gen-ai/openai/keys", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, openaiAPIKeyInfoResponse)
	})

	req := &OpenAIAPIKeyCreateRequest{
		Name:   "OpenAI One",
		ApiKey: "11111111-1111-1111-1111-111111111111",
	}

	key, resp, err := client.GenAI.CreateOpenAIAPIKey(ctx, req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "OpenAI One", key.Name)
	assert.Equal(t, "11111111-1111-1111-1111-111111111111", key.Uuid)
}

func TestGetOpenAIAPIKey(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/gen-ai/openai/keys/11111111-1111-1111-1111-111111111111", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, openaiAPIKeyInfoResponse)
	})

	key, resp, err := client.GenAI.GetOpenAIAPIKey(ctx, "11111111-1111-1111-1111-111111111111")
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "OpenAI One", key.Name)
	assert.Equal(t, "11111111-1111-1111-1111-111111111111", key.Uuid)
}

func TestUpdateOpenAIAPIKey(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/gen-ai/openai/keys/11111111-1111-1111-1111-111111111111", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		fmt.Fprint(w, openaiAPIKeyInfoResponse)
	})

	req := &OpenAIAPIKeyUpdateRequest{
		Name:       "OpenAI One",
		ApiKey:     "11111111-1111-1111-1111-111111111111",
		ApiKeyUuid: "11111111-1111-1111-1111-111111111111",
	}

	key, resp, err := client.GenAI.UpdateOpenAIAPIKey(ctx, "11111111-1111-1111-1111-111111111111", req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "OpenAI One", key.Name)
	assert.Equal(t, "11111111-1111-1111-1111-111111111111", key.Uuid)
}

func TestDeleteOpenAIAPIKey(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/gen-ai/openai/keys/11111111-1111-1111-1111-111111111111", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		fmt.Fprint(w, openaiAPIKeyInfoResponse)
	})

	key, resp, err := client.GenAI.DeleteOpenAIAPIKey(ctx, "11111111-1111-1111-1111-111111111111")
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "OpenAI One", key.Name)
	assert.Equal(t, "11111111-1111-1111-1111-111111111111", key.Uuid)
}

func TestListAgentsByOpenAIAPIKey(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/gen-ai/openai/keys/11111111-1111-1111-1111-111111111111/agents", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, listAgentsByOpenAIAPIKeyResponse)
	})

	agents, resp, err := client.GenAI.ListAgentsByOpenAIAPIKey(ctx, "11111111-1111-1111-1111-111111111111", nil)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, 2, len(agents))
	assert.Equal(t, "OpenAI Agent 1", agents[0].Name)
	assert.Equal(t, "00000000-0000-0000-0000-000000000000", agents[0].Uuid)
	assert.Equal(t, "OpenAI Agent 2", agents[1].Name)
	assert.Equal(t, "00000000-0000-0000-0000-000000000001", agents[1].Uuid)
	assert.NotNil(t, resp.Meta)
	assert.Equal(t, 2, resp.Meta.Total)
	assert.NotNil(t, resp.Links)
	assert.Equal(t, "https://api.digitalocean.com/v2/gen-ai/openai/keys?page=1&per_page=1", resp.Links.Pages.First)
}

func TestCreateFunctionRoute(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/gen-ai/agents/00000000-0000-0000-0000-000000000000/functions", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		// Respond with an agent object that includes a non-empty functions array
		fmt.Fprint(w, agentResponse)
	})

	req := &FunctionRouteCreateRequest{
		Description:   "Creating Function Route",
		AgentUuid:     "00000000-0000-0000-0000-000000000000",
		FaasName:      "godo-test-faasname",
		FaasNamespace: "fn-00000000-0000-0000-0000-000000000000",
		InputSchema: FunctionInputSchema{
			Parameters: []OpenAPIParameterSchema{
				{
					Name: "zipCode",
					In:   "query",
					Schema: NestedSchema{
						Type: "string",
					},
					Required:    false,
					Description: "The ZIP code for which to fetch the weather",
				},
				{
					Name: "measurement",
					In:   "query",
					Schema: NestedSchema{
						Type: "string",
						Enum: []string{"F", "C"},
					},
					Required:    false,
					Description: "The measurement unit for temperature (F or C)",
				},
			},
		},
		FunctionName: "godo-test-function",
		OutputSchema: json.RawMessage(`{
  "properties": [
    {
      "name": "temperature",
      "type": "number",
      "description": "The temperature for the specified location"
    },
    {
      "name": "measurement",
      "type": "string",
      "description": "The measurement unit used for the temperature (F or C)"
    },
    {
      "name": "conditions",
      "type": "string",
      "description": "A description of the current weather conditions (Sunny, Cloudy, etc)"
    }
  ]
}`),
	}

	agent, res, err := client.GenAI.CreateFunctionRoute(ctx, "00000000-0000-0000-0000-000000000000", req)
	if err != nil {
		t.Errorf("GenAI.Create returned error: %v", err)
	}

	assert.NoError(t, err)
	assert.Equal(t, 200, res.Response.StatusCode)
	t.Log(agent)
	assert.Equal(t, req.FunctionName, agent.Functions[0].Name)
}

func TestUpdateFunctionRoute(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/gen-ai/agents/00000000-0000-0000-0000-000000000000/functions/00000000-0000-0000-0000-000000000000", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		fmt.Fprint(w, agentUpdateResponse)
	})

	req := &FunctionRouteUpdateRequest{
		Description: "Updating Function Route",
		InputSchema: FunctionInputSchema{
			Parameters: []OpenAPIParameterSchema{
				{
					Name: "zipCode",
					In:   "query",
					Schema: NestedSchema{
						Type: "string",
					},
					Required:    false,
					Description: "The ZIP code for which to fetch the weather",
				},
				{
					Name: "measurement",
					In:   "query",
					Schema: NestedSchema{
						Type: "string",
						Enum: []string{"F", "C"},
					},
					Required:    false,
					Description: "The measurement unit for temperature (F or C)",
				},
			},
		},
		OutputSchema: json.RawMessage(`{
            "properties": [
                {
                    "name": "temperature",
                    "type": "number",
                    "description": "The temperature for the specified location"
                }
            ]
        }`),
	}

	agent, resp, err := client.GenAI.UpdateFunctionRoute(ctx, "00000000-0000-0000-0000-000000000000", "00000000-0000-0000-0000-000000000000", req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.Response.StatusCode)
	assert.Equal(t, req.Description, agent.Functions[0].Description)
}

func TestDeleteFunctionRoute(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/gen-ai/agents/00000000-0000-0000-0000-000000000000/functions/00000000-0000-0000-0000-000000000000", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{}`)
	})

	_, resp, err := client.GenAI.DeleteFunctionRoute(ctx, "00000000-0000-0000-0000-000000000000", "00000000-0000-0000-0000-000000000000")
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.Response.StatusCode)
}

func TestListAvailableModels(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/gen-ai/models", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		testFormValues(t, r, values{
			"page":     "1",
			"per_page": "1",
		})

		fmt.Fprint(w, listAvailableModelsResponse)
	})

	req := &ListOptions{
		Page:    1,
		PerPage: 1,
	}

	models, resp, err := client.GenAI.ListAvailableModels(ctx, req)
	if err != nil {
		t.Fatalf("GenAI ListAvailableModels returned error: %v", err)
	}

	assert.Equal(t, "Llama 3.3 Instruct (70B)", models[0].Name)
	assert.Equal(t, "00000000-0000-0000-0000-000000000000", models[0].Uuid)
	expectedString := fmt.Sprintf("%v", models[0])
	assert.Equal(t, 200, resp.Response.StatusCode)
	assert.Equal(t, expectedString, models[0].String())
}

func TestListDatacenterRegions(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/gen-ai/regions", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, listDatacenterRegionsResponse)
	})

	regions, resp, err := client.GenAI.ListDatacenterRegions(ctx, nil, nil)
	if err != nil {
		t.Fatalf("GenAI ListDatacenterRegions returned error: %v", err)
	}

	assert.Equal(t, 2, len(regions))
	assert.Equal(t, "tor1", regions[0].Region)
	assert.Equal(t, "https://tor1.gen-ai.digitalocean.com", regions[0].InferenceUrl)
	assert.Equal(t, true, regions[0].ServesBatch)
	assert.Equal(t, true, regions[0].ServesInference)
	assert.Equal(t, "https://tor1.gen-ai.digitalocean.com/stream", regions[0].StreamInferenceUrl)

	assert.Equal(t, "nyc3", regions[1].Region)
	assert.Equal(t, "https://nyc3.gen-ai.digitalocean.com", regions[1].InferenceUrl)
	assert.Equal(t, false, regions[1].ServesBatch)
	assert.Equal(t, true, regions[1].ServesInference)
	assert.Equal(t, "https://nyc3.gen-ai.digitalocean.com/stream", regions[1].StreamInferenceUrl)

	assert.Equal(t, 200, resp.Response.StatusCode)
}
