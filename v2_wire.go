package godo

import (
	"context"
	"net/http"

	"github.com/digitalocean/godo/v2"
)

// v2Invoker adapts *Client to v2.Invoker without an import cycle in package v2.
type v2Invoker struct {
	c *Client
}

func (a *v2Invoker) NewRequest(ctx context.Context, method, urlStr string, body interface{}) (*http.Request, error) {
	return a.c.NewRequest(ctx, method, urlStr, body)
}

func (a *v2Invoker) Do(ctx context.Context, req *http.Request, v interface{}) (*v2.Response, error) {
	resp, err := a.c.Do(ctx, req, v)
	if resp == nil {
		return nil, err
	}
	return &v2.Response{Response: resp.Response}, err
}

func newV2Client(c *Client) *v2.Client {
	return v2.New(&v2Invoker{c: c})
}
