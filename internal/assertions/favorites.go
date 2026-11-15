package assertions

import favoritesModels "api-tests-template/internal/managers/favorites/models"

// FindFavoriteByAdvertisementID ищет запись избранного по идентификатору объявления
func FindFavoriteByAdvertisementID(items []favoritesModels.Favorite, advertisementID string) (favoritesModels.Favorite, bool) {
	for _, item := range items {
		if item.Advertisement.ID == advertisementID {
			return item, true
		}
	}

	return favoritesModels.Favorite{}, false
}
