package models

import advertisementsModels "api-tests-template/internal/managers/advertisements/models"

// Favorite описывает запись избранного
type Favorite struct {
	FavoriteID    string                             `json:"favorite_id"`
	AddedAt       string                             `json:"added_at"`
	Advertisement advertisementsModels.Advertisement `json:"advertisement"`
}

// FavoritesList описывает страницу избранного
type FavoritesList struct {
	Items      []Favorite `json:"items"`
	Page       int        `json:"page"`
	PageSize   int        `json:"page_size"`
	Total      int        `json:"total"`
	TotalPages int        `json:"total_pages"`
}
