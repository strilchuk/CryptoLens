<?php

namespace App\Exceptions;

use App\Constants\Common;
use Illuminate\Database\QueryException;
use Illuminate\Foundation\Exceptions\Handler as ExceptionHandler;
use Illuminate\Support\Facades\Lang;
use Illuminate\Support\Facades\Log;
use Symfony\Component\HttpFoundation\Response as ResponseAlias;
use Throwable;

class Handler extends ExceptionHandler
{
    /**
     * The list of the inputs that are never flashed to the session on validation exceptions.
     *
     * @var array<int, string>
     */
    protected $dontFlash = [
        'current_password',
        'password',
        'password_confirmation',
    ];

    /**
     * Register the exception handling callbacks for the application.
     */
    public function register(): void
    {
        $this->reportable(function (Throwable $e) {
            //
        });
    }

    public function render($request, \Throwable $e)
    {
//        $result = parent::render($request, $e);
        $errorData = [];
        Log::error($e->getMessage());
        if ($e instanceof QueryException) {
            $errorData[Common::ERROR] = Lang::get('errors.database_error');
        } elseif ($e instanceof \Exception) {
            $errorData[Common::ERROR] = $e->getMessage();
        } else {
            $errorData[Common::ERROR] = Lang::get('errors.backend_error');
        }

        return response()->json(
            $errorData,
            ResponseAlias::HTTP_INTERNAL_SERVER_ERROR,
            ['Content-Type' => 'application/json;charset=UTF-8', 'Charset' => 'utf-8'],
            JSON_UNESCAPED_UNICODE
        );
    }

    /**
     * @param Throwable $e
     * @throws Throwable
     */
    public function report(Throwable $e): void
    {
        $ex_message = (!empty($e->getMessage()) ? trim($e->getMessage()) : 'App Error Exception');

        $log_message = "\"" . $ex_message . " in file '" . $e->getFile() . "' on line '" . $e->getLine() . "'" . "\"";

        if (!config('app.debug')) {
            Log::error($log_message);
        } else {
            parent::report($e);
        }
    }
}
