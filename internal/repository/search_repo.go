package repository

import (
	"context"
	"database/sql"
	"encoding/json"

	"map-backend/internal/model"
)

type SearchRepository interface {
	SaveHistory(ctx context.Context, query string, parsedQuery *model.AIIntent) error
	GetCache(ctx context.Context, queryKey string) (*model.SearchResponse, error)
	SaveCache(ctx context.Context, queryKey string, response *model.SearchResponse) error
}

type searchRepositoryImpl struct {
	db *sql.DB
}

func NewSearchRepository(db *sql.DB) SearchRepository {
	return &searchRepositoryImpl{db: db}
}

func (r *searchRepositoryImpl) SaveHistory(ctx context.Context, query string, parsedQuery *model.AIIntent) error {
	querySQL := `INSERT INTO search_history (query, parsed_query) VALUES ($1, $2)`
	
	parsedJSON, err := json.Marshal(parsedQuery)
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx, querySQL, query, parsedJSON)
	return err
}

func (r *searchRepositoryImpl) GetCache(ctx context.Context, queryKey string) (*model.SearchResponse, error) {
	querySQL := `SELECT response FROM places_cache WHERE query_key = $1`
	
	var responseJSON []byte
	err := r.db.QueryRowContext(ctx, querySQL, queryKey).Scan(&responseJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No cache hit
		}
		return nil, err
	}

	var response model.SearchResponse
	if err := json.Unmarshal(responseJSON, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

func (r *searchRepositoryImpl) SaveCache(ctx context.Context, queryKey string, response *model.SearchResponse) error {
	querySQL := `
		INSERT INTO places_cache (query_key, response) 
		VALUES ($1, $2)
		ON CONFLICT (query_key) DO UPDATE SET response = EXCLUDED.response, created_at = NOW()
	`
	
	responseJSON, err := json.Marshal(response)
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx, querySQL, queryKey, responseJSON)
	return err
}
