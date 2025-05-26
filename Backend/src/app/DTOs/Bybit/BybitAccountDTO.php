<?php

namespace App\DTOs\Bybit;

readonly class BybitAccountDTO
{
    /**
     * @param string $totalEquity
     * @param string $accountIMRate
     * @param string $totalMarginBalance
     * @param string $totalInitialMargin
     * @param string $accountType
     * @param string $totalAvailableBalance
     * @param string $accountMMRate
     * @param string $totalPerpUPL
     * @param string $totalWalletBalance
     * @param string $accountLTV
     * @param string $totalMaintenanceMargin
     * @param BybitCoinDTO[] $coins
     */
    public function __construct(
        public string $totalEquity,
        public string $accountIMRate,
        public string $totalMarginBalance,
        public string $totalInitialMargin,
        public string $accountType,
        public string $totalAvailableBalance,
        public string $accountMMRate,
        public string $totalPerpUPL,
        public string $totalWalletBalance,
        public string $accountLTV,
        public string $totalMaintenanceMargin,
        public array $coins
    ) {
    }

    public static function fromArray(array $data): self
    {
        return new self(
            totalEquity: $data['totalEquity'] ?? '0',
            accountIMRate: $data['accountIMRate'] ?? '0',
            totalMarginBalance: $data['totalMarginBalance'] ?? '0',
            totalInitialMargin: $data['totalInitialMargin'] ?? '0',
            accountType: $data['accountType'] ?? '',
            totalAvailableBalance: $data['totalAvailableBalance'] ?? '0',
            accountMMRate: $data['accountMMRate'] ?? '0',
            totalPerpUPL: $data['totalPerpUPL'] ?? '0',
            totalWalletBalance: $data['totalWalletBalance'] ?? '0',
            accountLTV: $data['accountLTV'] ?? '0',
            totalMaintenanceMargin: $data['totalMaintenanceMargin'] ?? '0',
            coins: array_map(
                fn(array $coinData) => BybitCoinDTO::fromArray($coinData),
                $data['coin'] ?? []
            )
        );
    }
} 