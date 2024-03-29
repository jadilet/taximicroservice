package transports

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-kit/kit/log"

	gt "github.com/go-kit/kit/transport/grpc"
	httptransport "github.com/go-kit/kit/transport/http"

	"github.com/go-kit/kit/transport"
	"github.com/gorilla/mux"
	"github.com/jadilet/taximicroservice/drivermanagement/endpoints"
	"github.com/jadilet/taximicroservice/drivermanagement/pb"
	"github.com/jadilet/taximicroservice/drivermanagement/service"
)

type errorer interface {
	error() error
}

var (
	// ErrBadRouting is returned when an expected path variable is missing.
	// It always indicates programmer error.
	ErrBadRouting = errors.New("inconsistent mapping between route and handler")

	ErrInconsistentIDs = errors.New("inconsistent IDs")
	ErrAlreadyExists   = errors.New("already exists")
	ErrNotFound        = errors.New("not found")
)

type gRPCServer struct {
	send gt.Handler
	pb.UnimplementedDriverServer
}

func NewGRPCServer(endpoint endpoints.EndpointGrpc, logger log.Logger) pb.DriverServer {
	return &gRPCServer{
		send: gt.NewServer(
			endpoint.Send,
			decodeSendRequest,
			encodeSendResponse,
		),
	}
}

func (s *gRPCServer) Send(ctx context.Context, req *pb.Request) (*pb.Response, error) {
	_, resp, err := s.send.ServeGRPC(ctx, req)

	if err != nil {
		return nil, err
	}

	return resp.(*pb.Response), nil
}

func decodeSendRequest(_ context.Context, request interface{}) (interface{}, error) {
	req := request.(*pb.Request)

	return endpoints.RideReq{DriverID: uint(req.Driverid),
		Dist: req.Dist,
		Lat:  req.Lat,
		Lon:  req.Lon}, nil
}

func encodeSendResponse(_ context.Context, response interface{}) (interface{}, error) {
	resp := response.(endpoints.RideResp)

	if resp.Err != nil {
		return &pb.Response{Msg: resp.Msg, Err: resp.Err.Error()}, nil
	}

	return &pb.Response{Msg: resp.Msg}, nil
}

func MakeHTTPHandler(s service.DriverService, logger log.Logger) http.Handler {
	r := mux.NewRouter()
	e := endpoints.MakeHttpEndpoint(s)

	options := []httptransport.ServerOption{
		httptransport.ServerErrorHandler(transport.NewLogErrorHandler(logger)),
		httptransport.ServerErrorEncoder(encodeError),
	}

	r.Methods("POST").Path("/driver/register/").Handler(
		httptransport.NewServer(
			e.Register,
			decodePostDriverRegisterReq,
			encodeResponse,
			options...,
		))

	r.Methods("POST").Path("/driver/accept/").Handler(
		httptransport.NewServer(
			e.Accept,
			decodePostDriverAcceptReq,
			encodeResponse,
			options...,
		))

	r.Methods("POST").Path("/driver/set/").Handler(
		httptransport.NewServer(
			e.Set,
			decodePostDriverSetReq,
			encodeResponse,
			options...,
		))

	return r
}

func encodeResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	if e, ok := response.(errorer); ok && e.error() != nil {
		encodeError(ctx, e.error(), w)
		return nil
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}

func decodePostDriverSetReq(_ context.Context, r *http.Request) (request interface{}, err error) {
	var req endpoints.LocReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, err
	}

	return req, nil
}

func decodePostDriverRegisterReq(_ context.Context, r *http.Request) (request interface{}, err error) {
	var req endpoints.DriverRegisterReq
	if err := json.NewDecoder(r.Body).Decode(&req.Driver); err != nil {
		return nil, err
	}

	return req, nil
}

func decodePostDriverAcceptReq(_ context.Context, r *http.Request) (request interface{}, err error) {
	var req endpoints.DriverAcceptReq
	if err := json.NewDecoder(r.Body).Decode(&req.Task); err != nil {
		return nil, err
	}

	return req, nil
}

func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	if err == nil {
		panic("encodeError with nil error")
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(codeFrom(err))

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}

func codeFrom(err error) int {
	switch err {
	case ErrNotFound:
		return http.StatusNotFound
	case ErrAlreadyExists, ErrInconsistentIDs:
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}
