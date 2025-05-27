<?php

use App\Constants\Api;
use App\Constants\Common;
use App\Http\Controllers\AuthController;
use App\Http\Controllers\BybitController;
use Illuminate\Support\Facades\Route;

/*
|--------------------------------------------------------------------------
| API Routes
|--------------------------------------------------------------------------
|
| Here is where you can register API routes for your application. These
| routes are loaded by the RouteServiceProvider and all of them will
| be assigned to the "api" middleware group. Make something great!
|
*/

Route::middleware('api')->prefix("v1")->group(function () {
    // Общедоступные маршруты
    Route::get(Api::STATUS, function () {
        return ['status' => Common::STATUS_SUCCESS];
    });
    Route::post(Api::USER . Api::REGISTER, [AuthController::class, 'register']);
    Route::post(Api::USER . Api::LOGIN, [AuthController::class, 'login']);
    Route::get('/bybit/instruments', [BybitController::class, 'getInstruments']);

    // Защищённые маршруты
    Route::middleware('auth:api')->group(function () {
        Route::post(Api::USER . Api::LOGOUT, [AuthController::class, 'logout']);
        Route::get(Api::USER . Api::ACCOUNT, [AuthController::class, 'me']);
        Route::get('/bybit/wallet-balance', [BybitController::class, 'getWalletBalance']);
        Route::get('/bybit/fee-rate', [BybitController::class, 'getFeeRate']);
    });
});