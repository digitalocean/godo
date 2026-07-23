package godo

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

func TestMicroDropletImages_List(t *testing.T) {
	setup()
	defer teardown()

	jBlob := `{
		"images": [
			{"id": "img-1", "name": "python-3.12", "source": "docker.io/library/python:3.12", "status": "IMAGE_AVAILABLE"},
			{"id": "img-2", "name": "node-20",     "source": "docker.io/library/node:20",     "status": "IMAGE_IMPORTING"}
		],
		"meta": {"total": 2}
	}`

	mux.HandleFunc("/v2/microdroplets/images", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, jBlob)
	})

	images, resp, err := client.MicroDropletImages.List(ctx, nil)
	if err != nil {
		t.Fatalf("MicroDropletImages.List returned error: %v", err)
	}

	expected := []MicroDropletImage{
		{ID: "img-1", Name: "python-3.12", Source: "docker.io/library/python:3.12", Status: MicroDropletImageStatusAvailable},
		{ID: "img-2", Name: "node-20", Source: "docker.io/library/node:20", Status: MicroDropletImageStatusImporting},
	}
	if !reflect.DeepEqual(images, expected) {
		t.Errorf("MicroDropletImages.List returned %+v, expected %+v", images, expected)
	}

	if resp.Meta == nil || resp.Meta.Total != 2 {
		t.Errorf("MicroDropletImages.List Meta not propagated: %+v", resp.Meta)
	}
}

func TestMicroDropletImages_List_Paginated(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/microdroplets/images", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		if got, want := r.URL.Query().Get("page"), "3"; got != want {
			t.Errorf("page query = %q, expected %q", got, want)
		}
		if got, want := r.URL.Query().Get("per_page"), "10"; got != want {
			t.Errorf("per_page query = %q, expected %q", got, want)
		}
		fmt.Fprint(w, `{"images": [], "meta": {"total": 0}}`)
	})

	if _, _, err := client.MicroDropletImages.List(ctx, &ListOptions{Page: 3, PerPage: 10}); err != nil {
		t.Fatalf("MicroDropletImages.List returned error: %v", err)
	}
}

func TestMicroDropletImages_Get(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/microdroplets/images/img-1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"image": {
				"id": "img-1",
				"name": "python-3.12",
				"source": "docker.io/library/python:3.12",
				"status": "IMAGE_AVAILABLE",
				"created_at": "2026-07-16T10:00:00Z"
			}
		}`)
	})

	image, _, err := client.MicroDropletImages.Get(ctx, "img-1")
	if err != nil {
		t.Fatalf("MicroDropletImages.Get returned error: %v", err)
	}

	expected := &MicroDropletImage{
		ID:      "img-1",
		Name:    "python-3.12",
		Source:  "docker.io/library/python:3.12",
		Status:  MicroDropletImageStatusAvailable,
		Created: "2026-07-16T10:00:00Z",
	}
	if !reflect.DeepEqual(image, expected) {
		t.Errorf("MicroDropletImages.Get returned %+v, expected %+v", image, expected)
	}
}

func TestMicroDropletImages_Get_EmptyID(t *testing.T) {
	_, _, err := (&MicroDropletImagesServiceOp{}).Get(ctx, "")
	if err == nil {
		t.Fatal("expected error for empty id")
	}
	if _, ok := err.(*ArgError); !ok {
		t.Errorf("expected *ArgError, got %T: %v", err, err)
	}
}

func TestMicroDropletImages_Create(t *testing.T) {
	setup()
	defer teardown()

	createRequest := &MicroDropletImageCreateRequest{
		Name:   "python-3.12",
		Source: "docker.io/library/python:3.12",
	}

	mux.HandleFunc("/v2/microdroplets/images", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)

		expected := map[string]interface{}{
			"name":   "python-3.12",
			"source": "docker.io/library/python:3.12",
		}
		var got map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if !reflect.DeepEqual(got, expected) {
			t.Errorf("Create body\n got=%#v\nwant=%#v", got, expected)
		}

		fmt.Fprint(w, `{"image": {"id": "img-1", "name": "python-3.12", "status": "IMAGE_IMPORTING"}}`)
	})

	image, _, err := client.MicroDropletImages.Create(ctx, createRequest)
	if err != nil {
		t.Fatalf("MicroDropletImages.Create returned error: %v", err)
	}

	if image.ID != "img-1" {
		t.Errorf("MicroDropletImages.Create returned ID %q, expected %q", image.ID, "img-1")
	}
	if image.Status != MicroDropletImageStatusImporting {
		t.Errorf("MicroDropletImages.Create returned Status %q, expected %q", image.Status, MicroDropletImageStatusImporting)
	}
}

func TestMicroDropletImages_Create_NilRequest(t *testing.T) {
	_, _, err := (&MicroDropletImagesServiceOp{}).Create(ctx, nil)
	if err == nil {
		t.Fatal("expected error for nil createRequest")
	}
	if _, ok := err.(*ArgError); !ok {
		t.Errorf("expected *ArgError, got %T: %v", err, err)
	}
}

func TestMicroDropletImages_Delete(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/microdroplets/images/img-1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		w.WriteHeader(http.StatusNoContent)
	})

	if _, err := client.MicroDropletImages.Delete(ctx, "img-1"); err != nil {
		t.Fatalf("MicroDropletImages.Delete returned error: %v", err)
	}
}

func TestMicroDropletImages_Delete_EmptyID(t *testing.T) {
	_, err := (&MicroDropletImagesServiceOp{}).Delete(ctx, "")
	if err == nil {
		t.Fatal("expected error for empty id")
	}
	if _, ok := err.(*ArgError); !ok {
		t.Errorf("expected *ArgError, got %T: %v", err, err)
	}
}

func TestMicroDropletImage_URN(t *testing.T) {
	img := MicroDropletImage{ID: "img-1"}
	want := "do:microdropletimage:img-1"
	if got := img.URN(); got != want {
		t.Errorf("MicroDropletImage.URN = %q, expected %q", got, want)
	}
}
