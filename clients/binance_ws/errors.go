package binance_ws

import (
	"errors"
)

var (
	ErrNotConnected            = errors.New("websocket not connected")
	ErrEmptyStream             = errors.New("stream name cannot be empty")
	ErrMaxSubscriptions        = errors.New("maximum subscriptions limit reached")
	ErrAlreadyConnected        = errors.New("already connected")
	ErrClientAlreadyConfigured = errors.New("client options cannot be changed after Connect")
)
