package v2

import (
	"net/http"
	"net/url"
	"reflect"

	"github.com/google/go-querystring/query"
)

// ListOptions specifies optional pagination parameters for List methods.
type ListOptions struct {
	Page    int `url:"page,omitempty"`
	PerPage int `url:"per_page,omitempty"`
}

// Response wraps the HTTP response and optional list metadata.
type Response struct {
	*http.Response
	Links *Links
	Meta  *Meta
}

// Meta describes generic information about a response.
type Meta struct {
	Total int `json:"total"`
}

// Links manages links returned with a list response.
type Links struct {
	Pages *Pages `json:"pages,omitempty"`
}

// Pages are pages specified in Links.
type Pages struct {
	First string `json:"first,omitempty"`
	Prev  string `json:"prev,omitempty"`
	Last  string `json:"last,omitempty"`
	Next  string `json:"next,omitempty"`
}

func addOptions(s string, opt interface{}) (string, error) {
	v := reflect.ValueOf(opt)
	if v.Kind() == reflect.Ptr && v.IsNil() {
		return s, nil
	}

	origURL, err := url.Parse(s)
	if err != nil {
		return s, err
	}

	origValues := origURL.Query()
	newValues, err := query.Values(opt)
	if err != nil {
		return s, err
	}
	for k, vals := range newValues {
		origValues[k] = vals
	}
	origURL.RawQuery = origValues.Encode()
	return origURL.String(), nil
}
