package service

import (
	"context"
	"fmt"

	"github.com/go-kit/kit/log"

	"gorm.io/gorm"
)

type Driver struct {
	gorm.Model
	UUID      string
	Status    string
	Name      string
	Email     string
	Telephone string
	Blocked   bool `gorm:"default:false"`
}

type DriverService interface {
	Register(ctx context.Context, driver Driver) (string, error)
}

type driverService struct {
	logger log.Logger
	db     *gorm.DB
}

func NewDriverService(log log.Logger, db *gorm.DB) DriverService {
	return &driverService{log, db}
}

func (s *driverService) Register(ctx context.Context, driver Driver) (string, error) {
	res := s.db.Create(&driver)

	if res.Error != nil {
		return "", res.Error
	}

	return fmt.Sprintf("Successfully added a new driver ID: %d", driver.ID), nil
}
