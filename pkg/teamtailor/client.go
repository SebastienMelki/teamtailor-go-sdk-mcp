// Package teamtailor is the public Go SDK for the Teamtailor API
// (https://docs.teamtailor.com/).
//
// It wraps the generated client under internal/gen with Teamtailor-specific
// defaults: regional base URLs, Token-auth header formatting, default
// X-Api-Version, and JSON:API content negotiation
// (application/vnd.api+json).
//
// Typical use:
//
//	c := teamtailor.NewClient(os.Getenv("TEAMTAILOR_API_KEY"))
//	resp, err := c.ListCandidates(ctx, &teamtailorv1.ListCandidatesRequest{
//	    PageSize: 10,
//	})
package teamtailor

import (
	teamtailorv1 "github.com/sebastienmelki/teamtailor-go-sdk-mcp/internal/gen/teamtailor/v1"
)

// ContentTypeJSONAPI is the JSON:API media type Teamtailor speaks.
const ContentTypeJSONAPI = "application/vnd.api+json"

// Client is a typed Teamtailor API client.
//
// It embeds the generated TeamtailorServiceClient interface, so all RPC
// methods (ListCandidates, GetCandidate, ...) are available directly on
// *Client.
type Client struct {
	teamtailorv1.TeamtailorServiceClient
}

// NewClient builds a Client authenticated with the given API key.
// The key is sent as "Authorization: Token token=<apiKey>" on every request.
//
// Defaults: EU region, X-Api-Version=20240904, http.DefaultClient,
// Content-Type/Accept=application/vnd.api+json.
func NewClient(apiKey string, opts ...Option) *Client {
	cfg := defaultOptions()
	for _, opt := range opts {
		opt(cfg)
	}

	inner := teamtailorv1.NewTeamtailorServiceClient(
		cfg.resolveBaseURL(),
		teamtailorv1.WithTeamtailorServiceHTTPClient(cfg.httpClient),
		teamtailorv1.WithTeamtailorServiceAuthorization("Token token="+apiKey),
		teamtailorv1.WithTeamtailorServiceApiVersion(cfg.apiVersion),
		teamtailorv1.WithTeamtailorServiceContentType(ContentTypeJSONAPI),
		teamtailorv1.WithTeamtailorServiceAccept(ContentTypeJSONAPI),
		teamtailorv1.WithTeamtailorServiceDiscardUnknownFields(cfg.discardUnknownFields),
	)

	return &Client{TeamtailorServiceClient: inner}
}
