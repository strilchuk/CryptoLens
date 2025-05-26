<?php

namespace Tests\helpers;

use App\Models\User;
use App\Models\UserType;
use Tests\TestCase;

/**
 *
 */
class TestUserHelpers
{
    /**
     * @return mixed
     */
    public static function createTestClientUser(): mixed
    {
        return User::factory()
            ->withNickName(TestCase::TEST_CLIENT_NICKNAME)
            ->withUserTypeByAlias(UserType::ALIAS_CLIENT)
            ->withEmail(TestCase::TEST_CLIENT_EMAIL)
            ->withPassword(TestCase::TEST_CLIENT_PASSWORD)
            ->create();
    }

    /**
     * @return mixed
     */
    public static function createTestModeratorUser(): mixed
    {
        return User::factory()
            ->withNickName(TestCase::TEST_MODERATOR_NICKNAME)
            ->withUserTypeByAlias(UserType::ALIAS_MODERATOR)
            ->withEmail(TestCase::TEST_MODERATOR_EMAIL)
            ->withPassword(TestCase::TEST_MODERATOR_PASSWORD)
            ->create();
    }

    /**
     * @return mixed
     */
    public static function createTestAdminUser(): mixed
    {
        return User::factory()
            ->withNickName(TestCase::TEST_ADMIN_NICKNAME)
            ->withUserTypeByAlias(UserType::ALIAS_ADMIN)
            ->withEmail(TestCase::TEST_ADMIN_EMAIL)
            ->withPassword(TestCase::TEST_ADMIN_PASSWORD)
            ->create();
    }

    /**
     * @return mixed
     */
    public static function createTestSaUser(): mixed
    {
        return User::factory()
            ->withNickName(TestCase::TEST_SA_NICKNAME)
            ->withUserTypeByAlias(UserType::ALIAS_SA)
            ->withEmail(TestCase::TEST_SA_EMAIL)
            ->withPassword(TestCase::TEST_SA_PASSWORD)
            ->create();
    }

    /**
     * @return void
     */
    public static function deleteTestClientUser(): void
    {
        User::query()->where([User::NICK_NAME => TestCase::TEST_CLIENT_NICKNAME])?->forceDelete();
    }

    /**
     * @return void
     */
    public static function deleteTestModeratorUser(): void
    {
        User::query()->where([User::NICK_NAME => TestCase::TEST_MODERATOR_NICKNAME])?->forceDelete();
    }

    /**
     * @return void
     */
    public static function deleteTestAdmintUser(): void
    {
        User::query()->where([User::NICK_NAME => TestCase::TEST_ADMIN_NICKNAME])?->forceDelete();
    }

    /**
     * @return void
     */
    public static function deleteTestSaUser(): void
    {
        User::query()->where([User::NICK_NAME => TestCase::TEST_SA_NICKNAME])?->forceDelete();
    }
}