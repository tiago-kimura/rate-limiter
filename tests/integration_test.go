package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tiago-kimura/rate-limiter/internal/middleware"
	"github.com/tiago-kimura/rate-limiter/internal/ratelimiter"
	"github.com/tiago-kimura/rate-limiter/internal/storage"
)

func TestRateLimiterMiddleware_IPLimiting(t *testing.T) {
	mockStorage := storage.NewMockStorage()
	config := ratelimiter.Config{
		Limit:     3,
		Window:    time.Second,
		BlockTime: time.Minute,
	}

	rateLimiter := ratelimiter.NewRateLimiter(mockStorage, config)
	middleware := middleware.NewRateLimiterMiddleware(rateLimiter)

	router := mux.NewRouter()
	router.Use(middleware.Handler)
	router.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "success"}`))
	}).Methods("GET")

	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"

		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.Contains(t, recorder.Header().Get("X-RateLimit-Limit"), "3")
		assert.Equal(t, fmt.Sprintf("%d", 3-i-1), recorder.Header().Get("X-RateLimit-Remaining"))
		assert.Equal(t, "ip", recorder.Header().Get("X-RateLimit-Type"))
	}

	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusTooManyRequests, recorder.Code)
	assert.Equal(t, "0", recorder.Header().Get("X-RateLimit-Remaining"))

	var response map[string]interface{}
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "you have reached the maximum number of requests or actions allowed within a certain time frame", response["message"])
}

func TestRateLimiterMiddleware_TokenLimiting(t *testing.T) {
	mockStorage := storage.NewMockStorage()
	ipConfig := ratelimiter.Config{
		Limit:     2,
		Window:    time.Second,
		BlockTime: time.Minute,
	}

	rateLimiter := ratelimiter.NewRateLimiter(mockStorage, ipConfig)

	tokenConfig := ratelimiter.Config{
		Limit:     5,
		Window:    time.Second,
		BlockTime: time.Minute,
	}
	rateLimiter.SetTokenConfig("test_token", tokenConfig)

	middleware := middleware.NewRateLimiterMiddleware(rateLimiter)

	router := mux.NewRouter()
	router.Use(middleware.Handler)
	router.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "success"}`))
	}).Methods("GET")

	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		req.Header.Set("API_KEY", "test_token")

		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.Equal(t, "token", recorder.Header().Get("X-RateLimit-Type"))
		assert.Equal(t, fmt.Sprintf("%d", 5-i-1), recorder.Header().Get("X-RateLimit-Remaining"))
	}

	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	req.Header.Set("API_KEY", "test_token")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusTooManyRequests, recorder.Code)
	assert.Equal(t, "token", recorder.Header().Get("X-RateLimit-Type"))
}

func TestRateLimiterMiddleware_DifferentIPs(t *testing.T) {
	mockStorage := storage.NewMockStorage()
	config := ratelimiter.Config{
		Limit:     2,
		Window:    time.Second,
		BlockTime: time.Minute,
	}

	rateLimiter := ratelimiter.NewRateLimiter(mockStorage, config)
	middleware := middleware.NewRateLimiterMiddleware(rateLimiter)

	router := mux.NewRouter()
	router.Use(middleware.Handler)
	router.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "success"}`))
	}).Methods("GET")

	ips := []string{"192.168.1.1:12345", "192.168.1.2:12345", "192.168.1.3:12345"}

	for _, ip := range ips {
		for i := 0; i < 2; i++ {
			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = ip

			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)

			assert.Equal(t, http.StatusOK, recorder.Code, "IP %s request %d should be allowed", ip, i+1)
		}

		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = ip

		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusTooManyRequests, recorder.Code, "IP %s 3rd request should be blocked", ip)
	}
}

func TestRateLimiterMiddleware_XForwardedFor(t *testing.T) {
	mockStorage := storage.NewMockStorage()
	config := ratelimiter.Config{
		Limit:     2,
		Window:    time.Second,
		BlockTime: time.Minute,
	}

	rateLimiter := ratelimiter.NewRateLimiter(mockStorage, config)
	middleware := middleware.NewRateLimiterMiddleware(rateLimiter)

	router := mux.NewRouter()
	router.Use(middleware.Handler)
	router.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "success"}`))
	}).Methods("GET")

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "10.0.0.1:12345"
		req.Header.Set("X-Forwarded-For", "203.0.113.1")

		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusOK, recorder.Code)
	}

	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "10.0.0.1:12345"
	req.Header.Set("X-Forwarded-For", "203.0.113.1")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusTooManyRequests, recorder.Code)
}
