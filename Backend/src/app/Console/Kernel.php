<?php

namespace App\Console;

use App\Console\Commands\TradeBybitSpot;
use App\Console\Commands\UpdateBybitInstruments;
use App\Console\Commands\UpdateBybitQuotes;
use App\Console\Commands\PlaceBybitOrder;
use Illuminate\Console\Scheduling\Schedule;
use Illuminate\Foundation\Console\Kernel as ConsoleKernel;
use Illuminate\Support\Facades\Log;

class Kernel extends ConsoleKernel
{
    /**
     * The Artisan commands provided by your application.
     *
     * @var array
     */
    protected $commands = [
        UpdateBybitInstruments::class,
        UpdateBybitQuotes::class,
        PlaceBybitOrder::class,
        TradeBybitSpot::class,
    ];

    protected function schedule(Schedule $schedule): void
    {
//        Log::info('Проверка расписания...', [
//            'time' => now()->toDateTimeString(),
//            'commands' => $this->commands,
//            'timezone' => config('app.timezone'),
//        ]);

        $schedulerLogPath = storage_path('logs/scheduler.log');
//        Log::info('Попытка записи в scheduler.log', ['path' => $schedulerLogPath]);

        $schedule->command('bybit:update-instruments')
            ->hourly()
            ->withoutOverlapping()
            ->appendOutputTo(storage_path('logs/scheduler.log'))
            ->before(function () {
//                Log::info('Подготовка к запуску bybit:update-instruments', ['time' => now()->toDateTimeString()]);
            })
            ->after(function () {
//                Log::info('Завершение bybit:update-instruments', ['time' => now()->toDateTimeString()]);
            });

//            $schedule->command('bybit:update-quotes')
//            ->everyMinute()
//            ->withoutOverlapping()
//            ->appendOutputTo(storage_path('logs/scheduler.log'));

//        Log::info('Запланированные команды', [
//            'scheduled' => array_map(fn($event) => [
//                'command' => $event->command,
//                'expression' => $event->expression,
//                'mutex' => $event->mutexName(),
//            ], $schedule->events())
//        ]);


//        $schedule->command('bybit:trade-spot BTCUSDT')
//            ->everyFiveMinutes()
//            ->withoutOverlapping()
//            ->appendOutputTo($schedulerLogPath)
//            ->before(function () {
//                Log::info('Подготовка к запуску bybit:trade-spot');
//            })
//            ->after(function () {
//                Log::info('Завершение bybit:trade-spot');
//            });
    }

    protected function commands(): void
    {
        $this->load(__DIR__ . '/Commands');
        require base_path('routes/console.php');
    }
}