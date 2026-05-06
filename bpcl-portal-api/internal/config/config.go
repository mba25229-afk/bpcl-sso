package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	DBUrl        string
	JWTSecret    string
	JWTExpiry    time.Duration
	Port         string
	UploadDir    string
	CORSOrigins  string
	MaxUploadMB  int
	LogLevel     string
	Env          string
	RateLimitRPM int
}

func Load() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("config: failed to read .env: %w", err)
	}

	required := []string{"BPCL_DB_URL", "BPCL_JWT_SECRET", "BPCL_PORT"}
	for _, key := range required {
		if viper.GetString(key) == "" {
			return nil, fmt.Errorf("config: required env var %s is missing or empty", key)
		}
	}

	expiry, err := time.ParseDuration(viper.GetString("BPCL_JWT_EXPIRY"))
	if err != nil {
		expiry = 24 * time.Hour
	}

	return &Config{
		DBUrl:        viper.GetString("BPCL_DB_URL"),
		JWTSecret:    viper.GetString("BPCL_JWT_SECRET"),
		JWTExpiry:    expiry,
		Port:         viper.GetString("BPCL_PORT"),
		UploadDir:    viper.GetString("BPCL_UPLOAD_DIR"),
		CORSOrigins:  viper.GetString("BPCL_CORS_ORIGINS"),
		MaxUploadMB:  viper.GetInt("BPCL_MAX_UPLOAD_MB"),
		LogLevel:     viper.GetString("BPCL_LOG_LEVEL"),
		Env:          viper.GetString("BPCL_ENV"),
		RateLimitRPM: viper.GetInt("BPCL_RATE_LIMIT_RPM"),
	}, nil
}

func (c *Config) Validate() error {
	if c.Env == "production" && len(c.JWTSecret) < 32 {
		return fmt.Errorf("config: BPCL_JWT_SECRET must be at least 32 characters in production")
	}
	if c.Env == "production" && c.UploadDir == "/tmp" {
		return fmt.Errorf("config: BPCL_UPLOAD_DIR cannot be /tmp in production")
	}
	return nil
}
