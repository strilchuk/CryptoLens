<?php

namespace App\Integration\Bybit;

use App\DTOs\Bybit\BybitInstrumentsResponseDTO;
use App\DTOs\Bybit\BybitTickersResponseDTO;
use App\DTOs\Bybit\BybitWalletBalanceDTO;
use App\Models\BybitAccount;
use App\DTOs\Bybit\BybitKlinesResponseDTO;
use App\DTOs\Bybit\BybitTradesResponseDTO;

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
} 