package oapidispatcher

import (
	"context"
	"crypto/rsa"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/kncept-oauth/simple-oidc/service/client"
	"github.com/kncept-oauth/simple-oidc/service/dao"
	"github.com/kncept-oauth/simple-oidc/service/dispatcherauth"
	"github.com/kncept-oauth/simple-oidc/service/gen/api"
	"github.com/kncept-oauth/simple-oidc/service/jwtutil"
	"github.com/kncept-oauth/simple-oidc/service/keys"
	"github.com/kncept-oauth/simple-oidc/service/params"
	"github.com/kncept-oauth/simple-oidc/service/session"
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

	grantType := strings.ToLower(tokenRequestBody.GrantType.Or(""))
	grantPayload := ""
	// validate requried sets
	switch grantType {
	case "authorization_code":
		// code must be set
		grantPayload = tokenRequestBody.Code.Or("")
		if grantPayload == "" {
			return nil, fmt.Errorf("query parameter \"code\" not set")
		}
	case "refresh_token":
		grantPayload = tokenRequestBody.RefreshToken.Or("")
		if grantPayload == "" {
			return nil, fmt.Errorf("query parameter \"code\" not set")
		}
	}

	ses, err := obj.mapToSession(ctx, grantType, grantPayload)
	if err != nil {
		return nil, err
	}
	if ses == nil {
		return nil, fmt.Errorf("Session Not Found")
	}

	keyPair, err := keys.GetCurrentKey(ctx, obj.DaoSource.GetKeyStore(ctx))
	if err != nil {
		return nil, err
	}

	decodedKey, err := keyPair.DecodePrivateKey()
	if err != nil {
		return nil, err
	}
	rsaKey, isRsaKey := decodedKey.(*rsa.PrivateKey)
	if !isRsaKey {
		return nil, errors.New("not an rsa key")
	}

	idToken, refreshToken := ses.IssueTokens(obj.Issuer, ses.ClientId)
	sessionStore := obj.DaoSource.GetSessionStore(ctx)
	err = sessionStore.SaveSession(ctx, ses)
	if err != nil {
		return nil, err
	}

	idTokenJwt, err := jwtutil.ClaimsToJwt(idToken, keyPair.Kid, rsaKey)
	if err != nil {
		return nil, err
	}

	refreshTokenJwt, err := jwtutil.ClaimsToJwt(refreshToken, keyPair.Kid, rsaKey)
	if err != nil {
		return nil, err
	}
	return &api.LoginTokensHeaders{
		AccessControlAllowOrigin: api.NewOptString("*"),
		Response: api.LoginTokens{
			IDToken:      idTokenJwt,
			AccessToken:  idTokenJwt,
			RefreshToken: refreshTokenJwt,
		},
	}, nil

}

func (obj *authorizationHandler) mapToSession(ctx context.Context, grantType string, grantPayload string) (*session.Session, error) {
	switch grantType {
	case "authorization_code":
		authCode, err := obj.DaoSource.GetAuthorizationCodeStore(ctx).GetAuthorizationCode(ctx, grantPayload)
		if err != nil {
			return nil, err
		}

		// obj.DaoSource.GetAuthorizationCodeStore().DeleteAuthorizationCode(ctx, tokenRequestBody.Code)
		acParams, err := params.OidcParamsFromQuery(authCode.OidcParams)
		if err != nil {
			return nil, err
		}
		userId := authCode.UserId
		clientId := acParams.ClientId

		ses, err := session.NewSession(userId, clientId)
		if err != nil {
			return nil, err
		}
		return ses, nil
	case "refresh_token":
		refreshClaims := &jwtutil.RefreshClaimsJwt{}
		err := jwtutil.ParseJwt(ctx, grantPayload, obj.DaoSource.GetKeyStore(ctx), obj.Issuer, refreshClaims)
		if err != nil {
			return nil, err
		}

		userId := refreshClaims.Sub
		sessionId := refreshClaims.Ses
		// clientId := refreshClaims.au

		ses, err := obj.DaoSource.GetSessionStore(ctx).LoadSession(ctx, sessionId, userId)
		if err != nil {
			return nil, err
		}
		if ses == nil {
			return nil, nil
		}

		if ses.RefreshCode != refreshClaims.Code {
			fmt.Printf("Refresh Code Mismatch:\nses %v\njwt %v\n", ses.RefreshCode, refreshClaims.Code)
			return nil, fmt.Errorf("refresh code mismatch")
		}
		return ses, nil
	}
	return nil, fmt.Errorf("Unknown grant type: %s", grantType)
}

func (obj authorizationHandler) AuthorizeGet(ctx context.Context, params api.AuthorizeGetParams) (api.AuthorizeGetRes, error) {
	client, err := obj.DaoSource.GetClientStore(ctx).GetClient(ctx, params.ClientID)
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
