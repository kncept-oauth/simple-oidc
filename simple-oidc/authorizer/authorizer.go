package authorizer

import (
	"context"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/kncept-oauth/simple-oidc/gen/api"
)

type Authorizer struct {
	store ClientStore
}

func (obj Authorizer) AuthorizeGet(ctx context.Context, params api.AuthorizeGetParams) (api.AuthorizeGetRes, error) {
	client, err := obj.store.GetClient(params.ClientID)
	if err != nil {
		return nil, err
	}
	if client == nil {
		return nil, fmt.Errorf("no such client " + params.ClientID)
	}
	fmt.Printf("client: %v\n", client)

	// TODO: what to do with this
	// params.ResponseType // id_token token

	// validate allowed scopes
	// allowedScopes := client.GetAllowedScopes()
	// if allowedScopes != nil && len(allowedScopes) != 0 {
	// }

	redirectUri := params.RedirectURI
	if !isValidRedirectUri(client, redirectUri) {
		return nil, fmt.Errorf("invalid redirect uri " + redirectUri)
	}

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
