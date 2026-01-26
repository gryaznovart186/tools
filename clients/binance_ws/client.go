package binance_ws

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
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
	conns  []*connection
	ctx    context.Context // Global context for all connections
	cancel context.CancelFunc

	// global stream map to find which connection serves which stream
	streams map[string]*connection
	reqID   int64

	// ping/pong
	pingInterval time.Duration
	pongTimeout  time.Duration

	// limits
	maxSubscriptions int

	// state
	configured bool

	// message channel for ReadLoop
	dataCh chan *StreamData
}

type connection struct {
	client *BinanceWebsocketClient
	conn   *websocket.Conn
	ctx    context.Context
	cancel context.CancelFunc

	writeCh chan writeRequest
	streams map[string]struct{}
	mu      sync.Mutex
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
		streams:          make(map[string]*connection),
		dataCh:           make(chan *StreamData, 1000),
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

func (c *BinanceWebsocketClient) newConnection() *connection {
	return &connection{
		client:  c,
		streams: make(map[string]struct{}),
	}
}

func (conn *connection) ensureConnected() error {
	conn.mu.Lock()
	defer conn.mu.Unlock()

	if conn.conn != nil {
		return nil
	}

	conn.ctx, conn.cancel = context.WithCancel(conn.client.ctx)

	c, _, err := conn.client.dialer.DialContext(conn.ctx, conn.client.endpoint, nil)
	if err != nil {
		return err
	}

	c.SetReadDeadline(time.Now().Add(conn.client.pongTimeout))
	c.SetPongHandler(func(string) error {
		c.SetReadDeadline(time.Now().Add(conn.client.pongTimeout))
		return nil
	})

	conn.conn = c
	conn.writeCh = make(chan writeRequest, 100)

	go conn.writePump()
	go conn.pingPump()

	// Restore subscriptions
	if len(conn.streams) > 0 {
		all := make([]string, 0, len(conn.streams))
		for s := range conn.streams {
			all = append(all, s)
		}
		// Send subscribe command
		go func() {
			_ = conn.send("SUBSCRIBE", all)
		}()
	}

	return nil
}

func (conn *connection) reconnect() {
	conn.mu.Lock()
	defer conn.mu.Unlock()

	if conn.cancel != nil {
		conn.cancel()
		conn.cancel = nil
	}

	if conn.writeCh != nil {
		close(conn.writeCh)
		conn.writeCh = nil
	}

	if conn.conn != nil {
		_ = conn.conn.Close()
		conn.conn = nil
	}
}

func (conn *connection) writePump() {
	for {
		select {
		case <-conn.ctx.Done():
			return
		case req, ok := <-conn.writeCh:
			if !ok {
				return
			}

			conn.mu.Lock()
			c := conn.conn
			conn.mu.Unlock()

			if c == nil {
				req.result <- ErrNotConnected
				continue
			}

			err := c.WriteMessage(websocket.TextMessage, req.data)
			req.result <- err
		}
	}
}

func (conn *connection) pingPump() {
	ticker := time.NewTicker(conn.client.pingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-conn.ctx.Done():
			return
		case <-ticker.C:
			conn.mu.Lock()
			c := conn.conn
			conn.mu.Unlock()

			if c == nil {
				return
			}

			if err := c.WriteControl(websocket.PingMessage, nil, time.Now().Add(10*time.Second)); err != nil {
				conn.reconnect()
				return
			}
		}
	}
}

func (conn *connection) send(method string, streams []string) error {
	id := atomic.AddInt64(&conn.client.reqID, 1)

	msg := map[string]any{
		"method": method,
		"params": streams,
		"id":     id,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	conn.mu.Lock()
	writeCh := conn.writeCh
	ctx := conn.ctx
	conn.mu.Unlock()

	if writeCh == nil {
		return ErrNotConnected
	}

	result := make(chan error, 1)
	select {
	case writeCh <- writeRequest{data: data, result: result}:
		return <-result
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (conn *connection) run() {
	for {
		select {
		case <-conn.client.ctx.Done():
			return
		default:
		}

		if err := conn.ensureConnected(); err != nil {
			time.Sleep(time.Second)
			continue
		}

		conn.mu.Lock()
		c := conn.conn
		conn.mu.Unlock()

		if c == nil {
			time.Sleep(time.Second)
			continue
		}

		_, msg, err := c.ReadMessage()
		if err != nil {
			conn.reconnect()
			time.Sleep(time.Second)
			continue
		}

		var streamMsg StreamData
		if err := json.Unmarshal(msg, &streamMsg); err != nil {
			continue
		}

		if streamMsg.Stream == "" {
			continue
		}

		select {
		case conn.client.dataCh <- &streamMsg:
		case <-conn.client.ctx.Done():
			return
		}
	}
}

func (c *BinanceWebsocketClient) Connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.ctx != nil {
		return ErrAlreadyConnected
	}

	c.ctx, c.cancel = context.WithCancel(ctx)
	c.configured = true

	conn := c.newConnection()
	c.conns = append(c.conns, conn)

	if err := conn.ensureConnected(); err != nil {
		return err
	}

	go conn.run()

	return nil
}

func (c *BinanceWebsocketClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.cancel != nil {
		c.cancel()
		c.cancel = nil
	}

	for _, conn := range c.conns {
		conn.reconnect()
	}
	c.conns = nil
	c.streams = make(map[string]*connection)

	return nil
}

func (c *BinanceWebsocketClient) Subscribe(streams ...string) error {
	c.mu.Lock()
	if c.ctx == nil {
		c.mu.Unlock()
		return ErrNotConnected
	}

	groups := make(map[*connection][]string)

	for _, s := range streams {
		s = strings.TrimSpace(s)
		if s == "" {
			c.mu.Unlock()
			return ErrEmptyStream
		}

		if _, ok := c.streams[s]; ok {
			continue
		}

		// Find connection with space
		var targetConn *connection
		for _, conn := range c.conns {
			conn.mu.Lock()
			count := len(conn.streams)
			conn.mu.Unlock()

			if count < c.maxSubscriptions {
				targetConn = conn
				break
			}
		}

		if targetConn == nil {
			targetConn = c.newConnection()
			c.conns = append(c.conns, targetConn)
			if err := targetConn.ensureConnected(); err != nil {
				c.mu.Unlock()
				return err
			}
			go targetConn.run()
		}

		targetConn.mu.Lock()
		targetConn.streams[s] = struct{}{}
		targetConn.mu.Unlock()

		c.streams[s] = targetConn
		groups[targetConn] = append(groups[targetConn], s)
	}
	c.mu.Unlock()

	// Send SUBSCRIBE messages outside of client lock
	for conn, sList := range groups {
		if err := conn.send("SUBSCRIBE", sList); err != nil {
			return err
		}
	}

	return nil
}

func (c *BinanceWebsocketClient) Unsubscribe(streams ...string) error {
	c.mu.Lock()
	if c.ctx == nil {
		c.mu.Unlock()
		return ErrNotConnected
	}

	// Group streams by connection
	groups := make(map[*connection][]string)

	for _, s := range streams {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}

		conn, ok := c.streams[s]
		if !ok {
			continue
		}

		groups[conn] = append(groups[conn], s)
		delete(c.streams, s)

		conn.mu.Lock()
		delete(conn.streams, s)
		conn.mu.Unlock()
	}
	c.mu.Unlock()

	for conn, s := range groups {
		if err := conn.send("UNSUBSCRIBE", s); err != nil {
			return err
		}
	}

	return nil
}

func (c *BinanceWebsocketClient) ReadLoop(ctx context.Context, handler func(ctx context.Context, data *StreamData)) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-c.ctx.Done():
			return c.ctx.Err()
		case data, ok := <-c.dataCh:
			if !ok {
				return nil
			}
			handler(ctx, data)
		}
	}
}

func (c *BinanceWebsocketClient) SubscriptionCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.streams)
}

func (c *BinanceWebsocketClient) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if len(c.conns) == 0 {
		return false
	}
	for _, conn := range c.conns {
		conn.mu.Lock()
		connected := conn.conn != nil
		conn.mu.Unlock()
		if !connected {
			return false
		}
	}
	return true
}
