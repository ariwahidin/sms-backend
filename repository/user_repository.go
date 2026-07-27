package repository

import (
	"sms-backend/database"
	"sms-backend/model"
)

type UserRepository struct{}

func NewUserRepository() *UserRepository {
	return &UserRepository{}
}

func (r *UserRepository) FindByUsername(username string) (*model.User, error) {
	var user model.User
	err := database.DB.Preload("Department").Where("username = ? AND is_active = 1", username).First(&user).Error
	return &user, err
}

func (r *UserRepository) FindByID(id uint) (*model.User, error) {
	var user model.User
	err := database.DB.Preload("Department").Where("id = ? AND is_active = 1", id).First(&user).Error
	return &user, err
}

func (r *UserRepository) FindAll() ([]model.User, error) {
	var users []model.User
	err := database.DB.Preload("Department").Order("created_at desc").Find(&users).Error
	return users, err
}

func (r *UserRepository) Create(user *model.User) error {
	return database.DB.Create(user).Error
}

func (r *UserRepository) Update(user *model.User) error {
	return database.DB.Omit("Department").Save(user).Error
}

func (r *UserRepository) ExistsByUsername(username string) bool {
	var count int64
	database.DB.Model(&model.User{}).Where("username = ?", username).Count(&count)
	return count > 0
}

// FindActivePICs returns all active PIC users across all departments.
// When filterByDepartment is true, only returns PICs in the given departmentID.
func (r *UserRepository) FindActivePICs(filterByDepartment bool, departmentID uint) ([]model.User, error) {
	var users []model.User
	q := database.DB.Preload("Department").Where("role = ? AND is_active = 1", model.RolePIC)
	if filterByDepartment && departmentID > 0 {
		q = q.Where("department_id = ?", departmentID)
	}
	err := q.Find(&users).Error
	return users, err
}

// FindActiveHODsByDepartment returns active HOD users for a specific department.
// If departmentID is 0, returns all active HODs.
func (r *UserRepository) FindActiveHODsByDepartment(departmentID uint) ([]model.User, error) {
	var users []model.User
	q := database.DB.Preload("Department").Where("role = ? AND is_active = 1", model.RoleHOD)
	if departmentID > 0 {
		q = q.Where("department_id = ?", departmentID)
	}
	err := q.Find(&users).Error
	return users, err
}
