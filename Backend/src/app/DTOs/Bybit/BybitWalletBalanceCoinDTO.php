<?php

namespace App\DTOs\Bybit;

readonly class BybitWalletBalanceCoinDTO
{
    public function __construct(
        public string $coin,
        public string $walletBalance,
        public string $equity,
        public string $usdValue,
        public string $unrealisedPnl,
        public string $cumRealisedPnl
    ) {
    }

    public function toArray(): array
    {
        return [
            'coin' => $this->coin,
            'walletBalance' => $this->walletBalance,
            'equity' => $this->equity,
            'usdValue' => $this->usdValue,
            'unrealisedPnl' => $this->unrealisedPnl,
            'cumRealisedPnl' => $this->cumRealisedPnl
        ];
    }
} 