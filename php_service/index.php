<?php

require_once 'vendor/autoload.php';
$config = require 'config.php';

use GeoIp2\Database\Reader;

// Create a function for logging
function log_message($message) {
    error_log($message);
}

// Get the IP address from GET parameters
$ip = isset($_GET['ip']) ? $_GET['ip'] : null;

// Log the received IP address
log_message("Received request for IP: $ip");

// Validate the IP address
if (!$ip || !filter_var($ip, FILTER_VALIDATE_IP)) {
    http_response_code(400);
    header('Content-Type: text/plain');
    echo 'Invalid IP address';
    log_message("Invalid IP address provided: $ip");
    exit;
}

// Initialize cache based on configuration
$cache = null;
if ($config['cache']['enabled']) {
    if ($config['cache']['driver'] === 'redis') {
        $cache = new Redis();
        try {
            $cache->connect($config['cache']['redis']['host'], $config['cache']['redis']['port']);
            log_message("Connected to Redis at {$config['cache']['redis']['host']}:{$config['cache']['redis']['port']}");
        } catch (Exception $e) {
            log_message("Failed to connect to Redis: " . $e->getMessage());
        }
    }
}

// Generate cache key for the IP address
$cacheKey = 'geoip:' . $ip;

// Try to retrieve data from cache
if ($cache && ($cachedData = $cache->get($cacheKey))) {
    $country = unserialize($cachedData);
    log_message("Cache hit for IP: $ip");
} else {
    log_message("Cache miss for IP: $ip. Performing GeoIP lookup.");
    try {
        // Create a new instance of GeoIP2 Reader
        $reader = new Reader($config['geoip']['database_file']);
        // Perform country lookup
        $record = $reader->country($ip);
        $country = [
            'iso_code' => $record->country->isoCode,
            'name' => $record->country->name,
        ];
        // Save the result in cache
        if ($cache) {
            $cache->set($cacheKey, serialize($country), $config['cache']['ttl']);
            log_message("Cached result for IP: $ip with TTL {$config['cache']['ttl']} seconds");
        }
    } catch (Exception $e) {
        // Handle exceptions (e.g., address not found in the database)
        http_response_code(400);
        header('Content-Type: text/plain');
        echo 'Unable to determine country for IP';
        log_message("Error determining country for IP $ip: " . $e->getMessage());
        exit;
    }
}

// Log successful data retrieval
log_message("Successfully retrieved country data for IP: $ip");

// Return the result in JSON format
header('Content-Type: application/json');
echo json_encode(['ip' => $ip, 'country' => $country]);
