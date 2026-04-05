package dao

import (
	"context"

	"github.com/kncept-oauth/simple-oidc/service/client"
	"github.com/kncept-oauth/simple-oidc/service/keys"
	"github.com/kncept-oauth/simple-oidc/service/session"
	"github.com/kncept-oauth/simple-oidc/service/users"
)

type DaoSource interface {
	GetDaoSourceDescription() string // name, type, etc

	// simple-oidc registered users. TODO: Move to a "User" and "UserAuth" model
	GetUserStore(ctx context.Context) users.UserStore

	// Clients are consumer of the auth service
	GetClientStore(ctx context.Context) client.ClientStore

	// Mapping of users to clients
	GetClientAuthorizationStore(ctx context.Context) client.ClientAuthorizationStore // client sessions

	// OIDC Authorization Codes for user-client authorizations
	GetAuthorizationCodeStore(ctx context.Context) client.AuthorizationCodeStore

	// Encryption keys (currently RSA)
	GetKeyStore(ctx context.Context) keys.Keystore

	// this is sessions against the simple-oidc service
	// this is NOT connectd to client-authorizations
	// eg: multiple client-authorizatons can come from a single simple-oidc session
	// TODO: rename to GetSimpleOidcSessionStore
	GetSessionStore(ctx context.Context) session.SessionStore
}
