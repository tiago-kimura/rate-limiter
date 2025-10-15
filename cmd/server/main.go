package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/tiago-kimura/rate-limiter/internal/config"
	"github.com/tiago-kimura/rate-limiter/internal/middleware"
	"github.com/tiago-kimura/rate-limiter/internal/ratelimiter"
	"github.com/tiago-kimura/rate-limiter/internal/storage"
)

type Response struct {
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	IP        string    `json:"ip,omitempty"`
	Token     string    `json:"token,omitempty"`
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	redisStorage, err := storage.NewRedisStorage(cfg.RedisURL)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisStorage.Close()

	rateLimiter := ratelimiter.NewRateLimiter(redisStorage, cfg.GetIPConfig())

	for token, tokenConfig := range cfg.TokenConfigs {
		rateLimiter.SetTokenConfig(token, ratelimiter.Config{
			Limit:     tokenConfig.Limit,
			Window:    tokenConfig.Window,
			BlockTime: tokenConfig.BlockTime,
		})
	}

	rateLimiterMiddleware := middleware.NewRateLimiterMiddleware(rateLimiter)

	router := mux.NewRouter()

	router.Use(rateLimiterMiddleware.Handler)

	router.HandleFunc("/health", healthHandler).Methods("GET")

	router.HandleFunc("/", homeHandler).Methods("GET")
	router.HandleFunc("/api/test", testHandler).Methods("GET", "POST")
	router.HandleFunc("/api/data", dataHandler).Methods("GET")

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("Rate Limiter Server starting on port %s", cfg.Port)
	log.Printf("Redis URL: %s", cfg.RedisURL)
	log.Printf("IP Rate Limit: %d requests per %v", cfg.IPRateLimit, cfg.IPRateWindow)
	log.Printf("Token Rate Limit: %d requests per %v", cfg.TokenRateLimit, cfg.TokenRateWindow)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := Response{
		Message:   "Rate limiter service is healthy",
		Timestamp: time.Now(),
	}
	json.NewEncoder(w).Encode(response)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ip := getClientIP(r)
	token := r.Header.Get("API_KEY")

	response := Response{
		Message:   "Welcome to the Rate Limiter API",
		Timestamp: time.Now(),
		IP:        ip,
		Token:     token,
	}
	json.NewEncoder(w).Encode(response)
}

func testHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ip := getClientIP(r)
	token := r.Header.Get("API_KEY")

	response := Response{
		Message:   fmt.Sprintf("Test endpoint accessed via %s", r.Method),
		Timestamp: time.Now(),
		IP:        ip,
		Token:     token,
	}
	json.NewEncoder(w).Encode(response)
}

func dataHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ip := getClientIP(r)
	token := r.Header.Get("API_KEY")

	data := map[string]interface{}{
		"message":   "Data retrieved successfully",
		"timestamp": time.Now(),
		"ip":        ip,
		"token":     token,
		"data": []map[string]interface{}{
			{"id": 1, "name": "Item 1", "value": 100},
			{"id": 2, "name": "Item 2", "value": 200},
			{"id": 3, "name": "Item 3", "value": 300},
		},
	}

	json.NewEncoder(w).Encode(data)
}

func getClientIP(r *http.Request) string {
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		return forwarded
	}

	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	return r.RemoteAddr
}
