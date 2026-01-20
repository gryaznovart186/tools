package binance_futures

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

// Sign вычисляет HMAC SHA256 подпись для данных с использованием secretKey.
func sign(data, secretKey string) string {
	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

func (c *Client) sign(data string) string {
	return sign(data, c.creds.apiSecret)
}

func (c *Client) withSignature(query string) string {
	timestamp := time.Now().UnixMilli()
	if query != "" {
		query += "&"
	}
	query += fmt.Sprintf("timestamp=%d", timestamp)

	if c.recvWindow > 0 {
		query += fmt.Sprintf("&recvWindow=%d", c.recvWindow)
	}

	signature := c.sign(query)
	return fmt.Sprintf("%s&signature=%s", query, signature)
}
