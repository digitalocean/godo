package v2

// Client is the OpenAPI-generated API surface.
// Populate with: make generate
type Client struct{}

// New returns a v2 Client backed by inv.
func New(inv Invoker) *Client {
	_ = inv
	return &Client{}
}
