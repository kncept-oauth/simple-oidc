package oapidispatcher

import (
	"context"
	"fmt"

	"github.com/kncept-oauth/simple-oidc/service/dao"
	"github.com/kncept-oauth/simple-oidc/service/gen/api"
)

type oapiDispatcher struct {
	authorizationHandler
	wellKnownHandler
	userInfoHandler
}

func NewOapiDispatcher(
	daoSource dao.DaoSource,
	urlPrefix string,
) api.Handler {
	return &oapiDispatcher{
		authorizationHandler: authorizationHandler{
			DaoSource: daoSource,
			Issuer:    urlPrefix,
		},
		wellKnownHandler: wellKnownHandler{
			DaoSource: daoSource,
			Issuer:    urlPrefix,
		},
		userInfoHandler: userInfoHandler{
			DaoSource: daoSource,
			Issuer:    urlPrefix,
		},
	}
}

func (obj *oapiDispatcher) NewError(ctx context.Context, err error) *api.ErrRespStatusCode {
	fmt.Printf("General error occurred: %v\n", err)
	return &api.ErrRespStatusCode{
		StatusCode: 500,
		Response:   fmt.Sprintf("%v", err),
	}
}
