package service

import (
	"context"
	"fmt"

	"github.com/go-kit/kit/log"
	"github.com/streadway/amqp"

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
	CheckResponse(ctx context.Context) error
	Send(ctx context.Context, driverID string, lat float64, lon float64, dist float64) (string, error)
}

type driverService struct {
	logger log.Logger
	db     *gorm.DB
	ch     *amqp.Channel
}

func NewDriverService(log log.Logger, db *gorm.DB, ch *amqp.Channel) DriverService {
	return &driverService{log, db, ch}
}

func (s *driverService) Send(ctx context.Context, driverID string, lat float64,
	lon float64, dist float64) (string, error) {
	s.logger.Log("Offered a ride to the driver ", driverID, " distance: ", dist, "km")
	return fmt.Sprintf("Offered a ride to the driver %s distance %v km", driverID, dist), nil
}

func (s *driverService) Register(ctx context.Context, driver Driver) (string, error) {
	res := s.db.Create(&driver)

	if res.Error != nil {
		return "", res.Error
	}

	return fmt.Sprintf("Successfully added a new driver ID: %d", driver.ID), nil
}

func (s *driverService) CheckResponse(ctx context.Context) error {

	return nil
}
