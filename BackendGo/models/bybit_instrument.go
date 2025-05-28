package models

import (
	"github.com/shopspring/decimal"
	"time"
)

// BybitInstrument представляет собой модель торгового инструмента на бирже Bybit
type BybitInstrument struct {
	ID            int64           `json:"id" db:"id"`                         // Уникальный идентификатор инструмента
	Symbol        string          `json:"symbol" db:"symbol"`                 // Символ торговой пары (например, BTCUSDT)
	Category      string          `json:"category" db:"category"`             // Категория инструмента (spot, linear, inverse)
	BaseCoin      string          `json:"base_coin" db:"base_coin"`           // Базовая валюта (например, BTC)
	QuoteCoin     string          `json:"quote_coin" db:"quote_coin"`         // Котируемая валюта (например, USDT)
	MinOrderQty   decimal.Decimal `json:"min_order_qty" db:"min_order_qty"`   // Минимальный размер ордера
	MaxOrderQty   decimal.Decimal `json:"max_order_qty" db:"max_order_qty"`   // Максимальный размер ордера
	MinPrice      decimal.Decimal `json:"min_price" db:"min_price"`           // Минимальная цена
	MaxPrice      decimal.Decimal `json:"max_price" db:"max_price"`           // Максимальная цена
	PriceScale    int             `json:"price_scale" db:"price_scale"`       // Количество знаков после запятой для цены
	QuantityScale int             `json:"quantity_scale" db:"quantity_scale"` // Количество знаков после запятой для количества
	Status        string          `json:"status" db:"status"`                 // Статус инструмента (Trading, Suspended и т.д.)
	CreatedAt     time.Time       `json:"created_at" db:"created_at"`         // Время создания записи
	UpdatedAt     time.Time       `json:"updated_at" db:"updated_at"`         // Время последнего обновления записи
}

// BybitInstrumentResponse представляет собой структуру ответа API для списка инструментов
type BybitInstrumentResponse struct {
	Status string            `json:"status"` // Статус ответа (success/error)
	Data   []BybitInstrument `json:"data"`   // Список инструментов
}
