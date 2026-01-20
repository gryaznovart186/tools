package models

type RateLimits struct {
	UsedWeight1m int
}

type ExchangeInfo struct {
	Timezone        string       `json:"timezone"`
	ServerTime      int64        `json:"serverTime"`
	FuturesType     string       `json:"futuresType"`
	RateLimits      []RateLimit  `json:"rateLimits"`
	ExchangeFilters []any        `json:"exchangeFilters"`
	Symbols         []SymbolInfo `json:"symbols"`
}

type RateLimit struct {
	Interval      string `json:"interval"`
	IntervalNum   int    `json:"intervalNum"`
	Limit         int    `json:"limit"`
	RateLimitType string `json:"rateLimitType"`
}

type SymbolInfo struct {
	Pair    string         `json:"pair"`
	Status  string         `json:"status"`
	Filters []SymbolFilter `json:"filters"`
}

type SymbolFilter struct {
	FilterType        string  `json:"filterType"`
	MaxPrice          float64 `json:"maxPrice,string,omitempty"`
	MinPrice          float64 `json:"minPrice,string,omitempty"`
	TickSize          float64 `json:"tickSize,string,omitempty"`
	MaxQty            float64 `json:"maxQty,string,omitempty"`
	MinQty            float64 `json:"minQty,string,omitempty"`
	StepSize          float64 `json:"stepSize,string,omitempty"`
	Limit             int     `json:"limit,omitempty"`
	Notional          float64 `json:"notional,string,omitempty"`
	MultiplierUp      float64 `json:"multiplierUp,string,omitempty"`
	MultiplierDown    float64 `json:"multiplierDown,string,omitempty"`
	MultiplierDecimal int     `json:"multiplierDecimal,string,omitempty"`
}

type PriceChangeStats struct {
	Symbol             string  `json:"symbol"`
	PriceChange        float64 `json:"priceChange,string"`
	PriceChangePercent float64 `json:"priceChangePercent,string"`
	LastPrice          float64 `json:"lastPrice,string"`
	Volume             float64 `json:"volume,string"`
}

type Kline struct {
	OpenTime                 int64   `json:"openTime"`
	Open                     float64 `json:"open,string"`
	High                     float64 `json:"high,string"`
	Low                      float64 `json:"low,string"`
	Close                    float64 `json:"close,string"`
	Volume                   float64 `json:"volume,string"`
	CloseTime                int64   `json:"closeTime"`
	QuoteAssetVolume         float64 `json:"quoteAssetVolume,string"`
	NumberOfTrades           int64   `json:"numberOfTrades"`
	TakerBuyBaseAssetVolume  float64 `json:"takerBuyBaseAssetVolume,string"`
	TakerBuyQuoteAssetVolume float64 `json:"takerBuyQuoteAssetVolume,string"`
}
