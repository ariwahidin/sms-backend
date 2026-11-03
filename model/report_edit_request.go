package model

import "time"

type ReportEditRequestStatus string

const (
	EditRequestPending  ReportEditRequestStatus = "pending"
	EditRequestApproved ReportEditRequestStatus = "approved"
	EditRequestRejected ReportEditRequestStatus = "rejected"
)

type ReportEditRequest struct {
	ID uint `gorm:"primaryKey;autoIncrement" json:"id"`

	ReportID uint        `gorm:"not null;index" json:"report_id"`
	Report   *RiskReport `gorm:"foreignKey:ReportID" json:"report,omitempty"`

	// =========================================================
	// Requested Detection Values
	// =========================================================

	LocationID *uint     `json:"location_id"`
	Location   *Location `gorm:"foreignKey:LocationID" json:"location,omitempty"`

	SafetyCategoryID *uint           `json:"safety_category_id"`
	SafetyCategory   *SafetyCategory `gorm:"foreignKey:SafetyCategoryID" json:"safety_category,omitempty"`

	DepartmentID *uint       `json:"department_id"`
	Department   *Department `gorm:"foreignKey:DepartmentID" json:"department,omitempty"`

	RiskLevel RiskLevel `gorm:"size:20;not null" json:"risk_level"`

	Description string `gorm:"type:nvarchar(max);not null" json:"description"`

	// =========================================================
	// Request Information
	// =========================================================

	Reason string `gorm:"type:nvarchar(max);not null" json:"reason"`

	RequestedBy   *uint  `json:"requested_by"`
	RequestedName string `gorm:"size:150" json:"requested_name"`

	// =========================================================
	// Approval
	// =========================================================

	Status ReportEditRequestStatus `gorm:"size:20;not null;default:'pending';index" json:"status"`

	ApprovedBy   *uint      `json:"approved_by"`
	ApprovedName string     `gorm:"size:150" json:"approved_name"`
	ApprovedAt   *time.Time `json:"approved_at"`

	RejectedBy   *uint      `json:"rejected_by"`
	RejectedName string     `gorm:"size:150" json:"rejected_name"`
	RejectedAt   *time.Time `json:"rejected_at"`

	ApprovalNote string `gorm:"type:nvarchar(max)" json:"approval_note"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
