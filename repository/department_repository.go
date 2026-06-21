package repository

import (
	"sms-backend/database"
	"sms-backend/model"
)

type DepartmentRepository struct{}

func NewDepartmentRepository() *DepartmentRepository {
	return &DepartmentRepository{}
}

func (r *DepartmentRepository) FindAll() ([]model.Department, error) {
	var depts []model.Department
	err := database.DB.Where("is_active = 1").Order("site, name").Find(&depts).Error
	return depts, err
}

func (r *DepartmentRepository) FindByID(id uint) (*model.Department, error) {
	var dept model.Department
	err := database.DB.Where("id = ? AND is_active = 1", id).First(&dept).Error
	return &dept, err
}

func (r *DepartmentRepository) Create(dept *model.Department) error {
	return database.DB.Create(dept).Error
}

func (r *DepartmentRepository) Update(dept *model.Department) error {
	return database.DB.Save(dept).Error
}
