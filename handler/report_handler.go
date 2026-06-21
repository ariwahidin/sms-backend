// // package handler

// // import (
// // 	"fmt"
// // 	"io"
// // 	"os"
// // 	"path/filepath"
// // 	"strconv"
// // 	"strings"
// // 	"time"

// // 	"github.com/gofiber/fiber/v2"

// // 	"sms-backend/config"
// // 	"sms-backend/middleware"
// // 	"sms-backend/model"
// // 	"sms-backend/repository"
// // 	"sms-backend/service"
// // )

// // type ReportHandler struct {
// // 	reportSvc  *service.ReportService
// // 	reportRepo *repository.ReportRepository
// // }

// // func NewReportHandler(reportSvc *service.ReportService, reportRepo *repository.ReportRepository) *ReportHandler {
// // 	return &ReportHandler{reportSvc: reportSvc, reportRepo: reportRepo}
// // }

// // func (h *ReportHandler) Declare(c *fiber.Ctx) error {
// // 	var input service.DeclareInput
// // 	if err := c.BodyParser(&input); err != nil {
// // 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
// // 	}

// // 	report, err := h.reportSvc.Declare(input)
// // 	if err != nil {
// // 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
// // 	}

// // 	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
// // 		"message": "Declaration successful",
// // 		"code":    report.Code,
// // 		"id":      report.ID,
// // 	})
// // }

// // func (h *ReportHandler) GetAll(c *fiber.Ctx) error {
// // 	page, _ := strconv.Atoi(c.Query("page", "1"))
// // 	pageSize, _ := strconv.Atoi(c.Query("page_size", "25"))

// // 	filters := map[string]interface{}{
// // 		"status":        c.Query("status"),
// // 		"site":          c.Query("site"),
// // 		"department_id": c.Query("department_id"),
// // 		"risk_level":    c.Query("risk_level"),
// // 		"date_from":     c.Query("date_from"),
// // 		"date_to":       c.Query("date_to"),
// // 		"page":          page,
// // 		"page_size":     pageSize,
// // 	}

// // 	reports, total, err := h.reportRepo.FindAll(filters)
// // 	if err != nil {
// // 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
// // 	}

// // 	return c.JSON(fiber.Map{
// // 		"data":      reports,
// // 		"total":     total,
// // 		"page":      page,
// // 		"page_size": pageSize,
// // 		"pages":     (total + int64(pageSize) - 1) / int64(pageSize),
// // 	})
// // }

// // func (h *ReportHandler) GetByID(c *fiber.Ctx) error {
// // 	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
// // 	if err != nil {
// // 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
// // 	}

// // 	report, err := h.reportRepo.FindByID(uint(id))
// // 	if err != nil {
// // 		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Report not found"})
// // 	}

// // 	return c.JSON(report)
// // }

// // func (h *ReportHandler) Classify(c *fiber.Ctx) error {
// // 	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
// // 	if err != nil {
// // 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
// // 	}

// // 	var input service.ClassifyInput
// // 	if err := c.BodyParser(&input); err != nil {
// // 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
// // 	}

// // 	claims := middleware.GetClaims(c)
// // 	input.ByUserID = claims.UserID
// // 	input.ByUserName = claims.Name

// // 	report, err := h.reportSvc.Classify(uint(id), input)
// // 	if err != nil {
// // 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
// // 	}

// // 	return c.JSON(fiber.Map{"message": "Classification confirmed", "report": report})
// // }

// // func (h *ReportHandler) Investigate(c *fiber.Ctx) error {
// // 	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
// // 	if err != nil {
// // 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
// // 	}

// // 	var input service.InvestigateInput
// // 	if err := c.BodyParser(&input); err != nil {
// // 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
// // 	}

// // 	claims := middleware.GetClaims(c)
// // 	input.ByUserID = claims.UserID
// // 	input.ByUserName = claims.Name

// // 	report, err := h.reportSvc.Investigate(uint(id), input)
// // 	if err != nil {
// // 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
// // 	}

// // 	return c.JSON(fiber.Map{"message": "Investigation submitted", "report": report})
// // }

// // func (h *ReportHandler) Approve(c *fiber.Ctx) error {
// // 	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
// // 	if err != nil {
// // 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
// // 	}

// // 	var body struct {
// // 		Comment string `json:"comment"`
// // 	}
// // 	c.BodyParser(&body)

// // 	claims := middleware.GetClaims(c)
// // 	report, err := h.reportSvc.Approve(uint(id), body.Comment, claims.UserID, claims.Name)
// // 	if err != nil {
// // 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
// // 	}

// // 	return c.JSON(fiber.Map{"message": "Report approved and closed", "report": report})
// // }

// // func (h *ReportHandler) Reject(c *fiber.Ctx) error {
// // 	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
// // 	if err != nil {
// // 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
// // 	}

// // 	var body struct {
// // 		Comment string `json:"comment"`
// // 	}
// // 	if err := c.BodyParser(&body); err != nil {
// // 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
// // 	}
// // 	if body.Comment == "" {
// // 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Rejection reason (comment) is required"})
// // 	}

// // 	claims := middleware.GetClaims(c)
// // 	report, err := h.reportSvc.Reject(uint(id), body.Comment, claims.UserID, claims.Name)
// // 	if err != nil {
// // 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
// // 	}

// // 	return c.JSON(fiber.Map{"message": "Report rejected, returned to investigation", "report": report})
// // }

// // func (h *ReportHandler) GetFiles(c *fiber.Ctx) error {
// // 	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
// // 	if err != nil {
// // 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
// // 	}

// // 	files, err := h.reportRepo.FindFiles(uint(id))
// // 	if err != nil {
// // 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
// // 	}

// // 	return c.JSON(files)
// // }

// // func (h *ReportHandler) UploadFile(c *fiber.Ctx) error {
// // 	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
// // 	if err != nil {
// // 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
// // 	}

// // 	// verify report exists
// // 	if _, err := h.reportRepo.FindByID(uint(id)); err != nil {
// // 		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Report not found"})
// // 	}

// // 	fileHeader, err := c.FormFile("file")
// // 	if err != nil {
// // 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "File is required"})
// // 	}

// // 	maxSize := config.Cfg.Storage.MaxFileSizeMB * 1024 * 1024
// // 	if fileHeader.Size > maxSize {
// // 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
// // 			"error": fmt.Sprintf("File too large. Max size: %dMB", config.Cfg.Storage.MaxFileSizeMB),
// // 		})
// // 	}

// // 	// ensure upload dir exists
// // 	uploadDir := filepath.Join(config.Cfg.Storage.UploadDir, fmt.Sprintf("report_%d", id))
// // 	if err := os.MkdirAll(uploadDir, 0755); err != nil {
// // 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create upload directory"})
// // 	}

// // 	// unique filename
// // 	ext := filepath.Ext(fileHeader.Filename)
// // 	uniqueName := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), sanitizeFileName(fileHeader.Filename), ext)
// // 	savePath := filepath.Join(uploadDir, uniqueName)

// // 	src, err := fileHeader.Open()
// // 	if err != nil {
// // 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to open file"})
// // 	}
// // 	defer src.Close()

// // 	dst, err := os.Create(savePath)
// // 	if err != nil {
// // 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to save file"})
// // 	}
// // 	defer dst.Close()

// // 	if _, err := io.Copy(dst, src); err != nil {
// // 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to write file"})
// // 	}

// // 	claims := middleware.GetClaims(c)
// // 	var uploaderID *uint
// // 	if claims != nil {
// // 		uid := claims.UserID
// // 		uploaderID = &uid
// // 	}

// // 	reportFile := &model.ReportFile{
// // 		ReportID:   uint(id),
// // 		FileName:   fileHeader.Filename,
// // 		FilePath:   savePath,
// // 		FileSize:   fileHeader.Size,
// // 		MimeType:   fileHeader.Header.Get("Content-Type"),
// // 		UploadedBy: uploaderID,
// // 	}

// // 	if err := h.reportRepo.AddFile(reportFile); err != nil {
// // 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to save file record"})
// // 	}

// // 	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
// // 		"message": "File uploaded successfully",
// // 		"file":    reportFile,
// // 	})
// // }

// // func (h *ReportHandler) Dashboard(c *fiber.Ctx) error {
// // 	filters := map[string]interface{}{
// // 		"site":      c.Query("site"),
// // 		"date_from": c.Query("date_from"),
// // 		"date_to":   c.Query("date_to"),
// // 	}

// // 	stats, err := h.reportRepo.DashboardStats(filters)
// // 	if err != nil {
// // 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
// // 	}

// // 	return c.JSON(stats)
// // }

// // func sanitizeFileName(name string) string {
// // 	base := strings.TrimSuffix(name, filepath.Ext(name))
// // 	safe := strings.Map(func(r rune) rune {
// // 		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
// // 			return r
// // 		}
// // 		return '_'
// // 	}, base)
// // 	if len(safe) > 50 {
// // 		safe = safe[:50]
// // 	}
// // 	return safe
// // }

// package handler

// import (
// 	"fmt"
// 	"strconv"

// 	"github.com/gofiber/fiber/v2"

// 	"sms-backend/config"
// 	"sms-backend/middleware"
// 	"sms-backend/model"
// 	"sms-backend/repository"
// 	"sms-backend/service"
// )

// type ReportHandler struct {
// 	reportSvc  *service.ReportService
// 	reportRepo *repository.ReportRepository
// }

// func NewReportHandler(reportSvc *service.ReportService, reportRepo *repository.ReportRepository) *ReportHandler {
// 	return &ReportHandler{reportSvc: reportSvc, reportRepo: reportRepo}
// }

// func (h *ReportHandler) Declare(c *fiber.Ctx) error {
// 	input := service.DeclareInput{
// 		Description:  c.FormValue("description"),
// 		RiskLevel:    model.RiskLevel(c.FormValue("risk_level")),
// 		ReporterName: c.FormValue("reporter_name"),
// 	}

// 	if locID := c.FormValue("location_id"); locID != "" {
// 		id, err := strconv.ParseUint(locID, 10, 64)
// 		if err != nil {
// 			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid location_id"})
// 		}
// 		uid := uint(id)
// 		input.LocationID = &uid
// 	}

// 	if deptID := c.FormValue("department_id"); deptID != "" {
// 		id, err := strconv.ParseUint(deptID, 10, 64)
// 		if err != nil {
// 			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid department_id"})
// 		}
// 		uid := uint(id)
// 		input.DepartmentID = &uid
// 	}

// 	// ambil semua files dari field "files"
// 	form, err := c.MultipartForm()
// 	if err == nil && form != nil && len(form.File["files"]) > 0 {
// 		maxFiles := 10
// 		if len(form.File["files"]) > maxFiles {
// 			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
// 				"error": fmt.Sprintf("Maximum %d files allowed", maxFiles),
// 			})
// 		}
// 		input.Files = form.File["files"]
// 	}

// 	report, err := h.reportSvc.Declare(input)
// 	if err != nil {
// 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
// 	}

// 	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
// 		"message": "Declaration successful",
// 		"code":    report.Code,
// 		"id":      report.ID,
// 	})
// }

// func (h *ReportHandler) GetAll(c *fiber.Ctx) error {
// 	page, _ := strconv.Atoi(c.Query("page", "1"))
// 	pageSize, _ := strconv.Atoi(c.Query("page_size", "25"))

// 	filters := map[string]interface{}{
// 		"status":        c.Query("status"),
// 		"site":          c.Query("site"),
// 		"department_id": c.Query("department_id"),
// 		"risk_level":    c.Query("risk_level"),
// 		"date_from":     c.Query("date_from"),
// 		"date_to":       c.Query("date_to"),
// 		"page":          page,
// 		"page_size":     pageSize,
// 	}

// 	reports, total, err := h.reportRepo.FindAll(filters)
// 	if err != nil {
// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
// 	}

// 	return c.JSON(fiber.Map{
// 		"data":      reports,
// 		"total":     total,
// 		"page":      page,
// 		"page_size": pageSize,
// 		"pages":     (total + int64(pageSize) - 1) / int64(pageSize),
// 	})
// }

// func (h *ReportHandler) GetByID(c *fiber.Ctx) error {
// 	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
// 	if err != nil {
// 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
// 	}

// 	report, err := h.reportRepo.FindByID(uint(id))
// 	if err != nil {
// 		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Report not found"})
// 	}

// 	return c.JSON(report)
// }

// func (h *ReportHandler) Classify(c *fiber.Ctx) error {
// 	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
// 	if err != nil {
// 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
// 	}

// 	var input service.ClassifyInput
// 	if err := c.BodyParser(&input); err != nil {
// 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
// 	}

// 	claims := middleware.GetClaims(c)
// 	input.ByUserID = claims.UserID
// 	input.ByUserName = claims.Name

// 	report, err := h.reportSvc.Classify(uint(id), input)
// 	if err != nil {
// 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
// 	}

// 	return c.JSON(fiber.Map{"message": "Classification confirmed", "report": report})
// }

// func (h *ReportHandler) Investigate(c *fiber.Ctx) error {
// 	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
// 	if err != nil {
// 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
// 	}

// 	var input service.InvestigateInput
// 	if err := c.BodyParser(&input); err != nil {
// 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
// 	}

// 	claims := middleware.GetClaims(c)
// 	input.ByUserID = claims.UserID
// 	input.ByUserName = claims.Name

// 	report, err := h.reportSvc.Investigate(uint(id), input)
// 	if err != nil {
// 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
// 	}

// 	return c.JSON(fiber.Map{"message": "Investigation submitted", "report": report})
// }

// func (h *ReportHandler) Approve(c *fiber.Ctx) error {
// 	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
// 	if err != nil {
// 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
// 	}

// 	var body struct {
// 		Comment string `json:"comment"`
// 	}
// 	c.BodyParser(&body)

// 	claims := middleware.GetClaims(c)
// 	report, err := h.reportSvc.Approve(uint(id), body.Comment, claims.UserID, claims.Name)
// 	if err != nil {
// 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
// 	}

// 	return c.JSON(fiber.Map{"message": "Report approved and closed", "report": report})
// }

// func (h *ReportHandler) Reject(c *fiber.Ctx) error {
// 	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
// 	if err != nil {
// 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
// 	}

// 	var body struct {
// 		Comment string `json:"comment"`
// 	}
// 	if err := c.BodyParser(&body); err != nil {
// 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
// 	}
// 	if body.Comment == "" {
// 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Rejection reason (comment) is required"})
// 	}

// 	claims := middleware.GetClaims(c)
// 	report, err := h.reportSvc.Reject(uint(id), body.Comment, claims.UserID, claims.Name)
// 	if err != nil {
// 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
// 	}

// 	return c.JSON(fiber.Map{"message": "Report rejected, returned to investigation", "report": report})
// }

// func (h *ReportHandler) GetFiles(c *fiber.Ctx) error {
// 	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
// 	if err != nil {
// 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
// 	}

// 	files, err := h.reportRepo.FindFiles(uint(id))
// 	if err != nil {
// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
// 	}

// 	return c.JSON(files)
// }

// func (h *ReportHandler) UploadFile(c *fiber.Ctx) error {
// 	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
// 	if err != nil {
// 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
// 	}

// 	if _, err := h.reportRepo.FindByID(uint(id)); err != nil {
// 		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Report not found"})
// 	}

// 	fileHeader, err := c.FormFile("file")
// 	if err != nil {
// 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "File is required"})
// 	}

// 	if err := service.ValidateFileSize(fileHeader, config.Cfg.Storage.MaxFileSizeMB); err != nil {
// 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
// 	}

// 	savePath, err := service.SaveUploadedFile(fileHeader, uint(id))
// 	if err != nil {
// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to save file"})
// 	}

// 	claims := middleware.GetClaims(c)
// 	var uploaderID *uint
// 	if claims != nil {
// 		uid := claims.UserID
// 		uploaderID = &uid
// 	}

// 	reportFile := &model.ReportFile{
// 		ReportID:   uint(id),
// 		FileName:   fileHeader.Filename,
// 		FilePath:   savePath,
// 		FileSize:   fileHeader.Size,
// 		MimeType:   fileHeader.Header.Get("Content-Type"),
// 		UploadedBy: uploaderID,
// 	}

// 	if err := h.reportRepo.AddFile(reportFile); err != nil {
// 		service.CleanupFiles([]string{savePath})
// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to save file record"})
// 	}

// 	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
// 		"message": "File uploaded successfully",
// 		"file":    reportFile,
// 	})
// }

// func (h *ReportHandler) Dashboard(c *fiber.Ctx) error {
// 	filters := map[string]interface{}{
// 		"site":      c.Query("site"),
// 		"date_from": c.Query("date_from"),
// 		"date_to":   c.Query("date_to"),
// 	}

// 	stats, err := h.reportRepo.DashboardStats(filters)
// 	if err != nil {
// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
// 	}

// 	return c.JSON(stats)
// }

package handler

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gofiber/fiber/v2"

	"sms-backend/config"
	"sms-backend/middleware"
	"sms-backend/model"
	"sms-backend/repository"
	"sms-backend/service"
)

type ReportHandler struct {
	reportSvc  *service.ReportService
	reportRepo *repository.ReportRepository
}

func NewReportHandler(reportSvc *service.ReportService, reportRepo *repository.ReportRepository) *ReportHandler {
	return &ReportHandler{reportSvc: reportSvc, reportRepo: reportRepo}
}

func (h *ReportHandler) Declare(c *fiber.Ctx) error {
	input := service.DeclareInput{
		Description:  c.FormValue("description"),
		RiskLevel:    model.RiskLevel(c.FormValue("risk_level")),
		ReporterName: c.FormValue("reporter_name"),
	}

	if locID := c.FormValue("location_id"); locID != "" {
		id, err := strconv.ParseUint(locID, 10, 64)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid location_id"})
		}
		uid := uint(id)
		input.LocationID = &uid
	}

	if deptID := c.FormValue("department_id"); deptID != "" {
		id, err := strconv.ParseUint(deptID, 10, 64)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid department_id"})
		}
		uid := uint(id)
		input.DepartmentID = &uid
	}

	// ambil semua files dari field "files"
	form, err := c.MultipartForm()
	if err == nil && form != nil && len(form.File["files"]) > 0 {
		maxFiles := 10
		if len(form.File["files"]) > maxFiles {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": fmt.Sprintf("Maximum %d files allowed", maxFiles),
			})
		}
		input.Files = form.File["files"]
	}

	report, err := h.reportSvc.Declare(input)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Declaration successful",
		"code":    report.Code,
		"id":      report.ID,
	})
}

func (h *ReportHandler) GetAll(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "25"))

	filters := map[string]interface{}{
		"status":        c.Query("status"),
		"site":          c.Query("site"),
		"department_id": c.Query("department_id"),
		"risk_level":    c.Query("risk_level"),
		"date_from":     c.Query("date_from"),
		"date_to":       c.Query("date_to"),
		"page":          page,
		"page_size":     pageSize,
	}

	reports, total, err := h.reportRepo.FindAll(filters)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"data":      reports,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
		"pages":     (total + int64(pageSize) - 1) / int64(pageSize),
	})
}

func (h *ReportHandler) GetByID(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
	}

	report, err := h.reportRepo.FindByID(uint(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Report not found"})
	}

	return c.JSON(report)
}

func (h *ReportHandler) Classify(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
	}

	var input service.ClassifyInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	claims := middleware.GetClaims(c)
	input.ByUserID = claims.UserID
	input.ByUserName = claims.Name

	report, err := h.reportSvc.Classify(uint(id), input)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Classification confirmed", "report": report})
}

func (h *ReportHandler) Investigate(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
	}

	var input service.InvestigateInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	claims := middleware.GetClaims(c)
	input.ByUserID = claims.UserID
	input.ByUserName = claims.Name

	report, err := h.reportSvc.Investigate(uint(id), input)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Investigation submitted", "report": report})
}

func (h *ReportHandler) Approve(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
	}

	var body struct {
		Comment string `json:"comment"`
	}
	c.BodyParser(&body)

	claims := middleware.GetClaims(c)
	report, err := h.reportSvc.Approve(uint(id), body.Comment, claims.UserID, claims.Name)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Report approved and closed", "report": report})
}

func (h *ReportHandler) Reject(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
	}

	var body struct {
		Comment string `json:"comment"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}
	if body.Comment == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Rejection reason (comment) is required"})
	}

	claims := middleware.GetClaims(c)
	report, err := h.reportSvc.Reject(uint(id), body.Comment, claims.UserID, claims.Name)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Report rejected, returned to investigation", "report": report})
}

func (h *ReportHandler) GetFiles(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
	}

	files, err := h.reportRepo.FindFiles(uint(id))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(files)
}

func (h *ReportHandler) UploadFile(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
	}

	if _, err := h.reportRepo.FindByID(uint(id)); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Report not found"})
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "File is required"})
	}

	if err := service.ValidateFileSize(fileHeader, config.Cfg.Storage.MaxFileSizeMB); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	savePath, err := service.SaveUploadedFile(fileHeader, uint(id))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to save file"})
	}

	claims := middleware.GetClaims(c)
	var uploaderID *uint
	if claims != nil {
		uid := claims.UserID
		uploaderID = &uid
	}

	reportFile := &model.ReportFile{
		ReportID:   uint(id),
		FileName:   fileHeader.Filename,
		FilePath:   savePath,
		FileSize:   fileHeader.Size,
		MimeType:   fileHeader.Header.Get("Content-Type"),
		UploadedBy: uploaderID,
	}

	if err := h.reportRepo.AddFile(reportFile); err != nil {
		service.CleanupFiles([]string{savePath})
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to save file record"})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "File uploaded successfully",
		"file":    reportFile,
	})
}

func (h *ReportHandler) Dashboard(c *fiber.Ctx) error {
	filters := map[string]interface{}{
		"site":      c.Query("site"),
		"date_from": c.Query("date_from"),
		"date_to":   c.Query("date_to"),
	}

	stats, err := h.reportRepo.DashboardStats(filters)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(stats)
}

func (h *ReportHandler) ServeFile(c *fiber.Ctx) error {
	reportID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid report ID"})
	}
	fileID, err := strconv.ParseUint(c.Params("file_id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid file ID"})
	}

	file, err := h.reportRepo.FindFileByID(uint(reportID), uint(fileID))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "File not found"})
	}

	// cek file ada di disk
	if _, err := os.Stat(file.FilePath); os.IsNotExist(err) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "File not found on disk"})
	}

	mimeType := file.MimeType
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	// convert ke absolute path supaya Fiber bisa serve dengan benar
	absPath, err := filepath.Abs(file.FilePath)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to resolve file path"})
	}

	c.Set("Content-Type", mimeType)
	c.Set("Content-Disposition", fmt.Sprintf(`inline; filename="%s"`, file.FileName))
	c.Set("Cache-Control", "private, max-age=3600")

	return c.SendFile(absPath)
}
