<?php

namespace App\Console\Commands;

use App\Integration\Bybit\BybitClientInterface;
use App\Models\BybitAccount;
use Illuminate\Console\Command;
use Illuminate\Support\Facades\Log;

class PlaceBybitOrder extends Command
{
    protected $signature = 'bybit:place-order 
        {symbol : Торговая пара, например BTCUSDT} 
        {side : Buy или Sell} 
        {type : Market или Limit} 
        {qty : Количество} 
        {--price= : Цена для лимитного ордера}';

    protected $description = 'Разместить ордер на спотовой торговле Bybit';

    public function __construct(
        private readonly BybitClientInterface $bybitClient
    ) {
        parent::__construct();
    }

    public function handle(): int
    {
        $symbol = $this->argument('symbol');
        $side = $this->argument('side');
        $orderType = $this->argument('type');
        $qty = $this->argument('qty');
        $price = $this->option('price');

        Log::info("Запуск команды bybit:place-order", [
            'symbol' => $symbol,
            'side' => $side,
            'orderType' => $orderType,
            'qty' => $qty,
            'price' => $price
        ]);

        try {
            // Получаем аккаунт из базы данных
            $account = BybitAccount::firstOrFail();

            $this->info("Размещение ордера: {$side} {$orderType} {$qty} {$symbol}...");
            $response = $this->bybitClient->createOrder(
                $account,
                $symbol,
                $side,
                $orderType,
                $qty,
                $price
            );

            $this->info("Ордер создан: ID={$response->orderId}");

            // Проверяем открытые ордера
            $orders = $this->bybitClient->getOpenOrders($account, $symbol);
            $this->info("Открытые ордера: " . count($orders->list));

            return self::SUCCESS;
        } catch (\Exception $e) {
            $this->error("Ошибка: {$e->getMessage()}");
            Log::error('Ошибка при размещении ордера Bybit', [
                'error' => $e->getMessage(),
                'trace' => $e->getTraceAsString()
            ]);
            return self::FAILURE;
        }
    }
} 