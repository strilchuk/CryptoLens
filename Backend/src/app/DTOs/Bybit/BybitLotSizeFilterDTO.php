<?php

namespace App\DTOs\Bybit;

readonly class BybitLotSizeFilterDTO
{
    public function __construct(
        public string $basePrecision,
        public string $quotePrecision,
        public string $minOrderQty,
        public string $maxOrderQty,
        public string $minOrderAmt,
        public string $maxOrderAmt,
    ) {
    }

    public static function fromArray(array $data): self
    {
        return new self(
            basePrecision: $data['basePrecision'],
            quotePrecision: $data['quotePrecision'],
            minOrderQty: $data['minOrderQty'],
            maxOrderQty: $data['maxOrderQty'],
            minOrderAmt: $data['minOrderAmt'],
            maxOrderAmt: $data['maxOrderAmt'],
        );
    }

    public function toArray(): array
    {
        return [
            'basePrecision' => $this->basePrecision,
            'quotePrecision' => $this->quotePrecision,
            'minOrderQty' => $this->minOrderQty,
            'maxOrderQty' => $this->maxOrderQty,
            'minOrderAmt' => $this->minOrderAmt,
            'maxOrderAmt' => $this->maxOrderAmt,
        ];
    }
} 