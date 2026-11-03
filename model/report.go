package model

import (
	"time"

	"gorm.io/gorm"
)

type ReportStatus string
type RiskLevel string
type Priority string

const (
	StatusDeclaration    ReportStatus = "declaration"
	StatusClassification ReportStatus = "classification"
	StatusInvestigation  ReportStatus = "investigation"
	StatusApproval       ReportStatus = "approval"
	StatusClosed         ReportStatus = "closed"

	RiskLow    RiskLevel = "low"
	RiskMedium RiskLevel = "medium"
	RiskHigh   RiskLevel = "high"

	PriorityLow    Priority = "low"
	PriorityMedium Priority = "medium"
	PriorityHigh   Priority = "high"
)

// SLA hours per priority per step (business hours from declaration time)
// Step: classification confirm, investigation submit, approval submit
var SLAHours = map[Priority]map[string]int{
	PriorityHigh: {
		"classification": 4,
		"investigation":  8,
		"approval":       10,
	},
	PriorityMedium: {
		"classification": 8,
		"investigation":  16,
		"approval":       18,
	},
	PriorityLow: {
		"classification": 8,
		"investigation":  24,
		"approval":       26,
	},
}

type RiskReport struct {
	ID   uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Code string `gorm:"uniqueIndex;size:20;not null" json:"code"`

	// SafetyCategoryID uint            `gorm:"not null;index" json:"safety_category_id"`
	// SafetyCategory   *SafetyCategory `gorm:"foreignKey:SafetyCategoryID" json:"safety_category,omitempty"`

	SafetyCategoryID *uint           `gorm:"index" json:"safety_category_id"`
	SafetyCategory   *SafetyCategory `gorm:"foreignKey:SafetyCategoryID" json:"safety_category,omitempty"`

	LocationID      *uint        `json:"location_id"`
	Location        *Location    `gorm:"foreignKey:LocationID" json:"location,omitempty"`
	DepartmentID    *uint        `json:"department_id"`
	Department      *Department  `gorm:"foreignKey:DepartmentID" json:"department,omitempty"`
	Description     string       `gorm:"type:nvarchar(max);not null" json:"description"`
	RiskLevelOrigin RiskLevel    `gorm:"size:20;not null;default:'low'" json:"risk_level_origin"`
	RiskLevel       RiskLevel    `gorm:"size:20;not null;default:'low'" json:"risk_level"`
	Priority        Priority     `gorm:"size:20" json:"priority"`
	Status          ReportStatus `gorm:"size:30;not null;default:'declaration'" json:"status"`
	ReporterName    string       `gorm:"size:150" json:"reporter_name"`

	// PIC fields (classification step)
	PICID        *uint      `json:"pic_id"`
	PIC          *User      `gorm:"foreignKey:PICID" json:"pic,omitempty"`
	PICNotes     string     `gorm:"type:nvarchar(max)" json:"pic_notes"`
	ClassifiedAt *time.Time `json:"classified_at"`
	ClassifiedBy *uint      `json:"classified_by"`

	// Investigation fields
	RootCause       string     `gorm:"type:nvarchar(max)" json:"root_cause"`
	Countermeasures string     `gorm:"type:nvarchar(max)" json:"countermeasures"`
	InvestigatedAt  *time.Time `json:"investigated_at"`
	InvestigatedBy  *uint      `json:"investigated_by"`

	// Approval fields
	HODComment string     `gorm:"type:nvarchar(max)" json:"hod_comment"`
	ApprovedAt *time.Time `json:"approved_at"`
	ApprovedBy *uint      `json:"approved_by"`
	RejectedAt *time.Time `json:"rejected_at"`
	RejectedBy *uint      `json:"rejected_by"`

	// SLA
	SLADeadlineClassification *time.Time `json:"sla_deadline_classification"`
	SLADeadlineInvestigation  *time.Time `json:"sla_deadline_investigation"`
	SLADeadlineApproval       *time.Time `json:"sla_deadline_approval"`
	IsOverdue                 bool       `gorm:"default:false" json:"is_overdue"`

	Files      []ReportFile     `gorm:"foreignKey:ReportID" json:"files,omitempty"`
	Activities []ReportActivity `gorm:"foreignKey:ReportID" json:"activities,omitempty"`

	EditRequests []ReportEditRequest `gorm:"foreignKey:ReportID" json:"edit_requests,omitempty"`

	// CreatedAt time.Time `json:"created_at"`
	// UpdatedAt time.Time `json:"updated_at"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type ReportFile struct {
	ID         uint            `gorm:"primaryKey;autoIncrement" json:"id"`
	ReportID   uint            `gorm:"not null;index" json:"report_id"`
	ActivityID *uint           `gorm:"index" json:"activity_id"`
	Activity   *ReportActivity `gorm:"foreignKey:ActivityID" json:"activity,omitempty"`
	FileName   string          `gorm:"size:255" json:"file_name"`
	FilePath   string          `gorm:"size:500" json:"file_path"`
	FileSize   int64           `json:"file_size"`
	MimeType   string          `gorm:"size:100" json:"mime_type"`
	UploadedBy *uint           `json:"uploaded_by"`
	CreatedAt  time.Time       `json:"created_at"`
}

type ReportActivity struct {
	ID         uint         `gorm:"primaryKey;autoIncrement" json:"id"`
	ReportID   uint         `gorm:"not null;index" json:"report_id"`
	Action     string       `gorm:"size:100" json:"action"`
	FromStatus ReportStatus `gorm:"size:30" json:"from_status"`
	ToStatus   ReportStatus `gorm:"size:30" json:"to_status"`
	Notes      string       `gorm:"type:nvarchar(max)" json:"notes"`
	ByUserID   *uint        `json:"by_user_id"`
	ByUserName string       `gorm:"size:150" json:"by_user_name"`
	CreatedAt  time.Time    `json:"created_at"`

	Files []ReportFile `gorm:"foreignKey:ActivityID" json:"files,omitempty"` // reverse relation, file-file yang lahir dari activity ini
}
