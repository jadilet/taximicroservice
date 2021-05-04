package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/go-kit/kit/log"
	"github.com/jadilet/taximicroservice/dispatcher/service"
	"github.com/jadilet/taximicroservice/location/pb"
	"github.com/joho/godotenv"
	"github.com/streadway/amqp"
	"google.golang.org/grpc"
)

func main() {

	err := godotenv.Load()

	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = log.With(logger, "caller", log.DefaultCaller)
	}

	if err != nil {
		logger.Log("Error while reading the .env file")
	}

	var radius float64

	radius, err = strconv.ParseFloat(os.Getenv("DISPATCHER_RADIUS"), 64)

	if err != nil {
		logger.Log("DISPATCHER_RADIUS environment variable required", err)
		return
	}

	url := fmt.Sprintf("amqp://%s:%s@%s:%s/",
		os.Getenv("RABBITMQ_USER"),
		os.Getenv("RABBITMQ_PASSWORD"),
		os.Getenv("RABBITMQ_HOST"),
		os.Getenv("RABBITMQ_PORT"),
	)

	conn, err := amqp.Dial(url)

	if err != nil {
		logger.Log("Failed to connect to RabbitMQ", err)
		return
	}

	defer conn.Close()

	ch, err := conn.Channel()

	if err != nil {

		logger.Log("Failed to open a channel", err)
		return
	}

	defer ch.Close()

	_, err = ch.QueueDeclare(
		"open_ride_queue", // name
		true,              // durable
		false,             // delete when unused
		false,             // exclusive
		false,             // no-wait
		nil,               // arguments
	)

	if err != nil {
		logger.Log("Failed to declare a queue", err)
		return
	}

	err = ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)

	if err != nil {
		logger.Log("Failed to set Qos", err)
		return
	}

	var opts []grpc.DialOption = []grpc.DialOption{grpc.WithInsecure()}
	addr := fmt.Sprintf(":%s",
		os.Getenv("GRPC_LOCATION_SRV_PORT"))

	grpcConn, err := grpc.Dial(addr, opts...)
	if err != nil {
		logger.Log("Can't connect to GRPC server", err)
		return
	}

	defer grpcConn.Close()

	client := pb.NewLocationClient(grpcConn)

	s := service.NewDispatcherService(logger, ch, client, radius)
	s.Dispatch(context.Background())

	errs := make(chan error)
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt)
		signal.Notify(sigChan, syscall.SIGTERM)
		errs <- fmt.Errorf("%s", <-sigChan)

	}()
	logger.Log("exit", <-errs)
}
