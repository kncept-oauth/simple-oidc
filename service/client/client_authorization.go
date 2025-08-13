package client

import (
	"context"
	"time"

	"github.com/kncept-oauth/simple-oidc/service/dao/ddbutil"
)

type ClientAuthorizationStore interface {
	ClientAuthorizationsByUser(ctx context.Context, userId string, scroller ddbutil.SimpleScroller[ClientAuthorization]) error

	// could easily be a very large amount.
	ClientAuthorizationsByClient(ctx context.Context, clientId string, scroller ddbutil.SimpleScroller[ClientAuthorization]) error

	GetClientAuthorization(ctx context.Context, userId string, clientId string) (*ClientAuthorization, error)
	SaveClientAuthorization(ctx context.Context, clientAuthorization *ClientAuthorization) error
	DeleteClientAuthorization(ctx context.Context, userId string, clientId string) error
}

type ClientAuthorization struct {
	UserId          string    `dynamodbav:"userId"`
	ClientId        string    `dynamodbav:"clientId"`
	AuthorizedAt    time.Time `dynamodbav:"authorizedAt"`
	LastRefreshedAt time.Time `dynamodbav:"lastRefreshedAt"`
}
