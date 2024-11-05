<?php

use PHPUnit\Framework\TestCase;
use GeoIp2\Database\Reader;
use GeoIp2\Exception\AddressNotFoundException;
use MaxMind\Db\Reader\InvalidDatabaseException;

class GeoIpTest extends TestCase
{
    private $reader;
    private $config;

    protected function setUp(): void
    {
        $this->config = require __DIR__ . '/../config.php';

        try {
            $this->reader = new Reader($this->config['geoip']['database_file']);
            error_log("Initialized GeoIP2 Reader with database file: {$this->config['geoip']['database_file']}");
        } catch (InvalidDatabaseException $e) {
            error_log("Invalid GeoIP2 database: " . $e->getMessage());
            $this->markTestSkipped('GeoIP2 Reader initialization failed: ' . $e->getMessage());
        } catch (Exception $e) {
            error_log("Failed to initialize GeoIP2 Reader: " . $e->getMessage());
            $this->markTestSkipped('GeoIP2 Reader initialization failed: ' . $e->getMessage());
        }
    }

    public function testInvalidIp()
    {
        $this->expectException(InvalidArgumentException::class);
        $ip = '999.999.999.999';
        $this->reader->country($ip);
    }

    public function testValidButUnknownIp()
    {
        $this->expectException(AddressNotFoundException::class);
        $ip = '203.0.113.1';
        $this->reader->country($ip);
    }

    public function testValidIpv6()
    {
        $ip = '2001:4860:4860::8888';

        try {
            $record = $this->reader->country($ip);
            $country = [
                'iso_code' => $record->country->isoCode,
                'name' => $record->country->name,
            ];

            $this->assertIsArray($country);
            $this->assertNotEmpty($country['iso_code'], "ISO код не должен быть пустым");
            $this->assertNotEmpty($country['name'], "Имя страны не должно быть пустым");
        } catch (AddressNotFoundException $e) {
            $this->fail("Адрес $ip не найден в базе данных GeoIP2.");
        } catch (Exception $e) {
            $this->fail("Произошла неожиданная ошибка: " . $e->getMessage());
        }
    }

    public function testValidIpv4()
    {
        $ip = '8.8.8.8'; 

        try {
            $record = $this->reader->country($ip);
            $country = [
                'iso_code' => $record->country->isoCode,
                'name' => $record->country->name,
            ];

            $this->assertIsArray($country);
            $this->assertNotEmpty($country['iso_code'], "ISO код не должен быть пустым");
            $this->assertNotEmpty($country['name'], "Имя страны не должно быть пустым");
        } catch (AddressNotFoundException $e) {
            $this->fail("Адрес $ip не найден в базе данных GeoIP2.");
        } catch (Exception $e) {
            $this->fail("Произошла неожиданная ошибка: " . $e->getMessage());
        }
    }

    public function testInvalidIpv6()
    {
        $this->expectException(InvalidArgumentException::class);
        $ip = 'gggg::gggg';
        $this->reader->country($ip);
    }

    public function testValidButUnknownIpv6()
    {
        $this->expectException(AddressNotFoundException::class);
        $ip = '2001:db8::1';
        $this->reader->country($ip);
    }

    protected function tearDown(): void
    {
        if ($this->reader) {
            $this->reader->close();
            error_log("Closed GeoIP2 Reader.");
        }
    }
}
