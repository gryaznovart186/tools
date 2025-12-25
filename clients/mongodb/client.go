package mongodb

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	connectionTimeout = 5 * time.Second
	minPoolSize       = 10
	maxPoolSize       = 100
	idleTimeout       = 30 * time.Second
)

type Opt func(*options.ClientOptions)

type Client struct {
	cl *mongo.Client
	db *mongo.Database
}

func New(uri, dbName string, clientOpts ...Opt) (*Client, error) {
	o := options.Client().
		ApplyURI(uri).
		SetMinPoolSize(minPoolSize).
		SetMaxPoolSize(maxPoolSize).
		SetConnectTimeout(connectionTimeout).
		SetMaxConnIdleTime(idleTimeout)
	for _, opt := range clientOpts {
		opt(o)
	}

	client, err := mongo.Connect(context.Background(), o)
	if err != nil {
		return nil, fmt.Errorf("mongo connect error: %w", err)
	}

	return &Client{cl: client, db: client.Database(dbName)}, nil
}

func (c *Client) DB() *mongo.Database {
	return c.db
}

func (c *Client) Ping(ctx context.Context) error {
	return c.cl.Ping(ctx, nil)
}

func (c *Client) Disconnect(ctx context.Context) error {
	return c.cl.Disconnect(ctx)
}

func WithMinPoolSize(v uint64) Opt {
	return func(o *options.ClientOptions) {
		o.SetMinPoolSize(v)
	}
}

func WithMaxPoolSize(v uint64) Opt {
	return func(o *options.ClientOptions) {
		o.SetMaxPoolSize(v)
	}
}

func WithConnectTimeout(v time.Duration) Opt {
	return func(o *options.ClientOptions) {
		o.SetConnectTimeout(v)
	}
}

func WithMaxConnIdleTime(v time.Duration) Opt {
	return func(o *options.ClientOptions) {
		o.SetMaxConnIdleTime(v)
	}
}
