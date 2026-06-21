package model

import (
	"time"
)

type Role string

const (
	RoleAdmin Role = "admin"
	RolePIC   Role = "pic"
	RoleHOD   Role = "hod"
)

type User struct {
	ID           uint        `gorm:"primaryKey;autoIncrement" json:"id"`
	Username     string      `gorm:"uniqueIndex;size:100;not null" json:"username"`
	Password     string      `gorm:"size:255;not null" json:"-"`
	Name         string      `gorm:"size:150;not null" json:"name"`
	Email        string      `gorm:"size:150" json:"email"`
	Role         Role        `gorm:"size:20;not null;default:'pic'" json:"role"`
	DepartmentID *uint       `json:"department_id"`
	Department   *Department `gorm:"foreignKey:DepartmentID" json:"department,omitempty"`
	IsActive     bool        `gorm:"default:true" json:"is_active"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
}
