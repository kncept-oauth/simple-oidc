package dispatcher

import (
	"context"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/kncept-oauth/simple-oidc/service/client"
	"github.com/kncept-oauth/simple-oidc/service/dao"
	"github.com/kncept-oauth/simple-oidc/service/dispatcherauth"
	"github.com/kncept-oauth/simple-oidc/service/gen/api"
	"github.com/kncept-oauth/simple-oidc/service/jwtutil"
	"github.com/kncept-oauth/simple-oidc/service/keys"
	"github.com/kncept-oauth/simple-oidc/service/params"
)

type authorizationHandler struct {
	DaoSource dao.DaoSource
	Issuer    string
}

// TokenPost implements api.AuthorizationHandler.
func (obj *authorizationHandler) TokenPost(ctx context.Context, req api.TokenPostReq) (api.TokenPostRes, error) {
	var tokenRequestBody *api.TokenRequestBody

	if jsonReq, ok := req.(*api.TokenPostApplicationJSON); ok {
		tokenRequestBody = (*api.TokenRequestBody)(jsonReq)
	}
	if formReq, ok := req.(*api.TokenPostApplicationXWwwFormUrlencoded); ok {
		tokenRequestBody = (*api.TokenRequestBody)(formReq)
	}
	if tokenRequestBody == nil {
		return nil, errors.New("unable to parse request body")
	}

	authCode, err := obj.DaoSource.GetAuthorizationCodeStore().GetAuthorizationCode(ctx, tokenRequestBody.Code)
	if err != nil {
		return nil, err
	}

	// obj.DaoSource.GetAuthorizationCodeStore().DeleteAuthorizationCode(ctx, tokenRequestBody.Code)
	acParams, err := params.OidcParamsFromQuery(authCode.OidcParams)
	if err != nil {
		return nil, err
	}

	keyPair, err := keys.GetCurrentKey(obj.DaoSource.GetKeyStore())
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC().Truncate(time.Second)
	claims := &jwtutil.IdClaimsJwt{
		Iss: obj.Issuer,
		Sub: authCode.UserId,
		Aud: acParams.ClientId, // TODO: use client configuration for audiences here
		Nbf: now.Unix(),
		Iat: now.Unix(),
		Exp: now.Add(time.Hour * 3).Unix(),
	}
	jwt, err := jwtutil.ClaimsToJwt(claims, keyPair.Kid, keyPair.Rsa)
	if err != nil {
		return nil, err
	}

	loginTokens := &api.LoginTokens{
		IDToken:     jwt,
		AccessToken: jwt,
		// RefreshToken: "refresh",
	}

	return loginTokens, nil
}
func (obj authorizationHandler) AuthorizeGet(ctx context.Context, params api.AuthorizeGetParams) (api.AuthorizeGetRes, error) {
	client, err := obj.DaoSource.GetClientStore().GetClient(ctx, params.ClientID)
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
