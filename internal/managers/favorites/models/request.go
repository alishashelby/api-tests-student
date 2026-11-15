package models

// AddFavoriteRequest содержит идентификатор добавляемого объявления
type AddFavoriteRequest struct {
	AdvertisementID string `json:"advertisement_id"`
}
