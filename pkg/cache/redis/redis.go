package redis

import (
	"context"
	"log"
	"time"

	"github.com/Mobo140/microservices/auth/internal/client/cache"
	"github.com/Mobo140/microservices/auth/internal/config"
	"github.com/gomodule/redigo/redis"
)

var _ cache.Cache = (*rd)(nil)

type rd struct {
	pool   *redis.Pool
	config config.RedisConfig
}

func NewRD(pool *redis.Pool, config config.RedisConfig) *rd {
	return &rd{
		pool:   pool,
		config: config,
	}
}

func (c *rd) Set(ctx context.Context, key string, value interface{}) error {
	err := c.execute(ctx, func(ctx context.Context, conn redis.Conn) error {
		var errEx error
		_, errEx = conn.Do("SET", key, value)
		if errEx != nil {
			return errEx
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func (c *rd) HashSet(ctx context.Context, key string, values interface{}) error {
	err := c.execute(ctx, func(ctx context.Context, conn redis.Conn) error {
		var errEx error
		_, errEx = conn.Do("HSET", redis.Args{key}.AddFlat(values)...)
		if errEx != nil {
			return errEx
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func (c *rd) Get(ctx context.Context, key string) (interface{}, error) {
	var value interface{}

	err := c.execute(ctx, func(ctx context.Context, conn redis.Conn) error {
		var errEx error
		value, errEx = conn.Do("GET", key)
		if errEx != nil {
			return errEx
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return value, nil
}

func (c *rd) HGetAll(ctx context.Context, key string) ([]interface{}, error) {
	var values []interface{}
	err := c.execute(ctx, func(ctx context.Context, conn redis.Conn) error {
		var errEx error
		values, errEx = redis.Values((conn.Do("HGETALL", key)))
		if errEx != nil {
			return errEx
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return values, nil
}

func (c *rd) Expire(ctx context.Context, key string, expiration time.Duration) error {
	err := c.execute(ctx, func(ctx context.Context, conn redis.Conn) error {
		var errEx error
		_, errEx = conn.Do("EXPIRE", key, int(expiration.Seconds()))
		if errEx != nil {
			return errEx
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func (c *rd) Ping(ctx context.Context) error {
	err := c.execute(ctx, func(ctx context.Context, conn redis.Conn) error {
		var errEx error
		_, errEx = conn.Do("PING")
		if errEx != nil {
			return errEx
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func (c *rd) execute(ctx context.Context, fn cache.Handler) error {
	conn, err := c.getConnect(ctx)
	if err != nil {
		return err
	}
	defer func() {
		err = conn.Close()
		if err != nil {
			log.Printf("failed to close redis connection: %v\n", err)
		}
	}()

	err = fn(ctx, conn)
	if err != nil {
		return err
	}

	return nil
}

func (c *rd) getConnect(ctx context.Context) (redis.Conn, error) {
	getConnTimeoutCtx, cancel := context.WithTimeout(ctx, c.config.ConnectionTimeout())
	defer cancel()

	conn, err := c.pool.GetContext(getConnTimeoutCtx)
	if err != nil {
		log.Printf("failed to get redis connection: %v\n", err)

		_ = conn.Close()

		return nil, err
	}

	return conn, nil
}

func (r *rd) Close() {
	r.pool.Close()
}
