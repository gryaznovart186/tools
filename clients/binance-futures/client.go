package binance_futures

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/gryaznovart186/tools/clients/binance-futures/models"
)

const (
	baseURL     = "https://fapi.binance.com"
	testBaseURL = "https://demo-fapi.binance.com"
)

type credentials struct {
	apiKey    string
	apiSecret string
}

type Opt func(*Client)

type Client struct {
	rc    *resty.Client
	creds credentials

	mu         sync.RWMutex
	rateLimits models.RateLimits
	recvWindow int64
}

func New(apiKey, secretKey string, opts ...Opt) *Client {
	c := &Client{
		rc:         resty.New().SetBaseURL(baseURL).SetTimeout(time.Second * 5),
		recvWindow: 5000,
		creds: credentials{
			apiKey:    apiKey,
			apiSecret: secretKey,
		},
		rateLimits: models.RateLimits{},
	}

	c.rc.OnAfterResponse(func(client *resty.Client, resp *resty.Response) error {
		c.updateRateLimits(resp.Header())
		return nil
	})

	for _, opt := range opts {
		opt(c)
	}

	return c
}

func (c *Client) checkCredentials() error {
	if c.creds.apiKey == "" || c.creds.apiSecret == "" {
		return ErrEmptyCredentials
	}
	return nil
}

func (c *Client) updateRateLimits(headers http.Header) {
	c.mu.Lock()
	defer c.mu.Unlock()

	val := headers.Get("X-Mbx-Used-Weight-1m")
	if val != "" {
		if weight, err := strconv.Atoi(val); err == nil {
			c.rateLimits.UsedWeight1m = weight
		}
	}
}

func (c *Client) UsedRateLimits() models.RateLimits {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.rateLimits
}

func WithTestNet() Opt {
	return func(c *Client) {
		c.rc.SetBaseURL(testBaseURL)
	}
}

func WithCustomBaseURL(baseURL string) Opt {
	return func(c *Client) {
		c.rc.SetBaseURL(baseURL)
	}
}

func WithTimeout(timeout time.Duration) Opt {
	return func(c *Client) {
		c.rc.SetTimeout(timeout)
	}
}

func WithProxy(proxy string) Opt {
	return func(c *Client) {
		c.rc.SetProxy(proxy)
	}
}

func WithRecvWindow(recvWindow int64) Opt {
	return func(c *Client) {
		c.recvWindow = recvWindow
	}
}
