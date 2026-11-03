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
	safetyCategoryH *handler.SafetyCategoryHandler,
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
	api.Get("/no-auth/safety-categories", safetyCategoryH.GetPublic)

	// ── Protected routes ──────────────────────────────────────────────────────
	protected := api.Group("", middleware.Auth(authSvc))

	// Reports
	protected.Get("/reports", reportH.GetAll)
	protected.Get("/reports/:id", reportH.GetByID)
	protected.Get("/reports/:id/files", reportH.GetFiles)
	protected.Get("/reports/:id/files/:file_id", reportH.ServeFile)
	protected.Post("/reports/:id/files", reportH.UploadFile)

	// Detection edit request
	protected.Post(
		"/reports/:id/edit-request",
		reportH.RequestDetectionEdit,
	)

	// PIC / Admin approval
	protected.Post(
		"/reports/:id/edit-request/:request_id/approve",
		middleware.RequireRole(model.RolePIC, model.RoleAdmin),
		reportH.ApproveDetectionEdit,
	)

	protected.Post(
		"/reports/:id/edit-request/:request_id/reject",
		middleware.RequireRole(model.RolePIC, model.RoleAdmin),
		reportH.RejectDetectionEdit,
	)

	// HOD / Admin only
	protected.Post("/reports/:id/approve", middleware.RequireRole(model.RoleHOD, model.RoleAdmin), reportH.Approve)
	protected.Post("/reports/:id/reject", middleware.RequireRole(model.RoleHOD, model.RoleAdmin), reportH.Reject)
	protected.Delete(
		"/reports/:id",
		middleware.RequireRole(model.RoleAdmin),
		reportH.Delete,
	)

	// PIC / Admin only
	protected.Patch("/reports/:id/classify", middleware.RequireRole(model.RolePIC, model.RoleAdmin), reportH.Classify)
	protected.Patch("/reports/:id/investigate", middleware.RequireRole(model.RolePIC, model.RoleAdmin), reportH.Investigate)

	// Dashboard
	protected.Get("/dashboard", reportH.Dashboard)

	// Departments
	protected.Get("/departments", middleware.RequireRole(model.RoleAdmin, model.RolePIC, model.RoleHOD), deptH.GetAll)
	protected.Post("/departments", middleware.RequireRole(model.RoleAdmin, model.RolePIC), deptH.Create)
	protected.Patch("/departments/:id", middleware.RequireRole(model.RoleAdmin, model.RolePIC), deptH.Update)

	// Locations
	protected.Get("/locations", middleware.RequireRole(model.RoleAdmin, model.RolePIC, model.RoleHOD), locationH.GetAll)
	protected.Post("/locations", middleware.RequireRole(model.RoleAdmin, model.RolePIC), locationH.Create)
	protected.Patch("/locations/:id", middleware.RequireRole(model.RoleAdmin, model.RolePIC), locationH.Update)

	// Safety Categories
	protected.Get("/safety-categories", middleware.RequireRole(model.RoleAdmin, model.RolePIC, model.RoleHOD), safetyCategoryH.GetAll)
	protected.Get("/safety-categories/:id", middleware.RequireRole(model.RoleAdmin, model.RolePIC, model.RoleHOD), safetyCategoryH.GetByID)
	protected.Post("/safety-categories", middleware.RequireRole(model.RoleAdmin, model.RolePIC), safetyCategoryH.Create)
	protected.Patch("/safety-categories/:id", middleware.RequireRole(model.RoleAdmin, model.RolePIC), safetyCategoryH.Update)
	protected.Delete("/safety-categories/:id", middleware.RequireRole(model.RoleAdmin, model.RolePIC), safetyCategoryH.Delete)

	// Users
	protected.Get("/users", middleware.RequireRole(model.RoleAdmin), userH.GetAll)
	protected.Post("/users", middleware.RequireRole(model.RoleAdmin), userH.Create)
	protected.Patch("/users/:id", middleware.RequireRole(model.RoleAdmin), userH.Update)
}
