package handler

import (
	"encoding/json"
	"net/http"

	"map-backend/internal/model"
	"map-backend/internal/service"
)

type FavoritesHandler struct {
	service service.FavoritesService
}

func NewFavoritesHandler(service service.FavoritesService) *FavoritesHandler {
	return &FavoritesHandler{service: service}
}

func (h *FavoritesHandler) AddFavorite(w http.ResponseWriter, r *http.Request) {
	// In a real app, user_id would come from auth middleware
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		// Mock userID for testing if not provided
		userID = "test-user-123"
	}

	var req model.FavoriteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.service.AddFavorite(r.Context(), userID, &req); err != nil {
		h.errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (h *FavoritesHandler) GetFavorites(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		userID = "test-user-123"
	}

	favorites, err := h.service.GetFavorites(r.Context(), userID)
	if err != nil {
		h.errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(favorites)
}

func (h *FavoritesHandler) errorResponse(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
