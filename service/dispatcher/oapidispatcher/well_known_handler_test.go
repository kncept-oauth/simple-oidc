package oapidispatcher

import (
	"github.com/kncept-oauth/simple-oidc/service/gen/api"
)

var _ api.WellKnownHandler = (*wellKnownHandler)(nil)
