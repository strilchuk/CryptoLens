<?php

use App\Models\UserType;
use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration {
    /**
     * Run the migrations.
     */
    public function up(): void
    {
        Schema::create(UserType::TABLE_NAME, function (Blueprint $table) {
            $table->uuid(UserType::ID)->primary();
            $table->string(UserType::NAME, 100)->nullable();
            $table->string(UserType::TYPE_ALIAS, 30)->nullable();
            $table->boolean(UserType::IS_SA)->default(false);
            $table->boolean(UserType::IS_ADMIN)->default(false);
            $table->boolean(UserType::IS_MODERATOR)->default(false);
            $table->boolean(UserType::IS_CLIENT)->default(false);
            $table->softDeletes();
        });
    }

    /**
     * Reverse the migrations.
     */
    public function down(): void
    {
        Schema::dropIfExists(UserType::TABLE_NAME);
    }
};
