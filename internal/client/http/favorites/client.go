package favorites

import (
	"fmt"
	"net/http"
	"strconv"
	"testing"

	"api-tests-template/internal/constants/path"
	apiRunner "api-tests-template/internal/helpers/api-runner"
	"api-tests-template/internal/managers/favorites/models"
)

// HttpPostFavorite отправляет запрос на добавление объявления в избранное
func HttpPostFavorite(t *testing.T, token, advertisementID string) *http.Response {
	request := models.AddFavoriteRequest{AdvertisementID: advertisementID}
	return apiRunner.GetRunner().Auth(token).Create().Post(path.FavoritesPath).
		ContentType("application/json").
		JSON(request).
		Expect(t).
		End().Response
}

// HttpGetFavorites отправляет запрос на получение списка избранного
func HttpGetFavorites(t *testing.T, token string, page, pageSize int) *http.Response {
	return apiRunner.GetRunner().Auth(token).Create().Get(path.FavoritesPath).
		Query("page", strconv.Itoa(page)).
		Query("page_size", strconv.Itoa(pageSize)).
		Expect(t).
		End().Response
}

// HttpDeleteFavorite отправляет запрос на удаление объявления из избранного
func HttpDeleteFavorite(t *testing.T, token, advertisementID string) *http.Response {
	favoritePath := fmt.Sprintf(path.FavoritePath, advertisementID)
	return apiRunner.GetRunner().Auth(token).Create().Delete(favoritePath).
		Expect(t).
		End().Response
}
