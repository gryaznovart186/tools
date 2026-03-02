package binance_futures

import (
	"context"
	"fmt"

	"github.com/gryaznovart186/tools/clients/binance-futures/models"
)

func (c *Client) TopLongShortPositionRatio(ctx context.Context, symbol, period string, startTime, endTime int64, limit int) ([]models.LongShortRatio, error) {
	return c.fetchLongShortRatio(ctx, "/futures/data/topLongShortPositionRatio", symbol, period, startTime, endTime, limit)
}

func (c *Client) TopLongShortAccountRatio(ctx context.Context, symbol, period string, startTime, endTime int64, limit int) ([]models.LongShortRatio, error) {
	return c.fetchLongShortRatio(ctx, "/futures/data/topLongShortAccountRatio", symbol, period, startTime, endTime, limit)
}

func (c *Client) GlobalLongShortAccountRatio(ctx context.Context, symbol, period string, startTime, endTime int64, limit int) ([]models.LongShortRatio, error) {
	return c.fetchLongShortRatio(ctx, "/futures/data/globalLongShortAccountRatio", symbol, period, startTime, endTime, limit)
}

func (c *Client) TakerLongShortRatio(ctx context.Context, symbol, period string, startTime, endTime int64, limit int) ([]models.TakerLongShortRatio, error) {
	req := c.rc.R().
		SetContext(ctx).
		SetQueryParams(map[string]string{
			"symbol": symbol,
			"period": period,
		})

	if limit > 0 {
		req.SetQueryParam("limit", fmt.Sprintf("%d", limit))
	}
	if startTime > 0 {
		req.SetQueryParam("startTime", fmt.Sprintf("%d", startTime))
	}
	if endTime > 0 {
		req.SetQueryParam("endTime", fmt.Sprintf("%d", endTime))
	}

	var res []models.TakerLongShortRatio
	resp, err := req.SetResult(&res).Get("/futures/data/takerlongshortRatio")
	if err != nil {
		return nil, fmt.Errorf("failed to get taker long/short ratio: %w", err)
	}
	if resp.IsError() {
		return nil, fmt.Errorf("api error, code: %d message: %s", resp.StatusCode(), resp.String())
	}

	return res, nil
}

func (c *Client) fetchLongShortRatio(ctx context.Context, path, symbol, period string, startTime, endTime int64, limit int) ([]models.LongShortRatio, error) {
	req := c.rc.R().
		SetContext(ctx).
		SetQueryParams(map[string]string{
			"symbol": symbol,
			"period": period,
		})

	if limit > 0 {
		req.SetQueryParam("limit", fmt.Sprintf("%d", limit))
	}
	if startTime > 0 {
		req.SetQueryParam("startTime", fmt.Sprintf("%d", startTime))
	}
	if endTime > 0 {
		req.SetQueryParam("endTime", fmt.Sprintf("%d", endTime))
	}

	var res []models.LongShortRatio
	resp, err := req.SetResult(&res).Get(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get %s: %w", path, err)
	}
	if resp.IsError() {
		return nil, fmt.Errorf("api error, code: %d message: %s", resp.StatusCode(), resp.String())
	}

	return res, nil
}
