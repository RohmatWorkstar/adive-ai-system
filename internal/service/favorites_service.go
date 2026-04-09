package service

import (
	"context"

	"map-backend/internal/model"
	"map-backend/internal/repository"
)

type FavoritesService interface {
	AddFavorite(ctx context.Context, userID string, favorite *model.FavoriteRequest) error
	GetFavorites(ctx context.Context, userID string) ([]model.Favorite, error)
}

type favoritesServiceImpl struct {
	repo repository.FavoritesRepository
}

func NewFavoritesService(repo repository.FavoritesRepository) FavoritesService {
	return &favoritesServiceImpl{repo: repo}
}

func (s *favoritesServiceImpl) AddFavorite(ctx context.Context, userID string, favorite *model.FavoriteRequest) error {
	return s.repo.AddFavorite(ctx, userID, favorite)
}

func (s *favoritesServiceImpl) GetFavorites(ctx context.Context, userID string) ([]model.Favorite, error) {
	return s.repo.GetFavorites(ctx, userID)
}
