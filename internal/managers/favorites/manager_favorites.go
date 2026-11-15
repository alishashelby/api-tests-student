package favorites

import (
	"net/http"
	"testing"

	favoritesClient "api-tests-template/internal/client/http/favorites"
	managerResponse "api-tests-template/internal/managers"
	"api-tests-template/internal/managers/favorites/models"
)

// AddFavorite добавляет объявление в избранное или возвращает модель ошибки
func AddFavorite(t *testing.T, token, advertisementID string, expectedStatusCode int) (models.Favorite, managerResponse.ErrorResponse) {
	rawResponse := favoritesClient.HttpPostFavorite(t, token, advertisementID)
	body := managerResponse.ReadBody(t, rawResponse, expectedStatusCode)

	if rawResponse.StatusCode == http.StatusCreated {
		return managerResponse.Unmarshal[models.Favorite](t, body), managerResponse.ErrorResponse{}
	}

	return models.Favorite{}, managerResponse.Unmarshal[managerResponse.ErrorResponse](t, body)
}

// GetFavorites возвращает страницу избранного или модель ошибки
func GetFavorites(t *testing.T, token string, page, pageSize, expectedStatusCode int) (models.FavoritesList, managerResponse.ErrorResponse) {
	rawResponse := favoritesClient.HttpGetFavorites(t, token, page, pageSize)
	body := managerResponse.ReadBody(t, rawResponse, expectedStatusCode)

	if rawResponse.StatusCode == http.StatusOK {
		return managerResponse.Unmarshal[models.FavoritesList](t, body), managerResponse.ErrorResponse{}
	}

	return models.FavoritesList{}, managerResponse.Unmarshal[managerResponse.ErrorResponse](t, body)
}

// DeleteFavorite удаляет объявление из избранного или возвращает модель ошибки
func DeleteFavorite(t *testing.T, token, advertisementID string, expectedStatusCodes ...int) managerResponse.ErrorResponse {
	rawResponse := favoritesClient.HttpDeleteFavorite(t, token, advertisementID)
	body := managerResponse.ReadBody(t, rawResponse, expectedStatusCodes...)

	if rawResponse.StatusCode == http.StatusNoContent {
		return managerResponse.ErrorResponse{}
	}

	return managerResponse.Unmarshal[managerResponse.ErrorResponse](t, body)
}
