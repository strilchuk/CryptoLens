<?php

namespace App\Services;

use App\DTOs\Bybit\BybitInstrumentDTO;
use App\DTOs\Bybit\BybitInstrumentsResponseDTO;
use App\DTOs\Bybit\BybitWalletBalanceDTO;
use App\Integration\Bybit\BybitClientInterface;
use App\Models\BybitAccount;
use App\Models\BybitInstrument;
use App\Models\User;
use Illuminate\Http\Client\ConnectionException;

class BybitService
{
    public function __construct(
        private readonly BybitClientInterface $bybitClient
    ) {
    }

    /**
     * Получить баланс кошелька
     *
     * @param User $user
     * @return BybitWalletBalanceDTO
     * @throws ConnectionException
     * @throws \Exception
     */
    public function getWalletBalance(User $user): BybitWalletBalanceDTO
    {
        $account = $user->activeBybitAccount();

        if (!$account) {
            throw new \Exception('No active Bybit account found');
        }

        return $this->bybitClient->getWalletBalance($account);
    }

    /**
     * Получить список доступных для торговли пар
     *
     * @param string $category Категория (spot, linear, inverse)
     * @return BybitInstrumentsResponseDTO
     * @throws ConnectionException
     */
    public function getInstruments(string $category = 'spot'): BybitInstrumentsResponseDTO
    {
        return $this->bybitClient->getInstruments($category);
    }

    /**
     * Получить список доступных для торговли пар из базы данных
     *
     * @param string $category Категория (spot, linear, inverse)
     * @return BybitInstrumentsResponseDTO
     */
    public function getInstrumentsFromDatabase(string $category = 'spot'): BybitInstrumentsResponseDTO
    {
        $instruments = BybitInstrument::all()->map(function (BybitInstrument $instrument) {
            return new BybitInstrumentDTO(
                symbol: $instrument->symbol,
                baseCoin: $instrument->base_coin,
                quoteCoin: $instrument->quote_coin,
                innovation: $instrument->innovation,
                status: $instrument->status,
                marginTrading: $instrument->margin_trading,
                stTag: $instrument->st_tag,
                lotSizeFilter: new \App\DTOs\Bybit\BybitLotSizeFilterDTO(
                    basePrecision: $instrument->base_precision,
                    quotePrecision: $instrument->quote_precision,
                    minOrderQty: $instrument->min_order_qty,
                    maxOrderQty: $instrument->max_order_qty,
                    minOrderAmt: $instrument->min_order_amt,
                    maxOrderAmt: $instrument->max_order_amt,
                ),
                priceFilter: new \App\DTOs\Bybit\BybitPriceFilterDTO(
                    tickSize: $instrument->tick_size,
                ),
                riskParameters: new \App\DTOs\Bybit\BybitRiskParametersDTO(
                    priceLimitRatioX: $instrument->price_limit_ratio_x,
                    priceLimitRatioY: $instrument->price_limit_ratio_y,
                ),
            );
        })->toArray();

        return new BybitInstrumentsResponseDTO(
            category: $category,
            list: $instruments
        );
    }
} 