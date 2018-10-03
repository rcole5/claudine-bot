package claudine_bot

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/go-kit/kit/log"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
)

var (
	ErrBadRouting = errors.New("inconsistent mapping between route and handler (programmer error)")
)

func MakeHTTPHandler(s Service, logger log.Logger) http.Handler {
	r := mux.NewRouter().StrictSlash(false).PathPrefix("/api/v1").Subrouter()
	e := MakeServerEndpoints(s)
	options := []httptransport.ServerOption{
		httptransport.ServerErrorLogger(logger),
		httptransport.ServerErrorEncoder(encodeError),
	}
	// Channels
	r.Methods("POST").Path("/channels").Handler(httptransport.NewServer(
		e.NewChannelEndpoint,
		decodeNewChannelRequest,
		encodeResponse,
		options...,
	))
	r.Methods("GET").Path("/channels").Handler(httptransport.NewServer(
		e.ListChannelEndpoint,
		decodeListChannelRequest,
		encodeResponse,
		options...,
	))

	// Commands
	r.Methods("POST").Path("/channels/{channel}/commands").Handler(httptransport.NewServer(
		e.NewCommandEndpoint,
		decodeNewCommandRequest,
		encodeResponse,
		options...,
	))
	r.Methods("GET").Path("/channels/{channel}/commands/{trigger}").Handler(httptransport.NewServer(
		e.GetCommandEndpoint,
		decodeGetCommandRequest,
		encodeResponse,
		options...,
	))
	r.Methods("GET").Path("/channels/{channel}/commands").Handler(httptransport.NewServer(
		e.ListCommandEndpoint,
		decodeListCommandRequest,
		encodeResponse,
		options...,
	))
	r.Methods("PUT").Path("/channels/{channel}/commands/{trigger}").Handler(httptransport.NewServer(
		e.UpdateCommandEndpoint,
		decodeUpdateCommandRequest,
		encodeResponse,
		options...,
	))
	r.Methods("DELETE").Path("/channels/{channel}/commands/{trigger}").Handler(httptransport.NewServer(
		e.DeleteCommandEndpoint,
		decodeDeleteCommandEndpoint,
		encodeResponse,
		options...,
	))
	return r
}

func decodeNewChannelRequest(_ context.Context, r *http.Request) (request interface{}, err error) {
	var req newChannelRequest
	if e := json.NewDecoder(r.Body).Decode(&req); e != nil {
		return nil, e
	}

	return req, nil
}

func decodeListChannelRequest(_ context.Context, r *http.Request) (request interface{}, err error) {
	return listChannelRequest{}, nil
}

func decodeNewCommandRequest(_ context.Context, r *http.Request) (request interface{}, err error) {
	var req newCommandRequest
	if e := json.NewDecoder(r.Body).Decode(&req.Command); e != nil {
		return nil, e
	}

	channel, ok := mux.Vars(r)["channel"]
	if !ok {
		return nil, ErrBadRouting
	}
	req.Channel = channel
	return req, nil
}

func decodeGetCommandRequest(_ context.Context, r *http.Request) (request interface{}, err error) {
	vars := mux.Vars(r)
	trigger, ok := vars["trigger"]
	if !ok {
		return nil, ErrBadRouting
	}

	channel, ok := vars["channel"]
	if !ok {
		return nil, ErrBadRouting
	}

	return getCommandRequest{Channel: channel, Trigger: trigger}, nil
}

func decodeListCommandRequest(_ context.Context, r *http.Request) (request interface{}, err error) {
	vars := mux.Vars(r)
	channel, ok := vars["channel"]
	if !ok {
		return nil, ErrBadRouting
	}
	return listCommandRequest{Channel: channel}, nil
}

func decodeUpdateCommandRequest(_ context.Context, r *http.Request) (request interface{}, err error) {
	vars := mux.Vars(r)
	trigger, ok := vars["trigger"]
	if !ok {
		return nil, ErrBadRouting
	}

	channel, ok := vars["channel"]
	if !ok {
		return nil, ErrBadRouting
	}

	var req updateCommandRequest
	if e := json.NewDecoder(r.Body).Decode(&req); e != nil {
		return nil, e
	}
	req.Trigger = trigger
	req.Channel = channel

	return req, nil
}

func decodeDeleteCommandEndpoint(_ context.Context, r *http.Request) (request interface{}, err error) {
	vars := mux.Vars(r)
	trigger, ok := vars["trigger"]
	channel, ok := vars["channel"]
	if !ok {
		return nil, ErrBadRouting
	}
	return deleteCommandRequest{Channel: channel, Trigger: trigger}, nil
}

func encodeNewCommandRequest(ctx context.Context, req *http.Request, request interface{}) error {
	req.URL.Path = "/profiles/"
	return encodeRequest(ctx, req, request)
}

type errorer interface {
	error() error
}

func encodeResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	if e, ok := response.(errorer); ok && e.error() != nil {
		encodeError(ctx, e.error(), w)
		return nil
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}

func encodeRequest(_ context.Context, req *http.Request, request interface{}) error {
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(request)
	if err != nil {
		return err
	}
	req.Body = ioutil.NopCloser(&buf)
	return nil
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
	case ErrNotFound:
		return http.StatusNotFound
	case ErrAlreadyExist:
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}
