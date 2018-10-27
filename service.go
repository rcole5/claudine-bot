package claudine_bot

import (
	"bytes"
	"context"
	"errors"
	bolt "github.com/etcd-io/bbolt"
	"strconv"
	"sync"
)

var (
	TRUE  = []byte{1}
	FALSE = []byte{0}
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

	NewRepeatCommand(ctx context.Context, channel string, trigger string, duration int) (RepeatCommand, error)
	GetRepeatCommand(ctx context.Context, channel string, trigger string) (RepeatCommand, error)
	ListRepeatCommand(ctx context.Context, channel string) ([]RepeatCommand, error)
	DeleteRepeatCommand(ctx context.Context, channel string, trigger string) error
}

type Command struct {
	Trigger string `json:"trigger"`
	Action  string `json:"action"`
}

type RepeatCommand struct {
	Trigger  string `json:"trigger"`
	Duration int    `json:"duration"`
}

type Channel string

var (
	ErrAlreadyExist = errors.New("already exists")
	ErrNotFound     = errors.New("not found")
	ErrGeneric      = errors.New("generic server error")
)

type claudineService struct {
	mtx          sync.RWMutex
	db           *bolt.DB
}

func NewClaudineService(db *bolt.DB) Service {
	return &claudineService{
		db:           db,
	}
}

// Channel Functions
func (s *claudineService) NewChannel(ctx context.Context, channel string) (Channel, error) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	// Create a channel bucket
	err := s.db.Update(func(tx *bolt.Tx) error {
		//b := tx.Bucket([]byte(channel))
		b, err := tx.CreateBucketIfNotExists([]byte(channel))
		if err != nil {
			return err
		}

		// Create the command bucket
		b.CreateBucketIfNotExists([]byte("commands"))

		// Enable the channel
		b.Put([]byte("enabled"), TRUE)

		return nil
	})

	if err != nil {
		return "", ErrAlreadyExist
	}

	return Channel(channel), nil
}

func (s *claudineService) ListChannel(ctx context.Context) ([]Channel, error) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	var channels []Channel
	err := s.db.View(func(tx *bolt.Tx) error {
		err := tx.ForEach(func(name []byte, b *bolt.Bucket) error {
			if bytes.Compare(b.Get([]byte("enabled")), TRUE) == 0 {
				channels = append(channels, Channel(name))
			}
			return nil
		})
		return err
	})
	if err != nil {
		return []Channel{}, err
	}
	return channels, nil
}

func (s *claudineService) DeleteChannel(ctx context.Context, channel string) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	err := s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(channel))
		if b == nil {
			return ErrNotFound
		}

		err := b.Put([]byte("enabled"), FALSE)

		return err
	})
	return err
}

// Command Functions
func (s *claudineService) NewCommand(ctx context.Context, channel string, c Command) (Command, error) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	err := s.db.Update(func(tx *bolt.Tx) error {
		cBucket, err := GetActiveCommandBucket(tx, channel)
		if err != nil {
			return err
		}

		// Check if command exists
		command := cBucket.Get([]byte(c.Action))
		if command != nil {
			return ErrAlreadyExist
		}

		// Create command
		err = cBucket.Put([]byte(c.Trigger), []byte(c.Action))
		return err
	})
	if err != nil {
		return Command{}, err
	}

	return c, nil
}

func (s *claudineService) GetCommand(ctx context.Context, channel string, trigger string) (Command, error) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	c := Command{
		Trigger: trigger,
	}

	err := s.db.View(func(tx *bolt.Tx) error {
		cBucket, err := GetActiveCommandBucket(tx, channel)
		if err != nil {
			return err
		}

		response := cBucket.Get([]byte(trigger))
		if response == nil {
			return ErrNotFound
		}

		c.Action = string(response)
		return nil
	})
	if err != nil {
		return Command{}, err
	}

	return c, nil
}

func (s *claudineService) ListCommand(ctx context.Context, channel string) ([]Command, error) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	var list []Command

	err := s.db.View(func(tx *bolt.Tx) error {
		cBucket, err := GetActiveCommandBucket(tx, channel)
		if err != nil {
			return err
		}

		err = cBucket.ForEach(func(trigger, action []byte) error {
			list = append(list, Command{
				Trigger: string(trigger),
				Action:  string(action),
			})
			return nil
		})
		return err
	})
	if err != nil {
		return []Command{}, err
	}

	return list, nil
}

func (s *claudineService) UpdateCommand(ctx context.Context, channel string, trigger string, action string) (Command, error) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	c := Command{
		Trigger: trigger,
	}

	err := s.db.Update(func(tx *bolt.Tx) error {
		cBucket, err := GetActiveCommandBucket(tx, channel)
		if err != nil {
			return err
		}

		response := cBucket.Get([]byte(trigger))
		if response == nil {
			return ErrNotFound
		}

		err = cBucket.Put([]byte(trigger), []byte(action))
		if err != nil {
			return ErrGeneric
		}

		c.Action = action
		return nil
	})
	if err != nil {
		return Command{}, err
	}

	return c, nil
}

func (s *claudineService) DeleteCommand(ctx context.Context, channel string, trigger string) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	err := s.db.Update(func(tx *bolt.Tx) error {
		cBucket, err := GetActiveCommandBucket(tx, channel)

		response := cBucket.Get([]byte(trigger))
		if response == nil {
			return ErrNotFound
		}

		err = cBucket.Delete([]byte(trigger))
		if err != nil {
			return ErrGeneric
		}

		return nil
	})

	return err
}

func (s *claudineService) NewRepeatCommand(ctx context.Context, channel string, trigger string, duration int) (RepeatCommand, error) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	err := s.db.Update(func(tx *bolt.Tx) error {
		rBucket, err := tx.CreateBucketIfNotExists([]byte("repeat"))
		if err != nil {
			return err
		}

		// TODO: Check if channel is active
		bucket, err := rBucket.CreateBucketIfNotExists([]byte(channel))
		if err != nil {
			return err
		}

		exist := bucket.Get([]byte(trigger))
		if exist != nil {
			return ErrAlreadyExist
		}

		bucket.Put([]byte(trigger), []byte(strconv.Itoa(duration)))
		return err
	})

	if err != nil {
		return RepeatCommand{}, err
	}

	command := RepeatCommand{
		Trigger: trigger,
		Duration:duration,
	}

	return command, nil
}

func (s *claudineService) GetRepeatCommand(ctx context.Context, channel string, trigger string) (RepeatCommand, error) {
	command := RepeatCommand{
		Trigger: trigger,
	}


	err := s.db.View(func(tx *bolt.Tx) error {
		rBucket := tx.Bucket([]byte("repeat"))
		if rBucket == nil {
			return ErrNotFound
		}

		cBucket := rBucket.Bucket([]byte(channel))
		if cBucket == nil {
			return ErrNotFound
		}

		duration := cBucket.Get([]byte(trigger))
		if duration == nil {
			return ErrNotFound
		}

		intDuration, err := strconv.Atoi(string(duration))
		if err != nil {
			return err
		}
		command.Duration = intDuration

		// TODO: Check if channel is active
		return err
	})
	if err != nil {
		return RepeatCommand{}, err
	}

	return command, err
}

func (s *claudineService) ListRepeatCommand(ctx context.Context, channel string) ([]RepeatCommand, error) {
	var list []RepeatCommand

	err := s.db.View(func(tx *bolt.Tx) error {
		rBucket := tx.Bucket([]byte("repeat"))
		if rBucket == nil {
			return ErrNotFound
		}

		cBucket := rBucket.Bucket([]byte(channel))
		if cBucket == nil {
			return ErrNotFound
		}


		err := cBucket.ForEach(func(trigger, duration []byte) error {
			intDuration, err := strconv.Atoi(string(duration))
			if err != nil {
				return err
			}
			list = append(list, RepeatCommand{
				Trigger: string(trigger),
				Duration:  intDuration,
			})
			return  nil
		})

		return err
	})
	if err != nil {
		return []RepeatCommand{}, err
	}

	return list, nil
}

func (s *claudineService) DeleteRepeatCommand(ctx context.Context, channel string, trigger string) error {
	err := s.db.Update(func(tx *bolt.Tx) error {
		rBucket := tx.Bucket([]byte("repeat"))
		if rBucket == nil {
			return ErrNotFound
		}

		cBucket := rBucket.Bucket([]byte(channel))
		if cBucket == nil {
			return ErrNotFound
		}

		err := cBucket.Delete([]byte(trigger))

		return err
	})
	return  err
}

func GetActiveCommandBucket(tx *bolt.Tx, channel string) (*bolt.Bucket, error) {
	// Get the channel bucket
	bucket := tx.Bucket([]byte(channel))
	if bucket == nil {
		return &bolt.Bucket{}, ErrNotFound
	}

	// Check if account is active
	active := bucket.Get([]byte("enabled"))
	if bytes.Compare(active, TRUE) != 0 {
		return &bolt.Bucket{}, ErrNotFound
	}

	// Get the commands bucket
	cBucket := bucket.Bucket([]byte("commands"))
	return cBucket, nil
}
