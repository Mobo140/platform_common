package cache

import (
	"context"
	"time"

	"github.com/gomodule/redigo/redis"
)

type Client interface {
	Cache() Cache
	Close() error
}

type Cache interface {
	Set(ctx context.Context, key string, value interface{}) error
	HashSet(ctx context.Context, key string, value interface{}) error
	Get(ctx context.Context, key string) (interface{}, error)
	HGetAll(ctx context.Context, key string) ([]interface{}, error)
	Expire(ctx context.Context, key string, expiration time.Duration) error
	Ping(ctx context.Context) error
	Close()
}

type Handler func(context.Context, redis.Conn) error
