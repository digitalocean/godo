package godo

import "testing"

func TestV2ClientWired(t *testing.T) {
	c := NewClient(nil)
	if c.V2 == nil {
		t.Fatal("Client.V2 should be wired")
	}
}
