<?php

namespace App\Console\Commands;

use App\Integration\Bybit\BybitClientInterface;
use App\Models\BybitInstrument;
use Illuminate\Console\Command;
use Illuminate\Support\Facades\Log;

class UpdateBybitInstruments extends Command
{
    protected $signature = 'bybit:update-instruments {--category=spot : Категория инструментов (spot, linear, inverse)}';
    protected $description = 'Обновить список инструментов Bybit';

    public function __construct(
        private readonly BybitClientInterface $bybitClient
    ) {
        parent::__construct();
    }

    public function handle(): int
    {
        $category = $this->option('category');
        Log::info("Запуск команды bybit:update-instruments", ['category' => $category]);
        try {
            $this->info("Получение списка инструментов категории {$category}...");
//            Log::info("Отправка запроса к Bybit API", ['category' => $category]);

            $response = $this->bybitClient->getInstruments($category);
            Log::info("Получен ответ от Bybit", ['count' => count($response->list)]);

            $count = count($response->list);
            $this->info("Найдено {$count} инструментов");
//            $bar = $this->output->createProgressBar($count);

            foreach ($response->list as $instrument) {
//                Log::debug("Обработка инструмента", ['symbol' => $instrument->symbol]);
                BybitInstrument::updateFromDTO($instrument);
//                $bar->advance();
            }

//            $bar->finish();
            $this->newLine();
            $this->info('Инструменты успешно обновлены');
            Log::info('Инструменты успешно обновлены');

            return self::SUCCESS;
        } catch (\Exception $e) {
            $this->error("Ошибка при обновлении инструментов: {$e->getMessage()}");
            Log::error('Ошибка при обновлении инструментов Bybit', [
                'error' => $e->getMessage(),
                'trace' => $e->getTraceAsString()
            ]);

            return self::FAILURE;
        }
    }
}