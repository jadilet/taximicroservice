package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/jadilet/taximicroservice/location/pb"
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

// Driver
type Task struct {
	gorm.Model
	Status   string
	DriverID uint
	RideID   uint
}

type Ride struct {
	UUID        string
	PassengerID string
	DriverID    string
	Lat         float64
	Lon         float64
	Addr        string
	ID          uint
	CreatedAt   time.Time
	UpdatedAt   time.Time
	SentAt      time.Time
}

type DriverService interface {
	Register(ctx context.Context, driver Driver) (string, error)
	CheckResponse(ctx context.Context) error
	Send(ctx context.Context, driverID uint, lat float64, lon float64, dist float64) (string, error)
	Accept(ctx context.Context, driverID, rideID uint) (string, error)
	Set(ctx context.Context, driverID uint, lat float64, lon float64) error
}

type DriverLocationService interface {
	Set(ctx context.Context, driverID uint, lat float64, lon float64) error
}

type driverService struct {
	logger    log.Logger
	master    *gorm.DB
	slave     *gorm.DB
	ch        *amqp.Channel
	mutex     sync.Mutex
	locClient pb.LocationClient
}

func NewDriverService(log log.Logger, master *gorm.DB, slave *gorm.DB, ch *amqp.Channel, locClient pb.LocationClient) DriverService {
	return &driverService{logger: log, master: master, slave: slave, ch: ch, locClient: locClient}
}

func (s *driverService) Set(ctx context.Context, driverID uint, lat float64, lon float64) error {
	_, err := s.locClient.Set(ctx, &pb.RequestLocation{Key: int32(driverID), P: &pb.Point{Latitude: lat, Longitude: lon}})

	return err
}

func (s *driverService) Accept(ctx context.Context, driverID, rideID uint) (string, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	var driver Driver

	if err := s.master.First(&driver, "id = ?", driverID).Error; err != nil {
		return "", err
	}

	if driver.Blocked {
		return "", errors.New("Driver blocked")
	}

	var task Task
	if resp := s.master.Where("ride_id = ?", rideID).First(&task); resp.Error != nil {

		if !errors.Is(resp.Error, gorm.ErrRecordNotFound) {
			return "", resp.Error
		}

		if err := s.master.Transaction(func(tx *gorm.DB) error {
			if err := tx.Create(&Task{DriverID: driverID, RideID: rideID}).Error; err != nil {
				return err
			}

			return nil
		}); err != nil {
			return "", err
		}

		return fmt.Sprintf("Driver %d accepted the ride %d", driverID, rideID), nil
	}

	return "", fmt.Errorf("Ride has been already accepted RideID=%d", rideID)
}

func (s *driverService) Send(ctx context.Context, driverID uint, lat float64,
	lon float64, dist float64) (string, error) {

	if err := s.slave.First(&Driver{}, "id = ?", driverID).Error; err != nil {
		return "", err
	}

	s.logger.Log("Offered a ride to the driver ", driverID, " distance: ", dist, "km")
	return fmt.Sprintf("Offered a ride to the driver %d distance %v km", driverID, dist), nil
}

func (s *driverService) Register(ctx context.Context, driver Driver) (string, error) {
	res := s.master.Create(&driver)

	if res.Error != nil {
		return "", res.Error
	}

	return fmt.Sprintf("Successfully added a new driver ID: %d", driver.ID), nil
}

func (s *driverService) CheckResponse(ctx context.Context) error {
	msgs, err := s.ch.Consume(
		"waiting_driver_response",
		"",
		false,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		s.logger.Log("Failed to register to waiting_driver_response a consumer", err)
		return err
	}

	forever := make(chan bool)

	var ride Ride
	go func() {
		for d := range msgs {
			err := json.Unmarshal(d.Body, &ride)

			if err != nil {
				s.logger.Log("Failed parse ride message body")
				if err := d.Nack(false, true); err != nil {
					s.logger.Log("Error negative acknowledging the ride", err)
				}
				continue
			}

			s.logger.Log("DriverManagement processing the ride ", ride.ID)

			// Wait for 15 seconds for driver response
			// If doesn't accept the ride then resend the ride
			// to the dispatcher service
			elapsed := time.Now().UTC().Sub(ride.SentAt)
			s.logger.Log("Checking response elapsed time", elapsed)
			if elapsed < time.Second*15 {
				s.logger.Log("Waiting driver response for RideID", ride.ID, "elapsed", time.Second*15-elapsed)
				time.Sleep(time.Second*15 - elapsed)
			}

			if elapsed > time.Hour*1 {
				if err := d.Ack(false); err != nil {
					s.logger.Log("Error ackowledging the ride", err)
				}
				s.logger.Log("Old ride acknowledged RideID=", ride.ID)
				continue
			}

			var task Task
			if resp := s.slave.Where("ride_id = ?", ride.ID).First(&task); resp.Error != nil && errors.Is(resp.Error, gorm.ErrRecordNotFound) {
				if err := s.ch.Publish(
					"",
					"open_ride_queue",
					false,
					false,
					amqp.Publishing{
						DeliveryMode: amqp.Persistent,
						ContentType:  "application/json",
						Body:         d.Body,
					}); err != nil {
					s.logger.Log("Error to write a ride to the waiting_driver_response channel ", ride.ID)
					if err := d.Nack(false, true); err != nil {
						s.logger.Log("Error negative acknowledging the ride", err)
					}
				}

				if err := d.Ack(false); err != nil {
					s.logger.Log("Error acknowledging ride")
				}

			} else {
				if err := d.Ack(false); err != nil {
					s.logger.Log("Error ackowledging the ride", err)
				}
				// TODO PUBLISH Message to other services
				s.logger.Log("Ride accepted by DriverID=", task.DriverID, "RideID=", task.RideID)
			}

		}
	}()

	s.logger.Log(" [*] Waiting waiting_driver_response for messages.")
	<-forever

	return nil
}
