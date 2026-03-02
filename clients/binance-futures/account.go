package binance_futures

import (
	"context"
	"fmt"

	"github.com/gryaznovart186/tools/clients/binance-futures/models"
)

func (c *Client) AccountBalance(ctx context.Context) (*models.AccountBalance, error) {
	if err := c.checkCredentials(); err != nil {
		return nil, err
	}
	fullQuery := c.withSignature("")

	var balances []models.AccountBalance
	resp, err := c.rc.R().
		SetContext(ctx).
		SetHeader("X-MBX-APIKEY", c.creds.apiKey).
		SetResult(&balances).
		Get("/fapi/v3/balance?" + fullQuery)

	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, fmt.Errorf("api error, code: %d message: %s", resp.StatusCode(), resp.String())
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

	var openedPositions []models.Position
	resp, err := c.rc.R().
		SetContext(ctx).
		SetHeader("X-MBX-APIKEY", c.creds.apiKey).
		SetResult(&openedPositions).
		Get("/fapi/v3/positionRisk?" + fullQuery)

	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, fmt.Errorf("api error, code: %d message: %s", resp.StatusCode(), resp.String())
	}

	return openedPositions, nil
}
