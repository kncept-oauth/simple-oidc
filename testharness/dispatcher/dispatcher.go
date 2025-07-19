package dispatcher

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/kncept-oauth/simple-oidc/service/client"
	"github.com/kncept-oauth/simple-oidc/service/dao"
	"github.com/kncept-oauth/simple-oidc/testharness/webcontent"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/template/html/v2"

	fiberoidc "github.com/kncept/fiber-oidc"
)

const staticClientId = "static-client-id"

func NewApplication(daoSource dao.DaoSource) *fiber.App {
	ctx := context.Background()
	fmt.Printf("New Testharness Application\n")
	// http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	viewEngine := html.NewFileSystem(http.FS(webcontent.Views), ".html")
	viewEngine.AddFunc("Clients", func() []*client.Client {
		clients, _ := daoSource.GetClientStore().ListClients(ctx)
		return clients
	})

	app := fiber.New(
		fiber.Config{
			Views:          viewEngine,
			ReadBufferSize: 4096 * 4,
		},
	)

	// most basic client
	c := &client.Client{
		ClientId: staticClientId,
		AllowedRedirectUris: []string{
			"https://localhost:3000/oauth2/callback",
		},
		// Audiences: []string{
		// 	"https://localhost:3000/",
		// },
	}
	daoSource.GetClientStore().SaveClient(ctx, c)

	// c = &client.Client{
	// 	ClientId: uuid.NewString(),
	// 	AllowedRedirectUris: []string{
	// 		"https://localhost:3000/oauth2/callback",
	// 	},
	// 	Audiences: []string{
	// 		"https://localhost:3000/",
	// 	},
	// }
	// daoSource.GetClientStore().SaveClient(c)

	fiberOidcConfig := &fiberoidc.Config{
		Issuer:         "https://localhost:8443",
		ClientId:       staticClientId,
		ClientSecret:   fmt.Sprintf("client-secret-%v", uuid.NewString()),
		RedirectUri:    "https://localhost:3000/oauth2/callback",
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
		bind["ClientId"] = staticClientId
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
		payload := struct {
			Op string
			Id string
		}{}
		err := c.BodyParser(&payload)
		if err != nil {
			return err
		}
		switch payload.Op {
		case "create":
			c := &client.Client{
				ClientId: uuid.NewString(),
				AllowedRedirectUris: []string{
					"https://localhost:3000/oauth2/callback",
				},
			}
			daoSource.GetClientStore().SaveClient(ctx, c)
		case "delete":
			c, err := daoSource.GetClientStore().GetClient(ctx, payload.Id)
			if err != nil {
				return err
			}
			if c == nil {
				return errors.New("no client with id " + payload.Id)
			}
			err = daoSource.GetClientStore().RemoveClient(ctx, payload.Id)
			if err != nil {
				return err
			}

		}

		// dynamically pull form type & details to perform operation
		// then redirect back to index
		return c.Redirect("/")
	})
	return app
}
