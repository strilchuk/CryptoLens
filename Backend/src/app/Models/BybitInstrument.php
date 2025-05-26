<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\SoftDeletes;

/**
 * Модель инструмента Bybit
 *
 * @property int $id Уникальный идентификатор
 * @property string $symbol Символ торговой пары (например, BTCUSDT)
 * @property string $base_coin Базовая валюта (например, BTC)
 * @property string $quote_coin Котируемая валюта (например, USDT)
 * @property string $innovation Флаг инновационной торговой пары
 * @property string $status Статус торговой пары (Trading, Suspended и т.д.)
 * @property string $margin_trading Тип маржинальной торговли
 * @property string $st_tag Специальный тег
 * 
 * @property string $base_precision Точность базовой валюты
 * @property string $quote_precision Точность котируемой валюты
 * @property string $min_order_qty Минимальный размер ордера в базовой валюте
 * @property string $max_order_qty Максимальный размер ордера в базовой валюте
 * @property string $min_order_amt Минимальная сумма ордера в котируемой валюте
 * @property string $max_order_amt Максимальная сумма ордера в котируемой валюте
 * 
 * @property string $tick_size Минимальный шаг цены
 * 
 * @property string $price_limit_ratio_x Коэффициент ограничения цены X
 * @property string $price_limit_ratio_y Коэффициент ограничения цены Y
 * 
 * @property \Carbon\Carbon $created_at
 * @property \Carbon\Carbon $updated_at
 * @property \Carbon\Carbon|null $deleted_at
 */
class BybitInstrument extends Model
{
    use SoftDeletes;

    protected $fillable = [
        'symbol',
        'base_coin',
        'quote_coin',
        'innovation',
        'status',
        'margin_trading',
        'st_tag',
        'base_precision',
        'quote_precision',
        'min_order_qty',
        'max_order_qty',
        'min_order_amt',
        'max_order_amt',
        'tick_size',
        'price_limit_ratio_x',
        'price_limit_ratio_y',
    ];

    /**
     * Создать или обновить инструмент из DTO
     */
    public static function updateFromDTO(\App\DTOs\Bybit\BybitInstrumentDTO $dto): self
    {
        return self::updateOrCreate(
            ['symbol' => $dto->symbol],
            [
                'base_coin' => $dto->baseCoin,
                'quote_coin' => $dto->quoteCoin,
                'innovation' => $dto->innovation,
                'status' => $dto->status,
                'margin_trading' => $dto->marginTrading,
                'st_tag' => $dto->stTag,
                'base_precision' => $dto->lotSizeFilter->basePrecision,
                'quote_precision' => $dto->lotSizeFilter->quotePrecision,
                'min_order_qty' => $dto->lotSizeFilter->minOrderQty,
                'max_order_qty' => $dto->lotSizeFilter->maxOrderQty,
                'min_order_amt' => $dto->lotSizeFilter->minOrderAmt,
                'max_order_amt' => $dto->lotSizeFilter->maxOrderAmt,
                'tick_size' => $dto->priceFilter->tickSize,
                'price_limit_ratio_x' => $dto->riskParameters->priceLimitRatioX,
                'price_limit_ratio_y' => $dto->riskParameters->priceLimitRatioY,
            ]
        );
    }
} 