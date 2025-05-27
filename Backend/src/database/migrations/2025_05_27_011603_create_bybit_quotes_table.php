<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::create('bybit_quotes', function (Blueprint $table) {
            $table->id();
            $table->string('symbol');
            $table->decimal('last_price', 20, 8);
            $table->decimal('bid_price', 20, 8);
            $table->decimal('ask_price', 20, 8);
            $table->decimal('volume_24h', 20, 8);
            $table->decimal('turnover_24h', 20, 8);
            $table->decimal('high_price_24h', 20, 8);
            $table->decimal('low_price_24h', 20, 8);
            $table->decimal('price_change_24h', 20, 8);
            $table->decimal('price_change_percent_24h', 20, 8);
            $table->timestamp('timestamp');
            $table->timestamps();

            $table->index(['symbol', 'timestamp']);
        });
    }

    public function down(): void
    {
        Schema::dropIfExists('bybit_quotes');
    }
}; 