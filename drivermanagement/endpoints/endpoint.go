package endpoints

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/jadilet/taximicroservice/drivermanagement/service"
)

type EndpointHttp struct {
	Register endpoint.Endpoint
}

type EndpointGrpc struct {
	Send endpoint.Endpoint
}

type DriverRequest struct {
	Driver service.Driver
}

type RideRequest struct {
	DriverID string
	Dist     float64
	Lat      float64
	Lon      float64
}
type RideResponse struct {
	Msg string `json:"msg"`
	Err error  `json:"error,omitempty"`
}
type DriverResponse struct {
	Msg string `json:"msg"`
	Err error  `json:"error,omitempty"`
}

func makeRegisterEndpoint(s service.DriverService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(DriverRequest)
		msg, err := s.Register(ctx, req.Driver)

		return DriverResponse{Msg: msg, Err: err}, err
	}
}

func makeSendEndpoint(s service.DriverService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(RideRequest)
		msg, err := s.Send(ctx, req.DriverID, req.Lat, req.Lon, req.Dist)

		return RideResponse{Msg: msg, Err: err}, nil
	}
}

func MakeGrpcEndpoint(s service.DriverService) EndpointGrpc {
	return EndpointGrpc{
		Send: makeSendEndpoint(s),
	}
}

func MakeHttpEndpoint(s service.DriverService) EndpointHttp {
	return EndpointHttp{
		Register: makeRegisterEndpoint(s),
	}
}
