package dispatcher

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"text/template"

	"github.com/kncept-oauth/simple-oidc/service/jwtutil"
	"github.com/kncept-oauth/simple-oidc/service/keys"
	"github.com/kncept-oauth/simple-oidc/service/params"
	"github.com/kncept-oauth/simple-oidc/service/session"
	"github.com/kncept-oauth/simple-oidc/service/users"
	"github.com/kncept-oauth/simple-oidc/service/webcontent"
)

type acceptOidcHandler struct {
	daoSource DaoSource
	urlPrefix string

	templates sync.Map // map[string]*template.Template
}

func (obj *acceptOidcHandler) myAccountHandler() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		claims := obj.userClaims(req)
		if claims == nil {
			res.Header().Add("Location", "/accept")
			res.WriteHeader(302)
		}
		obj.respondWithTemplate("account.html", 200, res, params.QueryParamsToMap(req.URL))
	}
}

func (obj *acceptOidcHandler) snippetHandler() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		fmt.Printf("snippetHandler req.RequestURI: %v\n", req.RequestURI)
		snippet := strings.TrimSpace(strings.TrimPrefix(req.RequestURI, "/snippet/"))
		if snippet == "" {
			res.WriteHeader(404)
			return
		}
		obj.respondWithTemplate(fmt.Sprintf("/snippet/%v", snippet), 200, res, params.QueryParamsToMap(req.URL))
	}
}

func (obj *acceptOidcHandler) userId(req *http.Request) string {
	claims := obj.userClaims(req)
	if claims == nil {
		return ""
	}
	return claims.Sub
}

func (obj *acceptOidcHandler) userClaims(req *http.Request) *session.AuthTokenJwt {
	soJwt, err := req.Cookie(LoginJwtCookieName) // Simple Oidc Session JWT (if present)
	if err != nil {
		return nil
	}
	if soJwt == nil {
		return nil
	}
	kid := jwtutil.JwtKeyId(soJwt.Value)
	if kid == "" {
		return nil
	}
	key, err := obj.daoSource.GetKeyStore().GetKey(kid)
	if err != nil || key == nil {
		return nil
	}
	session := &session.AuthTokenJwt{}
	err = jwtutil.JwtToClaims(soJwt.Value, &key.Rsa.PublicKey, session)
	if err != nil {
		return nil
	}
	return session

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
				SameSite: http.SameSiteDefaultMode,
			}
			http.SetCookie(res, soCurrentCookie)
			res.Header().Add("Location", "/accept")
			res.WriteHeader(302)
			return
		}

		if !soCurrent.IsValid() {
			// TODO: Send to an 'invalid state' page (unrecoverable)
			res.WriteHeader(400)
			return
		}

		userId := obj.userId(req)
		if userId != "" {
			if req.Method == http.MethodGet {
				obj.respondWithTemplate("accept_authenticated.html", 200, res, params.QueryParamsToMap(req.URL))
				return
			}
			// handle POST action (logout of account, )
		} else {
			if req.Method == http.MethodGet {
				obj.respondWithTemplate("accept_unauthenticated.html", 200, res, params.QueryParamsToMap(req.URL))
				return
			}
		}

		res.WriteHeader(500)
	}
}

// func printDir(fs embed.FS, prefix string, dirs []fs.DirEntry, err error) {
// 	if err != nil {
// 		fmt.Printf("err: %v\n", err)
// 		return
// 	}
// 	for _, dir := range dirs {
// 		name := fmt.Sprintf("%v/%v", prefix, dir.Name())
// 		if prefix == "" {
// 			name = dir.Name()
// 		}
// 		if dir.IsDir() {
// 			fmt.Printf("%v/\n", name)
// 			dirs, err = fs.ReadDir(name)
// 			printDir(fs, name, dirs, err)
// 		} else {
// 			fmt.Printf("%v\n", name)
// 		}
// 	}
// }

func (obj *acceptOidcHandler) template(filename string) *template.Template {
	t, ok := obj.templates.Load("_")
	if ok {
		return t.(*template.Template)
	}

	// dirs, err := webcontent.Fs.ReadDir(".")
	// printDir(webcontent.Fs, "", dirs, err)
	// fmt.Printf("__\n\n")

	templ, err := template.New(filename).ParseFS(webcontent.Fs, "*.html", "snippet/*.html", "*.snippet")
	if err != nil {
		panic(err)
	}
	obj.templates.Store("_", templ)
	return templ
}

func (obj *acceptOidcHandler) respondWithTemplate(filename string, statusCode int, res http.ResponseWriter, data any) {
	t := obj.template(filename)
	res.Header().Add("Content-Type", "text/html")
	// TODO: configurable option - execute template before writing to stream
	res.WriteHeader(statusCode)
	err := t.ExecuteTemplate(res, filename, data)
	// err := t.Execute(res, data)
	if err != nil {
		fmt.Printf("%v", err)
	}
}

// func (obj *acceptOidcHandler) respondWithStaticFile(filename string, statusCode int, res http.ResponseWriter) {
// 	file, err := webcontent.Fs.Open(filename)
// 	if err == nil {
// 		fileContent, err := io.ReadAll(file)
// 		if err == nil {
// 			res.Header().Add("Content-Type", "text/html")
// 			res.WriteHeader(statusCode)
// 			res.Write(fileContent)
// 			return
// 		}
// 	}
// }

func (obj *acceptOidcHandler) createUserSession(res http.ResponseWriter, user *users.OidcUser) {
	key, err := keys.GetCurrentKey(obj.daoSource.GetKeyStore())
	if err != nil {
		obj.respondWithTemplate("register.html", 500, res, map[string]any{
			"err": err,
		})
		return
	}

	// create a simple-oidc session
	ses := session.NewSession(user.Id)
	err = obj.daoSource.GetSessionStore().SaveSession(ses)
	if err != nil {
		obj.respondWithTemplate("register.html", 500, res, map[string]any{
			"err": err,
		})
		return
	}

	authJwt := ses.MakeAuthTokenJwt(user, obj.urlPrefix, obj.urlPrefix)
	refreshJwt := ses.MakeRefreshTokenJwt(*authJwt)
	// TRACE
	jwt, err := jwtutil.ClaimsToJwt(authJwt, key.Kid, key.Rsa)
	if err != nil {
		obj.respondWithTemplate("register.html", 500, res, map[string]any{
			"err": err,
		})
		return
	}
	rt, err := jwtutil.ClaimsToJwt(refreshJwt, key.Kid, key.Rsa)
	if err != nil {
		obj.respondWithTemplate("register.html", 500, res, map[string]any{
			"err": err,
		})
		return
	}

	loginCookie := &http.Cookie{
		Name:  LoginJwtCookieName,
		Value: jwt,
		// Path:     "/",
		MaxAge:   9 * 60 * 60, // 9 hours
		HttpOnly: true,
		SameSite: http.SameSiteDefaultMode,
	}
	refreshCookie := &http.Cookie{
		Name:  LoginRefreshTokenCookieName,
		Value: rt,
		// Path:     "/",
		MaxAge:   7 * 24 * 60 * 60, // 7 days
		HttpOnly: true,
		SameSite: http.SameSiteDefaultMode,
	}

	http.SetCookie(res, loginCookie)
	http.SetCookie(res, refreshCookie)

	// redirect back to the accept page, now that they are logged in
	res.Header().Add("Location", "/accept")
	res.WriteHeader(302)
}

func (obj *acceptOidcHandler) registerHandler() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		claims := obj.userClaims(req)
		if claims != nil {
			res.Header().Add("Location", "/account")
			res.WriteHeader(302)
			return
		}

		if req.Method == http.MethodGet {
			obj.respondWithTemplate("register.html", 200, res, nil)
			return
		}
		if req.Method == http.MethodPost {
			// handle LOGIN (or REGISTRATION)
			//
			req.ParseForm()

			userService := &users.UserService{
				UserStore: obj.daoSource.GetUserStore(),
			}

			username := req.Form.Get("username")
			password := req.Form.Get("password")

			user, err := userService.AttemptUserRegistration(username, password)
			if errors.Is(err, users.ErrUserExists) {
				obj.respondWithTemplate("register.html", 400, res, map[string]any{
					"err": "user already exists",
				})
				return
			} else if err != nil {
				obj.respondWithTemplate("register.html", 500, res, map[string]any{
					"err": err,
				})
				return
			}
			if user == nil {
				obj.respondWithTemplate("register.html", 400, res, map[string]any{
					"err": "unable to create user",
				})
				return
			}
			obj.createUserSession(res, user)
			return
		}

		res.WriteHeader(500)
		return
	}
}

func (obj *acceptOidcHandler) loginHandler() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		claims := obj.userClaims(req)
		if claims != nil {
			res.Header().Add("Location", "/account")
			res.WriteHeader(302)
			return
		}

		// todo: load customisation by client id
		clientId := req.URL.Query().Get("client_id")
		client, err := obj.daoSource.GetClientStore().GetClient(clientId)
		if err == nil {
			fmt.Printf("client %v\n", client)
		} else {
			fmt.Printf("error fetching client %v\n", err)
		}

		if req.Method == http.MethodGet {
			obj.respondWithTemplate("login.html", 200, res, nil)
			return
		}
		if req.Method == http.MethodPost {
			req.ParseForm()

			userService := &users.UserService{
				UserStore: obj.daoSource.GetUserStore(),
			}

			username := req.Form.Get("username")
			password := req.Form.Get("password")

			user, err := userService.AttemptUserLogin(username, password)
			if err != nil || user == nil {
				res.WriteHeader(500)
			} else {
				obj.createUserSession(res, user)
				return
			}
		}

		res.WriteHeader(500)
	}
}
