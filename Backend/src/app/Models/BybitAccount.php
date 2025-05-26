<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\Relations\BelongsTo;
use Illuminate\Database\Eloquent\SoftDeletes;

class BybitAccount extends Model
{
    use SoftDeletes;

    protected $keyType = 'string';

    protected $fillable = [
        'user_id',
        'api_key',
        'api_secret',
        'account_type',
        'is_active'
    ];

    protected $casts = [
        'is_active' => 'boolean',
        'user_id' => 'string'
    ];

    /**
     * Получить пользователя, которому принадлежит аккаунт
     */
    public function user(): BelongsTo
    {
        return $this->belongsTo(User::class);
    }
} 