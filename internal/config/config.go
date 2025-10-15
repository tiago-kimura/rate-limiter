package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"github.com/tiago-kimura/rate-limiter/internal/ratelimiter"
)

type Config struct {
	Port string

	RedisURL string

	IPRateLimit     int64
	IPRateWindow    time.Duration
	IPBlockTime     time.Duration
	TokenRateLimit  int64
	TokenRateWindow time.Duration
	TokenBlockTime  time.Duration

	TokenConfigs map[string]TokenConfig
}

type TokenConfig struct {
	Limit     int64
	Window    time.Duration
	BlockTime time.Duration
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	config := &Config{
		Port:     getEnvString("PORT", "8080"),
		RedisURL: getEnvString("REDIS_URL", "redis://localhost:6379/0"),

		IPRateLimit:     getEnvInt64("IP_RATE_LIMIT", 10),
		IPRateWindow:    getEnvDuration("IP_RATE_WINDOW", "1s"),
		IPBlockTime:     getEnvDuration("IP_BLOCK_TIME", "5m"),
		TokenRateLimit:  getEnvInt64("TOKEN_RATE_LIMIT", 100),
		TokenRateWindow: getEnvDuration("TOKEN_RATE_WINDOW", "1s"),
		TokenBlockTime:  getEnvDuration("TOKEN_BLOCK_TIME", "5m"),

		TokenConfigs: make(map[string]TokenConfig),
	}

	config.loadTokenConfigs()

	return config, nil
}

func (c *Config) loadTokenConfigs() {
	for _, env := range os.Environ() {
		if len(env) > 6 && env[:6] == "TOKEN_" {
			parts := parseTokenEnvVar(env)
			if len(parts) >= 3 && parts[2] == "LIMIT" {
				token := parts[1]
				if token != "" {
					limit := getEnvInt64(fmt.Sprintf("TOKEN_%s_LIMIT", token), c.TokenRateLimit)
					window := getEnvDuration(fmt.Sprintf("TOKEN_%s_WINDOW", token), c.TokenRateWindow.String())
					blockTime := getEnvDuration(fmt.Sprintf("TOKEN_%s_BLOCK_TIME", token), c.TokenBlockTime.String())

					c.TokenConfigs[token] = TokenConfig{
						Limit:     limit,
						Window:    window,
						BlockTime: blockTime,
					}
				}
			}
		}
	}
}

func parseTokenEnvVar(env string) []string {
	equalIndex := -1
	for i, char := range env {
		if char == '=' {
			equalIndex = i
			break
		}
	}

	if equalIndex == -1 {
		return nil
	}

	varName := env[:equalIndex]
	parts := []string{}
	current := ""
	underscoreCount := 0

	for _, char := range varName {
		if char == '_' {
			underscoreCount++
			if underscoreCount <= 2 || current == "" {
				if current != "" {
					parts = append(parts, current)
					current = ""
				}
			} else {
				current += string(char)
			}
		} else {
			current += string(char)
		}
	}

	if current != "" {
		parts = append(parts, current)
	}

	return parts
}

func (c *Config) GetIPConfig() ratelimiter.Config {
	return ratelimiter.Config{
		Limit:     c.IPRateLimit,
		Window:    c.IPRateWindow,
		BlockTime: c.IPBlockTime,
	}
}

func (c *Config) GetTokenConfig(token string) (ratelimiter.Config, bool) {
	if tokenConfig, exists := c.TokenConfigs[token]; exists {
		return ratelimiter.Config{
			Limit:     tokenConfig.Limit,
			Window:    tokenConfig.Window,
			BlockTime: tokenConfig.BlockTime,
		}, true
	}

	return ratelimiter.Config{
		Limit:     c.TokenRateLimit,
		Window:    c.TokenRateWindow,
		BlockTime: c.TokenBlockTime,
	}, false
}


func getEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue string) time.Duration {
	value := getEnvString(key, defaultValue)
	if duration, err := time.ParseDuration(value); err == nil {
		return duration
	}
	if seconds, err := strconv.ParseFloat(value, 64); err == nil {
		return time.Duration(seconds * float64(time.Second))
	}
	if duration, err := time.ParseDuration(defaultValue); err == nil {
		return duration
	}
	return time.Second
}
