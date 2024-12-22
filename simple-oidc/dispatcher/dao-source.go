package dispatcher

import "github.com/kncept-oauth/simple-oidc/authorizer"

type DaoSource interface {
	GetClientStore() authorizer.ClientStore
}
