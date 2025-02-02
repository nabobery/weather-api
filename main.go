package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/didip/tollbooth/v7"
	tollbooth_gin "github.com/didip/tollbooth_gin"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

// Global variables for the Redis client and configuration
var (
	ctx             = context.Background()
	redisClient     *redis.Client
	apiKey          string
	apiUrl          string
	cacheExpiration time.Duration
)

// init loads configuration settings and initializes the Redis client.
func init() {
	// Load environment variables from .env if available
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, proceeding with system environment variables")
	}

	// Retrieve environment configurations
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "localhost:6379"
	}
	apiKey = os.Getenv("VISUAL_CROSSING_API_KEY")
	if apiKey == "" {
		log.Fatal("VISUAL_CROSSING_API_KEY must be set")
	}
	apiUrl = os.Getenv("VISUAL_CROSSING_API_URL")
	if apiUrl == "" {
		log.Fatal("VISUAL_CROSSING_API_URL must be set")
	}
	expirationStr := os.Getenv("CACHE_EXPIRATION")
	if expirationStr == "" {
		expirationStr = "43200" // Default: 12 hours in seconds
	}
	expirationSec, err := strconv.Atoi(expirationStr)
	if err != nil {
		log.Printf("Invalid CACHE_EXPIRATION, defaulting to 43200 seconds")
		expirationSec = 43200
	}
	cacheExpiration = time.Duration(expirationSec) * time.Second

	// Initialize the Redis client
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		// Changed from log.Fatal to log.Fatalf to properly apply the format directive.
		log.Fatalf("failed to fetch weather data: %v", err)
	}
	redisClient = redis.NewClient(opt)
	// Test Redis connection
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis at %s: %v", redisURL, err)
	}
	log.Println("Connected to Redis at", redisURL)
}

// fetchWeatherData constructs the API URL using the provided location and fetches data
// from the third-party weather API (Visual Crossing).
func fetchWeatherData(location string) (map[string]interface{}, error) {
	// Build the Visual Crossing API URL.
	// Example: {API_URL}/{location}?key={API_KEY}&unitGroup=metric&include=days
	url := fmt.Sprintf("%s/%s?key=%s&unitGroup=metric&include=days", apiUrl, location, apiKey)
	log.Println("Fetching weather data from:", url)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch weather data: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to fetch weather data: status %d, response: %s", resp.StatusCode, string(bodyBytes))
	}

	var data map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode weather data: %v", err)
	}

	return data, nil
}

// getWeatherHandler handles GET /weather requests.
// It determines whether cached data exists for the requested location, and if not,
// it fetches the data from the weather API, caches it in Redis, and returns the result.
func getWeatherHandler(c *gin.Context) {
	location := c.Query("location")
	if location == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "location query parameter is required"})
		return
	}

	cacheKey := "weather:" + location

	// Attempt to retrieve cached weather data from Redis.
	cachedData, err := redisClient.Get(ctx, cacheKey).Result()
	var weatherData map[string]interface{}

	if err == redis.Nil {
		// Cache miss: fetch the weather data from the API.
		weatherData, err = fetchWeatherData(location)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Marshal the retrieved data into JSON and store it in Redis.
		jsonData, err := json.Marshal(weatherData)
		if err != nil {
			log.Printf("Error marshaling weather data: %v", err)
		} else {
			if err := redisClient.Set(ctx, cacheKey, jsonData, cacheExpiration).Err(); err != nil {
				log.Printf("Error caching weather data: %v", err)
			}
		}
		log.Printf("Fetched fresh weather data for location: %s", location)
	} else if err != nil {
		log.Printf("Error retrieving data from Redis: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	} else {
		// Cache hit: unmarshal the JSON data from the cache.
		if err := json.Unmarshal([]byte(cachedData), &weatherData); err != nil {
			log.Printf("Error unmarshaling cached data: %v", err)
			// Optionally, fetch fresh data if unmarshaling fails.
			weatherData, err = fetchWeatherData(location)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		} else {
			log.Printf("Serving cached weather data for location: %s", location)
		}
	}

	c.JSON(http.StatusOK, weatherData)
}

func main() {
	router := gin.Default()

	// --------------------------------------------------------------
	// RATE LIMITING SETUP:
	// Create a new limiter that allows, for example, 1 request per second.
	// Adjust the parameter to suit your needs.
	limiter := tollbooth.NewLimiter(1, nil)

	// Attach Tollbooth's Gin middleware to limit all incoming requests.
	router.Use(tollbooth_gin.LimitHandler(limiter))
	// --------------------------------------------------------------

	// Define the /weather endpoint.
	router.GET("/weather", getWeatherHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Server listening on port %s", port)
	// Check the error returned from router.Run.
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("failed to start the server: %v", err)
	}
}
