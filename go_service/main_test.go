package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"github.com/oschwald/geoip2-golang"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

type TestConfig struct {
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
	testCfg TestConfig
	testDb  *geoip2.Reader
	testRdb *redis.Client
	testCtx = context.Background()
)

func loadConfigTest() {
	data, err := os.ReadFile("config.yaml")
	if err != nil {
		log.Fatalf("Error reading configuration file: %v", err)
	}
	err = yaml.Unmarshal(data, &testCfg)
	if err != nil {
		log.Fatalf("Error parsing configuration file: %v", err)
	}

	if redisAddr := os.Getenv("REDIS_ADDRESS"); redisAddr != "" {
		testCfg.Cache.Address = redisAddr
	}
	log.Println("Configuration loaded successfully")
}

func init() {
	loadConfigTest()
	var err error
	testDb, err = geoip2.Open(testCfg.GeoIP.DBPath)
	log.Printf("Opening GeoIP database at %s", testCfg.GeoIP.DBPath)
	if err != nil {
		panic(err)
	}

	testRdb = redis.NewClient(&redis.Options{
		Addr:     testCfg.Cache.Address,
		Password: testCfg.Cache.Password,
		DB:       testCfg.Cache.DB,
	})

	_, err = testRdb.Ping(testCtx).Result()
	if err != nil {
		panic(err)
	}
}

func TestGetCountryByIP(t *testing.T) {
	testRdb.FlushDB(testCtx)

	req, err := http.NewRequest("GET", "/getcountry?ip=8.8.8.8", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	r := mux.NewRouter()
	r.HandleFunc("/getcountry", func(w http.ResponseWriter, r *http.Request) {
		getCountryByIPTest(w, r, testDb, testRdb)
	}).Methods("GET")

	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, "Ожидается статус 200 OK")

	var resp map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &resp)
	if err != nil {
		t.Fatalf("Ошибка при разборе ответа: %v", err)
	}

	assert.Equal(t, "8.8.8.8", resp["ip"], "IP должен совпадать")
	country, ok := resp["country"].(map[string]interface{})
	if !ok {
		t.Fatalf("Неверный формат поля 'country'")
	}
	assert.Equal(t, "US", country["iso_code"], "Код страны должен быть 'US'")

	cachedCountry, err := testRdb.Get(testCtx, "8.8.8.8").Result()
	assert.NoError(t, err, "Данные должны быть сохранены в кэше")
	assert.Equal(t, "US", cachedCountry, "Кэш должен содержать код страны 'US'")
}

func TestGetCountryByIP_InvalidIP(t *testing.T) {
	req, err := http.NewRequest("GET", "/getcountry?ip=invalid_ip", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	r := mux.NewRouter()
	r.HandleFunc("/getcountry", func(w http.ResponseWriter, r *http.Request) {
		getCountryByIPTest(w, r, testDb, testRdb)
	}).Methods("GET")

	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code, "Ожидается статус 400 Bad Request")
	assert.Equal(t, "Invalid IP address\n", rr.Body.String(), "Сообщение об ошибке должно быть 'Invalid IP address'")
}

func getCountryByIPTest(w http.ResponseWriter, r *http.Request, db *geoip2.Reader, rdb *redis.Client) {
	ip := r.URL.Query().Get("ip")
	log.Printf("Received request for IP: %s", ip)

	if ip == "" {
		log.Println("Parameter 'ip' is missing")
		http.Error(w, "Parameter 'ip' is missing", http.StatusBadRequest)
		return
	}

	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		log.Printf("Invalid IP address: %s", ip)
		http.Error(w, "Invalid IP address", http.StatusBadRequest)
		return
	}

	testCtx := context.Background()

	cachedCountry, err := rdb.Get(testCtx, ip).Result()
	if err == redis.Nil {
		log.Printf("IP %s not found in cache, querying GeoIP database", ip)
		record, err := db.Country(parsedIP)
		if err != nil {
			log.Printf("Error processing IP address %s: %v", ip, err)
			http.Error(w, fmt.Sprintf("Error processing IP address: %v", err), http.StatusInternalServerError)
			return
		}

		country := record.Country.IsoCode
		log.Printf("Retrieved country code '%s' for IP %s", country, ip)

		err = rdb.Set(testCtx, ip, country, 24*time.Hour).Err()
		if err != nil {
			log.Printf("Error saving to cache: %v", err)
		} else {
			log.Printf("Saved country code '%s' for IP %s to cache", country, ip)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"ip": ip, "country": map[string]string{"iso_code": country}})
	} else if err != nil {
		log.Printf("Error retrieving from cache: %v", err)
		http.Error(w, fmt.Sprintf("Cache error: %v", err), http.StatusInternalServerError)
		return
	} else {
		log.Printf("IP %s found in cache, returning cached country code '%s'", ip, cachedCountry)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"ip": ip, "country": map[string]string{"iso_code": cachedCountry}})
	}
}
