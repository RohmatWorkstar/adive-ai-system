# Travida — AI Map Recommendation Backend

A production-ready Go REST API that accepts natural language queries, uses multi-AI providers (Grok + Gemini) to extract structured intent, fetches real place data from **Foursquare API**, and returns clean, frontend-ready responses.

## Architecture

```
cmd/
  api/
    main.go                  # Entry point, dependency wiring

internal/
  client/
    ai_client.go             # AI orchestrator (Grok → Gemini → keyword fallback)
    grok_client.go           # Grok AI provider (via Groq API)
    gemini_client.go         # Gemini AI provider (OpenAI-compatible)
    foursquare_client.go     # Foursquare Places API v3 client
  config/
    config.go                # Environment config loader
  handler/
    search.go                # POST /api/search
    favorites.go             # GET/POST /api/favorites
    health.go                # GET /api/health
  middleware/
    rate_limiter.go          # Token-bucket rate limiter (30 req/min/IP)
    logging.go               # Structured request logging
  model/
    domain.go                # Data types (Place, AIIntent, SearchResponse, etc.)
  repository/
    db.go                    # PostgreSQL connection pooling
    search_repo.go           # Search cache & history (24h TTL)
    favorites_repo.go        # User favorites CRUD
  service/
    search_service.go        # Core search logic + in-memory cache
    favorites_service.go     # Favorites business logic
  utils/
    utils.go                 # Query normalization, ranking algorithm

sql/
  schema.sql                 # Database schema (Supabase PostgreSQL)
```

## Tech Stack

- **Language**: Go 1.25
- **Router**: [chi](https://github.com/go-chi/chi) v5
- **Database**: PostgreSQL (Supabase)
- **AI Providers**: Grok (Groq API) + Gemini (Google AI)
- **Places API**: Foursquare v3

## Quick Start

### 1. Configure Environment

```bash
cp .env.example .env
# Edit .env with your API keys
```

### 2. Install & Run

```bash
go mod tidy
go run cmd/api/main.go
```

### 3. Test

```bash
# Health check
curl http://localhost:8080/api/health

# Search
curl -X POST http://localhost:8080/api/search \
  -H "Content-Type: application/json" \
  -d '{"query": "tempat makan murah di bekasi"}'
```

## API Endpoints

| Method | Path              | Description                  |
|--------|-------------------|------------------------------|
| GET    | `/api/health`     | Health check                 |
| POST   | `/api/search`     | AI-powered place search      |
| GET    | `/api/favorites`  | Get user favorites           |
| POST   | `/api/favorites`  | Add a favorite place         |

### POST /api/search

**Request:**
```json
{
  "query": "tempat makan murah di bekasi"
}
```

**Response:**
```json
{
  "places": [
    {
      "name": "Warung Sederhana",
      "rating": 8.5,
      "address": "Jl. Ahmad Yani, Bekasi",
      "lat": -6.2383,
      "lng": 106.9756,
      "category": "Restaurant"
    }
  ],
  "summary": "These are popular and affordable dining spots near Bekasi."
}
```

## Multi-AI Strategy

| Purpose            | Primary | Fallback | Final Fallback       |
|--------------------|---------|----------|----------------------|
| Intent Extraction  | Grok    | Gemini   | Keyword extraction   |
| Summary Generation | Gemini  | Grok     | Static message       |

## Environment Variables

| Variable             | Required | Description                    |
|----------------------|----------|--------------------------------|
| `GROK_API_KEY`       | Optional | Groq API key (primary AI)      |
| `GEMINI_API_KEY`     | Optional | Google Gemini API key           |
| `FOURSQUARE_API_KEY` | Yes      | Foursquare Places API key       |
| `SUPABASE_DB_URL`    | Optional | PostgreSQL connection string    |
| `PORT`               | No       | Server port (default: 8080)     |
| `ENV`                | No       | Environment (default: development) |
