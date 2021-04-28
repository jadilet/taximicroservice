package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/go-redis/redis"
	"github.com/jadilet/taximicroservice/location/endpoints"
	"github.com/jadilet/taximicroservice/location/pb"
	"github.com/jadilet/taximicroservice/location/service"
	"github.com/jadilet/taximicroservice/location/transports"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	var logger log.Logger
	logger = log.NewJSONLogger(os.Stdout)
	logger = log.With(logger, "ts", log.DefaultTimestampUTC)
	logger = log.With(logger, "caller", log.DefaultCaller)

	err := godotenv.Load()

	if err != nil {
		level.Info(logger).Log("msg", "failed to read .env file", err)
	}

	// redis
	redisURL := fmt.Sprintf("redis://%s:%s@%s:%s/%s",
		os.Getenv("REDIS_USER"),
		os.Getenv("REDIS_PASSWORD"),
		os.Getenv("REDIS_HOST"),
		os.Getenv("REDIS_PORT"),
		os.Getenv("REDIS_DB"),
	)

	opt, err := redis.ParseURL(redisURL)

	if err != nil {
		level.Error(logger).Log("err", "Failed to connect redis server", err)
		return
	}

	rdb := redis.NewClient(opt)

	setservice := service.NewService(logger, rdb, "drivers.location")
	setendpoints := endpoints.MakeEndpoint(setservice)
	grpcServer := transports.NewGRPCServer(setendpoints, logger)

	grpcListener, err := net.Listen("tcp", ":50051")

	if err != nil {
		logger.Log("during", "Listen", "err", err)
		os.Exit(1)
	}

	go func() {
		baseServer := grpc.NewServer()
		reflection.Register(baseServer)
		pb.RegisterLocationServer(baseServer, grpcServer)
		level.Info(logger).Log("msg", "Server started successfully")
		err = baseServer.Serve(grpcListener)
		if err != nil {
			level.Error(logger).Log("msg", "Failed to serve grpc server")
			os.Exit(1)
		}

	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	signal.Notify(sigChan, syscall.SIGTERM)

	sig := <-sigChan

	level.Error(logger).Log("exit", sig)
}
