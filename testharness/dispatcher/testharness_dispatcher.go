package dispatcher

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/kncept-oauth/simple-oidc/service/client"
	"github.com/kncept-oauth/simple-oidc/service/dao"
	"github.com/kncept-oauth/simple-oidc/service/users"
	"github.com/kncept-oauth/simple-oidc/testharness/webcontent"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/template/html/v2"

	fiberoidc "github.com/kncept/fiber-oidc"
	"github.com/kncept/fiber-oidc/provider"
)

const staticClientId = "static-client-id"

type testApp struct {
	message         string
	fiberOidcConfig *fiberoidc.Config
	fiberOidc       fiberoidc.FiberOidc
	daoSource       dao.DaoSource
}

func NewApplication(daoSource dao.DaoSource) *fiber.App {
	testApp := &testApp{
		daoSource: daoSource,
	}
	ctx := context.Background()
	viewEngine := html.NewFileSystem(http.FS(webcontent.Views), ".html")
	viewEngine.AddFunc("Clients", func() []*client.Client {
		clients, _ := daoSource.GetClientStore(ctx).ListClients(ctx)
		return clients
	})
	app := fiber.New(
		fiber.Config{
			Views:          viewEngine,
			ReadBufferSize: 4096 * 4,
		},
	)
	staticClient, err := daoSource.GetClientStore(ctx).GetClient(ctx, staticClientId)
	if err != nil {
		panic(err)
	}
	if staticClient == nil {
		staticClient = &client.Client{
			ClientId: staticClientId,
			AllowedRedirectUris: []string{
				"https://localhost:3000/oauth2/callback",
			},
			// Audiences: []string{
			// 	"https://localhost:3000/",
			// },
		}
		daoSource.GetClientStore(ctx).SaveClient(ctx, staticClient)
	}

	testApp.fiberOidcConfig = &fiberoidc.Config{
		OidcProviderConfig: provider.OidcProviderConfig{
			Issuer:       "https://localhost:8443",
			ClientId:     staticClientId,
			ClientSecret: fmt.Sprintf("client-secret-%v", uuid.NewString()),
			RedirectUri:  "https://localhost:3000/oauth2/callback",
		},
		WebAppConfig: fiberoidc.WebAppConfig{
			AuthCookieName:        "bearer-auth",
			AuthRefreshCookieName: "refresh-auth",
		},
	}
	fiberOidc, err := fiberoidc.New(ctx, testApp.fiberOidcConfig)
	if err != nil {
		panic(err)
	}
	testApp.fiberOidc = fiberOidc

	app.Use(
		compress.New(),
	)

	app.Get(fiberOidc.CallbackPath(), fiberOidc.CallbackHandler())

	app.Use("/static", filesystem.New(filesystem.Config{
		Root:   http.FS(webcontent.Static),
		Browse: true,
	}))
	app.Get("/", fiberOidc.UnprotectedRoute(), testApp.GetIndex)
	// app.Get("/unprotected", fiberOidc.UnprotectedRoute(), testApp.GetIndex)
	// app.Get("/protected", fiberOidc.ProtectedRoute(), testApp.GetIndex)

	app.Post("/", fiberOidc.UnprotectedRoute(), testApp.PostIndex)
	// app.Post("/unprotected", fiberOidc.UnprotectedRoute(), testApp.PostIndex)
	// app.Post("/protected", fiberOidc.ProtectedRoute(), testApp.PostIndex)
	return app
}

func (obj *testApp) GetIndex(c *fiber.Ctx) error {
	idToken := fiberoidc.GoOidcToken(c)
	providerAuth := fiberoidc.ProviderAuth(c)
	ctx := c.Context()
	bind := make(map[string]any)
	bind["ClientId"] = staticClientId
	bind["Issuer"] = obj.fiberOidcConfig.Issuer
	bind["RedirectUri"] = obj.fiberOidcConfig.RedirectUri
	if idToken == nil {
		bind["LoggedIn"] = false
	} else {
		bind["LoggedIn"] = true
		bind["IdToken"] = idToken
		bind["Oauth2Token"] = providerAuth.GetOauth2Token()
	}

	// bind // users
	userStore := obj.daoSource.GetUserStore(ctx)
	allUsers := make([]*users.OidcUser, 0)
	userStore.EnumerateUsers(ctx, func(user *users.OidcUser) bool {
		allUsers = append(allUsers, user)
		return true
	})
	bind["AllUsers"] = allUsers
	if obj.message != "" {
		bind["Message"] = obj.message
		bind["HasMessage"] = true
		obj.message = "" // (globally shared) state only persists for ONE render
	} else {
		bind["HasMessage"] = false
	}
	bind["DatastoreType"] = obj.daoSource.GetDaoSourceDescription()

	return c.Render("index", bind)
}

func (obj *testApp) PostIndex(c *fiber.Ctx) error {
	ctx := c.Context()
	payload := struct {
		Op string
		Id string
	}{}
	err := c.BodyParser(&payload)
	if err != nil {
		return err
	}
	switch payload.Op {
	case "init":
		obj.fiberOidc.Providers().Initialize(c.Context())
	case "create":
		c := &client.Client{
			ClientId: uuid.NewString(),
			AllowedRedirectUris: []string{
				"https://localhost:3000/oauth2/callback",
			},
		}
		err = obj.daoSource.GetClientStore(ctx).SaveClient(ctx, c)
		if err == nil {
			obj.message = fmt.Sprintf("Successfully created client:\n%v", c.ClientId)
		} else {
			obj.message = fmt.Sprintf("Error occurred:\n%v", err)
		}
	case "delete":
		c, err := obj.daoSource.GetClientStore(ctx).GetClient(ctx, payload.Id)
		if err != nil {
			return err
		}
		if c == nil {
			return errors.New("no client with id " + payload.Id)
		}
		err = obj.daoSource.GetClientStore(ctx).RemoveClient(ctx, payload.Id)
		if err != nil {
			return err
		}
	case "logout":
		c.Cookie(&fiber.Cookie{
			Name:  obj.fiberOidcConfig.AuthCookieName,
			Value: "",
		})
	case "userinfo-callback":
		{

			goOidcProvider, err := obj.fiberOidc.Providers().GoOidcProvider(ctx)
			if err != nil {
				return err
			}

			userInfo, err := goOidcProvider.UserInfo(ctx, fiberoidc.Oauth2TokenSource(c))
			obj.message = fmt.Sprintf("%+v", userInfo)
		}
	}

	// dynamically pull form type & details to perform operation
	// then redirect back to index
	return c.Redirect("/")
}
