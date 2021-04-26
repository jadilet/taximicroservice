package service

import (
	"context"

	"github.com/go-kit/kit/log"
	"github.com/go-redis/redis"
	"github.com/golang/protobuf/ptypes/empty"
)

type service struct {
	logger log.Logger
	rdb    *redis.Client
	rdbKey string
}

// Service interface describe  a service that set locations
type Service interface {
	Set(ctx context.Context, key string, lat, lon float64) (*empty.Empty, error)
	Nearest(ctx context.Context, lon, lat, radius float64) ([]redis.GeoLocation, error)
}

func NewService(log log.Logger, rdb *redis.Client, key string) Service {
	return &service{log, rdb, key}
}

func (s service) Nearest(ctx context.Context, lon, lat, radius float64) ([]redis.GeoLocation, error) {

	res, err := s.rdb.GeoRadius(ctx, s.rdbKey, lon, lat, &redis.GeoRadiusQuery{
		Unit: "km", 
		WithDist: true, 
		Radius: radius,
		WithCoord:   true,
		Sort:        "ASC",

		}).Result()
	
	if err != nil {
		return []redis.GeoLocation{}, err
	}

	return res, nil
}

func (s service) Set(ctx context.Context, key string, lat, lon float64) (*empty.Empty, error) {
	var emp empty.Empty

	err := s.rdb.GeoAdd(ctx, s.rdbKey, &redis.GeoLocation{
		Name:      key,
		Longitude: lon,
		Latitude:  lat,
		Dist:      0,
		GeoHash:   0,
	}).Err()

	if err != nil {
		return &emp, err
	}

	return &emp, nil
}
