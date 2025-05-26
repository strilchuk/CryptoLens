<?php

namespace Database\Seeders;

use App\Models\User;
use App\Models\UserType;
use Faker\Provider\Uuid;
use Illuminate\Database\Seeder;
use Illuminate\Support\Facades\Hash;

class UserSeeder extends Seeder
{
    /**
     * Run the database seeds.
     */
    public function run(): void
    {
        $usersCount = User::query()->get()->count();
        $isProd =  in_array(config('app.env'), ['production', 'prod']);;

        if ($usersCount === 0 && !$isProd) {
            User::factory()->count(100)->create([
                'password' => Hash::make('43562345'),
            ]);
        }
    }
}
