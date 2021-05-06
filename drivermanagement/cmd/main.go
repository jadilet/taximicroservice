package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/jadilet/taximicroservice/drivermanagement/endpoints"
	"github.com/jadilet/taximicroservice/drivermanagement/pb"
	"github.com/jadilet/taximicroservice/drivermanagement/service"
	"github.com/jadilet/taximicroservice/drivermanagement/transports"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/joho/godotenv"
	"github.com/streadway/amqp"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	var (
		httpAddr = flag.String("http.addr", ":8081", "HTTP listen address")
	)
	flag.Parse()

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

	// waiting_driver_response: the dispatcher services find the nearest driver for the ride
	// then offers the ride to the drivers
	// waits for the driver's response
	// if driver doesn't accept the ride then the ride would be re-queued to the dispatcher service
	_, err = ch.QueueDeclare(
		"waiting_driver_response_queue", // name
		true,                            // durable
		false,                           // delete when unused
		false,                           // exclusive
		false,                           // no-wait
		nil,                             // arguments
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

	dns := fmt.Sprintf(
		"%s:%s@%s(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		os.Getenv("MYSQL_USER"),
		os.Getenv("MYSQL_PASSWORD"),
		os.Getenv("MYSQL_PROTOCOL"),
		os.Getenv("MYSQL_HOST"),
		os.Getenv("MYSQL_PORT"),
		os.Getenv("MYSQL_DBNAME"),
	)

	db, err := gorm.Open(mysql.Open(dns), &gorm.Config{})

	if err != nil {
		logger.Log(err)
		return
	}

	err = db.AutoMigrate(&service.Driver{})

	if err != nil {
		logger.Log(err)
	}

	sqlDB, err := db.DB()

	if err != nil {
		logger.Log(err)
		return
	}

	// SetMaxIdleConns sets the maximum number of connections in the idle connection pool.
	sqlDB.SetMaxIdleConns(15)

	// SetMaxOpenConns sets the maximum number of open connections to the database.
	sqlDB.SetMaxOpenConns(100)

	// SetConnMaxLifetime sets the maximum amount of time a connection may be reused.
	sqlDB.SetConnMaxLifetime(time.Hour)

	s := service.NewDriverService(logger, db, ch)
	h := transports.MakeHTTPHandler(s, log.With(logger, "component", "HTTP"))
	s.CheckResponse(context.Background())

	sendendpoints := endpoints.MakeGrpcEndpoint(s)
	grpcServer := transports.NewGRPCServer(sendendpoints, logger)

	grpcListener, err := net.Listen("tcp", fmt.Sprintf(":%s", os.Getenv("GRPC_DRIVERMANAGEMENT_SRV_PORT")))

	if err != nil {
		logger.Log("during", "Listen Driver Management GRPC Server", "err", err)
		os.Exit(1)
	}

	go func() {
		baseServer := grpc.NewServer()
		reflection.Register(baseServer)
		pb.RegisterDriverServer(baseServer, grpcServer)
		level.Info(logger).Log("msg", "Driver Management GRPC Server started successfully")
		err = baseServer.Serve(grpcListener)
		if err != nil {
			level.Error(logger).Log("msg", "Failed to serve Driver Management GRPC Server")
			os.Exit(1)
		}
	}()

	errs := make(chan error)
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt)
		signal.Notify(sigChan, syscall.SIGTERM)
		errs <- fmt.Errorf("%s", <-sigChan)

	}()

	go func() {
		logger.Log("transport", "HTTP", "addr", *httpAddr)
		errs <- http.ListenAndServe(*httpAddr, h)
	}()

	logger.Log("exit", <-errs)
}
