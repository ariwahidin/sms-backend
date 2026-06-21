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
	return database.DB.Save(user).Error
}

func (r *UserRepository) ExistsByUsername(username string) bool {
	var count int64
	database.DB.Model(&model.User{}).Where("username = ?", username).Count(&count)
	return count > 0
}
