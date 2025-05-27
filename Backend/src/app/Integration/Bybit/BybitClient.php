<?php

namespace App\Integration\Bybit;

use App\DTOs\Bybit\BybitInstrumentsResponseDTO;
use App\DTOs\Bybit\BybitResponseDTO;
use App\DTOs\Bybit\BybitTickersResponseDTO;
use App\DTOs\Bybit\BybitWalletBalanceDTO;
use App\DTOs\Bybit\BybitKlinesResponseDTO;
use App\DTOs\Bybit\BybitTradesResponseDTO;
use App\DTOs\Bybit\BybitOrderResponseDTO;
use App\DTOs\Bybit\BybitOrderListResponseDTO;
use App\DTOs\Bybit\BybitFeeRateResponseDTO;
use App\Models\BybitAccount;
use Illuminate\Http\Client\ConnectionException;
use Illuminate\Support\Facades\Http;
use Illuminate\Support\Facades\Log;

class BybitClient implements BybitClientInterface
{
    private string $baseUrl;
    private int $recvWindow;

    public function __construct()
    {
        $this->baseUrl = config('services.bybit.base_url', 'https://api.bybit.com');
        $this->recvWindow = config('services.bybit.recv_window', 5000);
    }

    /**
     * Получить баланс кошелька
     *
     * @param BybitAccount $account
     * @return BybitWalletBalanceDTO
     * @throws ConnectionException
     */
    public function getWalletBalance(BybitAccount $account): BybitWalletBalanceDTO
    {
        $timestamp = (string) round(microtime(true) * 1000);
        $queryParams = "accountType={$account->account_type}";
        
        $signature = $this->generateSignature($timestamp, $queryParams, $account);

        try {
            $response = Http::withHeaders([
                'X-BAPI-API-KEY' => $account->api_key,
                'X-BAPI-TIMESTAMP' => $timestamp,
                'X-BAPI-RECV-WINDOW' => (string) $this->recvWindow,
                'X-BAPI-SIGN' => $signature,
                'Content-Type' => 'application/json'
            ])->get("{$this->baseUrl}/v5/account/wallet-balance", [
                'accountType' => $account->account_type
            ]);

            $responseData = $response->json();
            $bybitResponse = BybitResponseDTO::fromArray($responseData);

            if (!$bybitResponse->isSuccess()) {
                throw new \Exception($bybitResponse->retMsg);
            }

            return BybitWalletBalanceDTO::fromArray($bybitResponse->result);
        } catch (\Exception $e) {
            Log::error('Bybit API Error: ' . $e->getMessage());
            throw $e;
        }
    }

    /**
     * Получить список доступных для торговли пар
     *
     * @param string $category Категория (spot, linear, inverse)
     * @return BybitInstrumentsResponseDTO
     * @throws ConnectionException
     */
    public function getInstruments(string $category = 'spot'): BybitInstrumentsResponseDTO
    {
        try {
            $response = Http::get("{$this->baseUrl}/v5/market/instruments-info", [
                'category' => $category
            ]);

            $responseData = $response->json();
            $bybitResponse = BybitResponseDTO::fromArray($responseData);

            if (!$bybitResponse->isSuccess()) {
                throw new \Exception($bybitResponse->retMsg);
            }

            return BybitInstrumentsResponseDTO::fromArray($bybitResponse->result);
        } catch (\Exception $e) {
            Log::error('Bybit API Error: ' . $e->getMessage());
            throw $e;
        }
    }

    public function getTickers(string $category = 'spot', ?string $symbol = null): BybitTickersResponseDTO
    {
        try {
            $query = ['category' => $category];
            if ($symbol) {
                $query['symbol'] = $symbol;
            }

            $response = Http::get("{$this->baseUrl}/v5/market/tickers", $query);
            $responseData = $response->json();
            $bybitResponse = BybitResponseDTO::fromArray($responseData);

            if (!$bybitResponse->isSuccess()) {
                throw new \Exception($bybitResponse->retMsg);
            }

            return BybitTickersResponseDTO::fromArray($bybitResponse->result);
        } catch (\Exception $e) {
            Log::error('Bybit API Error: ' . $e->getMessage());
            throw $e;
        }
    }

    /**
     * Получить исторические свечи
     *
     * @param string $category Категория (spot, linear, inverse)
     * @param string $symbol Символ торговой пары
     * @param string $interval Интервал (1,3,5,15,30,60,120,240,360,720,D,M,W)
     * @param int $limit Лимит записей (макс. 1000)
     * @param int|null $start Время начала в миллисекундах
     * @param int|null $end Время окончания в миллисекундах
     * @return BybitKlinesResponseDTO
     * @throws ConnectionException
     */
    public function getKlines(
        string $category,
        string $symbol,
        string $interval,
        int $limit = 200,
        ?int $start = null,
        ?int $end = null
    ): BybitKlinesResponseDTO {
        try {
            $query = [
                'category' => $category,
                'symbol' => $symbol,
                'interval' => $interval,
                'limit' => $limit
            ];

            if ($start) {
                $query['start'] = $start;
            }
            if ($end) {
                $query['end'] = $end;
            }

            $response = Http::get("{$this->baseUrl}/v5/market/kline", $query);
            $responseData = $response->json();
            $bybitResponse = BybitResponseDTO::fromArray($responseData);

            if (!$bybitResponse->isSuccess()) {
                throw new \Exception($bybitResponse->retMsg);
            }

            return BybitKlinesResponseDTO::fromArray($bybitResponse->result);
        } catch (\Exception $e) {
            Log::error('Bybit API Error: ' . $e->getMessage());
            throw $e;
        }
    }

    /**
     * Получить исторические сделки
     *
     * @param string $category Категория (spot, linear, inverse)
     * @param string $symbol Символ торговой пары
     * @param int $limit Лимит записей (макс. 1000)
     * @param string|null $orderId ID ордера
     * @return BybitTradesResponseDTO
     * @throws ConnectionException
     */
    public function getTrades(
        string $category,
        string $symbol,
        int $limit = 200,
        ?string $orderId = null
    ): BybitTradesResponseDTO {
        try {
            $query = [
                'category' => $category,
                'symbol' => $symbol,
                'limit' => $limit
            ];

            if ($orderId) {
                $query['orderId'] = $orderId;
            }

            $response = Http::get("{$this->baseUrl}/v5/market/recent-trade", $query);
            $responseData = $response->json();
            $bybitResponse = BybitResponseDTO::fromArray($responseData);

            if (!$bybitResponse->isSuccess()) {
                throw new \Exception($bybitResponse->retMsg);
            }

            return BybitTradesResponseDTO::fromArray($bybitResponse->result);
        } catch (\Exception $e) {
            Log::error('Bybit API Error: ' . $e->getMessage());
            throw $e;
        }
    }

    /**
     * Создать ордер
     *
     * @param BybitAccount $account
     * @param string $symbol Символ торговой пары
     * @param string $side Сторона (Buy/Sell)
     * @param string $orderType Тип ордера (Market/Limit)
     * @param string $qty Количество
     * @param string|null $price Цена (для лимитного ордера)
     * @param string $timeInForce Время действия (GTC/IOC/FOK)
     * @param string|null $orderLinkId Пользовательский ID ордера
     * @return BybitOrderResponseDTO
     * @throws ConnectionException
     */
    public function createOrder(
        BybitAccount $account,
        string $symbol,
        string $side,
        string $orderType,
        string $qty,
        ?string $price = null,
        string $timeInForce = 'GTC',
        ?string $orderLinkId = null
    ): BybitOrderResponseDTO {
        $timestamp = (string) round(microtime(true) * 1000);
        $payload = [
            'category' => 'spot',
            'symbol' => $symbol,
            'side' => $side,
            'orderType' => $orderType,
            'qty' => $qty,
            'timeInForce' => $timeInForce,
        ];

        if ($price) {
            $payload['price'] = $price;
        }
        if ($orderLinkId) {
            $payload['orderLinkId'] = $orderLinkId;
        }

        $signature = $this->generateSignature($timestamp, json_encode($payload), $account);

        try {
            $response = Http::withHeaders([
                'X-BAPI-API-KEY' => $account->api_key,
                'X-BAPI-TIMESTAMP' => $timestamp,
                'X-BAPI-RECV-WINDOW' => (string) $this->recvWindow,
                'X-BAPI-SIGN' => $signature,
                'Content-Type' => 'application/json'
            ])->post("{$this->baseUrl}/v5/order/create", $payload);

            $responseData = $response->json();
            $bybitResponse = BybitResponseDTO::fromArray($responseData);

            if (!$bybitResponse->isSuccess()) {
                throw new \Exception($bybitResponse->retMsg);
            }

            return BybitOrderResponseDTO::fromArray($bybitResponse->result);
        } catch (\Exception $e) {
            Log::error('Bybit API Error: ' . $e->getMessage());
            throw $e;
        }
    }

    /**
     * Изменить ордер
     *
     * @param BybitAccount $account
     * @param string $symbol Символ торговой пары
     * @param string $orderId ID ордера
     * @param string|null $price Новая цена
     * @param string|null $qty Новое количество
     * @return BybitOrderResponseDTO
     * @throws ConnectionException
     */
    public function amendOrder(
        BybitAccount $account,
        string $symbol,
        string $orderId,
        ?string $price = null,
        ?string $qty = null
    ): BybitOrderResponseDTO {
        $timestamp = (string) round(microtime(true) * 1000);
        $payload = [
            'category' => 'spot',
            'symbol' => $symbol,
            'orderId' => $orderId,
        ];

        if ($price) {
            $payload['price'] = $price;
        }
        if ($qty) {
            $payload['qty'] = $qty;
        }

        $signature = $this->generateSignature($timestamp, json_encode($payload), $account);

        try {
            $response = Http::withHeaders([
                'X-BAPI-API-KEY' => $account->api_key,
                'X-BAPI-TIMESTAMP' => $timestamp,
                'X-BAPI-RECV-WINDOW' => (string) $this->recvWindow,
                'X-BAPI-SIGN' => $signature,
                'Content-Type' => 'application/json'
            ])->post("{$this->baseUrl}/v5/order/amend", $payload);

            $responseData = $response->json();
            $bybitResponse = BybitResponseDTO::fromArray($responseData);

            if (!$bybitResponse->isSuccess()) {
                throw new \Exception($bybitResponse->retMsg);
            }

            return BybitOrderResponseDTO::fromArray($bybitResponse->result);
        } catch (\Exception $e) {
            Log::error('Bybit API Error: ' . $e->getMessage());
            throw $e;
        }
    }

    /**
     * Отменить ордер
     *
     * @param BybitAccount $account
     * @param string $symbol Символ торговой пары
     * @param string $orderId ID ордера
     * @return BybitOrderResponseDTO
     * @throws ConnectionException
     */
    public function cancelOrder(
        BybitAccount $account,
        string $symbol,
        string $orderId
    ): BybitOrderResponseDTO {
        $timestamp = (string) round(microtime(true) * 1000);
        $payload = [
            'category' => 'spot',
            'symbol' => $symbol,
            'orderId' => $orderId,
        ];

        $signature = $this->generateSignature($timestamp, json_encode($payload), $account);

        try {
            $response = Http::withHeaders([
                'X-BAPI-API-KEY' => $account->api_key,
                'X-BAPI-TIMESTAMP' => $timestamp,
                'X-BAPI-RECV-WINDOW' => (string) $this->recvWindow,
                'X-BAPI-SIGN' => $signature,
                'Content-Type' => 'application/json'
            ])->post("{$this->baseUrl}/v5/order/cancel", $payload);

            $responseData = $response->json();
            $bybitResponse = BybitResponseDTO::fromArray($responseData);

            if (!$bybitResponse->isSuccess()) {
                throw new \Exception($bybitResponse->retMsg);
            }

            return BybitOrderResponseDTO::fromArray($bybitResponse->result);
        } catch (\Exception $e) {
            Log::error('Bybit API Error: ' . $e->getMessage());
            throw $e;
        }
    }

    /**
     * Отменить все ордера
     *
     * @param BybitAccount $account
     * @param string $symbol Символ торговой пары
     * @return BybitOrderResponseDTO
     * @throws ConnectionException
     */
    public function cancelAllOrders(
        BybitAccount $account,
        string $symbol
    ): BybitOrderResponseDTO {
        $timestamp = (string) round(microtime(true) * 1000);
        $payload = [
            'category' => 'spot',
            'symbol' => $symbol,
        ];

        $signature = $this->generateSignature($timestamp, json_encode($payload), $account);

        try {
            $response = Http::withHeaders([
                'X-BAPI-API-KEY' => $account->api_key,
                'X-BAPI-TIMESTAMP' => $timestamp,
                'X-BAPI-RECV-WINDOW' => (string) $this->recvWindow,
                'X-BAPI-SIGN' => $signature,
                'Content-Type' => 'application/json'
            ])->post("{$this->baseUrl}/v5/order/cancel-all", $payload);

            $responseData = $response->json();
            $bybitResponse = BybitResponseDTO::fromArray($responseData);

            if (!$bybitResponse->isSuccess()) {
                throw new \Exception($bybitResponse->retMsg);
            }

            return BybitOrderResponseDTO::fromArray($bybitResponse->result);
        } catch (\Exception $e) {
            Log::error('Bybit API Error: ' . $e->getMessage());
            throw $e;
        }
    }

    /**
     * Получить открытые ордера
     *
     * @param BybitAccount $account
     * @param string $symbol Символ торговой пары
     * @param string|null $orderId ID ордера
     * @param int $limit Лимит записей
     * @return BybitOrderListResponseDTO
     * @throws ConnectionException
     */
    public function getOpenOrders(
        BybitAccount $account,
        string $symbol,
        ?string $orderId = null,
        int $limit = 20
    ): BybitOrderListResponseDTO {
        $timestamp = (string) round(microtime(true) * 1000);
        $queryParams = http_build_query([
            'category' => 'spot',
            'symbol' => $symbol,
            'orderId' => $orderId,
            'limit' => $limit,
        ]);

        $signature = $this->generateSignature($timestamp, $queryParams, $account);

        try {
            $response = Http::withHeaders([
                'X-BAPI-API-KEY' => $account->api_key,
                'X-BAPI-TIMESTAMP' => $timestamp,
                'X-BAPI-RECV-WINDOW' => (string) $this->recvWindow,
                'X-BAPI-SIGN' => $signature,
                'Content-Type' => 'application/json'
            ])->get("{$this->baseUrl}/v5/order/realtime", [
                'category' => 'spot',
                'symbol' => $symbol,
                'orderId' => $orderId,
                'limit' => $limit,
            ]);

            $responseData = $response->json();
            $bybitResponse = BybitResponseDTO::fromArray($responseData);

            if (!$bybitResponse->isSuccess()) {
                throw new \Exception($bybitResponse->retMsg);
            }

            return BybitOrderListResponseDTO::fromArray($bybitResponse->result);
        } catch (\Exception $e) {
            Log::error('Bybit API Error: ' . $e->getMessage());
            throw $e;
        }
    }

    /**
     * Получить ставки комиссии
     *
     * @param BybitAccount $account
     * @param string $category Категория (spot, linear, inverse)
     * @param string|null $symbol Символ торговой пары
     * @param string|null $baseCoin Базовая валюта
     * @return BybitFeeRateResponseDTO
     * @throws ConnectionException
     */
    public function getFeeRate(
        BybitAccount $account,
        string $category = 'spot',
        ?string $symbol = null,
        ?string $baseCoin = null
    ): BybitFeeRateResponseDTO {
        $timestamp = (string) round(microtime(true) * 1000);
        $queryParams = http_build_query(array_filter([
            'category' => $category,
            'symbol' => $symbol,
            'baseCoin' => $baseCoin,
        ]));

        $signature = $this->generateSignature($timestamp, $queryParams, $account);

        try {
            $response = Http::withHeaders([
                'X-BAPI-API-KEY' => $account->api_key,
                'X-BAPI-TIMESTAMP' => $timestamp,
                'X-BAPI-RECV-WINDOW' => (string) $this->recvWindow,
                'X-BAPI-SIGN' => $signature,
                'Content-Type' => 'application/json'
            ])->get("{$this->baseUrl}/v5/account/fee-rate", [
                'category' => $category,
                'symbol' => $symbol,
                'baseCoin' => $baseCoin,
            ]);

            $responseData = $response->json();
            $bybitResponse = BybitResponseDTO::fromArray($responseData);

            if (!$bybitResponse->isSuccess()) {
                throw new \Exception($bybitResponse->retMsg);
            }

            return BybitFeeRateResponseDTO::fromArray($bybitResponse->result);
        } catch (\Exception $e) {
            Log::error('Bybit API Error: ' . $e->getMessage());
            throw $e;
        }
    }

    /**
     * Генерация подписи для запроса
     *
     * @param string $timestamp
     * @param string $queryParams
     * @param BybitAccount $account
     * @return string
     */
    private function generateSignature(string $timestamp, string $queryParams, BybitAccount $account): string
    {
        $paramStr = $timestamp . $account->api_key . $this->recvWindow . $queryParams;
        return hash_hmac('sha256', $paramStr, $account->api_secret);
    }
} 