package internals

import (
	"github.com/MashukeAlam/grails/handlers"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func SetupRoutes(app *fiber.App, dbGorm *gorm.DB) {
	// Route to render the Slim template
	app.Get("/", func(c *fiber.Ctx) error {
		// Pass the title to the template
		return c.Render("index", fiber.Map{
			"Title": "Hello, Fiber with Slim!",
		}, "layouts/main")
	})

	// Dev routes
	Dev := app.Group("/dev")
	Dev.Get("/", handlers.GetDevView())
	Dev.Post("/", handlers.ProcessIncomingScaffoldData(dbGorm))
}
