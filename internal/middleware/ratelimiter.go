package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/tiago-kimura/rate-limiter/internal/ratelimiter"
)

type RateLimiterMiddleware struct {
	rateLimiter *ratelimiter.RateLimiter
}

func NewRateLimiterMiddleware(rateLimiter *ratelimiter.RateLimiter) *RateLimiterMiddleware {
	return &RateLimiterMiddleware{
		rateLimiter: rateLimiter,
	}
}

type ErrorResponse struct {
	Message string `json:"message"`
	Error   string `json:"error"`
}

func (m *RateLimiterMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		ip := getClientIP(r)

		apiKey := r.Header.Get("API_KEY")

		result, err := m.rateLimiter.CheckLimit(ctx, ip, apiKey)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", result.Limit))
		w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", result.Remaining))
		w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", result.ResetTime.Unix()))
		w.Header().Set("X-RateLimit-Type", string(result.LimitType))

		if !result.Allowed {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)

			response := ErrorResponse{
				Message: "you have reached the maximum number of requests or actions allowed within a certain time frame",
				Error:   "rate_limit_exceeded",
			}

			json.NewEncoder(w).Encode(response)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func getClientIP(r *http.Request) string {
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		ips := strings.Split(forwarded, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	return ip
}
