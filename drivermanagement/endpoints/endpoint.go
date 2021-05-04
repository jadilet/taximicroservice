package endpoints

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/jadilet/taximicroservice/drivermanagement/service"
)

type Endpoint struct {
	Register endpoint.Endpoint
}

type DriverRequest struct {
	Driver service.Driver
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

func MakeRegisterEndpoint(s service.DriverService) Endpoint {
	return Endpoint{
		Register: makeRegisterEndpoint(s),
	}
}
