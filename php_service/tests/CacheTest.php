<?php

use PHPUnit\Framework\TestCase;

class CacheTest extends TestCase
{
    private $cache;
    private $config;

    protected function setUp(): void
    {
        $this->config = require __DIR__ . '/../config.php';

        $this->cache = null;
        if ($this->config['cache']['enabled']) {
            if ($this->config['cache']['driver'] === 'redis') {
                $this->cache = new Redis();
                try {
                    $this->cache->connect($this->config['cache']['redis']['host'], $this->config['cache']['redis']['port']);
                    error_log("Connected to Redis at {$this->config['cache']['redis']['host']}:{$this->config['cache']['redis']['port']}");
                } catch (Exception $e) {
                    error_log("Failed to connect to Redis: " . $e->getMessage());
                    $this->markTestSkipped('Redis connection failed: ' . $e->getMessage());
                }
            }
        }
    }

    public function testSetAndGet()
    {
        if (!$this->cache) {
            $this->markTestSkipped('Redis is not enabled or not available.');
        }

        $key = 'test:key';
        $value = ['foo' => 'bar'];

        $setResult = $this->cache->set($key, serialize($value), $this->config['cache']['ttl']);
        $this->assertTrue($setResult, "Failed to set value in cache");

        $cachedData = $this->cache->get($key);
        $this->assertNotFalse($cachedData, "Failed to get value from cache");

        $result = unserialize($cachedData);

        $this->assertSame($value, $result);
    }

    protected function tearDown(): void
    {
        if ($this->cache) {
            $this->cache->flushAll();
            $this->cache->close();
        }
    }
}