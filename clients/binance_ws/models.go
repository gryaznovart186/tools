package binance_ws

import (
	"encoding/json"
)

type StreamData struct {
	Stream string          `json:"stream"`
	Data   json.RawMessage `json:"data"`
}
