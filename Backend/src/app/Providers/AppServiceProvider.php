<?php

namespace App\Providers;

use App\Http\Controllers\AuthController;
use App\Services\Contracts\UserServiceInterface;
use App\Services\UserService;
use App\Integration\Bybit\BybitClient;
use App\Integration\Bybit\BybitClientInterface;
use Illuminate\Support\ServiceProvider;

class AppServiceProvider extends ServiceProvider
{
    /**
     * Register any application services.
     */
    public function register(): void
    {
        $this->app->when(AuthController::class)
            ->needs(UserServiceInterface::class)
            ->give(UserService::class);
        $this->app->bind(BybitClientInterface::class, BybitClient::class);
    }

    /**
     * Bootstrap any application services.
     */
    public function boot(): void
    {
        //
    }
}
