<?php

use PHPUnit\Framework\TestCase;

class IpValidationTest extends TestCase
{
    /**
     * @dataProvider ipProvider
     */
    public function testValidIpAddresses($ip, $expected)
    {
        $result = filter_var($ip, FILTER_VALIDATE_IP) !== false;
        $this->assertSame($expected, $result);
    }

    public function ipProvider()
    {
        return [
            ['8.8.8.8', true],
            ['2001:4860:4860::8888', true],
            ['999.999.999.999', false],
            ['invalid-ip', false],
            ['', false],
            [null, false],
        ];
    }
}
