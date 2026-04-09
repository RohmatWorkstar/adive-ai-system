package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"map-backend/internal/client"
	"map-backend/internal/config"
	"map-backend/internal/handler"
	"map-backend/internal/repository"
	"map-backend/internal/service"
)

func main() {
	cfg := config.LoadConfig()

	// 1. Setup Database
	db, err := repository.NewPostgresDB(cfg.SupabaseDBURL)
	if err != nil {
		log.Printf("Warning: Failed to connect to database: %v. Running without DB.", err)
		// Note: in a real prod app, you might want to log.Fatalf here,
		// but allowing it to start for demo/testing without keys
	} else {
		defer db.Close()
		log.Println("Successfully connected to Supabase PostgreSQL")
	}

	// 2. Initialize Repositories
	searchRepo := repository.NewSearchRepository(db)
	favoritesRepo := repository.NewFavoritesRepository(db)

	// 3. Initialize Clients
	aiClient := client.NewAIClient(cfg.GroqAPIKey, cfg.GeminiAPIKey)
	googleClient := client.NewGoogleClient(cfg.GoogleMapsAPIKey)

	// 4. Initialize Services
	searchService := service.NewSearchService(searchRepo, aiClient, googleClient)
	favoritesService := service.NewFavoritesService(favoritesRepo)

	// 5. Initialize Handlers
	searchHandler := handler.NewSearchHandler(searchService)
	favoritesHandler := handler.NewFavoritesHandler(favoritesService)

	// 6. Init Router
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RealIP)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-User-ID"},
		AllowCredentials: true,
		MaxAge:           300, 
	}))

	// Routes
	r.Get("/api/health", handler.HealthCheck)
	
	r.Route("/api", func(r chi.Router) {
		r.Post("/search", searchHandler.Search)
		
		r.Route("/favorites", func(r chi.Router) {
			r.Get("/", favoritesHandler.GetFavorites)
			r.Post("/", favoritesHandler.AddFavorite)
		})
	})

	// Start Server
	log.Printf("Starting server on port %s in %s mode", cfg.Port, cfg.Env)
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
