package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"k8s.io/klog"
)

type Database struct {
	Client  *redis.Client
	Context *context.Context
}

var (
	ErrNil         = errors.New("No matching record found in redis database")
	defaultContext = context.TODO()
)

func NewRedisDatabase(options *redis.Options, ctx *context.Context) (*Database, error) {
	if ctx == nil {
		ctx = &defaultContext
	}
	client := redis.NewClient(options)

	if err := client.Ping(defaultContext).Err(); err != nil {
		return nil, err
	}

	return &Database{
		Client:  client,
		Context: ctx,
	}, nil
}

func (d *Database) Ping() (string, error) {
	klog.Info("Pinging database...")
	return d.Client.Ping(*d.Context).Result()
}

func (d *Database) Set(key string, value string, expiration int) (string, error) {
	klog.Info(fmt.Sprintf("Writing key [%s] with value [%s] and TTL of [%v] seconds to Redis cache...", key, value, time.Duration(expiration)*time.Second))
	return d.Client.Set(*d.Context, key, value, time.Duration(expiration)*time.Second).Result()
}

func (d *Database) Get(key string) (string, error) {
	klog.Info(fmt.Sprintf("Fetching key [%s] from the Redis cache...", key))
	return d.Client.Get(*d.Context, key).Result()
}
