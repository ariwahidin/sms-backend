package service

import (
	"errors"

	"sms-backend/model"
	"sms-backend/repository"
)

type DepartmentService struct {
	deptRepo *repository.DepartmentRepository
}

func NewDepartmentService(deptRepo *repository.DepartmentRepository) *DepartmentService {
	return &DepartmentService{deptRepo: deptRepo}
}

func (s *DepartmentService) GetAll() ([]model.Department, error) {
	return s.deptRepo.FindAll()
}

type CreateDepartmentInput struct {
	Name string `json:"name"`
	Site string `json:"site"`
}

func (s *DepartmentService) Create(input CreateDepartmentInput) (*model.Department, error) {
	if input.Name == "" {
		return nil, errors.New("name is required")
	}
	dept := &model.Department{Name: input.Name, Site: input.Site}
	return dept, s.deptRepo.Create(dept)
}

type UpdateDepartmentInput struct {
	Name string `json:"name"`
	Site string `json:"site"`
}

func (s *DepartmentService) Update(id uint, input UpdateDepartmentInput) (*model.Department, error) {
	if input.Name == "" {
		return nil, errors.New("name is required")
	}

	dept, err := s.deptRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("department not found")
	}

	dept.Name = input.Name
	dept.Site = input.Site

	return dept, s.deptRepo.Update(dept)
}
