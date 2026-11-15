package advertisements

import (
	"fmt"
	"net/http"
	"strconv"
	"testing"

	"api-tests-template/internal/constants/path"
	apiRunner "api-tests-template/internal/helpers/api-runner"
	"api-tests-template/internal/managers/advertisements/models"
)

func HttpGetAdvertisements(t *testing.T, token, textForSearch string) *http.Response {
	return apiRunner.GetRunner().Auth(token).Create().
		Get(path.AdvertisementsPath).
		Query("search", textForSearch).
		Expect(t).
		Status(http.StatusOK).
		End().Response
}

// HttpPostAdvertisement отправляет запрос на создание объявления
func HttpPostAdvertisement(
	t *testing.T,
	token string,
	request models.CreateAdvertisementRequest,
) *http.Response {
	httpRequest := apiRunner.GetRunner().Auth(token).Create().Post(path.AdvertisementPath)

	if request.Title != nil {
		httpRequest.MultipartFormData("title", *request.Title)
	}
	if request.Description != nil {
		httpRequest.MultipartFormData("description", *request.Description)
	}
	if request.Price != nil {
		httpRequest.MultipartFormData("price", strconv.FormatFloat(*request.Price, 'f', -1, 64))
	}
	if request.Quantity != nil {
		httpRequest.MultipartFormData("quantity", strconv.Itoa(*request.Quantity))
	}
	if len(request.PhotoPaths) > 0 {
		httpRequest.MultipartFile("photos", request.PhotoPaths...)
	}

	return httpRequest.Expect(t).End().Response
}

// HttpGetAdvertisement отправляет запрос на получение объявления по идентификатору
func HttpGetAdvertisement(t *testing.T, id string) *http.Response {
	return apiRunner.GetRunner().Create().Get(path.AdvertisementPath).
		Query("id", id).
		Expect(t).
		End().Response
}

// HttpDeleteAdvertisement отправляет запрос на удаление объявления
func HttpDeleteAdvertisement(t *testing.T, token, id string) *http.Response {
	return apiRunner.GetRunner().Auth(token).Create().Delete(path.AdvertisementPath).
		Query("id", id).
		Expect(t).
		End().Response
}

// HttpGetAdvertisementPhotos отправляет запрос на получение фотографий объявления
func HttpGetAdvertisementPhotos(t *testing.T, token, id string) *http.Response {
	photosPath := fmt.Sprintf(path.AdvertisementPhotosPath, id)

	return apiRunner.GetRunner().Auth(token).Create().Get(photosPath).
		Expect(t).
		End().Response
}
