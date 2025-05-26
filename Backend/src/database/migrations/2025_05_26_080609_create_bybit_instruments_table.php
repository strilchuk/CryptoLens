<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    /**
     * Run the migrations.
     */
    public function up(): void
    {
        Schema::create('bybit_instruments', function (Blueprint $table) {
            $table->id();
            $table->string('symbol')->unique();
            $table->string('base_coin');
            $table->string('quote_coin');
            $table->string('innovation');
            $table->string('status');
            $table->string('margin_trading');
            $table->string('st_tag');
            
            // Lot Size Filter
            $table->string('base_precision');
            $table->string('quote_precision');
            $table->string('min_order_qty');
            $table->string('max_order_qty');
            $table->string('min_order_amt');
            $table->string('max_order_amt');
            
            // Price Filter
            $table->string('tick_size');
            
            // Risk Parameters
            $table->string('price_limit_ratio_x');
            $table->string('price_limit_ratio_y');
            
            $table->timestamps();
            $table->softDeletes();
        });
    }

    /**
     * Reverse the migrations.
     */
    public function down(): void
    {
        Schema::dropIfExists('bybit_instruments');
    }
}; 