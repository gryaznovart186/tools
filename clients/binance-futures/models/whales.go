package models

type LongShortRatio struct {
	Symbol         string  `json:"symbol"`
	LongShortRatio float64 `json:"longShortRatio,string"`
	LongAccount    float64 `json:"longAccount,string"`
	ShortAccount   float64 `json:"shortAccount,string"`
	Timestamp      int64   `json:"timestamp"`
}

type TakerLongShortRatio struct {
	BuySellRatio float64 `json:"buySellRatio,string"`
	BuyVol       float64 `json:"buyVol,string"`
	SellVol      float64 `json:"sellVol,string"`
	Timestamp    int64   `json:"timestamp"`
}
