package model

import "time"

type Location struct {
	ID       uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Site     string `gorm:"size:100;not null" json:"site"`
	AreaName string `gorm:"size:150;not null" json:"area_name"`
	QRToken  string `gorm:"uniqueIndex;size:255;not null" json:"qr_token"`
	// DepartmentID: departemen yang bertanggung jawab/memiliki area ini
	// (Occurrence Department sesuai SOP) — tetap per lokasi, ikut ter-prefill
	// otomatis saat QR di-scan. Boleh diubah manual oleh pelapor jika keliru.
	DepartmentID *uint       `json:"department_id"`
	Department   *Department `gorm:"foreignKey:DepartmentID" json:"department,omitempty"`
	IsActive     bool        `gorm:"default:true" json:"is_active"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
}
