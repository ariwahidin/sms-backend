package handler

import (
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"strconv"
	"strings"

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
	userRepo     *repository.UserRepository
	emailService *email.EmailService
}

func NewReportHandler(
	reportSvc *service.ReportService,
	reportRepo *repository.ReportRepository,
	userRepo *repository.UserRepository,
	emailService *email.EmailService,
) *ReportHandler {
	return &ReportHandler{
		reportSvc:    reportSvc,
		reportRepo:   reportRepo,
		userRepo:     userRepo,
		emailService: emailService,
	}
}

func (h *ReportHandler) Declare(c *fiber.Ctx) error {

	safetyCategoryID, err := strconv.ParseUint(c.FormValue("safety_category_id"), 10, 32)
	if err != nil || safetyCategoryID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Safety category is required",
		})
	}

	input := service.DeclareInput{
		SafetyCategoryID: uint(safetyCategoryID),
		Description:      c.FormValue("description"),
		RiskLevel:        model.RiskLevel(c.FormValue("risk_level")),
		ReporterName:     c.FormValue("reporter_name"),
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

	if page < 1 {
		page = 1
	}

	if pageSize < 1 {
		pageSize = 25
	}

	if pageSize > 100 {
		pageSize = 100
	}

	// ============================================================
	// GET CURRENT USER
	// ============================================================

	claims := middleware.GetClaims(c)

	if claims == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	// ============================================================
	// BASE FILTERS
	// ============================================================

	filters := map[string]interface{}{
		"status":     c.Query("status"),
		"site":       c.Query("site"),
		"risk_level": c.Query("risk_level"),
		"date_from":  c.Query("date_from"),
		"date_to":    c.Query("date_to"),
		"page":       page,
		"page_size":  pageSize,
	}

	// ============================================================
	// DEPARTMENT ACCESS CONTROL
	// ============================================================

	if claims.Role == model.RoleAdmin || claims.Role == model.RolePIC {

		// Admin & PIC boleh melihat semua department.
		//
		// Kalau frontend mengirim department_id,
		// maka digunakan sebagai filter.
		filters["department_id"] = c.Query("department_id")

	} else {

		// User selain Admin/PIC hanya boleh melihat
		// report dari department miliknya sendiri.

		user, err := h.userRepo.FindByID(claims.UserID)

		if err != nil {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "User not found or access denied",
			})
		}

		if user.DepartmentID == nil {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "User has no assigned department",
			})
		}

		// Jangan gunakan department_id dari query parameter.
		filters["department_id"] = *user.DepartmentID
	}

	// ============================================================
	// GET REPORTS
	// ============================================================

	reports, total, err := h.reportRepo.FindAll(filters)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// ============================================================
	// RESPONSE
	// ============================================================

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

func (h *ReportHandler) RequestDetectionEdit(c *fiber.Ctx) error {

	// =========================================================
	// PARSE REPORT ID
	// =========================================================

	reportID, err := strconv.ParseUint(
		c.Params("id"),
		10,
		64,
	)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			fiber.Map{
				"error": "Invalid report ID",
			},
		)
	}

	// =========================================================
	// REQUEST PAYLOAD
	// =========================================================

	type RequestDetectionEditPayload struct {
		SafetyCategoryID *uint  `json:"safety_category_id"`
		LocationID       *uint  `json:"location_id"`
		DepartmentID     *uint  `json:"department_id"`
		RiskLevel        string `json:"risk_level"`
		Description      string `json:"description"`
		Reason           string `json:"reason"`
	}

	var payload RequestDetectionEditPayload

	// =========================================================
	// PARSE JSON BODY
	// =========================================================

	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			fiber.Map{
				"error": "Invalid request body",
			},
		)
	}

	// =========================================================
	// VALIDATION
	// =========================================================

	if strings.TrimSpace(payload.Description) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(
			fiber.Map{
				"error": "description is required",
			},
		)
	}

	if strings.TrimSpace(payload.Reason) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(
			fiber.Map{
				"error": "reason is required",
			},
		)
	}

	// =========================================================
	// BUILD SERVICE INPUT
	// =========================================================

	input := service.RequestDetectionEditInput{
		Description:      payload.Description,
		Reason:           payload.Reason,
		RiskLevel:        model.RiskLevel(payload.RiskLevel),
		LocationID:       payload.LocationID,
		DepartmentID:     payload.DepartmentID,
		SafetyCategoryID: payload.SafetyCategoryID,
	}

	// =========================================================
	// USER CLAIMS
	// =========================================================

	claims := middleware.GetClaims(c)

	input.ByUserID = claims.UserID
	input.ByUserName = claims.Name

	// =========================================================
	// REQUEST EDIT
	// =========================================================

	request, err := h.reportSvc.RequestDetectionEdit(
		uint(reportID),
		input,
	)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			fiber.Map{
				"error": err.Error(),
			},
		)
	}

	// =========================================================
	// RESPONSE
	// =========================================================

	return c.Status(fiber.StatusCreated).JSON(
		fiber.Map{
			"message": "Detection edit request submitted",
			"request": request,
		},
	)
}

func (h *ReportHandler) ApproveDetectionEdit(c *fiber.Ctx) error {

	reportID, err := strconv.ParseUint(
		c.Params("id"),
		10,
		64,
	)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			fiber.Map{"error": "Invalid report ID"},
		)
	}

	requestID, err := strconv.ParseUint(
		c.Params("request_id"),
		10,
		64,
	)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			fiber.Map{"error": "Invalid request ID"},
		)
	}

	claims := middleware.GetClaims(c)

	report, err := h.reportSvc.ApproveDetectionEdit(
		uint(reportID),
		uint(requestID),
		claims.UserID,
		claims.Name,
	)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			fiber.Map{"error": err.Error()},
		)
	}

	return c.JSON(
		fiber.Map{
			"message": "Detection edit approved",
			"report":  report,
		},
	)
}

func (h *ReportHandler) RejectDetectionEdit(c *fiber.Ctx) error {
	// ============================================================
	// PARSE REPORT ID
	// ============================================================

	reportID, err := strconv.ParseUint(
		c.Params("id"),
		10,
		64,
	)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			fiber.Map{
				"error": "Invalid report ID",
			},
		)
	}

	// ============================================================
	// PARSE REQUEST ID
	// ============================================================

	requestID, err := strconv.ParseUint(
		c.Params("request_id"),
		10,
		64,
	)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			fiber.Map{
				"error": "Invalid request ID",
			},
		)
	}

	// ============================================================
	// REQUEST PAYLOAD
	// ============================================================

	type RejectDetectionEditPayload struct {
		Note string `json:"note"`
	}

	var payload RejectDetectionEditPayload

	// ============================================================
	// PARSE JSON BODY
	// ============================================================

	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			fiber.Map{
				"error": "Invalid request body",
			},
		)
	}

	// ============================================================
	// VALIDATION
	// ============================================================

	payload.Note = strings.TrimSpace(payload.Note)

	if payload.Note == "" {
		return c.Status(fiber.StatusBadRequest).JSON(
			fiber.Map{
				"error": "Rejection reason is required",
			},
		)
	}

	// ============================================================
	// GET USER CLAIMS
	// ============================================================

	claims := middleware.GetClaims(c)

	if claims == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(
			fiber.Map{
				"error": "Unauthorized",
			},
		)
	}

	// ============================================================
	// REJECT DETECTION EDIT
	// ============================================================

	err = h.reportSvc.RejectDetectionEdit(
		uint(reportID),
		uint(requestID),
		payload.Note,
		claims.UserID,
		claims.Name,
	)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			fiber.Map{
				"error": err.Error(),
			},
		)
	}

	// ============================================================
	// RESPONSE
	// ============================================================

	return c.JSON(
		fiber.Map{
			"message": "Detection edit rejected",
		},
	)
}

func (h *ReportHandler) Delete(c *fiber.Ctx) error {
	// Validate report ID
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil || id == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid report ID",
		})
	}

	// Get authenticated user
	claims := middleware.GetClaims(c)
	if claims == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	// Only Admin can delete reports
	if claims.Role != model.RoleAdmin {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Only administrators can delete reports",
		})
	}

	// Execute deletion through service
	if err := h.reportSvc.Delete(uint(id), claims.UserID, claims.Name); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Report deleted successfully",
		"id":      uint(id),
	})
}
