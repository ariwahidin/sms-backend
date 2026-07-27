package handler

import (
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gofiber/fiber/v2"

	"sms-backend/config"
	"sms-backend/email"
	"sms-backend/middleware"
	"sms-backend/model"
	"sms-backend/repository"
	"sms-backend/service"
)

type ReportHandler struct {
	reportSvc    *service.ReportService
	reportRepo   *repository.ReportRepository
	emailService *email.EmailService // baru
}

func NewReportHandler(reportSvc *service.ReportService, reportRepo *repository.ReportRepository, emailService *email.EmailService) *ReportHandler {
	return &ReportHandler{
		reportSvc:    reportSvc,
		reportRepo:   reportRepo,
		emailService: emailService,
	}
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

func (h *ReportHandler) GetByID(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
	}

	report, err := h.reportRepo.FindByID(uint(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Report not found"})
	}

	recipients := h.emailService.GetRecipientDetails(report)

	return c.JSON(fiber.Map{
		"report":     report,
		"recipients": recipients,
	})
}

func (h *ReportHandler) Classify(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
	}

	input := service.ClassifyInput{
		RiskLevel: model.RiskLevel(c.FormValue("risk_level")),
		Priority:  model.Priority(c.FormValue("priority")),
		Notes:     c.FormValue("notes"),
	}

	if picID := c.FormValue("pic_id"); picID != "" {
		pid, err := strconv.ParseUint(picID, 10, 64)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid pic_id"})
		}
		uid := uint(pid)
		input.PICID = &uid
	}

	if form, err := c.MultipartForm(); err == nil && form != nil && len(form.File["files"]) > 0 {
		if len(form.File["files"]) > 10 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Maximum 10 files allowed"})
		}
		input.Files = form.File["files"]
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

	input := service.InvestigateInput{
		RootCause:       c.FormValue("root_cause"),
		Countermeasures: c.FormValue("countermeasures"),
	}

	if form, err := c.MultipartForm(); err == nil && form != nil && len(form.File["files"]) > 0 {
		if len(form.File["files"]) > 10 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Maximum 10 files allowed"})
		}
		input.Files = form.File["files"]
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

	comment := c.FormValue("comment")

	var files []*multipart.FileHeader
	if form, err := c.MultipartForm(); err == nil && form != nil && len(form.File["files"]) > 0 {
		if len(form.File["files"]) > 10 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Maximum 10 files allowed"})
		}
		files = form.File["files"]
	}

	claims := middleware.GetClaims(c)
	report, err := h.reportSvc.Approve(uint(id), comment, files, claims.UserID, claims.Name)
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

	comment := c.FormValue("comment")
	if comment == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Rejection reason (comment) is required"})
	}

	var files []*multipart.FileHeader
	if form, err := c.MultipartForm(); err == nil && form != nil && len(form.File["files"]) > 0 {
		if len(form.File["files"]) > 10 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Maximum 10 files allowed"})
		}
		files = form.File["files"]
	}

	claims := middleware.GetClaims(c)
	report, err := h.reportSvc.Reject(uint(id), comment, files, claims.UserID, claims.Name)
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

	report, err := h.reportRepo.FindByID(uint(id))
	if err != nil {
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
	uploaderName := ""
	if claims != nil {
		uid := claims.UserID
		uploaderID = &uid
		uploaderName = claims.Name
	}

	// buat activity dulu, supaya dapat ID-nya untuk dikaitkan ke file
	activity := &model.ReportActivity{
		ReportID:   uint(id),
		Action:     "upload_file",
		FromStatus: report.Status,
		ToStatus:   report.Status,
		Notes:      fmt.Sprintf("Uploaded file: %s", fileHeader.Filename),
		ByUserID:   uploaderID,
		ByUserName: uploaderName,
	}
	if err := h.reportRepo.AddActivity(activity); err != nil {
		service.CleanupFiles([]string{savePath})
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to log activity"})
	}

	reportFile := &model.ReportFile{
		ReportID:   uint(id),
		ActivityID: &activity.ID,
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

	// muat ulang relasi Activity supaya response ke frontend lengkap
	reportFile.Activity = activity

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
