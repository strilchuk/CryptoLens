<?php

namespace App\DTOs\Bybit;

class BybitFeeRateResponseDTO
{
    /**
     * @param string $category Категория
     * @param array $list Список комиссий
     */
    public function __construct(
        public readonly string $category,
        public readonly array $list
    ) {
    }

    public static function fromArray(array $data): self
    {
        return new self(
            category: $data['category'],
            list: array_map(fn($fee) => (object) [
                'symbol' => $fee['symbol'],
                'takerFeeRate' => $fee['takerFeeRate'],
                'makerFeeRate' => $fee['makerFeeRate'],
            ], $data['list'])
        );
    }
} 