package dispatcher

import (
	"context"
	"fmt"

	"github.com/kncept-oauth/simple-oidc/service/dao"
	"github.com/kncept-oauth/simple-oidc/service/gen/api"
)

type oapiDispatcher struct {
	authorizationHandler
	wellKnownHandler
}

type wellKnownHandler struct {
	DaoSource dao.DaoSource
	Issuer    string
}

func (obj *wellKnownHandler) Jwks(ctx context.Context) (*api.JWKSetResponse, error) {
	keyStore := obj.DaoSource.GetKeyStore(ctx)
	keys, err := keyStore.ListKeys(ctx)
	if err != nil {
		return nil, err
	}
	responseKeys := make([]api.JWKResponse, 0)
	for _, key := range keys {
		jwkKey, err := key.ToJwkDetails()
		if err != nil {
			return nil, err
		}
		switch jwkKey.Kty {
		case "RSA":
			responseKeys = append(responseKeys, api.JWKResponse{
				Kty: api.NewOptString(jwkKey.Kty),
				Kid: api.NewOptString(jwkKey.Kid),
				Use: api.NewOptString(jwkKey.Use),
				Alg: api.NewOptString(jwkKey.Alg),
				N:   api.NewOptString(jwkKey.N),
				E:   api.NewOptString(jwkKey.E),
			})
		default:
			panic("unable to make key type of " + key.Kty)
		}

	}
	return &api.JWKSetResponse{
		Keys: responseKeys,
	}, nil
}

// additional rURL eferences and examples:
// https://appleid.apple.com/.well-known/openid-configuration
// https://accounts.google.com/.well-known/openid-configuration
func (obj *wellKnownHandler) OpenIdConfiguration(ctx context.Context) (*api.OpenIDProviderMetadataResponse, error) {
	return &api.OpenIDProviderMetadataResponse{
		Issuer:                obj.Issuer,
		AuthorizationEndpoint: fmt.Sprintf("%v/authorize", obj.Issuer),
		TokenEndpoint:         fmt.Sprintf("%v/token", obj.Issuer),
		JwksURI:               fmt.Sprintf("%v/.well-known/jwks.json", obj.Issuer),
	}, nil

	// fmt.Printf("TODO: OpenIdConfiguration\n")
	// return nil, errors.ErrUnsupported
}

func (obj *oapiDispatcher) NewError(ctx context.Context, err error) *api.ErrRespStatusCode {
	fmt.Printf("General error occurred: %v\n", err)
	return &api.ErrRespStatusCode{
		StatusCode: 500,
		Response:   fmt.Sprintf("%v", err),
	}
}
