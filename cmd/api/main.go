package main

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"map-backend/internal/client"
	"map-backend/internal/config"
	"map-backend/internal/handler"
	"map-backend/internal/middleware"
	"map-backend/internal/repository"
	"map-backend/internal/service"
)

func main() {
	cfg := config.LoadConfig()

	// ── 1. Setup Database ──
	db, err := repository.NewPostgresDB(cfg.SupabaseDBURL)
	if err != nil {
		log.Printf("⚠️  Warning: Failed to connect to database: %v. Running without DB.", err)
		// Allow startup without DB for development/testing
	} else {
		defer db.Close()
		log.Println("✅ Successfully connected to Supabase PostgreSQL")
	}

	// ── 2. Initialize Repositories ──
	searchRepo := repository.NewSearchRepository(db)
	favoritesRepo := repository.NewFavoritesRepository(db)

	// ── 3. Initialize Clients ──
	grokClient := client.NewGrokClient(cfg.GrokAPIKey)
	geminiClient := client.NewGeminiClient(cfg.GeminiAPIKey)
	aiClient := client.NewAIClient(grokClient, geminiClient)
	fsqClient := client.NewFoursquareClient(cfg.FoursquareAPIKey)

	// Log AI provider availability
	if grokClient.IsAvailable() {
		log.Println("✅ Grok AI provider configured")
	} else {
		log.Println("⚠️  Grok AI provider not configured (missing GROK_API_KEY)")
	}
	if geminiClient.IsAvailable() {
		log.Println("✅ Gemini AI provider configured")
	} else {
		log.Println("⚠️  Gemini AI provider not configured (missing GEMINI_API_KEY)")
	}
	if cfg.FoursquareAPIKey != "" {
		log.Println("✅ Foursquare API configured")
	} else {
		log.Println("⚠️  Foursquare API not configured (missing FOURSQUARE_API_KEY)")
	}

	// ── 4. Initialize Services ──
	searchService := service.NewSearchService(searchRepo, aiClient, fsqClient)
	favoritesService := service.NewFavoritesService(favoritesRepo)

	// ── 5. Initialize Handlers ──
	searchHandler := handler.NewSearchHandler(searchService)
	favoritesHandler := handler.NewFavoritesHandler(favoritesService)

	// ── 6. Initialize Router ──
	r := chi.NewRouter()

	// Rate Limiter: 30 requests per minute per IP
	rateLimiter := middleware.NewRateLimiter(30, 1*time.Minute)

	// Middleware stack
	r.Use(chimw.RealIP)
	r.Use(middleware.RequestLogger)
	r.Use(chimw.Recoverer)
	r.Use(rateLimiter.Middleware)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-User-ID"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// ── 7. Routes ──
	r.Get("/api/health", handler.HealthCheck)

	r.Route("/api", func(r chi.Router) {
		r.Post("/search", searchHandler.Search)

		r.Route("/favorites", func(r chi.Router) {
			r.Get("/", favoritesHandler.GetFavorites)
			r.Post("/", favoritesHandler.AddFavorite)
		})
	})

	// ── 8. Start Server ──
	log.Printf("🚀 Starting AI Map Recommendation API on port %s [%s mode]", cfg.Port, cfg.Env)
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Fatalf("❌ Failed to start server: %v", err)
	}
}
