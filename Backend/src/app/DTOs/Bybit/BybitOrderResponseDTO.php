<?php

namespace App\DTOs\Bybit;

class BybitOrderResponseDTO
{
    /**
     * @param string $orderId ID ордера
     * @param string $orderLinkId Пользовательский ID ордера
     */
    public function __construct(
        public readonly string $orderId,
        public readonly string $orderLinkId
    ) {
    }

    public static function fromArray(array $data): self
    {
        return new self(
            orderId: $data['orderId'],
            orderLinkId: $data['orderLinkId']
        );
    }
} 