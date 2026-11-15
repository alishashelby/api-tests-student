package advertisementCreation

import (
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"api-tests-template/internal/assertions"
	advertisementsManager "api-tests-template/internal/managers/advertisements"
	advertisementsModels "api-tests-template/internal/managers/advertisements/models"
	"api-tests-template/internal/managers/auth"
	authModels "api-tests-template/internal/managers/auth/models"
	"api-tests-template/internal/utils"
	base "api-tests-template/tests"
)

const (
	invalidToken   = "incorrect_token"
	searchTimeout  = 10 * time.Second
	searchInterval = 200 * time.Millisecond
)

type TestSuite struct {
	suite.Suite
	loginData authModels.LoginOkResponse
	photoPath string
}

func TestSuiteRun(t *testing.T) {
	suite.Run(t, &TestSuite{})
}

func (s *TestSuite) SetupSuite() {
	base.SetupSuite()

	base.Precondition("Авторизация пользователя и подготовка тестовой фотографии")
	s.loginData = auth.Login(s.T(), os.Getenv("TEST_LOGIN"), os.Getenv("TEST_PASSWORD"))
	s.photoPath = base.TestPhotoPath(s.T())
}

func (s *TestSuite) TearDownSuite() {
	base.TearDownSuite()
}

func (s *TestSuite) TestCreateAdvertisementPositive() {
	request := base.NewAdvertisementRequest(s.photoPath)
	testContext := s.T()

	var createdAdvertisement advertisementsModels.Advertisement
	s.Run("Создаем объявление со всеми параметрами", func() {
		var errorResponse advertisementsModels.ErrorResponse
		createdAdvertisement, errorResponse = advertisementsManager.CreateAdvertisement(
			s.T(),
			s.loginData.Token,
			request,
			http.StatusCreated,
		)

		require.Empty(s.T(), errorResponse.Error, "в позитивном ответе присутствует тип ошибки")
		require.Empty(s.T(), errorResponse.Message, "в позитивном ответе присутствует сообщение об ошибке")
		require.Nil(s.T(), errorResponse.Details, "в позитивном ответе присутствуют детали ошибки")
		require.NotEmpty(s.T(), createdAdvertisement.ID, "сервер не вернул id созданного объявления")

		registerAdvertisementCleanup(testContext, s.loginData.Token, createdAdvertisement.ID)
		assertAdvertisementMatchesRequest(s.T(), createdAdvertisement, request)

		s.T().Logf("Объявление %s успешно создано", createdAdvertisement.ID)
	})

	s.Run("Получаем объявление по id и сверяем его параметры", func() {
		advertisement := advertisementsManager.GetAdvertisement(s.T(), createdAdvertisement.ID, http.StatusOK)

		assertAdvertisementMatchesRequest(s.T(), advertisement, request)
		require.Equal(s.T(), createdAdvertisement.ID, advertisement.ID, "id полученного объявления не совпадает")
		require.Equal(s.T(), s.loginData.User.Id, advertisement.UserID, "user_id объявления не совпадает")

		s.T().Logf("Параметры объявления %s совпадают с запросом", advertisement.ID)
	})

	s.Run("Ждем появления объявления в поиске по полному названию", func() {
		var foundAdvertisement advertisementsModels.Advertisement
		utils.WaitUntil(s.T(), searchTimeout, searchInterval, "объявление не появилось в поиске", func() bool {
			list := advertisementsManager.GetAdvertisementsList(s.T(), s.loginData.Token, *request.Title)
			var found bool
			foundAdvertisement, found = assertions.FindAdvertisementByID(list.Items, createdAdvertisement.ID)
			return found
		})

		assertAdvertisementMatchesRequest(s.T(), foundAdvertisement, request)
		s.T().Logf("Объявление %s найдено по полному названию", foundAdvertisement.ID)
	})

	s.Run("Проверяем загруженную фотографию", func() {
		photos := advertisementsManager.GetAdvertisementPhotos(s.T(), s.loginData.Token, createdAdvertisement.ID, http.StatusOK)

		require.Len(s.T(), photos, 1, "количество фотографий объявления не совпадает")
		require.NotEmpty(s.T(), photos[0].ID, "id фотографии пустой")
		require.NotEmpty(s.T(), photos[0].URL, "URL фотографии пустой")
		require.NotEmpty(s.T(), photos[0].CreatedAt, "время создания фотографии пустое")

		s.T().Logf("Фотография %s доступна через API", photos[0].ID)
	})
}

func (s *TestSuite) TestCreateAdvertisementValidation() {
	testCases := []struct {
		name            string
		changeRequest   func(*advertisementsModels.CreateAdvertisementRequest)
		expectedError   string
		expectedMessage string
	}{
		{
			name: "отсутствует title",
			changeRequest: func(request *advertisementsModels.CreateAdvertisementRequest) {
				request.Title = nil
			},
			expectedError:   "invalid_request",
			expectedMessage: "title is required",
		},
		{
			name: "пустой title",
			changeRequest: func(request *advertisementsModels.CreateAdvertisementRequest) {
				emptyTitle := ""
				request.Title = &emptyTitle
			},
			expectedError:   "invalid_request",
			expectedMessage: "title is required",
		},
		{
			name: "слишком длинный title",
			changeRequest: func(request *advertisementsModels.CreateAdvertisementRequest) {
				longTitle := strings.Repeat("a", 51)
				request.Title = &longTitle
			},
			expectedError:   "invalid_request",
			expectedMessage: "title must not exceed 50 characters provided 51",
		},
		{
			name: "отсутствует description",
			changeRequest: func(request *advertisementsModels.CreateAdvertisementRequest) {
				request.Description = nil
			},
			expectedError:   "invalid_request",
			expectedMessage: "description is required",
		},
		{
			name: "отрицательная цена",
			changeRequest: func(request *advertisementsModels.CreateAdvertisementRequest) {
				negativePrice := -1.0
				request.Price = &negativePrice
			},
			expectedError:   "invalid_request",
			expectedMessage: "price must be between 0 and 1000000 inclusive",
		},
		{
			name: "нулевое количество",
			changeRequest: func(request *advertisementsModels.CreateAdvertisementRequest) {
				zeroQuantity := 0
				request.Quantity = &zeroQuantity
			},
			expectedError:   "invalid_request",
			expectedMessage: "quantity must be between 1 and 100 inclusive",
		},
		{
			name: "отсутствует фотография",
			changeRequest: func(request *advertisementsModels.CreateAdvertisementRequest) {
				request.PhotoPaths = nil
			},
			expectedError:   "invalid_request",
			expectedMessage: "at least 1 photo is required",
		},
	}

	for _, testCase := range testCases {
		s.Run(testCase.name, func() {
			request := base.NewAdvertisementRequest(s.photoPath)
			testCase.changeRequest(&request)

			unexpectedAdvertisement, errorResponse := advertisementsManager.CreateAdvertisement(
				s.T(),
				s.loginData.Token,
				request,
				http.StatusBadRequest,
			)

			require.Empty(s.T(), unexpectedAdvertisement.ID, "при ошибке сервер вернул id объявления")
			require.Equal(s.T(), testCase.expectedError, errorResponse.Error, "тип ошибки не совпадает")
			require.Equal(s.T(), testCase.expectedMessage, errorResponse.Message, "сообщение ошибки не совпадает")

			s.T().Logf("Сервер отклонил запрос ожидаемой ошибкой: %s", errorResponse.Message)
		})
	}
}

func (s *TestSuite) TestCreateAdvertisementUnauthorized() {
	testCases := []struct {
		name            string
		token           string
		expectedMessage string
	}{
		{name: "токен отсутствует", token: "", expectedMessage: "Authorization header required"},
		{name: "токен неверный", token: invalidToken, expectedMessage: "Invalid or expired token"},
	}

	for _, testCase := range testCases {
		s.Run(testCase.name, func() {
			unexpectedAdvertisement, errorResponse := advertisementsManager.CreateAdvertisement(
				s.T(),
				testCase.token,
				base.NewAdvertisementRequest(s.photoPath),
				http.StatusUnauthorized,
			)

			require.Empty(s.T(), unexpectedAdvertisement.ID, "без авторизации сервер вернул id объявления")
			require.Equal(s.T(), "unauthorized", errorResponse.Error, "тип ошибки не совпадает")
			require.Equal(s.T(), testCase.expectedMessage, errorResponse.Message, "сообщение ошибки не совпадает")

			s.T().Logf("Сервер вернул ожидаемый 401: %s", errorResponse.Message)
		})
	}
}

// assertAdvertisementMatchesRequest сравнивает объявление с исходным запросом
func assertAdvertisementMatchesRequest(
	t *testing.T,
	advertisement advertisementsModels.Advertisement,
	request advertisementsModels.CreateAdvertisementRequest,
) {
	t.Helper()
	require.Equal(t, *request.Title, advertisement.Title, "title объявления не совпадает")
	require.Equal(t, *request.Description, advertisement.Description, "description объявления не совпадает")
	require.InDelta(t, *request.Price, advertisement.Price, 0.001, "price объявления не совпадает")
	require.Equal(t, *request.Quantity, advertisement.Quantity, "quantity объявления не совпадает")
}

// registerAdvertisementCleanup регистрирует резервное удаление объявления
func registerAdvertisementCleanup(t *testing.T, token, advertisementID string) {
	t.Helper()
	t.Cleanup(func() {
		advertisementsManager.DeleteAdvertisement(
			t,
			token,
			advertisementID,
			http.StatusNoContent,
			http.StatusNotFound,
		)
	})
}
