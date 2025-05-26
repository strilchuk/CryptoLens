<?php

namespace App\Console;

use App\Console\Commands\UpdateBybitInstruments;
use Illuminate\Console\Scheduling\Schedule;
use Illuminate\Foundation\Console\Kernel as ConsoleKernel;
use Illuminate\Support\Facades\Log;

class Kernel extends ConsoleKernel
{
    protected $commands = [
        UpdateBybitInstruments::class,
    ];

    protected function schedule(Schedule $schedule): void
    {
        Log::info('Проверка расписания...', [
            'time' => now()->toDateTimeString(),
            'commands' => $this->commands,
            'timezone' => config('app.timezone'),
        ]);

        $schedulerLogPath = storage_path('logs/scheduler.log');
        Log::info('Попытка записи в scheduler.log', ['path' => $schedulerLogPath]);

        $schedule->command('bybit:update-instruments')
            ->everyMinute()
            ->withoutOverlapping()
            ->appendOutputTo($schedulerLogPath)
            ->before(function () {
                Log::info('Подготовка к запуску bybit:update-instruments', ['time' => now()->toDateTimeString()]);
            })
            ->after(function () {
                Log::info('Завершение bybit:update-instruments', ['time' => now()->toDateTimeString()]);
            });

        Log::info('Запланированные команды', [
            'scheduled' => array_map(fn($event) => [
                'command' => $event->command,
                'expression' => $event->expression,
                'mutex' => $event->mutexName(),
            ], $schedule->events())
        ]);
    }

    protected function commands(): void
    {
        $this->load(__DIR__ . '/Commands');
        require base_path('routes/console.php');
    }
}