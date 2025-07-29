package dispatcher

import (
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/kncept-oauth/simple-oidc/service/client"
	"github.com/kncept-oauth/simple-oidc/service/dao"
	"github.com/kncept-oauth/simple-oidc/service/jwtutil"
	"github.com/kncept-oauth/simple-oidc/service/keys"
	"github.com/kncept-oauth/simple-oidc/service/params"
	"github.com/kncept-oauth/simple-oidc/service/session"
	"github.com/kncept-oauth/simple-oidc/service/users"
	"github.com/kncept-oauth/simple-oidc/service/webcontent"
)

type acceptOidcHandler struct {
	daoSource dao.DaoSource
	urlPrefix string
	mu        sync.Mutex
	tmpl      *template.Template
}

func (obj *acceptOidcHandler) myAccountHandler() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		claims := obj.userClaims(req)
		if claims == nil {
			res.Header().Add("Location", "/login")
			res.WriteHeader(302)
			return
		}

		userId := claims.Sub

		user, err := obj.daoSource.GetUserStore().GetUser(userId)
		if err != nil {
			fmt.Printf("%v\n", err)
			res.WriteHeader(500)
		}

		clientAuthorizations := &client.DepaginatedScroller{}
		err = obj.daoSource.GetClientAuthorizationStore().ClientAuthorizationsByUser(userId, clientAuthorizations.Scroller)
		if err != nil {
			fmt.Printf("%v\n", err)
			res.WriteHeader(500)
			return
		}

		type account_page_params struct {
			User                 *users.OidcUser
			ClientAuthorizations []*client.ClientAuthorization
		}

		params := account_page_params{
			User:                 user,
			ClientAuthorizations: clientAuthorizations.Results,
		}

		obj.respondWithTemplate("account.html", 200, res, params)
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
		obj.respondWithTemplate(fmt.Sprintf("%v.snippet", snippet), 200, res, params.QueryParamsToMap(req.URL))
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

// show the 'accept page' so that the user KNOWS where they are going
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

		type accept_page_params struct {
			Params                params.OidcAuthCodeFlowParams
			ExistingAuthorization *client.ClientAuthorization
			ClientAuthorizations  []*client.ClientAuthorization
		}

		acceptPageParams := accept_page_params{
			Params: *soCurrent,
		}

		userId := obj.userId(req)
		if userId != "" {
			if req.Method == http.MethodGet {

				clientAuthorizationsScroller := &client.DepaginatedScroller{}
				err := obj.daoSource.GetClientAuthorizationStore().ClientAuthorizationsByUser(userId, clientAuthorizationsScroller.Scroller)
				if err != nil {
					fmt.Printf("%v\n", err)
					res.WriteHeader(500)
					return
				}
				clientAuthorizations := make([]*client.ClientAuthorization, 0, len(clientAuthorizationsScroller.Results))
				acceptPageParams.ClientAuthorizations = clientAuthorizationsScroller.Results
				for _, existingAuth := range clientAuthorizationsScroller.Results {
					if existingAuth.ClientId == soCurrent.ClientId {
						acceptPageParams.ExistingAuthorization = existingAuth
					} else {
						clientAuthorizations = append(clientAuthorizations, existingAuth)
					}
				}
				acceptPageParams.ClientAuthorizations = clientAuthorizations
				obj.respondWithTemplate("accept_authenticated.html", 200, res, acceptPageParams)
				return
			}
			// handle POST action (logout of account, etc)
		} else {
			if req.Method == http.MethodGet {
				obj.respondWithTemplate("accept_unauthenticated.html", 200, res, acceptPageParams)
				return
			}
		}

		res.WriteHeader(500)
	}
}

// click 'confirm' ==> redirect back to app
func (obj *acceptOidcHandler) confirmLogin() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
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
		if userId == "" {
			res.WriteHeader(400)
			// res.Header().Add("Location", "/authorize")
			// res.WriteHeader(302)
			return
		}

		clientAuthorizations := &client.DepaginatedScroller{}
		err = obj.daoSource.GetClientAuthorizationStore().ClientAuthorizationsByUser(userId, clientAuthorizations.Scroller)
		if err != nil {
			fmt.Printf("%v\n", err)
			res.WriteHeader(500)
			return
		}
		newAuthorization := true
		for _, existingAuth := range clientAuthorizations.Results {
			if existingAuth.ClientId == soCurrent.ClientId {
				newAuthorization = false
			}
		}
		now := time.Now().UTC()
		if newAuthorization {
			err = obj.daoSource.GetClientAuthorizationStore().SaveClientAuthorization(&client.ClientAuthorization{
				ClientId:               soCurrent.ClientId,
				UserId:                 userId,
				AuthorizedAt:           now,
				LastRefreshedAt:        now,
				AuthorizationSessionId: uuid.NewString(),
			})
			if err != nil {
				fmt.Printf("%v\n", err)
				res.WriteHeader(500)
				return
			}
		}

		authCodeStore := obj.daoSource.GetAuthorizationCodeStore()
		authCode, err := client.NewAuthorizationCode(userId, soCurrent.ToQueryParams())
		if err != nil {
			fmt.Printf("%v\n", err)
			res.WriteHeader(500)
			return
		}
		err = authCodeStore.SaveAuthorizationCode(ctx, authCode)
		if err != nil {
			fmt.Printf("%v\n", err)
			res.WriteHeader(500)
			return
		}
		code := authCode.Code
		state := soCurrent.State

		u, err := url.Parse(soCurrent.RedirectUri)
		if err != nil {
			fmt.Printf("%v\n", err)
			res.WriteHeader(500)
			return
		}
		q := u.Query()
		q.Add("code", code)
		if state != "" {
			q.Add("state", state)
		}
		u.RawQuery = q.Encode()

		res.Header().Add("Location", u.String())
		res.WriteHeader(302)
	}
}

func (obj *acceptOidcHandler) templates() *template.Template {
	obj.mu.Lock()
	defer obj.mu.Unlock()
	if obj.tmpl != nil {
		return obj.tmpl
	}

	templ := template.New("_")

	f := map[string]any{
		"Wrap": func(keyPairs ...any) any {
			keyPairsLen := len(keyPairs)
			if keyPairsLen%2 != 0 {
				panic("Must supply a full set of key pairs")
			}
			m := map[any]any{}
			i := 0
			for i < keyPairsLen {
				m[keyPairs[i]] = keyPairs[i+1]
				i = i + 2
			}
			return m
		},
		"Coalesce": func(str ...string) string {
			for _, s := range str {
				if s != "" {
					return s
				}
			}
			return ""
		},

		"time": func(args ...any) string {
			if len(args) == 0 {
				return ""
			}
			if t, ok := args[0].(time.Time); ok {
				if len(args) == 1 {
					return t.String()
				}
				if args[1] == "since" {
					duration := time.Since(t)

					seconds := int(duration.Seconds())
					if seconds < 60 {
						return fmt.Sprintf("%d seconds ago", seconds)
					}

					minutes := int(duration.Minutes())
					if minutes < 60 {
						return fmt.Sprintf("%d minutes ago", minutes)
					}

					hours := int(duration.Hours())
					if hours < 24 {
						return fmt.Sprintf("%d hours ago", hours)
					}

					days := int(hours / 24)
					if days < 30 {
						return fmt.Sprintf("%d days ago", days)
					}

					return fmt.Sprintf("%d months ago", int(days/30))

				}
				panic("unknown use of time function")
			} else {
				return ""
			}

		},
	}
	templ = templ.Funcs(f)
	templ, err := templ.ParseFS(webcontent.Fs, "*.html", "snippet/*.snippet")
	if err != nil {
		panic(err)
	}

	obj.tmpl = templ
	return templ
}

func (obj *acceptOidcHandler) respondWithTemplate(filename string, statusCode int, res http.ResponseWriter, data any) {
	t := obj.templates()
	res.Header().Add("Content-Type", "text/html")
	// TODO: configurable option - execute template before writing to stream
	res.WriteHeader(statusCode)
	err := t.ExecuteTemplate(res, filename, data)
	// err := t.Execute(res, data)
	if err != nil {
		fmt.Printf("%v", err)
	}
}

func (obj *acceptOidcHandler) respondWithStaticFile(filename string, contentType string, statusCode int) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		if !strings.HasPrefix(filename, "/") || strings.Contains(filename, "..") {
			res.WriteHeader(http.StatusNotFound)
			return
		}
		file, err := webcontent.Fs.Open(fmt.Sprintf("static%v", filename))
		if err == nil {
			fileContent, err := io.ReadAll(file)
			if err == nil {
				if contentType != "" {
					res.Header().Add("Content-Type", contentType)
				}
				res.WriteHeader(statusCode)
				res.Write(fileContent)
				return
			}
		}
		if errors.Is(err, os.ErrNotExist) {
			res.WriteHeader(http.StatusNotFound)
			return
		}
		res.WriteHeader(http.StatusInternalServerError)
	}
}

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
		// clientId := req.URL.Query().Get("client_id")
		// client, err := obj.daoSource.GetClientStore().GetClient(clientId)
		// if err == nil {
		// 	fmt.Printf("client %v\n", client)
		// } else {
		// 	fmt.Printf("error fetching client %v\n", err)
		// }
		// if client == nil || err != nil {
		// 	res.WriteHeader(500)
		// 	return
		// }

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
				return
			} else {
				obj.createUserSession(res, user)
				return
			}
		}

		res.WriteHeader(500)
	}
}

func (obj *acceptOidcHandler) logoutHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// expire login cookie
		loginCookie := &http.Cookie{
			Name:     LoginJwtCookieName,
			Value:    "",
			MaxAge:   0, // expire cookie
			HttpOnly: true,
			SameSite: http.SameSiteDefaultMode,
		}
		http.SetCookie(w, loginCookie)
		http.Redirect(w, r, "/", http.StatusFound)
	}
}
