package config

import (
	"bufio"
	"errors"
	"os"
	"strings"
)

type Config struct {
	Host          string
	Port          string
	FrontendURL   string
	OpenRouterKey string
	AppName       string
}

func Load() Config {
	_ = loadEnvFile(".env")
	_ = loadEnvFile("../.env")

	return Config{
		Host:          valueOrDefault("HOST", "127.0.0.1"),
		Port:          valueOrDefault("PORT", "8080"),
		FrontendURL:   valueOrDefault("FRONTEND_URL", "http://localhost:5173"),
		OpenRouterKey: strings.TrimSpace(os.Getenv("OPENROUTER_API_KEY")),
		AppName:       valueOrDefault("OPENROUTER_APP_NAME", "Allai"),
	}
}

func loadEnvFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}

		key = strings.TrimSpace(key)
		value = strings.Trim(strings.TrimSpace(value), "\"'")
		if key != "" {
			if _, exists := os.LookupEnv(key); !exists {
				_ = os.Setenv(key, value)
			}
		}
	}

	return scanner.Err()
}

func valueOrDefault(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}
