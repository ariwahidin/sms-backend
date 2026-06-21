package service

import (
	"errors"

	"sms-backend/model"
	"sms-backend/repository"
)

type UserService struct {
	userRepo *repository.UserRepository
}

func NewUserService(userRepo *repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

func (s *UserService) GetAll() ([]model.User, error) {
	return s.userRepo.FindAll()
}

type CreateUserInput struct {
	Username     string     `json:"username"`
	Password     string     `json:"password"`
	Name         string     `json:"name"`
	Email        string     `json:"email"`
	Role         model.Role `json:"role"`
	DepartmentID *uint      `json:"department_id"`
}

type UpdateUserInput struct {
	Name         string     `json:"name"`
	Email        string     `json:"email"`
	Role         model.Role `json:"role"`
	DepartmentID *uint      `json:"department_id"`
	IsActive     *bool      `json:"is_active"`
	Password     string     `json:"password"` // opsional, kosong = tidak diubah
}

func (s *UserService) Update(id uint, input UpdateUserInput) (*model.User, error) {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("user not found")
	}
	if input.Name != "" {
		user.Name = input.Name
	}
	user.Email = input.Email
	if input.Role != "" {
		user.Role = input.Role
	}
	user.DepartmentID = input.DepartmentID
	if input.IsActive != nil {
		user.IsActive = *input.IsActive
	}
	if input.Password != "" {
		hashed, err := HashPassword(input.Password)
		if err != nil {
			return nil, errors.New("failed to hash password")
		}
		user.Password = hashed
	}
	return user, s.userRepo.Update(user)
}

func (s *UserService) Create(input CreateUserInput) (*model.User, error) {
	if input.Username == "" || input.Password == "" || input.Name == "" {
		return nil, errors.New("username, password, and name are required")
	}
	if s.userRepo.ExistsByUsername(input.Username) {
		return nil, errors.New("username already exists")
	}

	hashed, err := HashPassword(input.Password)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	user := &model.User{
		Username:     input.Username,
		Password:     hashed,
		Name:         input.Name,
		Email:        input.Email,
		Role:         input.Role,
		DepartmentID: input.DepartmentID,
	}
	return user, s.userRepo.Create(user)
}
