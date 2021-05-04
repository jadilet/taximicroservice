package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/jadilet/taximicroservice/location/pb"
	"github.com/streadway/amqp"
)

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
}

type DispatcherService interface {
	Dispatch(ctx context.Context) error
}

type dispatcherService struct {
	logger log.Logger
	ch     *amqp.Channel
	locationSrv pb.LocationClient
	radius float64
}

func NewDispatcherService(log log.Logger, ch *amqp.Channel,
	client pb.LocationClient, radius float64) DispatcherService {
	return &dispatcherService{log, ch, client, radius}
}

func (s *dispatcherService) Dispatch(ctx context.Context) error {

	msgs, err := s.ch.Consume(
		"open_ride_queue", // queue
		"",                // consumer
		false,             // auto-ack
		false,             // exclusive
		false,             // no-local
		false,             // no-wait
		nil,               // args
	)

	if err != nil {
		s.logger.Log("Failed to register a consumer", err)
		return err
	}

	forever := make(chan bool)

	var ride Ride
	go func() {
		for d := range msgs {
			err := json.Unmarshal(d.Body, &ride)

			if err != nil {
				s.logger.Log("Failed to parse sub message body", err)
				d.Nack(false, true)
				continue
			}

			s.logger.Log("Processing the ride ", ride.ID, ride.Addr)

			resp, err := s.locationSrv.Nearest(ctx, &pb.RequestGeo{Lat: ride.Lat,
				Lon: ride.Lon, Radius: s.radius})

			if err != nil {
				s.logger.Log("Error grcp call Nearest: ", err)
				d.Nack(false, true)
				continue
			}

			if len(resp.Locations) != 0 {
				loc := resp.Locations[0]

				s.logger.Log("Trip with ID", ride.ID, " assigned to Driver ", loc.Name)
				d.Ack(false)
			} else {
				s.logger.Log("Not found drivers for trip id ", ride.ID)
				time.Sleep(10 * time.Second)
				d.Nack(false, true)
			}
		}
	}()

	s.logger.Log(" [*] Waiting for messages. ")
	<-forever

	return nil
}
