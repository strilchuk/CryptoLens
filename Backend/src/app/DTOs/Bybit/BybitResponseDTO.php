<?php

namespace App\DTOs\Bybit;

readonly class BybitResponseDTO
{
    public function __construct(
        public int $retCode,
        public string $retMsg,
        public array $result,
        public array $retExtInfo,
        public int $time
    ) {
    }

    public static function fromArray(array $data): self
    {
        return new self(
            retCode: $data['retCode'],
            retMsg: $data['retMsg'],
            result: $data['result'],
            retExtInfo: $data['retExtInfo'],
            time: $data['time']
        );
    }

    public function isSuccess(): bool
    {
        return $this->retCode === 0;
    }
} 