package godo

import (
	"context"
	"errors"
	"strings"

	"golang.org/x/oauth2"
)

const (
	// defaultOrcaRouterBaseURL is the OpenAI-compatible gateway endpoint
	// exposed by OrcaRouter.
	defaultOrcaRouterBaseURL = "https://api.orcarouter.ai/"
)

// OrcaRouterClient is a client for OrcaRouter, an OpenAI-compatible AI gateway.
// It embeds a *Client whose inference services (Chat, Messages, Models,
// Embeddings, Responses, ...) are resolved against OrcaRouter's endpoint using
// an OrcaRouter API key.
//
// See https://www.orcarouter.ai for more information.
type OrcaRouterClient struct {
	*Client
}

// NewOrcaRouterClient returns a client for OrcaRouter's OpenAI-compatible
// gateway. apiKey is an OrcaRouter API key. The caller may pass additional
// ClientOpt values, for example WithRetryAndBackoffs to enable automatic
// retries or SetBaseURL to target a self-hosted OrcaRouter instance.
func NewOrcaRouterClient(apiKey string, opts ...ClientOpt) (*OrcaRouterClient, error) {
	cleanKey := strings.TrimSpace(apiKey)
	if cleanKey == "" {
		return nil, errors.New("orcarouter: api key is required")
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: cleanKey})
	oauthClient := oauth2.NewClient(ctx, ts)

	client, err := New(oauthClient, append([]ClientOpt{
		SetBaseURL(defaultOrcaRouterBaseURL),
	}, opts...)...)
	if err != nil {
		return nil, err
	}

	// Point the shared OpenAI-compatible inference services at OrcaRouter's
	// endpoint instead of DigitalOcean's Serverless Inference API.
	newInferenceServices(client, client.BaseURL)

	return &OrcaRouterClient{Client: client}, nil
}
