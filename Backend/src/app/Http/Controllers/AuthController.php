<?php

namespace App\Http\Controllers;

use App\Constants\Common;
use App\Models\User;
use App\Models\UserType;
use App\Repositories\UserTypeRepository;
use App\Services\Contracts\UserServiceInterface;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\Auth;
use Symfony\Component\HttpFoundation\Response as ResponseAlias;

/**
 *
 */
class AuthController extends Controller
{
    /**
     * @var UserServiceInterface
     */
    private UserServiceInterface $userService;

    /**
     * @var UserTypeRepository
     */
    private UserTypeRepository $userTypeRepository;

    /**
     * @param UserServiceInterface $userService
     * @param UserTypeRepository $userTypeRepository
     */
    public function __construct(
        UserServiceInterface $userService,
        UserTypeRepository $userTypeRepository
    ) {
        $this->userService = $userService;
        $this->userTypeRepository = $userTypeRepository;
    }

    /**
     * @param Request $request
     * @return JsonResponse
     */
    public function register(Request $request): JsonResponse
    {
        $data = $request->json()?->all();
        $clientUserType = $this->userTypeRepository->getByTypeAlias(UserType::ALIAS_CLIENT);
        $data[User::USER_TYPE_ID] = $clientUserType[UserType::ID];
        $result = $this->userService->store($data);

        if (!empty($result[Common::ERROR])) {
            return response()->json($result, ResponseAlias::HTTP_BAD_REQUEST);
        }

        $user = User::find($result[User::ID]);
        $token = Auth::login($user);

        return response()->json(compact('user', 'token'), 201);
    }

    /**
     * @param Request $request
     * @return JsonResponse
     */
    public function login(Request $request): JsonResponse
    {
        $credentials = $request->only('email', 'password');

        if (!$token = Auth::attempt($credentials)) {
            return response()->json(['error' => 'Unauthorized'], 401);
        }

        return $this->respondWithToken($token);
    }

    /**
     * @return JsonResponse
     */
    public function logout(): JsonResponse
    {
        Auth::logout();

        return response()->json([
            'status' => Common::STATUS_SUCCESS,
            'message' => 'Successfully logged out'
        ]);
    }

    /**
     * @return JsonResponse
     */
    public function me(): JsonResponse
    {
        return response()->json(Auth::user());
    }

    /**
     * @param $token
     * @return JsonResponse
     */
    protected function respondWithToken($token): JsonResponse
    {
        return response()->json([
            'access_token' => $token,
            'token_type' => 'bearer',
            'expires_in' => Auth::factory()->getTTL() * 60
        ]);
    }
}
