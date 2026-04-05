package oapidispatcher

import (
	"context"
	"fmt"

	"github.com/kncept-oauth/simple-oidc/service/dao"
	"github.com/kncept-oauth/simple-oidc/service/dispatcherauth"
	"github.com/kncept-oauth/simple-oidc/service/gen/api"
	"github.com/kncept-oauth/simple-oidc/service/jwtutil"
)

type userInfoHandler struct {
	DaoSource dao.DaoSource
	Issuer    string
}

// UserinfoGet implements [api.UserInfoHandler].
func (obj *userInfoHandler) UserinfoGet(ctx context.Context) (*api.UserInfo, error) {
	jwt := dispatcherauth.GetAnyAuth(ctx)
	claims, err := jwtutil.ParseIdToken(ctx, jwt, obj.DaoSource.GetKeyStore(ctx), obj.Issuer)
	if err != nil {
		return nil, err
	}
	if claims == nil {
		return nil, fmt.Errorf("Not Logged In")
	}

	return &api.UserInfo{
		Sub: claims.Sub,
	}, nil
}
