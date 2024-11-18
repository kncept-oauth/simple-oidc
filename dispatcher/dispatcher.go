package dispatcher

import (
	"context"
	"errors"
	"fmt"

	"github.com/kncept-oauth/simple-oidc/cmd/gen/api"
)

func NewApplication() (*api.Server, error) {
	fmt.Printf("NewApplication\n")
	return api.NewServer(&dispatcherHandler{})
}

type dispatcherHandler struct {
}

func (obj *dispatcherHandler) AuthorizeGet(ctx context.Context, params api.AuthorizeGetParams) (api.AuthorizeGetRes, error) {
	// lookup and validate tennant
	// also handle default tennant
	return nil, errors.ErrUnsupported
}

func (obj *dispatcherHandler) Index(ctx context.Context) (api.IndexRes, error) {
	return nil, errors.ErrUnsupported
}

func (obj *dispatcherHandler) LoginGet(ctx context.Context) error {
	return errors.ErrUnsupported
}
func (obj *dispatcherHandler) Jwks(ctx context.Context) (api.JwksRes, error) {
	return nil, errors.ErrUnsupported
}
func (obj *dispatcherHandler) OpenIdConfiguration(ctx context.Context) (api.OpenIdConfigurationRes, error) {
	return nil, errors.ErrUnsupported
}
