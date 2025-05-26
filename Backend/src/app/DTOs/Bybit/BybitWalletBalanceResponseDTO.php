<?php

namespace App\DTOs\Bybit;

readonly class BybitWalletBalanceResponseDTO
{
    /**
     * @param BybitWalletBalanceAccountDTO[] $accounts
     */
    public function __construct(
        public array $accounts
    ) {
    }

    public function toArray(): array
    {
        return [
            'accounts' => array_map(
                fn(BybitWalletBalanceAccountDTO $account) => $account->toArray(),
                $this->accounts
            )
        ];
    }
} 