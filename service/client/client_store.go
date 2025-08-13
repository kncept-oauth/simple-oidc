package client

import "context"

type Client struct {
	ClientId string `dynamodbav:"clientId"` // assigned guid

	// can be an empty array (all scopes)
	AllowedScopes []string `dynamodbav:"allowedScopes"`

	// treat the response from GetAllowedRedirectUris as a regex?
	// RECOMMENDED TO BE FALSE
	AllowRegexForRedirectUri bool `dynamodbav:"regexForRedirectUri"`

	// regex scripts for redirect uris
	AllowedRedirectUris []string `dynamodbav:"allowedRedirectUris"`

	PublicName    string `dynamodbav:"publicName"`
	PublicWebsite string `dynamodbav:"publicWebsite"`
	Description   string `dynamodbav:"description"`
}

type ClientStore interface {
	GetClient(ctx context.Context, clientId string) (*Client, error)
	SaveClient(ctx context.Context, client *Client) error
	ListClients(ctx context.Context) ([]*Client, error)
	RemoveClient(ctx context.Context, clientId string) error
}
