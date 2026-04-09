package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                   string
	Env                    string
	GroqAPIKey             string
	GeminiAPIKey           string
	GoogleMapsAPIKey       string
	SupabaseDBURL          string
	SupabaseAnonKey        string
	SupabaseServiceRoleKey string
}

func LoadConfig() *Config {
	// Load .env file if it exists (useful for local development)
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	return &Config{
		Port:                   getEnv("PORT", "8080"),
		Env:                    getEnv("ENV", "development"),
		GroqAPIKey:             getEnv("GROQ_API_KEY", ""),
		GeminiAPIKey:           getEnv("GEMINI_API_KEY", ""),
		GoogleMapsAPIKey:       getEnv("GOOGLE_MAPS_API_KEY", ""),
		SupabaseDBURL:          getEnv("SUPABASE_DB_URL", ""),
		SupabaseAnonKey:        getEnv("SUPABASE_ANON_KEY", ""),
		SupabaseServiceRoleKey: getEnv("SUPABASE_SERVICE_ROLE_KEY", ""),
	}
}

func getEnv(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}
