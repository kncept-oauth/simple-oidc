package dao

import (
	"github.com/kncept-oauth/simple-oidc/service/client"
	"github.com/kncept-oauth/simple-oidc/service/keys"
	"github.com/kncept-oauth/simple-oidc/service/session"
	"github.com/kncept-oauth/simple-oidc/service/users"
)

type DaoSource interface {
	GetClientStore() client.ClientStore                           // clients
	GetClientAuthorizationStore() client.ClientAuthorizationStore // client sessions

	GetAuthorizationCodeStore() client.AuthorizationCodeStore

	GetKeyStore() keys.Keystore            // encryption keys
	GetUserStore() users.UserStore         // simple-oidc registered users. TODO: Move to a "User" and "UserAuth" model
	GetSessionStore() session.SessionStore // simple-oidc sessions (not connected to auth clients)
}
