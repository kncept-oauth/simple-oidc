package dao

import (
	"context"

	"github.com/kncept-oauth/simple-oidc/service/client"
	"github.com/kncept-oauth/simple-oidc/service/keys"
	"github.com/kncept-oauth/simple-oidc/service/session"
	"github.com/kncept-oauth/simple-oidc/service/users"
)

type DaoSource interface {
	GetClientStore(ctx context.Context) client.ClientStore                           // clients
	GetClientAuthorizationStore(ctx context.Context) client.ClientAuthorizationStore // client sessions
	GetAuthorizationCodeStore(ctx context.Context) client.AuthorizationCodeStore     // user-client authorizations
	GetKeyStore(ctx context.Context) keys.Keystore                                   // encryption keys
	GetUserStore(ctx context.Context) users.UserStore                                // simple-oidc registered users. TODO: Move to a "User" and "UserAuth" model
	GetSessionStore(ctx context.Context) session.SessionStore                        // simple-oidc sessions (not connected to auth clients)
}
