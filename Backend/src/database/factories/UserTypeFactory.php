<?php

namespace Database\Factories;

use App\Models\UserType;
use Faker\Provider\Uuid;
use Illuminate\Database\Eloquent\Factories\Factory;
use Illuminate\Support\Str;

/**
 * @extends \Illuminate\Database\Eloquent\Factories\Factory<\App\Models\UserType>
 */
class UserTypeFactory extends Factory
{
    /**
     * The name of the factory's corresponding model.
     *
     * @var string
     */
    protected $model = UserType::class;

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
        return [
            UserType::ID => Uuid::uuid(),
            UserType::NAME => $this->faker->name,
            UserType::TYPE_ALIAS => Str::lower($this->faker->word),
            UserType::IS_SA => $this->faker->boolean,
            UserType::IS_ADMIN => $this->faker->boolean,
            UserType::IS_MODERATOR => $this->faker->boolean,
            UserType::IS_CLIENT => $this->faker->boolean,
        ];
    }
}
