<?php

namespace App\DTOs\Bybit;

class BybitTickersResponseDTO
{
    /**
     * @param string $category Категория
     * @param array $list Список тикеров
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
            list: array_map(fn($ticker) => (object) [
                'symbol' => $ticker['symbol'],
                'lastPrice' => $ticker['lastPrice'],
                'highPrice24h' => $ticker['highPrice24h'],
                'lowPrice24h' => $ticker['lowPrice24h'],
                'prevPrice24h' => $ticker['prevPrice24h'],
                'volume24h' => $ticker['volume24h'],
                'turnover24h' => $ticker['turnover24h'],
                'price24hPcnt' => $ticker['price24hPcnt'],
                'price1hPcnt' => $ticker['price1hPcnt'],
                'markPrice' => $ticker['markPrice'],
                'indexPrice' => $ticker['indexPrice'],
                'openInterest' => $ticker['openInterest'],
                'openInterestValue' => $ticker['openInterestValue'],
                'totalTurnover' => $ticker['totalTurnover'],
                'totalVolume' => $ticker['totalVolume'],
                'fundingRate' => $ticker['fundingRate'],
                'nextFundTime' => $ticker['nextFundTime'],
                'bid1Price' => $ticker['bid1Price'],
                'bid1Size' => $ticker['bid1Size'],
                'ask1Price' => $ticker['ask1Price'],
                'ask1Size' => $ticker['ask1Size'],
            ], $data['list'])
        );
    }
} 