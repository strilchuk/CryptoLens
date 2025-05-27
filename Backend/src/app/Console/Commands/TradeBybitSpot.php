<?php

namespace App\Console\Commands;

use App\Integration\Bybit\BybitClientInterface;
use App\Models\BybitAccount;
use App\Models\BybitInstrument;
use Illuminate\Console\Command;
use Illuminate\Support\Facades\Log;

class TradeBybitSpot extends Command
{
    protected $signature = 'bybit:trade-spot {symbol : Торговая пара, например BTCUSDT}';
    protected $description = 'Автоматическая спотовая торговля на Bybit по стратегии EMA+RSI+ATR';

    public function __construct(
        private readonly BybitClientInterface $bybitClient
    ) {
        parent::__construct();
    }

    public function handle(): int
    {
        $symbol = $this->argument('symbol');
        Log::info("Запуск торговой стратегии для {$symbol}");

        try {
            $account = BybitAccount::firstOrFail();

            // Получение комиссий
            $feeRate = $this->bybitClient->getFeeRate($account, 'spot', $symbol);
            $totalFee = (float)$feeRate->list[0]->takerFeeRate * 2; // Вход + выход

            // Получение свечей (5-минутный таймфрейм, 30 свечей для EMA, RSI, ATR)
            $klines = $this->bybitClient->getKlines('spot', $symbol, '5', 30);
            $candles = array_reverse($klines->list); // Свечи от новых к старым

            // Расчет индикаторов
            $prices = array_map(fn($candle) => (float)$candle->close, $candles);
            $ema9 = $this->calculateEMA($prices, 9);
            $ema21 = $this->calculateEMA($prices, 21);
            $rsi = $this->calculateRSI($prices, 14);
            $atr = $this->calculateATR($candles, 14);
            $volumes = array_map(fn($candle) => (float)$candle->volume, $candles);
            $avgVolume = array_sum(array_slice($volumes, 0, 20)) / 20;

            // Проверка тренда и условий входа
            $currentPrice = $prices[0];
            $localLow = min(array_slice($prices, 0, 10));
            $isUptrend = $ema9[0] > $ema21[0];
            $isValidRSI = $rsi[0] < 45 && $rsi[0] > 30;
            $isVolatile = $atr > 0.005 * $currentPrice;
            $hasVolume = $volumes[0] > $avgVolume;

            Log::info('Анализ рынка', [
                'symbol' => $symbol,
                'price' => $currentPrice,
                'ema9' => $ema9[0],
                'ema21' => $ema21[0],
                'rsi' => $rsi[0],
                'atr' => $atr,
                'volume' => $volumes[0],
                'avgVolume' => $avgVolume,
            ]);

            if ($isUptrend && $isValidRSI && $isVolatile && $hasVolume && $currentPrice > $localLow) {
                $this->info("Сигнал на покупку: {$symbol} по {$currentPrice}");

                // Расчет позиции (1% риска от баланса)
                $balance = $this->bybitClient->getWalletBalance($account);
                $usdtBalance = (float)$balance->accounts[0]->totalEquity; // Предполагаем USDT
                $riskAmount = 0.01 * $usdtBalance; // 1% риска
                $slDistance = 2 * $atr;
                $positionSize = $riskAmount / $slDistance; // Кол-во в базовой валюте
                $qty = number_format($positionSize, 6); // Округляем до 6 знаков

                // Проверка минимального объема ордера
                $instrument = BybitInstrument::where('symbol', $symbol)->first();
                if ($qty < (float)$instrument->min_order_qty) {
                    $this->error("Недостаточный объем: {$qty} < {$instrument->min_order_qty}");
                    return self::FAILURE;
                }

                // Размещение рыночного ордера на покупку
                $order = $this->bybitClient->createOrder($account, $symbol, 'Buy', 'Market', $qty);
                $this->info("Ордер создан: ID={$order->orderId}");

                // Установка SL и TP
                $slPrice = number_format($currentPrice - $slDistance, 2);
                $tpPrice = number_format($currentPrice + 2 * $slDistance, 2);

                // Размещение лимитного ордера для TP
                $tpOrder = $this->bybitClient->createOrder($account, $symbol, 'Sell', 'Limit', $qty, $tpPrice);
                $this->info("TP ордер создан: ID={$tpOrder->orderId}");

                // Проверка безубыточности
                $this->monitorPosition($account, $symbol, $order->orderId, $currentPrice, $slPrice, $tpPrice, $qty);
            } else {
                $this->info("Сигналов для входа нет");
            }

            return self::SUCCESS;
        } catch (\Exception $e) {
            $this->error("Ошибка: {$e->getMessage()}");
            Log::error('Ошибка в торговой стратегии', [
                'error' => $e->getMessage(),
                'trace' => $e->getTraceAsString()
            ]);
            return self::FAILURE;
        }
    }

    private function calculateEMA(array $prices, int $period): array
    {
        $multiplier = 2 / ($period + 1);
        $ema = [];
        $ema[] = array_sum(array_slice($prices, 0, $period)) / $period; // Начальная SMA

        for ($i = 1; $i < count($prices); $i++) {
            $ema[] = ($prices[$i] - $ema[$i - 1]) * $multiplier + $ema[$i - 1];
        }

        return array_reverse($ema); // От новых к старым
    }

    private function calculateRSI(array $prices, int $period): array
    {
        $rsi = [];
        $gains = $losses = [];

        for ($i = 1; $i < count($prices); $i++) {
            $diff = $prices[$i - 1] - $prices[$i];
            $gains[] = $diff > 0 ? $diff : 0;
            $losses[] = $diff < 0 ? -$diff : 0;
        }

        for ($i = $period; $i < count($gains); $i++) {
            $avgGain = array_sum(array_slice($gains, $i - $period, $period)) / $period;
            $avgLoss = array_sum(array_slice($losses, $i - $period, $period)) / $period;
            $rs = $avgLoss == 0 ? 100 : $avgGain / $avgLoss;
            $rsi[] = 100 - (100 / (1 + $rs));
        }

        return array_pad(array_reverse($rsi), count($prices), 0);
    }

    private function calculateATR(array $candles, int $period): float
    {
        $tr = [];
        foreach ($candles as $i => $candle) {
            if ($i == count($candles) - 1) continue; // Пропускаем последнюю свечу
            $high = (float)$candle->high;
            $low = (float)$candle->low;
            $prevClose = (float)$candles[$i + 1]->close;
            $tr[] = max($high - $low, abs($high - $prevClose), abs($low - $prevClose));
        }
        return array_sum(array_slice($tr, 0, $period)) / $period;
    }

    private function monitorPosition(
        BybitAccount $account,
        string $symbol,
        string $orderId,
        float $entryPrice,
        float $slPrice,
        float $tpPrice,
        string $qty
    ) {
        $this->info("Мониторинг позиции: {$symbol}, Вход={$entryPrice}, SL={$slPrice}, TP={$tpPrice}");

        while (true) {
            sleep(60); // Проверка каждую минуту
            $klines = $this->bybitClient->getKlines('spot', $symbol, '5', 1);
            $currentPrice = (float)$klines->list[0]->close;

            Log::info("Проверка позиции", ['symbol' => $symbol, 'currentPrice' => $currentPrice]);

            // Безубыточность: если достигли 50% TP, перемещаем SL в точку входа
            if ($currentPrice >= $entryPrice + ($tpPrice - $entryPrice) / 2) {
                $slPrice = $entryPrice;
                $this->info("Перемещён SL в точку безубыточности: {$slPrice}");
                // Частичное закрытие: продаём 30% позиции
                $partialQty = number_format((float)$qty * 0.3, 6);
                $this->bybitClient->createOrder($account, $symbol, 'Sell', 'Market', $partialQty);
                $qty = number_format((float)$qty * 0.7, 6); // Оставшиеся 70%
            }

            // Проверка выхода
            if ($currentPrice <= $slPrice) {
                $this->info("Сработал стоп-лосс: {$slPrice}");
                $this->bybitClient->createOrder($account, $symbol, 'Sell', 'Market', $qty);
                break;
            }

            if ($currentPrice >= $tpPrice) {
                $this->info("Сработал тейк-профит: {$tpPrice}");
                $this->bybitClient->createOrder($account, $symbol, 'Sell', 'Market', $qty);
                break;
            }

            // Проверка смены тренда
            $klines = $this->bybitClient->getKlines('spot', $symbol, '5', 10);
            $prices = array_map(fn($candle) => (float)$candle->close, array_reverse($klines->list));
            $ema9 = $this->calculateEMA($prices, 9);
            $ema21 = $this->calculateEMA($prices, 21);
            if ($ema9[0] < $ema21[0]) {
                $this->info("Тренд сменился на нисходящий, закрытие позиции");
                $this->bybitClient->createOrder($account, $symbol, 'Sell', 'Market', $qty);
                break;
            }
        }
    }
} 