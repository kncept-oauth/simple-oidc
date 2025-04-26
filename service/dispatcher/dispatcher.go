package dispatcher

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/klauspost/compress/gzhttp"

	"github.com/kncept-oauth/simple-oidc/service/authorizer"
	"github.com/kncept-oauth/simple-oidc/service/dispatcherauth"
	"github.com/kncept-oauth/simple-oidc/service/gen/api"
)

const CurrentOperationParamsCookieName = "so-current"
const CurrentOperationNameCookieName = "so-op"

const LoginJwtCookieName = "so-jwt"
const LoginRefreshTokenCookieName = "so-ts"

func NewApplication(daoSource DaoSource, urlPrefix string) (http.HandlerFunc, error) {
	fmt.Printf("NewApplication: %v\n", urlPrefix)
	serveMux := http.NewServeMux()

	acceptOidcHandler := acceptOidcHandler{
		urlPrefix: urlPrefix,
		daoSource: daoSource,
	}

	serveMux.Handle("/snippet/", acceptOidcHandler.snippetHandler())
	serveMux.Handle("/accept", acceptOidcHandler.acceptLogin())
	serveMux.Handle("/login", acceptOidcHandler.loginHandler())
	serveMux.Handle("/register", acceptOidcHandler.registerHandler())
	serveMux.Handle("/me", acceptOidcHandler.myAccountHandler()) // TODO: Redirect to /account (or /login)
	serveMux.Handle("/account", acceptOidcHandler.myAccountHandler())

	server, err := api.NewServer(
		&oapiDispatcher{
			authorizer: authorizer.NewAuthorizer(
				daoSource.GetClientStore(),
			),
			Issuer:    strings.TrimSuffix(urlPrefix, "/"),
			daoSource: daoSource,
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
