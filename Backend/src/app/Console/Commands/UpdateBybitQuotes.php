<?php

namespace App\Console\Commands;

use App\Integration\Bybit\BybitClientInterface;
use App\Models\BybitInstrument;
use App\Models\BybitQuote;
use Illuminate\Console\Command;
use Illuminate\Support\Facades\Log;

class UpdateBybitQuotes extends Command
{
    protected $signature = 'bybit:update-quotes {--symbol= : Символ торговой пары}';
    protected $description = 'Обновить котировки Bybit';

    public function __construct(
        private readonly BybitClientInterface $bybitClient
    ) {
        parent::__construct();
    }

    public function handle(): int
    {
        $symbol = $this->option('symbol');
        Log::info('Запуск команды bybit:update-quotes', ['symbol' => $symbol]);

        try {
            if ($symbol) {
                $this->updateQuotesForSymbol($symbol);
            } else {
                $this->updateAllQuotes();
            }

            return self::SUCCESS;
        } catch (\Exception $e) {
            $this->error("Ошибка: {$e->getMessage()}");
            Log::error('Ошибка при обновлении котировок Bybit', [
                'error' => $e->getMessage(),
                'trace' => $e->getTraceAsString()
            ]);
            return self::FAILURE;
        }
    }

    private function updateQuotesForSymbol(string $symbol): void
    {
        $this->info("Получение котировок для {$symbol}...");
        $tickers = $this->bybitClient->getTickers('spot', $symbol);

        foreach ($tickers->list as $ticker) {
            $this->saveQuote($ticker);
        }
    }

    private function updateAllQuotes(): void
    {
        $instruments = BybitInstrument::where('status', 'Trading')->get();
        $total = $instruments->count();
        $this->info("Найдено {$total} активных инструментов");

        $bar = $this->output->createProgressBar($total);
        $bar->start();

        foreach ($instruments as $instrument) {
            try {
                $tickers = $this->bybitClient->getTickers('spot', $instrument->symbol);
                foreach ($tickers->list as $ticker) {
                    $this->saveQuote($ticker);
                }
            } catch (\Exception $e) {
                Log::error("Ошибка при обновлении котировок для {$instrument->symbol}", [
                    'error' => $e->getMessage()
                ]);
            }
            $bar->advance();
        }

        $bar->finish();
        $this->newLine();
    }

    private function saveQuote(object $ticker): void
    {
        BybitQuote::create([
            'symbol' => $ticker->symbol,
            'last_price' => $ticker->lastPrice,
            'bid_price' => $ticker->bid1Price,
            'ask_price' => $ticker->ask1Price,
            'volume_24h' => $ticker->volume24h,
            'turnover_24h' => $ticker->turnover24h,
            'high_price_24h' => $ticker->highPrice24h,
            'low_price_24h' => $ticker->lowPrice24h,
            'price_change_24h' => $ticker->lastPrice - $ticker->prevPrice24h,
            'price_change_percent_24h' => $ticker->price24hPcnt,
            'timestamp' => now(),
        ]);
    }
} 