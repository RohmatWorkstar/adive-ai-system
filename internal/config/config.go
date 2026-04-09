package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port             string
	Env              string
	GrokAPIKey       string
	GeminiAPIKey     string
	FoursquareAPIKey string
	SupabaseDBURL    string
}

func LoadConfig() *Config {
	// Load .env file if it exists (useful for local development)
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	return &Config{
		Port:             getEnv("PORT", "8080"),
		Env:              getEnv("ENV", "development"),
		GrokAPIKey:       getEnv("GROK_API_KEY", ""),
		GeminiAPIKey:     getEnv("GEMINI_API_KEY", ""),
		FoursquareAPIKey: getEnv("FOURSQUARE_API_KEY", ""),
		SupabaseDBURL:    getEnv("SUPABASE_DB_URL", ""),
	}
}

func getEnv(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}
