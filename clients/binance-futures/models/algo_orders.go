package models

type AlgoType string

const (
	AlgoTypeConditional AlgoType = "CONDITIONAL"
)

type AlgoStatus string

const (
	AlgoStatusNew       AlgoStatus = "NEW"
	AlgoStatusFilled    AlgoStatus = "FILLED"
	AlgoStatusCanceled  AlgoStatus = "CANCELLED"
	AlgoStatusExpired   AlgoStatus = "EXPIRED"
	AlgoStatusTriggered AlgoStatus = "TRIGGERED"
)

type AlgoOrder struct {
	AlgoID        int64        `json:"algoId"`
	ClientAlgoID  string       `json:"clientAlgoId"`
	AlgoType      AlgoType     `json:"algoType"`
	OrderType     OrderType    `json:"orderType"`
	Symbol        string       `json:"symbol"`
	Side          OrderSide    `json:"side"`
	PositionSide  PositionSide `json:"positionSide"`
	TimeInForce   TimeInForce  `json:"timeInForce"`
	Quantity      float64      `json:"quantity,string"`
	AlgoStatus    AlgoStatus   `json:"algoStatus"`
	ActualOrderID int64        `json:"actualOrderId"`
	ActualPrice   float64      `json:"actualPrice,string"`
	TriggerPrice  float64      `json:"triggerPrice,string"`
	Price         float64      `json:"price,string"`
	WorkingType   WorkingType  `json:"workingType"`
	PriceMatch    string       `json:"priceMatch"`
	ClosePosition bool         `json:"closePosition"`
	PriceProtect  bool         `json:"priceProtect"`
	ReduceOnly    bool         `json:"reduceOnly"`
	CreateTime    int64        `json:"createTime"`
	UpdateTime    int64        `json:"updateTime"`
	TriggerTime   int64        `json:"triggerTime"`
	GoodTillDate  int64        `json:"goodTillDate"`
}

type CreateAlgoOrderRequest struct {
	Symbol       string
	Side         OrderSide
	Type         OrderType
	AlgoType     AlgoType
	Quantity     *float64
	Price        *float64
	TriggerPrice *float64
	TimeInForce  TimeInForce
	ReduceOnly   *bool
	WorkingType  WorkingType
	PriceProtect *bool
	PositionSide PositionSide
	ClosePosition *bool
	GoodTillDate *int64
}

func NewCreateAlgoOrderRequest(symbol string, side OrderSide, orderType OrderType) *CreateAlgoOrderRequest {
	return &CreateAlgoOrderRequest{
		Symbol:   symbol,
		Side:     side,
		Type:     orderType,
		AlgoType: AlgoTypeConditional,
	}
}

func (r *CreateAlgoOrderRequest) WithQuantity(quantity float64) *CreateAlgoOrderRequest {
	r.Quantity = &quantity
	return r
}

func (r *CreateAlgoOrderRequest) WithPrice(price float64) *CreateAlgoOrderRequest {
	r.Price = &price
	return r
}

func (r *CreateAlgoOrderRequest) WithTriggerPrice(triggerPrice float64) *CreateAlgoOrderRequest {
	r.TriggerPrice = &triggerPrice
	return r
}

func (r *CreateAlgoOrderRequest) WithTimeInForce(tif TimeInForce) *CreateAlgoOrderRequest {
	r.TimeInForce = tif
	return r
}

func (r *CreateAlgoOrderRequest) WithReduceOnly(reduceOnly bool) *CreateAlgoOrderRequest {
	r.ReduceOnly = &reduceOnly
	return r
}

func (r *CreateAlgoOrderRequest) WithWorkingType(wt WorkingType) *CreateAlgoOrderRequest {
	r.WorkingType = wt
	return r
}

func (r *CreateAlgoOrderRequest) WithPriceProtect(priceProtect bool) *CreateAlgoOrderRequest {
	r.PriceProtect = &priceProtect
	return r
}

func (r *CreateAlgoOrderRequest) WithPositionSide(ps PositionSide) *CreateAlgoOrderRequest {
	r.PositionSide = ps
	return r
}

func (r *CreateAlgoOrderRequest) WithClosePosition(closePosition bool) *CreateAlgoOrderRequest {
	r.ClosePosition = &closePosition
	return r
}

func (r *CreateAlgoOrderRequest) WithGoodTillDate(gtd int64) *CreateAlgoOrderRequest {
	r.GoodTillDate = &gtd
	return r
}
