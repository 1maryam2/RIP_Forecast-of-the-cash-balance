package config

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt"
)

type JWTConfig struct {
	Token         string
	SigningMethod jwt.SigningMethod
	ExpiresIn     time.Duration
	RefreshExpiry time.Duration
}

type RedisConfig struct {
	Host        string
	Password    string
	Port        int
	User        string
	DialTimeout time.Duration
	ReadTimeout time.Duration
}

type Config struct {
	ServiceHost string
	ServicePort int
	JWT         JWTConfig
	Redis       RedisConfig
}

const (
	envServiceHost   = "SERVICE_HOST"
	envServicePort   = "SERVICE_PORT"
	envJWTToken      = "JWT_TOKEN"
	envJWTExpires    = "JWT_EXPIRES_IN"
	envJWTRefreshExp = "JWT_REFRESH_EXPIRES_IN"
	envRedisHost     = "REDIS_HOST"
	envRedisPort     = "REDIS_PORT"
	envRedisUser     = "REDIS_USER"
	envRedisPass     = "REDIS_PASSWORD"
)

func NewConfig(ctx context.Context) (*Config, error) {
	cfg := &Config{}

	cfg.ServiceHost = os.Getenv(envServiceHost)
	if cfg.ServiceHost == "" {
		cfg.ServiceHost = "0.0.0.0"
	}

	portStr := os.Getenv(envServicePort)
	if portStr == "" {
		cfg.ServicePort = 8080
	} else {
		port, err := strconv.Atoi(portStr)
		if err != nil {
			return nil, fmt.Errorf("service port must be int: %w", err)
		}
		cfg.ServicePort = port
	}
	cfg.JWT.Token = os.Getenv(envJWTToken)
	if cfg.JWT.Token == "" {
		cfg.JWT.Token = "test_secret"
	}

	expiresStr := os.Getenv(envJWTExpires)
	if expiresStr == "" {
		cfg.JWT.ExpiresIn = time.Hour
	} else {
		expires, err := strconv.Atoi(expiresStr)
		if err != nil {
			return nil, fmt.Errorf("jwt expires must be int: %w", err)
		}
		cfg.JWT.ExpiresIn = time.Duration(expires) * time.Second
	}
	refreshExpStr := os.Getenv(envJWTRefreshExp)
	if refreshExpStr == "" {
		cfg.JWT.RefreshExpiry = time.Hour * 24 * 7
	} else {
		refreshExp, err := strconv.Atoi(refreshExpStr)
		if err != nil {
			return nil, fmt.Errorf("jwt refresh expires must be int: %w", err)
		}
		cfg.JWT.RefreshExpiry = time.Duration(refreshExp) * time.Second
	}

	cfg.JWT.SigningMethod = jwt.SigningMethodHS256
	cfg.Redis.Host = os.Getenv(envRedisHost)
	if cfg.Redis.Host == "" {
		cfg.Redis.Host = "localhost"
	}

	redisPortStr := os.Getenv(envRedisPort)
	if redisPortStr == "" {
		cfg.Redis.Port = 6379
	} else {
		port, err := strconv.Atoi(redisPortStr)
		if err != nil {
			return nil, fmt.Errorf("redis port must be int: %w", err)
		}
		cfg.Redis.Port = port
	}

	cfg.Redis.User = os.Getenv(envRedisUser)
	cfg.Redis.Password = os.Getenv(envRedisPass)
	cfg.Redis.DialTimeout = 10 * time.Second
	cfg.Redis.ReadTimeout = 10 * time.Second

	return cfg, nil
}
