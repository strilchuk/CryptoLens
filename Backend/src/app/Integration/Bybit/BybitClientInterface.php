<?php

namespace App\Integration\Bybit;

use App\DTOs\Bybit\BybitInstrumentsResponseDTO;
use App\DTOs\Bybit\BybitTickersResponseDTO;
use App\DTOs\Bybit\BybitWalletBalanceDTO;
use App\DTOs\Bybit\BybitOrderResponseDTO;
use App\DTOs\Bybit\BybitOrderListResponseDTO;
use App\Models\BybitAccount;
use App\DTOs\Bybit\BybitKlinesResponseDTO;
use App\DTOs\Bybit\BybitTradesResponseDTO;
use App\DTOs\Bybit\BybitFeeRateResponseDTO;

interface BybitClientInterface
{
    /**
     * Получить баланс кошелька
     *
     * @param BybitAccount $account
     * @return BybitWalletBalanceDTO
     */
    public function getWalletBalance(BybitAccount $account): BybitWalletBalanceDTO;

    /**
     * Получить список доступных для торговли пар
     *
     * @param string $category Категория (spot, linear, inverse)
     * @return BybitInstrumentsResponseDTO
     */
    public function getInstruments(string $category = 'spot'): BybitInstrumentsResponseDTO;

    /**
     * Получить текущие котировки
     *
     * @param string $category Категория (spot, linear, inverse)
     * @param string|null $symbol Символ торговой пары
     * @return BybitTickersResponseDTO
     */
    public function getTickers(string $category = 'spot', ?string $symbol = null): BybitTickersResponseDTO;

    /**
     * Получить исторические свечи
     *
     * @param string $category Категория (spot, linear, inverse)
     * @param string $symbol Символ торговой пары
     * @param string $interval Интервал (1,3,5,15,30,60,120,240,360,720,D,M,W)
     * @param int $limit Лимит записей (макс. 1000)
     * @param int|null $start Время начала в миллисекундах
     * @param int|null $end Время окончания в миллисекундах
     * @return BybitKlinesResponseDTO
     */
    public function getKlines(
        string $category,
        string $symbol,
        string $interval,
        int $limit = 200,
        ?int $start = null,
        ?int $end = null
    ): BybitKlinesResponseDTO;

    /**
     * Получить исторические сделки
     *
     * @param string $category Категория (spot, linear, inverse)
     * @param string $symbol Символ торговой пары
     * @param int $limit Лимит записей (макс. 1000)
     * @param string|null $orderId ID ордера
     * @return BybitTradesResponseDTO
     */
    public function getTrades(
        string $category,
        string $symbol,
        int $limit = 200,
        ?string $orderId = null
    ): BybitTradesResponseDTO;

    /**
     * Создать ордер
     *
     * @param BybitAccount $account
     * @param string $symbol Символ торговой пары
     * @param string $side Сторона (Buy/Sell)
     * @param string $orderType Тип ордера (Market/Limit)
     * @param string $qty Количество
     * @param string|null $price Цена (для лимитного ордера)
     * @param string $timeInForce Время действия (GTC/IOC/FOK)
     * @param string|null $orderLinkId Пользовательский ID ордера
     * @return BybitOrderResponseDTO
     */
    public function createOrder(
        BybitAccount $account,
        string $symbol,
        string $side,
        string $orderType,
        string $qty,
        ?string $price = null,
        string $timeInForce = 'GTC',
        ?string $orderLinkId = null
    ): BybitOrderResponseDTO;

    /**
     * Изменить ордер
     *
     * @param BybitAccount $account
     * @param string $symbol Символ торговой пары
     * @param string $orderId ID ордера
     * @param string|null $price Новая цена
     * @param string|null $qty Новое количество
     * @return BybitOrderResponseDTO
     */
    public function amendOrder(
        BybitAccount $account,
        string $symbol,
        string $orderId,
        ?string $price = null,
        ?string $qty = null
    ): BybitOrderResponseDTO;

    /**
     * Отменить ордер
     *
     * @param BybitAccount $account
     * @param string $symbol Символ торговой пары
     * @param string $orderId ID ордера
     * @return BybitOrderResponseDTO
     */
    public function cancelOrder(
        BybitAccount $account,
        string $symbol,
        string $orderId
    ): BybitOrderResponseDTO;

    /**
     * Отменить все ордера
     *
     * @param BybitAccount $account
     * @param string $symbol Символ торговой пары
     * @return BybitOrderResponseDTO
     */
    public function cancelAllOrders(
        BybitAccount $account,
        string $symbol
    ): BybitOrderResponseDTO;

    /**
     * Получить открытые ордера
     *
     * @param BybitAccount $account
     * @param string $symbol Символ торговой пары
     * @param string|null $orderId ID ордера
     * @param int $limit Лимит записей
     * @return BybitOrderListResponseDTO
     */
    public function getOpenOrders(
        BybitAccount $account,
        string $symbol,
        ?string $orderId = null,
        int $limit = 20
    ): BybitOrderListResponseDTO;

    /**
     * Получить ставки комиссии
     *
     * @param BybitAccount $account
     * @param string $category Категория (spot, linear, inverse)
     * @param string|null $symbol Символ торговой пары
     * @param string|null $baseCoin Базовая валюта
     * @return BybitFeeRateResponseDTO
     */
    public function getFeeRate(
        BybitAccount $account,
        string $category = 'spot',
        ?string $symbol = null,
        ?string $baseCoin = null
    ): BybitFeeRateResponseDTO;
} 