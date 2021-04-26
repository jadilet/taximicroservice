package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/go-kit/kit/log"
	"github.com/streadway/amqp"
	"gorm.io/gorm"
)

var (
	ErrInconsistentIDs = errors.New("inconsistent IDs")
	ErrAlreadyExists   = errors.New("already exists")
	ErrNotFound        = errors.New("not found")
)

type Ride struct {
	gorm.Model
	UUID        string
	PassengerID string
	DriverID    string
	Lat         float64
	Lon         float64
	Addr        string
}

type TripService interface {
	AddRide(ctx context.Context, ride Ride) (string, error)
}

type tripService struct {
	logger log.Logger
	db     *gorm.DB
	ch     *amqp.Channel
}

func NewTripService(log log.Logger, db *gorm.DB, ch *amqp.Channel) TripService {
	return &tripService{logger: log, db: db, ch: ch}
}

func (srv *tripService) AddRide(ctx context.Context, ride Ride) (string, error) {
	res := srv.db.Create(&ride)

	if res.Error != nil {
		return "", res.Error
	}

	data, err := json.Marshal(&ride)

	if err != nil {
		return "", err
	}

	err = srv.ch.Publish(
		"",                // exchange
		"open_ride_queue", // routing key
		false,             // mandatory
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         data,
		})

	if err != nil {
		return "", err
	}

	return fmt.Sprintf("ride added id: %d", ride.ID), nil
}
