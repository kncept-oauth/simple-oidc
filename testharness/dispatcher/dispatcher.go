package dispatcher

import (
	"fmt"
	"net/http"

	servicedispatcher "github.com/kncept-oauth/simple-oidc/service/dispatcher"
	"github.com/kncept-oauth/simple-oidc/testharness/webcontent"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/template/html/v2"
)

func NewApplication(daoSource servicedispatcher.DaoSource) *fiber.App {
	fmt.Printf("New Testharness Application\n")

	viewEngine := html.NewFileSystem(http.FS(webcontent.Fs), ".html")

	app := fiber.New(
		fiber.Config{
			Views: viewEngine,
		},
	)
	app.Use(
		compress.New(),
	)

	app.Get("/", Index)

	return app
}

func Index(c *fiber.Ctx) error {
	return c.Render("index", nil)
}
