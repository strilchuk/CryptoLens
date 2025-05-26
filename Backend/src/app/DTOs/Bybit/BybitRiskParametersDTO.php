<?php

namespace App\DTOs\Bybit;

readonly class BybitRiskParametersDTO
{
    public function __construct(
        public string $priceLimitRatioX,
        public string $priceLimitRatioY,
    ) {
    }

    public static function fromArray(array $data): self
    {
        return new self(
            priceLimitRatioX: $data['priceLimitRatioX'],
            priceLimitRatioY: $data['priceLimitRatioY'],
        );
    }

    public function toArray(): array
    {
        return [
            'priceLimitRatioX' => $this->priceLimitRatioX,
            'priceLimitRatioY' => $this->priceLimitRatioY,
        ];
    }
} 