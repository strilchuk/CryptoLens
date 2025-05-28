package bybit

import "time"

// BybitResponse представляет базовый ответ от API Bybit
type BybitResponse struct {
	RetCode    int         `json:"retCode"`
	RetMsg     string      `json:"retMsg"`
	Result     interface{} `json:"result"`
	RetExtInfo interface{} `json:"retExtInfo"`
	Time       int64       `json:"time"`
}

// IsSuccess проверяет успешность ответа
func (r *BybitResponse) IsSuccess() bool {
	return r.RetCode == 0
}

// BybitAccount представляет аккаунт Bybit
type BybitAccount struct {
	ID          int64      `json:"id"`
	UserID      string     `json:"user_id"`
	APIKey      string     `json:"api_key"`
	APISecret   string     `json:"api_secret"`
	AccountType string     `json:"account_type"`
	IsActive    bool       `json:"is_active"`
	CreatedAt   *time.Time `json:"-"`
	UpdatedAt   *time.Time `json:"-"`
	DeletedAt   *time.Time `json:"-"`
}

// BybitWalletBalance представляет баланс кошелька
type BybitWalletBalance struct {
	List []BybitAccountBalance `json:"list"`
}

// BybitAccountBalance представляет баланс аккаунта
type BybitAccountBalance struct {
	AccountIMRate          string              `json:"accountIMRate"`
	AccountLTV            string              `json:"accountLTV"`
	AccountMMRate         string              `json:"accountMMRate"`
	AccountType           string              `json:"accountType"`
	Coins                 []BybitCoinBalance  `json:"coin"`
	TotalAvailableBalance string              `json:"totalAvailableBalance"`
	TotalEquity           string              `json:"totalEquity"`
	TotalInitialMargin    string              `json:"totalInitialMargin"`
	TotalMaintenanceMargin string             `json:"totalMaintenanceMargin"`
	TotalMarginBalance    string              `json:"totalMarginBalance"`
	TotalPerpUPL          string              `json:"totalPerpUPL"`
	TotalWalletBalance    string              `json:"totalWalletBalance"`
}

// BybitCoinBalance представляет баланс монеты
type BybitCoinBalance struct {
	AccruedInterest    string `json:"accruedInterest"`
	AvailableToBorrow  string `json:"availableToBorrow"`
	AvailableToWithdraw string `json:"availableToWithdraw"`
	Bonus              string `json:"bonus"`
	BorrowAmount       string `json:"borrowAmount"`
	Coin               string `json:"coin"`
	CollateralSwitch   bool   `json:"collateralSwitch"`
	CumRealisedPnl     string `json:"cumRealisedPnl"`
	Equity             string `json:"equity"`
	Locked             string `json:"locked"`
	MarginCollateral   bool   `json:"marginCollateral"`
	SpotHedgingQty     string `json:"spotHedgingQty"`
	TotalOrderIM       string `json:"totalOrderIM"`
	TotalPositionIM    string `json:"totalPositionIM"`
	TotalPositionMM    string `json:"totalPositionMM"`
	UnrealisedPnl      string `json:"unrealisedPnl"`
	USDValue           string `json:"usdValue"`
	WalletBalance      string `json:"walletBalance"`
}

// BybitInstrumentsResponse представляет ответ со списком инструментов
type BybitInstrumentsResponse struct {
	Category string              `json:"category"`
	List     []BybitInstrument  `json:"list"`
}

// BybitInstrument представляет торговый инструмент
type BybitInstrument struct {
	Symbol        string              `json:"symbol"`
	BaseCoin      string              `json:"baseCoin"`
	QuoteCoin     string              `json:"quoteCoin"`
	Innovation    string              `json:"innovation"`
	Status        string              `json:"status"`
	MarginTrading string              `json:"marginTrading"`
	StTag         string              `json:"stTag"`
	LotSizeFilter BybitLotSizeFilter  `json:"lotSizeFilter"`
	PriceFilter   BybitPriceFilter    `json:"priceFilter"`
	RiskParameters BybitRiskParameters `json:"riskParameters"`
}

// BybitLotSizeFilter представляет фильтр размера лота
type BybitLotSizeFilter struct {
	BasePrecision  string `json:"basePrecision"`
	QuotePrecision string `json:"quotePrecision"`
	MinOrderQty    string `json:"minOrderQty"`
	MaxOrderQty    string `json:"maxOrderQty"`
	MinOrderAmt    string `json:"minOrderAmt"`
	MaxOrderAmt    string `json:"maxOrderAmt"`
}

// BybitPriceFilter представляет фильтр цены
type BybitPriceFilter struct {
	TickSize string `json:"tickSize"`
}

// BybitRiskParameters представляет параметры риска
type BybitRiskParameters struct {
	PriceLimitRatioX string `json:"priceLimitRatioX"`
	PriceLimitRatioY string `json:"priceLimitRatioY"`
}

// BybitTickersResponse представляет ответ со списком тикеров
type BybitTickersResponse struct {
	Category string         `json:"category"`
	List     []BybitTicker `json:"list"`
}

// BybitTicker представляет тикер
type BybitTicker struct {
	Symbol            string `json:"symbol"`
	LastPrice         string `json:"lastPrice"`
	HighPrice24h      string `json:"highPrice24h"`
	LowPrice24h       string `json:"lowPrice24h"`
	PrevPrice24h      string `json:"prevPrice24h"`
	Volume24h         string `json:"volume24h"`
	Turnover24h       string `json:"turnover24h"`
	Price24hPcnt      string `json:"price24hPcnt"`
	Price1hPcnt       string `json:"price1hPcnt"`
	MarkPrice         string `json:"markPrice"`
	IndexPrice        string `json:"indexPrice"`
	OpenInterest      string `json:"openInterest"`
	OpenInterestValue string `json:"openInterestValue"`
	TotalTurnover     string `json:"totalTurnover"`
	TotalVolume       string `json:"totalVolume"`
	FundingRate       string `json:"fundingRate"`
	NextFundTime      string `json:"nextFundTime"`
	Bid1Price         string `json:"bid1Price"`
	Bid1Size          string `json:"bid1Size"`
	Ask1Price         string `json:"ask1Price"`
	Ask1Size          string `json:"ask1Size"`
}

// BybitKlinesResponse представляет ответ со свечами
type BybitKlinesResponse struct {
	Category string        `json:"category"`
	Symbol   string        `json:"symbol"`
	Interval string        `json:"interval"`
	List     []BybitKline `json:"list"`
}

// BybitKline представляет свечу
type BybitKline struct {
	StartTime string `json:"startTime"`
	Open      string `json:"open"`
	High      string `json:"high"`
	Low       string `json:"low"`
	Close     string `json:"close"`
	Volume    string `json:"volume"`
	Turnover  string `json:"turnover"`
}

// BybitTradesResponse представляет ответ со сделками
type BybitTradesResponse struct {
	Category string        `json:"category"`
	Symbol   string        `json:"symbol"`
	List     []BybitTrade `json:"list"`
}

// BybitTrade представляет сделку
type BybitTrade struct {
	ExecID      string `json:"execId"`
	Symbol      string `json:"symbol"`
	Price       string `json:"price"`
	Size        string `json:"size"`
	Side        string `json:"side"`
	Time        string `json:"time"`
	IsBlockTrade bool   `json:"isBlockTrade"`
}

// BybitOrderResponse представляет ответ на создание/изменение ордера
type BybitOrderResponse struct {
	OrderID     string `json:"orderId"`
	OrderLinkID string `json:"orderLinkId"`
}

// BybitOrderListResponse представляет ответ со списком ордеров
type BybitOrderListResponse struct {
	List []BybitOrder `json:"list"`
}

// BybitOrder представляет ордер
type BybitOrder struct {
	OrderID       string `json:"orderId"`
	OrderLinkID   string `json:"orderLinkId"`
	Symbol        string `json:"symbol"`
	Side          string `json:"side"`
	OrderType     string `json:"orderType"`
	Price         string `json:"price"`
	Qty           string `json:"qty"`
	TimeInForce   string `json:"timeInForce"`
	OrderStatus   string `json:"orderStatus"`
	LeavesQty     string `json:"leavesQty"`
	CumExecQty    string `json:"cumExecQty"`
	CumExecValue  string `json:"cumExecValue"`
	CumExecFee    string `json:"cumExecFee"`
	CreateTime    string `json:"createTime"`
	UpdateTime    string `json:"updateTime"`
}

// BybitFeeRateResponse представляет ответ со ставками комиссии
type BybitFeeRateResponse struct {
	Category string         `json:"category"`
	List     []BybitFeeRate `json:"list"`
}

// BybitFeeRate представляет ставку комиссии
type BybitFeeRate struct {
	Symbol       string `json:"symbol"`
	TakerFeeRate string `json:"takerFeeRate"`
	MakerFeeRate string `json:"makerFeeRate"`
} 