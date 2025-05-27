<?php

namespace App\Models;

// use Illuminate\Contracts\Auth\MustVerifyEmail;
use Illuminate\Database\Eloquent\Factories\HasFactory;
use Illuminate\Database\Eloquent\SoftDeletes;
use Illuminate\Foundation\Auth\User as Authenticatable;
use Illuminate\Notifications\Notifiable;
use Laravel\Sanctum\HasApiTokens;
use Tymon\JWTAuth\Contracts\JWTSubject;

class User extends Authenticatable implements JWTSubject
{
    use HasApiTokens;
    use HasFactory;
    use Notifiable;
    use SoftDeletes;

    public const TABLE_NAME = 'users';

    public const ID = 'id';
    public const USER_TYPE_ID = 'user_type_id';
    public const NICK_NAME = 'nickname';
    public const EMAIL = 'email';

    public const PASSWORD = 'password';
    public const EMAIL_VERIFIED_AT = 'email_verified_at';
    public const REMEMBER_TOKEN = 'remember_token';
    public const CREATED_AT = 'created_at';
    public const UPDATED_AT = 'updated_at';
    public const DELETED_AT = 'deleted_at';

    public const PASSWORD_CONFIRMATION = 'password_confirmation';

    protected $keyType = 'string';

    /**
     * The attributes that are mass assignable.
     *
     * @var array<int, string>
     */
    protected $fillable = [
        self::USER_TYPE_ID,
        self::NICK_NAME,
        self::EMAIL,
        self::PASSWORD,
        self::CREATED_AT,
        self::UPDATED_AT,
        self::DELETED_AT,
    ];

    /**
     * The attributes that should be hidden for serialization.
     *
     * @var array<int, string>
     */
    protected $hidden = [
        self::PASSWORD,
        self::REMEMBER_TOKEN,
        self::CREATED_AT,
        self::UPDATED_AT,
    ];

    /**
     * The attributes that should be cast.
     *
     * @var array<string, string>
     */
    protected $casts = [
        self::ID => 'string',
        self::EMAIL_VERIFIED_AT => 'datetime',
        self::DELETED_AT => 'datetime',
        self::PASSWORD => 'hashed',
    ];

    /**
     * Get the identifier that will be stored in the subject claim of the JWT.
     *
     * @return mixed
     */
    public function getJWTIdentifier()
    {
        return $this->getKey();
    }

    /**
     * Return a key value array, containing any custom claims to be added to the JWT.
     *
     * @return array
     */
    public function getJWTCustomClaims()
    {
        return [];
    }

    /**
     * Получить аккаунты Bybit пользователя
     */
    public function bybitAccounts()
    {
        return $this->hasMany(BybitAccount::class);
    }

    /**
     * Получить активный аккаунт Bybit пользователя
     */
    public function activeBybitAccount()
    {
        return $this->bybitAccounts()->where('is_active', true)->first();
    }

    /**
     * Получить аккаунт Bybit пользователя
     */
    public function bybitAccount()
    {
        return $this->hasOne(BybitAccount::class);
    }
}
