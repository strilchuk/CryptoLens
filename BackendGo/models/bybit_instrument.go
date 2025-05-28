package models

import (
	"github.com/shopspring/decimal"
	"time"
)

// BybitInstrument представляет собой модель торгового инструмента на бирже Bybit
type BybitInstrument struct {
	Symbol           string          `json:"symbol" db:"symbol"`                 // Символ торговой пары (например, BTCUSDT)
	Category         string          `json:"category" db:"category"`             // Категория инструмента (spot, linear, inverse)
	BaseCoin         string          `json:"base_coin" db:"base_coin"`           // Базовая валюта (например, BTC)
	QuoteCoin        string          `json:"quote_coin" db:"quote_coin"`         // Котируемая валюта (например, USDT)
	PricePrecision   int             `json:"price_precision" db:"price_precision"` // Количество знаков после запятой для цены
	QuantityPrecision int            `json:"quantity_precision" db:"quantity_precision"` // Количество знаков после запятой для количества
	MinPrice         decimal.Decimal `json:"min_price" db:"min_price"`           // Минимальная цена
	MaxPrice         decimal.Decimal `json:"max_price" db:"max_price"`           // Максимальная цена
	MinQuantity      decimal.Decimal `json:"min_quantity" db:"min_quantity"`     // Минимальный размер ордера
	MaxQuantity      decimal.Decimal `json:"max_quantity" db:"max_quantity"`     // Максимальный размер ордера
	QuantityStep     decimal.Decimal `json:"quantity_step" db:"quantity_step"`   // Шаг размера ордера
	PriceStep        decimal.Decimal `json:"price_step" db:"price_step"`         // Шаг цены
	Status           string          `json:"status" db:"status"`                 // Статус инструмента (Trading, Suspended и т.д.)
	CreatedAt        time.Time       `json:"created_at" db:"created_at"`         // Время создания записи
	UpdatedAt        time.Time       `json:"updated_at" db:"updated_at"`         // Время последнего обновления записи
}

// BybitInstrumentResponse представляет собой структуру ответа API для списка инструментов
type BybitInstrumentResponse struct {
	Status string            `json:"status"` // Статус ответа (success/error)
	Data   []BybitInstrument `json:"data"`   // Список инструментов
}
