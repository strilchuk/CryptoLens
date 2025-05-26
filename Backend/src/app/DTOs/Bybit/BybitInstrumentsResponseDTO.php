<?php

namespace App\DTOs\Bybit;

readonly class BybitInstrumentsResponseDTO
{
    /**
     * @param BybitInstrumentDTO[] $list
     */
    public function __construct(
        public string $category,
        public array $list
    ) {
    }

    public static function fromArray(array $data): self
    {
        return new self(
            category: $data['category'],
            list: array_map(
                fn(array $item) => BybitInstrumentDTO::fromArray($item),
                $data['list']
            )
        );
    }

    public function toArray(): array
    {
        return [
            'category' => $this->category,
            'list' => array_map(
                fn(BybitInstrumentDTO $item) => $item->toArray(),
                $this->list
            )
        ];
    }
} 