package claudine_bot

import (
	"context"
	"github.com/go-kit/kit/endpoint"
)

type Endpoints struct {
	NewCommandEndpoint    endpoint.Endpoint
	GetCommandEndpoint    endpoint.Endpoint
	ListCommandEndpoint   endpoint.Endpoint
	UpdateCommandEndpoint endpoint.Endpoint
	DeleteCommandEndpoint endpoint.Endpoint
}

func MakeServerEndpoints(s Service) Endpoints {
	return Endpoints{
		NewCommandEndpoint:    MakeNewCommandEndpoint(s),
		GetCommandEndpoint:    MakeGetCommandEndpoint(s),
		ListCommandEndpoint:   MakeListCommandEndpoint(s),
		UpdateCommandEndpoint: MakeUpdateCommandEndpoint(s),
		DeleteCommandEndpoint: MakeDeleteCommandEndpoint(s),
	}
}

func (e Endpoints) NewCommand(ctx context.Context, c Command) (Command, error) {
	request := newCommandRequest{Command: c}
	response, err := e.NewCommandEndpoint(ctx, request)
	if err != nil {
		return Command{}, err
	}
	resp := response.(newCommandResponse)
	return resp.Command, resp.Error
}

func (e Endpoints) GetCommand(ctx context.Context, t string) (Command, error) {
	request := getCommandRequest{Trigger: t}
	response, err := e.GetCommandEndpoint(ctx, request)
	if err != nil {
		return Command{}, err
	}
	resp := response.(getCommandResponse)
	return resp.Command, resp.Error
}

func (e Endpoints) ListCommand(ctx context.Context) ([]Command, error) {
	request := listCommandRequest{}
	response, err := e.ListCommandEndpoint(ctx, request)
	if err != nil {
		return []Command{}, err
	}
	resp := response.(listCommandResponse)
	return resp.Commands, resp.Error
}

func (e Endpoints) UpdateCommand(ctx context.Context, c Command) (Command, error) {
	request := updateCommandRequest{Trigger: c.Trigger, Action: c.Action}
	response, err := e.UpdateCommandEndpoint(ctx, request)
	if err != nil {
		return Command{}, err
	}
	resp := response.(updateCommandResponse)
	return resp.Command, resp.Error
}

func (e Endpoints) DeleteCommand(ctx context.Context, trigger string) error {
	request := updateCommandRequest{Trigger: trigger}
	response, err := e.DeleteCommandEndpoint(ctx, request)
	if err != nil {
		return err
	}
	resp := response.(deleteCommandResponse)
	return resp.Error
}

func MakeNewCommandEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(newCommandRequest)
		c, e := s.NewCommand(ctx, req.Channel, req.Command)
		return newCommandResponse{Command: c, Error: e}, nil
	}
}

func MakeGetCommandEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(getCommandRequest)
		c, e := s.GetCommand(ctx, req.Channel, req.Trigger)
		return getCommandResponse{Command: c, Error: e}, nil
	}
}

func MakeListCommandEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(listCommandRequest)
		c, e := s.ListCommand(ctx, req.Channel)
		return listCommandResponse{Commands: c, Error: e}, nil
	}
}

func MakeUpdateCommandEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(updateCommandRequest)
		c, e := s.UpdateCommand(ctx, req.Channel, req.Trigger, req.Action)
		return updateCommandResponse{Command: c, Error: e}, nil
	}
}

func MakeDeleteCommandEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(deleteCommandRequest)
		e := s.DeleteCommand(ctx, req.Channel, req.Trigger)
		return deleteCommandResponse{Error: e}, nil
	}
}

// New Command
type newCommandRequest struct {
	Command Command
	Channel string
}

type newCommandResponse struct {
	Command Command `json:"command"`
	Error   error   `json:"error"`
}

func (r newCommandResponse) error() error { return r.Error }

// Get Command
type getCommandRequest struct {
	Trigger string
	Channel string
}

type getCommandResponse struct {
	Command Command `json:"command"`
	Error   error   `json:"error"`
}

type listCommandRequest struct {
	Channel string
}

type listCommandResponse struct {
	Commands []Command `json:"commands"`
	Error    error     `json:"error"`
}

type updateCommandRequest struct {
	Trigger string `json:"trigger"`
	Action  string `json:"action"`
	Channel string `json:"channel"`
}

type updateCommandResponse struct {
	Command Command `json:"command"`
	Error   error   `json:"error"`
}

type deleteCommandRequest struct {
	Trigger string `json:"trigger"`
	Channel string `json:"channel"`
}

type deleteCommandResponse struct {
	Error error `json:"error"`
}

func (r getCommandResponse) error() error { return r.Error }
