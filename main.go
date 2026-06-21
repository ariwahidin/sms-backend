package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"sms-backend/config"
	"sms-backend/database"
	"sms-backend/email"
	"sms-backend/handler"
	"sms-backend/middleware"
	"sms-backend/model"
	"sms-backend/repository"
	"sms-backend/routes"
	"sms-backend/scheduler"
	"sms-backend/service"
)

func main() {
	// ── 1. Load config ────────────────────────────────────────────────────────
	config.Load()

	// ── 2. Connect DB + AutoMigrate ───────────────────────────────────────────
	database.Connect()

	// ── 3. Seed admin user ────────────────────────────────────────────────────
	seedAdmin()

	// ── 4. Wire dependencies ──────────────────────────────────────────────────
	emailSvc := email.NewEmailService()

	userRepo := repository.NewUserRepository()
	locationRepo := repository.NewLocationRepository()
	deptRepo := repository.NewDepartmentRepository()
	reportRepo := repository.NewReportRepository()

	authSvc := service.NewAuthService(userRepo)
	locationSvc := service.NewLocationService(locationRepo)
	deptSvc := service.NewDepartmentService(deptRepo)
	userSvc := service.NewUserService(userRepo)
	reportSvc := service.NewReportService(reportRepo, emailSvc)

	authH := handler.NewAuthHandler(authSvc)
	reportH := handler.NewReportHandler(reportSvc, reportRepo)
	locationH := handler.NewLocationHandler(locationSvc)
	deptH := handler.NewDepartmentHandler(deptSvc)
	userH := handler.NewUserHandler(userSvc)

	// ── 5. Setup Fiber ────────────────────────────────────────────────────────
	app := fiber.New(fiber.Config{
		AppName:      "SMS Safety Management System v1.0",
		ErrorHandler: customErrorHandler,
	})

	app.Use(recover.New(recover.Config{
		EnableStackTrace: true,
	}))
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, POST, PUT, PATCH, DELETE, OPTIONS",
	}))
	app.Use(middleware.Logger())

	// Static file serving for uploads
	app.Static("/uploads", config.Cfg.Storage.UploadDir)

	// ── 6. Register routes ────────────────────────────────────────────────────
	routes.Setup(app, authSvc, authH, reportH, locationH, deptH, userH)

	for _, route := range app.GetRoutes() {
		fmt.Printf("[%s] %s\n", route.Method, route.Path)
	}

	// 404 fallback
	app.Use(func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Route not found",
			"path":  c.Path(),
		})
	})

	// ── 7. Start scheduler ────────────────────────────────────────────────────
	slaChecker := scheduler.NewSLAChecker(reportRepo, emailSvc)
	slaChecker.Start()

	// ── 8. Start server ───────────────────────────────────────────────────────
	go func() {
		port := config.Cfg.App.Port
		log.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		log.Printf("  SMS Safety Management System")
		log.Printf("  Server  : http://localhost:%s", port)
		log.Printf("  Env     : %s", config.Cfg.App.Env)
		log.Printf("  DB      : %s:%s/%s", config.Cfg.DB.Host, config.Cfg.DB.Port, config.Cfg.DB.Name)
		log.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

		if err := app.Listen(":" + port); err != nil {
			log.Fatalf("[SERVER] Failed to start: %v", err)
		}
	}()

	// ── 9. Graceful shutdown ──────────────────────────────────────────────────
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("[SERVER] Shutting down gracefully...")
	slaChecker.Stop()
	app.Shutdown()
	log.Println("[SERVER] Bye!")
}

func seedAdmin() {
	userRepo := repository.NewUserRepository()
	cfg := config.Cfg.Seeder

	if userRepo.ExistsByUsername(cfg.AdminUsername) {
		log.Printf("[SEEDER] Admin user '%s' already exists, skipping", cfg.AdminUsername)
		return
	}

	hashed, err := service.HashPassword(cfg.AdminPassword)
	if err != nil {
		log.Fatalf("[SEEDER] Failed to hash admin password: %v", err)
	}

	admin := &model.User{
		Username: cfg.AdminUsername,
		Password: hashed,
		Name:     cfg.AdminName,
		Email:    cfg.AdminEmail,
		Role:     model.RoleAdmin,
		IsActive: true,
	}

	if err := userRepo.Create(admin); err != nil {
		log.Fatalf("[SEEDER] Failed to create admin user: %v", err)
	}

	log.Printf("[SEEDER] Admin user created: %s (role: admin)", cfg.AdminUsername)
}

func customErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}
	log.Printf("[ERROR] %s %s → %d: %v", c.Method(), c.Path(), code, err)
	return c.Status(code).JSON(fiber.Map{"error": err.Error()})
}
