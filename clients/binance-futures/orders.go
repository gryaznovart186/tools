package binance_futures

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/gryaznovart186/tools/clients/binance-futures/models"
)

func (c *Client) CreateOrder(ctx context.Context, req *models.CreateOrderRequest) (*models.Order, error) {
	if err := c.checkCredentials(); err != nil {
		return nil, err
	}
	params := url.Values{}
	params.Set("symbol", req.Symbol)
	params.Set("side", string(req.Side))
	params.Set("type", string(req.Type))

	if req.Quantity != nil {
		params.Set("quantity", fmt.Sprintf("%f", *req.Quantity))
	}
	if req.Price != nil {
		params.Set("price", fmt.Sprintf("%f", *req.Price))
	}
	if req.StopPrice != nil {
		params.Set("stopPrice", fmt.Sprintf("%f", *req.StopPrice))
	}
	if req.TimeInForce != "" {
		params.Set("timeInForce", string(req.TimeInForce))
	}
	if req.ReduceOnly != nil {
		params.Set("reduceOnly", fmt.Sprintf("%t", *req.ReduceOnly))
	}
	if req.WorkingType != "" {
		params.Set("workingType", string(req.WorkingType))
	}
	if req.PriceProtect != nil {
		params.Set("priceProtect", fmt.Sprintf("%t", *req.PriceProtect))
	}
	if req.PositionSide != "" {
		params.Set("positionSide", string(req.PositionSide))
	}
	if req.ClosePosition != nil {
		params.Set("closePosition", fmt.Sprintf("%t", *req.ClosePosition))
	}
	if req.GoodTillDate != nil {
		params.Set("goodTillDate", fmt.Sprintf("%d", *req.GoodTillDate))
	}

	fullQuery := c.withSignature(params.Encode())

	resp, err := c.rc.R().
		SetContext(ctx).
		SetHeader("X-MBX-APIKEY", c.creds.apiKey).
		Post("/fapi/v1/order?" + fullQuery)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode(), resp.Body())
	}

	var order models.Order
	err = json.Unmarshal(resp.Body(), &order)
	if err != nil {
		return nil, ErrInvalidResponse
	}

	return &order, nil
}

func (c *Client) GetOrder(ctx context.Context, symbol string, orderID int64) (*models.Order, error) {
	if err := c.checkCredentials(); err != nil {
		return nil, err
	}
	params := url.Values{}
	params.Set("symbol", symbol)
	if orderID != 0 {
		params.Set("orderId", fmt.Sprintf("%d", orderID))
	}

	fullQuery := c.withSignature(params.Encode())

	resp, err := c.rc.R().
		SetContext(ctx).
		SetHeader("X-MBX-APIKEY", c.creds.apiKey).
		Get("/fapi/v1/order?" + fullQuery)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode(), resp.Body())
	}

	var order models.Order
	err = json.Unmarshal(resp.Body(), &order)
	if err != nil {
		return nil, ErrInvalidResponse
	}

	return &order, nil
}

func (c *Client) CancelOrder(ctx context.Context, symbol string, orderID int64) (*models.Order, error) {
	if err := c.checkCredentials(); err != nil {
		return nil, err
	}
	params := url.Values{}
	params.Set("symbol", symbol)
	if orderID != 0 {
		params.Set("orderId", fmt.Sprintf("%d", orderID))
	}

	fullQuery := c.withSignature(params.Encode())

	resp, err := c.rc.R().
		SetContext(ctx).
		SetHeader("X-MBX-APIKEY", c.creds.apiKey).
		Delete("/fapi/v1/order?" + fullQuery)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode(), resp.Body())
	}

	var order models.Order
	err = json.Unmarshal(resp.Body(), &order)
	if err != nil {
		return nil, ErrInvalidResponse
	}

	return &order, nil
}
