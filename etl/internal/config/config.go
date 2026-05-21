package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	DBUrl                 string
	MicrosoftFilePath     string
	MicrosoftClientID     string
	MicrosoftClientSecret string
	MicrosoftTenantID     string
	ETLEnabled            bool
	ETLCron               string
	TargetPeriodMonth     time.Time
}

func Load() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("config: failed to read .env: %w", err)
	}

	dbURL := viper.GetString("BPCL_DB_URL")
	if dbURL == "" {
		return nil, fmt.Errorf("config: BPCL_DB_URL is required")
	}

	filePath := viper.GetString("BPCL_MICROSOFT_FILE_PATH")
	if filePath == "" {
		return nil, fmt.Errorf("config: BPCL_MICROSOFT_FILE_PATH is required")
	}

	clientID := viper.GetString("BPCL_MICROSOFT_CLIENT_ID")
	if clientID == "" {
		return nil, fmt.Errorf("config: BPCL_MICROSOFT_CLIENT_ID is required")
	}

	clientSecret := viper.GetString("BPCL_MICROSOFT_CLIENT_SECRET")
	if clientSecret == "" {
		return nil, fmt.Errorf("config: BPCL_MICROSOFT_CLIENT_SECRET is required")
	}

	tenantID := viper.GetString("BPCL_MICROSOFT_TENANT_ID")
	if tenantID == "" {
		return nil, fmt.Errorf("config: BPCL_MICROSOFT_TENANT_ID is required")
	}

	periodStr := viper.GetString("BPCL_ETL_TARGET_PERIOD")
	var targetPeriod time.Time
	if periodStr != "" {
		var err error
		targetPeriod, err = time.Parse("2006-01-02", periodStr)
		if err != nil {
			return nil, fmt.Errorf("config: BPCL_ETL_TARGET_PERIOD must be YYYY-MM-DD")
		}
	} else {
		targetPeriod = time.Date(2025, 5, 1, 0, 0, 0, 0, time.UTC)
	}

	return &Config{
		DBUrl:                 dbURL,
		MicrosoftFilePath:     filePath,
		MicrosoftClientID:     clientID,
		MicrosoftClientSecret: clientSecret,
		MicrosoftTenantID:     tenantID,
		ETLEnabled:            viper.GetBool("BPCL_ETL_ENABLED"),
		ETLCron:               viper.GetString("BPCL_ETL_CRON"),
		TargetPeriodMonth:     targetPeriod,
	}, nil
}