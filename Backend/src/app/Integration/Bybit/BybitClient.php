<?php

namespace App\Integration\Bybit;

use App\DTOs\Bybit\BybitInstrumentsResponseDTO;
use App\DTOs\Bybit\BybitResponseDTO;
use App\DTOs\Bybit\BybitWalletBalanceDTO;
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