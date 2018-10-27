package claudine_bot

import (
	"context"
	"github.com/go-kit/kit/endpoint"
)

type Endpoints struct {
	NewChannelEndpoint    endpoint.Endpoint
	ListChannelEndpoint   endpoint.Endpoint
	DeleteChannelEndpoint endpoint.Endpoint

	NewCommandEndpoint    endpoint.Endpoint
	GetCommandEndpoint    endpoint.Endpoint
	ListCommandEndpoint   endpoint.Endpoint
	UpdateCommandEndpoint endpoint.Endpoint
	DeleteCommandEndpoint endpoint.Endpoint

	NewRepeatEndpoint    endpoint.Endpoint
	GetRepeatEndpoint    endpoint.Endpoint
	ListRepeatEndpoint   endpoint.Endpoint
	DeleteRepeatEndpoint endpoint.Endpoint
}

func MakeServerEndpoints(s Service) Endpoints {
	return Endpoints{
		NewChannelEndpoint:    MakeNewChannelEndpoint(s),
		ListChannelEndpoint:   MakeListChannelEndpoint(s),
		DeleteChannelEndpoint: MakeDeleteChannelEndpoint(s),

		NewCommandEndpoint:    MakeNewCommandEndpoint(s),
		GetCommandEndpoint:    MakeGetCommandEndpoint(s),
		ListCommandEndpoint:   MakeListCommandEndpoint(s),
		UpdateCommandEndpoint: MakeUpdateCommandEndpoint(s),
		DeleteCommandEndpoint: MakeDeleteCommandEndpoint(s),

		NewRepeatEndpoint:  MakeNewRepeatEndpoint(s),
		GetRepeatEndpoint:  MakeGetRepeatEndpoint(s),
		ListRepeatEndpoint: MakeListRepeatEndpoint(s),
		DeleteRepeatEndpoint: MakeDeleteRepeatEndpoint(s),
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

func MakeNewChannelEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(newChannelRequest)
		c, e := s.NewChannel(ctx, req.Channel)
		return newChannelResponse{Channel: c, Error: e}, nil
	}
}

func MakeListChannelEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		_ = request.(listChannelRequest)
		c, e := s.ListChannel(ctx)
		return listChannelResponse{Channels: c, Error: e}, nil
	}
}

func MakeDeleteChannelEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		c := request.(deleteChannelRequest)
		e := s.DeleteChannel(ctx, string(c.Channel))
		return deleteChannelResponse{Error: e}, nil
	}
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

func MakeNewRepeatEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(newRepeatRequest)
		resp, e := s.NewRepeatCommand(ctx, req.Channel, req.Trigger, req.Duration)
		return newRepeatCommandResponse{RepeatCommand: resp, Error: e}, nil
	}
}

func MakeGetRepeatEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(getRepeatRequest)
		resp, e := s.GetRepeatCommand(ctx, req.Channel, req.Trigger)
		return getRepeatCommandResponse{RepeatCommand: resp, Error: e}, nil
	}
}

func MakeListRepeatEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(listRepeatRequest)
		resp, e := s.ListRepeatCommand(ctx, req.Channel)
		return listRepeatResponse{RepeatCommands: resp, Error: e}, nil
	}
}

func MakeDeleteRepeatEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(deleteRepeatRequest)
		e := s.DeleteRepeatCommand(ctx, req.Channel, req.Trigger)
		return deleteRepeatResponse{Error: e}, nil
	}
}

// New Command
type newCommandRequest struct {
	Command Command
	Channel string
}

type newChannelRequest struct {
	Channel string `json:"channel"`
}

type newCommandResponse struct {
	Command Command `json:"command"`
	Error   error   `json:"error"`
}

type newChannelResponse struct {
	Channel Channel `json:"channel"`
	Error   error   `json:"error"`
}

type listChannelRequest struct{}

type listChannelResponse struct {
	Channels []Channel `json:"channels"`
	Error    error     `json:"error"`
}

type deleteChannelRequest struct {
	Channel Channel `json:"channel"`
}

type deleteChannelResponse struct {
	Error error `json:"error"`
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

type getRepeatRequest struct {
	Channel string `json:"channel"`
	Trigger string `json:"trigger"`
}

type deleteRepeatRequest struct {
	Channel string `json:"channel"`
	Trigger string `json:"trigger"`
}

type getRepeatCommandResponse struct {
	RepeatCommand RepeatCommand `json:"repeat_command"`
	Error         error         `json:"error"`
}

type newRepeatRequest struct {
	Channel  string `json:"channel"`
	Trigger  string `json:"trigger"`
	Duration int    `json:"duration"`
}

type newRepeatCommandResponse struct {
	RepeatCommand RepeatCommand `json:"repeat_command"`
	Error         error         `json:"error"`
}

type listRepeatRequest struct {
	Channel string `json:"channel"`
}

type listRepeatResponse struct {
	RepeatCommands []RepeatCommand `json:"repeat_commands"`
	Error          error           `json:"error"`
}

type deleteRepeatResponse struct {
	Error error `json:"error"`
}

func (r getCommandResponse) error() error { return r.Error }
