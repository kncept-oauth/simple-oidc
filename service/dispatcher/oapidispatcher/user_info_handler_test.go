package oapidispatcher

import "github.com/kncept-oauth/simple-oidc/service/gen/api"

var _ api.UserInfoHandler = (*userInfoHandler)(nil)
