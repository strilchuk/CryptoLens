<?php

namespace App\DTOs\Bybit;

class BybitOrderListResponseDTO
{
    /**
     * @param array $list Список ордеров
     */
    public function __construct(
        public readonly array $list
    ) {
    }

    public static function fromArray(array $data): self
    {
        return new self(
            list: array_map(fn($order) => (object) [
                'orderId' => $order['orderId'],
                'orderLinkId' => $order['orderLinkId'],
                'symbol' => $order['symbol'],
                'side' => $order['side'],
                'orderType' => $order['orderType'],
                'price' => $order['price'],
                'qty' => $order['qty'],
                'timeInForce' => $order['timeInForce'],
                'orderStatus' => $order['orderStatus'],
                'leavesQty' => $order['leavesQty'],
                'cumExecQty' => $order['cumExecQty'],
                'cumExecValue' => $order['cumExecValue'],
                'cumExecFee' => $order['cumExecFee'],
                'createTime' => $order['createTime'],
                'updateTime' => $order['updateTime'],
            ], $data['list'])
        );
    }
} 