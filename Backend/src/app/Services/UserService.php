<?php

namespace App\Services;

use App\Constants\Common;
use App\Models\User;
use App\Models\UserType;
use App\Repositories\UserRepository;
use App\Services\Contracts\UserServiceInterface;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Hash;
use Illuminate\Support\Facades\Lang;
use Illuminate\Support\Facades\Validator;

/**
 *
 */
class UserService implements UserServiceInterface
{
    /**
     * @var UserRepository
     */
    private UserRepository $userRepository;

    /**
     * @param UserRepository $userRepository
     */
    public function __construct(UserRepository $userRepository)
    {
        $this->userRepository = $userRepository;
    }

    /**
     * @param array $data
     * @return array
     */
    public function store(array $data): array
    {
        DB::beginTransaction();

        $validator = Validator::make($data, [
            User::USER_TYPE_ID => 'required|exists:' . UserType::TABLE_NAME . ',id',
            User::NICK_NAME => 'required|string|max:255',
            User::EMAIL => 'required|string|email|max:255|unique:users',
            User::PASSWORD => 'required|string|min:8|confirmed',
        ]);

        if ($validator->fails()) {
            DB::rollBack();
            return [
                Common::ERROR => Lang::get('errors.invalid_fields'),
                Common::MESSAGE => $validator->errors()->toArray(),
            ];
        }

        $data[User::CREATED_AT] = now()->toString();
        $data[User::UPDATED_AT] = now()->toString();
        $data[User::PASSWORD] = Hash::make($data[User::PASSWORD]);
        unset($data['password_confirmation']);

        $result = $this->userRepository->create($data);

        if (empty($result[User::ID])) {
            DB::rollBack();
            return $result;
        }

        DB::commit();
        return $result;
    }

    /**
     * @param string|null $id
     * @return bool
     */
    public function delete(?string $id): bool
    {
        // TODO: Implement delete() method.
    }

    /**
     * @param string|null $id
     * @return bool
     */
    public function restore(?string $id): bool
    {
        // TODO: Implement restore() method.
    }

    /**
     * @param array $params
     * @param int $size
     * @return array
     */
    public function pageableListByParams(array $params, int $size): array
    {
        // TODO: Implement pageableListByParams() method.
    }

    /**
     * @param int $size
     * @param bool $deleted
     * @return array
     */
    public function pageableListAll(int $size, bool $deleted): array
    {
        // TODO: Implement pageableListAll() method.
    }

    /**
     * @param array $data
     * @return array
     */
    public function update(array $data): array
    {
        // TODO: Implement update() method.
    }
}
