package endpoints

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/jadilet/taximicroservice/drivermanagement/service"
)

type EndpointHttp struct {
	Register endpoint.Endpoint
	Accept   endpoint.Endpoint
	Set      endpoint.Endpoint
}

type EndpointGrpc struct {
	Send endpoint.Endpoint
}

type DriverRegisterReq struct {
	Driver service.Driver
}

type DriverAcceptReq struct {
	Task service.Task
}

type RideReq struct {
	DriverID uint
	Dist     float64
	Lat      float64
	Lon      float64
}

type LocReq struct {
	DriverID uint
	Lat      float64
	Lon      float64
}

type LocResp struct {
	Msg string `json:"msg,omitempty"`
	Err error  `json:"err,omitempty"`
}
type RideResp struct {
	Msg string `json:"msg"`
	Err error  `json:"error,omitempty"`
}
type DriverResp struct {
	Msg string `json:"msg"`
	Err error  `json:"error,omitempty"`
}

func makeSetEndpoint(s service.DriverLocationService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(LocReq)
		err = s.Set(ctx, req.DriverID, req.Lat, req.Lon)

		return LocResp{Err: err}, err
	}
}

func makeRegisterEndpoint(s service.DriverService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(DriverRegisterReq)
		msg, err := s.Register(ctx, req.Driver)

		return DriverResp{Msg: msg, Err: err}, err
	}
}

func makeAcceptEndpoint(s service.DriverService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(DriverAcceptReq)
		msg, err := s.Accept(ctx, req.Task.DriverID, req.Task.RideID)

		return DriverResp{Msg: msg, Err: err}, err
	}
}

func makeSendEndpoint(s service.DriverService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(RideReq)
		msg, err := s.Send(ctx, req.DriverID, req.Lat, req.Lon, req.Dist)

		return RideResp{Msg: msg, Err: err}, nil
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
		Accept:   makeAcceptEndpoint(s),
		Set:      makeSetEndpoint(s),
	}
}