package models

type OrderSide string

const (
	SideBuy  OrderSide = "BUY"
	SideSell OrderSide = "SELL"
)

type OrderType string

const (
	OrderTypeLimit              OrderType = "LIMIT"
	OrderTypeMarket             OrderType = "MARKET"
	OrderTypeStop               OrderType = "STOP"
	OrderTypeStopMarket         OrderType = "STOP_MARKET"
	OrderTypeTakeProfit         OrderType = "TAKE_PROFIT"
	OrderTypeTakeProfitMarket   OrderType = "TAKE_PROFIT_MARKET"
	OrderTypeTrailingStopMarket OrderType = "TRAILING_STOP_MARKET"
)

type PositionSide string

const (
	PositionSideBoth  PositionSide = "BOTH"
	PositionSideLong  PositionSide = "LONG"
	PositionSideShort PositionSide = "SHORT"
)

type TimeInForce string

const (
	TimeInForceGTC TimeInForce = "GTC" // Good Till Cancel
	TimeInForceIOC TimeInForce = "IOC" // Immediate Or Cancel
	TimeInForceFOK TimeInForce = "FOK" // Fill Or Kill
	TimeInForceGTX TimeInForce = "GTX" // Good Till Crossing (Post Only)
	TimeInForceGTD TimeInForce = "GTD" // Good Till Date
)

type WorkingType string

const (
	WorkingTypeMarkPrice     WorkingType = "MARK_PRICE"
	WorkingTypeContractPrice WorkingType = "CONTRACT_PRICE"
)

type OrderStatus string

const (
	OrderStatusNew             OrderStatus = "NEW"
	OrderStatusPartiallyFilled OrderStatus = "PARTIALLY_FILLED"
	OrderStatusFilled          OrderStatus = "FILLED"
	OrderStatusCanceled        OrderStatus = "CANCELED"
	OrderStatusExpired         OrderStatus = "EXPIRED"
)

type Order struct {
	Symbol           string       `json:"symbol"`
	OrderID          int64        `json:"orderId"`
	ClientOrderID    string       `json:"clientOrderId"`
	Price            float64      `json:"price,string"`
	OrigQty          float64      `json:"origQty,string"`
	ExecutedQty      float64      `json:"executedQty,string"`
	Status           OrderStatus  `json:"status"`
	TimeInForce      TimeInForce  `json:"timeInForce"`
	Type             OrderType    `json:"type"`
	Side             OrderSide    `json:"side"`
	StopPrice        float64      `json:"stopPrice,string"`
	WorkingType      WorkingType  `json:"workingType"`
	PriceProtect     bool         `json:"priceProtect"`
	OrigType         OrderType    `json:"origType"`
	UpdateTime       int64        `json:"updateTime"`
	AvgPrice         float64      `json:"avgPrice,string"`
	ReduceOnly       bool         `json:"reduceOnly"`
	ClosePosition    bool         `json:"closePosition"`
	PositionSide     PositionSide `json:"positionSide"`
	StopMain         bool         `json:"stopMain"`
	PriceMatch       string       `json:"priceMatch"`
	SelfTradePrevMod string       `json:"selfTradePreventionMode"`
	GoodTillDate     int64        `json:"goodTillDate"`
}

type CreateOrderRequest struct {
	Symbol           string
	Side             OrderSide
	Type             OrderType
	Quantity         *float64
	Price            *float64
	StopPrice        *float64
	TimeInForce      TimeInForce
	ReduceOnly       *bool
	WorkingType      WorkingType
	PriceProtect     *bool
	PositionSide     PositionSide
	ClosePosition    *bool
	PriceMatch       string
	SelfTradePrevMod string
	GoodTillDate     *int64
}

func NewCreateOrderRequest(symbol string, side OrderSide, orderType OrderType) *CreateOrderRequest {
	return &CreateOrderRequest{
		Symbol: symbol,
		Side:   side,
		Type:   orderType,
	}
}

func (r *CreateOrderRequest) WithQuantity(quantity float64) *CreateOrderRequest {
	r.Quantity = &quantity
	return r
}

func (r *CreateOrderRequest) WithPrice(price float64) *CreateOrderRequest {
	r.Price = &price
	return r
}

func (r *CreateOrderRequest) WithStopPrice(stopPrice float64) *CreateOrderRequest {
	r.StopPrice = &stopPrice
	return r
}

func (r *CreateOrderRequest) WithTimeInForce(tif TimeInForce) *CreateOrderRequest {
	r.TimeInForce = tif
	return r
}

func (r *CreateOrderRequest) WithReduceOnly(reduceOnly bool) *CreateOrderRequest {
	r.ReduceOnly = &reduceOnly
	return r
}

func (r *CreateOrderRequest) WithWorkingType(wt WorkingType) *CreateOrderRequest {
	r.WorkingType = wt
	return r
}

func (r *CreateOrderRequest) WithPriceProtect(priceProtect bool) *CreateOrderRequest {
	r.PriceProtect = &priceProtect
	return r
}

func (r *CreateOrderRequest) WithPositionSide(ps PositionSide) *CreateOrderRequest {
	r.PositionSide = ps
	return r
}

func (r *CreateOrderRequest) WithClosePosition(closePosition bool) *CreateOrderRequest {
	r.ClosePosition = &closePosition
	return r
}

func (r *CreateOrderRequest) WithGoodTillDate(gtd int64) *CreateOrderRequest {
	r.GoodTillDate = &gtd
	return r
}
