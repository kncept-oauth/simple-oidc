package dispatcher

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/klauspost/compress/gzhttp"

	"github.com/kncept-oauth/simple-oidc/cmd/gen/api"
	"github.com/kncept-oauth/simple-oidc/webcontent"
)

func NewApplication() (http.HandlerFunc, error) {
	fmt.Printf("NewApplication\n")
	server, err := api.NewServer(&dispatcherHandler{})
	if err != nil {
		return nil, err
	}
	handler := gzhttp.GzipHandler(server)
	return handler, err
}

type dispatcherHandler struct {
}

func (obj *dispatcherHandler) AuthorizeGet(ctx context.Context, params api.AuthorizeGetParams) (api.AuthorizeGetRes, error) {
	// lookup and validate tennant
	// also handle default tennant
	return nil, errors.ErrUnsupported
}

func (obj *dispatcherHandler) Index(ctx context.Context) (api.IndexRes, error) {
	found, err := webcontent.Fs.ReadDir(".")
	if err != nil {
		return nil, err
	}
	for _, val := range found {
		fmt.Printf("found %v\n", val)
	}
	file, err := webcontent.Fs.Open("index.html")
	if err != nil {
		return nil, err
	}
	return &api.IndexOK{
		Data: file,
	}, nil
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
