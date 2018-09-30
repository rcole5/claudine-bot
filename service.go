package claudine_bot

import (
	"context"
	"errors"
	"fmt"
	"github.com/jinzhu/gorm"
	"sync"
)

type Service interface {
	// Command functions
	NewCommand(ctx context.Context, c Command) (Command, error)
	GetCommand(ctx context.Context, trigger string) (Command, error)
	ListCommand(ctx context.Context) ([]Command, error)
	//UpdateCommand(ctx context.Context, trigger string) (Command, error)
	//DeleteCommand(ctx context.Context, trigger string) error
}

type Command struct {
	Trigger string `json:"trigger"`
	Action  string `json:"action"`
}

var (
	ErrAlreadyExist = errors.New("already exists")
	ErrNotFound      = errors.New("not found")
	//ErrGeneric       = errors.New("generic server error")
)

type claudineService struct {
	mtx          sync.RWMutex
	commands     map[string]Command
	commandExist map[string]struct{}
	db           *gorm.DB
}

func NewClaudineService() Service {
	return &claudineService{
		commands: make(map[string]Command),
		commandExist: make(map[string]struct{}),
	}
}

func (s *claudineService) NewCommand(ctx context.Context, c Command) (Command, error) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	if _, ok := s.commandExist[c.Action]; ok {
		return Command{}, ErrAlreadyExist
	}
	s.commands[c.Trigger] = c
	s.commandExist[c.Trigger] = struct{}{}
	return c, nil
}

func (s *claudineService) GetCommand(ctx context.Context, trigger string) (Command, error) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	c, ok := s.commands[trigger]
	fmt.Println(s.commands)
	if !ok {
		return Command{}, ErrNotFound
	}
	return c, nil
}

func (s *claudineService) ListCommand(ctx context.Context) ([]Command, error) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	var list []Command
	for _, command := range s.commands {
		list = append(list, command)
	}
	return list, nil
}

//func (s *claudineService) UpdateCommand(ctx context.Context, trigger string) (Command, error) {}
//
//func (s *claudineService) DeleteCommand(ctx context.Context, trigger string) error {}
