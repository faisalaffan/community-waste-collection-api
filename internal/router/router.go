package router

import (
	"github.com/gofiber/fiber/v3"

	"github.com/faisalaffan/community-waste-collection-api/docs"
	"github.com/faisalaffan/community-waste-collection-api/internal/handler"
	"github.com/faisalaffan/community-waste-collection-api/internal/middleware"
)

// swaggerHTML is the Swagger UI page served at /swagger.
const swaggerHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>Swagger UI - Community Waste Collection API</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css" />
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js" crossorigin></script>
  <script>
    SwaggerUIBundle({
      url: "/swagger/doc.json",
      dom_id: "#swagger-ui",
    });
  </script>
</body>
</html>`

func Setup(
	hh *handler.HouseholdHandler,
	ph *handler.PickupHandler,
	pmh *handler.PaymentHandler,
	rh *handler.ReportHandler,
) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName: "Community Waste Collection API",
	})

	// Frontend SPA & Assets
	app.Get("/assets/:filename", func(c fiber.Ctx) error {
		return c.SendFile("assets/" + c.Params("filename"))
	})
	app.Get("/", func(c fiber.Ctx) error {
		c.Type("html", "utf-8")
		return c.SendFile("web/index.html")
	})
	app.Get("/app.js", func(c fiber.Ctx) error {
		c.Type("js", "utf-8")
		return c.SendFile("web/app.js")
	})

	api := app.Group("/api")

	// Households
	households := api.Group("/households")
	households.Post("/", hh.Create)
	households.Get("/", hh.List)
	households.Get("/:id", hh.Get)
	households.Put("/:id", hh.Update)
	households.Delete("/:id", hh.Delete)

	// Pickups
	pickups := api.Group("/pickups")
	pickups.Post("/", middleware.RateLimitPickup(), ph.Create)
	pickups.Get("/", ph.List)
	pickups.Put("/:id", ph.Update)
	pickups.Delete("/:id", ph.Delete)
	pickups.Put("/:id/schedule", ph.Schedule)
	pickups.Put("/:id/complete", ph.Complete)
	pickups.Put("/:id/cancel", ph.Cancel)

	// Payments
	payments := api.Group("/payments")
	payments.Post("/", pmh.Create)
	payments.Get("/", pmh.List)
	payments.Put("/:id/confirm", pmh.Confirm)

	// Reports
	reports := api.Group("/reports")
	reports.Get("/waste-summary", rh.WasteSummary)
	reports.Get("/payment-summary", rh.PaymentSummary)
	reports.Get("/households/:id/history", rh.HouseholdHistory)

	// Swagger UI
	app.Get("/swagger/doc.json", func(c fiber.Ctx) error {
		docs.SwaggerInfo.Host = c.Host()
		c.Type("json", "utf-8")
		return c.SendString(docs.SwaggerInfo.ReadDoc())
	})
	app.Get("/swagger", func(c fiber.Ctx) error {
		c.Type("html", "utf-8")
		return c.SendString(swaggerHTML)
	})

	return app
}
