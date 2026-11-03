package service

import (
	"errors"
	"strings"

	"sms-backend/model"

	"gorm.io/gorm"
)

var (
	ErrSafetyCategoryNotFound     = errors.New("safety category not found")
	ErrSafetyCategoryCodeUsed     = errors.New("category code already exists")
	ErrSafetyCategoryInactive     = errors.New("category is already inactive")
	ErrSafetyCategoryCodeRequired = errors.New("category code is required")
	ErrSafetyCategoryNameRequired = errors.New("category name is required")
)

type SafetyCategoryRepository interface {
	FindAll() ([]model.SafetyCategory, error)
	FindByID(id uint) (*model.SafetyCategory, error)
	FindByCode(code string) (*model.SafetyCategory, error)
	FindByCodeExceptID(code string, id uint) (*model.SafetyCategory, error)
	Create(category *model.SafetyCategory) error
	Update(category *model.SafetyCategory) error
	Deactivate(id uint) error
	FindAllActive() ([]model.SafetyCategory, error)
}

type SafetyCategoryService struct {
	repo SafetyCategoryRepository
}

func NewSafetyCategoryService(repo SafetyCategoryRepository) *SafetyCategoryService {
	return &SafetyCategoryService{
		repo: repo,
	}
}

type CreateSafetyCategoryInput struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type UpdateSafetyCategoryInput struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsActive    *bool  `json:"is_active"`
}

func (s *SafetyCategoryService) GetAll() ([]model.SafetyCategory, error) {
	return s.repo.FindAll()
}

func (s *SafetyCategoryService) GetByID(id uint) (*model.SafetyCategory, error) {
	category, err := s.repo.FindByID(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrSafetyCategoryNotFound
	}
	if err != nil {
		return nil, err
	}

	return category, nil
}

func (s *SafetyCategoryService) Create(input CreateSafetyCategoryInput) (*model.SafetyCategory, error) {
	code := strings.ToUpper(strings.TrimSpace(input.Code))
	name := strings.TrimSpace(input.Name)
	description := strings.TrimSpace(input.Description)

	if code == "" {
		return nil, ErrSafetyCategoryCodeRequired
	}

	if name == "" {
		return nil, ErrSafetyCategoryNameRequired
	}

	existing, err := s.repo.FindByCode(code)

	if err == nil && existing != nil {
		return nil, ErrSafetyCategoryCodeUsed
	}

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	category := &model.SafetyCategory{
		Code:        code,
		Name:        name,
		Description: description,
		IsActive:    true,
	}

	if err := s.repo.Create(category); err != nil {
		return nil, err
	}

	return category, nil
}

func (s *SafetyCategoryService) Update(
	id uint,
	input UpdateSafetyCategoryInput,
) (*model.SafetyCategory, error) {

	category, err := s.repo.FindByID(id)

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrSafetyCategoryNotFound
	}

	if err != nil {
		return nil, err
	}

	code := strings.ToUpper(strings.TrimSpace(input.Code))
	name := strings.TrimSpace(input.Name)
	description := strings.TrimSpace(input.Description)

	if code == "" {
		return nil, ErrSafetyCategoryCodeRequired
	}

	if name == "" {
		return nil, ErrSafetyCategoryNameRequired
	}

	existing, err := s.repo.FindByCodeExceptID(code, id)

	if err == nil && existing != nil {
		return nil, ErrSafetyCategoryCodeUsed
	}

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	category.Code = code
	category.Name = name
	category.Description = description

	if input.IsActive != nil {
		category.IsActive = *input.IsActive
	}

	if err := s.repo.Update(category); err != nil {
		return nil, err
	}

	return category, nil
}

func (s *SafetyCategoryService) Delete(id uint) error {
	category, err := s.repo.FindByID(id)

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrSafetyCategoryNotFound
	}

	if err != nil {
		return err
	}

	if !category.IsActive {
		return ErrSafetyCategoryInactive
	}

	return s.repo.Deactivate(id)
}

func (s *SafetyCategoryService) GetActive() ([]model.SafetyCategory, error) {
	return s.repo.FindAllActive()
}

// func (s *SafetyCategoryService) GetActive() ([]model.SafetyCategory, error) {
// 	var categories []model.SafetyCategory

// 	err := s.db.
// 		Where("is_active = ?", true).
// 		Order("name ASC").
// 		Find(&categories).Error

// 	return categories, err
// }
