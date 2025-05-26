<?php

namespace App\DTOs\Bybit;

readonly class BybitWalletBalanceDTO
{
    /**
     * @param BybitAccountDTO[] $accounts
     */
    public function __construct(
        public array $accounts
    ) {
    }

    public static function fromArray(array $data): self
    {
        return new self(
            accounts: array_map(
                fn(array $accountData) => BybitAccountDTO::fromArray($accountData),
                $data['list'] ?? []
            )
        );
    }
} 