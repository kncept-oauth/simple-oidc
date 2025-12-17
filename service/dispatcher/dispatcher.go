package dispatcher

import (
	"net/http"
	"strings"

	"github.com/klauspost/compress/gzhttp"

	"github.com/kncept-oauth/simple-oidc/service/dao"
	"github.com/kncept-oauth/simple-oidc/service/dispatcherauth"
	"github.com/kncept-oauth/simple-oidc/service/gen/api"
)

const CurrentOperationParamsCookieName = "so-cp"
const CurrentOperationNameCookieName = "so-op"

const LoginJwtCookieName = "so-jwt" // contains the jwt
const LoginRefreshTokenCookieName = "so-ts"

func NewApplication(daoSource dao.DaoSource, urlPrefix string) (http.HandlerFunc, error) {
	urlPrefix = strings.TrimSuffix(urlPrefix, "/")
	serveMux := http.NewServeMux()

	acceptOidcHandler := acceptOidcHandler{
		urlPrefix: urlPrefix,
		daoSource: daoSource,
	}

	serveMux.Handle("/snippet/", acceptOidcHandler.snippetHandler())
	serveMux.Handle("/accept", acceptOidcHandler.acceptLogin())
	serveMux.Handle("/login", acceptOidcHandler.loginHandler())
	serveMux.Handle("/logout", acceptOidcHandler.logoutHandler())
	serveMux.Handle("/register", acceptOidcHandler.registerHandler())
	serveMux.Handle("/me", acceptOidcHandler.myAccountHandler()) // TODO: Redirect to /account (or /login)
	serveMux.Handle("/account", acceptOidcHandler.myAccountHandler())
	// serveMux.Handle("/style.css", acceptOidcHandler.respondWithStaticFile("style.css", "text/css", 200))
	// serveMux.Handle("/htmx.js", acceptOidcHandler.respondWithStaticFile("htmx.js", "application/javascript", 200))
	// serveMux.Handle("/header.js", acceptOidcHandler.respondWithStaticFile("header.js", "application/javascript", 200))
	serveMux.Handle("/confirm", acceptOidcHandler.confirmLogin())

	serveMux.Handle("/deauthorize/", acceptOidcHandler.deauthClientHandler())

	server, err := api.NewServer(
		&oapiDispatcher{
			authorizationHandler: authorizationHandler{
				DaoSource: daoSource,
				Issuer:    urlPrefix,
			},
			wellKnownHandler: wellKnownHandler{
				DaoSource: daoSource,
				Issuer:    urlPrefix,
			},
		},
		&dispatcherauth.Handler{},
		api.WithNotFound(func(w http.ResponseWriter, r *http.Request) {

			if strings.HasSuffix(r.URL.Path, ".js") {
				acceptOidcHandler.respondWithStaticFile(r.URL.Path, "application/javascript", 200)(w, r)
				return
			}
			if strings.HasSuffix(r.URL.Path, ".css") {
				acceptOidcHandler.respondWithStaticFile(r.URL.Path, "test/css", 200)(w, r)
				return
			}

			if r.URL.Path == "/" {
				// TODO: detect login, and set flag (eg: display login or account details link)
				acceptOidcHandler.respondWithTemplate("index.html", 200, w, nil)
			} else {
				w.WriteHeader(404)
				// acceptOidcHandler.respondWithTemplate("notfound.html", 404, w, nil) // todo
			}
		}),
	)
	if err != nil {
		return nil, err
	}
	serveMux.Handle("/", server)
	handler := gzhttp.GzipHandler(serveMux)
	return handler, err
}
