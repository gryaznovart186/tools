package binance_ws

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

const (
	wsBaseEndpoint     = "wss://fstream.binance.com/market/stream"
	testWsBaseEndpoint = "wss://stream.binancefuture.com/market/stream"

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

	// message channel for ReadLoop
	dataCh chan *StreamData
	wg     sync.WaitGroup
}

type connection struct {
	client *BinanceWebsocketClient
	conn   *websocket.Conn
	ctx    context.Context
	cancel context.CancelFunc

	writeCh chan writeRequest
	pumpWg  sync.WaitGroup // tracks writePump and pingPump for this connection
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

func (c *BinanceWebsocketClient) newConnection() *connection {
	return &connection{
		client:  c,
		streams: make(map[string]struct{}),
	}
}

func (conn *connection) ensureConnected() error {
	conn.mu.Lock()

	if conn.conn != nil {
		conn.mu.Unlock()
		return nil
	}

	// If there's an old cancel, cancel it and wait for old pumps to finish
	if conn.cancel != nil {
		conn.cancel()
		conn.cancel = nil
		pumpWg := &conn.pumpWg
		conn.mu.Unlock()
		pumpWg.Wait()
		conn.mu.Lock()
	}

	conn.ctx, conn.cancel = context.WithCancel(conn.client.ctx)

	c, _, err := conn.client.dialer.DialContext(conn.ctx, conn.client.endpoint, nil)
	if err != nil {
		conn.cancel()
		conn.cancel = nil
		conn.ctx = nil
		conn.mu.Unlock()
		return err
	}

	c.SetReadDeadline(time.Now().Add(conn.client.pongTimeout))
	c.SetPongHandler(func(string) error {
		c.SetReadDeadline(time.Now().Add(conn.client.pongTimeout))
		return nil
	})

	conn.conn = c
	conn.writeCh = make(chan writeRequest, 100)

	conn.pumpWg.Add(2)
	go conn.writePump()
	go conn.pingPump()

	// Collect streams to restore while still holding the lock
	var streamsToRestore []string
	if len(conn.streams) > 0 {
		streamsToRestore = make([]string, 0, len(conn.streams))
		for s := range conn.streams {
			streamsToRestore = append(streamsToRestore, s)
		}
	}

	conn.mu.Unlock()

	// Restore subscriptions outside the lock
	if len(streamsToRestore) > 0 {
		if err := conn.send("SUBSCRIBE", streamsToRestore); err != nil {
			conn.closeConn()
			return err
		}
	}

	return nil
}

// closeConn closes the underlying websocket connection and waits for pump goroutines.
// Safe to call concurrently.
func (conn *connection) closeConn() {
	conn.mu.Lock()

	if conn.cancel != nil {
		conn.cancel()
		conn.cancel = nil
	}

	// Don't close writeCh — writePump exits via ctx.Done().
	// Just nil it out so send() returns ErrNotConnected.
	conn.writeCh = nil

	if conn.conn != nil {
		_ = conn.conn.Close()
		conn.conn = nil
	}

	pumpWg := &conn.pumpWg
	conn.mu.Unlock()

	// Wait for writePump and pingPump to finish outside the lock
	pumpWg.Wait()
}

func (conn *connection) reconnect() {
	conn.closeConn()
}

func (conn *connection) writePump() {
	defer conn.pumpWg.Done()

	conn.mu.Lock()
	writeCh := conn.writeCh
	ctx := conn.ctx
	conn.mu.Unlock()

	if writeCh == nil || ctx == nil {
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		case req := <-writeCh:
			conn.mu.Lock()
			c := conn.conn
			conn.mu.Unlock()

			if c == nil {
				select {
				case req.result <- ErrNotConnected:
				default:
				}
				continue
			}

			c.SetWriteDeadline(time.Now().Add(10 * time.Second))
			err := c.WriteMessage(websocket.TextMessage, req.data)
			select {
			case req.result <- err:
			default:
			}
		}
	}
}

func (conn *connection) pingPump() {
	defer conn.pumpWg.Done()

	conn.mu.Lock()
	ctx := conn.ctx
	conn.mu.Unlock()

	if ctx == nil {
		return
	}

	ticker := time.NewTicker(conn.client.pingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			conn.mu.Lock()
			c := conn.conn
			conn.mu.Unlock()

			if c == nil {
				return
			}

			if err := c.WriteControl(websocket.PingMessage, nil, time.Now().Add(10*time.Second)); err != nil {
				// Don't call reconnect()/closeConn() from pingPump —
				// it would deadlock on pumpWg.Wait() (waiting for itself).
				// Just close the underlying net.Conn to unblock ReadMessage() in run().
				// run() will handle the reconnection.
				conn.mu.Lock()
				if conn.conn != nil {
					_ = conn.conn.Close()
					conn.conn = nil
				}
				conn.mu.Unlock()
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

	result := make(chan error, 1)
	req := writeRequest{data: data, result: result}

	conn.mu.Lock()
	if conn.writeCh == nil {
		conn.mu.Unlock()
		return ErrNotConnected
	}

	ctx := conn.ctx

	// Try non-blocking write under mutex (buffer is 100, usually succeeds)
	select {
	case conn.writeCh <- req:
		conn.mu.Unlock()
	default:
		// Buffer full — release mutex and wait
		writeCh := conn.writeCh
		conn.mu.Unlock()

		select {
		case writeCh <- req:
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	// Wait for result outside the lock
	select {
	case err := <-result:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

var errConnClosed = errors.New("websocket connection closed")

// readMessage safely reads from the websocket connection, recovering from
// panics caused by reading a connection that was closed by another goroutine.
func (conn *connection) readMessage() (msg []byte, err error) {
	conn.mu.Lock()
	c := conn.conn
	conn.mu.Unlock()

	if c == nil {
		return nil, errConnClosed
	}

	defer func() {
		if r := recover(); r != nil {
			msg = nil
			err = fmt.Errorf("websocket read panic: %v", r)
		}
	}()

	_, msg, err = c.ReadMessage()
	return msg, err
}

func (conn *connection) run() {
	for {
		select {
		case <-conn.client.ctx.Done():
			return
		default:
		}

		if err := conn.ensureConnected(); err != nil {
			select {
			case <-conn.client.ctx.Done():
				return
			case <-time.After(time.Second):
				continue
			}
		}

		msg, err := conn.readMessage()
		if err != nil {
			conn.closeConn()
			select {
			case <-conn.client.ctx.Done():
				return
			case <-time.After(time.Second):
				continue
			}
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
	if c.dataCh == nil {
		c.dataCh = make(chan *StreamData, 1000)
	}

	conn := c.newConnection()
	c.conns = append(c.conns, conn)

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		conn.run()
	}()

	if err := conn.ensureConnected(); err != nil {
		return err
	}

	return nil
}

func (c *BinanceWebsocketClient) Close() error {
	c.mu.Lock()
	if c.cancel != nil {
		c.cancel()
		c.cancel = nil
	}

	// Copy conns slice under lock, then release lock before calling closeConn
	// to avoid deadlock (closeConn takes conn.mu, and run() goroutines may
	// hold conn.mu while trying to send to dataCh which needs client context).
	conns := make([]*connection, len(c.conns))
	copy(conns, c.conns)
	c.mu.Unlock()

	// Close each connection and wait for their pump goroutines
	for _, conn := range conns {
		conn.closeConn()
	}

	// Wait for all run() goroutines
	c.wg.Wait()

	c.mu.Lock()
	c.ctx = nil
	c.conns = nil
	c.streams = make(map[string]*connection)

	if c.dataCh != nil {
		close(c.dataCh)
		c.dataCh = nil
	}
	c.mu.Unlock()

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

			c.wg.Add(1)
			go func() {
				defer c.wg.Done()
				targetConn.run()
			}()

			if err := targetConn.ensureConnected(); err != nil {
				c.mu.Unlock()
				return err
			}
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
	c.mu.RLock()
	clientCtx := c.ctx
	dataCh := c.dataCh
	c.mu.RUnlock()

	if clientCtx == nil {
		return ErrNotConnected
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-clientCtx.Done():
			return clientCtx.Err()
		case data, ok := <-dataCh:
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
