package dispatcher

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/klauspost/compress/gzhttp"

	"github.com/kncept-oauth/simple-oidc/authorizer"
	"github.com/kncept-oauth/simple-oidc/gen/api"
	"github.com/kncept-oauth/simple-oidc/webcontent"
)

func NewApplication() (http.HandlerFunc, error) {
	fmt.Printf("NewApplication\n")

	serveMux := http.NewServeMux()

	serveMux.Handle("/snippet/", &snippetsHandler{})

	server, err := api.NewServer(&dispatcherHandler{})
	if err != nil {
		return nil, err
	}
	serveMux.Handle("/", server)

	handler := gzhttp.GzipHandler(serveMux)
	return handler, err
}

type snippetsHandler struct {
}

func (obj *snippetsHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	fmt.Printf("req.RequestURI: %v\n", req.RequestURI)
	snippet := strings.TrimSpace(strings.TrimPrefix(req.RequestURI, "/snippet/"))
	if snippet == "" {
		res.WriteHeader(404)
		return
	}
	fmt.Printf("snippet %v\n", snippet)
	res.WriteHeader(400)
}

type dispatcherHandler struct {
	a authorizer.Authorizer
}

func (obj *dispatcherHandler) AuthorizeGet(ctx context.Context, params api.AuthorizeGetParams) (api.AuthorizeGetRes, error) {
	return obj.a.AuthorizeGet(ctx, params)
}

func (obj *dispatcherHandler) Index(ctx context.Context) (api.IndexOK, error) {
	file, err := webcontent.Fs.Open("index.html")
	if err != nil {
		return api.IndexOK{}, err
	}
	return api.IndexOK{
		Data: file,
	}, nil
}

func (obj *dispatcherHandler) LoginGet(ctx context.Context) error {
	return errors.ErrUnsupported
}
func (obj *dispatcherHandler) Jwks(ctx context.Context) (*api.JWKSetResponse, error) {
	return nil, errors.ErrUnsupported
}
func (obj *dispatcherHandler) OpenIdConfiguration(ctx context.Context) (*api.OpenIDProviderMetadataResponse, error) {
	return nil, errors.ErrUnsupported
}

func (obj *dispatcherHandler) NewError(ctx context.Context, err error) *api.ErrRespStatusCode {
	fmt.Printf("General error occurred: %v\n", err)
	return &api.ErrRespStatusCode{
		StatusCode: 500,
		Response:   fmt.Sprintf("%v", err),
	}
}
