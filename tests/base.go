package tests

import (
	"log"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	advertisementsModels "api-tests-template/internal/managers/advertisements/models"
	"api-tests-template/internal/utils"
)

const (
	advertisementDescription = "Mini pig test advertisement description"
	advertisementPrice       = 100500.5
	advertisementQuantity    = 4
)

// SetupSuite действия перед запуском сьюта с тестами
func SetupSuite() {
	log.Println("Init environment variables")
	utils.LoadEnv()
}

// TearDownSuite действия после запуска сьюта с тестами
func TearDownSuite() {
	log.Println("Tear down suite")
}

// Precondition выводит текст в консоль с меткой [Precondition]
func Precondition(text string) {
	utils.LogWithLabelAndTimeStamp("Precondition", text)
}

// NewAdvertisementRequest создает валидные уникальные данные объявления
func NewAdvertisementRequest(photoPath string) advertisementsModels.CreateAdvertisementRequest {
	title := uuid.New().String()
	description := advertisementDescription
	price := advertisementPrice
	quantity := advertisementQuantity

	return advertisementsModels.CreateAdvertisementRequest{
		Title:       &title,
		Description: &description,
		Price:       &price,
		Quantity:    &quantity,
		PhotoPaths:  []string{photoPath},
	}
}

// TestPhotoPath возвращает путь к тестовой фотографии и проверяет файл
func TestPhotoPath(t *testing.T) string {
	t.Helper()

	_, currentFile, _, ok := runtime.Caller(0)
	require.True(t, ok, "не удалось определить путь к тестовым данным")

	photoPath := filepath.Join(filepath.Dir(currentFile), "scenarios", "advertisementCreation", "testData", "img.png")
	photoInfo, err := os.Stat(photoPath)
	require.NoError(t, err, "тестовая фотография отсутствует: %s", photoPath)
	require.False(t, photoInfo.IsDir(), "вместо тестовой фотографии найдена директория: %s", photoPath)
	require.Positive(t, photoInfo.Size(), "тестовая фотография пустая: %s", photoPath)

	return photoPath
}
