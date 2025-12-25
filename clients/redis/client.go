package redis

import (
	"github.com/redis/go-redis/v9"
)

type Opt func(*redis.Options)

type Client struct {
	rc *redis.Client
}

func NewClient(addr string, opts ...Opt) *Client {
	o := &redis.Options{Addr: addr}
	for _, opt := range opts {
		opt(o)
	}

	rc := redis.NewClient(o)
	return &Client{rc: rc}
}

func WithDB(db int) Opt {
	return func(o *redis.Options) {
		o.DB = db
	}
}

func WithPassword(password string) Opt {
	return func(o *redis.Options) {
		o.Password = password
	}
}

func WithPoolSize(poolSize int) Opt {
	return func(o *redis.Options) {
		o.PoolSize = poolSize
	}
}
