package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/streadway/amqp"
)

type Ride struct {
	ID          uint
	UUID        string
	PassengerID string
	DriverID    string
	Lat         float64
	Lon         float64
	Addr        string
}

type DispatcherService interface {
	Dispatch(ctx context.Context) error
}

type dispatcherService struct {
	logger log.Logger
	ch     *amqp.Channel
}

func NewDispatcherService(log log.Logger, ch *amqp.Channel) DispatcherService {
	return &dispatcherService{log, ch}
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
			json.Unmarshal(d.Body, &ride)
			s.logger.Log("Processing the ride ", ride.ID, ride.Addr)
			time.Sleep(5 * time.Second)
			d.Nack(false, true)

			//d.Ack(false) // successfull delivered
			/*if ride.Addr == "accept" {
				d.Ack(false)
			} else {
				d.Nack(false, true)
			}*/
			//d.Acknowledger.Reject(d.DeliveryTag, true)
			//d.Nack(false,true)

		}
	}()

	s.logger.Log(" [*] Waiting for messages. ")
	<-forever

	return nil
}
