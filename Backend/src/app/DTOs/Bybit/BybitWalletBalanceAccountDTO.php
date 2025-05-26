<?php

namespace App\DTOs\Bybit;

readonly class BybitWalletBalanceAccountDTO
{
    /**
     * @param string $totalEquity
     * @param string $accountType
     * @param string $totalWalletBalance
     * @param string $totalAvailableBalance
     * @param BybitWalletBalanceCoinDTO[] $coins
     */
    public function __construct(
        public string $totalEquity,
        public string $accountType,
        public string $totalWalletBalance,
        public string $totalAvailableBalance,
        public array $coins
    ) {
    }

    public function toArray(): array
    {
        return [
            'totalEquity' => $this->totalEquity,
            'accountType' => $this->accountType,
            'totalWalletBalance' => $this->totalWalletBalance,
            'totalAvailableBalance' => $this->totalAvailableBalance,
            'coins' => array_map(
                fn(BybitWalletBalanceCoinDTO $coin) => $coin->toArray(),
                $this->coins
            )
        ];
    }
} 