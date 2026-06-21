package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"

	"sms-backend/model"
	"sms-backend/repository"
)

type LocationService struct {
	locationRepo *repository.LocationRepository
}

func NewLocationService(locationRepo *repository.LocationRepository) *LocationService {
	return &LocationService{locationRepo: locationRepo}
}

func (s *LocationService) GetAll() ([]model.Location, error) {
	return s.locationRepo.FindAll()
}

func (s *LocationService) GetByQRToken(token string) (*model.Location, error) {
	loc, err := s.locationRepo.FindByQRToken(token)
	if err != nil {
		return nil, errors.New("location not found")
	}
	return loc, nil
}

type CreateLocationInput struct {
	Site         string `json:"site"`
	AreaName     string `json:"area_name"`
	DepartmentID *uint  `json:"department_id"`
}

func (s *LocationService) Create(input CreateLocationInput) (*model.Location, error) {
	if input.Site == "" || input.AreaName == "" {
		return nil, errors.New("site and area_name are required")
	}

	token, err := generateQRToken()
	if err != nil {
		return nil, errors.New("failed to generate QR token")
	}

	loc := &model.Location{
		Site:         input.Site,
		AreaName:     input.AreaName,
		QRToken:      token,
		DepartmentID: input.DepartmentID,
	}
	return loc, s.locationRepo.Create(loc)
}

type UpdateLocationInput struct {
	Site         string `json:"site"`
	AreaName     string `json:"area_name"`
	DepartmentID *uint  `json:"department_id"`
}

func (s *LocationService) Update(id uint, input UpdateLocationInput) (*model.Location, error) {
	if input.Site == "" || input.AreaName == "" {
		return nil, errors.New("site and area_name are required")
	}

	loc, err := s.locationRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("location not found")
	}

	// QRToken sengaja tidak diubah — QR code yang sudah dicetak/ditempel di
	// lapangan harus tetap valid walau site/area name/department diedit.
	loc.Site = input.Site
	loc.AreaName = input.AreaName
	loc.DepartmentID = input.DepartmentID

	return loc, s.locationRepo.Update(loc)
}

func generateQRToken() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
