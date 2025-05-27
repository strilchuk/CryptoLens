<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\Relations\BelongsTo;

/**
 * @property int $id
 * @property string $symbol Символ торговой пары
 * @property float $last_price Последняя цена
 * @property float $bid_price Лучшая цена покупки
 * @property float $ask_price Лучшая цена продажи
 * @property float $volume_24h Объем торгов за 24 часа
 * @property float $turnover_24h Оборот за 24 часа
 * @property float $high_price_24h Максимальная цена за 24 часа
 * @property float $low_price_24h Минимальная цена за 24 часа
 * @property float $price_change_24h Изменение цены за 24 часа
 * @property float $price_change_percent_24h Процент изменения цены за 24 часа
 * @property \Carbon\Carbon $timestamp Время котировки
 * @property \Carbon\Carbon $created_at
 * @property \Carbon\Carbon $updated_at
 */
class BybitQuote extends Model
{
    protected $fillable = [
        'symbol',
        'last_price',
        'bid_price',
        'ask_price',
        'volume_24h',
        'turnover_24h',
        'high_price_24h',
        'low_price_24h',
        'price_change_24h',
        'price_change_percent_24h',
        'timestamp',
    ];

    protected $casts = [
        'last_price' => 'decimal:8',
        'bid_price' => 'decimal:8',
        'ask_price' => 'decimal:8',
        'volume_24h' => 'decimal:8',
        'turnover_24h' => 'decimal:8',
        'high_price_24h' => 'decimal:8',
        'low_price_24h' => 'decimal:8',
        'price_change_24h' => 'decimal:8',
        'price_change_percent_24h' => 'decimal:8',
        'timestamp' => 'datetime',
    ];

    public function instrument(): BelongsTo
    {
        return $this->belongsTo(BybitInstrument::class, 'symbol', 'symbol');
    }
} 