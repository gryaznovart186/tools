package binance_ws

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	wsBaseEndpoint     = "wss://fstream.binance.com/stream"
	testWsBaseEndpoint = "wss://stream.binancefuture.com/stream"

	defaultHandshakeTimeout = 10 * time.Second
	defaultReadBufferSize   = 8192
	defaultWriteBufferSize  = 8192
	defaultPingInterval     = 30 * time.Second
	defaultPongTimeout      = 60 * time.Second

	defaultMaxSubscriptions = 200 // Binance limit
)

// Option is a functional option for configuring BinanceWebsocketClient
type Option func(*BinanceWebsocketClient)

// WithTestnet configures the client to use testnet endpoint
func WithTestnet() Option {
	return func(c *BinanceWebsocketClient) {
		c.endpoint = testWsBaseEndpoint
	}
}

// WithEndpoint sets a custom endpoint
func WithEndpoint(endpoint string) Option {
	return func(c *BinanceWebsocketClient) {
		c.endpoint = endpoint
	}
}

// WithHandshakeTimeout sets the handshake timeout
func WithHandshakeTimeout(timeout time.Duration) Option {
	return func(c *BinanceWebsocketClient) {
		c.dialer.HandshakeTimeout = timeout
	}
}

// WithPingInterval sets the ping interval
func WithPingInterval(interval time.Duration) Option {
	return func(c *BinanceWebsocketClient) {
		c.pingInterval = interval
	}
}

// WithMaxSubscriptions sets the maximum number of subscriptions
func WithMaxSubscriptions(max int) Option {
	return func(c *BinanceWebsocketClient) {
		c.maxSubscriptions = max
	}
}

type BinanceWebsocketClient struct {
	endpoint string
	dialer   *websocket.Dialer

	mu     sync.RWMutex
	conn   *websocket.Conn
	ctx    context.Context
	cancel context.CancelFunc

	// write operations channel for serialization
	writeCh chan writeRequest

	// active subscriptions (set)
	streams map[string]struct{}
	reqID   int

	// ping/pong
	pingInterval time.Duration
	pongTimeout  time.Duration

	// limits
	maxSubscriptions int

	// state
	configured bool
}

type writeRequest struct {
	data   []byte
	result chan error
}

func NewClient(opts ...Option) *BinanceWebsocketClient {
	c := &BinanceWebsocketClient{
		endpoint:         wsBaseEndpoint,
		pingInterval:     defaultPingInterval,
		pongTimeout:      defaultPongTimeout,
		maxSubscriptions: defaultMaxSubscriptions,
		streams:          make(map[string]struct{}),
		dialer: &websocket.Dialer{
			Proxy:            http.ProxyFromEnvironment,
			HandshakeTimeout: defaultHandshakeTimeout,
			ReadBufferSize:   defaultReadBufferSize,
			WriteBufferSize:  defaultWriteBufferSize,
		},
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// WithTestnet configures the client to connect to the Binance Testnet by setting the WebSocket endpoint.
func (c *BinanceWebsocketClient) WithTestnet() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.configured {
		return ErrClientNotConfigured
	}

	c.endpoint = testWsBaseEndpoint
	return nil
}

func (c *BinanceWebsocketClient) Connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {
		return ErrAlreadyConnected
	}

	return c.connectLocked(ctx)
}

// connectLocked establishes connection, must be called with mu locked
func (c *BinanceWebsocketClient) connectLocked(ctx context.Context) error {
	// Cancel previous context if exists
	if c.cancel != nil {
		c.cancel()
	}

	c.ctx, c.cancel = context.WithCancel(ctx)
	c.configured = true

	conn, _, err := c.dialer.DialContext(c.ctx, c.endpoint, nil)
	if err != nil {
		return err
	}

	conn.SetReadDeadline(time.Now().Add(c.pongTimeout))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(c.pongTimeout))
		return nil
	})

	c.conn = conn
	c.writeCh = make(chan writeRequest, 100)

	// Start write pump
	go c.writePump()

	// Start ping pump
	go c.pingPump()

	// Restore subscriptions
	if len(c.streams) > 0 {
		all := make([]string, 0, len(c.streams))
		for s := range c.streams {
			all = append(all, s)
		}
		return c.sendLocked("SUBSCRIBE", all)
	}

	return nil
}

// writePump handles all write operations to avoid concurrent writes
func (c *BinanceWebsocketClient) writePump() {
	for {
		select {
		case <-c.ctx.Done():
			return
		case req, ok := <-c.writeCh:
			if !ok {
				return
			}

			c.mu.RLock()
			conn := c.conn
			c.mu.RUnlock()

			if conn == nil {
				req.result <- ErrNotConnected
				continue
			}

			err := conn.WriteMessage(websocket.TextMessage, req.data)
			req.result <- err
		}
	}
}

// pingPump sends periodic ping messages
func (c *BinanceWebsocketClient) pingPump() {
	ticker := time.NewTicker(c.pingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			c.mu.RLock()
			conn := c.conn
			c.mu.RUnlock()

			if conn == nil {
				return
			}

			if err := conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(10*time.Second)); err != nil {
				// Ping failed, close connection to trigger reconnection
				c.mu.Lock()
				if c.conn != nil {
					_ = c.closeLocked()
				}
				c.mu.Unlock()
				return
			}
		}
	}
}

func (c *BinanceWebsocketClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.closeLocked()
}

// closeLocked closes connection, must be called with mu locked
func (c *BinanceWebsocketClient) closeLocked() error {
	if c.cancel != nil {
		c.cancel()
		c.cancel = nil
	}

	if c.writeCh != nil {
		close(c.writeCh)
		c.writeCh = nil
	}

	if c.conn == nil {
		return nil
	}

	_ = c.conn.WriteControl(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
		time.Now().Add(time.Second),
	)

	err := c.conn.Close()
	c.conn = nil
	return err
}

func (c *BinanceWebsocketClient) Subscribe(streams ...string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn == nil {
		return ErrNotConnected
	}

	var toSubscribe []string

	for _, s := range streams {
		// Validate stream name
		s = strings.TrimSpace(s)
		if s == "" {
			return ErrEmptyStream
		}

		if _, ok := c.streams[s]; ok {
			continue
		}

		// Check subscription limit
		if len(c.streams) >= c.maxSubscriptions {
			return ErrMaxSubscriptions
		}

		c.streams[s] = struct{}{}
		toSubscribe = append(toSubscribe, s)
	}

	if len(toSubscribe) == 0 {
		return nil
	}

	return c.sendLocked("SUBSCRIBE", toSubscribe)
}

func (c *BinanceWebsocketClient) Unsubscribe(streams ...string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn == nil {
		return ErrNotConnected
	}

	var toUnsubscribe []string

	for _, s := range streams {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}

		if _, ok := c.streams[s]; !ok {
			continue
		}
		delete(c.streams, s)
		toUnsubscribe = append(toUnsubscribe, s)
	}

	if len(toUnsubscribe) == 0 {
		return nil
	}

	return c.sendLocked("UNSUBSCRIBE", toUnsubscribe)
}

// sendLocked prepares and sends a message, must be called with mu locked
// It unlocks the mutex before sending to avoid deadlock with writePump
func (c *BinanceWebsocketClient) sendLocked(method string, streams []string) error {
	c.reqID++

	msg := map[string]any{
		"method": method,
		"params": streams,
		"id":     c.reqID,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	// Use write channel to avoid concurrent writes
	if c.writeCh == nil {
		return ErrNotConnected
	}

	writeCh := c.writeCh
	ctx := c.ctx

	// Unlock before sending to avoid deadlock with writePump
	c.mu.Unlock()
	defer c.mu.Lock()

	result := make(chan error, 1)
	select {
	case writeCh <- writeRequest{data: data, result: result}:
		return <-result
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (c *BinanceWebsocketClient) read() ([]byte, error) {
	c.mu.RLock()
	conn := c.conn
	c.mu.RUnlock()

	if conn == nil {
		return nil, ErrNotConnected
	}

	_, msg, err := conn.ReadMessage()
	return msg, err
}

// ReadLoop reads messages from the WebSocket connection and passes them to the handler.
// It returns an error when the connection is closed or an error occurs.
// Reconnection logic should be handled by the caller.
func (c *BinanceWebsocketClient) ReadLoop(ctx context.Context, handler func(ctx context.Context, data *StreamData)) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		msg, err := c.read()
		if err != nil {
			return err
		}
		var streamMsg StreamData
		if err := json.Unmarshal(msg, &streamMsg); err != nil {
			return err
		}

		// Ignore subscribe/unsubcribe confirmation messages
		if streamMsg.Stream == "" {
			continue
		}

		handler(ctx, &streamMsg)
	}
}

// SubscriptionCount returns the current number of active subscriptions
func (c *BinanceWebsocketClient) SubscriptionCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.streams)
}

// IsConnected returns true if the client is currently connected
func (c *BinanceWebsocketClient) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.conn != nil
}
