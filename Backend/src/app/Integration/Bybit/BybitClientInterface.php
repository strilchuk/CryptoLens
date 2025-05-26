<?php

namespace App\Integration\Bybit;

use App\DTOs\Bybit\BybitInstrumentsResponseDTO;
use App\DTOs\Bybit\BybitWalletBalanceDTO;
use App\Models\BybitAccount;

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
} 