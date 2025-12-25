package telegram

import (
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
)

const (
	telegramAPIURL = "https://api.telegram.org"

	// timeout defines the maximum duration allowed for HTTP requests before they are canceled or timed out.
	timeout = 5 * time.Second
)

// Client represents a Telegram client configured with a token for authenticated REST API operations.
type Client struct {
	restyClient *resty.Client
	token       string
	chatID      string
	timeout     time.Duration
}

type Opt func(*Client)

// NewClient creates and returns a new Telegram client using the provided token, chat ID, and optional configurations.
func NewClient(token, chatID string, opts ...Opt) *Client {
	c := &Client{restyClient: resty.New(), token: token, chatID: chatID, timeout: timeout}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// WithTimeout sets a timeout duration for the client's HTTP requests.
func WithTimeout(timeout time.Duration) Opt {
	return func(c *Client) {
		c.timeout = timeout
	}
}

func (c *Client) botUrl() string {
	return fmt.Sprintf("%s/bot%s", telegramAPIURL, c.token)
}

// SendText sends a text message to the chat associated with the client and returns an error if the operation fails.
func (c *Client) SendText(msg string) error {
	params := map[string]string{
		"chat_id": c.chatID,
		"text":    msg,
	}

	reps, err := c.restyClient.
		SetTimeout(c.timeout).
		R().
		SetQueryParams(params).
		Post(c.botUrl() + "/sendMessage")
	if err != nil {
		return err
	}
	if reps.IsError() {
		return fmt.Errorf("api error, code: %d message: %s", reps.StatusCode(), reps.String())
	}

	return nil
}
