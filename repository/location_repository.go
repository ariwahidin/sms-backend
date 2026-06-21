package repository

import (
	"sms-backend/database"
	"sms-backend/model"
)

type LocationRepository struct{}

func NewLocationRepository() *LocationRepository {
	return &LocationRepository{}
}

func (r *LocationRepository) FindAll() ([]model.Location, error) {
	var locations []model.Location
	err := database.DB.Preload("Department").Where("is_active = 1").Order("site, area_name").Find(&locations).Error
	return locations, err
}

func (r *LocationRepository) FindByID(id uint) (*model.Location, error) {
	var loc model.Location
	err := database.DB.Preload("Department").Where("id = ? AND is_active = 1", id).First(&loc).Error
	return &loc, err
}

func (r *LocationRepository) FindByQRToken(token string) (*model.Location, error) {
	var loc model.Location
	err := database.DB.Preload("Department").Where("qr_token = ? AND is_active = 1", token).First(&loc).Error
	return &loc, err
}

func (r *LocationRepository) Create(loc *model.Location) error {
	return database.DB.Create(loc).Error
}

func (r *LocationRepository) Update(loc *model.Location) error {
	return database.DB.Save(loc).Error
}
