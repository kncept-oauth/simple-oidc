package dispatcher

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/klauspost/compress/gzhttp"

	"github.com/kncept-oauth/simple-oidc/service/authorizer"
	"github.com/kncept-oauth/simple-oidc/service/dispatcherauth"
	"github.com/kncept-oauth/simple-oidc/service/gen/api"
	"github.com/kncept-oauth/simple-oidc/service/webcontent"
)

func NewApplication(daoSource DaoSource) (http.HandlerFunc, error) {
	fmt.Printf("NewApplication\n")

	serveMux := http.NewServeMux()

	serveMux.Handle("/snippet/", &snippetsHandler{})
	server, err := api.NewServer(
		&dispatcherHandler{
			authorizer: authorizer.NewAuthorizer(
				daoSource.GetClientStore(),
			),
		},
		&dispatcherauth.Handler{},
	)
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
	authorizer authorizer.Authorizer
}

func (obj *dispatcherHandler) AuthorizeGet(ctx context.Context, params api.AuthorizeGetParams) (api.AuthorizeGetRes, error) {
	return obj.authorizer.AuthorizeGet(ctx, params)
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
	fmt.Printf("TODO: LoginGet\n")
	return errors.ErrUnsupported
}
func (obj *dispatcherHandler) Jwks(ctx context.Context) (*api.JWKSetResponse, error) {
	fmt.Printf("TODO: Jwks\n")
	return nil, errors.ErrUnsupported
}
func (obj *dispatcherHandler) OpenIdConfiguration(ctx context.Context) (*api.OpenIDProviderMetadataResponse, error) {
	return &api.OpenIDProviderMetadataResponse{
		Issuer:                "http://localhost:8080", // todo :/
		AuthorizationEndpoint: "/authorize",
		TokenEndpoint:         "todo",
		JwksURI:               "/.well-known/jwks.json",
	}, nil

	// fmt.Printf("TODO: OpenIdConfiguration\n")
	// return nil, errors.ErrUnsupported
}

func (obj *dispatcherHandler) Me(ctx context.Context) (api.MeOK, error) {
	return api.MeOK{}, errors.ErrUnsupported
}

func (obj *dispatcherHandler) NewError(ctx context.Context, err error) *api.ErrRespStatusCode {
	fmt.Printf("General error occurred: %v\n", err)
	return &api.ErrRespStatusCode{
		StatusCode: 500,
		Response:   fmt.Sprintf("%v", err),
	}
}
