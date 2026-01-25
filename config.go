package main

import (
	"flag"
	"os"
	"time"
)

type Config struct {
	Port    string
	BaseURL string
	Timeout time.Duration
}

func LoadConfig() Config {
	port := flag.String("port", "", "Server port (e.g. :8000)")
	baseURL := flag.String("base-url", "", "API base URL (e.g. https://api.example.com)")
	timeout := flag.Duration("timeout", 0, "API timeout (e.g. 10s)")

	flag.Parse()

	return Config{
		Port:    getStringWithFallback(*port, "PORT", ":8000"),
		BaseURL: getStringWithFallback(*baseURL, "BASE_URL", "https://consumer-api.development.dev.woltapi.com/home-assignment-api/v1/venues/"),
		Timeout: getDurationWithFallback(*timeout, "TIMEOUT", 10*time.Second),
	}
}

func getStringWithFallback(cliFlag, envVar, defaultVal string) string {

	if cliFlag != "" {
		return cliFlag
	}

	if envVal := os.Getenv(envVar); envVal != "" {
		return envVal
	}

	return defaultVal
}

func getDurationWithFallback(cliFlag time.Duration, envVar string, defaultVal time.Duration) time.Duration {

	if cliFlag != 0 {
		return cliFlag
	}

	if envVal := os.Getenv(envVar); envVal != "" {
		if d, err := time.ParseDuration(envVal); err == nil {
			return d
		}
	}

	return defaultVal
}
