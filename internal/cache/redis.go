package cache

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

type Database struct {
	Client  *redis.Client
	context context.Context
}

var (
	ErrNil         = errors.New("No matching record found in redis database")
	defaultContext = context.TODO()
)

func NewRedisDatabase(options *redis.Options, ctx context.Context) (*Database, error) {
	client := redis.NewClient(options)

	if err := client.Ping(defaultContext).Err(); err != nil {
		return nil, err
	}

	return &Database{
		Client:  client,
		context: ctx,
	}, nil
}

func (d *Database) Ping() (string, error) {
	return d.Client.Ping(d.context).Result()
}

func (d *Database) Set(key string, value string, expiration time.Duration) (string, error) {
	return d.Client.Set(d.context, key, value, expiration).Result()
}
