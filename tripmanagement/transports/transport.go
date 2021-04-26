package transports

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/transport"
	"github.com/gorilla/mux"
	"github.com/jadilet/taximicroservice/tripmanagement/endpoints"
	"github.com/jadilet/taximicroservice/tripmanagement/service"

	httptransport "github.com/go-kit/kit/transport/http"
)

type errorer interface {
	error() error
}

var (
	// ErrBadRouting is returned when an expected path variable is missing.
	// It always indicates programmer error.
	ErrBadRouting = errors.New("inconsistent mapping between route and handler (programmer error)")
)

func MakeHTTPHandler(s service.TripService, logger log.Logger) http.Handler {
	r := mux.NewRouter()
	e := endpoints.MakeEndpoint(s)

	options := []httptransport.ServerOption{
		httptransport.ServerErrorHandler(transport.NewLogErrorHandler(logger)),
		httptransport.ServerErrorEncoder(encodeError),
	}

	r.Methods("POST").Path("/trip/").Handler(
		httptransport.NewServer(
			e.AddRide,
			decodePostTripRequest,
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

func decodePostTripRequest(_ context.Context, r *http.Request) (request interface{}, err error) {
	var req endpoints.RideRequest
	if e := json.NewDecoder(r.Body).Decode(&req.Ride); e != nil {
		return nil, e
	}

	return req, nil
}

func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	if err == nil {
		panic("encodeError with nil error")
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(codeFrom(err))

	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}

func codeFrom(err error) int {
	switch err {
	case service.ErrNotFound:
		return http.StatusNotFound
	case service.ErrAlreadyExists, service.ErrInconsistentIDs:
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}
