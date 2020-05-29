package godo

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testOneClick = &OneClick{
	Slug: "test-slug",
	Type: "droplet",
}

var testOneClickJSON = `
    {
      "slug":"test-slug",
      "type":"droplet"
    }
`

func TestOneClick_List(t *testing.T) {
	setup()
	defer teardown()

	svc := client.OneClick
	path := "/v2/1-click"
	want := []*OneClick{
		testOneClick,
	}

	jsonBlob := `
{
  "list": [
` + testOneClickJSON + `
  ]
}
`
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, jsonBlob)
	})

	got, _, err := svc.List(ctx, "")
	require.NoError(t, err)
	assert.Equal(t, want, got)
}
