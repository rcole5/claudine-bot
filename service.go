package claudine_bot

import (
	"context"
	"errors"
	"github.com/jinzhu/gorm"
	"sync"
)

type Service interface {
	// Command functions
	NewCommand(ctx context.Context, c Command) (Command, error)
	GetCommand(ctx context.Context, trigger string) (Command, error)
	ListCommand(ctx context.Context) ([]Command, error)
	UpdateCommand(ctx context.Context, trigger string) (Command, error)
	DeleteCommand(ctx context.Context, trigger string) error
}

type Command struct {
	Trigger string `json:"trigger"`
	Action  string `json:"action"`
}

var (
	ErrAlreadyExists = errors.New("already exists")
	ErrNotFound      = errors.New("not found")
	ErrGeneric       = errors.New("generic server error")
)

type claudineService struct {
	mtx          sync.RWMutex
	commands     map[string]string
	commandExist map[string]struct{}
	db           *gorm.DB
}

func NewClaudineService() Service {
	return &claudineService{
		commands: map[string]string,
		commandExist: map[string]struct{},
	}
}

func (s *claudineService) NewCommand(ctx context.Context, c Command) (Command, error) {}

func (s *claudineService) GetCommand(ctx context.Context, trigger string) (Command, error) {}

func (s *claudineService) ListCommand(ctx context.Context) ([]Command, error) {}

func (s *claudineService) UpdateCommand(ctx context.Context, trigger string) (Command, error) {}

func (s *claudineService) DeleteCommand(ctx context.Context, trigger string) error {}
