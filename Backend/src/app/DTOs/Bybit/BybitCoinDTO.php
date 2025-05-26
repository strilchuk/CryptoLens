<?php

namespace App\DTOs\Bybit;

readonly class BybitCoinDTO
{
    public function __construct(
        public string $availableToBorrow,
        public string $bonus,
        public string $accruedInterest,
        public string $availableToWithdraw,
        public string $totalOrderIM,
        public string $equity,
        public string $totalPositionMM,
        public string $usdValue,
        public string $unrealisedPnl,
        public bool $collateralSwitch,
        public string $spotHedgingQty,
        public string $borrowAmount,
        public string $totalPositionIM,
        public string $walletBalance,
        public string $cumRealisedPnl,
        public string $locked,
        public bool $marginCollateral,
        public string $coin
    ) {
    }

    public static function fromArray(array $data): self
    {
        return new self(
            availableToBorrow: $data['availableToBorrow'] ?? '',
            bonus: $data['bonus'] ?? '0',
            accruedInterest: $data['accruedInterest'] ?? '0',
            availableToWithdraw: $data['availableToWithdraw'] ?? '',
            totalOrderIM: $data['totalOrderIM'] ?? '0',
            equity: $data['equity'] ?? '0',
            totalPositionMM: $data['totalPositionMM'] ?? '0',
            usdValue: $data['usdValue'] ?? '0',
            unrealisedPnl: $data['unrealisedPnl'] ?? '0',
            collateralSwitch: $data['collateralSwitch'] ?? false,
            spotHedgingQty: $data['spotHedgingQty'] ?? '0',
            borrowAmount: $data['borrowAmount'] ?? '0',
            totalPositionIM: $data['totalPositionIM'] ?? '0',
            walletBalance: $data['walletBalance'] ?? '0',
            cumRealisedPnl: $data['cumRealisedPnl'] ?? '0',
            locked: $data['locked'] ?? '0',
            marginCollateral: $data['marginCollateral'] ?? false,
            coin: $data['coin'] ?? ''
        );
    }
} 