package dispatcher

import (
	"context"
	"fmt"

	"github.com/kncept-oauth/simple-oidc/service/authorizer"
	"github.com/kncept-oauth/simple-oidc/service/dao"
	"github.com/kncept-oauth/simple-oidc/service/gen/api"
	"github.com/kncept-oauth/simple-oidc/service/webcontent"
)

type oapiDispatcher struct {
	authorizer authorizer.Authorizer
	Issuer     string
	daoSource  dao.DaoSource
}

func (obj *oapiDispatcher) AuthorizeGet(ctx context.Context, params api.AuthorizeGetParams) (api.AuthorizeGetRes, error) {
	return obj.authorizer.AuthorizeGet(ctx, params)
}

func (obj *oapiDispatcher) Index(ctx context.Context) (api.IndexOK, error) {
	file, err := webcontent.Fs.Open("index.html")
	if err != nil {
		return api.IndexOK{}, err
	}
	return api.IndexOK{
		Data: file,
	}, nil
}

func (obj *oapiDispatcher) Jwks(ctx context.Context) (*api.JWKSetResponse, error) {
	keyStore := obj.daoSource.GetKeyStore()
	keys, err := keyStore.ListKeys()
	if err != nil {
		return nil, err
	}
	responseKeys := make([]api.JWKResponse, 0)
	for _, kid := range keys {
		key, err := keyStore.GetKey(kid)
		if err != nil {
			return nil, err
		}
		jwkKey := key.ToJwkDetails()
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
func (obj *oapiDispatcher) OpenIdConfiguration(ctx context.Context) (*api.OpenIDProviderMetadataResponse, error) {
	return &api.OpenIDProviderMetadataResponse{
		Issuer:                obj.Issuer,
		AuthorizationEndpoint: "/authorize",
		TokenEndpoint:         "todo",
		JwksURI:               "/.well-known/jwks.json",
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
