package v2

import (
	"context"
	"net/http"
)

// Invoker is the HTTP transport used by generated services.
// The root godo.Client satisfies this via a thin adapter.
type Invoker interface {
	NewRequest(ctx context.Context, method, urlStr string, body interface{}) (*http.Request, error)
	Do(ctx context.Context, req *http.Request, v interface{}) (*Response, error)
}
