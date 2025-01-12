package dispatcherauth

import (
	"testing"

	"github.com/kncept-oauth/simple-oidc/service/gen/api"
)

func TestAuthHandlerType(t *testing.T) {
	assertIsApiSecurityHandler(&Handler{})
}

func assertIsApiSecurityHandler(receiver api.SecurityHandler) {}
