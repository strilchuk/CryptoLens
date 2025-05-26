<?php

namespace App\Http\Controllers;

use App\DTOs\Bybit\BybitWalletBalanceAccountDTO;
use App\DTOs\Bybit\BybitWalletBalanceCoinDTO;
use App\DTOs\Bybit\BybitWalletBalanceDTO;
use App\DTOs\Bybit\BybitWalletBalanceResponseDTO;
use App\Services\BybitService;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;

class BybitController extends Controller
{
    private BybitService $bybitService;

    public function __construct(BybitService $bybitService)
    {
        $this->bybitService = $bybitService;
    }

    /**
     * Получить баланс кошелька
     *
     * @param Request $request
     * @return JsonResponse
     */
    public function getWalletBalance(Request $request): JsonResponse
    {
        try {
            $balance = $this->bybitService->getWalletBalance($request->user());
            
            $response = new BybitWalletBalanceResponseDTO(
                accounts: array_map(
                    fn($account) => new BybitWalletBalanceAccountDTO(
                        totalEquity: $account->totalEquity,
                        accountType: $account->accountType,
                        totalWalletBalance: $account->totalWalletBalance,
                        totalAvailableBalance: $account->totalAvailableBalance,
                        coins: array_map(
                            fn($coin) => new BybitWalletBalanceCoinDTO(
                                coin: $coin->coin,
                                walletBalance: $coin->walletBalance,
                                equity: $coin->equity,
                                usdValue: $coin->usdValue,
                                unrealisedPnl: $coin->unrealisedPnl,
                                cumRealisedPnl: $coin->cumRealisedPnl
                            ),
                            $account->coins
                        )
                    ),
                    $balance->accounts
                )
            );

            return response()->json($response->toArray());
        } catch (\Exception $e) {
            return response()->json([
                'error' => $e->getMessage()
            ], 500);
        }
    }

    /**
     * Получить список доступных для торговли пар
     *
     * @param Request $request
     * @return JsonResponse
     */
    public function getInstruments(Request $request): JsonResponse
    {
        try {
            $category = $request->get('category', 'spot');
            $instruments = $this->bybitService->getInstrumentsFromDatabase($category);

            return response()->json($instruments->toArray());
        } catch (\Exception $e) {
            return response()->json([
                'error' => $e->getMessage()
            ], 500);
        }
    }
} 