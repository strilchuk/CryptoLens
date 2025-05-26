<?php

namespace App\DTOs\Bybit;

readonly class BybitPriceFilterDTO
{
    public function __construct(
        public string $tickSize,
    ) {
    }

    public static function fromArray(array $data): self
    {
        return new self(
            tickSize: $data['tickSize'],
        );
    }

    public function toArray(): array
    {
        return [
            'tickSize' => $this->tickSize,
        ];
    }
} 