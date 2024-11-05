package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"github.com/oschwald/geoip2-golang"
	"gopkg.in/yaml.v2"
)

// Config holds the application's configuration settings
type Config struct {
	Server struct {
		Port string `yaml:"port"`
	} `yaml:"server"`
	GeoIP struct {
		DBPath string `yaml:"db_path"`
	} `yaml:"geoip"`
	Cache struct {
		Address  string `yaml:"address"`
		Password string `yaml:"password"`
		DB       int    `yaml:"db"`
	} `yaml:"cache"`
}

var (
	cfg Config
	db  *geoip2.Reader
	rdb *redis.Client
	ctx = context.Background()
)

// loadConfig reads and parses the configuration file
func loadConfig() {
	data, err := os.ReadFile("config.yaml")
	if err != nil {
		log.Fatalf("Error reading configuration file: %v", err)
	}
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		log.Fatalf("Error parsing configuration file: %v", err)
	}

	// Override Redis address if provided via environment variable
	if redisAddr := os.Getenv("REDIS_ADDRESS"); redisAddr != "" {
		cfg.Cache.Address = redisAddr
	}
	log.Println("Configuration loaded successfully")
}

// initGeoIP initializes the GeoIP database reader
func initGeoIP() {
	var err error
	db, err = geoip2.Open(cfg.GeoIP.DBPath)
	if err != nil {
		log.Fatalf("Failed to open GeoIP database: %v", err)
	}
	log.Println("GeoIP database initialized successfully")
}

// initRedis initializes the Redis client
func initRedis() {
	rdb = redis.NewClient(&redis.Options{
		Addr:     cfg.Cache.Address,
		Password: cfg.Cache.Password,
		DB:       cfg.Cache.DB,
	})

	// Ping Redis to ensure the connection is established
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Println("Connected to Redis successfully")
}

// getCountryByIP handles HTTP requests to retrieve the country code by IP address
func getCountryByIP(w http.ResponseWriter, r *http.Request) {
	ip := r.URL.Query().Get("ip")
	log.Printf("Received request for IP: %s", ip)

	// Check if the 'ip' parameter is provided
	if ip == "" {
		log.Println("Parameter 'ip' is missing")
		http.Error(w, "Parameter 'ip' is missing", http.StatusBadRequest)
		return
	}

	// Parse and validate the IP address
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		log.Printf("Invalid IP address: %s", ip)
		http.Error(w, "Invalid IP address", http.StatusBadRequest)
		return
	}

	// Check if the country code is cached in Redis
	cachedCountry, err := rdb.Get(ctx, ip).Result()
	if err == redis.Nil {
		// IP not found in cache, query the GeoIP database
		log.Printf("IP %s not found in cache, querying GeoIP database", ip)
		record, err := db.Country(parsedIP)
		if err != nil {
			log.Printf("Error processing IP address %s: %v", ip, err)
			http.Error(w, fmt.Sprintf("Error processing IP address: %v", err), http.StatusInternalServerError)
			return
		}

		country := record.Country.IsoCode
		log.Printf("Retrieved country code '%s' for IP %s", country, ip)

		// Save the country code to cache with a TTL of 24 hours
		err = rdb.Set(ctx, ip, country, 24*time.Hour).Err()
		if err != nil {
			log.Printf("Error saving to cache: %v", err)
		} else {
			log.Printf("Saved country code '%s' for IP %s to cache", country, ip)
		}

		// Send the response with the country code
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"country": country})
	} else if err != nil {
		// Handle errors when retrieving from cache
		log.Printf("Error retrieving from cache: %v", err)
		http.Error(w, fmt.Sprintf("Cache error: %v", err), http.StatusInternalServerError)
		return
	} else {
		// IP found in cache, return the cached country code
		log.Printf("IP %s found in cache, returning cached country code '%s'", ip, cachedCountry)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"country": cachedCountry})
	}
}

func main() {
	// Load application configuration
	loadConfig()

	// Initialize the GeoIP database
	initGeoIP()
	defer db.Close()

	// Initialize the Redis client
	initRedis()
	defer rdb.Close()

	// Create a new router
	r := mux.NewRouter()
	// Define the route for getting country by IP
	r.HandleFunc("/getcountry", getCountryByIP).Methods("GET")

	// Start the HTTP server
	log.Printf("Server started on port %s", cfg.Server.Port)
	if err := http.ListenAndServe(cfg.Server.Port, r); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
