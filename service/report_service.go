package service

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"sms-backend/config"
	"sms-backend/database"
	"sms-backend/email"
	"sms-backend/model"
	"sms-backend/repository"

	"gorm.io/gorm"
)

type ReportService struct {
	reportRepo         *repository.ReportRepository
	safetyCategoryRepo *repository.SafetyCategoryRepository
	emailSvc           *email.EmailService
}

// func NewReportService(reportRepo *repository.ReportRepository, emailSvc *email.EmailService) *ReportService {
// 	return &ReportService{reportRepo: reportRepo, emailSvc: emailSvc}
// }

func NewReportService(reportRepo *repository.ReportRepository, safetyCategoryRepo *repository.SafetyCategoryRepository, emailSvc *email.EmailService) *ReportService {
	return &ReportService{reportRepo: reportRepo, safetyCategoryRepo: safetyCategoryRepo, emailSvc: emailSvc}
}

// type DeclareInput struct {
// 	LocationID   *uint                   `json:"location_id"`
// 	DepartmentID *uint                   `json:"department_id"`
// 	Description  string                  `json:"description"`
// 	RiskLevel    model.RiskLevel         `json:"risk_level"`
// 	ReporterName string                  `json:"reporter_name"`
// 	Files        []*multipart.FileHeader // diisi dari handler, bukan JSON
// }

type DeclareInput struct {
	SafetyCategoryID uint
	LocationID       *uint
	DepartmentID     *uint
	Description      string
	RiskLevel        model.RiskLevel
	ReporterName     string
	Files            []*multipart.FileHeader
}

// func (s *ReportService) Declare(input DeclareInput) (*model.RiskReport, error) {
// 	if input.Description == "" {
// 		return nil, errors.New("description is required")
// 	}
// 	if input.RiskLevel == "" {
// 		input.RiskLevel = model.RiskLow
// 	}

// 	// validasi ukuran semua file sebelum mulai transaksi
// 	for _, fh := range input.Files {
// 		if err := ValidateFileSize(fh, config.Cfg.Storage.MaxFileSizeMB); err != nil {
// 			return nil, err
// 		}
// 	}

// 	code, err := s.reportRepo.GenerateCode()
// 	if err != nil {
// 		return nil, err
// 	}

// 	report := &model.RiskReport{
// 		Code:            code,
// 		LocationID:      input.LocationID,
// 		DepartmentID:    input.DepartmentID,
// 		Description:     input.Description,
// 		RiskLevelOrigin: input.RiskLevel,
// 		RiskLevel:       input.RiskLevel,
// 		Status:          model.StatusDeclaration,
// 		ReporterName:    input.ReporterName,
// 	}

// 	tx := database.DB.Begin()
// 	if tx.Error != nil {
// 		return nil, tx.Error
// 	}

// 	if err := tx.Create(report).Error; err != nil {
// 		tx.Rollback()
// 		return nil, err
// 	}

// 	// buat activity dulu, supaya dapat ID-nya untuk dikaitkan ke file
// 	activity := &model.ReportActivity{
// 		ReportID: report.ID,
// 		Action:   "declared",
// 		ToStatus: model.StatusDeclaration,
// 		Notes:    input.Description,
// 	}
// 	if err := tx.Create(activity).Error; err != nil {
// 		tx.Rollback()
// 		return nil, err
// 	}

// 	// upload files jika ada, di-link ke activity "declared" di atas
// 	var savedPaths []string
// 	for _, fh := range input.Files {
// 		savePath, err := SaveUploadedFile(fh, report.ID)
// 		if err != nil {
// 			tx.Rollback()
// 			CleanupFiles(savedPaths)
// 			return nil, fmt.Errorf("failed to save file %s: %w", fh.Filename, err)
// 		}
// 		savedPaths = append(savedPaths, savePath)

// 		rf := &model.ReportFile{
// 			ReportID:   report.ID,
// 			ActivityID: &activity.ID,
// 			FileName:   fh.Filename,
// 			FilePath:   savePath,
// 			FileSize:   fh.Size,
// 			MimeType:   fh.Header.Get("Content-Type"),
// 		}
// 		if err := tx.Create(rf).Error; err != nil {
// 			tx.Rollback()
// 			CleanupFiles(savedPaths)
// 			return nil, err
// 		}
// 	}

// 	if err := tx.Commit().Error; err != nil {
// 		CleanupFiles(savedPaths)
// 		return nil, err
// 	}

// 	go s.emailSvc.SendDeclarationNotice(report)

// 	return report, nil
// }

func (s *ReportService) Declare(input DeclareInput) (*model.RiskReport, error) {
	input.Description = strings.TrimSpace(input.Description)

	if input.Description == "" {
		return nil, errors.New("description is required")
	}

	if input.SafetyCategoryID == 0 {
		return nil, errors.New("safety category is required")
	}

	if input.RiskLevel == "" {
		input.RiskLevel = model.RiskLow
	}

	// Validate selected safety category.
	category, err := s.safetyCategoryRepo.FindActiveByID(input.SafetyCategoryID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("selected safety category is invalid or inactive")
		}
		return nil, err
	}

	// Validate all file sizes before starting the transaction.
	for _, fh := range input.Files {
		if err := ValidateFileSize(fh, config.Cfg.Storage.MaxFileSizeMB); err != nil {
			return nil, err
		}
	}

	code, err := s.reportRepo.GenerateCode()
	if err != nil {
		return nil, err
	}

	report := &model.RiskReport{
		Code:             code,
		SafetyCategoryID: &category.ID,
		LocationID:       input.LocationID,
		DepartmentID:     input.DepartmentID,
		Description:      input.Description,
		RiskLevelOrigin:  input.RiskLevel,
		RiskLevel:        input.RiskLevel,
		Status:           model.StatusDeclaration,
		ReporterName:     input.ReporterName,
	}

	tx := database.DB.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	if err := tx.Create(report).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	activity := &model.ReportActivity{
		ReportID: report.ID,
		Action:   "declared",
		ToStatus: model.StatusDeclaration,
		Notes:    input.Description,
	}

	if err := tx.Create(activity).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	var savedPaths []string

	for _, fh := range input.Files {
		savePath, err := SaveUploadedFile(fh, report.ID)
		if err != nil {
			tx.Rollback()
			CleanupFiles(savedPaths)
			return nil, fmt.Errorf("failed to save file %s: %w", fh.Filename, err)
		}

		savedPaths = append(savedPaths, savePath)

		rf := &model.ReportFile{
			ReportID:   report.ID,
			ActivityID: &activity.ID,
			FileName:   fh.Filename,
			FilePath:   savePath,
			FileSize:   fh.Size,
			MimeType:   fh.Header.Get("Content-Type"),
		}

		if err := tx.Create(rf).Error; err != nil {
			tx.Rollback()
			CleanupFiles(savedPaths)
			return nil, err
		}
	}

	if err := tx.Commit().Error; err != nil {
		CleanupFiles(savedPaths)
		return nil, err
	}

	// Load relation for email notification.
	report.SafetyCategory = category

	go s.emailSvc.SendDeclarationNotice(report)

	return report, nil
}

type ClassifyInput struct {
	LocationID   *uint
	DepartmentID *uint
	RiskLevel    model.RiskLevel
	Priority     model.Priority
	PICID        *uint
	Notes        string
	Files        []*multipart.FileHeader
	ByUserID     uint
	ByUserName   string
}

func (s *ReportService) Classify(reportID uint, input ClassifyInput) (*model.RiskReport, error) {
	report, err := s.reportRepo.FindByID(reportID)
	if err != nil {
		return nil, errors.New("report not found")
	}
	if report.Status != model.StatusDeclaration {
		return nil, errors.New("report is not in declaration status")
	}
	for _, fh := range input.Files {
		if err := ValidateFileSize(fh, config.Cfg.Storage.MaxFileSizeMB); err != nil {
			return nil, err
		}
	}

	now := time.Now()
	sla := model.SLAHours[input.Priority]

	report.Status = model.StatusClassification
	report.Priority = input.Priority
	report.RiskLevel = input.RiskLevel
	report.PICID = input.PICID
	report.PICNotes = input.Notes
	report.ClassifiedAt = &now
	report.ClassifiedBy = &input.ByUserID
	if input.LocationID != nil {
		report.LocationID = input.LocationID
	}
	if input.DepartmentID != nil {
		report.DepartmentID = input.DepartmentID
	}
	investigationDeadline := now.Add(time.Duration(sla["investigation"]) * time.Hour)
	approvalDeadline := now.Add(time.Duration(sla["approval"]) * time.Hour)
	report.SLADeadlineInvestigation = &investigationDeadline
	report.SLADeadlineApproval = &approvalDeadline

	tx := database.DB.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	if err := tx.Save(report).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	activity := &model.ReportActivity{
		ReportID:   report.ID,
		Action:     "classified",
		FromStatus: model.StatusDeclaration,
		ToStatus:   model.StatusClassification,
		Notes:      input.Notes,
		ByUserID:   &input.ByUserID,
		ByUserName: input.ByUserName,
	}
	if err := tx.Create(activity).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	var savedPaths []string
	for _, fh := range input.Files {
		savePath, err := SaveUploadedFile(fh, report.ID)
		if err != nil {
			tx.Rollback()
			CleanupFiles(savedPaths)
			return nil, fmt.Errorf("failed to save file %s: %w", fh.Filename, err)
		}
		savedPaths = append(savedPaths, savePath)
		rf := &model.ReportFile{
			ReportID:   report.ID,
			ActivityID: &activity.ID,
			FileName:   fh.Filename,
			FilePath:   savePath,
			FileSize:   fh.Size,
			MimeType:   fh.Header.Get("Content-Type"),
			UploadedBy: &input.ByUserID,
		}
		if err := tx.Create(rf).Error; err != nil {
			tx.Rollback()
			CleanupFiles(savedPaths)
			return nil, err
		}
	}

	if err := tx.Commit().Error; err != nil {
		CleanupFiles(savedPaths)
		return nil, err
	}

	go s.emailSvc.SendStatusUpdateNotice(report, "Classification confirmed")
	return report, nil
}

type InvestigateInput struct {
	RootCause       string
	Countermeasures string
	Files           []*multipart.FileHeader
	ByUserID        uint
	ByUserName      string
}

func (s *ReportService) Investigate(reportID uint, input InvestigateInput) (*model.RiskReport, error) {
	report, err := s.reportRepo.FindByID(reportID)
	if err != nil {
		return nil, errors.New("report not found")
	}
	if report.Status != model.StatusClassification && report.Status != model.StatusInvestigation && report.Status != model.StatusApproval {
		return nil, errors.New("report cannot be investigated at current status")
	}
	if input.RootCause == "" || input.Countermeasures == "" {
		return nil, errors.New("root cause and countermeasures are required")
	}
	for _, fh := range input.Files {
		if err := ValidateFileSize(fh, config.Cfg.Storage.MaxFileSizeMB); err != nil {
			return nil, err
		}
	}

	now := time.Now()
	fromStatus := report.Status
	report.Status = model.StatusInvestigation
	report.RootCause = input.RootCause
	report.Countermeasures = input.Countermeasures
	report.InvestigatedAt = &now
	report.InvestigatedBy = &input.ByUserID

	tx := database.DB.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	if err := tx.Save(report).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	activity := &model.ReportActivity{
		ReportID:   report.ID,
		Action:     "investigated",
		FromStatus: fromStatus,
		ToStatus:   model.StatusInvestigation,
		Notes:      "Cause : " + input.RootCause + ", `\nCountermeasures : " + input.Countermeasures,
		ByUserID:   &input.ByUserID,
		ByUserName: input.ByUserName,
	}
	if err := tx.Create(activity).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	var savedPaths []string
	for _, fh := range input.Files {
		savePath, err := SaveUploadedFile(fh, report.ID)
		if err != nil {
			tx.Rollback()
			CleanupFiles(savedPaths)
			return nil, fmt.Errorf("failed to save file %s: %w", fh.Filename, err)
		}
		savedPaths = append(savedPaths, savePath)
		rf := &model.ReportFile{
			ReportID:   report.ID,
			ActivityID: &activity.ID,
			FileName:   fh.Filename,
			FilePath:   savePath,
			FileSize:   fh.Size,
			MimeType:   fh.Header.Get("Content-Type"),
			UploadedBy: &input.ByUserID,
		}
		if err := tx.Create(rf).Error; err != nil {
			tx.Rollback()
			CleanupFiles(savedPaths)
			return nil, err
		}
	}

	if err := tx.Commit().Error; err != nil {
		CleanupFiles(savedPaths)
		return nil, err
	}

	go s.emailSvc.SendReviewNotice(report)
	return report, nil
}

func (s *ReportService) Approve(reportID uint, comment string, files []*multipart.FileHeader, byUserID uint, byUserName string) (*model.RiskReport, error) {
	report, err := s.reportRepo.FindByID(reportID)
	if err != nil {
		return nil, errors.New("report not found")
	}
	if report.Status != model.StatusInvestigation {
		return nil, errors.New("report is not ready for approval")
	}
	for _, fh := range files {
		if err := ValidateFileSize(fh, config.Cfg.Storage.MaxFileSizeMB); err != nil {
			return nil, err
		}
	}

	now := time.Now()
	report.Status = model.StatusClosed
	report.HODComment = comment
	report.ApprovedAt = &now
	report.ApprovedBy = &byUserID

	tx := database.DB.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	if err := tx.Save(report).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	activity := &model.ReportActivity{
		ReportID:   report.ID,
		Action:     "approved",
		FromStatus: model.StatusInvestigation,
		ToStatus:   model.StatusClosed,
		Notes:      comment,
		ByUserID:   &byUserID,
		ByUserName: byUserName,
	}
	if err := tx.Create(activity).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	var savedPaths []string
	for _, fh := range files {
		savePath, err := SaveUploadedFile(fh, report.ID)
		if err != nil {
			tx.Rollback()
			CleanupFiles(savedPaths)
			return nil, fmt.Errorf("failed to save file %s: %w", fh.Filename, err)
		}
		savedPaths = append(savedPaths, savePath)
		rf := &model.ReportFile{
			ReportID:   report.ID,
			ActivityID: &activity.ID,
			FileName:   fh.Filename,
			FilePath:   savePath,
			FileSize:   fh.Size,
			MimeType:   fh.Header.Get("Content-Type"),
			UploadedBy: &byUserID,
		}
		if err := tx.Create(rf).Error; err != nil {
			tx.Rollback()
			CleanupFiles(savedPaths)
			return nil, err
		}
	}

	if err := tx.Commit().Error; err != nil {
		CleanupFiles(savedPaths)
		return nil, err
	}

	go s.emailSvc.SendClosingNotice(report)
	return report, nil
}

func (s *ReportService) Reject(reportID uint, comment string, files []*multipart.FileHeader, byUserID uint, byUserName string) (*model.RiskReport, error) {
	report, err := s.reportRepo.FindByID(reportID)
	if err != nil {
		return nil, errors.New("report not found")
	}
	if report.Status != model.StatusInvestigation {
		return nil, errors.New("report is not ready for review")
	}
	for _, fh := range files {
		if err := ValidateFileSize(fh, config.Cfg.Storage.MaxFileSizeMB); err != nil {
			return nil, err
		}
	}

	reportStatus := model.StatusClassification

	if report.Status == model.StatusDeclaration {
		reportStatus = model.StatusClassification
	}

	if report.Status == model.StatusInvestigation {
		reportStatus = model.StatusApproval
	}

	now := time.Now()
	report.Status = reportStatus
	report.HODComment = comment
	report.RejectedAt = &now
	report.RejectedBy = &byUserID

	tx := database.DB.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	if err := tx.Save(report).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	activity := &model.ReportActivity{
		ReportID:   report.ID,
		Action:     "rejected",
		FromStatus: model.StatusInvestigation,
		ToStatus:   reportStatus,
		Notes:      comment,
		ByUserID:   &byUserID,
		ByUserName: byUserName,
	}
	if err := tx.Create(activity).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	var savedPaths []string
	for _, fh := range files {
		savePath, err := SaveUploadedFile(fh, report.ID)
		if err != nil {
			tx.Rollback()
			CleanupFiles(savedPaths)
			return nil, fmt.Errorf("failed to save file %s: %w", fh.Filename, err)
		}
		savedPaths = append(savedPaths, savePath)
		rf := &model.ReportFile{
			ReportID:   report.ID,
			ActivityID: &activity.ID,
			FileName:   fh.Filename,
			FilePath:   savePath,
			FileSize:   fh.Size,
			MimeType:   fh.Header.Get("Content-Type"),
			UploadedBy: &byUserID,
		}
		if err := tx.Create(rf).Error; err != nil {
			tx.Rollback()
			CleanupFiles(savedPaths)
			return nil, err
		}
	}

	if err := tx.Commit().Error; err != nil {
		CleanupFiles(savedPaths)
		return nil, err
	}

	go s.emailSvc.SendRejectionNotice(report)
	return report, nil
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func ValidateFileSize(fh *multipart.FileHeader, maxMB int64) error {
	if fh.Size > maxMB*1024*1024 {
		return fmt.Errorf("file '%s' melebihi ukuran maksimal %dMB", fh.Filename, maxMB)
	}
	return nil
}

func SaveUploadedFile(fh *multipart.FileHeader, reportID uint) (string, error) {
	uploadDir := filepath.Join(config.Cfg.Storage.UploadDir, fmt.Sprintf("report_%d", reportID))
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return "", err
	}

	ext := filepath.Ext(fh.Filename)
	uniqueName := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), sanitizeFileName(fh.Filename), ext)
	savePath := filepath.Join(uploadDir, uniqueName)

	src, err := fh.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	dst, err := os.Create(savePath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return "", err
	}
	return savePath, nil
}

func CleanupFiles(paths []string) {
	for _, p := range paths {
		os.Remove(p)
	}
}

func sanitizeFileName(name string) string {
	base := strings.TrimSuffix(name, filepath.Ext(name))
	safe := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			return r
		}
		return '_'
	}, base)
	if len(safe) > 50 {
		safe = safe[:50]
	}
	return safe
}

type RequestDetectionEditInput struct {
	SafetyCategoryID *uint
	LocationID       *uint
	DepartmentID     *uint
	RiskLevel        model.RiskLevel
	Description      string
	Reason           string

	ByUserID   uint
	ByUserName string
}

func (s *ReportService) RequestDetectionEdit(
	reportID uint,
	input RequestDetectionEditInput,
) (*model.ReportEditRequest, error) {

	report, err := s.reportRepo.FindByID(reportID)
	if err != nil {
		return nil, errors.New("report not found")
	}

	// Jangan boleh membuat request baru
	// kalau masih ada pending request.
	existing, err := s.reportRepo.FindPendingEditRequest(reportID)

	if err == nil && existing != nil {
		return nil, errors.New(
			"this report already has a pending detection edit request",
		)
	}

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// Validation
	if strings.TrimSpace(input.Description) == "" {
		return nil, errors.New("description is required")
	}

	if strings.TrimSpace(input.Reason) == "" {
		return nil, errors.New("reason is required")
	}

	switch input.RiskLevel {
	case model.RiskLow,
		model.RiskMedium,
		model.RiskHigh:
	default:
		return nil, errors.New("invalid risk level")
	}

	request := &model.ReportEditRequest{
		ReportID: report.ID,

		SafetyCategoryID: input.SafetyCategoryID,
		LocationID:       input.LocationID,
		DepartmentID:     input.DepartmentID,

		RiskLevel:   input.RiskLevel,
		Description: input.Description,

		Reason: input.Reason,

		RequestedBy:   &input.ByUserID,
		RequestedName: input.ByUserName,

		Status: model.EditRequestPending,
	}

	if err := s.reportRepo.CreateEditRequest(request); err != nil {
		return nil, err
	}

	// Audit
	activity := &model.ReportActivity{
		ReportID: report.ID,

		Action:     "request_detection_edit",
		FromStatus: report.Status,
		ToStatus:   report.Status,

		Notes: fmt.Sprintf(
			"Detection edit requested. Reason: %s",
			input.Reason,
		),

		ByUserID:   &input.ByUserID,
		ByUserName: input.ByUserName,
	}

	if err := s.reportRepo.AddActivity(activity); err != nil {
		return nil, err
	}

	return request, nil
}

func (s *ReportService) ApproveDetectionEdit(
	reportID uint,
	requestID uint,
	userID uint,
	userName string,
) (*model.RiskReport, error) {

	tx := database.DB.Begin()

	if tx.Error != nil {
		return nil, tx.Error
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var request model.ReportEditRequest

	if err := tx.
		Preload("Location").
		Preload("Department").
		Where(
			"id = ? AND report_id = ?",
			requestID,
			reportID,
		).
		First(&request).Error; err != nil {

		tx.Rollback()

		return nil, errors.New("edit request not found")
	}

	if request.Status != model.EditRequestPending {
		tx.Rollback()

		return nil, errors.New(
			"edit request has already been processed",
		)
	}

	var report model.RiskReport

	if err := tx.First(&report, reportID).Error; err != nil {
		tx.Rollback()

		return nil, errors.New("report not found")
	}

	// =========================================================
	// APPLY REQUESTED DETECTION
	// =========================================================

	report.LocationID = request.LocationID
	report.SafetyCategoryID = request.SafetyCategoryID
	report.DepartmentID = request.DepartmentID
	report.RiskLevel = request.RiskLevel
	report.Description = request.Description

	if err := tx.Save(&report).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// =========================================================
	// APPROVE REQUEST
	// =========================================================

	now := time.Now()

	request.Status = model.EditRequestApproved
	request.ApprovedBy = &userID
	request.ApprovedName = userName
	request.ApprovedAt = &now

	if err := tx.Save(&request).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// =========================================================
	// AUDIT
	// =========================================================

	activity := &model.ReportActivity{
		ReportID: report.ID,

		Action:     "approve_detection_edit",
		FromStatus: report.Status,
		ToStatus:   report.Status,

		Notes: fmt.Sprintf(
			"Detection edit approved. Request ID: %d",
			request.ID,
		),

		ByUserID:   &userID,
		ByUserName: userName,
	}

	if err := tx.Create(&activity).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return &report, nil
}

func (s *ReportService) RejectDetectionEdit(
	reportID uint,
	requestID uint,
	note string,
	userID uint,
	userName string,
) error {

	if strings.TrimSpace(note) == "" {
		return errors.New("rejection reason is required")
	}

	request, err := s.reportRepo.FindEditRequestByID(
		reportID,
		requestID,
	)

	if err != nil {
		return errors.New("edit request not found")
	}

	if request.Status != model.EditRequestPending {
		return errors.New(
			"edit request has already been processed",
		)
	}

	now := time.Now()

	request.Status = model.EditRequestRejected

	request.RejectedBy = &userID
	request.RejectedName = userName
	request.RejectedAt = &now

	request.ApprovalNote = note

	if err := s.reportRepo.UpdateEditRequest(request); err != nil {
		return err
	}

	report, err := s.reportRepo.FindByID(reportID)
	if err != nil {
		return err
	}

	activity := &model.ReportActivity{
		ReportID: report.ID,

		Action:     "reject_detection_edit",
		FromStatus: report.Status,
		ToStatus:   report.Status,

		Notes: fmt.Sprintf(
			"Detection edit rejected. Request ID: %d. Reason: %s",
			request.ID,
			note,
		),

		ByUserID:   &userID,
		ByUserName: userName,
	}

	return s.reportRepo.AddActivity(activity)
}

func (s *ReportService) Delete(
	reportID uint,
	userID uint,
	userName string,
) error {

	if reportID == 0 {
		return errors.New("invalid report ID")
	}

	// Check whether report exists.
	report, err := s.reportRepo.FindByID(reportID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("report not found")
		}

		return err
	}

	// Prepare deletion audit activity.
	activity := &model.ReportActivity{
		ReportID:   report.ID,
		Action:     "deleted",
		FromStatus: report.Status,
		ToStatus:   report.Status,
		Notes:      fmt.Sprintf("Report %s deleted by administrator", report.Code),
		ByUserID:   &userID,
		ByUserName: userName,
	}

	// Delete report and create activity atomically.
	if err := s.reportRepo.DeleteWithActivity(report.ID, activity); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("report not found")
		}

		return fmt.Errorf("failed to delete report: %w", err)
	}

	return nil
}
