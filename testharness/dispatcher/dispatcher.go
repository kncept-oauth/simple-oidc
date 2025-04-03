package dispatcher

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/kncept-oauth/simple-oidc/service/authorizer"
	servicedispatcher "github.com/kncept-oauth/simple-oidc/service/dispatcher"
	"github.com/kncept-oauth/simple-oidc/testharness/webcontent"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/template/html/v2"

	fiberoidc "github.com/kncept/fiber-oidc"
)

func NewApplication(daoSource servicedispatcher.DaoSource) *fiber.App {
	ctx := context.Background()
	fmt.Printf("New Testharness Application\n")
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	viewEngine := html.NewFileSystem(http.FS(webcontent.Views), ".html")
	viewEngine.AddFunc("Clients", func() []authorizer.Client {
		clients, _ := daoSource.GetClientStore().List()
		return clients
	})

	app := fiber.New(
		fiber.Config{
			Views: viewEngine,
		},
	)

	// most basic client
	client := authorizer.ClientStruct{
		ClientId: "random-client-id", //uuid.NewString(),
		AllowedRedirectUris: []string{
			"http://localhost:3000/oauth2/callback",
		},
		// ClientSecret: uuid.NewString(),
	}
	daoSource.GetClientStore().Save(client)

	fiberOidcConfig := &fiberoidc.Config{
		Issuer:         "https://localhost:8443",
		ClientId:       client.ClientId,
		ClientSecret:   "todo",
		RedirectUri:    "http://localhost:3000/oauth2/callback",
		AuthCookieName: "bearer-auth",
	}
	fiberOidc, err := fiberoidc.New(ctx, fiberOidcConfig)
	if err != nil {
		panic(err)
	}

	app.Use(
		compress.New(),
	)

	app.Get(fiberOidc.CallbackPath(), fiberOidc.CallbackHandler())

	app.Use("/static", filesystem.New(filesystem.Config{
		Root:   http.FS(webcontent.Static),
		Browse: true,
	}))
	app.Get("/", fiberOidc.UnprotectedRoute(), func(c *fiber.Ctx) error {
		idToken := fiberoidc.IdTokenFromContext(c)
		bind := make(map[string]any)
		bind["ClientId"] = client.ClientId
		bind["Issuer"] = fiberOidcConfig.Issuer
		bind["RedirectUri"] = fiberOidcConfig.RedirectUri
		if idToken == nil {
			bind["LoggedIn"] = false
		} else {
			bind["LoggedIn"] = true
			bind["IdToken"] = fmt.Sprintf("%+v", idToken)
		}
		return c.Render("index", bind)
	})

	app.Post("/", func(c *fiber.Ctx) error {
		// dynamically pull form type & details to perform operation
		// then redirect back to index
		return c.Redirect("/")
	})
	return app
}
