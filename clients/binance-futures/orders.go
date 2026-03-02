package binance_futures

import (
	"context"
	"fmt"
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

	var order models.Order
	resp, err := c.rc.R().
		SetContext(ctx).
		SetHeader("X-MBX-APIKEY", c.creds.apiKey).
		SetResult(&order).
		Post("/fapi/v1/order?" + fullQuery)

	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, fmt.Errorf("api error, code: %d message: %s", resp.StatusCode(), resp.String())
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

	var order models.Order
	resp, err := c.rc.R().
		SetContext(ctx).
		SetHeader("X-MBX-APIKEY", c.creds.apiKey).
		SetResult(&order).
		Get("/fapi/v1/order?" + fullQuery)

	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, fmt.Errorf("api error, code: %d message: %s", resp.StatusCode(), resp.String())
	}

	return &order, nil
}

func (c *Client) OpenOrders(ctx context.Context, symbol string) ([]models.Order, error) {
	if err := c.checkCredentials(); err != nil {
		return nil, err
	}
	params := url.Values{}
	if symbol != "" {
		params.Set("symbol", symbol)
	}

	fullQuery := c.withSignature(params.Encode())

	var orders []models.Order
	resp, err := c.rc.R().
		SetContext(ctx).
		SetHeader("X-MBX-APIKEY", c.creds.apiKey).
		SetResult(&orders).
		Get("/fapi/v1/openOrders?" + fullQuery)

	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, fmt.Errorf("api error, code: %d message: %s", resp.StatusCode(), resp.String())
	}

	return orders, nil
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

	var order models.Order
	resp, err := c.rc.R().
		SetContext(ctx).
		SetHeader("X-MBX-APIKEY", c.creds.apiKey).
		SetResult(&order).
		Delete("/fapi/v1/order?" + fullQuery)

	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, fmt.Errorf("api error, code: %d message: %s", resp.StatusCode(), resp.String())
	}

	return &order, nil
}
