package binance_futures

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gryaznovart186/tools/clients/binance-futures/models"
)

func (c *Client) AccountBalance(ctx context.Context) (*models.AccountBalance, error) {
	if err := c.checkCredentials(); err != nil {
		return nil, err
	}
	fullQuery := c.withSignature("")

	resp, err := c.rc.R().
		SetContext(ctx).
		SetHeader("X-MBX-APIKEY", c.creds.apiKey).
		Get("/fapi/v3/balance?" + fullQuery)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode(), resp.Body())
	}
	var balances []models.AccountBalance
	err = json.Unmarshal(resp.Body(), &balances)
	if err != nil {
		return nil, ErrInvalidResponse
	}

	for _, b := range balances {
		if b.Asset == "USDT" {
			return &b, nil
		}
	}

	return nil, ErrUSDTBalanceNotFound
}

func (c *Client) OpenedPositions(ctx context.Context) ([]models.Position, error) {
	if err := c.checkCredentials(); err != nil {
		return nil, err
	}
	fullQuery := c.withSignature("")

	resp, err := c.rc.R().
		SetContext(ctx).
		SetHeader("X-MBX-APIKEY", c.creds.apiKey).
		Get("/fapi/v3/positionRisk?" + fullQuery)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode(), resp.Body())
	}

	var openedPositions []models.Position
	err = json.Unmarshal(resp.Body(), &openedPositions)
	if err != nil {
		return nil, ErrInvalidResponse
	}

	return openedPositions, nil
}
