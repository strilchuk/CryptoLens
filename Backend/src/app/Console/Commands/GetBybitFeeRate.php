<?php

namespace App\Console\Commands;

use App\Integration\Bybit\BybitClient;
use App\Models\BybitAccount;
use Illuminate\Console\Command;
use Illuminate\Support\Facades\Log;

class GetBybitFeeRate extends Command
{
    protected $signature = 'bybit:fee-rate {symbol?} {--category=spot} {--base-coin=}';
    protected $description = 'Получить ставки комиссии для спотовой торговли на Bybit';

    public function handle(BybitClient $bybitClient)
    {
        $symbol = $this->argument('symbol');
        $category = $this->option('category');
        $baseCoin = $this->option('base-coin');

        $this->info("Получение комиссий для категории: {$category}");
        if ($symbol) {
            $this->info("Символ: {$symbol}");
        }
        if ($baseCoin) {
            $this->info("Базовая валюта: {$baseCoin}");
        }

        try {
            $account = BybitAccount::first();
            if (!$account) {
                $this->error('Аккаунт Bybit не найден');
                return 1;
            }

            $feeRate = $bybitClient->getFeeRate($account, $category, $symbol, $baseCoin);

            $this->info("\nРезультат:");
            $this->info("Категория: {$feeRate->category}");
            $this->info("\nКомиссии:");
            foreach ($feeRate->list as $fee) {
                $this->info("Символ: {$fee->symbol}");
                $this->info("Комиссия тейкера: {$fee->takerFeeRate}");
                $this->info("Комиссия мейкера: {$fee->makerFeeRate}");
                $this->info("---");
            }

            return 0;
        } catch (\Exception $e) {
            $this->error("Ошибка: {$e->getMessage()}");
            Log::error('Bybit Fee Rate Error: ' . $e->getMessage());
            return 1;
        }
    }
} 