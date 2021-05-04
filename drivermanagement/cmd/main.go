package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/jadilet/taximicroservice/drivermanagement/service"
	"github.com/jadilet/taximicroservice/drivermanagement/transports"
	"github.com/joho/godotenv"
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

	s := service.NewDriverService(logger, db)
	h := transports.MakeHTTPHandler(s, log.With(logger, "component", "HTTP"))

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
