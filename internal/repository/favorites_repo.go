package repository

import (
	"context"
	"database/sql"

	"map-backend/internal/model"
)

type FavoritesRepository interface {
	AddFavorite(ctx context.Context, userID string, favorite *model.FavoriteRequest) error
	GetFavorites(ctx context.Context, userID string) ([]model.Favorite, error)
}

type favoritesRepositoryImpl struct {
	db *sql.DB
}

func NewFavoritesRepository(db *sql.DB) FavoritesRepository {
	return &favoritesRepositoryImpl{db: db}
}

func (r *favoritesRepositoryImpl) AddFavorite(ctx context.Context, userID string, favorite *model.FavoriteRequest) error {
	querySQL := `
		INSERT INTO favorites (user_id, place_name, lat, lng, address) 
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.db.ExecContext(ctx, querySQL, userID, favorite.PlaceName, favorite.Lat, favorite.Lng, favorite.Address)
	return err
}

func (r *favoritesRepositoryImpl) GetFavorites(ctx context.Context, userID string) ([]model.Favorite, error) {
	querySQL := `
		SELECT id, user_id, place_name, lat, lng, address, created_at 
		FROM favorites 
		WHERE user_id = $1 
		ORDER BY created_at DESC
	`
	
	rows, err := r.db.QueryContext(ctx, querySQL, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var favorites []model.Favorite
	for rows.Next() {
		var f model.Favorite
		err := rows.Scan(&f.ID, &f.UserID, &f.PlaceName, &f.Lat, &f.Lng, &f.Address, &f.CreatedAt)
		if err != nil {
			return nil, err
		}
		favorites = append(favorites, f)
	}

	return favorites, nil
}
