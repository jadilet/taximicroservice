package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-kit/kit/log"
	dClient "github.com/jadilet/taximicroservice/drivermanagement/pb"
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
	SentAt      time.Time
}

type DispatcherService interface {
	Dispatch(ctx context.Context) error
}

type dispatcherService struct {
	logger         log.Logger
	ch             *amqp.Channel
	locationClient pb.LocationClient
	driverClient   dClient.DriverClient
	radius         float64
}

func NewDispatcherService(log log.Logger, ch *amqp.Channel,
	locationClient pb.LocationClient, driverClient dClient.DriverClient, radius float64) DispatcherService {
	return &dispatcherService{log, ch, locationClient, driverClient, radius}
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
		s.logger.Log("Failed to register to open_ride_queue a consumer", err)
		return err
	}

	forever := make(chan bool)

	
	go func() {
		for d := range msgs {
			var ride Ride
			err := json.Unmarshal(d.Body, &ride)

			if err != nil {
				s.logger.Log("Failed parse ride message body", err)
				if err := d.Nack(false, true); err != nil {
					s.logger.Log("Error negative acknowledging ride", err)
				}
				continue
			}

			s.logger.Log("Dispatcher processing the ride ", ride.ID, ride.Addr)

			resp, err := s.locationClient.Nearest(ctx, &pb.GeoRequest{Lat: ride.Lat,
				Lon: ride.Lon, Radius: s.radius})

			if err != nil {
				s.logger.Log("Error grcp call Nearest: ", err)
				if err := d.Nack(false, true); err != nil {
					s.logger.Log("Error negative acknowledging the ride", err)
				}
				continue
			}

			if len(resp.Locations) != 0 {
				count := 0
				for _, driver := range resp.Locations {
					_, err := s.driverClient.Send(ctx, &dClient.Request{Driverid: driver.Id,
						Dist: driver.Dist,
						Lat:  ride.Lat,
						Lon:  ride.Lon,
					})

					if err != nil {
						count++
						s.logger.Log("Can't send ride to driver:", driver.Name, " ride ID", ride.ID,
							"err", err)
					}
				}

				// if doesn't send any driver then set negative acknowledge about the ride
				if len(resp.Locations) == count {
					if err := d.Nack(false, true); err != nil {
						s.logger.Log("Error negative acknowledging the ride", err)
					}
					continue
				}

				ride.SentAt = time.Now()
				data, err := json.Marshal(&ride)

				if err != nil {
					s.logger.Log("Error encoding ride to json ", err)
					if err := d.Nack(false, true); err != nil {
						s.logger.Log("Error negative acknowledging the ride", err)
					}
					continue
				}

				err = s.ch.Publish(
					"",
					"waiting_driver_response",
					false,
					false,
					amqp.Publishing{
						DeliveryMode: amqp.Persistent,
						ContentType:  "application/json",
						Body:         data,
					})
				if err != nil {
					s.logger.Log("Error to write a ride to the waiting_driver_response channel ", ride.ID)
					if err := d.Nack(false, true); err != nil {
						s.logger.Log("Error negative acknowledging the ride", err)
					}
					continue
				}

				if err := d.Ack(false); err != nil {
					s.logger.Log("Error ackowledging ride", err)
				}

				continue
			}

			s.logger.Log("Not found drivers for trip id ", ride.ID)
			time.Sleep(10 * time.Second)
			if err := d.Nack(false, true); err != nil {
				s.logger.Log("Error negative acknowledging the ride", err)
			}

		}
	}()

	s.logger.Log(" [*] Waiting open_ride_queue for messages. ")
	<-forever

	return nil
}
