package client

import "context"

type Client struct {
	ClientId string

	// can be an empty array (all scopes)
	AllowedScopes []string

	// treat the response from GetAllowedRedirectUris as a regex?
	// RECOMMENDED TO BE FALSE
	AllowRegexForRedirectUri bool

	// regex scripts for redirect uris
	AllowedRedirectUris []string
}

type ClientStore interface {
	GetClient(ctx context.Context, clientId string) (*Client, error)
	SaveClient(ctx context.Context, client *Client) error
	ListClients(ctx context.Context) ([]*Client, error)
	RemoveClient(ctx context.Context, clientId string) error
}
