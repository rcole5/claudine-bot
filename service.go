package claudine_bot

import (
	"context"
	"errors"
	"github.com/jinzhu/gorm"
	"github.com/rcole5/claudine-bot/models"
	"sync"
)

type Service interface {
	// Channel functions
	NewChannel(ctx context.Context, channel string) (Channel, error)
	ListChannel(ctx context.Context) ([]Channel, error)
	DeleteChannel(ctx context.Context, channel string) error

	// Command functions
	NewCommand(ctx context.Context, channel string, c Command) (Command, error)
	GetCommand(ctx context.Context, channel string, trigger string) (Command, error)
	ListCommand(ctx context.Context, channel string) ([]Command, error)
	UpdateCommand(ctx context.Context, channel string, trigger string, action string) (Command, error)
	DeleteCommand(ctx context.Context, channel string, trigger string) error
}

type Command struct {
	Trigger string `json:"trigger"`
	Action  string `json:"action"`
}

type Channel string

var (
	ErrAlreadyExist = errors.New("already exists")
	ErrNotFound     = errors.New("not found")
	//ErrGeneric       = errors.New("generic server error")
)

type claudineService struct {
	mtx          sync.RWMutex
	commands     map[string]map[string]Command
	commandExist map[string]map[string]struct{}
	db           *gorm.DB
}

func NewClaudineService(db *gorm.DB) Service {
	return &claudineService{
		commands:     make(map[string]map[string]Command),
		commandExist: make(map[string]map[string]struct{}),
		db:           db,
	}
}

// Channel Functions
func (s *claudineService) NewChannel(ctx context.Context, channel string) (Channel, error) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	if _, ok := s.commands[channel]; ok {
		return "", ErrAlreadyExist
	}

	if s.commands[channel] == nil {
		s.commands[channel] = make(map[string]Command)
		s.commandExist[channel] = make(map[string]struct{})
	}

	return Channel(channel), nil
}

func (s *claudineService) ListChannel(ctx context.Context) ([]Channel, error) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	var channels []Channel
	for ch := range s.commands {
		channels = append(channels, Channel(ch))
	}
	return channels, nil
}

func (s *claudineService) DeleteChannel(ctx context.Context, channel string) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	_, ok := s.commands[channel]
	if !ok {
		return ErrNotFound
	}
	delete(s.commands, channel)
	delete(s.commandExist, channel)
	return nil
}

// Command Functions
func (s *claudineService) NewCommand(ctx context.Context, channel string, c Command) (Command, error) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	if _, ok := s.commands[channel][c.Trigger]; ok {
		return Command{}, ErrAlreadyExist
	}

	if s.commands[channel] == nil {
		s.commands[channel] = make(map[string]Command)
		s.commandExist[channel] = make(map[string]struct{})
	}

	go func() {
		var existing models.Command
		if s.db.Where("Channel = ? AND Trigger = ?", channel, c.Trigger).Find(&existing).RowsAffected == 0 {
			s.db.Create(&models.Command{
				Trigger: c.Trigger,
				Action:  c.Action,
				Channel: channel,
			})
		}
	}()

	s.commands[channel][c.Trigger] = c
	s.commandExist[channel][c.Trigger] = struct{}{}
	return c, nil
}

func (s *claudineService) GetCommand(ctx context.Context, channel string, trigger string) (Command, error) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	c, ok := s.commands[channel][trigger]
	if !ok {
		return Command{}, ErrNotFound
	}
	return c, nil
}

func (s *claudineService) ListCommand(ctx context.Context, channel string) ([]Command, error) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	var list []Command
	for _, command := range s.commands[channel] {
		list = append(list, command)
	}
	return list, nil
}

func (s *claudineService) UpdateCommand(ctx context.Context, channel string, trigger string, action string) (Command, error) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	_, ok := s.commands[channel][trigger]
	if !ok {
		return Command{}, ErrNotFound

	}
	s.commands[channel][trigger] = Command{
		Trigger: trigger,
		Action:  action,
	}

	return s.commands[channel][trigger], nil
}

func (s *claudineService) DeleteCommand(ctx context.Context, channel string, trigger string) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	_, ok := s.commands[channel][trigger]
	if !ok {
		return ErrNotFound
	}
	delete(s.commands[channel], trigger)
	delete(s.commandExist[channel], trigger)
	return nil
}
