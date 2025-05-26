<?php

use App\Models\User;
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
        Schema::create(User::TABLE_NAME, function (Blueprint $table) {
            $table->uuid(User::ID)->primary();

            $table->uuid(User::USER_TYPE_ID)->nullable();
            $table->foreign(User::USER_TYPE_ID)
                ->references(UserType::ID)
                ->on(UserType::TABLE_NAME)
                ->cascadeOnUpdate()->nullOnDelete();

            $table->string(User::NICK_NAME, 100)->nullable();
            $table->string(User::EMAIL, 100)->unique();
            $table->timestamp(User::EMAIL_VERIFIED_AT)->nullable();
            $table->string(User::PASSWORD, 255);
            $table->rememberToken();
            $table->timestamps();
            $table->softDeletes();
        });
    }

    /**
     * Reverse the migrations.
     */
    public function down(): void
    {
        Schema::dropIfExists(User::TABLE_NAME);
    }
};
