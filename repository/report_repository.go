package repository

import (
	"fmt"
	"time"

	"sms-backend/database"
	"sms-backend/model"
)

type ReportRepository struct{}

func NewReportRepository() *ReportRepository {
	return &ReportRepository{}
}

func (r *ReportRepository) GenerateCode() (string, error) {
	now := time.Now()
	year := now.Year() % 100 // ambil 2 digit terakhir (2026 -> 26)

	// Hitung laporan yang dibuat di tahun ini saja
	var count int64
	startOfYear := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
	endOfYear := time.Date(now.Year()+1, 1, 1, 0, 0, 0, 0, now.Location())

	if err := database.DB.Model(&model.RiskReport{}).
		Where("created_at >= ? AND created_at < ?", startOfYear, endOfYear).
		Count(&count).Error; err != nil {
		return "", err
	}

	// Format: YI + 2 digit tahun + 4 digit nomor urut
	return fmt.Sprintf("YI%02d%04d", year, count+1), nil
}

func (r *ReportRepository) Create(report *model.RiskReport) error {
	return database.DB.Create(report).Error
}

func (r *ReportRepository) FindAll(filters map[string]interface{}) ([]model.RiskReport, int64, error) {
	var reports []model.RiskReport
	var total int64

	query := database.DB.Model(&model.RiskReport{}).
		Preload("Location").
		Preload("Department").
		Preload("PIC")

	if status, ok := filters["status"]; ok && status != "" {
		query = query.Where("status = ?", status)
	}
	if site, ok := filters["site"]; ok && site != "" {
		query = query.Joins("JOIN locations ON locations.id = risk_reports.location_id").
			Where("locations.site = ?", site)
	}
	if deptID, ok := filters["department_id"]; ok && deptID != "" {
		query = query.Where("department_id = ?", deptID)
	}
	if level, ok := filters["risk_level"]; ok && level != "" {
		query = query.Where("risk_level = ?", level)
	}
	if from, ok := filters["date_from"]; ok && from != "" {
		query = query.Where("created_at >= ?", from)
	}
	if to, ok := filters["date_to"]; ok && to != "" {
		query = query.Where("created_at <= ?", to)
	}

	query.Count(&total)

	page := 1
	pageSize := 25
	if p, ok := filters["page"].(int); ok && p > 0 {
		page = p
	}
	if ps, ok := filters["page_size"].(int); ok && ps > 0 {
		pageSize = ps
	}

	offset := (page - 1) * pageSize
	err := query.Order("created_at desc").Offset(offset).Limit(pageSize).Find(&reports).Error
	return reports, total, err
}

func (r *ReportRepository) Update(report *model.RiskReport) error {
	return database.DB.Save(report).Error
}

func (r *ReportRepository) AddActivity(activity *model.ReportActivity) error {
	return database.DB.Create(activity).Error
}

func (r *ReportRepository) AddFile(file *model.ReportFile) error {
	return database.DB.Create(file).Error
}

func (r *ReportRepository) FindByID(id uint) (*model.RiskReport, error) {
	var report model.RiskReport
	err := database.DB.
		Preload("Location").
		Preload("Department").
		Preload("PIC").
		Preload("Files").
		Preload("Files.Activity").
		Preload("Activities").
		First(&report, id).Error
	return &report, err
}

func (r *ReportRepository) FindFiles(reportID uint) ([]model.ReportFile, error) {
	var files []model.ReportFile
	err := database.DB.Preload("Activity").Where("report_id = ?", reportID).Find(&files).Error
	return files, err
}

func (r *ReportRepository) FindFileByID(reportID, fileID uint) (*model.ReportFile, error) {
	var file model.ReportFile
	err := database.DB.
		Preload("Activity").
		Where("id = ? AND report_id = ?", fileID, reportID).
		First(&file).Error
	return &file, err
}

func (r *ReportRepository) FindOverdue() ([]model.RiskReport, error) {
	var reports []model.RiskReport
	now := time.Now()

	err := database.DB.
		Preload("PIC").
		Preload("Department").
		Where("status NOT IN (?) AND ("+
			"(status = ? AND sla_deadline_classification < ?) OR "+
			"(status = ? AND sla_deadline_investigation < ?) OR "+
			"(status = ? AND sla_deadline_approval < ?)"+
			")",
			[]string{string(model.StatusClosed)},
			model.StatusClassification, now,
			model.StatusInvestigation, now,
			model.StatusApproval, now,
		).Find(&reports).Error
	return reports, err
}

func (r *ReportRepository) DashboardStats(filters map[string]interface{}) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// by_status → flat object
	var statusStats []struct {
		Status string
		Count  int64
	}
	database.DB.Model(&model.RiskReport{}).
		Select("status, count(*) as count").
		Group("status").
		Scan(&statusStats)

	byStatus := map[string]int64{
		"declaration":    0,
		"classification": 0,
		"investigation":  0,
		"approval":       0,
		"closed":         0,
	}
	var total int64
	for _, s := range statusStats {
		byStatus[s.Status] = s.Count
		total += s.Count
	}
	stats["by_status"] = byStatus
	stats["total"] = total

	// by_risk → flat object
	var levelStats []struct {
		RiskLevel string
		Count     int64
	}
	database.DB.Model(&model.RiskReport{}).
		Select("risk_level, count(*) as count").
		Group("risk_level").
		Scan(&levelStats)

	byRisk := map[string]int64{"high": 0, "medium": 0, "low": 0}
	for _, l := range levelStats {
		byRisk[l.RiskLevel] = l.Count
	}
	stats["by_risk"] = byRisk

	// overdue = off_track
	var onTrack, offTrack int64
	database.DB.Model(&model.RiskReport{}).Where("is_overdue = 0 AND status != ?", model.StatusClosed).Count(&onTrack)
	database.DB.Model(&model.RiskReport{}).Where("is_overdue = 1").Count(&offTrack)
	stats["on_track"] = onTrack
	stats["off_track"] = offTrack
	stats["overdue"] = offTrack // ← alias yang dipakai frontend

	// sisanya tetap sama
	var siteStats []struct {
		Site  string
		Count int64
	}
	database.DB.Model(&model.RiskReport{}).
		Joins("JOIN locations ON locations.id = risk_reports.location_id").
		Select("locations.site as site, count(*) as count").
		Group("locations.site").
		Scan(&siteStats)
	stats["by_site"] = siteStats

	var deptStats []struct {
		Name  string
		Count int64
	}
	database.DB.Model(&model.RiskReport{}).
		Joins("JOIN departments ON departments.id = risk_reports.department_id").
		Select("departments.name as name, count(*) as count").
		Group("departments.name").
		Scan(&deptStats)
	stats["by_department"] = deptStats

	var monthStats []struct {
		Month     int
		RiskLevel string
		Count     int64
	}
	database.DB.Model(&model.RiskReport{}).
		Select("MONTH(created_at) as month, risk_level, count(*) as count").
		Where("YEAR(created_at) = YEAR(GETDATE())").
		Group("MONTH(created_at), risk_level").
		Order("month").
		Scan(&monthStats)
	stats["monthly"] = monthStats

	return stats, nil
}
