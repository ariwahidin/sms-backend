// package service

// import (
// 	"errors"
// 	"time"

// 	"sms-backend/email"
// 	"sms-backend/model"
// 	"sms-backend/repository"
// )

// type ReportService struct {
// 	reportRepo *repository.ReportRepository
// 	emailSvc   *email.EmailService
// }

// func NewReportService(reportRepo *repository.ReportRepository, emailSvc *email.EmailService) *ReportService {
// 	return &ReportService{reportRepo: reportRepo, emailSvc: emailSvc}
// }

// type DeclareInput struct {
// 	LocationID   *uint           `json:"location_id"`
// 	DepartmentID *uint           `json:"department_id"`
// 	Description  string          `json:"description"`
// 	RiskLevel    model.RiskLevel `json:"risk_level"`
// 	ReporterName string          `json:"reporter_name"`
// }

// func (s *ReportService) Declare(input DeclareInput) (*model.RiskReport, error) {
// 	if input.Description == "" {
// 		return nil, errors.New("description is required")
// 	}
// 	if input.RiskLevel == "" {
// 		input.RiskLevel = model.RiskLow
// 	}

// 	code, err := s.reportRepo.GenerateCode()
// 	if err != nil {
// 		return nil, err
// 	}

// 	report := &model.RiskReport{
// 		Code:         code,
// 		LocationID:   input.LocationID,
// 		DepartmentID: input.DepartmentID,
// 		Description:  input.Description,
// 		RiskLevel:    input.RiskLevel,
// 		Status:       model.StatusDeclaration,
// 		ReporterName: input.ReporterName,
// 	}

// 	if err := s.reportRepo.Create(report); err != nil {
// 		return nil, err
// 	}

// 	s.reportRepo.AddActivity(&model.ReportActivity{
// 		ReportID: report.ID,
// 		Action:   "declared",
// 		ToStatus: model.StatusDeclaration,
// 		Notes:    "Risk declared by " + safeStr(input.ReporterName, "anonymous"),
// 	})

// 	go s.emailSvc.SendDeclarationNotice(report)

// 	return report, nil
// }

// type ClassifyInput struct {
// 	LocationID   *uint           `json:"location_id"`
// 	DepartmentID *uint           `json:"department_id"`
// 	RiskLevel    model.RiskLevel `json:"risk_level"`
// 	Priority     model.Priority  `json:"priority"`
// 	PICID        *uint           `json:"pic_id"`
// 	Notes        string          `json:"notes"`
// 	ByUserID     uint
// 	ByUserName   string
// }

// func (s *ReportService) Classify(reportID uint, input ClassifyInput) (*model.RiskReport, error) {
// 	report, err := s.reportRepo.FindByID(reportID)
// 	if err != nil {
// 		return nil, errors.New("report not found")
// 	}
// 	if report.Status != model.StatusDeclaration {
// 		return nil, errors.New("report is not in declaration status")
// 	}

// 	now := time.Now()
// 	sla := model.SLAHours[input.Priority]

// 	report.Status = model.StatusClassification
// 	report.Priority = input.Priority
// 	report.RiskLevel = input.RiskLevel
// 	report.PICID = input.PICID
// 	report.PICNotes = input.Notes
// 	report.ClassifiedAt = &now
// 	report.ClassifiedBy = &input.ByUserID

// 	if input.LocationID != nil {
// 		report.LocationID = input.LocationID
// 	}
// 	if input.DepartmentID != nil {
// 		report.DepartmentID = input.DepartmentID
// 	}

// 	investigationDeadline := now.Add(time.Duration(sla["investigation"]) * time.Hour)
// 	approvalDeadline := now.Add(time.Duration(sla["approval"]) * time.Hour)
// 	report.SLADeadlineInvestigation = &investigationDeadline
// 	report.SLADeadlineApproval = &approvalDeadline

// 	if err := s.reportRepo.Update(report); err != nil {
// 		return nil, err
// 	}

// 	s.reportRepo.AddActivity(&model.ReportActivity{
// 		ReportID:   report.ID,
// 		Action:     "classified",
// 		FromStatus: model.StatusDeclaration,
// 		ToStatus:   model.StatusClassification,
// 		Notes:      input.Notes,
// 		ByUserID:   &input.ByUserID,
// 		ByUserName: input.ByUserName,
// 	})

// 	go s.emailSvc.SendStatusUpdateNotice(report, "Classification confirmed")

// 	return report, nil
// }

// type InvestigateInput struct {
// 	RootCause       string `json:"root_cause"`
// 	Countermeasures string `json:"countermeasures"`
// 	ByUserID        uint
// 	ByUserName      string
// }

// func (s *ReportService) Investigate(reportID uint, input InvestigateInput) (*model.RiskReport, error) {
// 	report, err := s.reportRepo.FindByID(reportID)
// 	if err != nil {
// 		return nil, errors.New("report not found")
// 	}
// 	if report.Status != model.StatusClassification && report.Status != model.StatusInvestigation {
// 		return nil, errors.New("report cannot be investigated at current status")
// 	}
// 	if input.RootCause == "" || input.Countermeasures == "" {
// 		return nil, errors.New("root cause and countermeasures are required")
// 	}

// 	now := time.Now()
// 	report.Status = model.StatusInvestigation
// 	report.RootCause = input.RootCause
// 	report.Countermeasures = input.Countermeasures
// 	report.InvestigatedAt = &now
// 	report.InvestigatedBy = &input.ByUserID

// 	if err := s.reportRepo.Update(report); err != nil {
// 		return nil, err
// 	}

// 	s.reportRepo.AddActivity(&model.ReportActivity{
// 		ReportID:   report.ID,
// 		Action:     "investigated",
// 		FromStatus: model.StatusClassification,
// 		ToStatus:   model.StatusInvestigation,
// 		Notes:      "Root cause & countermeasures submitted",
// 		ByUserID:   &input.ByUserID,
// 		ByUserName: input.ByUserName,
// 	})

// 	go s.emailSvc.SendReviewNotice(report)

// 	return report, nil
// }

// func (s *ReportService) Approve(reportID uint, comment string, byUserID uint, byUserName string) (*model.RiskReport, error) {
// 	report, err := s.reportRepo.FindByID(reportID)
// 	if err != nil {
// 		return nil, errors.New("report not found")
// 	}
// 	if report.Status != model.StatusInvestigation {
// 		return nil, errors.New("report is not ready for approval")
// 	}

// 	now := time.Now()
// 	report.Status = model.StatusClosed
// 	report.HODComment = comment
// 	report.ApprovedAt = &now
// 	report.ApprovedBy = &byUserID

// 	if err := s.reportRepo.Update(report); err != nil {
// 		return nil, err
// 	}

// 	s.reportRepo.AddActivity(&model.ReportActivity{
// 		ReportID:   report.ID,
// 		Action:     "approved",
// 		FromStatus: model.StatusInvestigation,
// 		ToStatus:   model.StatusClosed,
// 		Notes:      comment,
// 		ByUserID:   &byUserID,
// 		ByUserName: byUserName,
// 	})

// 	go s.emailSvc.SendClosingNotice(report)

// 	return report, nil
// }

// func (s *ReportService) Reject(reportID uint, comment string, byUserID uint, byUserName string) (*model.RiskReport, error) {
// 	report, err := s.reportRepo.FindByID(reportID)
// 	if err != nil {
// 		return nil, errors.New("report not found")
// 	}
// 	if report.Status != model.StatusInvestigation {
// 		return nil, errors.New("report is not ready for review")
// 	}

// 	now := time.Now()
// 	report.Status = model.StatusClassification
// 	report.HODComment = comment
// 	report.RejectedAt = &now
// 	report.RejectedBy = &byUserID

// 	if err := s.reportRepo.Update(report); err != nil {
// 		return nil, err
// 	}

// 	s.reportRepo.AddActivity(&model.ReportActivity{
// 		ReportID:   report.ID,
// 		Action:     "rejected",
// 		FromStatus: model.StatusInvestigation,
// 		ToStatus:   model.StatusClassification,
// 		Notes:      comment,
// 		ByUserID:   &byUserID,
// 		ByUserName: byUserName,
// 	})

// 	go s.emailSvc.SendRejectionNotice(report)

// 	return report, nil
// }

// func safeStr(s, fallback string) string {
// 	if s == "" {
// 		return fallback
// 	}
// 	return s
// }

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
)

type ReportService struct {
	reportRepo *repository.ReportRepository
	emailSvc   *email.EmailService
}

func NewReportService(reportRepo *repository.ReportRepository, emailSvc *email.EmailService) *ReportService {
	return &ReportService{reportRepo: reportRepo, emailSvc: emailSvc}
}

type DeclareInput struct {
	LocationID   *uint                   `json:"location_id"`
	DepartmentID *uint                   `json:"department_id"`
	Description  string                  `json:"description"`
	RiskLevel    model.RiskLevel         `json:"risk_level"`
	ReporterName string                  `json:"reporter_name"`
	Files        []*multipart.FileHeader // diisi dari handler, bukan JSON
}

func (s *ReportService) Declare(input DeclareInput) (*model.RiskReport, error) {
	if input.Description == "" {
		return nil, errors.New("description is required")
	}
	if input.RiskLevel == "" {
		input.RiskLevel = model.RiskLow
	}

	// validasi ukuran semua file sebelum mulai transaksi
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
		Code:            code,
		LocationID:      input.LocationID,
		DepartmentID:    input.DepartmentID,
		Description:     input.Description,
		RiskLevelOrigin: input.RiskLevel,
		RiskLevel:       input.RiskLevel,
		Status:          model.StatusDeclaration,
		ReporterName:    input.ReporterName,
	}

	tx := database.DB.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	if err := tx.Create(report).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// simpan activity dalam tx
	if err := tx.Create(&model.ReportActivity{
		ReportID: report.ID,
		Action:   "declared",
		ToStatus: model.StatusDeclaration,
		Notes:    "Risk declared by " + safeStr(input.ReporterName, "anonymous"),
	}).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// upload files jika ada
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
			ReportID: report.ID,
			FileName: fh.Filename,
			FilePath: savePath,
			FileSize: fh.Size,
			MimeType: fh.Header.Get("Content-Type"),
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

	go s.emailSvc.SendDeclarationNotice(report)

	return report, nil
}

type ClassifyInput struct {
	LocationID   *uint           `json:"location_id"`
	DepartmentID *uint           `json:"department_id"`
	RiskLevel    model.RiskLevel `json:"risk_level"`
	Priority     model.Priority  `json:"priority"`
	PICID        *uint           `json:"pic_id"`
	Notes        string          `json:"notes"`
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

	if err := s.reportRepo.Update(report); err != nil {
		return nil, err
	}

	s.reportRepo.AddActivity(&model.ReportActivity{
		ReportID:   report.ID,
		Action:     "classified",
		FromStatus: model.StatusDeclaration,
		ToStatus:   model.StatusClassification,
		Notes:      input.Notes,
		ByUserID:   &input.ByUserID,
		ByUserName: input.ByUserName,
	})

	go s.emailSvc.SendStatusUpdateNotice(report, "Classification confirmed")

	return report, nil
}

type InvestigateInput struct {
	RootCause       string `json:"root_cause"`
	Countermeasures string `json:"countermeasures"`
	ByUserID        uint
	ByUserName      string
}

func (s *ReportService) Investigate(reportID uint, input InvestigateInput) (*model.RiskReport, error) {
	report, err := s.reportRepo.FindByID(reportID)
	if err != nil {
		return nil, errors.New("report not found")
	}
	if report.Status != model.StatusClassification && report.Status != model.StatusInvestigation {
		return nil, errors.New("report cannot be investigated at current status")
	}
	if input.RootCause == "" || input.Countermeasures == "" {
		return nil, errors.New("root cause and countermeasures are required")
	}

	now := time.Now()
	report.Status = model.StatusInvestigation
	report.RootCause = input.RootCause
	report.Countermeasures = input.Countermeasures
	report.InvestigatedAt = &now
	report.InvestigatedBy = &input.ByUserID

	if err := s.reportRepo.Update(report); err != nil {
		return nil, err
	}

	s.reportRepo.AddActivity(&model.ReportActivity{
		ReportID:   report.ID,
		Action:     "investigated",
		FromStatus: model.StatusClassification,
		ToStatus:   model.StatusInvestigation,
		Notes:      "Root cause & countermeasures submitted",
		ByUserID:   &input.ByUserID,
		ByUserName: input.ByUserName,
	})

	go s.emailSvc.SendReviewNotice(report)

	return report, nil
}

func (s *ReportService) Approve(reportID uint, comment string, byUserID uint, byUserName string) (*model.RiskReport, error) {
	report, err := s.reportRepo.FindByID(reportID)
	if err != nil {
		return nil, errors.New("report not found")
	}
	if report.Status != model.StatusInvestigation {
		return nil, errors.New("report is not ready for approval")
	}

	now := time.Now()
	report.Status = model.StatusClosed
	report.HODComment = comment
	report.ApprovedAt = &now
	report.ApprovedBy = &byUserID

	if err := s.reportRepo.Update(report); err != nil {
		return nil, err
	}

	s.reportRepo.AddActivity(&model.ReportActivity{
		ReportID:   report.ID,
		Action:     "approved",
		FromStatus: model.StatusInvestigation,
		ToStatus:   model.StatusClosed,
		Notes:      comment,
		ByUserID:   &byUserID,
		ByUserName: byUserName,
	})

	go s.emailSvc.SendClosingNotice(report)

	return report, nil
}

func (s *ReportService) Reject(reportID uint, comment string, byUserID uint, byUserName string) (*model.RiskReport, error) {
	report, err := s.reportRepo.FindByID(reportID)
	if err != nil {
		return nil, errors.New("report not found")
	}
	if report.Status != model.StatusInvestigation {
		return nil, errors.New("report is not ready for review")
	}

	now := time.Now()
	report.Status = model.StatusClassification
	report.HODComment = comment
	report.RejectedAt = &now
	report.RejectedBy = &byUserID

	if err := s.reportRepo.Update(report); err != nil {
		return nil, err
	}

	s.reportRepo.AddActivity(&model.ReportActivity{
		ReportID:   report.ID,
		Action:     "rejected",
		FromStatus: model.StatusInvestigation,
		ToStatus:   model.StatusClassification,
		Notes:      comment,
		ByUserID:   &byUserID,
		ByUserName: byUserName,
	})

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

func safeStr(s, fallback string) string {
	if s == "" {
		return fallback
	}
	return s
}
