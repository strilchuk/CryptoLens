<?php

namespace App\Repositories;

use App\Models\User;
use Faker\Provider\Uuid;

/**
 *
 */
class UserRepository
{
    /**
     * @var User
     */
    private User $user;

    /**
     * @param User $user
     */
    public function __construct(User $user)
    {
        $this->user = $user;
    }

    /**
     * @param array $data
     * @return array
     */
    final public function create(array $data): array
    {
        $id = Uuid::uuid();
        $data[User::ID] = $id;
        $this->user->newQuery()->insert($data);

        return $this->user->newQuery()
            ->where([User::ID => $id])->first()->toArray() ?? [];
    }
}
