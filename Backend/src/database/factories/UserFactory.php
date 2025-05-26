<?php

namespace Database\Factories;

use App\Models\User;
use App\Models\UserType;
use Faker\Provider\Uuid;
use Illuminate\Database\Eloquent\Factories\Factory;
use Illuminate\Support\Facades\Hash;
use Illuminate\Support\Str;

/**
 * @extends \Illuminate\Database\Eloquent\Factories\Factory<\App\Models\User>
 */
class UserFactory extends Factory
{
    /**
     * The current password being used by the factory.
     */
    protected static ?string $password;

    /**
     * The name of the factory's corresponding model.
     *
     * @var string
     */
    protected $model = User::class;

    protected static function boot()
    {
        parent::boot();

        static::creating(function ($model) {
            if (empty($model->{$model->getKeyName()})) {
                $model->{$model->getKeyName()} = (string)Str::uuid();
            }
        });
    }

    /**
     * Define the model's default state.
     *
     * @return array<string, mixed>
     */
    public function definition(): array
    {
        $randomUserTypeId = UserType::inRandomOrder()->value(UserType::ID);

        return [
            User::ID => Uuid::uuid(),
            User::USER_TYPE_ID => $randomUserTypeId,
            User::NICK_NAME => fake()->word,
            User::EMAIL => fake()->unique()->safeEmail(),
            User::EMAIL_VERIFIED_AT => now(),
            User::PASSWORD => static::$password ??= Hash::make('password'),
            User::REMEMBER_TOKEN => Str::random(10),
        ];
    }

    /**
     * Indicate that the model's email address should be unverified.
     */
    public function unverified(): static
    {
        return $this->state(fn(array $attributes) => [
            User::EMAIL_VERIFIED_AT => null,
        ]);
    }

    /**
     * Set the nick-name for the user.
     */
    public function withNickName(string $nickName): static
    {
        return $this->state(fn(array $attributes) => [
            User::NICK_NAME => $nickName,
        ]);
    }

    /**
     * Set the email for the user.
     */
    public function withEmail(string $email): static
    {
        return $this->state(fn(array $attributes) => [
            User::EMAIL => $email,
        ]);
    }

    /**
     * Set the password for the user.
     */
    public function withPassword(string $password): static
    {
        return $this->state(fn(array $attributes) => [
            User::PASSWORD => Hash::make($password),
        ]);
    }

    /**
     * Set the password for the user.
     */
    public function withUserTypeByAlias(string $alias): static
    {
        return $this->state(fn(array $attributes) => [
            User::USER_TYPE_ID => UserType::where(UserType::TYPE_ALIAS, $alias)->first()
        ]);
    }

}
