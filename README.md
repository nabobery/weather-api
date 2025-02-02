# Weather API

https://roadmap.sh/projects/weather-api-wrapper-service

In this project, instead of relying on our own weather data, we will build a weather API that fetches and returns weather data from a third-party API. This project will help you understand how to work with 3rd party APIs, caching, and environment variables. As a suggestion, we are using Visual Crossing’s API—which is completely free and easy to use—for fetching weather data.

The API is implemented in Go using the [Gin](https://github.com/gin-gonic/gin) web framework. Data is cached in Redis for 12 hours to improve performance and reduce unnecessary API calls. We also include sample code for rate limiting using [Tollbooth](https://github.com/didip/tollbooth) should you wish to prevent abuse of your API.

## Features

- Fetch weather data from [Visual Crossing](https://weather.visualcrossing.com/VisualCrossingWebServices/rest/services/timeline).
- Redis caching with a 12-hour expiration period.
- Environment variable configuration to easily change settings.
- (Optional) Rate limiting to prevent abuse of the API.

## Prerequisites

- [Go 1.21+](https://golang.org/dl/)
- A Redis server (local or cloud, for example provided by Onrender)
- An API key from Visual Crossing

## Environment Configuration

Create a `.env` file in the project root with the following variables:

```env
# Visual Crossing API settings
VISUAL_CROSSING_API_KEY="your_visual_crossing_api_key"
VISUAL_CROSSING_API_URL="https://weather.visualcrossing.com/VisualCrossingWebServices/rest/services/timeline"

# Redis connection string
# For Onrender or other cloud providers, use the full URL (e.g., redis://default:yourpassword@your-redis-host:6379)
REDIS_URL="redis://default:yourpassword@your-redis-host:6379"

# Cache expiration time in seconds (43200 seconds = 12 hours)
CACHE_EXPIRATION="43200"

# (Optional) The port the API server will listen on
PORT="8080"
```

**Note:** Update each variable accordingly based on your configuration.

## Installation

1. **Clone the Repository**

   ```bash
   git clone https://github.com/nabobery/weather-api.git
   cd weather-api
   ```

2. **Install Dependencies**

   Ensure you have the necessary modules (Gin, go-redis, godotenv) installed. You can download them easily with:

   ```bash
   go mod tidy
   ```

3. **Run the Application**

   Start the server with:

   ```bash
   go run main.go
   ```

   If your environment variables are correctly configured, you should see a log message similar to:

   ```
   Connected to Redis at redis://default:yourpassword@your-redis-host:6379
   Server listening on port 8080
   ```

## Usage

The API exposes a single endpoint: `/weather`. You need to pass a location via the query parameter.

### Flow

[![](https://mermaid.ink/img/pako:eNp9ksFqwzAMhl9F-NwGtm6XHAKjHayHQWkphZGLsLXGNHE82-lWSt99SpM0KYHmZEuffv1SfBayVCRi4emnIiNpoXHvsEgN8Bd0yAl2hCEjB2-rJewcWsvnDbmjlgS_OmSwJqU9zFFm1NRZdEFLbdEE2HrG0fcq1sJ2OeZq9SHWNxuz11Y1_bDz-18gZzBvlWdOwYrTp1uPdojUNLW102mSMB6zMK_Dhxu6wIANxWmGri1jeIpgnpE8gOwtXFPTTug5at068rY0voUw76b40KEJdeJcWTupPYTKmQZTAweUe2qLP7X399VJMhg7hlk0GoXjfckAvll-ie7-wvrO-GgFrxFs8EidTaarPPjHI423SkaJiSjIFagVP8dzHU4FUwWlIuajQndIRWouzGEVys3JSBEHV9FEuLLaZ92lsgpD945F_I28rongF_FVlt398g-6A_ME?type=png)](https://mermaid.live/edit#pako:eNp9ksFqwzAMhl9F-NwGtm6XHAKjHayHQWkphZGLsLXGNHE82-lWSt99SpM0KYHmZEuffv1SfBayVCRi4emnIiNpoXHvsEgN8Bd0yAl2hCEjB2-rJewcWsvnDbmjlgS_OmSwJqU9zFFm1NRZdEFLbdEE2HrG0fcq1sJ2OeZq9SHWNxuz11Y1_bDz-18gZzBvlWdOwYrTp1uPdojUNLW102mSMB6zMK_Dhxu6wIANxWmGri1jeIpgnpE8gOwtXFPTTug5at068rY0voUw76b40KEJdeJcWTupPYTKmQZTAweUe2qLP7X399VJMhg7hlk0GoXjfckAvll-ie7-wvrO-GgFrxFs8EidTaarPPjHI423SkaJiSjIFagVP8dzHU4FUwWlIuajQndIRWouzGEVys3JSBEHV9FEuLLaZ92lsgpD945F_I28rongF_FVlt398g-6A_ME)

### Sample cURL Command

To test the service locally, run the following command:

```bash
curl --location 'http://localhost:8080/weather?location=London'
```

## Expected Output

- **First Request (Cache MISS):**

  On the first request for a given location (e.g., London), the application will fetch fresh weather data from the Visual Crossing API and cache it in Redis. The server logs will indicate something like:

  ```
  Fetching weather data from: https://weather.visualcrossing.com/VisualCrossingWebServices/rest/services/timeline/London?key=your_visual_crossing_api_key&unitGroup=metric&include=days
  Fetched fresh weather data for location: London
  ```

  The JSON response returned to the client will contain the weather information returned by the Visual Crossing API.

- **Subsequent Request (Cache HIT):**

  If you run the same cURL command again before the cache expires, the service will retrieve data from Redis. The logs will now show:

  ```
  Serving cached weather data for location: London
  ```

  And the returned JSON should be identical to the previous call.

## About

This project demonstrates how to build a RESTful service with Go that integrates:

- 3rd party API consumption
- Redis caching with expiration
- Rate limiting (optional)
- Environment variable-based configuration

It serves as a great starting point for building more robust microservices that need to interact with external APIs while keeping performance and scalability in mind.

Happy coding!

## License

This project is open-source and free to use under the [MIT License](LICENSE). Contributions are welcome!
