package repository

import (
	"sms-backend/database"
	"sms-backend/model"
)

type SafetyCategoryRepository struct{}

func NewSafetyCategoryRepository() *SafetyCategoryRepository {
	return &SafetyCategoryRepository{}
}

func (r *SafetyCategoryRepository) FindAllActive() ([]model.SafetyCategory, error) {
	var categories []model.SafetyCategory

	err := database.DB.
		Where("is_active = ?", true).
		Order("name ASC").
		Find(&categories).Error

	return categories, err
}

// func (r *SafetyCategoryRepository) FindAllActive() ([]model.SafetyCategory, error) {
//     var categories []model.SafetyCategory

//     err := database.DB.
//         Where("is_active = ?", true).
//         Order("name ASC").
//         Find(&categories).Error

//     return categories, err
// }

func (r *SafetyCategoryRepository) FindActiveByID(id uint) (*model.SafetyCategory, error) {
	var category model.SafetyCategory

	err := database.DB.
		Where("id = ? AND is_active = ?", id, true).
		First(&category).Error

	if err != nil {
		return nil, err
	}

	return &category, nil
}

func (r *SafetyCategoryRepository) FindAll() ([]model.SafetyCategory, error) {
	var categories []model.SafetyCategory

	err := database.DB.
		Order("name ASC").
		Find(&categories).Error

	return categories, err
}

func (r *SafetyCategoryRepository) FindByID(id uint) (*model.SafetyCategory, error) {
	var category model.SafetyCategory

	err := database.DB.
		First(&category, id).Error

	if err != nil {
		return nil, err
	}

	return &category, nil
}

func (r *SafetyCategoryRepository) FindByCode(code string) (*model.SafetyCategory, error) {
	var category model.SafetyCategory

	err := database.DB.
		Where("code = ?", code).
		First(&category).Error

	if err != nil {
		return nil, err
	}

	return &category, nil
}

func (r *SafetyCategoryRepository) FindByCodeExceptID(code string, id uint) (*model.SafetyCategory, error) {
	var category model.SafetyCategory

	err := database.DB.
		Where("code = ? AND id <> ?", code, id).
		First(&category).Error

	if err != nil {
		return nil, err
	}

	return &category, nil
}

func (r *SafetyCategoryRepository) Create(category *model.SafetyCategory) error {
	return database.DB.Create(category).Error
}

func (r *SafetyCategoryRepository) Update(category *model.SafetyCategory) error {
	return database.DB.Save(category).Error
}

func (r *SafetyCategoryRepository) Deactivate(id uint) error {
	return database.DB.
		Model(&model.SafetyCategory{}).
		Where("id = ?", id).
		Update("is_active", false).Error
}
