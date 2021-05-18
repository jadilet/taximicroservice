package endpoints

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/jadilet/taximicroservice/tripmanagement/service"
)

type Endpoint struct {
	AddRide endpoint.Endpoint
}

type RideReq struct {
	Ride service.Ride
}

type RideResp struct {
	Msg string `json:"msg"`
	Err error  `json:"error,omitempty"`
}

func makeAddRideEndpoint(s service.TripService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(RideReq)
		msg, err := s.AddRide(ctx, req.Ride)
		return RideResp{Msg: msg, Err: err}, err
	}
}
func MakeEndpoint(s service.TripService) Endpoint {
	return Endpoint{
		AddRide: makeAddRideEndpoint(s),
	}
}
