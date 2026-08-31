package godo

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	testScenarioSetUUID = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	testRunUUID         = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	testJourneyUUID     = "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"
	testLibraryUUID     = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
)

func TestCreateScenarioSetUploadPresignedURLs(t *testing.T) {
	setup()
	defer teardown()

	createReq := &CreateScenarioSetUploadPresignedURLsRequest{
		Files: []*PresignedUrlFile{
			{FileName: "scenarios.jsonl", FileSize: "1024"},
		},
	}

	mux.HandleFunc("/v2/gen-ai/scenario_sets/file_upload_presigned_urls", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)

		v := new(CreateScenarioSetUploadPresignedURLsRequest)
		assert.NoError(t, json.NewDecoder(r.Body).Decode(v))
		assert.Len(t, v.Files, 1)
		assert.Equal(t, "scenarios.jsonl", v.Files[0].FileName)
		assert.Equal(t, "1024", v.Files[0].FileSize)

		fmt.Fprint(w, `{
			"request_id": "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
			"uploads": [{
				"object_key": "obj-1",
				"original_file_name": "scenarios.jsonl",
				"presigned_url": "https://example.com/upload/scenarios.jsonl",
				"expires_at": "2025-05-08T03:37:28Z"
			}]
		}`)
	})

	out, resp, err := client.GradientAI.CreateScenarioSetUploadPresignedURLs(ctx, createReq)
	assert.NoError(t, err)
	assert.NotNil(t, out)
	assert.Equal(t, 200, resp.Response.StatusCode)
	assert.Equal(t, "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb", out.RequestID)
	assert.Len(t, out.Uploads, 1)
	assert.Equal(t, "scenarios.jsonl", out.Uploads[0].OriginalFileName)
	assert.Equal(t, "https://example.com/upload/scenarios.jsonl", out.Uploads[0].PresignedURL)
	assert.NotNil(t, out.Uploads[0].ExpiresAt)
}

func TestCreateScenarioSetUploadPresignedURLsNilRequest(t *testing.T) {
	setup()
	defer teardown()

	out, resp, err := client.GradientAI.CreateScenarioSetUploadPresignedURLs(ctx, nil)
	assert.Error(t, err)
	assert.Nil(t, out)
	assert.Nil(t, resp)
}

func TestCreateScenarioSetUploadPresignedURLsMissingFiles(t *testing.T) {
	setup()
	defer teardown()

	out, resp, err := client.GradientAI.CreateScenarioSetUploadPresignedURLs(ctx, &CreateScenarioSetUploadPresignedURLsRequest{})
	assert.Error(t, err)
	assert.Nil(t, out)
	assert.Nil(t, resp)
}

func TestCreateScenarioSetUploadPresignedURLsMissingFileName(t *testing.T) {
	setup()
	defer teardown()

	out, resp, err := client.GradientAI.CreateScenarioSetUploadPresignedURLs(ctx, &CreateScenarioSetUploadPresignedURLsRequest{
		Files: []*PresignedUrlFile{{FileSize: "1024"}},
	})
	assert.Error(t, err)
	assert.Nil(t, out)
	assert.Nil(t, resp)
}

func TestCreateScenarioSet(t *testing.T) {
	setup()
	defer teardown()

	createReq := &CreateScenarioSetRequest{
		Name: "support-scenarios",
		Scenarios: []*Scenario{
			{Name: "billing dispute", Description: "user disputes a charge", MaxTurns: 5},
		},
	}

	mux.HandleFunc("/v2/gen-ai/scenario_sets", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)

		v := new(CreateScenarioSetRequest)
		assert.NoError(t, json.NewDecoder(r.Body).Decode(v))
		assert.Equal(t, createReq.Name, v.Name)
		assert.Len(t, v.Scenarios, 1)
		assert.Equal(t, "billing dispute", v.Scenarios[0].Name)
		assert.Nil(t, v.FileUploadScenarioSet)

		fmt.Fprint(w, `{
			"scenario_set": {
				"scenario_set_uuid": "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
				"name": "support-scenarios",
				"status": "SCENARIO_SET_STATUS_READY",
				"source_kind": "SCENARIO_SET_SOURCE_KIND_USER_UPLOAD",
				"scenario_count": 1
			}
		}`)
	})

	out, resp, err := client.GradientAI.CreateScenarioSet(ctx, createReq)
	assert.NoError(t, err)
	assert.NotNil(t, out)
	assert.Equal(t, 200, resp.Response.StatusCode)
	assert.Equal(t, testScenarioSetUUID, out.ScenarioSetUUID)
	assert.Equal(t, "support-scenarios", out.Name)
	assert.Equal(t, ScenarioSetStatusReady, out.Status)
	assert.Equal(t, ScenarioSetSourceKindUserUpload, out.SourceKind)
	assert.Equal(t, uint32(1), out.ScenarioCount)
}

func TestCreateScenarioSetFromFileUpload(t *testing.T) {
	setup()
	defer teardown()

	createReq := &CreateScenarioSetRequest{
		Name: "support-scenarios",
		FileUploadScenarioSet: &FileUploadDataSource{
			OriginalFileName: "scenarios.jsonl",
			Size:             "1024",
			StoredObjectKey:  "obj-1",
		},
	}

	mux.HandleFunc("/v2/gen-ai/scenario_sets", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)

		v := new(CreateScenarioSetRequest)
		assert.NoError(t, json.NewDecoder(r.Body).Decode(v))
		assert.Empty(t, v.Scenarios)
		assert.NotNil(t, v.FileUploadScenarioSet)
		assert.Equal(t, "obj-1", v.FileUploadScenarioSet.StoredObjectKey)

		fmt.Fprint(w, `{
			"scenario_set": {
				"scenario_set_uuid": "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
				"name": "support-scenarios",
				"status": "SCENARIO_SET_STATUS_READY",
				"source_kind": "SCENARIO_SET_SOURCE_KIND_USER_UPLOAD"
			}
		}`)
	})

	out, resp, err := client.GradientAI.CreateScenarioSet(ctx, createReq)
	assert.NoError(t, err)
	assert.NotNil(t, out)
	assert.Equal(t, 200, resp.Response.StatusCode)
	assert.Equal(t, testScenarioSetUUID, out.ScenarioSetUUID)
}

func TestCreateScenarioSetNilRequest(t *testing.T) {
	setup()
	defer teardown()

	out, resp, err := client.GradientAI.CreateScenarioSet(ctx, nil)
	assert.Error(t, err)
	assert.Nil(t, out)
	assert.Nil(t, resp)
}

func TestCreateScenarioSetMissingName(t *testing.T) {
	setup()
	defer teardown()

	out, resp, err := client.GradientAI.CreateScenarioSet(ctx, &CreateScenarioSetRequest{
		Scenarios: []*Scenario{{Description: "user disputes a charge"}},
	})
	assert.Error(t, err)
	assert.Nil(t, out)
	assert.Nil(t, resp)
}

func TestCreateScenarioSetRequiresExactlyOneSource(t *testing.T) {
	setup()
	defer teardown()

	// Neither scenarios nor an uploaded file.
	out, resp, err := client.GradientAI.CreateScenarioSet(ctx, &CreateScenarioSetRequest{
		Name: "support-scenarios",
	})
	assert.Error(t, err)
	assert.Nil(t, out)
	assert.Nil(t, resp)

	// Both scenarios and an uploaded file.
	out, resp, err = client.GradientAI.CreateScenarioSet(ctx, &CreateScenarioSetRequest{
		Name:                  "support-scenarios",
		Scenarios:             []*Scenario{{Description: "user disputes a charge"}},
		FileUploadScenarioSet: &FileUploadDataSource{StoredObjectKey: "obj-1"},
	})
	assert.Error(t, err)
	assert.Nil(t, out)
	assert.Nil(t, resp)
}

func TestGenerateScenarioSet(t *testing.T) {
	setup()
	defer teardown()

	generateReq := &GenerateScenarioSetRequest{
		Name:               "goal-scenarios",
		GoalDescription:    "test refund flows",
		NumScenarios:       3,
		GeneratorModelUUID: "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
	}

	mux.HandleFunc("/v2/gen-ai/scenario_sets/generate", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)

		v := new(GenerateScenarioSetRequest)
		assert.NoError(t, json.NewDecoder(r.Body).Decode(v))
		assert.Equal(t, generateReq.Name, v.Name)
		assert.Equal(t, generateReq.GoalDescription, v.GoalDescription)
		assert.Equal(t, generateReq.NumScenarios, v.NumScenarios)
		assert.Equal(t, generateReq.GeneratorModelUUID, v.GeneratorModelUUID)

		fmt.Fprint(w, `{
			"scenario_set": {
				"scenario_set_uuid": "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
				"name": "goal-scenarios",
				"status": "SCENARIO_SET_STATUS_GENERATING",
				"source_kind": "SCENARIO_SET_SOURCE_KIND_GOAL_GENERATED",
				"source_goal_description": "test refund flows",
				"generator_model_uuid": "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"
			}
		}`)
	})

	out, resp, err := client.GradientAI.GenerateScenarioSet(ctx, generateReq)
	assert.NoError(t, err)
	assert.NotNil(t, out)
	assert.Equal(t, 200, resp.Response.StatusCode)
	assert.Equal(t, testScenarioSetUUID, out.ScenarioSetUUID)
	assert.Equal(t, ScenarioSetStatusGenerating, out.Status)
	assert.Equal(t, ScenarioSetSourceKindGoalGenerated, out.SourceKind)
	assert.Equal(t, "test refund flows", out.SourceGoalDescription)
}

func TestGenerateScenarioSetNilRequest(t *testing.T) {
	setup()
	defer teardown()

	out, resp, err := client.GradientAI.GenerateScenarioSet(ctx, nil)
	assert.Error(t, err)
	assert.Nil(t, out)
	assert.Nil(t, resp)
}

func TestGenerateScenarioSetMissingName(t *testing.T) {
	setup()
	defer teardown()

	out, resp, err := client.GradientAI.GenerateScenarioSet(ctx, &GenerateScenarioSetRequest{
		GoalDescription: "test refund flows",
	})
	assert.Error(t, err)
	assert.Nil(t, out)
	assert.Nil(t, resp)
}

func TestGenerateScenarioSetMissingGoalDescription(t *testing.T) {
	setup()
	defer teardown()

	out, resp, err := client.GradientAI.GenerateScenarioSet(ctx, &GenerateScenarioSetRequest{
		Name: "goal-scenarios",
	})
	assert.Error(t, err)
	assert.Nil(t, out)
	assert.Nil(t, resp)
}

func TestListScenarioSets(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/gen-ai/scenario_sets", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		q := r.URL.Query()
		assert.Equal(t, []string{
			string(ScenarioSetStatusReady),
			string(ScenarioSetStatusGenerating),
		}, q["statuses"])
		assert.Equal(t, []string{
			string(ScenarioSetSourceKindUserUpload),
			string(ScenarioSetSourceKindLibrary),
		}, q["source_kinds"])
		assert.Equal(t, "support", q.Get("search"))
		assert.Equal(t, string(ScenarioSetSortFieldCreatedAt), q.Get("sort_by"))
		assert.Equal(t, string(GenAISortDirectionDesc), q.Get("sort_direction"))
		assert.Equal(t, "1", q.Get("page"))
		assert.Equal(t, "10", q.Get("per_page"))

		fmt.Fprint(w, `{
			"scenario_sets": [{
				"scenario_set_uuid": "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
				"name": "support-scenarios",
				"status": "SCENARIO_SET_STATUS_READY",
				"source_kind": "SCENARIO_SET_SOURCE_KIND_USER_UPLOAD",
				"scenario_count": 2
			}],
			"meta": {"total": 1},
			"available_statuses": ["SCENARIO_SET_STATUS_READY"],
			"available_source_kinds": ["SCENARIO_SET_SOURCE_KIND_USER_UPLOAD"],
			"available_sort_by": ["SCENARIO_SET_SORT_FIELD_CREATED_AT"],
			"available_sort_directions": ["SORT_DIRECTION_DESC"]
		}`)
	})

	out, resp, err := client.GradientAI.ListScenarioSets(ctx, &ScenarioSetListOptions{
		Statuses: []ScenarioSetStatus{
			ScenarioSetStatusReady,
			ScenarioSetStatusGenerating,
		},
		SourceKinds: []ScenarioSetSourceKind{
			ScenarioSetSourceKindUserUpload,
			ScenarioSetSourceKindLibrary,
		},
		Search:        "support",
		SortBy:        ScenarioSetSortFieldCreatedAt,
		SortDirection: GenAISortDirectionDesc,
		ListOptions: ListOptions{
			Page:    1,
			PerPage: 10,
		},
	})
	assert.NoError(t, err)
	assert.NotNil(t, out)
	assert.Equal(t, 200, resp.Response.StatusCode)
	assert.Len(t, out.ScenarioSets, 1)
	assert.Equal(t, testScenarioSetUUID, out.ScenarioSets[0].ScenarioSetUUID)
	assert.Equal(t, "support-scenarios", out.ScenarioSets[0].Name)
	assert.NotNil(t, out.Meta)
	assert.Equal(t, 1, out.Meta.Total)
	assert.Equal(t, []ScenarioSetStatus{ScenarioSetStatusReady}, out.AvailableStatuses)
}

func TestGetScenarioSet(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/gen-ai/scenario_sets/"+testScenarioSetUUID, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"scenario_set": {
				"scenario_set_uuid": "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
				"name": "support-scenarios",
				"status": "SCENARIO_SET_STATUS_READY",
				"source_kind": "SCENARIO_SET_SOURCE_KIND_USER_UPLOAD",
				"scenario_count": 2
			}
		}`)
	})

	out, resp, err := client.GradientAI.GetScenarioSet(ctx, testScenarioSetUUID)
	assert.NoError(t, err)
	assert.NotNil(t, out)
	assert.Equal(t, 200, resp.Response.StatusCode)
	assert.Equal(t, testScenarioSetUUID, out.ScenarioSetUUID)
	assert.Equal(t, ScenarioSetStatusReady, out.Status)
}

func TestGetScenarioSetMissingUUID(t *testing.T) {
	setup()
	defer teardown()

	out, resp, err := client.GradientAI.GetScenarioSet(ctx, "")
	assert.Error(t, err)
	assert.Nil(t, out)
	assert.Nil(t, resp)
}

func TestListScenarios(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/gen-ai/scenario_sets/"+testScenarioSetUUID+"/scenarios", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		q := r.URL.Query()
		assert.Equal(t, "billing", q.Get("search"))
		assert.Equal(t, string(ScenarioSortFieldName), q.Get("sort_by"))
		assert.Equal(t, string(GenAISortDirectionAsc), q.Get("sort_direction"))
		assert.Equal(t, "2", q.Get("page"))
		assert.Equal(t, "25", q.Get("per_page"))

		fmt.Fprint(w, `{
			"scenarios": [{
				"scenario_uuid": "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
				"name": "billing dispute",
				"description": "user disputes a charge",
				"max_turns": 5
			}],
			"meta": {"total": 1}
		}`)
	})

	out, resp, err := client.GradientAI.ListScenarios(ctx, testScenarioSetUUID, &ScenarioListOptions{
		Search:        "billing",
		SortBy:        ScenarioSortFieldName,
		SortDirection: GenAISortDirectionAsc,
		ListOptions: ListOptions{
			Page:    2,
			PerPage: 25,
		},
	})
	assert.NoError(t, err)
	assert.NotNil(t, out)
	assert.Equal(t, 200, resp.Response.StatusCode)
	assert.Len(t, out.Scenarios, 1)
	assert.Equal(t, testJourneyUUID, out.Scenarios[0].ScenarioUUID)
	assert.Equal(t, "billing dispute", out.Scenarios[0].Name)
}

func TestListScenariosMissingUUID(t *testing.T) {
	setup()
	defer teardown()

	out, resp, err := client.GradientAI.ListScenarios(ctx, "", nil)
	assert.Error(t, err)
	assert.Nil(t, out)
	assert.Nil(t, resp)
}

func TestGetScenarioSetDownloadURL(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/gen-ai/scenario_sets/"+testScenarioSetUUID+"/download_url", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"download_url": "https://example.com/download/scenarios.jsonl",
			"expires_at": "2025-05-08T03:37:28Z"
		}`)
	})

	out, resp, err := client.GradientAI.GetScenarioSetDownloadURL(ctx, testScenarioSetUUID)
	assert.NoError(t, err)
	assert.NotNil(t, out)
	assert.Equal(t, 200, resp.Response.StatusCode)
	assert.Equal(t, "https://example.com/download/scenarios.jsonl", out.DownloadURL)
	assert.NotNil(t, out.ExpiresAt)
}

func TestGetScenarioSetDownloadURLMissingUUID(t *testing.T) {
	setup()
	defer teardown()

	out, resp, err := client.GradientAI.GetScenarioSetDownloadURL(ctx, "")
	assert.Error(t, err)
	assert.Nil(t, out)
	assert.Nil(t, resp)
}

func TestUpdateScenarioSet(t *testing.T) {
	setup()
	defer teardown()

	updateReq := &UpdateScenarioSetRequest{
		ScenarioSetUUID: testScenarioSetUUID,
		Name:            "renamed-scenarios",
		Scenarios: []*Scenario{
			{Name: "updated scenario", MaxTurns: 8},
		},
	}

	mux.HandleFunc("/v2/gen-ai/scenario_sets/"+testScenarioSetUUID, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)

		v := new(UpdateScenarioSetRequest)
		assert.NoError(t, json.NewDecoder(r.Body).Decode(v))
		assert.Equal(t, updateReq.Name, v.Name)
		assert.Len(t, v.Scenarios, 1)
		assert.Equal(t, "updated scenario", v.Scenarios[0].Name)

		fmt.Fprint(w, `{
			"scenario_set": {
				"scenario_set_uuid": "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
				"name": "renamed-scenarios",
				"status": "SCENARIO_SET_STATUS_READY",
				"scenario_count": 1
			}
		}`)
	})

	out, resp, err := client.GradientAI.UpdateScenarioSet(ctx, testScenarioSetUUID, updateReq)
	assert.NoError(t, err)
	assert.NotNil(t, out)
	assert.Equal(t, 200, resp.Response.StatusCode)
	assert.Equal(t, "renamed-scenarios", out.Name)
}

func TestUpdateScenarioSetNilRequest(t *testing.T) {
	setup()
	defer teardown()

	out, resp, err := client.GradientAI.UpdateScenarioSet(ctx, testScenarioSetUUID, nil)
	assert.Error(t, err)
	assert.Nil(t, out)
	assert.Nil(t, resp)
}

func TestUpdateScenarioSetMissingUUID(t *testing.T) {
	setup()
	defer teardown()

	out, resp, err := client.GradientAI.UpdateScenarioSet(ctx, "", &UpdateScenarioSetRequest{Name: "x"})
	assert.Error(t, err)
	assert.Nil(t, out)
	assert.Nil(t, resp)
}

func TestUpdateScenarioSetEmptyUpdate(t *testing.T) {
	setup()
	defer teardown()

	out, resp, err := client.GradientAI.UpdateScenarioSet(ctx, testScenarioSetUUID, &UpdateScenarioSetRequest{})
	assert.Error(t, err)
	assert.Nil(t, out)
	assert.Nil(t, resp)
}

func TestDeleteScenarioSet(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/gen-ai/scenario_sets/"+testScenarioSetUUID, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		fmt.Fprint(w, `{"scenario_set_uuid":"aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"}`)
	})

	out, resp, err := client.GradientAI.DeleteScenarioSet(ctx, testScenarioSetUUID)
	assert.NoError(t, err)
	assert.NotNil(t, out)
	assert.Equal(t, 200, resp.Response.StatusCode)
	assert.Equal(t, testScenarioSetUUID, out.ScenarioSetUUID)
}

func TestDeleteScenarioSetMissingUUID(t *testing.T) {
	setup()
	defer teardown()

	out, resp, err := client.GradientAI.DeleteScenarioSet(ctx, "")
	assert.Error(t, err)
	assert.Nil(t, out)
	assert.Nil(t, resp)
}

func TestListScenarioLibrary(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/gen-ai/scenario_library", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		q := r.URL.Query()
		assert.Equal(t, "support", q.Get("category"))
		assert.Equal(t, "refund", q.Get("search"))
		assert.Equal(t, string(ScenarioLibrarySortFieldName), q.Get("sort_by"))
		assert.Equal(t, string(GenAISortDirectionAsc), q.Get("sort_direction"))
		assert.Equal(t, "1", q.Get("page"))
		assert.Equal(t, "20", q.Get("per_page"))

		fmt.Fprint(w, `{
			"scenarios": [{
				"library_scenario_uuid": "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
				"name": "refund library",
				"category": "support",
				"status": "SCENARIO_LIBRARY_ENTRY_STATUS_ACTIVE",
				"scenario_count": 4
			}],
			"available_categories": ["support"],
			"meta": {"total": 1}
		}`)
	})

	out, resp, err := client.GradientAI.ListScenarioLibrary(ctx, &ScenarioLibraryListOptions{
		Category:      "support",
		Search:        "refund",
		SortBy:        ScenarioLibrarySortFieldName,
		SortDirection: GenAISortDirectionAsc,
		ListOptions: ListOptions{
			Page:    1,
			PerPage: 20,
		},
	})
	assert.NoError(t, err)
	assert.NotNil(t, out)
	assert.Equal(t, 200, resp.Response.StatusCode)
	assert.Len(t, out.Scenarios, 1)
	assert.Equal(t, testLibraryUUID, out.Scenarios[0].LibraryScenarioUUID)
	assert.Equal(t, ScenarioLibraryEntryStatusActive, out.Scenarios[0].Status)
	assert.Equal(t, []string{"support"}, out.AvailableCategories)
}

func TestListScenarioLibraryScenarios(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/gen-ai/scenario_library/"+testLibraryUUID+"/scenarios", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		q := r.URL.Query()
		assert.Equal(t, "persona", q.Get("search"))
		assert.Equal(t, string(ScenarioSortFieldFileOrder), q.Get("sort_by"))
		assert.Equal(t, string(GenAISortDirectionAsc), q.Get("sort_direction"))

		fmt.Fprint(w, `{
			"scenarios": [{
				"scenario_uuid": "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
				"name": "angry customer",
				"user_persona": "frustrated payer"
			}]
		}`)
	})

	out, resp, err := client.GradientAI.ListScenarioLibraryScenarios(ctx, testLibraryUUID, &ScenarioListOptions{
		Search:        "persona",
		SortBy:        ScenarioSortFieldFileOrder,
		SortDirection: GenAISortDirectionAsc,
	})
	assert.NoError(t, err)
	assert.NotNil(t, out)
	assert.Equal(t, 200, resp.Response.StatusCode)
	assert.Len(t, out.Scenarios, 1)
	assert.Equal(t, "angry customer", out.Scenarios[0].Name)
}

func TestListScenarioLibraryScenariosMissingUUID(t *testing.T) {
	setup()
	defer teardown()

	out, resp, err := client.GradientAI.ListScenarioLibraryScenarios(ctx, "", nil)
	assert.Error(t, err)
	assert.Nil(t, out)
	assert.Nil(t, resp)
}

func TestCreateScenarioSetFromLibrary(t *testing.T) {
	setup()
	defer teardown()

	createReq := &CreateScenarioSetFromLibraryRequest{
		LibraryScenarioUUID: testLibraryUUID,
		Name:                "from-library",
	}

	mux.HandleFunc("/v2/gen-ai/scenario_library/"+testLibraryUUID+"/create_scenario_set", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)

		v := new(CreateScenarioSetFromLibraryRequest)
		assert.NoError(t, json.NewDecoder(r.Body).Decode(v))
		assert.Equal(t, testLibraryUUID, v.LibraryScenarioUUID)
		assert.Equal(t, "from-library", v.Name)

		fmt.Fprint(w, `{
			"scenario_set": {
				"scenario_set_uuid": "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
				"name": "from-library",
				"status": "SCENARIO_SET_STATUS_READY",
				"source_kind": "SCENARIO_SET_SOURCE_KIND_LIBRARY",
				"library_scenario_uuid": "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
			}
		}`)
	})

	out, resp, err := client.GradientAI.CreateScenarioSetFromLibrary(ctx, testLibraryUUID, createReq)
	assert.NoError(t, err)
	assert.NotNil(t, out)
	assert.Equal(t, 200, resp.Response.StatusCode)
	assert.Equal(t, testJourneyUUID, out.ScenarioSetUUID)
	assert.Equal(t, ScenarioSetSourceKindLibrary, out.SourceKind)
	assert.Equal(t, testLibraryUUID, out.LibraryScenarioUUID)
}

func TestCreateScenarioSetFromLibraryMissingUUID(t *testing.T) {
	setup()
	defer teardown()

	out, resp, err := client.GradientAI.CreateScenarioSetFromLibrary(ctx, "", &CreateScenarioSetFromLibraryRequest{Name: "x"})
	assert.Error(t, err)
	assert.Nil(t, out)
	assert.Nil(t, resp)
}

func TestCreateSimulationRun(t *testing.T) {
	setup()
	defer teardown()

	threshold := float32(0.8)
	createReq := &CreateSimulationRunRequest{
		ScenarioSetUUID: testScenarioSetUUID,
		Name:            "sim-run-1",
		AgentConfig: &CandidateAgentConfig{
			AgentUUID: "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
			Name:      "candidate",
		},
		UserSimulatorModelUUID: "cccccccc-cccc-cccc-cccc-cccccccccccc",
		JudgeModelUUID:         "dddddddd-dddd-dddd-dddd-dddddddddddd",
		ExplorationBudget:      2,
		MaxTurns:               10,
		EvaluationConfig: &SimulationEvaluationConfig{
			MetricUUIDs: []string{"eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee"},
			StarMetric: &StarMetric{
				MetricUUID:       "eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee",
				Name:             "accuracy",
				SuccessThreshold: &threshold,
			},
		},
	}

	mux.HandleFunc("/v2/gen-ai/simulation_runs", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)

		v := new(CreateSimulationRunRequest)
		assert.NoError(t, json.NewDecoder(r.Body).Decode(v))
		assert.Equal(t, createReq.Name, v.Name)
		assert.Equal(t, createReq.ScenarioSetUUID, v.ScenarioSetUUID)
		assert.NotNil(t, v.AgentConfig)
		assert.Equal(t, "candidate", v.AgentConfig.Name)
		assert.NotNil(t, v.EvaluationConfig)
		assert.Len(t, v.EvaluationConfig.MetricUUIDs, 1)
		assert.NotNil(t, v.EvaluationConfig.StarMetric)
		assert.Equal(t, "accuracy", v.EvaluationConfig.StarMetric.Name)

		fmt.Fprint(w, `{
			"simulation_run": {
				"run_uuid": "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
				"name": "sim-run-1",
				"scenario_set_uuid": "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
				"status": "SIMULATION_RUN_STATUS_PENDING",
				"agent_config": {
					"agent_uuid": "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
					"name": "candidate"
				},
				"exploration_budget": 2,
				"max_turns": 10
			}
		}`)
	})

	out, resp, err := client.GradientAI.CreateSimulationRun(ctx, createReq)
	assert.NoError(t, err)
	assert.NotNil(t, out)
	assert.Equal(t, 200, resp.Response.StatusCode)
	assert.Equal(t, testRunUUID, out.RunUUID)
	assert.Equal(t, "sim-run-1", out.Name)
	assert.Equal(t, SimulationRunStatusPending, out.Status)
	assert.NotNil(t, out.AgentConfig)
	assert.Equal(t, "candidate", out.AgentConfig.Name)
}

func TestCreateSimulationRunNilRequest(t *testing.T) {
	setup()
	defer teardown()

	out, resp, err := client.GradientAI.CreateSimulationRun(ctx, nil)
	assert.Error(t, err)
	assert.Nil(t, out)
	assert.Nil(t, resp)
}

func TestCreateSimulationRunMissingScenarioSetUUID(t *testing.T) {
	setup()
	defer teardown()

	out, resp, err := client.GradientAI.CreateSimulationRun(ctx, &CreateSimulationRunRequest{
		AgentConfig: &CandidateAgentConfig{AgentUUID: "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"},
	})
	assert.Error(t, err)
	assert.Nil(t, out)
	assert.Nil(t, resp)
}

func TestCreateSimulationRunMissingAgentUUID(t *testing.T) {
	setup()
	defer teardown()

	// No agent config at all.
	out, resp, err := client.GradientAI.CreateSimulationRun(ctx, &CreateSimulationRunRequest{
		ScenarioSetUUID: testScenarioSetUUID,
	})
	assert.Error(t, err)
	assert.Nil(t, out)
	assert.Nil(t, resp)

	// Agent config without an agent UUID.
	out, resp, err = client.GradientAI.CreateSimulationRun(ctx, &CreateSimulationRunRequest{
		ScenarioSetUUID: testScenarioSetUUID,
		AgentConfig:     &CandidateAgentConfig{Name: "candidate"},
	})
	assert.Error(t, err)
	assert.Nil(t, out)
	assert.Nil(t, resp)
}

func TestListSimulationRuns(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/gen-ai/simulation_runs", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		q := r.URL.Query()
		assert.Equal(t, testScenarioSetUUID, q.Get("scenario_set_uuid"))
		assert.Equal(t, []string{
			string(SimulationRunStatusRunning),
			string(SimulationRunStatusSucceeded),
		}, q["statuses"])
		assert.Equal(t, "sim", q.Get("search"))
		assert.Equal(t, string(SimulationRunSortFieldCreatedAt), q.Get("sort_by"))
		assert.Equal(t, string(GenAISortDirectionDesc), q.Get("sort_direction"))
		assert.Equal(t, "1", q.Get("page"))
		assert.Equal(t, "10", q.Get("per_page"))

		fmt.Fprint(w, `{
			"simulation_runs": [{
				"run_uuid": "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
				"name": "sim-run-1",
				"status": "SIMULATION_RUN_STATUS_RUNNING",
				"scenario_set_uuid": "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
			}],
			"meta": {"total": 1},
			"available_statuses": ["SIMULATION_RUN_STATUS_RUNNING"],
			"available_sort_by": ["SIMULATION_RUN_SORT_FIELD_CREATED_AT"],
			"available_sort_directions": ["SORT_DIRECTION_DESC"]
		}`)
	})

	out, resp, err := client.GradientAI.ListSimulationRuns(ctx, &SimulationRunListOptions{
		ScenarioSetUUID: testScenarioSetUUID,
		Statuses: []SimulationRunStatus{
			SimulationRunStatusRunning,
			SimulationRunStatusSucceeded,
		},
		Search:        "sim",
		SortBy:        SimulationRunSortFieldCreatedAt,
		SortDirection: GenAISortDirectionDesc,
		ListOptions: ListOptions{
			Page:    1,
			PerPage: 10,
		},
	})
	assert.NoError(t, err)
	assert.NotNil(t, out)
	assert.Equal(t, 200, resp.Response.StatusCode)
	assert.Len(t, out.SimulationRuns, 1)
	assert.Equal(t, testRunUUID, out.SimulationRuns[0].RunUUID)
	assert.Equal(t, SimulationRunStatusRunning, out.SimulationRuns[0].Status)
	assert.Equal(t, []SimulationRunStatus{SimulationRunStatusRunning}, out.AvailableStatuses)
}

func TestGetSimulationRun(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/gen-ai/simulation_runs/"+testRunUUID, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"simulation_run": {
				"run_uuid": "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
				"name": "sim-run-1",
				"status": "SIMULATION_RUN_STATUS_SUCCEEDED",
				"scenario_set_uuid": "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
				"total_journeys": 2,
				"journeys_finished": 2,
				"result_summary": {
					"verdict_counts": {"success_count": 2},
					"total_duration_sec": "42"
				}
			},
			"scenario_results": [{
				"scenario_uuid": "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
				"total_journeys": 2,
				"journeys_finished": 2,
				"verdict_counts": {"success_count": 2}
			}]
		}`)
	})

	out, resp, err := client.GradientAI.GetSimulationRun(ctx, testRunUUID)
	assert.NoError(t, err)
	assert.NotNil(t, out)
	assert.Equal(t, 200, resp.Response.StatusCode)
	assert.NotNil(t, out.SimulationRun)
	assert.Equal(t, testRunUUID, out.SimulationRun.RunUUID)
	assert.Equal(t, SimulationRunStatusSucceeded, out.SimulationRun.Status)
	assert.NotNil(t, out.SimulationRun.ResultSummary)
	assert.Equal(t, "42", out.SimulationRun.ResultSummary.TotalDurationSec)
	assert.Len(t, out.ScenarioResults, 1)
	assert.Equal(t, testJourneyUUID, out.ScenarioResults[0].ScenarioUUID)
	assert.Equal(t, uint32(2), out.ScenarioResults[0].TotalJourneys)
}

func TestGetSimulationRunMissingUUID(t *testing.T) {
	setup()
	defer teardown()

	out, resp, err := client.GradientAI.GetSimulationRun(ctx, "")
	assert.Error(t, err)
	assert.Nil(t, out)
	assert.Nil(t, resp)
}

func TestUpdateSimulationRun(t *testing.T) {
	setup()
	defer teardown()

	updateReq := &UpdateSimulationRunRequest{
		RunUUID: testRunUUID,
		Name:    "renamed-run",
	}

	mux.HandleFunc("/v2/gen-ai/simulation_runs/"+testRunUUID, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)

		v := new(UpdateSimulationRunRequest)
		assert.NoError(t, json.NewDecoder(r.Body).Decode(v))
		assert.Equal(t, "renamed-run", v.Name)

		fmt.Fprint(w, `{
			"simulation_run": {
				"run_uuid": "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
				"name": "renamed-run",
				"status": "SIMULATION_RUN_STATUS_SUCCEEDED"
			}
		}`)
	})

	out, resp, err := client.GradientAI.UpdateSimulationRun(ctx, testRunUUID, updateReq)
	assert.NoError(t, err)
	assert.NotNil(t, out)
	assert.Equal(t, 200, resp.Response.StatusCode)
	assert.Equal(t, "renamed-run", out.Name)
}

func TestUpdateSimulationRunNilRequest(t *testing.T) {
	setup()
	defer teardown()

	out, resp, err := client.GradientAI.UpdateSimulationRun(ctx, testRunUUID, nil)
	assert.Error(t, err)
	assert.Nil(t, out)
	assert.Nil(t, resp)
}

func TestUpdateSimulationRunMissingUUID(t *testing.T) {
	setup()
	defer teardown()

	out, resp, err := client.GradientAI.UpdateSimulationRun(ctx, "", &UpdateSimulationRunRequest{Name: "x"})
	assert.Error(t, err)
	assert.Nil(t, out)
	assert.Nil(t, resp)
}

func TestUpdateSimulationRunMissingName(t *testing.T) {
	setup()
	defer teardown()

	out, resp, err := client.GradientAI.UpdateSimulationRun(ctx, testRunUUID, &UpdateSimulationRunRequest{})
	assert.Error(t, err)
	assert.Nil(t, out)
	assert.Nil(t, resp)
}

func TestCancelSimulationRun(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/gen-ai/simulation_runs/"+testRunUUID+"/cancel", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPatch)
		fmt.Fprint(w, `{
			"simulation_run": {
				"run_uuid": "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
				"name": "sim-run-1",
				"status": "SIMULATION_RUN_STATUS_CANCELLED"
			}
		}`)
	})

	out, resp, err := client.GradientAI.CancelSimulationRun(ctx, testRunUUID)
	assert.NoError(t, err)
	assert.NotNil(t, out)
	assert.Equal(t, 200, resp.Response.StatusCode)
	assert.Equal(t, SimulationRunStatusCancelled, out.Status)
}

func TestCancelSimulationRunMissingUUID(t *testing.T) {
	setup()
	defer teardown()

	out, resp, err := client.GradientAI.CancelSimulationRun(ctx, "")
	assert.Error(t, err)
	assert.Nil(t, out)
	assert.Nil(t, resp)
}

func TestDeleteSimulationRun(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/gen-ai/simulation_runs/"+testRunUUID, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		fmt.Fprint(w, `{"run_uuid":"aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"}`)
	})

	out, resp, err := client.GradientAI.DeleteSimulationRun(ctx, testRunUUID)
	assert.NoError(t, err)
	assert.NotNil(t, out)
	assert.Equal(t, 200, resp.Response.StatusCode)
	assert.Equal(t, testRunUUID, out.RunUUID)
}

func TestDeleteSimulationRunMissingUUID(t *testing.T) {
	setup()
	defer teardown()

	out, resp, err := client.GradientAI.DeleteSimulationRun(ctx, "")
	assert.Error(t, err)
	assert.Nil(t, out)
	assert.Nil(t, resp)
}

func TestListSimulationJourneys(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/gen-ai/simulation_runs/"+testRunUUID+"/journeys", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		q := r.URL.Query()
		assert.Equal(t, testJourneyUUID, q.Get("scenario_uuid"))
		assert.Equal(t, []string{
			string(SimulationJourneyStatusFinished),
			string(SimulationJourneyStatusFailed),
		}, q["statuses"])
		assert.Equal(t, []string{
			string(SimulationJourneyVerdictSuccess),
			string(SimulationJourneyVerdictFailure),
		}, q["verdicts"])
		assert.Equal(t, "billing", q.Get("search"))
		assert.Equal(t, string(SimulationJourneySortFieldCreatedAt), q.Get("sort_by"))
		assert.Equal(t, string(GenAISortDirectionDesc), q.Get("sort_direction"))
		assert.Equal(t, "1", q.Get("page"))
		assert.Equal(t, "50", q.Get("per_page"))

		fmt.Fprint(w, `{
			"journeys": [{
				"journey_uuid": "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
				"run_uuid": "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
				"scenario_uuid": "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
				"status": "SIMULATION_JOURNEY_STATUS_FINISHED",
				"verdict": "SIMULATION_JOURNEY_VERDICT_SUCCESS",
				"duration_sec": "12"
			}],
			"meta": {"total": 1}
		}`)
	})

	out, resp, err := client.GradientAI.ListSimulationJourneys(ctx, testRunUUID, &SimulationJourneyListOptions{
		ScenarioUUID: testJourneyUUID,
		Statuses: []SimulationJourneyStatus{
			SimulationJourneyStatusFinished,
			SimulationJourneyStatusFailed,
		},
		Verdicts: []SimulationJourneyVerdict{
			SimulationJourneyVerdictSuccess,
			SimulationJourneyVerdictFailure,
		},
		Search:        "billing",
		SortBy:        SimulationJourneySortFieldCreatedAt,
		SortDirection: GenAISortDirectionDesc,
		ListOptions: ListOptions{
			Page:    1,
			PerPage: 50,
		},
	})
	assert.NoError(t, err)
	assert.NotNil(t, out)
	assert.Equal(t, 200, resp.Response.StatusCode)
	assert.Len(t, out.Journeys, 1)
	assert.Equal(t, testJourneyUUID, out.Journeys[0].JourneyUUID)
	assert.Equal(t, SimulationJourneyStatusFinished, out.Journeys[0].Status)
	assert.Equal(t, SimulationJourneyVerdictSuccess, out.Journeys[0].Verdict)
}

func TestListSimulationJourneysMissingUUID(t *testing.T) {
	setup()
	defer teardown()

	out, resp, err := client.GradientAI.ListSimulationJourneys(ctx, "", nil)
	assert.Error(t, err)
	assert.Nil(t, out)
	assert.Nil(t, resp)
}

func TestGetSimulationJourney(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc(fmt.Sprintf("/v2/gen-ai/simulation_runs/%s/journeys/%s", testRunUUID, testJourneyUUID), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"journey": {
				"journey_uuid": "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
				"run_uuid": "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
				"scenario_uuid": "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
				"status": "SIMULATION_JOURNEY_STATUS_FINISHED",
				"verdict": "SIMULATION_JOURNEY_VERDICT_SUCCESS",
				"judge_reasoning": "criteria met",
				"token_usage": {"total_tokens": "100"}
			}
		}`)
	})

	out, resp, err := client.GradientAI.GetSimulationJourney(ctx, testRunUUID, testJourneyUUID)
	assert.NoError(t, err)
	assert.NotNil(t, out)
	assert.Equal(t, 200, resp.Response.StatusCode)
	assert.Equal(t, testJourneyUUID, out.JourneyUUID)
	assert.Equal(t, SimulationJourneyVerdictSuccess, out.Verdict)
	assert.Equal(t, "criteria met", out.JudgeReasoning)
	assert.NotNil(t, out.TokenUsage)
	assert.Equal(t, "100", out.TokenUsage.TotalTokens)
}

func TestGetSimulationJourneyMissingRunUUID(t *testing.T) {
	setup()
	defer teardown()

	out, resp, err := client.GradientAI.GetSimulationJourney(ctx, "", testJourneyUUID)
	assert.Error(t, err)
	assert.Nil(t, out)
	assert.Nil(t, resp)
}

func TestGetSimulationJourneyMissingJourneyUUID(t *testing.T) {
	setup()
	defer teardown()

	out, resp, err := client.GradientAI.GetSimulationJourney(ctx, testRunUUID, "")
	assert.Error(t, err)
	assert.Nil(t, out)
	assert.Nil(t, resp)
}

func TestGetSimulationJourneyTrajectoryURL(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc(fmt.Sprintf("/v2/gen-ai/simulation_runs/%s/journeys/%s/trajectory_url", testRunUUID, testJourneyUUID), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"download_url": "https://example.com/trajectory.json",
			"expires_at": "2025-05-08T03:37:28Z"
		}`)
	})

	out, resp, err := client.GradientAI.GetSimulationJourneyTrajectoryURL(ctx, testRunUUID, testJourneyUUID)
	assert.NoError(t, err)
	assert.NotNil(t, out)
	assert.Equal(t, 200, resp.Response.StatusCode)
	assert.Equal(t, "https://example.com/trajectory.json", out.DownloadURL)
	assert.NotNil(t, out.ExpiresAt)
}

func TestGetSimulationJourneyTrajectoryURLMissingRunUUID(t *testing.T) {
	setup()
	defer teardown()

	out, resp, err := client.GradientAI.GetSimulationJourneyTrajectoryURL(ctx, "", testJourneyUUID)
	assert.Error(t, err)
	assert.Nil(t, out)
	assert.Nil(t, resp)
}

func TestGetSimulationJourneyTrajectoryURLMissingJourneyUUID(t *testing.T) {
	setup()
	defer teardown()

	out, resp, err := client.GradientAI.GetSimulationJourneyTrajectoryURL(ctx, testRunUUID, "")
	assert.Error(t, err)
	assert.Nil(t, out)
	assert.Nil(t, resp)
}

func TestGetSimulationJourneyTrajectory(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc(fmt.Sprintf("/v2/gen-ai/simulation_runs/%s/journeys/%s/trajectory", testRunUUID, testJourneyUUID), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"trajectory": {
				"journey_uuid": "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
				"run_uuid": "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
				"scenario_uuid": "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
				"status": "SIMULATION_TRAJECTORY_STATUS_COMPLETED",
				"verdict": "SIMULATION_JOURNEY_VERDICT_SUCCESS",
				"turn_count": 2,
				"messages": [
					{"turn_index": 0, "role": "user", "content": "hello"},
					{"turn_index": 1, "role": "assistant", "content": "hi"}
				],
				"judge": {
					"verdict": "SIMULATION_JOURNEY_VERDICT_SUCCESS",
					"reasoning": "ok"
				}
			}
		}`)
	})

	out, resp, err := client.GradientAI.GetSimulationJourneyTrajectory(ctx, testRunUUID, testJourneyUUID)
	assert.NoError(t, err)
	assert.NotNil(t, out)
	assert.Equal(t, 200, resp.Response.StatusCode)
	assert.Equal(t, testJourneyUUID, out.JourneyUUID)
	assert.Equal(t, SimulationTrajectoryStatusCompleted, out.Status)
	assert.Equal(t, SimulationJourneyVerdictSuccess, out.Verdict)
	assert.Equal(t, uint32(2), out.TurnCount)
	assert.Len(t, out.Messages, 2)
	assert.Equal(t, "user", out.Messages[0].Role)
	assert.NotNil(t, out.Judge)
	assert.Equal(t, "ok", out.Judge.Reasoning)
}

func TestGetSimulationJourneyTrajectoryMissingRunUUID(t *testing.T) {
	setup()
	defer teardown()

	out, resp, err := client.GradientAI.GetSimulationJourneyTrajectory(ctx, "", testJourneyUUID)
	assert.Error(t, err)
	assert.Nil(t, out)
	assert.Nil(t, resp)
}

func TestGetSimulationJourneyTrajectoryMissingJourneyUUID(t *testing.T) {
	setup()
	defer teardown()

	out, resp, err := client.GradientAI.GetSimulationJourneyTrajectory(ctx, testRunUUID, "")
	assert.Error(t, err)
	assert.Nil(t, out)
	assert.Nil(t, resp)
}
