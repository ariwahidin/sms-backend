package routes

import (
	"github.com/gofiber/fiber/v2"

	"sms-backend/handler"
	"sms-backend/middleware"
	"sms-backend/model"
	"sms-backend/service"
)

func Setup(
	app *fiber.App,
	authSvc *service.AuthService,
	authH *handler.AuthHandler,
	reportH *handler.ReportHandler,
	locationH *handler.LocationHandler,
	deptH *handler.DepartmentHandler,
	userH *handler.UserHandler,
) {
	api := app.Group("/api")

	// ── Public routes ─────────────────────────────────────────────────────────
	api.Post("/auth/login", authH.Login)

	// QR token lookup (for prefill form, no auth needed)
	api.Get("/locations/:qr_token", locationH.GetByQRToken)

	// Declaration (no auth — anonymous allowed)
	api.Post("/reports", reportH.Declare)
	api.Get("/no-auth/departments", deptH.GetAll)
	api.Get("/no-auth/locations", locationH.GetAll)

	// ── Protected routes ──────────────────────────────────────────────────────
	protected := api.Group("", middleware.Auth(authSvc))

	// Reports
	protected.Get("/reports", reportH.GetAll)
	protected.Get("/reports/:id", reportH.GetByID)
	protected.Get("/reports/:id/files", reportH.GetFiles)
	protected.Get("/reports/:id/files/:file_id", reportH.ServeFile)
	protected.Post("/reports/:id/files", reportH.UploadFile)

	// HOD / Admin only
	protected.Post("/reports/:id/approve", middleware.RequireRole(model.RoleHOD, model.RoleAdmin), reportH.Approve)
	protected.Post("/reports/:id/reject", middleware.RequireRole(model.RoleHOD, model.RoleAdmin), reportH.Reject)

	// PIC / Admin only
	protected.Patch("/reports/:id/classify", middleware.RequireRole(model.RolePIC, model.RoleAdmin), reportH.Classify)
	protected.Patch("/reports/:id/investigate", middleware.RequireRole(model.RolePIC, model.RoleAdmin), reportH.Investigate)

	// Dashboard
	protected.Get("/dashboard", reportH.Dashboard)

	// Departments
	protected.Get("/departments", deptH.GetAll)
	protected.Post("/departments", middleware.RequireRole(model.RoleAdmin), deptH.Create)
	protected.Patch("/departments/:id", middleware.RequireRole(model.RoleAdmin), deptH.Update)

	// Locations
	protected.Get("/locations", middleware.RequireRole(model.RoleAdmin), locationH.GetAll)
	protected.Post("/locations", middleware.RequireRole(model.RoleAdmin), locationH.Create)
	protected.Patch("/locations/:id", middleware.RequireRole(model.RoleAdmin), locationH.Update)

	// Users
	protected.Get("/users", middleware.RequireRole(model.RoleAdmin), userH.GetAll)
	protected.Post("/users", middleware.RequireRole(model.RoleAdmin), userH.Create)
	protected.Patch("/users/:id", middleware.RequireRole(model.RoleAdmin), userH.Update)
}
