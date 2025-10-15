# Rate Limiter

A robust and configurable rate limiter implemented in Go that can limit requests based on IP address or access token.

## 🚀 Features

- **IP-based Limiting**: Controls requests by IP address
- **Token-based Limiting**: Allows custom limits for specific tokens
- **Token Precedence**: Token configurations override IP limitations
- **Storage Strategy**: Flexible interface with Redis implementation
- **HTTP Middleware**: Easy integration with any HTTP server
- **Flexible Configuration**: Via environment variables or .env file
- **Docker Ready**: Includes Dockerfile and docker-compose
- **Comprehensive Testing**: Unit and integration tests

## 📋 Requirements

- Go 1.21+
- Redis (for rate limiter data storage)
- Docker and Docker Compose (optional)

## 🏗️ Architecture

```
├── cmd/server/          # Main application
├── internal/
│   ├── config/         # Configuration management
│   ├── middleware/     # Rate limiter HTTP middleware
│   ├── ratelimiter/    # Core rate limiter logic
│   └── storage/        # Storage interface and implementations
├── tests/              # Integration tests
├── scripts/            # Test and utility scripts
└── docker-compose.yml  # Docker configuration
```

## ⚙️ Configuration

### Environment Variables

Create a `.env` file in the project root or configure the following environment variables:

```env
# Server Configuration
PORT=8080

# Redis Configuration
REDIS_URL=redis://localhost:6379/0

# IP Rate Limiting
IP_RATE_LIMIT=10          # Maximum requests per second per IP
IP_RATE_WINDOW=1s         # Time window for counting
IP_BLOCK_TIME=5m          # Block time after exceeding limit

# Token Rate Limiting (default)
TOKEN_RATE_LIMIT=100      # Default limit for tokens
TOKEN_RATE_WINDOW=1s      # Default time window
TOKEN_BLOCK_TIME=5m       # Default block time

# Token-specific configurations
TOKEN_abc123_LIMIT=50
TOKEN_abc123_WINDOW=1s
TOKEN_abc123_BLOCK_TIME=10m

TOKEN_vip_token_LIMIT=1000
TOKEN_vip_token_WINDOW=1s
TOKEN_vip_token_BLOCK_TIME=1m
```

### Time Formats

- **Seconds**: `1s`, `30s`
- **Minutes**: `1m`, `5m`, `30m`
- **Hours**: `1h`, `2h`
- **Combinations**: `1h30m`, `2m30s`

## 🚀 Running

### With Docker (Recommended)

```bash
# Start all services (Redis + Rate Limiter)
make run

# Build only
make build

# Clean up containers and images
make clean
```

### Local Execution

1. **Start Redis:**
```bash
docker run -d -p 6379:6379 redis:7-alpine
```

2. **Run application:**
```bash
go run cmd/server/main.go
```

## 🔧 Usage

### Request Headers

To use token-based limiting, include the header:
```
API_KEY: your_token_here
```

### Response Headers

The rate limiter adds the following headers to responses:

```
X-RateLimit-Limit: 10        # Maximum limit
X-RateLimit-Remaining: 7     # Remaining requests
X-RateLimit-Reset: 1634567890 # Unix timestamp for reset
X-RateLimit-Type: ip         # Limiting type (ip/token)
```

### Available Endpoints

- `GET /health` - Health check
- `GET /` - Main endpoint
- `GET|POST /api/test` - Test endpoint
- `GET /api/data` - Data endpoint

### Rate Limit Exceeded Response

When the limit is exceeded, the API returns:

**Status Code:** `429 Too Many Requests`

**Response Body:**
```json
{
  "message": "you have reached the maximum number of requests or actions allowed within a certain time frame",
  "error": "rate_limit_exceeded"
}
```

## 🧪 Testing

### Run All Tests

```bash
make test
```

### Load Testing

Execute the load test script:

```bash
# Make sure the server is running
make run

# Run load test
./scripts/load_test.sh
```

## 📊 Functional Testing Examples

### Example 1: IP-based Limiting

```bash
# Make multiple requests from the same IP
for i in {1..15}; do
  curl -i http://localhost:8080/api/test
  echo "Request $i completed"
done
```

### Example 2: Token-based Limiting

```bash
# With specific token
curl -H "API_KEY: abc123" http://localhost:8080/api/test

# Without token (uses IP limiting)
curl http://localhost:8080/api/test
```

### Example 3: Different IPs Testing

```bash
# Simulate different IPs with X-Forwarded-For
curl -H "X-Forwarded-For: 192.168.1.100" http://localhost:8080/api/test
curl -H "X-Forwarded-For: 192.168.1.101" http://localhost:8080/api/test
```

## 🛠️ Development

### Code Structure

1. **Storage Interface** (`internal/storage/interface.go`):
   - Defines interface for data persistence
   - Allows easy switching between Redis and other implementations

2. **Rate Limiter Core** (`internal/ratelimiter/ratelimiter.go`):
   - Main rate limiting logic
   - Separated from middleware for reusability

3. **HTTP Middleware** (`internal/middleware/ratelimiter.go`):
   - Integration with HTTP servers
   - IP and token extraction
   - Response header addition

4. **Configuration** (`internal/config/config.go`):
   - Environment variable loading
   - Token-specific configuration parsing

### Adding New Storage Implementation

1. Implement the `storage.Storage` interface
2. Add the new implementation in `internal/storage/`
3. Update initialization in `cmd/server/main.go`

### Example New Implementation

```go
type MemoryStorage struct {
    // in-memory implementation
}

func (m *MemoryStorage) Get(ctx context.Context, key string) (int64, error) {
    // implementation
}

func (m *MemoryStorage) Increment(ctx context.Context, key string, expiration time.Duration) (int64, error) {
    // implementation
}

// ... other interface methods
```

## 📦 Available Make Commands

- `make run` - Start all services with Docker Compose (builds if needed)
- `make test` - Run all tests with coverage and race detection
- `make build` - Build Docker image
- `make clean` - Stop containers and clean up Docker resources
