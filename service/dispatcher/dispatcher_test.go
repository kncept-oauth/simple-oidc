package dispatcher

import (
	"testing"

	"github.com/kncept-oauth/simple-oidc/service/gen/api"
)

func TestDispatcherHandlerType(t *testing.T) {
	assertIsApiHandler(&dispatcherHandler{})
}
func assertIsApiHandler(receiver api.Handler) {}
