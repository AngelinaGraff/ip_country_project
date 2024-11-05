<?php

return [
    'cache' => [
        'enabled' => true,
        'driver' => 'redis',
        'redis' => [
            'host' => getenv('REDIS_HOST') ?: 'redis',
            'port' => getenv('REDIS_PORT') ?: 6379,
        ],
        'ttl' => 3600,
    ],
    'geoip' => [
        'database_file' => '/var/www/html/GeoLite2-Country.mmdb',
    ],
];
