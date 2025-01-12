package authorizer

import (
	"context"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/kncept-oauth/simple-oidc/service/dispatcherauth"
	"github.com/kncept-oauth/simple-oidc/service/gen/api"
)

func NewAuthorizer(store ClientStore) Authorizer {
	return Authorizer{
		store: store,
	}
}

type Authorizer struct {
	store ClientStore
}

func (obj Authorizer) AuthorizeGet(ctx context.Context, params api.AuthorizeGetParams) (api.AuthorizeGetRes, error) {
	client, err := obj.store.Get(params.ClientID)
	if err != nil {
		return nil, err
	}
	if client == nil {
		return nil, fmt.Errorf("no such client: %v", params.ClientID)
	}

	// TODO: what to do with this
	// params.ResponseType // id_token token     or     code

	// validate allowed scopes
	// allowedScopes := client.GetAllowedScopes()
	// if allowedScopes != nil && len(allowedScopes) != 0 {
	// }

	redirectUri := params.RedirectURI
	if !isValidRedirectUri(client, redirectUri) {
		return nil, fmt.Errorf("invalid redirect uri: %v", redirectUri)
	}

	fmt.Printf("now what\n")

	// fetch simple-oidc(soidc) state cookie
	loginCookie := dispatcherauth.GetLoginCookie(ctx)
	if loginCookie != "" {
		// validate
		// handle logged in (must click to approve access)
	}
	// if they don't, then make them loginregister (TODO: according to client config)

	// redirect to a login/auth page

	return nil, errors.ErrUnsupported
}

func isValidRedirectUri(client Client, redirectUri string) bool {

	if client.IsAllowedRedirectUriRegex() {
		// TOOD: Consider caching?
		for _, allowedRedirectUri := range client.GetAllowedRedirectUris() {
			// regex match https://pkg.go.dev/regexp
			r, err := regexp.Compile(allowedRedirectUri)
			if err != nil {
				log.Printf("%v\n", err)
				continue
			}
			if r.Match([]byte(redirectUri)) {
				return true
			}
		}
		return false
	}

	// else this is just a 'starts with'... MUCH simpler
	for _, allowedRedirectUri := range client.GetAllowedRedirectUris() {
		if strings.HasPrefix(redirectUri, allowedRedirectUri) {
			return true
		}
	}

	return false
}
