package models

import (
	"encoding/json"
	"github.com/shopspring/decimal"
	"time"
)

// BybitInstrument представляет собой модель торгового инструмента на бирже Bybit
type BybitInstrument struct {
	Symbol           string          `json:"symbol"`           // Символ торговой пары (например, BTCUSDT)
	Category         string          `json:"category"`         // Категория инструмента (spot, linear, inverse)
	BaseCoin         string          `json:"baseCoin"`         // Базовая валюта (например, BTC)
	QuoteCoin        string          `json:"quoteCoin"`        // Котируемая валюта (например, USDT)
	Innovation       string          `json:"innovation"`       // Инновационный статус инструмента
	Status           string          `json:"status"`           // Статус инструмента (Trading, Suspended и т.д.)
	MarginTrading    string          `json:"marginTrading"`    // Статус маржинальной торговли
	StTag            string          `json:"stTag"`            // Специальный тег инструмента
	BasePrecision    decimal.Decimal `json:"basePrecision"`    // Точность базовой валюты
	QuotePrecision   decimal.Decimal `json:"quotePrecision"`   // Точность котируемой валюты
	MinOrderQty      decimal.Decimal `json:"minOrderQty"`      // Минимальный размер ордера в базовой валюте
	MaxOrderQty      decimal.Decimal `json:"maxOrderQty"`      // Максимальный размер ордера в базовой валюте
	MinOrderAmt      decimal.Decimal `json:"minOrderAmt"`      // Минимальная сумма ордера в котируемой валюте
	MaxOrderAmt      decimal.Decimal `json:"maxOrderAmt"`      // Максимальная сумма ордера в котируемой валюте
	TickSize         decimal.Decimal `json:"tickSize"`         // Минимальный шаг цены
	PriceLimitRatioX decimal.Decimal `json:"priceLimitRatioX"` // Коэффициент ограничения цены X
	PriceLimitRatioY decimal.Decimal `json:"priceLimitRatioY"` // Коэффициент ограничения цены Y
	CreatedAt        time.Time       `json:"created_at"`       // Время создания записи
	UpdatedAt        time.Time       `json:"updated_at"`       // Время последнего обновления записи
}

// UnmarshalJSON реализует кастомную десериализацию JSON для корректной обработки строковых значений в decimal.Decimal
func (d *BybitInstrument) UnmarshalJSON(data []byte) error {
	type Alias BybitInstrument
	aux := &struct {
		BasePrecision    string `json:"basePrecision"`
		QuotePrecision   string `json:"quotePrecision"`
		MinOrderQty      string `json:"minOrderQty"`
		MaxOrderQty      string `json:"maxOrderQty"`
		MinOrderAmt      string `json:"minOrderAmt"`
		MaxOrderAmt      string `json:"maxOrderAmt"`
		TickSize         string `json:"tickSize"`
		PriceLimitRatioX string `json:"priceLimitRatioX"`
		PriceLimitRatioY string `json:"priceLimitRatioY"`
		*Alias
	}{
		Alias: (*Alias)(d),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	d.BasePrecision = decimal.RequireFromString(aux.BasePrecision)
	d.QuotePrecision = decimal.RequireFromString(aux.QuotePrecision)
	d.MinOrderQty = decimal.RequireFromString(aux.MinOrderQty)
	d.MaxOrderQty = decimal.RequireFromString(aux.MaxOrderQty)
	d.MinOrderAmt = decimal.RequireFromString(aux.MinOrderAmt)
	d.MaxOrderAmt = decimal.RequireFromString(aux.MaxOrderAmt)
	d.TickSize = decimal.RequireFromString(aux.TickSize)
	d.PriceLimitRatioX = decimal.RequireFromString(aux.PriceLimitRatioX)
	d.PriceLimitRatioY = decimal.RequireFromString(aux.PriceLimitRatioY)
	return nil
}

// BybitInstrumentResponse представляет собой структуру ответа API для списка инструментов
type BybitInstrumentResponse struct {
	Status string            `json:"status"` // Статус ответа (success/error)
	Data   []BybitInstrument `json:"data"`   // Список инструментов
}
