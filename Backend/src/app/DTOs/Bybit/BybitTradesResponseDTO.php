<?php

namespace App\DTOs\Bybit;

class BybitTradesResponseDTO
{
    /**
     * @param string $category Категория
     * @param string $symbol Символ торговой пары
     * @param array $list Список сделок
     */
    public function __construct(
        public readonly string $category,
        public readonly string $symbol,
        public readonly array $list
    ) {
    }

    public static function fromArray(array $data): self
    {
        return new self(
            category: $data['category'],
            symbol: $data['symbol'],
            list: $data['list']
        );
    }
} 