<?php

namespace App\Services\Contracts;

interface UserServiceInterface
{
    public function store(array $data): array;

    public function delete(?string $id): bool;

    public function restore(?string $id): bool;

    public function pageableListByParams(array $params, int $size): array;

    public function pageableListAll(int $size, bool $deleted): array;

    public function update(array $data): array;
}
