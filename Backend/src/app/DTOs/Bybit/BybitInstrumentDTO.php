<?php

namespace App\DTOs\Bybit;

readonly class BybitInstrumentDTO
{
    public function __construct(
        public string $symbol,
        public string $baseCoin,
        public string $quoteCoin,
        public string $innovation,
        public string $status,
        public string $marginTrading,
        public string $stTag,
        public BybitLotSizeFilterDTO $lotSizeFilter,
        public BybitPriceFilterDTO $priceFilter,
        public BybitRiskParametersDTO $riskParameters,
    ) {
    }

    public static function fromArray(array $data): self
    {
        return new self(
            symbol: $data['symbol'],
            baseCoin: $data['baseCoin'],
            quoteCoin: $data['quoteCoin'],
            innovation: $data['innovation'],
            status: $data['status'],
            marginTrading: $data['marginTrading'],
            stTag: $data['stTag'],
            lotSizeFilter: BybitLotSizeFilterDTO::fromArray($data['lotSizeFilter']),
            priceFilter: BybitPriceFilterDTO::fromArray($data['priceFilter']),
            riskParameters: BybitRiskParametersDTO::fromArray($data['riskParameters']),
        );
    }

    public function toArray(): array
    {
        return [
            'symbol' => $this->symbol,
            'baseCoin' => $this->baseCoin,
            'quoteCoin' => $this->quoteCoin,
            'innovation' => $this->innovation,
            'status' => $this->status,
            'marginTrading' => $this->marginTrading,
            'stTag' => $this->stTag,
            'lotSizeFilter' => $this->lotSizeFilter->toArray(),
            'priceFilter' => $this->priceFilter->toArray(),
            'riskParameters' => $this->riskParameters->toArray(),
        ];
    }
} 