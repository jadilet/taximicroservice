package endpoints

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-redis/redis"

	"github.com/jadilet/taximicroservice/location/service"
)

type Endpoint struct {
	Set     endpoint.Endpoint
	Nearest endpoint.Endpoint
}

type Point struct {
	Lat float64
	Lon float64
}

type RequestLocation struct {
	Key string
	P   Point
}

type RequestGeo struct {
	Lat    float64
	Lon    float64
	Radius float64
}

type ResponseGeo struct {
	Locations []redis.GeoLocation
	Err       string
}

type ResponseLocation struct {
	Err error
}

func MakeEndpoint(s service.Service) Endpoint {
	return Endpoint{
		Set:     makeSetEndpoint(s),
		Nearest: makeNearestEndpoint(s),
	}
}

func makeSetEndpoint(s service.Service) endpoint.Endpoint {
	return func(ctx context.Context, req interface{}) (resp interface{}, err error) {
		request := req.(RequestLocation)
		resp, e := s.Set(ctx, request.Key, request.P.Lat, request.P.Lon)

		return resp, e
	}
}

func makeNearestEndpoint(s service.Service) endpoint.Endpoint {
	return func(ctx context.Context, req interface{}) (resp interface{}, err error) {
		request := req.(RequestGeo)

		locations, e := s.Nearest(ctx, request.Lon, request.Lat, request.Radius)

		if e != nil {
			return ResponseGeo{Locations: locations, Err: e.Error()}, e
		}

		return ResponseGeo{Locations: locations}, nil
	}
}
