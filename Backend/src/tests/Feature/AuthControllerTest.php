<?php

namespace Tests\Feature;

use App\Constants\Api;
use App\Models\User;
use App\Models\UserType;
use App\Repositories\UserTypeRepository;
use Illuminate\Support\Facades\Auth;
use Illuminate\Support\Facades\DB;
use Symfony\Component\HttpFoundation\Response as ResponseAlias;
use Tests\helpers\TestUserHelpers;
use Tests\TestCase;

class AuthControllerTest extends TestCase
{
    /** @test */
    public function testRegister(): void
    {
        $url = Api::APIV1 . Api::USER . Api::REGISTER;
        $this->console('--------------------');
        $this->console(">> User registration\n POST $url \n");

        $data = [
            User::NICK_NAME => self::TEST_CLIENT_NICKNAME,
            User::EMAIL => self::TEST_CLIENT_EMAIL,
            User::PASSWORD => self::TEST_CLIENT_PASSWORD,
            User::PASSWORD_CONFIRMATION => self::TEST_CLIENT_PASSWORD
        ];

        $response = $this->sendPost($url, $data);
        $response->assertStatus(ResponseAlias::HTTP_CREATED);

        $this->assertDatabaseHas(User::TABLE_NAME, [
            User::NICK_NAME => self::TEST_CLIENT_NICKNAME,
            User::EMAIL => self::TEST_CLIENT_EMAIL,
        ]);

        $this->clearData();
    }

    /** @test */
    public function testLogin(): void
    {
        $url = Api::APIV1 . Api::USER . Api::LOGIN;
        $this->console('--------------------');
        $this->console(">> User login\n POST $url \n");

        $user = TestUserHelpers::createTestClientUser();

        $data = [
            User::EMAIL => self::TEST_CLIENT_EMAIL,
            User::PASSWORD => self::TEST_CLIENT_PASSWORD,
        ];

        $response = $this->sendPost($url, $data);
        $response->assertStatus(ResponseAlias::HTTP_OK);

        $this->clearData();
    }

    /** @test */
    public function testLogout(): void
    {
        $url = Api::APIV1 . Api::USER . Api::LOGOUT;
        $this->console('--------------------');
        $this->console(">> User logout\n POST $url \n");

        $user = TestUserHelpers::createTestClientUser();

        $token = Auth::attempt([
            'email' => self::TEST_CLIENT_EMAIL,
            'password' => self::TEST_CLIENT_PASSWORD,
        ]);

        $response = $this->sendPost($url, [], ['Authorization' => "Bearer $token"]);
        $response->assertStatus(ResponseAlias::HTTP_OK);

        $this->clearData();
    }

    /** @test */
    public function testAccount(): void
    {
        $url = Api::APIV1 . Api::USER . Api::ACCOUNT;
        $this->console('--------------------');
        $this->console(">> User account\n GET $url \n");

        $user = TestUserHelpers::createTestClientUser();

        $token = Auth::attempt([
            'email' => self::TEST_CLIENT_EMAIL,
            'password' => self::TEST_CLIENT_PASSWORD,
        ]);

        $response = $this->sendGet($url, ['Authorization' => "Bearer $token"]);
        $response->assertStatus(ResponseAlias::HTTP_OK);

        $userTypeClient = app(UserTypeRepository::class)->getByTypeAlias(UserType::ALIAS_CLIENT);

        $expectedData = [
            User::USER_TYPE_ID => $userTypeClient[UserType::ID],
            User::NICK_NAME => self::TEST_CLIENT_NICKNAME,
            User::EMAIL => self::TEST_CLIENT_EMAIL,
        ];
        $response->assertJsonStructure([
            User::ID,
            User::USER_TYPE_ID,
            User::NICK_NAME,
            User::EMAIL,
            User::EMAIL_VERIFIED_AT,
            User::DELETED_AT,
        ]);
        $response->assertJson($expectedData);
//        $response->assertExactJson($expectedData);

        $this->clearData();
    }

    protected function setUp(): void
    {
        parent::setUp();

        $this->clearData();
    }

    private function clearData(): void
    {
        DB::beginTransaction();
        TestUserHelpers::deleteTestClientUser();
        TestUserHelpers::deleteTestModeratorUser();
        TestUserHelpers::deleteTestAdmintUser();
        TestUserHelpers::deleteTestSaUser();
        DB::commit();
    }
}
