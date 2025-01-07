package dispatcher

import "github.com/kncept-oauth/simple-oidc/service/authorizer"

type DaoSource interface {
	GetClientStore() authorizer.ClientStore
}
