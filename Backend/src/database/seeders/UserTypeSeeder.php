<?php

namespace Database\Seeders;

use App\Models\UserType;
use Faker\Provider\Uuid;
use Illuminate\Database\Seeder;

class UserTypeSeeder extends Seeder
{
    /**
     * Run the database seeds.
     */
    public function run(): void
    {
        $userTypes = UserType::query()->get()->count();

        $data = [
            [
                UserType::ID => Uuid::uuid(),
                UserType::NAME => 'Супер администратор',
                UserType::TYPE_ALIAS => UserType::ALIAS_SA,
                UserType::IS_SA => true,
                UserType::IS_ADMIN => true,
                UserType::IS_MODERATOR => true,
                UserType::IS_CLIENT => true,
            ],
            [
                UserType::ID => Uuid::uuid(),
                UserType::NAME => 'Администратор',
                UserType::TYPE_ALIAS => UserType::ALIAS_ADMIN,
                UserType::IS_SA => false,
                UserType::IS_ADMIN => true,
                UserType::IS_MODERATOR => true,
                UserType::IS_CLIENT => false,
            ],
            [
                UserType::ID => Uuid::uuid(),
                UserType::NAME => 'Модератор',
                UserType::TYPE_ALIAS => UserType::ALIAS_MODERATOR,
                UserType::IS_SA => false,
                UserType::IS_ADMIN => false,
                UserType::IS_MODERATOR => true,
                UserType::IS_CLIENT => false,
            ],
            [
                UserType::ID => Uuid::uuid(),
                UserType::NAME => 'Клиент',
                UserType::TYPE_ALIAS => UserType::ALIAS_CLIENT,
                UserType::IS_SA => false,
                UserType::IS_ADMIN => false,
                UserType::IS_MODERATOR => true,
                UserType::IS_CLIENT => false,
            ],
        ];

        if ($userTypes === 0) {
            UserType::query()->insert($data);
        }
    }
}
