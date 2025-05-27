<?php

namespace App\DTOs\Bybit;

class BybitKlinesResponseDTO
{
    /**
     * @param string $category Категория
     * @param string $symbol Символ торговой пары
     * @param string $interval Интервал
     * @param array $list Список свечей
     */
    public function __construct(
        public readonly string $category,
        public readonly string $symbol,
        public readonly string $interval,
        public readonly array $list
    ) {
    }

    public static function fromArray(array $data): self
    {
        return new self(
            category: $data['category'],
            symbol: $data['symbol'],
            interval: $data['interval'],
            list: $data['list']
        );
    }
} 