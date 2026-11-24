package godo

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"testing"
)

var (
	firstPageLinksJSONBlob = []byte(`{
		"links": {
			"pages": {
				"last": "https://api.digitalocean.com/v2/droplets/?page=3",
				"next": "https://api.digitalocean.com/v2/droplets/?page=2"
			}
		}
	}`)
	otherPageLinksJSONBlob = []byte(`{
		"links": {
			"pages": {
				"first": "https://api.digitalocean.com/v2/droplets/?page=1",
				"prev": "https://api.digitalocean.com/v2/droplets/?page=1",
				"last": "https://api.digitalocean.com/v2/droplets/?page=3",
				"next": "https://api.digitalocean.com/v2/droplets/?page=3"
			}
		}
	}`)
	lastPageLinksJSONBlob = []byte(`{
		"links": {
			"pages": {
				"first": "https://api.digitalocean.com/v2/droplets/?page=1",
				"prev": "https://api.digitalocean.com/v2/droplets/?page=2"
			}
		}
	}`)
	projectsLastPageLinksJSONBlob = []byte(`{
		"links": {
			"pages": {
				"first": "https://api.digitalocean.com/v2/projects?page=1",
				"prev": "https://api.digitalocean.com/v2/projects?page=2",
				"last": "https://api.digitalocean.com/v2/projects?page=3"
			}
		}
	}`)

	missingLinksJSONBlob = []byte(`{ }`)
)

type godoList struct {
	Links Links `json:"links"`
}

func loadLinksJSON(t *testing.T, j []byte) Links {
	var list godoList
	err := json.Unmarshal(j, &list)
	if err != nil {
		t.Fatal(err)
	}

	return list.Links
}

func TestLinks_ParseFirst(t *testing.T) {
	links := loadLinksJSON(t, firstPageLinksJSONBlob)
	_, err := links.CurrentPage()
	if err != nil {
		t.Fatal(err)
	}

	r := &Response{Links: &links}
	checkCurrentPage(t, r, 1)

	if links.IsLastPage() {
		t.Fatalf("shouldn't be last page")
	}
}

func TestLinks_ParseMiddle(t *testing.T) {
	links := loadLinksJSON(t, otherPageLinksJSONBlob)
	_, err := links.CurrentPage()
	if err != nil {
		t.Fatal(err)
	}

	r := &Response{Links: &links}
	checkCurrentPage(t, r, 2)

	if links.IsLastPage() {
		t.Fatalf("shouldn't be last page")
	}
}

func TestLinks_ParseLast(t *testing.T) {
	links := loadLinksJSON(t, lastPageLinksJSONBlob)
	_, err := links.CurrentPage()
	if err != nil {
		t.Fatal(err)
	}

	r := &Response{Links: &links}
	checkCurrentPage(t, r, 3)
	if !links.IsLastPage() {
		t.Fatalf("expected last page")
	}
}

func TestLinks_ParseProjectsLast(t *testing.T) {
	links := loadLinksJSON(t, projectsLastPageLinksJSONBlob)
	_, err := links.CurrentPage()
	if err != nil {
		t.Fatal(err)
	}

	r := &Response{Links: &links}
	checkCurrentPage(t, r, 3)
	if !links.IsLastPage() {
		t.Fatalf("expected last page")
	}
}

func TestLinks_ParseMissing(t *testing.T) {
	links := loadLinksJSON(t, missingLinksJSONBlob)
	_, err := links.CurrentPage()
	if err != nil {
		t.Fatal(err)
	}

	r := &Response{Links: &links}
	checkCurrentPage(t, r, 1)
}

func TestLinks_ParseURL(t *testing.T) {
	type linkTest struct {
		name, url         string
		expectedPage      int
		expectedPageToken string
	}

	linkTests := []linkTest{
		{
			name:         "prev",
			url:          "https://api.digitalocean.com/v2/droplets/?page=1",
			expectedPage: 1,
		},
		{
			name:         "last",
			url:          "https://api.digitalocean.com/v2/droplets/?page=5",
			expectedPage: 5,
		},
		{
			name:         "next",
			url:          "https://api.digitalocean.com/v2/droplets/?page=2",
			expectedPage: 2,
		},
		{
			name:              "page token",
			url:               "https://api.digitalocean.com/v2/droplets/?page=2&page_token=aaa",
			expectedPage:      2,
			expectedPageToken: "aaa",
		},
	}

	for _, lT := range linkTests {
		p, err := pageForURL(lT.url)
		if err != nil {
			t.Fatal(err)
		}

		if p != lT.expectedPage {
			t.Errorf("expected page for '%s' to be '%d', was '%d'",
				lT.url, lT.expectedPage, p)
		}

		pageToken, err := pageTokenFromURL(lT.url)
		if err != nil {
			t.Fatal(err)
		}
		if pageToken != lT.expectedPageToken {
			t.Errorf("expected pageToken for '%s' to be '%s', was '%s'",
				lT.url, lT.expectedPageToken, pageToken)
		}
	}

}

func TestLinks_ParseEmptyString(t *testing.T) {
	type linkTest struct {
		name, url string
		expected  int
	}

	linkTests := []linkTest{
		{
			name:     "none",
			url:      "http://example.com",
			expected: 0,
		},
		{
			name:     "bad",
			url:      "no url",
			expected: 0,
		},
		{
			name:     "empty",
			url:      "",
			expected: 0,
		},
	}

	for _, lT := range linkTests {
		_, err := pageForURL(lT.url)
		if err == nil {
			t.Fatalf("expected error for test '%s', but received none", lT.name)
		}
	}
}

func TestLinks_NextPageToken(t *testing.T) {
	t.Run("happy token", func(t *testing.T) {
		checkNextPageToken(t, &Response{Links: &Links{
			Pages: &Pages{
				Next: "https://api.digitalocean.com/v2/droplets/?page_token=aaa",
			},
		}}, "aaa")
	})
	t.Run("empty token", func(t *testing.T) {
		checkNextPageToken(t, &Response{Links: &Links{
			Pages: &Pages{
				Next: "https://api.digitalocean.com/v2/droplets/",
			},
		}}, "")
	})
	t.Run("no next page", func(t *testing.T) {
		checkNextPageToken(t, &Response{Links: &Links{
			Pages: &Pages{},
		}}, "")
	})
}

func TestLinkAction_UnmarshalOversizedID(t *testing.T) {
	// The API can return action IDs that do not fit into int, for example on
	// 32-bit platforms or once the counter outgrows 64-bit ints. Decoding
	// must not fail: the full ID stays available through HREF.
	jBlob := []byte(`{
		"links": {
			"actions": [
				{"id": 9223372036854775808, "rel": "create", "href": "https://api.digitalocean.com/v2/actions/9223372036854775808"}
			]
		}
	}`)

	var list godoList
	if err := json.Unmarshal(jBlob, &list); err != nil {
		t.Fatalf("unmarshal LinkAction with oversized id: %v", err)
	}

	if len(list.Links.Actions) != 1 {
		t.Fatalf("expected 1 link action, got %d", len(list.Links.Actions))
	}

	action := list.Links.Actions[0]
	if action.ID != 0 {
		t.Errorf("expected ID to stay 0 for an ID that overflows int, got %d", action.ID)
	}
	if action.Rel != "create" {
		t.Errorf("expected Rel %q, got %q", "create", action.Rel)
	}
	if action.HREF != "https://api.digitalocean.com/v2/actions/9223372036854775808" {
		t.Errorf("unexpected HREF %q", action.HREF)
	}
}

func TestLinkAction_UnmarshalLargeID(t *testing.T) {
	// 2771062950 exceeds 32-bit int; it fits int on 64-bit platforms.
	jBlob := []byte(`{
		"links": {
			"actions": [
				{"id": 2771062950, "rel": "create", "href": "https://api.digitalocean.com/v2/actions/2771062950"}
			]
		}
	}`)

	var list godoList
	if err := json.Unmarshal(jBlob, &list); err != nil {
		t.Fatalf("unmarshal LinkAction with large id: %v", err)
	}

	action := list.Links.Actions[0]
	if strconv.IntSize == 64 && action.ID != 2771062950 {
		t.Errorf("expected ID 2771062950 on a 64-bit platform, got %d", action.ID)
	}
	if strconv.IntSize == 32 && action.ID != 0 {
		t.Errorf("expected ID to stay 0 on a 32-bit platform, got %d", action.ID)
	}
}

func TestLinkAction_GetFallsBackToHREF(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/actions/9223372036854775808", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{"action": {"id": 9223372036854775808, "status": "completed", "type": "create"}}`)
	})

	// An action ID that overflowed int leaves ID at 0 but is still
	// reachable through the link's HREF.
	linkAction := &LinkAction{
		Rel:  "create",
		HREF: "https://api.digitalocean.com/v2/actions/9223372036854775808",
	}

	action, _, err := linkAction.Get(ctx, client)
	if err != nil {
		t.Fatalf("LinkAction.Get returned error: %v", err)
	}
	if action.Status != ActionCompleted {
		t.Errorf("expected status %q, got %q", ActionCompleted, action.Status)
	}
}

func TestLinks_ParseNextPageToken(t *testing.T) {
	t.Run("happy token", func(t *testing.T) {
		checkPreviousPageToken(t, &Response{Links: &Links{
			Pages: &Pages{
				Prev: "https://api.digitalocean.com/v2/droplets/?page_token=aaa",
			},
		}}, "aaa")
	})
	t.Run("empty token", func(t *testing.T) {
		checkPreviousPageToken(t, &Response{Links: &Links{
			Pages: &Pages{
				Prev: "https://api.digitalocean.com/v2/droplets/",
			},
		}}, "")
	})
	t.Run("no next page", func(t *testing.T) {
		checkPreviousPageToken(t, &Response{Links: &Links{
			Pages: &Pages{},
		}}, "")
	})
}
