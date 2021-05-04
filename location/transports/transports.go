package transports

import (
	"context"
	"errors"

	"github.com/go-kit/kit/log"
	gt "github.com/go-kit/kit/transport/grpc"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/jadilet/taximicroservice/location/endpoints"
	"github.com/jadilet/taximicroservice/location/pb"
)

type gRPCServer struct {
	set     gt.Handler
	nearest gt.Handler
	pb.UnimplementedLocationServer
}

func NewGRPCServer(endpoint endpoints.Endpoint, logger log.Logger) pb.LocationServer {
	return &gRPCServer{
		set: gt.NewServer(
			endpoint.Set,
			decodeSetRequest,
			encodeSetResponse,
		),
		nearest: gt.NewServer(
			endpoint.Nearest,
			decodeNearestRequest,
			encodeNearestResponse,
		),
	}
}

func (s *gRPCServer) Nearest(ctx context.Context, req *pb.RequestGeo) (*pb.ResponseGeo, error) {
	_, resp, err := s.nearest.ServeGRPC(ctx, req)

	if err != nil {
		return nil, err
	}

	return resp.(*pb.ResponseGeo), nil
}

func (s *gRPCServer) Set(ctx context.Context, req *pb.RequestLocation) (*empty.Empty, error) {
	_, resp, err := s.set.ServeGRPC(ctx, req)

	if err != nil {
		return nil, err
	}

	return resp.(*empty.Empty), nil
}

func decodeSetRequest(_ context.Context, request interface{}) (interface{}, error) {
	req := request.(*pb.RequestLocation)

	if req.P == nil {
		return nil, errors.New("pb.Locaion.Set request location can't be blank")
	}

	p := endpoints.Point{Lat: req.P.Latitude, Lon: req.P.Longitude}
	return endpoints.RequestLocation{Key: req.Key, P: p}, nil
}

func decodeNearestRequest(_ context.Context, request interface{}) (interface{}, error) {
	req := request.(*pb.RequestGeo)

	return endpoints.RequestGeo{Lat: req.Lat, Lon: req.Lon, Radius: req.Radius}, nil
}

func encodeNearestResponse(_ context.Context, response interface{}) (interface{}, error) {
	resp := response.(endpoints.ResponseGeo)

	res := []*pb.GeoLocation{}
	for _, v := range resp.Locations {
		res = append(res, &pb.GeoLocation{
			Name:      v.Name,
			Longitude: v.Longitude,
			Latitude:  v.Latitude,
			Dist:      v.Dist,
			Geohash:   v.GeoHash})
	}

	return &pb.ResponseGeo{Locations: res, Err: resp.Err}, nil
}

func encodeSetResponse(_ context.Context, response interface{}) (interface{}, error) {

	return response, nil

}
