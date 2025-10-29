package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)


type Config struct {
	DBName	string
	DBHost	string
	DBPassword	string
	DBUser	string
	DBPort	string
	ServerPort	string
	CountriesAPIURL	string
	ExchangeAPIURL	string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{	
		DBName: getEnv("DB_NAME", "country_currency_db"),
		DBHost: getEnv("DB_HOST", "localhost"),
		DBPassword: getEnv("DB_PASSWORD"),
		DBUser: getEnv("DB_USER", "root"),
		DBPort: getEnv("DB_PORT", "3306"),
		ServerPort: getEnv("PORT", "8080"),
		CountriesAPIURL: getEnv("COUNTRIES_API_URL", "https://restcountries.com/v2/all?fields=name,capital,region,population,flag,currencies"),
		ExchangeAPIURL: getEnv("EXCHANGE_API_URL", "https://open.er-api.com/v6/latest/USD"),
	}

	return cfg, nil
}


func getEnv(key string, defaultValue ...string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return ""
}


func (c *Config) Validate() error {
	if c.DBName == "" {
		return fmt.Errorf("DB_NAME is required")
	}
	if c.DBHost == "" {
		return fmt.Errorf("DB_HOST is required")
	}
	if c.DBUser == "" {
		return fmt.Errorf("DB_USER is required")
	}
	if c.DBPort == "" {
		return fmt.Errorf("DB_PORT is required")
	}
	if c.ServerPort == "" {
		return fmt.Errorf("PORT is required")
	}
	return nil
}
