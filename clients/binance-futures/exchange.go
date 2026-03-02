package binance_futures

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/gryaznovart186/tools/clients/binance-futures/models"
)

func (c *Client) ExchangeSymbolsInfo(ctx context.Context, isOnlyInTrading bool) ([]models.SymbolInfo, error) {
	var res models.ExchangeInfo
	resp, err := c.rc.R().
		SetContext(ctx).
		SetResult(&res).
		Get("/fapi/v1/exchangeInfo")
	if err != nil {
		return nil, fmt.Errorf("failed to get exchange info: %w", err)
	}
	if resp.IsError() {
		return nil, fmt.Errorf("api error, code: %d message: %s", resp.StatusCode(), resp.String())
	}

	if isOnlyInTrading {
		var filteredSymbols []models.SymbolInfo
		for _, symbol := range res.Symbols {
			if symbol.Status == "TRADING" {
				filteredSymbols = append(filteredSymbols, symbol)
			}
		}
		res.Symbols = filteredSymbols
	}

	return res.Symbols, nil
}

func (c *Client) PriceChangeStats(ctx context.Context) ([]models.PriceChangeStats, error) {
	var res []models.PriceChangeStats
	resp, err := c.rc.R().
		SetContext(ctx).
		SetResult(&res).
		Get("/fapi/v1/ticker/24hr")
	if err != nil {
		return nil, fmt.Errorf("failed to get price change stats: %w", err)
	}
	if resp.IsError() {
		return nil, fmt.Errorf("api error, code: %d message: %s", resp.StatusCode(), resp.String())
	}

	return res, nil
}

func (c *Client) OpenInterest(ctx context.Context, symbol string) (models.OpenInterest, error) {
	var res models.OpenInterest
	resp, err := c.rc.R().
		SetContext(ctx).
		SetQueryParam("symbol", symbol).
		SetResult(&res).
		Get("/fapi/v1/openInterest")
	if err != nil {
		return models.OpenInterest{}, fmt.Errorf("failed to get open interest: %w", err)
	}
	if resp.IsError() {
		return models.OpenInterest{}, fmt.Errorf("api error, code: %d message: %s", resp.StatusCode(), resp.String())
	}

	return res, nil
}

func (c *Client) OpenInterestHist(ctx context.Context, symbol, period string, startTime, endTime int64, limit int) ([]models.OpenInterestHist, error) {
	req := c.rc.R().
		SetContext(ctx).
		SetQueryParams(map[string]string{
			"symbol": symbol,
			"period": period,
		})

	if limit <= 0 {
		limit = 500
	}
	req.SetQueryParam("limit", fmt.Sprintf("%d", limit))

	if startTime > 0 {
		req.SetQueryParam("startTime", fmt.Sprintf("%d", startTime))
	}
	if endTime > 0 {
		req.SetQueryParam("endTime", fmt.Sprintf("%d", endTime))
	}

	var res []models.OpenInterestHist
	resp, err := req.SetResult(&res).Get("/futures/data/openInterestHist")
	if err != nil {
		return nil, fmt.Errorf("failed to get open interest hist: %w", err)
	}
	if resp.IsError() {
		return nil, fmt.Errorf("api error, code: %d message: %s", resp.StatusCode(), resp.String())
	}

	return res, nil
}

// Klines uses manual unmarshal because the API returns a raw 2D array, not a JSON object.
func (c *Client) Klines(ctx context.Context, symbol, interval string, limit int, startTime, endTime int64) ([]models.Kline, error) {
	req := c.rc.R().
		SetContext(ctx).
		SetQueryParams(map[string]string{
			"symbol":   symbol,
			"interval": interval,
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

	resp, err := req.Get("/fapi/v1/klines")
	if err != nil {
		return nil, fmt.Errorf("failed to get klines: %w", err)
	}
	if resp.IsError() {
		return nil, fmt.Errorf("api error, code: %d message: %s", resp.StatusCode(), resp.String())
	}

	var rawKlines [][]any
	if err := json.Unmarshal(resp.Body(), &rawKlines); err != nil {
		return nil, fmt.Errorf("failed to unmarshal klines: %w", err)
	}

	klines := make([]models.Kline, 0, len(rawKlines))
	for _, r := range rawKlines {
		if len(r) < 11 {
			continue
		}

		k := models.Kline{
			OpenTime:  int64(r[0].(float64)),
			CloseTime: int64(r[6].(float64)),
		}

		if s, ok := r[1].(string); ok {
			k.Open, _ = strconv.ParseFloat(s, 64)
		}
		if s, ok := r[2].(string); ok {
			k.High, _ = strconv.ParseFloat(s, 64)
		}
		if s, ok := r[3].(string); ok {
			k.Low, _ = strconv.ParseFloat(s, 64)
		}
		if s, ok := r[4].(string); ok {
			k.Close, _ = strconv.ParseFloat(s, 64)
		}
		if s, ok := r[5].(string); ok {
			k.Volume, _ = strconv.ParseFloat(s, 64)
		}
		if s, ok := r[7].(string); ok {
			k.QuoteAssetVolume, _ = strconv.ParseFloat(s, 64)
		}
		k.NumberOfTrades = int64(r[8].(float64))
		if s, ok := r[9].(string); ok {
			k.TakerBuyBaseAssetVolume, _ = strconv.ParseFloat(s, 64)
		}
		if s, ok := r[10].(string); ok {
			k.TakerBuyQuoteAssetVolume, _ = strconv.ParseFloat(s, 64)
		}

		klines = append(klines, k)
	}

	return klines, nil
}
