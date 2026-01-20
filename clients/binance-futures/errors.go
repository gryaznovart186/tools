package binance_futures

import (
	"errors"
)

var (
	ErrInvalidResponse     = errors.New("invalid response from Binance Futures API")
	ErrUSDTBalanceNotFound = errors.New("USDT balance not found")
	ErrEmptyCredentials    = errors.New("empty credentials")
)
