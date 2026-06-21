package middleware

import (
	"fmt"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
)

// routeInfo maps route pattern → {file, function}
var routeInfo = map[string]struct{ File, Func string }{
	// Auth
	"POST /api/auth/login": {"handler/auth_handler.go", "Login()"},

	// Reports
	"POST /api/reports":                  {"handler/report_handler.go", "Declare()"},
	"GET /api/reports":                   {"handler/report_handler.go", "GetAll()"},
	"GET /api/reports/:id":               {"handler/report_handler.go", "GetByID()"},
	"PATCH /api/reports/:id/classify":    {"handler/report_handler.go", "Classify()"},
	"PATCH /api/reports/:id/investigate": {"handler/report_handler.go", "Investigate()"},
	"POST /api/reports/:id/approve":      {"handler/report_handler.go", "Approve()"},
	"POST /api/reports/:id/reject":       {"handler/report_handler.go", "Reject()"},
	"GET /api/reports/:id/files":         {"handler/report_handler.go", "GetFiles()"},
	"POST /api/reports/:id/files":        {"handler/report_handler.go", "UploadFile()"},

	// Dashboard
	"GET /api/dashboard": {"handler/report_handler.go", "Dashboard()"},

	// Locations
	"GET /api/locations":           {"handler/location_handler.go", "GetAll()"},
	"POST /api/locations":          {"handler/location_handler.go", "Create()"},
	"GET /api/locations/:qr_token": {"handler/location_handler.go", "GetByQRToken()"},

	// Departments
	"GET /api/departments":  {"handler/department_handler.go", "GetAll()"},
	"POST /api/departments": {"handler/department_handler.go", "Create()"},

	// Users
	"GET /api/users":  {"handler/user_handler.go", "GetAll()"},
	"POST /api/users": {"handler/user_handler.go", "Create()"},
}

func Logger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		err := c.Next()

		duration := time.Since(start)
		method := c.Method()
		path := c.Path()
		status := c.Response().StatusCode()
		ip := c.IP()

		// build route key to look up file/func info
		routeKey := fmt.Sprintf("%s %s", method, c.Route().Path)
		info, ok := routeInfo[routeKey]
		if !ok {
			info = struct{ File, Func string }{"unknown", "unknown"}
		}

		// color status
		statusColor := colorStatus(status)

		log.Printf("[SMS] %s | %s %s → %s | %s | %s | %s | %dms | ip:%s",
			time.Now().Format("2006/01/02 15:04:05"),
			method,
			path,
			statusColor,
			info.File,
			info.Func,
			c.Route().Path,
			duration.Milliseconds(),
			ip,
		)

		return err
	}
}

func colorStatus(status int) string {
	switch {
	case status >= 500:
		return fmt.Sprintf("🔴 %d", status)
	case status >= 400:
		return fmt.Sprintf("🟡 %d", status)
	case status >= 300:
		return fmt.Sprintf("🔵 %d", status)
	default:
		return fmt.Sprintf("🟢 %d", status)
	}
}
