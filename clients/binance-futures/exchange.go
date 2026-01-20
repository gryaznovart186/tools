package binance_futures

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/gryaznovart186/tools/clients/binance-futures/models"
)

func (c *Client) ExchangeSymbolsInfo(ctx context.Context, isOnlyInTrading bool) ([]models.SymbolInfo, error) {
	resp, err := c.rc.R().
		SetContext(ctx).
		Get("/fapi/v1/exchangeInfo")
	if err != nil {
		return nil, fmt.Errorf("failed to get exchange info: %w", err)
	}

	var res models.ExchangeInfo
	if err := json.Unmarshal(resp.Body(), &res); err != nil {
		return nil, fmt.Errorf("failed to unmarshal exchange info: %w", err)
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
	req := c.rc.R().SetContext(ctx)

	resp, err := req.Get("/fapi/v1/ticker/24hr")
	if err != nil {
		return nil, fmt.Errorf("failed to get price change stats: %w", err)
	}

	var res []models.PriceChangeStats
	if err := json.Unmarshal(resp.Body(), &res); err != nil {
		return nil, fmt.Errorf("failed to unmarshal price change stats: %w", err)
	}

	return res, nil
}

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
