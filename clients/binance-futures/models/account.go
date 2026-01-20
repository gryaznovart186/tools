package models

type AccountBalance struct {
	Asset            string  `json:"asset"`
	Balance          float64 `json:"balance,string"`
	AvailableBalance float64 `json:"availableBalance,string"`
	MarginAvailable  bool    `json:"marginAvailable"`
}

type Position struct {
	Symbol           string  `json:"symbol"`
	PositionAmt      float64 `json:"positionAmt,string"`
	EntryPrice       float64 `json:"entryPrice,string"`
	MarkPrice        float64 `json:"markPrice,string"`
	UnRealizedProfit float64 `json:"unRealizedProfit,string"`
	LiquidationPrice float64 `json:"liquidationPrice,string"`
	Leverage         float64 `json:"leverage,string"`
	MaxNotionalValue float64 `json:"maxNotionalValue,string"`
	MarginType       string  `json:"marginType"`
	IsolatedMargin   float64 `json:"isolatedMargin,string"`
	IsAutoAddMargin  bool    `json:"isAutoAddMargin"`
	PositionSide     string  `json:"positionSide"`
	Notional         float64 `json:"notional,string"`
	IsolatedWallet   float64 `json:"isolatedWallet,string"`
	UpdateHeight     int64   `json:"updateHeight"`
}
