package dispatcher

import (
	"github.com/kncept-oauth/simple-oidc/service/authorizer"
	"github.com/kncept-oauth/simple-oidc/service/keys"
	"github.com/kncept-oauth/simple-oidc/service/users"
)

type DaoSource interface {
	GetClientStore() authorizer.ClientStore
	GetKeyStore() keys.Keystore
	GetUserStore() users.UserStore
}
