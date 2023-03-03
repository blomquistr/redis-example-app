package cache

import (
	"context"
	"errors"

	"github.com/redis/go-redis/v9"
)

type Database struct {
	Client *redis.Client
}

var (
	ErrNil         = errors.New("No matching record found in redis database")
	defaultContext = context.TODO()
)

func NewRedisDatabase(options *redis.Options) (*Database, error) {
	client := redis.NewClient(options)

	if err := client.Ping(defaultContext).Err(); err != nil {
		return nil, err
	}

	return &Database{
		Client: client,
	}, nil
}

func (d *Database) Set(key string, value string) error {

}
