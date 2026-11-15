package favorites

import (
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"api-tests-template/internal/assertions"
	managerResponse "api-tests-template/internal/managers"
	advertisementsManager "api-tests-template/internal/managers/advertisements"
	advertisementsModels "api-tests-template/internal/managers/advertisements/models"
	"api-tests-template/internal/managers/auth"
	authModels "api-tests-template/internal/managers/auth/models"
	favoritesManager "api-tests-template/internal/managers/favorites"
	"api-tests-template/internal/utils"
	base "api-tests-template/tests"
)

const (
	invalidToken      = "incorrect_token"
	favoritesPage     = 1
	favoritesPageSize = 50
	waitTimeout       = 10 * time.Second
	waitInterval      = 200 * time.Millisecond
)

type TestSuite struct {
	suite.Suite
	loginData       authModels.LoginOkResponse
	helperLoginData authModels.LoginOkResponse
	photoPath       string
}

func TestSuiteRun(t *testing.T) {
	suite.Run(t, &TestSuite{})
}

func (s *TestSuite) SetupSuite() {
	base.SetupSuite()

	base.Precondition("Авторизация основного и вспомогательного пользователей")
	s.loginData = auth.Login(s.T(), os.Getenv("TEST_LOGIN"), os.Getenv("TEST_PASSWORD"))
	s.helperLoginData = auth.Login(s.T(), os.Getenv("TEST_FAVORITES_LOGIN"), os.Getenv("TEST_FAVORITES_PASSWORD"))
	require.NotEqual(s.T(), s.loginData.User.Id, s.helperLoginData.User.Id, "основной и вспомогательный аккаунты должны быть разными")

	s.photoPath = base.TestPhotoPath(s.T())
}

func (s *TestSuite) TearDownSuite() {
	base.TearDownSuite()
}

func (s *TestSuite) TestFavoritePositive() {
	testContext := s.T()
	var disableAdvertisementCleanup func()
	var disableFavoriteCleanup func()

	var advertisement advertisementsModels.Advertisement
	s.Run("Создаем объявление основного пользователя", func() {
		advertisement = createAdvertisement(s.T(), s.loginData.Token, s.photoPath)
		disableAdvertisementCleanup = registerAdvertisementCleanup(testContext, s.loginData.Token, advertisement.ID)
	})

	s.Run("Добавляем чужое объявление в избранное", func() {
		disableFavoriteCleanup = registerFavoriteCleanup(testContext, s.helperLoginData.Token, advertisement.ID)
		favorite, errorResponse := favoritesManager.AddFavorite(s.T(), s.helperLoginData.Token, advertisement.ID, http.StatusCreated)

		requireEmptyErrorResponse(s.T(), errorResponse)
		require.NotEmpty(s.T(), favorite.FavoriteID, "сервер не вернул favorite_id")
		require.Equal(s.T(), advertisement.ID, favorite.Advertisement.ID, "id объявления в избранном не совпадает")
	})

	s.Run("Проверяем объявление в списке избранного", func() {
		list, errorResponse := favoritesManager.GetFavorites(s.T(), s.helperLoginData.Token, favoritesPage, favoritesPageSize, http.StatusOK)

		requireEmptyErrorResponse(s.T(), errorResponse)
		favorite, found := assertions.FindFavoriteByAdvertisementID(list.Items, advertisement.ID)

		require.True(s.T(), found, "объявление отсутствует в избранном")
		require.NotEmpty(s.T(), favorite.AddedAt, "added_at избранного пустой")
	})

	s.Run("Удаляем объявление из избранного", func() {
		errorResponse := favoritesManager.DeleteFavorite(s.T(), s.helperLoginData.Token, advertisement.ID, http.StatusNoContent)
		requireEmptyErrorResponse(s.T(), errorResponse)
		if disableFavoriteCleanup != nil {
			disableFavoriteCleanup()
		}
	})

	s.Run("Проверяем отсутствие удаленного избранного", func() {
		waitForFavoriteState(s.T(), s.helperLoginData.Token, advertisement.ID, false)
	})

	s.Run("Повторно добавляем объявление в избранное", func() {
		disableFavoriteCleanup = registerFavoriteCleanup(testContext, s.helperLoginData.Token, advertisement.ID)
		favorite, errorResponse := favoritesManager.AddFavorite(s.T(), s.helperLoginData.Token, advertisement.ID, http.StatusCreated)

		requireEmptyErrorResponse(s.T(), errorResponse)
		require.NotEmpty(s.T(), favorite.FavoriteID, "сервер не вернул favorite_id")
	})

	s.Run("Удаляем само объявление", func() {
		advertisementsManager.DeleteAdvertisement(s.T(), s.loginData.Token, advertisement.ID, http.StatusNoContent)
		if disableAdvertisementCleanup != nil {
			disableAdvertisementCleanup()
		}
	})

	s.Run("Проверяем каскадное удаление из избранного", func() {
		waitForFavoriteState(s.T(), s.helperLoginData.Token, advertisement.ID, false)
		if disableFavoriteCleanup != nil {
			disableFavoriteCleanup()
		}
	})
}

func (s *TestSuite) TestFavoriteNotFound() {
	missingAdvertisementID := uuid.New().String()

	s.Run("Добавляем несуществующее объявление", func() {
		unexpectedFavorite, errorResponse := favoritesManager.AddFavorite(s.T(), s.helperLoginData.Token, missingAdvertisementID, http.StatusNotFound)

		require.Empty(s.T(), unexpectedFavorite.FavoriteID, "сервер вернул favorite_id для несуществующего объявления")
		require.Equal(s.T(), "not_found", errorResponse.Error, "тип ошибки не совпадает")
		require.Equal(s.T(), "Advertisement not found", errorResponse.Message, "сообщение ошибки не совпадает")
	})

	s.Run("Удаляем несуществующее избранное", func() {
		errorResponse := favoritesManager.DeleteFavorite(s.T(), s.helperLoginData.Token, missingAdvertisementID, http.StatusNotFound)

		require.Equal(s.T(), "not_found", errorResponse.Error, "тип ошибки не совпадает")
		require.Equal(s.T(), "Favorite entry not found", errorResponse.Message, "сообщение ошибки не совпадает")
	})
}

func (s *TestSuite) TestFavoriteUnauthorized() {
	tokens := []struct {
		name            string
		value           string
		expectedMessage string
	}{
		{name: "токен отсутствует", value: "", expectedMessage: "Authorization header required"},
		{name: "токен неверный", value: invalidToken, expectedMessage: "Invalid or expired token"},
	}

	for _, token := range tokens {
		s.Run(token.name, func() {
			s.Run("POST favorites", func() {
				unexpectedFavorite, errorResponse := favoritesManager.AddFavorite(s.T(), token.value, uuid.New().String(), http.StatusUnauthorized)

				require.Empty(s.T(), unexpectedFavorite.FavoriteID, "без авторизации сервер вернул favorite_id")
				requireUnauthorized(s.T(), errorResponse, token.expectedMessage)
			})

			s.Run("GET favorites", func() {
				unexpectedList, errorResponse := favoritesManager.GetFavorites(s.T(), token.value, favoritesPage, favoritesPageSize, http.StatusUnauthorized)

				require.Empty(s.T(), unexpectedList.Items, "без авторизации сервер вернул список избранного")
				requireUnauthorized(s.T(), errorResponse, token.expectedMessage)
			})

			s.Run("DELETE favorites", func() {
				errorResponse := favoritesManager.DeleteFavorite(s.T(), token.value, uuid.New().String(), http.StatusUnauthorized)
				requireUnauthorized(s.T(), errorResponse, token.expectedMessage)
			})
		})
	}
}

func (s *TestSuite) TestFavoriteValidation() {
	invalidAdvertisementID := "invalid-advertisement-id"

	unexpectedFavorite, errorResponse := favoritesManager.AddFavorite(
		s.T(),
		s.helperLoginData.Token,
		invalidAdvertisementID,
		http.StatusBadRequest,
	)

	require.Empty(s.T(), unexpectedFavorite.FavoriteID, "сервер вернул favorite_id для некорректного advertisement_id")
	require.Equal(s.T(), "invalid_request", errorResponse.Error, "тип ошибки не совпадает")
	require.Equal(s.T(), "invalid UUID length: 24", errorResponse.Message, "сообщение ошибки не совпадает")
}

func (s *TestSuite) TestFavoriteBusinessRestrictions() {
	s.Run("Нельзя добавить свое объявление", func() {
		testContext := s.T()
		advertisement := createAdvertisement(s.T(), s.loginData.Token, s.photoPath)
		registerAdvertisementCleanup(testContext, s.loginData.Token, advertisement.ID)

		unexpectedFavorite, errorResponse := favoritesManager.AddFavorite(s.T(), s.loginData.Token, advertisement.ID, http.StatusForbidden)

		require.Empty(s.T(), unexpectedFavorite.FavoriteID, "сервер вернул favorite_id для своего объявления")
		require.Equal(s.T(), "forbidden", errorResponse.Error, "тип ошибки не совпадает")
		require.Equal(s.T(), "Cannot add own advertisement to favorites", errorResponse.Message, "сообщение ошибки не совпадает")
	})

	s.Run("Нельзя повторно добавить объявление", func() {
		testContext := s.T()
		advertisement := createAdvertisement(s.T(), s.loginData.Token, s.photoPath)
		registerAdvertisementCleanup(testContext, s.loginData.Token, advertisement.ID)
		registerFavoriteCleanup(testContext, s.helperLoginData.Token, advertisement.ID)

		favorite, errorResponse := favoritesManager.AddFavorite(s.T(), s.helperLoginData.Token, advertisement.ID, http.StatusCreated)

		requireEmptyErrorResponse(s.T(), errorResponse)
		require.NotEmpty(s.T(), favorite.FavoriteID, "сервер не вернул favorite_id")

		unexpectedFavorite, duplicateError := favoritesManager.AddFavorite(s.T(), s.helperLoginData.Token, advertisement.ID, http.StatusConflict)

		require.Empty(s.T(), unexpectedFavorite.FavoriteID, "при конфликте сервер вернул новый favorite_id")
		require.Equal(s.T(), "already_exists", duplicateError.Error, "тип ошибки не совпадает")
		require.Equal(s.T(), "Advertisement is already in favorites", duplicateError.Message, "сообщение ошибки не совпадает")
	})
}

// createAdvertisement создает объявление для предусловий сценария
func createAdvertisement(t *testing.T, token, photoPath string) advertisementsModels.Advertisement {
	t.Helper()

	advertisement, errorResponse := advertisementsManager.CreateAdvertisement(t, token, base.NewAdvertisementRequest(photoPath), http.StatusCreated)

	requireEmptyErrorResponse(t, errorResponse)
	require.NotEmpty(t, advertisement.ID, "сервер не вернул id объявления")

	return advertisement
}

// waitForFavoriteState ожидает нужное состояние объявления в избранном
func waitForFavoriteState(t *testing.T, token, advertisementID string, expected bool) {
	t.Helper()

	utils.WaitUntil(t, waitTimeout, waitInterval, "состояние избранного не обновилось", func() bool {
		list, errorResponse := favoritesManager.GetFavorites(
			t,
			token,
			favoritesPage,
			favoritesPageSize,
			http.StatusOK,
		)
		requireEmptyErrorResponse(t, errorResponse)
		_, found := assertions.FindFavoriteByAdvertisementID(list.Items, advertisementID)
		return found == expected
	})
}

// requireEmptyErrorResponse проверяет отсутствие ошибки в позитивном ответе
func requireEmptyErrorResponse(t *testing.T, errorResponse managerResponse.ErrorResponse) {
	t.Helper()
	require.Empty(t, errorResponse.Error, "в позитивном ответе присутствует тип ошибки")
	require.Empty(t, errorResponse.Message, "в позитивном ответе присутствует сообщение ошибки")
	require.Nil(t, errorResponse.Details, "в позитивном ответе присутствуют детали ошибки")
}

// requireUnauthorized проверяет стандартный ответ авторизации
func requireUnauthorized(t *testing.T, errorResponse managerResponse.ErrorResponse, expectedMessage string) {
	t.Helper()
	require.Equal(t, "unauthorized", errorResponse.Error, "тип ошибки не совпадает")
	require.Equal(t, expectedMessage, errorResponse.Message, "сообщение ошибки не совпадает")
}

// registerAdvertisementCleanup регистрирует резервное удаление объявления
func registerAdvertisementCleanup(t *testing.T, token, advertisementID string) func() {
	t.Helper()
	enabled := true
	t.Cleanup(func() {
		if enabled {
			advertisementsManager.DeleteAdvertisement(t, token, advertisementID, http.StatusNoContent, http.StatusNotFound)
		}
	})
	return func() {
		enabled = false
	}
}

// registerFavoriteCleanup регистрирует резервное удаление избранного
func registerFavoriteCleanup(t *testing.T, token, advertisementID string) func() {
	t.Helper()
	enabled := true
	t.Cleanup(func() {
		if enabled {
			favoritesManager.DeleteFavorite(t, token, advertisementID, http.StatusNoContent, http.StatusNotFound)
		}
	})
	return func() {
		enabled = false
	}
}
