package authorizer

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/kncept-oauth/simple-oidc/service/client"
	"github.com/kncept-oauth/simple-oidc/service/dispatcherauth"
	"github.com/kncept-oauth/simple-oidc/service/gen/api"
)

func NewAuthorizer(store client.ClientStore) Authorizer {
	return Authorizer{
		store: store,
	}
}

type Authorizer struct {
	store client.ClientStore
}

func (obj Authorizer) AuthorizeGet(ctx context.Context, params api.AuthorizeGetParams) (api.AuthorizeGetRes, error) {
	client, err := obj.store.GetClient(ctx, params.ClientID)
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

	// fetch simple-oidc(soidc) state cookie
	loginCookie := dispatcherauth.GetLoginCookie(ctx)
	if loginCookie != "" {
		// validate
		// handle logged in (must click to approve access)
	}

	// redirect to /accept endpoint - no zero click logins are allowed
	redirectLocation := fmt.Sprintf(
		"/accept?response_type=%s&client_id=%s&scope=%s&redirect_uri=%s",
		params.ResponseType,
		params.ClientID,
		params.Scope,
		params.RedirectURI,
	)
	if params.State.Set {
		redirectLocation = fmt.Sprintf("%s&state=%s", redirectLocation, params.State.Value)
	}
	if params.Nonce.Set {
		redirectLocation = fmt.Sprintf("%s&nonce=%s", redirectLocation, params.Nonce.Value)
	}
	return &api.AuthorizeGetFound{
		Location: redirectLocation,
	}, nil

	// redirect to a login/auth page
}

func isValidRedirectUri(client *client.Client, redirectUri string) bool {
	if client.AllowRegexForRedirectUri {
		// TOOD: Consider caching?
		for _, allowedRedirectUri := range client.AllowedRedirectUris {
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
	for _, allowedRedirectUri := range client.AllowedRedirectUris {
		if strings.HasPrefix(redirectUri, allowedRedirectUri) {
			return true
		}
	}

	return false
}
