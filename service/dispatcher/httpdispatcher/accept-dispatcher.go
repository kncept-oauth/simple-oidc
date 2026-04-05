package httpdispatcher

import (
	"context"
	"crypto/rsa"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/kncept-oauth/simple-oidc/service/client"
	"github.com/kncept-oauth/simple-oidc/service/dao"
	"github.com/kncept-oauth/simple-oidc/service/dao/ddbutil"
	"github.com/kncept-oauth/simple-oidc/service/jwtutil"
	"github.com/kncept-oauth/simple-oidc/service/keys"
	"github.com/kncept-oauth/simple-oidc/service/params"
	"github.com/kncept-oauth/simple-oidc/service/session"
	"github.com/kncept-oauth/simple-oidc/service/users"
)

const CurrentOperationParamsCookieName = "so-cp"
const CurrentOperationNameCookieName = "so-op"

const LoginJwtCookieName = "so-jwt" // contains the jwt
const LoginRefreshTokenCookieName = "so-ts"

type acceptOidcHandler struct {
	daoSource dao.DaoSource
	urlPrefix string

	templateDispatcher *TemplateDispatcher
}

func NewAcceptOidcHandler(
	daoSource dao.DaoSource,
	urlPrefix string,
	devModeLiveFilesystemBase *string,
) *http.ServeMux {
	serveMux := http.NewServeMux()

	acceptOidcHandler := acceptOidcHandler{
		urlPrefix:          urlPrefix,
		daoSource:          daoSource,
		templateDispatcher: NewTemplateDispatcher(devModeLiveFilesystemBase),
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

	return serveMux
}

func (obj *acceptOidcHandler) myAccountHandler() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		claims := obj.userClaims(req)
		if claims == nil {
			res.Header().Add("Location", "/login")
			res.WriteHeader(302)
			return
		}

		userId := claims.Sub

		user, err := obj.daoSource.GetUserStore(ctx).GetUser(ctx, userId)
		if err != nil {
			fmt.Printf("%v\n", err)
			res.WriteHeader(500)
		}

		clientAuthorizations := &ddbutil.DepaginatedScroller[client.ClientAuthorization]{}
		err = obj.daoSource.GetClientAuthorizationStore(ctx).ClientAuthorizationsByUser(ctx, userId, clientAuthorizations)
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

		obj.templateDispatcher.RespondWithTemplate("account.html", 200, res, params)
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
		obj.templateDispatcher.RespondWithTemplate(fmt.Sprintf("%v.snippet", snippet), 200, res, params.QueryParamsToMap(req.URL))
	}
}

func (obj *acceptOidcHandler) userId(req *http.Request) string {
	claims := obj.userClaims(req)
	if claims == nil {
		return ""
	}
	return claims.Sub
}

func (obj *acceptOidcHandler) userClaims(req *http.Request) *jwtutil.IdToken {
	ctx := req.Context()
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
	key, err := obj.daoSource.GetKeyStore(ctx).GetKey(ctx, kid)
	if err != nil || key == nil {
		return nil
	}

	decodedKey, err := key.DecodePrivateKey()
	if err != nil {
		return nil
	}
	rsaKey, isRsaKey := decodedKey.(*rsa.PrivateKey)
	if !isRsaKey {
		return nil
	}

	jwtToken := &jwtutil.IdToken{}
	err = jwtutil.JwtToClaims(soJwt.Value, &rsaKey.PublicKey, jwtToken)
	if err != nil {
		return nil
	}

	// // jwtToken.SessionId
	// session, err := obj.daoSource.GetSessionStore(ctx).LoadSession(ctx, jwtToken.SessionId)
	// if err == nil {
	// 	return nil
	// }
	// // session has expired, or been revoked
	// if session == nil {
	// 	return nil
	// }

	return jwtToken
}

// show the 'accept page' so that the user KNOWS where they are going
func (obj *acceptOidcHandler) acceptLogin() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		ctx := req.Context()

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
				clientAuthorizationsScroller := &ddbutil.DepaginatedScroller[client.ClientAuthorization]{}
				err := obj.daoSource.GetClientAuthorizationStore(ctx).ClientAuthorizationsByUser(ctx, userId, clientAuthorizationsScroller)
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
				obj.templateDispatcher.RespondWithTemplate("accept_authenticated.html", 200, res, acceptPageParams)
				return
			}
			// handle POST action (logout of account, etc)
		} else {
			if req.Method == http.MethodGet {
				obj.templateDispatcher.RespondWithTemplate("accept_unauthenticated.html", 200, res, acceptPageParams)
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

		clientAuthorizations := &ddbutil.DepaginatedScroller[client.ClientAuthorization]{}
		err = obj.daoSource.GetClientAuthorizationStore(ctx).ClientAuthorizationsByUser(ctx, userId, clientAuthorizations)
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
			err = obj.daoSource.GetClientAuthorizationStore(ctx).SaveClientAuthorization(ctx, &client.ClientAuthorization{
				ClientId:     soCurrent.ClientId,
				UserId:       userId,
				AuthorizedAt: now,
			})
			if err != nil {
				fmt.Printf("%v\n", err)
				res.WriteHeader(500)
				return
			}
		}

		authCodeStore := obj.daoSource.GetAuthorizationCodeStore(ctx)
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

func (obj *acceptOidcHandler) createUserSession(ctx context.Context, res http.ResponseWriter, user *users.OidcUser) {
	key, err := keys.GetCurrentKey(ctx, obj.daoSource.GetKeyStore(ctx))

	decodedKey, err := key.DecodePrivateKey()
	if err != nil {
		obj.templateDispatcher.RespondWithTemplate("register.html", 500, res, map[string]any{
			"err": err,
		})
		return
	}
	rsaKey, isRsaKey := decodedKey.(*rsa.PrivateKey)
	if !isRsaKey {
		obj.templateDispatcher.RespondWithTemplate("register.html", 500, res, map[string]any{
			"err": errors.New("not an rsa key"),
		})
		return
	}

	if err != nil {
		obj.templateDispatcher.RespondWithTemplate("register.html", 500, res, map[string]any{
			"err": err,
		})
		return
	}

	// create a simple-oidc session
	ses, err := session.NewSession(user.Id, client.ClientId_SimpleOidc)
	if err != nil {
		obj.templateDispatcher.RespondWithTemplate("register.html", 500, res, map[string]any{
			"err": err,
		})
		return
	}
	err = obj.daoSource.GetSessionStore(ctx).SaveSession(ctx, ses)
	if err != nil {
		obj.templateDispatcher.RespondWithTemplate("register.html", 500, res, map[string]any{
			"err": err,
		})
		return
	}

	idToken, refreshToken := ses.IssueTokens(obj.urlPrefix, obj.urlPrefix)

	jwt, err := jwtutil.ClaimsToJwt(idToken, key.Kid, rsaKey)
	if err != nil {
		obj.templateDispatcher.RespondWithTemplate("register.html", 500, res, map[string]any{
			"err": err,
		})
		return
	}
	rt, err := jwtutil.ClaimsToJwt(refreshToken, key.Kid, rsaKey)
	if err != nil {
		obj.templateDispatcher.RespondWithTemplate("register.html", 500, res, map[string]any{
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
		ctx := req.Context()
		claims := obj.userClaims(req)
		if claims != nil {
			res.Header().Add("Location", "/account")
			res.WriteHeader(302)
			return
		}

		if req.Method == http.MethodGet {
			obj.templateDispatcher.RespondWithTemplate("register.html", 200, res, nil)
			return
		}
		if req.Method == http.MethodPost {
			// handle LOGIN (or REGISTRATION)
			//
			req.ParseForm()

			userService := &users.UserService{
				UserStore: obj.daoSource.GetUserStore(ctx),
			}

			username := req.Form.Get("username")
			password := req.Form.Get("password")

			user, err := userService.AttemptUserRegistration(ctx, username, password)
			if errors.Is(err, users.ErrUserExists) {
				obj.templateDispatcher.RespondWithTemplate("register.html", 400, res, map[string]any{
					"err": "user already exists",
				})
				return
			} else if err != nil {
				obj.templateDispatcher.RespondWithTemplate("register.html", 500, res, map[string]any{
					"err": err,
				})
				return
			}
			if user == nil {
				obj.templateDispatcher.RespondWithTemplate("register.html", 400, res, map[string]any{
					"err": "unable to create user",
				})
				return
			}
			obj.createUserSession(ctx, res, user)
			return
		}

		res.WriteHeader(500)
	}
}

func (obj *acceptOidcHandler) loginHandler() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
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
			obj.templateDispatcher.RespondWithTemplate("login.html", 200, res, nil)
			return
		}
		if req.Method == http.MethodPost {
			req.ParseForm()
			userService := &users.UserService{
				UserStore: obj.daoSource.GetUserStore(ctx),
			}

			username := req.Form.Get("username")
			password := req.Form.Get("password")
			user, err := userService.AttemptUserLogin(ctx, username, password)
			if err != nil || user == nil {
				res.WriteHeader(500)
				return
			} else {
				obj.createUserSession(ctx, res, user)
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

func (obj *acceptOidcHandler) deauthClientHandler() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		// fmt.Printf("snippetHandler req.RequestURI: %v\n", 1req.RequestURI)
		clientId := strings.TrimSpace(strings.TrimPrefix(req.RequestURI, "/deauthorize/"))
		if clientId == "" {
			res.WriteHeader(http.StatusBadRequest)
			return
		}
		userId := obj.userId(req)
		if userId == "" {
			res.WriteHeader(http.StatusBadRequest)
			return
		}
		ctx := req.Context()
		scroller := &ddbutil.DepaginatedScroller[client.ClientAuthorization]{}
		err := obj.daoSource.GetClientAuthorizationStore(ctx).ClientAuthorizationsByUser(ctx, userId, scroller)
		if err != nil {
			fmt.Printf("%v\n", err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
		for _, authorizedClient := range scroller.Results {
			if authorizedClient.ClientId == clientId {
				err = obj.daoSource.GetClientAuthorizationStore(ctx).DeleteClientAuthorization(ctx, userId, clientId)
				if err != nil {
					fmt.Printf("%v\n", err)
					res.WriteHeader(http.StatusInternalServerError)
					return
				}

				// postback?!?
				res.Header().Add("Location", "/me")
				res.WriteHeader(http.StatusFound)
				return
			}
		}
		res.WriteHeader(http.StatusNotFound)
	}
}
