<?php

namespace Tests;

use Illuminate\Foundation\Testing\TestCase as BaseTestCase;
use Illuminate\Testing\TestResponse;
use Symfony\Component\Console\Output\ConsoleOutput;

abstract class TestCase extends BaseTestCase
{
    use CreatesApplication;

    public const TEST_SA_NICKNAME = "test_feature_sa";
    public const TEST_SA_EMAIL = "test_feature_sa@xmail.xx";
    public const TEST_SA_PASSWORD = "cm1208qmp3k484hrykvsc4";

    public const TEST_ADMIN_NICKNAME = "test_feature_admin";
    public const TEST_ADMIN_EMAIL = "test_feature_admin@xmail.xx";
    public const TEST_ADMIN_PASSWORD = "cm1208dhl2k484mlrgs7p8";

    public const TEST_MODERATOR_NICKNAME = "test_feature_moderator";
    public const TEST_MODERATOR_EMAIL = "test_feature_moderator@xmail.xx";
    public const TEST_MODERATOR_PASSWORD = "cm1205d2q1k484ppdvx0na";

    public const TEST_CLIENT_NICKNAME = "test_feature_client";
    public const TEST_CLIENT_EMAIL = "test_feature_client@xmail.xx";
    public const TEST_CLIENT_PASSWORD = "cm1203hz2k4847d2t5uak";

    private const CHECK_PREFIX = '+ OK';

    /**
     * @var ConsoleOutput
     */
    private $output;

    /**
     * TestCase constructor.
     * @param string|null $name
     * @param array $data
     * @param string $dataName
     */
    public function __construct(?string $name = null, array $data = [], $dataName = '')
    {
        parent::__construct($name, $data, $dataName);

        $this->output = new ConsoleOutput();
    }

    /**
     * @param string $reason
     */
    final public function okMsg(string $reason = ''): void
    {
        if (!empty($reason) === '') {
            $this->console("> " . $reason);
        }
        $this->console(self::CHECK_PREFIX);
    }

    /**
     * @param string $msg
     * @return void
     */
    final public function console(string $msg): void
    {
        $this->output->writeln($msg);
    }

    /**
     * @return ConsoleOutput
     */
    public function getOutput(): ConsoleOutput
    {
        return $this->output;
    }

    /**
     * @param string $url
     * @param array $data
     * @param array $headers
     * @return TestResponse
     */
    public function sendPost(string $url, array $data = [], array $headers = []): TestResponse
    {
        if ($headers)
            $response = $this->withHeaders($headers)->postJson($url, $data);
        else
            $response = $this->postJson($url, $data);

        if ($data)
            $this->console('Data: ' . json_encode($data));
        $this->console('Status: ' . $response->status());
        $this->console('Response: ' . $response->content());
        return $response;
    }

    /**
     * @param string $url
     * @param array $headers
     * @return TestResponse
     */
    public function sendGet(string $url, array $headers = []): TestResponse
    {
        if ($headers)
            $response = $this->withHeaders($headers)->get($url);
        else
            $response = $this->get($url);

        $this->console('Status: ' . $response->status());
        $this->console('Response: ' . $response->content());
        return $response;
    }
}
