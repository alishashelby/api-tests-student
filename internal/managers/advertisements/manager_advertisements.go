package advertisements

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	advertisementsClient "api-tests-template/internal/client/http/advertisements"
	managerResponse "api-tests-template/internal/managers"
	"api-tests-template/internal/managers/advertisements/models"
)

func GetAdvertisementsList(t *testing.T, token, text string) models.AdvertisementsList {
	rawResponse := advertisementsClient.HttpGetAdvertisements(t, token, text)
	defer rawResponse.Body.Close()

	body, err := io.ReadAll(rawResponse.Body)
	require.NoError(t, err, "не удалось прочитать тело ответа поиска объявлений")

	response := models.AdvertisementsList{}
	require.NoError(t, json.Unmarshal(body, &response), "не удалось разобрать JSON ответа поиска объявлений")

	return response
}

// CreateAdvertisement создает объявление или возвращает модель ошибки
func CreateAdvertisement(
	t *testing.T,
	token string,
	request models.CreateAdvertisementRequest,
	expectedStatusCode int,
) (models.Advertisement, models.ErrorResponse) {
	rawResponse := advertisementsClient.HttpPostAdvertisement(t, token, request)
	body := managerResponse.ReadBody(t, rawResponse, expectedStatusCode)

	if rawResponse.StatusCode == http.StatusCreated {
		return managerResponse.Unmarshal[models.Advertisement](t, body), models.ErrorResponse{}
	}

	return models.Advertisement{}, managerResponse.Unmarshal[models.ErrorResponse](t, body)
}

// GetAdvertisement возвращает объявление по идентификатору
func GetAdvertisement(t *testing.T, id string, expectedStatusCode int) models.Advertisement {
	rawResponse := advertisementsClient.HttpGetAdvertisement(t, id)
	body := managerResponse.ReadBody(t, rawResponse, expectedStatusCode)

	return managerResponse.Unmarshal[models.Advertisement](t, body)
}

// DeleteAdvertisement удаляет объявление и проверяет статус ответа
func DeleteAdvertisement(t *testing.T, token, id string, expectedStatusCodes ...int) {
	rawResponse := advertisementsClient.HttpDeleteAdvertisement(t, token, id)
	_ = managerResponse.ReadBody(t, rawResponse, expectedStatusCodes...)
}

// GetAdvertisementPhotos возвращает фотографии объявления
func GetAdvertisementPhotos(t *testing.T, token, id string, expectedStatusCode int) []models.Photo {
	rawResponse := advertisementsClient.HttpGetAdvertisementPhotos(t, token, id)
	body := managerResponse.ReadBody(t, rawResponse, expectedStatusCode)

	return managerResponse.Unmarshal[[]models.Photo](t, body)
}
