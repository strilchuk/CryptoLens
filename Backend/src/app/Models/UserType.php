<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Factories\HasFactory;
use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\SoftDeletes;

class UserType extends Model
{
    use HasFactory;
    use SoftDeletes;

    public const TABLE_NAME = 'user_types';

    public const ID = 'id';
    public const NAME = 'name';
    public const TYPE_ALIAS = 'type_alias';
    public const IS_SA = 'is_sa';
    public const IS_ADMIN = 'is_ADMIN';
    public const IS_MODERATOR = 'is_MODERATOR';
    public const IS_CLIENT = 'is_client';
    public const DELETED_AT = 'deleted_at';

    public const ALIAS_SA = 'sa';
    public const ALIAS_ADMIN = 'admin';
    public const ALIAS_MODERATOR = 'moderator';
    public const ALIAS_CLIENT = 'client';
    /**
     * Indicates if the model should be timestamped.
     *
     * @var bool
     */
    public $timestamps = false;

    protected $keyType = 'string';

    /**
     * The attributes that are mass assignable.
     *
     * @var array<int, string>
     */
    protected $fillable = [
        self::NAME,
        self::TYPE_ALIAS,
        self::IS_SA,
        self::IS_ADMIN,
        self::IS_MODERATOR,
        self::IS_CLIENT,
    ];
    /**
     * The attributes that should be cast.
     *
     * @var array<string, string>
     */
    protected $casts = [
        self::ID => 'string',
        self::DELETED_AT => 'datetime',
    ];
}
