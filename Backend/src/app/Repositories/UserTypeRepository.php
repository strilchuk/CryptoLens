<?php

namespace App\Repositories;

use App\Models\UserType;

/**
 *
 */
class UserTypeRepository
{
    /**
     * @var UserType
     */
    private UserType $userType;

    /**
     * @param UserType $userType
     */
    public function __construct(UserType $userType)
    {
        $this->userType = $userType;
    }

    /**
     * @param string $alias
     * @return array
     */
    final public function getByTypeAlias(string $alias): array
    {
        return $this->userType->newQuery()->where([UserType::TYPE_ALIAS => $alias])->first()->toArray() ?? [];
    }
}
