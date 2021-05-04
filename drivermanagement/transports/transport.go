package transports

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-kit/kit/log"

	httptransport "github.com/go-kit/kit/transport/http"

	"github.com/go-kit/kit/transport"
	"github.com/gorilla/mux"
	"github.com/jadilet/taximicroservice/drivermanagement/endpoints"
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

func MakeHTTPHandler(s service.DriverService, logger log.Logger) http.Handler {
	r := mux.NewRouter()
	e := endpoints.MakeRegisterEndpoint(s)

	options := []httptransport.ServerOption{
		httptransport.ServerErrorHandler(transport.NewLogErrorHandler(logger)),
		httptransport.ServerErrorEncoder(encodeError),
	}

	r.Methods("POST").Path("/driver/register/").Handler(
		httptransport.NewServer(
			e.Register,
			decodePostDriverRequest,
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

func decodePostDriverRequest(_ context.Context, r *http.Request) (request interface{}, err error) {
	var req endpoints.DriverRequest
	if e := json.NewDecoder(r.Body).Decode(&req.Driver); e != nil {
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
