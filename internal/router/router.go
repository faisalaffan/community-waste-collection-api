package router

import (
	"github.com/gofiber/fiber/v3"

	"github.com/faisalaffan/community-waste-collection-api/internal/handler"
	"github.com/faisalaffan/community-waste-collection-api/internal/middleware"
)

func Setup(
	hh *handler.HouseholdHandler,
	ph *handler.PickupHandler,
	pmh *handler.PaymentHandler,
	rh *handler.ReportHandler,
) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName: "Community Waste Collection API",
	})

	api := app.Group("/api")

	// Households
	households := api.Group("/households")
	households.Post("/", hh.Create)
	households.Get("/", hh.List)
	households.Get("/:id", hh.Get)
	households.Delete("/:id", hh.Delete)

	// Pickups
	pickups := api.Group("/pickups")
	pickups.Post("/", middleware.RateLimitPickup(), ph.Create)
	pickups.Get("/", ph.List)
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

	return app
}
