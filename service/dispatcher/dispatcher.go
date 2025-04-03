package dispatcher

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"net/http"
	"strings"
	"sync"

	"github.com/klauspost/compress/gzhttp"

	"github.com/kncept-oauth/simple-oidc/service/authorizer"
	"github.com/kncept-oauth/simple-oidc/service/dispatcherauth"
	"github.com/kncept-oauth/simple-oidc/service/gen/api"
	"github.com/kncept-oauth/simple-oidc/service/params"
	"github.com/kncept-oauth/simple-oidc/service/webcontent"

	"github.com/golang-jwt/jwt/v5"
)

const CurrentOperationParamsCookieName = "so-current"
const CurrentOperationNameCookieName = "so-op"
const LoginJwtCookieName = "so-jwt"

func NewApplication(daoSource DaoSource, urlPrefix string) (http.HandlerFunc, error) {
	fmt.Printf("NewApplication\n")

	serveMux := http.NewServeMux()

	acceptOidcHandler := acceptOidcHandler{
		daoSource: daoSource,
	}

	serveMux.Handle("/snippet/", acceptOidcHandler.snippetHandler())
	serveMux.Handle("/accept", acceptOidcHandler.acceptLogin())
	serveMux.Handle("/login", acceptOidcHandler.loginHandler())
	serveMux.Handle("/register", acceptOidcHandler.registerHandler())

	server, err := api.NewServer(
		&oapiDispatcher{
			authorizer: authorizer.NewAuthorizer(
				daoSource.GetClientStore(),
			),
			Issuer: strings.TrimSuffix(urlPrefix, "/"),
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

type acceptOidcHandler struct {
	daoSource DaoSource
	templates sync.Map // map[string]*template.Template
}

func (obj *acceptOidcHandler) snippetHandler() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		fmt.Printf("snippetHandler req.RequestURI: %v\n", req.RequestURI)
		snippet := strings.TrimSpace(strings.TrimPrefix(req.RequestURI, "/snippet/"))
		if snippet == "" {
			res.WriteHeader(404)
			return
		}
		obj.respondWithFile(fmt.Sprintf("/snippet/%v", snippet), 200, res, params.QueryParamsToMap(req.URL))
	}
}

func (obj *acceptOidcHandler) userId(req *http.Request) string {
	claims := obj.userClaims(req)
	if claims == nil {
		return ""
	}
	sub, err := claims.GetSubject()
	if err != nil {
		return sub
	}
	return ""
}

func (obj *acceptOidcHandler) userClaims(req *http.Request) jwt.Claims {
	soJwt, err := req.Cookie(LoginJwtCookieName) // Simple Oidc Session JWT (if present)
	if err != nil {
		return nil
	}
	if soJwt == nil {
		return nil
	}
	token, err := jwt.Parse(strings.TrimSpace(soJwt.Value), func(token *jwt.Token) (interface{}, error) {
		kid, ok := token.Header["kid"]
		if !ok || kid == "" {
			return nil, nil
		}

		// audiences, err := token.Claims.GetAudience()
		// if err != nil {
		// 	return nil, err
		// }
		// if len(audiences) != 1 || audiences[0] != "simple-oidc" { // this needs to be an ENV VAR
		// 	return nil, nil
		// }

		// fetch key by id
		keyStore := obj.daoSource.GetKeyStore()
		keypair, err := keyStore.GetKey(kid.(string))
		if err != nil {
			return nil, err
		}
		return keypair.Pvt, nil
	})
	if err == nil && token != nil {
		return token.Claims
	}
	return nil
}

func (obj *acceptOidcHandler) acceptLogin() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {

		// current auth attempt details
		// if _anything_ in the query params is different, update
		// if _anything_ is in the query params, save and redirect back with a clean url path
		soCurrentParams := ""
		soCurrentCookie, _ := req.Cookie(CurrentOperationParamsCookieName)
		if soCurrentCookie != nil {
			soCurrentParams = soCurrentCookie.Value
		}
		soCurrent, err := params.OidcParamsFromQuery(soCurrentParams)
		if err != nil {
			fmt.Printf("%v\n", err)
			res.WriteHeader(500)
			return
		}

		stateMap := params.QueryParamsToMap(req.URL)
		if len(stateMap) != 0 {
			updatedOidcParams := params.OidcParamsFromMap(stateMap)
			soCurrent.Merge(updatedOidcParams)

			soCurrentCookie = &http.Cookie{
				Name:  CurrentOperationParamsCookieName,
				Value: soCurrent.ToQueryParams(),
				// Path:     "/",
				MaxAge:   15 * 60, // 15 min
				HttpOnly: true,
				SameSite: http.SameSiteStrictMode,
			}
			err = soCurrentCookie.Valid()
			if err != nil {
				fmt.Printf("invalid cookie: %v\n", err)
			}
			http.SetCookie(res, soCurrentCookie)
			res.Header().Add("Location", "/accept")
			res.WriteHeader(302)
			return
		}

		if !soCurrent.IsValid() {
			res.WriteHeader(400)
			return
		}

		// todo: load customisation by client id
		// client, err := obj.daoSource.GetClientStore().Get(oidcParams.ClientId)
		// if err == nil {
		// 	fmt.Printf("client %v\n", client)
		// } else {
		// 	fmt.Printf("error fetching client %v\n", err)
		// }

		userId := obj.userId(req)
		if userId != "" {
			if req.Method == http.MethodGet {
				obj.respondWithFile("accept_authenticated.html", 200, res, params.QueryParamsToMap(req.URL))
				return
			}
			// handle POST action (logout of account, )
		} else {
			if req.Method == http.MethodGet {
				obj.respondWithFile("accept_unauthenticated.html", 200, res, params.QueryParamsToMap(req.URL))
				return
			}
		}

		res.WriteHeader(500)
	}
}

func printDir(fs embed.FS, prefix string, dirs []fs.DirEntry, err error) {
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}
	for _, dir := range dirs {
		name := fmt.Sprintf("%v/%v", prefix, dir.Name())
		if prefix == "" {
			name = dir.Name()
		}
		if dir.IsDir() {
			fmt.Printf("%v/\n", name)
			dirs, err = fs.ReadDir(name)
			printDir(fs, name, dirs, err)
		} else {
			fmt.Printf("%v\n", name)
		}
	}
}

func (obj *acceptOidcHandler) template(filename string) *template.Template {
	t, ok := obj.templates.Load("_")
	if ok {
		return t.(*template.Template)
	}

	dirs, err := webcontent.Fs.ReadDir(".")
	printDir(webcontent.Fs, "", dirs, err)
	fmt.Printf("__\n\n")

	templ, err := template.New(filename).ParseFS(webcontent.Fs, "*.html", "snippet/*.html", "*.snippet")
	if err != nil {
		panic(err)
	}
	obj.templates.Store("_", templ)
	return templ
}

func (obj *acceptOidcHandler) respondWithFile(filename string, statusCode int, res http.ResponseWriter, data any) {
	t := obj.template(filename)
	res.Header().Add("Content-Type", "text/html")
	// TODO: configurable option - execute template before writing to stream
	res.WriteHeader(statusCode)
	err := t.ExecuteTemplate(res, filename, data)
	// err := t.Execute(res, data)
	if err != nil {
		fmt.Printf("%v", err)
	}
	return

}

func (obj *acceptOidcHandler) registerHandler() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		// todo: load customisation by client id
		clientId := req.URL.Query().Get("client_id")
		client, err := obj.daoSource.GetClientStore().Get(clientId)
		if err == nil {
			fmt.Printf("client %v\n", client)
		} else {
			fmt.Printf("error fetching client %v\n", err)
		}

		if req.Method == http.MethodGet {
			file, err := webcontent.Fs.Open("register.html")
			if err == nil {
				fileContent, err := io.ReadAll(file)
				if err == nil {
					res.Header().Add("Content-Type", "text/html")
					res.WriteHeader(200)
					res.Write(fileContent)
					return
				}
			}
		}
		if req.Method == http.MethodPost {
			// handle LOGIN (or REGISTRATION)
			//
		}

		res.WriteHeader(500)
	}
}
func (obj *acceptOidcHandler) loginHandler() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		// todo: load customisation by client id
		clientId := req.URL.Query().Get("client_id")
		client, err := obj.daoSource.GetClientStore().Get(clientId)
		if err == nil {
			fmt.Printf("client %v\n", client)
		} else {
			fmt.Printf("error fetching client %v\n", err)
		}

		if req.Method == http.MethodGet {
			file, err := webcontent.Fs.Open("login.html")
			if err == nil {
				fileContent, err := io.ReadAll(file)
				if err == nil {
					res.Header().Add("Content-Type", "text/html")
					res.WriteHeader(200)
					res.Write(fileContent)
					return
				}
			}
		}
		if req.Method == http.MethodPost {
			// handle LOGIN (or REGISTRATION)
			//
		}

		res.WriteHeader(500)
	}
}

type oapiDispatcher struct {
	authorizer authorizer.Authorizer
	Issuer     string
}

func (obj *oapiDispatcher) AuthorizeGet(ctx context.Context, params api.AuthorizeGetParams) (api.AuthorizeGetRes, error) {
	return obj.authorizer.AuthorizeGet(ctx, params)
}

func (obj *oapiDispatcher) Index(ctx context.Context) (api.IndexOK, error) {
	file, err := webcontent.Fs.Open("index.html")
	if err != nil {
		return api.IndexOK{}, err
	}
	return api.IndexOK{
		Data: file,
	}, nil
}

func (obj *oapiDispatcher) LoginGet(ctx context.Context) error {
	fmt.Printf("TODO: LoginGet\n")
	return errors.ErrUnsupported
}
func (obj *oapiDispatcher) Jwks(ctx context.Context) (*api.JWKSetResponse, error) {
	fmt.Printf("TODO: Jwks\n")
	return nil, errors.ErrUnsupported
}
func (obj *oapiDispatcher) OpenIdConfiguration(ctx context.Context) (*api.OpenIDProviderMetadataResponse, error) {
	return &api.OpenIDProviderMetadataResponse{
		// Issuer:                "https://localhost:8443", // todo :/
		Issuer:                obj.Issuer,
		AuthorizationEndpoint: "/authorize",
		TokenEndpoint:         "todo",
		JwksURI:               "/.well-known/jwks.json",
	}, nil

	// fmt.Printf("TODO: OpenIdConfiguration\n")
	// return nil, errors.ErrUnsupported
}

func (obj *oapiDispatcher) Me(ctx context.Context) (api.MeOK, error) {
	return api.MeOK{}, errors.ErrUnsupported
}

func (obj *oapiDispatcher) NewError(ctx context.Context, err error) *api.ErrRespStatusCode {
	fmt.Printf("General error occurred: %v\n", err)
	return &api.ErrRespStatusCode{
		StatusCode: 500,
		Response:   fmt.Sprintf("%v", err),
	}
}
